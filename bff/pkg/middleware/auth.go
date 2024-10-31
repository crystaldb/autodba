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
		// Only authenticate requests to /api paths
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			authHeader := r.Header.Get("ACCESS_KEY")
			if authHeader == "" {
				http.Error(w, "Unauthorized - Missing ACCESS_KEY header", http.StatusUnauthorized)
				return
			}

			if authHeader != a.accessKey {
				http.Error(w, "Unauthorized - Invalid ACCESS_KEY", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
