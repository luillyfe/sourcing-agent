package observability

import (
	"net/http"

	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

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
