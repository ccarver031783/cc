# cc - Personal Dev Tool CLI

A simple, shortcut-based CLI tool for simplifying and automating common development workflows including git operations, PR management, and Terraform validation.

## Project Structure

```
cc/
├── cmd/
│   └── cc/
│       └── main.go              # Entry point
├── internal/
│   ├── git/                     # Git operations
│   │   └── git.go              # Branch, rebase, clean, status
│   ├── pr/                      # PR creation/management (GitHub CLI)
│   ├── setup/                   # Homebrew package management
│   │   └── setup.go            # Check, install, upgrade packages
│   ├── terraform/               # Terraform operations
│   │   └── terraform.go        # Format, scan, validate
│   ├── explain/                 # AI-powered code explanations
│   │   ├── tf_explain.go       # Terraform module analysis
│   │   ├── claude.go           # Claude API integration
│   │   ├── ollama.go           # Local Ollama integration
│   │   └── README.md           # Setup instructions
│   └── shell/                   # Shell execution utilities
│       └── shell.go            # Command execution helpers
├── examples/
│   └── terraform-templates/     # Terraform scaffolding templates
│       ├── aws.yaml
│       ├── azure.yaml
│       ├── gcp.yaml
│       └── README.md
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## Core Features

### 1. Setup & Package Management (`setup` command)

```bash
cc setup                         # Check and manage Homebrew packages
```

The setup command:
- **Checks Homebrew installation** - Installs Homebrew if not present
- **Detects required packages** - Checks if packages are installed (via Homebrew or manually)
- **Migrates manual installations** - For command-line tools, installs via Homebrew alongside manual versions and ensures Homebrew takes precedence via PATH
- **Upgrades outdated packages** - Identifies and offers to upgrade packages with available updates
- **Installs missing packages** - Offers to install packages not yet present

**Migration behavior:**
- **GUI Apps (Casks)**: Removes manual installation and reinstalls via Homebrew
- **Command-Line Tools (Formulas)**: Installs via Homebrew alongside manual installation, allowing Homebrew to "assume control" via PATH precedence

**Required packages checked:**
- Go (`go`)
- Sequel Ace (`sequel-ace`)
- UTM (`utm`)

### 2. Git Operations (`git` command)

```bash
cc git branch <name>         # Create new branch from clean main/master
cc git rebase <target-branch> # Rebase current branch onto specified branch
cc git clean                  # Clean working directory (stash changes, reset)
cc git status                 # Enhanced git status with branch info
```

### 3. PR Management (`pr` command)

```bash
cc pr create [--draft]        # Create PR from current branch using GitHub CLI
cc pr list                    # List open PRs (via gh CLI)
cc pr view <number>           # View PR details (via gh CLI)
```

**Note:** Commands must work whether user or AI/automation creates the PR.

### 4. Terraform Operations (`terraform` or `tf` command)

```bash
cc tf fmt                     # Format Terraform files (changed files only)
cc tf scan                    # Run security scan with tfsec or tflint (changed files only)
cc tf validate                # Validate Terraform config
cc tf pre-push                # Run fmt + scan + validate on changed files before push
cc tf init-dir <path>         # Scaffold a new Terraform directory
cc tf new <resource-name>     # Create multi-provider resource structure
```

### 5. AI-Powered Explanations (`explain` command)

```bash
cc explain tf [path]          # Explain Terraform modules using AI
cc explain tf . --local       # Use local Ollama instead of Claude API
```

The explain command analyzes Terraform modules and provides clear explanations including:
- Purpose and functionality
- Resources created
- Key variables and outputs
- Dependencies and use cases

**Dual AI Support:**
- **Claude API** (cloud) - High quality, requires API key
- **Ollama** (local) - Free, unlimited, works offline

See [internal/explain/README.md](internal/explain/README.md) for setup details.

### 6. Pre-Push Hook

```bash
cc hook install                # Install git pre-push hook
```

Hook automatically runs: `terraform fmt`, `scan` (tfsec/tflint), `validate` on changed files before allowing push.

Hook must work in both manual and automated (AI/CI) contexts.

## Technology Stack

- **CLI Framework:** `github.com/urfave/cli/v2` - Command-line interface structure
- **AI Integration:**
  - `github.com/anthropics/anthropic-sdk-go` - Claude API client
  - HTTP client for Ollama local LLM integration
- **Git Operations:** Shell commands via `os/exec`
- **GitHub API:** `github.com/cli/cli` (gh CLI) wrapper
- **Terraform:** Shell execution of terraform/tfsec/tflint binaries
- **Package Management:** Homebrew via shell commands

## Installation

### Prerequisites

- macOS (primary support) or Linux
- Go 1.21+ (for building from source)
- Homebrew (will be installed automatically if missing)

### Building from Source

```bash
git clone <your-repo>
cd cc
go build -o cc ./cmd/cc
./cc --help
```

### Optional Dependencies

**For AI explanations:**
- Claude API key (set `ANTHROPIC_API_KEY`)
- OR Ollama installed locally (`brew install ollama`)

**For Terraform operations:**
- `terraform` - Terraform CLI
- `tflint` or `tfsec` - Security scanning

## Quick Start

```bash
# Set up your environment
cc setup

# Check git status
cc git status

# Explain a Terraform module
export ANTHROPIC_API_KEY=your_key
cc explain tf ./terraform-module

# Format Terraform files
cc tf fmt
```

## Configuration

### Environment Variables

- `ANTHROPIC_API_KEY` - Claude API key for AI explanations
- `AWS_PROFILE` - AWS profile for Terraform operations

### Shell Profile Setup

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# Claude API key for cc explain
export ANTHROPIC_API_KEY=your_key_here

# Ensure Homebrew is in PATH (for M1/M2 Macs)
export PATH="/opt/homebrew/bin:$PATH"
```

## Contributing

This is a personal development tool. Feel free to fork and customize for your own needs!

## License

Personal use - modify as needed for your workflow.
