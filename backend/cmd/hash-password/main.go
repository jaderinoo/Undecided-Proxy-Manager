package main

import (
	"fmt"
	"os"

	"upm-backend/internal/auth"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run cmd/hash-password/main.go <password>")
		os.Exit(1)
	}

	password := os.Args[1]
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Original password: %s\n", password)
	fmt.Printf("Hashed password: %s\n", hashedPassword)
	fmt.Printf("\nAdd this to your .env file:\n")
	fmt.Printf("ADMIN_PASSWORD=%s\n", hashedPassword)
}
