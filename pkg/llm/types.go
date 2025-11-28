package llm

// Request represents the generic request payload for LLM API
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

// Message represents a generic message in the conversation
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentBlock
}

// ContentBlock represents a generic content block
type ContentBlock struct {
	Type             string      `json:"type"`
	Text             string      `json:"text,omitempty"`
	ID               string      `json:"id,omitempty"`
	Name             string      `json:"name,omitempty"`
	Input            interface{} `json:"input,omitempty"`
	ToolUseID        string      `json:"tool_use_id,omitempty"`
	Content          string      `json:"content,omitempty"`
	ThoughtSignature string      `json:"thought_signature,omitempty"`
}

// Tool represents a generic tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines the input parameters for a tool
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property defines a property in the tool input schema
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     int    `json:"default,omitempty"`
}

// Response represents the generic response from LLM API
type Response struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	Model      string         `json:"model"`
	StopReason string         `json:"stop_reason"`
	Usage      Usage          `json:"usage"`
}

// Usage represents token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
