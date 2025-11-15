# LinkedIn Headline Generator

A Go-based AI workflow that generates professional LinkedIn headlines from your professional background description.

## Overview

This application uses Anthropic's Claude AI to analyze your professional background and generate compelling, concise LinkedIn headlines (max 220 characters) that highlight your key skills and roles.

## Features

- 🤖 AI-powered headline generation using Claude 3.5 Sonnet
- 📝 Multi-line text input support
- ✅ Character count validation (LinkedIn's 220 character limit)
- 🔒 Secure API key management via environment variables

## Prerequisites

- Go 1.21 or higher
- Anthropic API key ([Get one here](https://console.anthropic.com/))

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

4. Edit `.env` and add your Anthropic API key:
```
ANTHROPIC_API_KEY=your_actual_api_key_here
```

## Usage

Run the application:
```bash
go run main.go
```

Or build and run:
```bash
go build -o linkedin-headline
./linkedin-headline
```

### Example

```
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

## How It Works

1. **Input**: You provide a description of your professional background, skills, and experience
2. **Processing**: The application sends your input to Anthropic's Claude API with a specialized prompt
3. **Generation**: Claude analyzes your background and generates a professional LinkedIn headline
4. **Output**: You receive a polished headline that's ready to use on LinkedIn

## Project Structure

```
sourcing-agent/
├── main.go           # Main application code
├── go.mod            # Go module dependencies
├── go.sum            # Dependency checksums
├── .env.example      # Environment variable template
├── .env              # Your actual API keys (gitignored)
├── .gitignore        # Git ignore rules
├── LICENSE           # License file
└── README.md         # This file
```

## API Information

This application uses the Anthropic Messages API:
- **Model**: claude-3-5-sonnet-20241022
- **Max Tokens**: 1024
- **API Version**: 2023-06-01

## Error Handling

The application handles several error cases:
- Missing API key
- Empty input
- API request failures
- Invalid responses

## Security

- API keys are stored in `.env` files (excluded from git)
- Never commit your `.env` file
- Keep your API key secure and don't share it

## License

See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues or questions, please open an issue on the GitHub repository.
