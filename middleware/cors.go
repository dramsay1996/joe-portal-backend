package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment variable
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:3000" // Default to Next.js development server
		}

		// Get the origin from the request
		origin := r.Header.Get("Origin")
		log.Printf("CORS: Received request from origin: %s", origin)

		// Check if the origin is in the allowed list
		allowed := false
		for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(allowedOrigin) == origin {
				allowed = true
				break
			}
		}

		if allowed {
			log.Printf("CORS: Allowing origin: %s", origin)
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
		} else {
			log.Printf("CORS: Rejecting origin: %s (not in allowed list: %s)", origin, allowedOrigins)
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Handling preflight request for %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
