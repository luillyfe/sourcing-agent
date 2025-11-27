# GitHub Developer Sourcing Agent - Stage 1

A Go-based AI agent that searches GitHub for developers matching hiring requirements using Claude Sonnet 4.5 / Gemini 3 Pro and the Augmented LLM pattern.

## Overview

This is a **Stage 1: Single-Shot Sourcing Agent** - the simplest possible implementation of an AI-powered developer sourcing system. It uses the **Augmented LLM pattern** where Claude has access to tools and can search GitHub in a single conversation turn.

### What Makes This Stage 1?

- **One Query, One Response**: No loops, no iteration, no multi-step orchestration
- **Single Tool**: One tool (`search_github_developers`) that does all the work
- **Augmented LLM**: Claude + tool access in a single invocation
- **Foundation**: Simple pattern that can evolve into more complex architectures

## Features

- 🤖 AI-powered developer search using Claude Sonnet 4.5
- 🔍 Natural language query processing
- 🐙 GitHub API integration for comprehensive developer profiles
- 📊 Rich candidate information (bio, repos, followers, location)
- ⚡ Single-shot execution (< 30 seconds)
- 🎯 Keyword filtering in developer bios
- 🔒 Secure API key management

## Prerequisites

- **Go 1.21 or higher**
- **Anthropic API key** - [Get one here](https://console.anthropic.com/)
- **GitHub Personal Access Token** - [Get one here](https://github.com/settings/tokens)
  - Required scope: `read:user`

## Installation

1. Clone the repository:
```bash
git clone https://github.com/luillyfe/sourcing-agent.git
cd sourcing-agent
```

2. Install dependencies:
```bash
go mod download
```

3. Set up your environment variables:
```bash
cp .env.example .env
```

4. Edit `.env` and add your API keys:
```env
ANTHROPIC_API_KEY=your_actual_anthropic_api_key_here
GITHUB_TOKEN=your_actual_github_token_here
```

## Usage

### Basic Usage

Run the application with a natural language query:

```bash
go run main.go "Find Go developers in Lima"
```

### Build and Run

```bash
go build -o sourcing-agent
./sourcing-agent "Looking for Python engineers in Peru"
```

### Example Queries

```bash
# Search by language and location
go run main.go "Find Go developers in Lima"

# Search with keywords
go run main.go "Looking for Python engineers with machine learning experience"

# Search by language only
go run main.go "Need React developers with TypeScript experience"

# Search in broader regions
go run main.go "Find JavaScript developers in Latin America"
```

## Example Output

```
=== GitHub Developer Sourcing Agent ===
Query: Find Go developers in Lima with microservices experience

Searching...

I found 8 Go developers in Lima with microservices experience:

1. fermin_tech (@fermin_tech)
   - Lima, Peru
   - Cloud Engineer | Go microservices | MongoDB
   - 25 public repos, 89 followers
   - https://github.com/fermin_tech

2. carlos_backend (@carlos_backend)
   - Lima, Peru
   - Backend Developer | Go | Kubernetes | Microservices
   - 32 public repos, 156 followers
   - https://github.com/carlos_backend

[... 6 more candidates]
```

## How It Works

### Architecture

```
User Query → Claude (LLM) + Tool → GitHub API → Result
```

### Process Flow

1. **User provides natural language query**
   - Example: "Find Go developers in Lima"

2. **Claude parses the query**
   - Extracts: language="go", location="lima"

3. **Claude calls the search_github_developers tool**
   - Passes extracted parameters

4. **Tool searches GitHub**
   - Builds search query
   - Calls GitHub Search API
   - Enriches results with user details
   - Filters by keywords if specified

5. **Claude formats and presents results**
   - Clear, readable candidate profiles
   - Ready to use

### Tool: search_github_developers

The system has ONE tool that does all the work:

**Parameters:**
- `language` (required) - Programming language (e.g., 'python', 'go', 'javascript')
- `location` (optional) - Geographic location (e.g., 'lima', 'peru', 'san francisco')
- `keywords` (optional) - Keywords in user bio (e.g., 'microservices', 'mongodb')
- `min_repos` (default: 5) - Minimum public repositories
- `max_results` (default: 10) - Maximum candidates to return

**Returns:**
```json
{
  "candidates": [
    {
      "username": "dev_user",
      "name": "John Doe",
      "location": "Lima, Peru",
      "bio": "Go developer building microservices",
      "public_repos": 45,
      "followers": 120,
      "github_url": "https://github.com/dev_user",
      "avatar_url": "https://..."
    }
  ],
  "total_found": 10,
  "search_criteria": {
    "language": "go",
    "location": "lima",
    "keywords": "microservices"
  }
}
```

## Project Structure

```
sourcing-agent/
├── main.go           # Main implementation
│   ├── GitHub API integration
│   ├── Anthropic Claude API integration
│   ├── Tool definition and execution
│   └── Sourcing agent logic
├── main_test.go      # Unit tests
├── go.mod            # Go module dependencies
├── go.sum            # Dependency checksums
├── .env.example      # Environment variable template
├── .env              # Your actual API keys (gitignored)
├── .gitignore        # Git ignore rules
├── LICENSE           # MIT License
└── README.md         # This file
```

## API Details

### Anthropic Claude API

- **Model**: claude-sonnet-4-20250514
- **Max Tokens**: 4096
- **API Version**: 2023-06-01
- **Pattern**: Augmented LLM with tool use

### GitHub API

- **Endpoints Used**:
  - `GET /search/users` - Find developers
  - `GET /users/{username}` - Get user details
- **Rate Limits**:
  - Authenticated: 5,000 requests/hour
  - Search API: 30 requests/minute
- **Authentication**: Personal Access Token

## Testing

### Run Tests

```bash
go test -v
```

### Run Benchmarks

```bash
go test -bench=.
```

### Test Coverage

```bash
go test -cover
```

The test suite includes:
- Unit tests for data structures
- Tool definition validation
- JSON marshaling/unmarshaling tests
- Input validation tests
- Benchmark tests

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | Yes | Your Anthropic API key |
| `GITHUB_TOKEN` | Yes | Your GitHub Personal Access Token |

### Default Settings

- **Max Results**: 10 candidates
- **Min Repos**: 5 public repositories
- **Timeout**: 30 seconds for GitHub API, 60 seconds for Claude
- **Tool Calls**: 1-2 per query (typically just 1)

## Error Handling

The agent handles several error cases gracefully:

- **Missing API Keys**: Clear error message with setup instructions
- **Rate Limit Exceeded**: Returns error with retry suggestion
- **Invalid Location**: Returns available results
- **No Results Found**: Informs user and suggests broader criteria
- **API Failures**: Clear error messages with debugging information

## Limitations (By Design)

Stage 1 does NOT include:

- ❌ No iteration or refinement
- ❌ No deep repository analysis
- ❌ No cross-platform search (only GitHub)
- ❌ No evaluation/scoring logic
- ❌ No follow-up questions
- ❌ No conversation history

These are features for future stages (Stage 2+).

## Future Stages

This Stage 1 implementation is the foundation for more advanced patterns:

- **Stage 2**: Multi-prompt with prompt chaining (sequential prompts)
- **Stage 3**: Reflective agent with evaluation and refinement
- **Stage 4**: Integrated system with multiple sources (LinkedIn, internal ATS)
- **Stage 5**: Autonomous agent with goal-driven behavior

### Why Start with Stage 1?

> "Always try simpler patterns first. Most problems don't need full autonomy."
>
> — Anthropic's Building Effective Agents Guide

Stage 1 proves the concept works before investing in complex architectures.

## Building Effective Agents - Pattern Mapping

This implementation follows Anthropic's recommended progression:

1. **✓ Augmented LLM (Stage 1)** ← You are here
   - LLM + tool access
   - Single conversation turn
   - Simplest possible implementation

2. **Prompt Chaining (Stage 2)**
   - Multiple sequential prompts
   - Each step builds on previous

3. **Routing (Stage 3)**
   - Classify query → route to specialist

4. **Orchestrator-Workers (Stage 4)**
   - Coordinator + specialized agents

5. **Full Autonomy (Stage 5)**
   - Goal-driven behavior
   - Self-correction and adaptation

## Security

- API keys stored in `.env` files (excluded from git)
- Never commit `.env` file or expose API keys
- Input validation for all user inputs
- Safe HTTP client with timeouts
- No execution of arbitrary code

## Performance

- **Execution Time**: < 30 seconds typical
- **API Calls**: 1 Claude call + N+1 GitHub calls (N = number of candidates)
- **Rate Limits**: Well within GitHub's limits for single searches
- **Memory**: Minimal (<50MB typical)

## Troubleshooting

### "ANTHROPIC_API_KEY environment variable is not set"

Create a `.env` file with your API key:
```env
ANTHROPIC_API_KEY=sk-ant-...
```

### "GITHUB_TOKEN environment variable is not set"

Add your GitHub token to `.env`:
```env
GITHUB_TOKEN=ghp_...
```

### "GitHub API request failed with status 403"

Check your GitHub token:
- Ensure it has `read:user` scope
- Verify it hasn't expired
- Check rate limits: https://api.github.com/rate_limit

### "No results found"

Try broader search criteria:
- Remove location filter
- Use more general keywords
- Lower `min_repos` requirement

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Guidelines

1. Follow Go best practices
2. Add tests for new features
3. Update documentation
4. Keep it simple (Stage 1 philosophy)

## License

MIT License - see the [LICENSE](LICENSE) file for details.

## Resources

- [Anthropic API Documentation](https://docs.anthropic.com/)
- [GitHub API Documentation](https://docs.github.com/rest)
- [Building Effective Agents (Anthropic)](https://docs.anthropic.com/en/docs/build-with-claude/develop-tests)
- [Go Documentation](https://golang.org/doc/)

## Support

For issues or questions:
- Open an issue on the GitHub repository
- Check existing issues for solutions
- Review the troubleshooting section

## Acknowledgments

This implementation follows the **Augmented LLM pattern** as described in Anthropic's "Building Effective Agents" guide, emphasizing simplicity and effectiveness over complexity.

---

**Built with Claude Sonnet 4.5 and Go 1.21+**

**Pattern**: Augmented LLM (Stage 1 of 5)

**Version**: 1.0.0
