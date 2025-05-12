package handlers

import (
	"auth_service/logger"
	"auth_service/tokens"
	"context"
	"encoding/base64"
	"net/http"
	"time"
)

type AuthContextKey struct {
	Val string
}

type responseWrapper struct {
	http.ResponseWriter
	status      int
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		if tokenHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		token, err := tokens.VerifyToken(tokenHeader)
		if err != nil {
			http.Error(w, "Token verification failed", http.StatusUnauthorized)
			return
		}

		refreshCookie, err := ExtractRefreshCookie(w, r)
		if err != nil { 
			return 
		}

		refreshBytes, _ := base64.StdEncoding.DecodeString(refreshCookie.Value)
		refreshPlain := string(refreshBytes)

		claims := token.Claims.(*tokens.JwtClaims)
		ctx := context.WithValue(r.Context(), AuthContextKey{Val: "refreshToken"}, refreshPlain)
        ctx = context.WithValue(ctx, AuthContextKey{Val: "user"}, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWrapper{ ResponseWriter: w, status: 200 }
		next.ServeHTTP(wrapper, r)
		logger.Info.Printf("[%s] %s - %d, %s %s\n",
			r.Method, r.URL.Path, wrapper.status, r.RemoteAddr, time.Since(start))
	})
}
