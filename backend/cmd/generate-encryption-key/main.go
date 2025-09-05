package main

import (
	"fmt"
	"upm-backend/internal/utils"
)

func main() {
	key, err := utils.GenerateRandomKey()
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		return
	}

	fmt.Printf("Generated 32-byte encryption key:\n")
	fmt.Printf("ENCRYPTION_KEY=%s\n", key)
	fmt.Printf("\nAdd this to your .env file for production use.\n")
}
