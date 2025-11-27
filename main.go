package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/luillyfe/sourcing-agent/pkg/agent"
	"github.com/luillyfe/sourcing-agent/pkg/anthropic"
	"github.com/luillyfe/sourcing-agent/pkg/github"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Get API keys from environment
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable is not set")
		fmt.Println("Please create a .env file with your API key or set it as an environment variable")
		os.Exit(1)
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Println("Error: GITHUB_TOKEN environment variable is not set")
		fmt.Println("Please create a .env file with your GitHub token or set it as an environment variable")
		os.Exit(1)
	}

	// Check for command line arguments
	if len(os.Args) < 2 {
		fmt.Println("=== GitHub Developer Sourcing Agent ===")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run main.go \"<your query>\"")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run main.go \"Find Go developers in Lima\"")
		fmt.Println("  go run main.go \"Looking for Python engineers in Peru\"")
		fmt.Println("  go run main.go \"Need React developers with TypeScript experience\"")
		fmt.Println()
		os.Exit(0)
	}

	// Get query from command line
	query := strings.Join(os.Args[1:], " ")

	fmt.Println("=== GitHub Developer Sourcing Agent ===")
	fmt.Printf("Query: %s\n\n", query)
	fmt.Println("Searching...")
	fmt.Println()

	// Initialize clients
	githubClient := github.NewClient(githubToken)
	anthropicClient := anthropic.NewClient(anthropicKey)

	// Run the sourcing agent
	result, err := agent.Run(anthropicClient, githubClient, query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display result
	fmt.Println(result)
}
