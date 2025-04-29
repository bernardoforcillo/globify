package globify

import (
	"fmt"
	"log"
	"os"

	"github.com/bernardoforcillo/globify/internal/app"
	"github.com/joho/godotenv"
)

// Main is the exported entry point for the application
func Main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Create and run app
	globify, err := app.NewApp()
	if err != nil {
		log.Fatalf("Error initializing application: %v", err)
	}

	if err := globify.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	Main()
}