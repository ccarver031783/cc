# cc - Personal Dev Tool CLI

A simple, shortcut-based CLI tool for simplifying and automating common development workflows including git operations, PR management, and Terraform validation.

## Project Structure

```
cc/
├── cmd/
│   └── cc/
│       └── main.go          # Entry point
├── internal/
│   ├── git/                 # Git operations
│   ├── pr/                  # PR creation/management (GitHub CLI)
│   ├── setup/               # Homebrew package management
│   ├── terraform/           # Terraform operations
│   └── shell/               # Shell execution utilities
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## Core Features

### 1. Setup (`setup` command)

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

### 2. Git Operations (`git` command)

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

**Note:** Commands must work whether user or AI/automation creates the PR.

### 3. Terraform Operations (`terraform` or `tf` command)

```bash
cc tf fmt                     # Format Terraform files (changed files only)
cc tf scan                    # Run security scan with tfsec or tflint (changed files only)
cc tf validate                # Validate Terraform config
cc tf pre-push                # Run fmt + scan + validate on changed files before push
cc tf init-dir <path>         # Scaffold a new Terraform directory
cc tf new <resource-name>     # Create multi-provider resource structure
```

### 4. Pre-Push Hook

```bash
cc hook install                # Install git pre-push hook
```

Hook automatically runs: `terraform fmt`, `scan` (tfsec/tflint), `validate` on changed files before allowing push.

Hook must work in both manual and automated (AI/CI) contexts.

## Technology Stack

- **CLI Framework:** `github.com/urfave/cli/v2`
- **Git Operations:** `github.com/go-git/go-git/v5` or shell commands
- **GitHub API:** `github.com/cli/cli` (gh CLI) or `github.com/google/go-github`
- **Terraform:** Shell execution of terraform/tfsec/checkov binaries
- **Logging:** Simple fmt or structured logging

## Implementation Steps

### Project Setup
- Initialize Go module
- Set up basic CLI structure with urfave/cli
- Create command scaffolding

### Git Module
- Implement branch creation
- Implement rebase operations
- Add working directory checks

### PR Module
- Integrate with GitHub CLI or API
- Implement PR creation with proper base branch detection
- Add PR listing/viewing

### Terraform Module
- Implement fmt command
- Integrate security scanner (tfsec recommended)
- Add validation
- Create pre-push check command

### Git Hooks
- Create pre-push hook script
- Install/uninstall functionality

### Documentation
- README with usage examples
- Command help text

### Configuration
- Config file: `~/.devtools/config.yaml` (optional)
- Default branch names (main/master)
- Terraform scanner preference
- GitHub repo defaults
