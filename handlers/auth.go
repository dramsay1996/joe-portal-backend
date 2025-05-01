package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Password string `json:"password"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the hashed password from environment variable
	hashedPassword := os.Getenv("APP_PASSWORD_HASH")
	if hashedPassword == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get JWT key from environment
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		fmt.Println("Error: JWT_SECRET_KEY is not set")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		UserID: "user", // For now, we'll use a static user ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the token in an HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   true, // Set to true in production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

// Logout handles user logout by clearing the JWT cookie
func Logout(w http.ResponseWriter, r *http.Request) {
	// Create an expired cookie to clear the JWT
	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Immediately expire the cookie
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	// Set the expired cookie
	http.SetCookie(w, &cookie)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Logged out successfully",
	})
}
