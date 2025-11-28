# Stage 1: Single-Shot Sourcing Agent
## Design Specification (Go Implementation)

## 1. Overview
* **Pattern:** Augmented LLM
* **Purpose:** Search for developers on GitHub in a single invocation
* **Complexity:** Minimal - one query, one response

## 2. System Architecture
`User Query → LLM + Tool(s) → Result`

That's it. No loops, no iteration, no multi-step orchestration.

## 3. Core Components

### 3.1 LLM
* **Model:** Claude Sonnet 4.5 / Gemini 3 Pro
* **Job:** Parse query → call tool → format results

### 3.2 Tool
* **One tool:** `search_github_developers`
* **No orchestration logic**

### 3.3 Flow
1.  User asks: "Find Go developers in Lima"
2.  LLM calls: `search_github_developers(...)`
3.  LLM formats results
4.  Done

## 4. GitHub API Integration

### 4.1 Authentication
```go
githubToken := os.Getenv("GITHUB_TOKEN")
```

### 4.2 Endpoints Used
* `GET /search/users` - Find developers by criteria
* `GET /users/{username}` - Get detailed profile
* `GET /users/{username}/repos` - Get user repositories

### 4.3 Rate Limits
* *Authenticated*: 5,000 requests/hour
* *Search API*: 30 requests/minute
* *Strategy*: Cache results, batch requests

### 5. Tool Definition
#### 5.1 search_github_developers
*Purpose*: Search GitHub and return ready-to-present candidate profiles
*Parameters*
| Parameter | Description |
| --- | --- |
| language | string (required) - e.g., 'python', 'go', 'javascript' |
| location | string (optional) - e.g., 'lima', 'peru', 'san francisco' |
| keywords | string (optional) - e.g., 'microservices', 'mongodb', 'react' |
| min_repos | integer (default: 5) - minimum public repositories |
| max_results | integer (default: 10) - maximum candidates to return |

*Returns*
| Field | Description |
| --- | --- |
| candidates | Array of candidate profiles |
| total_found | Number of candidates found |
| search_criteria | Object containing search parameters used |

### 6. LLM Prompt Design
#### 6.1 System Prompt
You are a developer sourcing assistant. Your job is to search GitHub for developers 
matching hiring requirements. You have ONE tool: `search_github_developers`.

*Process*:
1. Extract: programming language, location, and relevant keywords from the query
2. Call `search_github_developers` with appropriate parameters
3. Present the results in a clear, readable format Keep it simple. One search, one response.

#### 6.2 User Query Examples
- "Find Go developers in Lima"
- "Looking for Python engineers in Peru"
- "Need React developers with TypeScript experience"

### Execution Flow
*User*: "Find Go developers in Lima with microservices experience" *↓ LLM parses*
- language: "go"
- location: "lima"
- keywords: "microservices" *↓ LLM calls*: search_github_developers(language="go", 
  location="lima", keywords="microservices", max_results=10) ↓ Tool returns: 
  10 candidate profiles *↓ LLM formats response ↓ Done*.

### Key Point: The LLM does NOT:
- Make multiple tool calls
- Analyze results and search again
- Call additional tools for enrichment
- Iterate or loop

### 8. Implementation
#### 8.1 Technology Stack
- *Language*: Go 1.21+
- *LLM API*: Anthropic Claude / Gemini 3 Pro (via REST API)
- *HTTP Client*: Standard library net/http
- *Dependencies*: None - standard library only

#### 8.2 Key Functions
- `searchGitHubDevelopers()` - Main GitHub API wrapper
- `executeTool()` - Tool execution dispatcher
- `runSourcingAgent()` - Main agent orchestrator

#### 8.3 Environment Setup
```bash
export GITHUB_TOKEN="your_github_token_here"
export ANTHROPIC_API_KEY="your_anthropic_api_key_here"
```

### 9. Success Criteria
- Single invocation: One query → one response
- Tool usage: LLM successfully calls search_github_developers
- Results: Returns 5-10 relevant candidates
- Speed: Completes in < 15 seconds
- Format: Clear, readable candidate list

### 10. Limitations (By Design)
- No iteration or refinement
- No deep repository analysis
- No cross-platform search
- No evaluation/scoring logic
- No follow-up questions

These are features for Stage 2+

### 11. Example Usage
Query 1: Find Python developers in Peru
User: "Find Python developers in Peru"
Response: I found 10 Python developers in Peru:

carlos_dev (@carlos_dev) - Lima, Peru - Python backend developer | Django enthusiast - 32 repos, 45 followers - https://www.google.com/search?q=https://github.com/carlos_dev
maria_code (@maria_code) - Arequipa, Peru - ML Engineer | Python | Data Science - 18 repos, 67 followers - https://www.google.com/search?q=https://github.com/maria_code [... 8 more]

Query 2: Go developers with microservices
User: "Looking for Go developers with microservices experience"
Response: I found 8 Go developers with microservices experience:

fermin_tech (@fermin_tech) - Lima, Peru - Cloud Engineer | Go microservices | MongoDB - 25 repos, 89 followers - https://www.google.com/search?q=https://github.com/fermin_tech [... 7 more]

### 12. What Makes This Stage 1?
*Augmented LLM Pattern*
- LLM has access to tools
- Single conversation turn
- No complex orchestration

*Simplicity*
- One tool
- Direct execution
- No iteration

*Foundation for Future Stages*
- Stage 2 will add prompt chaining
- Stage 3 will add reflection
- Stage 4 will add multiple sources

This is the real Stage 1. Simple, direct, functional.