package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs incoming HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the request
		log.Printf(
			"Started %s %s from %s (Origin: %s)",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.Header.Get("Origin"),
		)

		// Create a custom response writer to capture the status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log the response and CORS headers
		duration := time.Since(start)
		log.Printf(
			"Completed %s %s in %v - Status: %d\n"+
				"CORS Headers:\n"+
				"  Access-Control-Allow-Origin: %s\n"+
				"  Access-Control-Allow-Methods: %s\n"+
				"  Access-Control-Allow-Headers: %s\n"+
				"  Access-Control-Allow-Credentials: %s",
			r.Method,
			r.URL.Path,
			duration,
			lrw.statusCode,
			lrw.Header().Get("Access-Control-Allow-Origin"),
			lrw.Header().Get("Access-Control-Allow-Methods"),
			lrw.Header().Get("Access-Control-Allow-Headers"),
			lrw.Header().Get("Access-Control-Allow-Credentials"),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
