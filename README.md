# cc - Personal Dev Tool CLI

A simple, shortcut-based CLI tool for simplifying and automating common development workflows including git operations, PR management, and Terraform validation.

## Project Structure

```
cc/
├── .github/
│   └── workflows/
│       └── auto-tag.yml         # Auto-tagging on push to main
├── cmd/
│   └── cc/
│       └── main.go              # Entry point
├── eksupgrade/                  # EKS upgrade helper (used by cc CLI)
│   ├── config.go                # YAML config and target versions
│   ├── cmd.go                   # eks upgrade subcommand
│   ├── diff.go                  # Diff logic and formatting
│   ├── diff_step.go             # Per-step diff filtering
│   ├── discover.go              # Cluster directory discovery
│   └── parse.go                 # Parse terraform.tfvars, eks.tf, etc.
├── eks_upgrade/                 # Sanitized, open-source-ready copy of eksupgrade
│   ├── config.go
│   ├── cmd.go
│   ├── diff.go
│   ├── diff_step.go
│   ├── discover.go
│   ├── parse.go
│   ├── eks-upgrade-config.example.yaml   # Example config (generic)
│   └── README.md                # Package documentation
├── internal/
│   ├── git/                     # Git operations
│   │   └── git.go              # Branch, rebase, clean, status
│   ├── pr/                      # PR creation/management (GitHub CLI)
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
├── eks-upgrade-config.yaml      # EKS upgrade target config (used by cc eks upgrade)
├── .mise.toml                   # Mise tool configuration
├── CHANGELOG.md                 # Version history and changes
├── ROADMAP.md                   # Planned features and priorities
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## Prerequisites & Installation

### Install Mise (Tool Version Manager)

cc uses [mise](https://mise.jdx.dev/) to manage all development dependencies.

```bash
# Install mise (if not already installed)
curl https://mise.run | sh

# Add mise to your shell (add to ~/.zshrc or ~/.bashrc)
eval "$(mise activate zsh)"  # or bash
```

### Install Dependencies

```bash
# Clone the repository
git clone <your-repo>
cd cc

# Install all dependencies via mise
mise install
```

This installs:
- **go** - Go compiler (latest)
- **terraform** - Terraform CLI (pinned to 1.5.7)
- **tflint** - Terraform linter
- **tfsec** - Terraform security scanner
- **gh** - GitHub CLI
- **ollama** - Local AI for explanations
- **golangci-lint** - Go linter

### Build the CLI

```bash
# Using mise task
mise run build

# Or directly
go build -o cc ./cmd/cc

