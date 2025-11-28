package llm

// Client defines the interface for interacting with an LLM
type Client interface {
	CallAPI(messages []Message, tools []Tool) (*Response, error)
}
