package main

import (
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	fmt.Printf("Analytics starting...\n")
}
