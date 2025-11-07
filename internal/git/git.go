package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/christopher.carver/cc/internal/shell"
	ufcli "github.com/urfave/cli/v2"
)

// NewGitCmd creates the git command
func NewGitCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "git",
		Usage: "Git operations and shortcuts",
		Subcommands: []*ufcli.Command{
			NewGitBranchCmd(),
			NewGitRebaseCmd(),
			NewGitCleanCmd(),
			NewGitStatusCmd(),
		},
	}
}

// NewGitBranchCmd creates a new branch from main/master
func NewGitBranchCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:      "branch",
		Usage:     "Create a new branch from clean main/master",
		ArgsUsage: "<branch-name>",
		Action: func(c *ufcli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("branch name is required")
			}
			branchName := c.Args().First()
			ctx := c.Context

			// Check if we're in a git repo
			if _, err := shell.Run(ctx, "git", "rev-parse", "--git-dir"); err != nil {
				return fmt.Errorf("not in a git repository")
			}

			// Get the default branch (main or master)
			defaultBranch, err := getDefaultBranch(ctx)
			if err != nil {
				return fmt.Errorf("failed to determine default branch: %w", err)
			}

			fmt.Printf("Creating branch '%s' from '%s'...\n", branchName, defaultBranch)

			// Stash any uncommitted changes
			hasChanges, err := hasUncommittedChanges(ctx)
			if err != nil {
				return fmt.Errorf("failed to check for uncommitted changes: %w", err)
			}

			if hasChanges {
				fmt.Println("Stashing uncommitted changes...")
				if _, err := shell.Run(ctx, "git", "stash"); err != nil {
					return fmt.Errorf("failed to stash changes: %w", err)
				}
				defer func() {
					fmt.Println("Restoring stashed changes...")
					shell.Run(ctx, "git", "stash", "pop")
				}()
			}

			// Checkout default branch and pull latest
			fmt.Printf("Checking out '%s' and pulling latest...\n", defaultBranch)
			if _, err := shell.Run(ctx, "git", "checkout", defaultBranch); err != nil {
				return fmt.Errorf("failed to checkout %s: %w", defaultBranch, err)
			}
			if _, err := shell.Run(ctx, "git", "pull"); err != nil {
				return fmt.Errorf("failed to pull latest: %w", err)
			}

			// Create and checkout new branch
			fmt.Printf("Creating and checking out branch '%s'...\n", branchName)
			if _, err := shell.Run(ctx, "git", "checkout", "-b", branchName); err != nil {
				return fmt.Errorf("failed to create branch: %w", err)
			}

			fmt.Printf("✓ Successfully created branch '%s' from '%s'\n", branchName, defaultBranch)
			return nil
		},
	}
}

// NewGitRebaseCmd rebases current branch onto target branch with 3-step workflow
func NewGitRebaseCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:      "rebase",
		Usage:     "Rebase current branch onto target branch (fetch, rebase, force push)",
		ArgsUsage: "<target-branch>",
		Action: func(c *ufcli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("target branch is required")
			}
			targetBranch := c.Args().First()
			ctx := c.Context

			// Check if we're in a git repo
			if _, err := shell.Run(ctx, "git", "rev-parse", "--git-dir"); err != nil {
				return fmt.Errorf("not in a git repository")
			}

			// Get current branch
			currentBranch, err := getCurrentBranch(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			if currentBranch == targetBranch {
				return fmt.Errorf("cannot rebase branch onto itself")
			}

			fmt.Printf("Rebasing '%s' onto '%s'...\n", currentBranch, targetBranch)

			// Check for uncommitted changes
			hasChanges, err := hasUncommittedChanges(ctx)
			if err != nil {
				return fmt.Errorf("failed to check for uncommitted changes: %w", err)
			}

			if hasChanges {
				return fmt.Errorf("uncommitted changes detected. Please commit or stash before rebasing")
			}

			// Step 1: Fetch origin targetBranch:targetBranch
			fmt.Printf("Step 1: Fetching origin %s:%s...\n", targetBranch, targetBranch)
			if _, err := shell.Run(ctx, "git", "fetch", "origin", fmt.Sprintf("%s:%s", targetBranch, targetBranch)); err != nil {
				return fmt.Errorf("failed to fetch origin %s:%s: %w", targetBranch, targetBranch, err)
			}

			// Step 2: Rebase onto target branch
			fmt.Printf("Step 2: Rebasing onto '%s'...\n", targetBranch)
			if err := shell.RunInteractive(ctx, "git", "rebase", targetBranch); err != nil {
				return fmt.Errorf("rebase failed: %w", err)
			}

			// Step 3: Force push to origin
			fmt.Printf("Step 3: Force pushing '%s' to origin...\n", currentBranch)
			if _, err := shell.Run(ctx, "git", "push", "-f", "origin", currentBranch); err != nil {
				return fmt.Errorf("failed to force push: %w", err)
			}

			fmt.Printf("✓ Successfully rebased and pushed '%s' onto '%s'\n", currentBranch, targetBranch)
			return nil
		},
	}
}

