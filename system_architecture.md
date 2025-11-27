```text
        ┌─────────────────┐
        │     Client      │
        │     (API)       │
        └───────┬─────────┘
                │ HTTP
                ▼
┌───────────────────────────────────┐
│        Agent Service (Go)         │
│           Single Process          │
│                                   │
│  ┌───────────────────────────┐    │
│  │ LLM Intent Analysis       │────│──▶ Vertex AI
│  └───────────────────────────┘    │
│                                   │   (External)
│  ┌───────────────────────────┐    │
│  │ Tool Call Construction    │    │
│  └───────────────────────────┘    │
│                                   │
│  ┌───────────────────────────┐    │
│  │ GitHub Tool Executor      │────│──▶ GitHub Search API
│  └───────────────────────────┘    │
│                                   │
│  ┌───────────────────────────┐    │
│  │ LLM Final Synthesis       │────│──▶ Vertex AI
│  └───────────────────────────┘    │
│                                   │
│  Logging: fmt.Println()           │
└───────────────┬───────────────────┘
                │ HTTP
                ▼
         Formatted Response
```
