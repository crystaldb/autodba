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
			authHeader := r.Header.Get("Autodba-Access-Key")
			if authHeader == "" {
				http.Error(w, "Unauthorized - Missing Autodba-Access-Key header", http.StatusUnauthorized)
				return
			}

			if authHeader != a.accessKey {
				http.Error(w, "Unauthorized - Invalid Autodba-Access-Key", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
