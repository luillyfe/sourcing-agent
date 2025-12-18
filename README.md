# GitHub Developer Sourcing Agent - Stage 2

A Go-based AI agent that discovers and qualifies developers using **Prompt Chaining** with **Gemini 3 Pro** on **Google Cloud Vertex AI**.

## Overview

This is a **Stage 2: Multi-Prompt Sourcing Agent**, representing a significant evolution from the single-shot approach. It implements the **Prompt Chaining** pattern to break down the complex task of sourcing into a deterministic pipeline of specialized steps: requirements analysis, strategy generation, candidate enrichment, and final ranking.

### What Makes This "Stage 2"?

- **Prompt Chaining Architecture**: A sequential pipeline of 4 specialized LLM prompts.
- **Dynamic Search Strategies**: Intelliqently generates primary and fallback search strategies based on requirements.
- **Candidate Enrichment**: deeply analyzes developer repositories to assess technical fit beyond just bio matching.
- **Weighted Scoring**: Programmatically calculates match scores based on skills, repo relevance, experience, and profile quality.
- **Observability**: Built-in tracking for Token Usage, API Call Counts, and Memory execution stats.

## Features

- 🧠 **Smart Requirements Analysis**: Parses natural language into structured technical requirements.
- 🎯 **Strategic Searching**: Generates optimal GitHub search queries + fallback options if results are scarse.
- 🔬 **Deep Repository Analysis**: Fetches and analyzes user repositories to verify claimed skills.
- 📊 **Programmatic Ranking**: Scores candidates on a weighted scale (Skills 40%, Repos 30%, Experience 20%, Quality 10%).
- 🛡️ **Rate Limit Aware**: Optimized to work within GitHub's API constraints.
- 👁️ **Full Observability**: Reports execution time, token usage, and API call counts for every run.

## System Architecture

<img src="Sourcing%20Agent%20Architecture%20Stage%202.png" width="1200" alt="Stage 2 System Architecture">

## Workflow: The 4-Step Pipeline

The agent follows a strict 4-step linear workflow:

<img src="Sourcing%20Agent%20Stage%202.png" width="250" alt="Stage 2 Workflow">

1.  **Requirements Analyzer**: "Find Go devs in Lima" -> `{Skills: ["Go"], Location: "Lima"}`
2.  **Strategy Generator**: Creates a primary search (strict) and fallback searches (broader).
3.  **Candidate Enricher**: Executes GitHub searches, then fetches repositories for each candidate to analyze code relevance.
4.  **Ranker**: Calculates final scores and formats the top candidates.

## Prerequisites

- **Go 1.21 or higher**
- **Google Cloud Project** with Vertex AI API enabled
- **GitHub Personal Access Token**
    - Required scope: `read:user`

## Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/luillyfe/sourcing-agent.git
    cd sourcing-agent
    ```

2.  Install dependencies:
    ```bash
    go mod download
    ```

3.  Set up your environment variables:
    ```bash
    cp .env.example .env
    ```

4.  Edit `.env` and add your configuration:
    ```env
    VERTEX_PROJECT_ID=your_google_cloud_project_id
    VERTEX_REGION=us-central1
    GITHUB_TOKEN=your_actual_github_token_here
    ```

## Usage

### Basic Usage

Run the application with a natural language query:

```bash
go run main.go "Find Go developers in Lima"
```

### Example Output

The agent provides real-time progress updates and a detailed JSON final report:

```text
=== GitHub Developer Sourcing Agent ===
Query: Find Go developers in Lima

Searching...

Step 1: Analyzing requirements...
Requirements analysis took 1.2s
  Usage: 154 input, 45 output tokens

Step 2: Generating search strategy...
Strategy generation took 1.5s
  Usage: 450 input, 320 output tokens

Step 3: Finding and enriching candidates...
Found 12 candidates, analyzed 12
Candidate search and enrichment took 4.5s

Step 4: Ranking and presenting...
Ranking took 2.1s
  Usage: 2100 input, 800 output tokens

--------------------------------------------------
Total Token Usage: 2704 input + 1165 output = 3869 total
--------------------------------------------------

{
  "top_candidates": [
    {
      "rank": 1,
      "username": "fermin_tech",
      "final_match_score": 0.92,
      "match_breakdown": {
        "required_skills_score": 1.0,
        "repository_relevance_score": 0.9,
        "experience_score": 0.8,
        "profile_quality_score": 0.9
      },
      "key_qualifications": ["Go", "Microservices", "Cloud"],
      "match_reasoning": "Strong match with active Go repositories..."
    }
  ],
  "summary": {
    "total_candidates_found": 12,
    "candidates_presented": 10,
    "average_match_score": 0.78,
    "search_quality": "excellent"
  }
}

Total execution time: 9.35 seconds
Total LLM calls: 3
Total GitHub API calls: 14
Memory usage: Alloc = 25 MiB...
```

## Project Structure

```
sourcing-agent/
├── main.go               # Entry point, client initialization, observability setup
├── pkg/
│   ├── agent/            # Core Agent Logic
│   │   ├── agent.go      # Pipeline orchestration (RunStage2)
│   │   ├── prompts.go    # System prompts for each step
│   │   └── types.go      # Data structures (Requirements, Strategy, etc.)
│   ├── github/           # GitHub API Client
│   ├── llm/              # LLM Interface definition
│   ├── observability/    # Metrics (CountingTransport, CountingLLMClient)
│   └── vertexai/         # Vertex AI specific implementation
└── docs/                 # Design documents (Stage 1, Stage 2)
```

## Configuration

### Environment Variables

| Variable | Required | Description |
| :--- | :--- | :--- |
| `VERTEX_PROJECT_ID` | Yes | Your Google Cloud Project ID |
| `VERTEX_REGION` | Yes | Your Google Cloud Region (e.g., us-central1) |
| `GITHUB_TOKEN` | Yes | Your GitHub Personal Access Token |

## License

MIT License - see the [LICENSE](LICENSE) file for details.
