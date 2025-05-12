package handlers

import (
	"auth_service/database"
	"auth_service/logger"
	"auth_service/tokens"
	"bytes"
	"fmt"
	"os"
	"time"

	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SuccessAuthResponse struct {
	Jwt string `json:"jwt"`
}

type DeauthorizeResponse struct {
	Msg string `json:"msg"`
}

type UserIdResponse struct {
	Id string `json:"guid"`
}

type AuthHandler struct {
	router *mux.Router
	repository database.ITokenRepository
}

func NewAuthHandler(router *mux.Router, repository database.ITokenRepository) *AuthHandler {
	return &AuthHandler{
		router: router,
		repository: repository,
	}
}

func (h *AuthHandler) SetUpRoutes() {
	uuidRegexp := `[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}`

	refreshHandler := AuthMiddleware(http.HandlerFunc(h.RefreshTokens))
	userIdHandler := AuthMiddleware(http.HandlerFunc(h.GetUserId))
	deauthorizationHandler := AuthMiddleware(http.HandlerFunc(h.Deauthorize))

	h.router.HandleFunc(fmt.Sprintf("/authorize/{userId:%s}", uuidRegexp), h.AuthorizeById).Methods("POST")
	h.router.Handle("/refresh_tokens", refreshHandler).Methods("POST")
	h.router.Handle("/guid", userIdHandler).Methods("GET")
	h.router.Handle("/deauthorize", deauthorizationHandler).Methods("POST")
}

func (h *AuthHandler) AuthorizeById(w http.ResponseWriter, r *http.Request) {
	userIdstr := mux.Vars(r)["userId"]
	userAgent := r.UserAgent()
	ip := GetClientIp(r)

	accessToken, err := tokens.CreateToken(ip, userAgent, userIdstr)
	if err != nil {
		http.Error(w, "can't generate access token:" + err.Error(), http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(userIdstr)
	if err != nil {
		http.Error(w, "invalid userId", http.StatusBadRequest)
		return
	}

	refreshToken := &tokens.RefreshToken{
		Id: uuid.New(),
		UserId: userId,
		ClientIp: ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 1, 0),
	}
	
	refreshToken.GenerateToken()

	if err := h.repository.SaveRefreshToken(r.Context(), refreshToken); err != nil {
		switch {
		case database.IsUniqueConstraintError(err):
			http.Error(w, "this user has been already authorized", http.StatusForbidden)
		default:
			http.Error(w, "authorization failed", http.StatusInternalServerError)
		}
		return
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(refreshToken.Token))

	refreshTokenCookie := http.Cookie{
		Name: "refresh",
		Value: encoded,
		HttpOnly: true,
		Secure: true,
		Path: "/",
	}

	http.SetCookie(w, &refreshTokenCookie)
	RenderJSON(w, SuccessAuthResponse{Jwt: accessToken})
}

func (h *AuthHandler) findRefreshToken(userId string, w http.ResponseWriter, r *http.Request) (*tokens.RefreshToken, error) {
	refreshPlain := r.Context().Value(AuthContextKey{Val: "refreshToken"}).(string)
	refreshToken, err := h.repository.GetRefreshToken(r.Context(), refreshPlain)
	if err != nil {
		http.Error(w, "wrong refresh token", http.StatusUnauthorized)
		return nil, err
	}

	if userId != refreshToken.UserId.String() {
		http.Error(w, "jwt subject is different from refresh token creator", http.StatusUnauthorized)
		return nil, fmt.Errorf("wrong sub")
	}
	return refreshToken, nil
}

func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(AuthContextKey{Val: "user"}).(string)

	refreshFromDB, err := h.findRefreshToken(userId, w, r)
	if err != nil { return }

	userAgent := r.UserAgent()
	ip := GetClientIp(r)
	if userAgent != refreshFromDB.UserAgent {
		http.Error(w, "user agent has changed", http.StatusUnauthorized)
		return
	}

	if refreshFromDB.ExpiresAt.Before(time.Now()) {
		http.Error(w, "refresh token has expired", http.StatusUnauthorized)
		return
	}

	if refreshFromDB.ClientIp != ip {
		postBody, _ := json.Marshal(map[string]string{
			"userId":  userId,
			"newIp": ip,
		})
		req, _ := http.NewRequest("POST", os.Getenv("IP_CHANGE_WEBHOOK"), bytes.NewBuffer(postBody))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logger.Err.Printf("Error making request: %v\n", err)
		}
		defer resp.Body.Close()
		refreshFromDB.ClientIp = ip
	}

	if err := h.repository.DeleteRefreshToken(r.Context(), refreshFromDB.Id); err != nil {
		http.Error(w, "refresh operation failed", http.StatusInternalServerError)
		return
	}

	refreshFromDB.Id = uuid.New()
	refreshFromDB.CreatedAt = time.Now()
	refreshFromDB.ExpiresAt = time.Now().AddDate(0, 1, 0)
	refreshFromDB.GenerateToken()

	if err := h.repository.SaveRefreshToken(r.Context(), refreshFromDB); err != nil {
		http.Error(w, "refresh operation failed", http.StatusInternalServerError)
		return
	}

	newAccessToken, err := tokens.CreateToken(ip, userAgent, userId)
	if err != nil {
		http.Error(w, "can't generate access token:" + err.Error(), http.StatusInternalServerError)
		return
	}

	updRefreshCookie := http.Cookie{
		Name: "refresh",
		Value: base64.StdEncoding.EncodeToString([]byte(refreshFromDB.Token)),
		HttpOnly: true,
		Secure: true,
		Path: "/",
	}

	http.SetCookie(w, &updRefreshCookie)
	RenderJSON(w, SuccessAuthResponse{Jwt: newAccessToken})
}

func (h *AuthHandler) GetUserId(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(AuthContextKey{Val: "user"}).(string)
	_, err := h.findRefreshToken(userId, w, r)
	if err != nil { return }
	RenderJSON(w, UserIdResponse{Id: userId})
}

func (h *AuthHandler) Deauthorize(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(AuthContextKey{Val: "user"}).(string)

	refreshFromDB, err := h.findRefreshToken(userId, w, r)
	if err != nil { return }

	if err := h.repository.DeleteRefreshToken(r.Context(), refreshFromDB.Id); err != nil {
		http.Error(w, "deauthorization failed: can't delete refresh token", http.StatusInternalServerError)
		return
	}

	updRefreshCookie := http.Cookie{
		Name: "refresh",
		Value: base64.StdEncoding.EncodeToString([]byte(refreshFromDB.Token)),
		MaxAge: -1,
		HttpOnly: true,
		Secure: true,
		Path: "/",
	}

	http.SetCookie(w, &updRefreshCookie)
	w.WriteHeader(http.StatusUnauthorized)
	RenderJSON(w, DeauthorizeResponse{Msg: "successfully deauthorized"})
}
