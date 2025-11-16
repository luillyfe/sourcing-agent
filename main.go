package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	modelName       = "claude-3-7-sonnet-20250219"
	maxTokens       = 1024
)

// AnthropicRequest represents the request payload for Anthropic API
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the response from Anthropic API
type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// generateLinkedInHeadline sends a request to the Anthropic API to generate a LinkedIn headline
func generateLinkedInHeadline(apiKey, userText string) (string, error) {
	// Create the prompt
	prompt := fmt.Sprintf(`Based on the following professional background, create a concise and compelling LinkedIn headline (max 220 characters). The headline should be professional, highlight key skills or roles, and be attention-grabbing.

Professional background:
%s

Respond with ONLY the headline text, nothing else.`, userText)

	// Create the request payload
	requestBody := AnthropicRequest{
		Model:     modelName,
		MaxTokens: maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Marshal the request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var apiResponse AnthropicResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the headline from the response
	if len(apiResponse.Content) > 0 {
		headline := strings.TrimSpace(apiResponse.Content[0].Text)
		return headline, nil
	}

	return "", fmt.Errorf("no content in API response")
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable is not set")
		fmt.Println("Please create a .env file with your API key or set it as an environment variable")
		os.Exit(1)
	}

	fmt.Println("=== LinkedIn Headline Generator ===")
	fmt.Println("Enter your professional background (press Enter twice when done):")
	fmt.Println()

	// Read multi-line input from user
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	emptyLineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			emptyLineCount++
			if emptyLineCount >= 2 {
				break
			}
		} else {
			emptyLineCount = 0
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	userText := strings.TrimSpace(strings.Join(lines, "\n"))
	if userText == "" {
		fmt.Println("Error: No input provided")
		os.Exit(1)
	}

	fmt.Println("\nGenerating your LinkedIn headline...")
	fmt.Println()

	// Generate the LinkedIn headline
	headline, err := generateLinkedInHeadline(apiKey, userText)
	if err != nil {
		fmt.Printf("Error generating headline: %v\n", err)
		os.Exit(1)
	}

	// Display the result
	fmt.Println("✓ Generated LinkedIn Headline:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(headline)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\nCharacter count: %d/220\n", len(headline))
}