// NewGitCleanCmd cleans the working directory
func NewGitCleanCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "clean",
		Usage: "Clean working directory (stash changes, reset to HEAD)",
		Flags: []ufcli.Flag{
			&ufcli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force clean (discard all changes)",
			},
		},
		Action: func(c *ufcli.Context) error {
			ctx := c.Context

			// Check if we're in a git repo
			if _, err := shell.Run(ctx, "git", "rev-parse", "--git-dir"); err != nil {
				return fmt.Errorf("not in a git repository")
			}

			// Check for uncommitted changes
			hasChanges, err := hasUncommittedChanges(ctx)
			if err != nil {
				return fmt.Errorf("failed to check for uncommitted changes: %w", err)
			}

			if !hasChanges {
				fmt.Println("Working directory is already clean")
				return nil
			}

			if c.Bool("force") {
				fmt.Println("Discarding all changes...")
				if _, err := shell.Run(ctx, "git", "reset", "--hard", "HEAD"); err != nil {
					return fmt.Errorf("failed to reset: %w", err)
				}
				if _, err := shell.Run(ctx, "git", "clean", "-fd"); err != nil {
					return fmt.Errorf("failed to clean: %w", err)
				}
				fmt.Println("✓ Working directory cleaned")
			} else {
				fmt.Println("Stashing changes...")
				if _, err := shell.Run(ctx, "git", "stash"); err != nil {
					return fmt.Errorf("failed to stash: %w", err)
				}
				fmt.Println("✓ Changes stashed")
			}

			return nil
		},
	}
}

// NewGitStatusCmd shows enhanced git status
func NewGitStatusCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "status",
		Usage: "Show enhanced git status with branch info",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context

			// Check if we're in a git repo
			if _, err := shell.Run(ctx, "git", "rev-parse", "--git-dir"); err != nil {
				return fmt.Errorf("not in a git repository")
			}

			// Get current branch
			currentBranch, err := getCurrentBranch(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Get default branch
			defaultBranch, err := getDefaultBranch(ctx)
			if err != nil {
				return fmt.Errorf("failed to determine default branch: %w", err)
			}

			// Show branch info
			fmt.Printf("Current branch: %s\n", currentBranch)
			if currentBranch != defaultBranch {
				// Check if branch is ahead/behind
				ahead, behind, err := getBranchStatus(ctx, currentBranch, defaultBranch)
				if err == nil {
					if ahead > 0 {
						fmt.Printf("  Ahead of %s by %d commit(s)\n", defaultBranch, ahead)
					}
					if behind > 0 {
						fmt.Printf("  Behind %s by %d commit(s)\n", defaultBranch, behind)
					}
				}
			}
			fmt.Println()

			// Show git status
			status, err := shell.Run(ctx, "git", "status", "--short")
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			if status == "" {
				fmt.Println("Working directory is clean")
			} else {
				fmt.Println("Changes:")
				fmt.Println(status)
			}

			return nil
		},
	}
}

// Helper functions

func getCurrentBranch(ctx context.Context) (string, error) {
	output, err := shell.Run(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return output, nil
}

func getDefaultBranch(ctx context.Context) (string, error) {
	// Try to get the default branch from remote
	output, err := shell.Run(ctx, "git", "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		// Parse output like "refs/remotes/origin/main"
		parts := strings.Split(output, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if main exists, otherwise use master
	if _, err := shell.Run(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/main"); err == nil {
		return "main", nil
	}
	return "master", nil
}

func hasUncommittedChanges(ctx context.Context) (bool, error) {
	output, err := shell.Run(ctx, "git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

func getBranchStatus(ctx context.Context, current, target string) (ahead, behind int, err error) {
	// Get the merge base
	mergeBase, err := shell.Run(ctx, "git", "merge-base", current, target)
	if err != nil {
		return 0, 0, err
	}

	// Count commits ahead
	aheadOutput, err := shell.Run(ctx, "git", "rev-list", "--count", fmt.Sprintf("%s..%s", mergeBase, current))
	if err != nil {
		return 0, 0, err
	}
	fmt.Sscanf(aheadOutput, "%d", &ahead)

	// Count commits behind
	behindOutput, err := shell.Run(ctx, "git", "rev-list", "--count", fmt.Sprintf("%s..%s", mergeBase, target))
	if err != nil {
		return 0, 0, err
	}
	fmt.Sscanf(behindOutput, "%d", &behind)

	return ahead, behind, nil
}
