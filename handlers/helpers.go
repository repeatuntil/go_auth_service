package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func GetClientIp(r *http.Request) (ip string) {
	ipList := r.Header.Get("X-Forwarded-For")
	if ipList != "" {
		ip = strings.Split(ipList, ",")[0]
	} else {
		ip = r.RemoteAddr
	}
	return ip[:strings.IndexRune(ip, ':')]
}

func RenderJSON(w http.ResponseWriter, object interface{}) {
	js, err := json.Marshal(object)
	if err != nil {
		http.Error(w, "Can't render JSON from object:", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func ExtractRefreshCookie(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	tokenCookie, err := r.Cookie("refresh")
	if err != nil {
		switch {
        case errors.Is(err, http.ErrNoCookie):
            http.Error(w, "refresh token missing", http.StatusUnauthorized)
        default:
           	fmt.Println(err)
            http.Error(w, "server error", http.StatusInternalServerError)
        }
	}
	return tokenCookie, err
}
