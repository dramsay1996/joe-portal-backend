package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Auth middleware checking path: %s\n", r.URL.Path)

		// Skip auth for login and public routes
		if r.URL.Path == "/api/login" || strings.HasPrefix(r.URL.Path, "/public/") {
			fmt.Println("Skipping auth for public route")
			next.ServeHTTP(w, r)
			return
		}

		// Get token from cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			fmt.Println("No token cookie found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		fmt.Printf("Found token cookie: %s\n", cookie.Value)

		// Parse and validate token
		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil || !token.Valid {
			fmt.Printf("Token validation failed: %v\n", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			fmt.Println("Invalid token claims")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "userId", claims.UserID)
		r = r.WithContext(ctx)

		fmt.Printf("Token validated successfully for user: %s\n", claims.UserID)
		next.ServeHTTP(w, r)
	})
}
