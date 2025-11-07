package pr

import (
	"context"
	"fmt"
	"strings"

	"github.com/christopher.carver/cc/internal/shell"
	ufcli "github.com/urfave/cli/v2"
)

// NewPRCmd creates the pr command
func NewPRCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "pr",
		Usage: "Pull request operations using GitHub CLI",
		Subcommands: []*ufcli.Command{
			NewPRCreateCmd(),
			NewPRListCmd(),
			NewPRViewCmd(),
		},
	}
}

// NewPRCreateCmd creates a new pull request
func NewPRCreateCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:      "create",
		Usage:     "Create a new pull request",
		ArgsUsage: "[title]",
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:    "body",
				Aliases: []string{"b"},
				Usage:   "Body text of the pull request",
			},
			&ufcli.StringFlag{
				Name:    "base",
				Aliases: []string{"B"},
				Usage:   "Base branch (default: main/master)",
			},
			&ufcli.StringFlag{
				Name:    "head",
				Aliases: []string{"H"},
				Usage:   "Head branch (default: current branch)",
			},
			&ufcli.BoolFlag{
				Name:    "draft",
				Aliases: []string{"d"},
				Usage:   "Create as draft pull request",
			},
			&ufcli.StringSliceFlag{
				Name:    "reviewer",
				Aliases: []string{"r"},
				Usage:   "Request review from users (can be specified multiple times)",
			},
			&ufcli.StringSliceFlag{
				Name:    "label",
				Aliases: []string{"l"},
				Usage:   "Add labels to the pull request (can be specified multiple times)",
			},
		},
		Action: func(c *ufcli.Context) error {
			ctx := c.Context

			// Check if we're in a git repo
			if _, err := shell.Run(ctx, "git", "rev-parse", "--git-dir"); err != nil {
				return fmt.Errorf("not in a git repository")
			}

			// Check if gh CLI is installed
			if _, err := shell.Run(ctx, "gh", "--version"); err != nil {
				return fmt.Errorf("GitHub CLI (gh) is not installed. Please install it from https://cli.github.com")
			}

			// Get current branch if head not specified
			headBranch := c.String("head")
			if headBranch == "" {
				currentBranch, err := getCurrentBranch(ctx)
				if err != nil {
					return fmt.Errorf("failed to get current branch: %w", err)
				}
				headBranch = currentBranch
			}

			// Get base branch if not specified
			baseBranch := c.String("base")
			if baseBranch == "" {
				defaultBranch, err := getDefaultBranch(ctx)
				if err != nil {
					return fmt.Errorf("failed to determine default branch: %w", err)
				}
				baseBranch = defaultBranch
			}

			// Get PR title
			title := c.Args().First()
			if title == "" {
				// Try to get title from last commit message
				lastCommit, err := shell.Run(ctx, "git", "log", "-1", "--pretty=%s")
				if err != nil || lastCommit == "" {
					return fmt.Errorf("PR title is required. Either provide it as an argument or ensure you have a recent commit")
				}
				title = lastCommit
				fmt.Printf("Using commit message as PR title: %s\n", title)
			}

			// Build gh pr create command
			args := []string{"pr", "create"}

			// Add title
			args = append(args, "--title", title)

			// Add body if provided
			if body := c.String("body"); body != "" {
				args = append(args, "--body", body)
			}

			// Add base branch
			args = append(args, "--base", baseBranch)

			// Add head branch
			args = append(args, "--head", headBranch)

			// Add draft flag
			if c.Bool("draft") {
				args = append(args, "--draft")
			}

			// Add reviewers
			for _, reviewer := range c.StringSlice("reviewer") {
				args = append(args, "--reviewer", reviewer)
			}

			// Add labels
			for _, label := range c.StringSlice("label") {
				args = append(args, "--label", label)
			}

			fmt.Printf("Creating pull request '%s' from '%s' to '%s'...\n", title, headBranch, baseBranch)

			// Execute gh pr create
			if err := shell.RunInteractive(ctx, "gh", args...); err != nil {
				return fmt.Errorf("failed to create pull request: %w", err)
			}

			fmt.Println("✓ Pull request created successfully")
			return nil
		},
	}
}

// NewPRListCmd lists pull requests
func NewPRListCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "list",
		Usage: "List pull requests",
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:    "state",
				Aliases: []string{"s"},
				Usage:   "Filter by state: open, closed, or all",
				Value:   "open",
			},
			&ufcli.StringFlag{
				Name:    "author",
				Aliases: []string{"a"},
				Usage:   "Filter by author",
			},
			&ufcli.IntFlag{
				Name:    "limit",
				Aliases: []string{"L"},
				Usage:   "Maximum number of items to fetch",
				Value:   30,
			},
		},
		Action: func(c *ufcli.Context) error {
			ctx := c.Context

			// Check if gh CLI is installed
			if _, err := shell.Run(ctx, "gh", "--version"); err != nil {
				return fmt.Errorf("GitHub CLI (gh) is not installed. Please install it from https://cli.github.com")
			}

			args := []string{"pr", "list"}

			if state := c.String("state"); state != "" {
				args = append(args, "--state", state)
			}

			if author := c.String("author"); author != "" {
				args = append(args, "--author", author)
			}

			if limit := c.Int("limit"); limit > 0 {
				args = append(args, "--limit", fmt.Sprintf("%d", limit))
			}

			return shell.RunInteractive(ctx, "gh", args...)
		},
	}
}

// NewPRViewCmd views a pull request
func NewPRViewCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:      "view",
		Usage:     "View a pull request",
		ArgsUsage: "[number]",
		Flags: []ufcli.Flag{
			&ufcli.BoolFlag{
				Name:    "web",
				Aliases: []string{"w"},
				Usage:   "Open pull request in web browser",
			},
		},
		Action: func(c *ufcli.Context) error {
			ctx := c.Context

			// Check if gh CLI is installed
			if _, err := shell.Run(ctx, "gh", "--version"); err != nil {
				return fmt.Errorf("GitHub CLI (gh) is not installed. Please install it from https://cli.github.com")
			}

			prNumber := c.Args().First()
			args := []string{"pr", "view"}

			if prNumber != "" {
				args = append(args, prNumber)
			}

			if c.Bool("web") {
				args = append(args, "--web")
			}

			return shell.RunInteractive(ctx, "gh", args...)
		},
	}
}

// Helper functions
func getCurrentBranch(ctx context.Context) (string, error) {
	output, err := shell.Run(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func getDefaultBranch(ctx context.Context) (string, error) {
	// Try main first
	if _, err := shell.Run(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/main"); err == nil {
		return "main", nil
	}
	// Fall back to master
	if _, err := shell.Run(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/master"); err == nil {
		return "master", nil
	}
	return "", fmt.Errorf("could not determine default branch (tried main and master)")
}

