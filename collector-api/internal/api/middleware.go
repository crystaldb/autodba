package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs all incoming HTTP requests if debugging is enabled
func LoggingMiddleware(next http.Handler, debug bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if debug {
			start := time.Now()

			// Log the incoming request details
			log.Printf("Started %s %s %s", r.Method, r.RequestURI, r.RemoteAddr)

			// Log headers
			for name, values := range r.Header {
				for _, value := range values {
					log.Printf("Header: %s = %s", name, value)
				}
			}

			// Log request body if it's a JSON request
			if r.Header.Get("Content-Type") == "application/json" {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					log.Printf("Request Body: %s", string(body))

					// Reset the body so it can be read again in the handler
					r.Body = io.NopCloser(bytes.NewBuffer(body))
				} else {
					log.Printf("Error reading body: %v", err)
				}
			}

			// Call the next handler in the chain
			next.ServeHTTP(w, r)

			// Log the request completion
			log.Printf("Completed in %v", time.Since(start))
		} else {
			// If debug is false, skip logging and just call the next handler
			next.ServeHTTP(w, r)
		}
	})
}
