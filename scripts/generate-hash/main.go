package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
)

func main() {
	// Parse command line arguments
	password := flag.String("password", "", "Password to hash")
	flag.Parse()

	if *password == "" {
		fmt.Println("Error: Password is required")
		fmt.Println("Usage: go run main.go -password <your-password>")
		os.Exit(1)
	}

	// Generate hash
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		os.Exit(1)
	}

	// Print the hash
	fmt.Printf("Hashed password: %s\n", string(hash))
	fmt.Println("\nAdd this to your .env file:")
	fmt.Printf("APP_PASSWORD_HASH=\"%s\"\n", string(hash))
}
