# AI-Powered Sourcing Agent

A Go-based AI workflow application with two powerful features:
1. **LinkedIn Headline Generator** - Create compelling LinkedIn headlines from your professional background
2. **Developer Search** - Find and rank GitHub developers based on skills and experience

## Overview

This application uses Anthropic's Claude AI with function calling capabilities to power intelligent workflows for talent sourcing and professional branding.

## Features

### LinkedIn Headline Generator
- 🤖 AI-powered headline generation using Claude 3.7 Sonnet
- 📝 Multi-line text input support
- ✅ Character count validation (LinkedIn's 220 character limit)
- 🔒 Secure API key management via environment variables

### Developer Search (GitHub)
- 🔍 Natural language search queries (e.g., "Find Go developers in Lima with microservices experience")
- 🤖 AI-powered query interpretation using Claude's function calling
- 📊 Automated repository analysis for skill verification
- 🏆 Intelligent ranking based on project relevance
- 🔧 GitHub API integration with rate limit handling

## Prerequisites

- Go 1.21 or higher
- Anthropic API key ([Get one here](https://console.anthropic.com/))
- GitHub Personal Access Token (optional, but recommended for higher rate limits - [Get one here](https://github.com/settings/tokens))

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
```
ANTHROPIC_API_KEY=your_actual_api_key_here
GITHUB_TOKEN=your_github_token_here  # Optional but recommended
```

## Usage

Run the application:
```bash
go run main.go developer_search.go
```

Or build and run:
```bash
go build -o sourcing-agent
./sourcing-agent
```

You'll be presented with a menu to choose between workflows:
```
╔════════════════════════════════════════════════════╗
║         AI-Powered Sourcing Agent                  ║
╚════════════════════════════════════════════════════╝

Select a workflow:
  1. LinkedIn Headline Generator
  2. Developer Search (GitHub)

Enter your choice (1 or 2):
```

### Example 1: LinkedIn Headline Generator

```
Enter your choice (1 or 2): 1

=== LinkedIn Headline Generator ===
Enter your professional background (press Enter twice when done):

I'm a software engineer with 5 years of experience in cloud computing
and distributed systems. I specialize in Go, Kubernetes, and AWS.
I've led teams to build scalable microservices architectures.


Generating your LinkedIn headline...

✓ Generated LinkedIn Headline:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Senior Software Engineer | Cloud Architecture & Distributed Systems Expert | Go, Kubernetes & AWS Specialist | Building Scalable Solutions
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Character count: 143/220
```

### Example 2: Developer Search

```
Enter your choice (1 or 2): 2

=== Developer Search ===
Enter your search query (e.g., 'Find Go developers in Lima with microservices experience'): Find Go developers in Lima with microservices experience

🔍 Analyzing your search query...

📞 Calling tool: search_github_developers
✓ Tool executed successfully

📞 Calling tool: get_user_repos
✓ Tool executed successfully

📞 Calling tool: get_user_repos
✓ Tool executed successfully

============================================================
🎯 RESULTS
============================================================

Based on my analysis, here are the top Go developers in Lima with microservices experience:

1. **developer1** (@developer1)
   - Location: Lima, Peru
   - Relevant Projects:
     * microservices-go: A complete microservices architecture built with Go
     * k8s-deployment: Kubernetes deployment configurations
   - Profile: https://github.com/developer1

2. **developer2** (@developer2)
   - Location: Lima, Peru
   - Relevant Projects:
     * go-api-gateway: API gateway for microservices
     * docker-compose-services: Multi-container microservices setup
   - Profile: https://github.com/developer2

============================================================
```

## How It Works

### LinkedIn Headline Generator
1. **Input**: You provide a description of your professional background, skills, and experience
2. **Processing**: The application sends your input to Anthropic's Claude API with a specialized prompt
3. **Generation**: Claude analyzes your background and generates a professional LinkedIn headline
4. **Output**: You receive a polished headline that's ready to use on LinkedIn

### Developer Search Workflow
1. **Query Input**: You provide a natural language search query
2. **AI Analysis**: Claude interprets the query and identifies:
   - Programming language (e.g., "Go")
   - Location (e.g., "Lima")
   - Required skills (e.g., "microservices")
3. **Function Calling**: Claude autonomously calls these functions:
   - `search_github_developers(language, location)` - Finds matching developers
   - `get_user_repos(username)` - Analyzes each developer's repositories
4. **Skill Verification**: The AI examines repository descriptions, topics, and metadata for skill keywords
5. **Ranking**: Candidates are scored based on:
   - Relevance of projects to requested skills
   - Repository popularity (stars, forks)
   - Number of matching projects
6. **Results**: You receive a ranked list of the best candidates with links to their profiles

## Project Structure

```
sourcing-agent/
├── main.go              # Main application with menu and LinkedIn workflow
├── developer_search.go  # Developer search workflow with AI function calling
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── .env.example         # Environment variable template
├── .env                 # Your actual API keys (gitignored)
├── .gitignore           # Git ignore rules
├── LICENSE              # License file
└── README.md            # This file
```

## API Information

This application integrates with two APIs:

### Anthropic Claude API
- **Model**: claude-3-7-sonnet-20250219
- **Max Tokens**: 1024 (LinkedIn), 4096 (Developer Search)
- **API Version**: 2023-06-01
- **Features Used**: Messages API, Function Calling (Tool Use)

### GitHub API
- **Version**: v3 (REST API)
- **Endpoints Used**:
  - `/search/users` - Search for developers
  - `/users/{username}/repos` - Get user repositories
- **Authentication**: Optional Personal Access Token (recommended for higher rate limits)

## Error Handling

The application handles various error cases:
- Missing API keys (Anthropic and GitHub)
- Empty or invalid input
- API request failures and rate limiting
- Invalid JSON responses
- Network timeouts
- GitHub API authentication errors

## Security

- API keys are stored in `.env` files (excluded from git via `.gitignore`)
- Never commit your `.env` file
- Keep your API keys secure and don't share them
- GitHub token only requires `public_repo` scope (read-only access)
- All API requests use HTTPS

## Rate Limits

### Without GitHub Token
- 60 requests per hour (GitHub API)

### With GitHub Token
- 5,000 requests per hour (GitHub API)

**Recommendation**: Set up a GitHub token to avoid hitting rate limits when searching for multiple developers.

## License

See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues or questions, please open an issue on the GitHub repository.
