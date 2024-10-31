package middleware

import (
	"net/http"
)

type AuthMiddleware struct {
	accessKey string
}

func NewAuthMiddleware(accessKey string) *AuthMiddleware {
	return &AuthMiddleware{
		accessKey: accessKey,
	}
}

func (a *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for static files
		if r.URL.Path == "/" || r.URL.Path == "/favicon.ico" || r.URL.Path == "/index.html" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("ACCESS_KEY")
		if authHeader == "" {
			http.Error(w, "Unauthorized - Missing ACCESS_KEY header", http.StatusUnauthorized)
			return
		}

		if authHeader != a.accessKey {
			http.Error(w, "Unauthorized - Invalid ACCESS_KEY", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
