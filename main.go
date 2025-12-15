package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/luillyfe/sourcing-agent/pkg/agent"
	"github.com/luillyfe/sourcing-agent/pkg/github"
	"github.com/luillyfe/sourcing-agent/pkg/llm"
	"github.com/luillyfe/sourcing-agent/pkg/vertexai"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Get API keys from environment
	projectID := os.Getenv("VERTEX_PROJECT_ID")
	if projectID == "" {
		fmt.Println("Error: VERTEX_PROJECT_ID environment variable is not set")
		fmt.Println("Please create a .env file with your Project ID or set it as an environment variable")
		os.Exit(1)
	}

	region := os.Getenv("VERTEX_REGION")
	if region == "" {
		fmt.Println("Error: VERTEX_REGION environment variable is not set")
		fmt.Println("Please create a .env file with your Region or set it as an environment variable")
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
	// 1. GitHub Client with Observability
	countingTransport := &CountingTransport{Transport: http.DefaultTransport}
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: countingTransport,
	}

	githubClient := github.NewClient(githubToken)
	githubClient.HTTPClient = httpClient

	ctx := context.Background()
	vertexClient, err := vertexai.NewClient(ctx, projectID, region)
	if err != nil {
		fmt.Printf("Error initializing Vertex AI client: %v\n", err)
		os.Exit(1)
	}
	defer vertexClient.Close()

	// 2. LLM Client with Observability
	countingLLMClient := &CountingLLMClient{Wrapped: vertexClient}

	// Run the sourcing agent
	startTime := time.Now()
	result, err := agent.RunStage2(countingLLMClient, githubClient, query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	duration := time.Since(startTime)

	// Display result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(resultJSON))
	fmt.Printf("\nTotal execution time: %.2f seconds\n", duration.Seconds())
	fmt.Printf("Total LLM calls: %d\n", countingLLMClient.Count)
	fmt.Printf("Total GitHub API calls: %d\n", countingTransport.Count)
}

// --- Observability Types ---

// CountingTransport tracks the number of HTTP requests
type CountingTransport struct {
	Transport http.RoundTripper
	Count     int
}

func (t *CountingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.Count++
	// Use default transport if nil
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

// CountingLLMClient tracks the number of LLM API calls
type CountingLLMClient struct {
	Wrapped llm.Client
	Count   int
}

func (c *CountingLLMClient) CallAPI(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	c.Count++
	return c.Wrapped.CallAPI(messages, tools)
}
