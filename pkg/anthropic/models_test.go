package anthropic

import (
	"testing"
)

func TestContentBlock(t *testing.T) {
	t.Run("TextBlock", func(t *testing.T) {
		block := ContentBlock{
			Type: "text",
			Text: "Hello, world!",
		}
		if block.Type != "text" {
			t.Errorf("Expected type 'text', got '%s'", block.Type)
		}
		if block.Text != "Hello, world!" {
			t.Errorf("Expected text 'Hello, world!', got '%s'", block.Text)
		}
	})

	t.Run("ToolUseBlock", func(t *testing.T) {
		block := ContentBlock{
			Type: "tool_use",
			Name: "search_github_developers",
		}
		if block.Type != "tool_use" {
			t.Errorf("Expected type 'tool_use', got '%s'", block.Type)
		}
		if block.Name != "search_github_developers" {
			t.Errorf("Expected name 'search_github_developers', got '%s'", block.Name)
		}
	})

	t.Run("ToolResultBlock", func(t *testing.T) {
		block := ContentBlock{
			Type:      "tool_result",
			ToolUseID: "toolu_123",
		}
		if block.Type != "tool_result" {
			t.Errorf("Expected type 'tool_result', got '%s'", block.Type)
		}
		if block.ToolUseID != "toolu_123" {
			t.Errorf("Expected tool_use_id 'toolu_123', got '%s'", block.ToolUseID)
		}
	})
}