# Verify it works
./cc --help
```

## Mise Tasks

The project includes mise tasks for common operations:

```bash
mise run build    # Build the cc binary
mise run test     # Run all tests
mise run install  # Install cc to GOBIN
mise run fmt      # Format Go code
mise run lint     # Lint code with golangci-lint
mise run run      # Run without building
mise run clean    # Remove built binary
```

## Core Features

### 1. Git Operations (`git` command)

```bash
cc git branch <name>         # Create new branch from clean main/master
cc git rebase <target-branch> # Rebase current branch onto specified branch
cc git clean                  # Clean working directory (stash changes, reset)
cc git status                 # Enhanced git status with branch info
```

### 2. PR Management (`pr` command)

```bash
cc pr create [--draft]        # Create PR from current branch using GitHub CLI
cc pr list                    # List open PRs (via gh CLI)
cc pr view <number>           # View PR details (via gh CLI)
```

**Note:** These commands should work whether these are user or AI/automated PR creation.

### 3. Terraform Operations (`terraform` or `tf` command)

```bash
cc tf fmt                     # Format Terraform files (changed files only)
cc tf scan                    # Run security scan with tfsec or tflint (changed files only)
cc tf validate                # Validate Terraform config
cc tf pre-push                # Run fmt + scan + validate on changed files before push
cc tf init-dir <path>         # Scaffold a new Terraform directory
cc tf new <resource-name>     # Create multi-provider resource structure
```

### 4. AI-Powered Explanations (`explain` command)

```bash
cc explain tf [path]          # Explain Terraform modules using AI
cc explain tf . --local       # Use local Ollama instead of Claude API
```

The explain command analyzes Terraform modules and provides clear explanations, including:
- Purpose and functionality
- Resources created
- Key variables and outputs
- Dependencies and use cases

**Dual AI Support:**
- **Claude API** (cloud) - High quality, requires API key
- **Ollama** (local) - Free, unlimited, works offline

See [internal/explain/README.md](internal/explain/README.md) for setup details.

### 5. Environment Context (`env` command)

```bash
cc env                   # Show current AWS, K8s, and Git context
cc env list              # List available Kubernetes contexts
cc env switch <context>  # Switch Kubernetes context
```

Displays your current environment at a glance:
- AWS profile, account, and identity
- Kubernetes context and namespace
- Git branch and remote

### 6. Cleanup (`clean` command)

```bash
cc clean                 # Interactive cleanup of all cruft
cc clean docker          # Remove dangling images and stopped containers
cc clean terraform       # Remove .terraform directories
cc clean go              # Clean Go build cache
cc clean --dry-run       # Preview what would be cleaned
```

### 7. EKS Upgrade Helper (`eks` command)

```bash
cc eks upgrade                    # Diff all clusters vs target config
cc eks upgrade --step 1            # Step 1 only (addons)
cc eks upgrade --step 2            # Step 2 only (control plane, module, provider)
cc eks upgrade --step 3            # Step 3 only (remaining addons, eks.tf blocks)
cc eks upgrade --config path.yaml # Custom config
cc eks upgrade --infra-path /path  # Path to infrastructure-terraform repo
cc eks upgrade --json              # JSON output for automation
```

Discovers EKS cluster directories in your Terraform repo, parses current state, and diffs against `eks-upgrade-config.yaml`. Supports step-based upgrades (addons → control plane → remaining config). See [eks_upgrade/README.md](eks_upgrade/README.md) for details.

### 8. Pre-Push Hook

```bash
cc hook install                # Install git pre-push hook
```

Hook automatically runs: `terraform fmt`, `scan` (tfsec/tflint), `validate` on changed files before allowing push.

This Hook should work in both manual and automated (AI/CI) contexts.

## Technology Stack

- **Tool Management:** [mise](https://mise.jdx.dev/) - Polyglot runtime manager
- **CLI Framework:** `github.com/urfave/cli/v2` - Command-line interface structure
- **AI Integration:**
  - `github.com/anthropics/anthropic-sdk-go` - Claude API client
  - HTTP client for Ollama local LLM integration
- **Git Operations:** Shell commands via `os/exec`
- **GitHub API:** `github.com/cli/cli` (gh CLI) wrapper
- **Terraform:** Shell execution of terraform/tfsec/tflint binaries

## Quick Start

```bash
# Install dependencies
mise install

# Build the CLI
mise run build

# Check git status
./cc git status

# Explain a Terraform module
export ANTHROPIC_API_KEY=your_key
./cc explain tf ./terraform-module

# Format Terraform files
./cc tf fmt
```

## Configuration

### Environment Variables

- `ANTHROPIC_API_KEY` - Claude API key for AI explanations

### Due to the various mechanisms used to manage AWS access, exporting and setting an AWS_PROFILE is not part of this tool by design. That step will need to be performed manually, if necessary.

### Shell Profile Setup

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# Mise activation (required)
eval "$(mise activate zsh)"  # or bash

# Claude API key for cc explain (optional)
export ANTHROPIC_API_KEY=your_key_here
```

## Contributing

Contributions are welcome! When making changes:

1. Update the `[Unreleased]` section in `CHANGELOG.md` with your changes
2. Follow the existing code style (run `mise run fmt` and `mise run lint`)
3. Test your changes with `mise run test`

### Versioning

This project follows [Semantic Versioning](https://semver.org/) with automated patch tagging:

| Component | How It's Managed |
|-----------|------------------|
| **MAJOR** | Manual - update `Version` in main.go for breaking changes |
| **MINOR** | Manual - update `Version` in main.go for new features |
| **PATCH** | Automatic - increments on each push to main |

**Example:**
```
Version = "1.0.0" in main.go
Push → v1.0.0
Push → v1.0.1
Push → v1.0.2

Update to Version = "1.1.0" (new feature)
Push → v1.1.0
Push → v1.1.1
```

To release a new minor version, simply update the `Version` constant in `cmd/cc/main.go`.

## License

Personal use - modify as needed for your workflow. Just have fun with it!
