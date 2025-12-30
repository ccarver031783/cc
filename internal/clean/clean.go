package clean

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/christopher.carver/cc/internal/shell"
	ufcli "github.com/urfave/cli/v2"
)

// CleanResult holds the result of a cleanup operation
type CleanResult struct {
	ItemsFound   int
	ItemsCleaned int
	SpaceFreed   string
}

// promptConfirmation asks the user for confirmation
func promptConfirmation(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message + " (y/n): ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// cleanDocker removes dangling Docker images and stopped containers
func cleanDocker(ctx context.Context, dryRun bool) error {
	fmt.Println("🐳 Docker Cleanup")
	fmt.Println(strings.Repeat("-", 40))

	// Check if Docker is available
	_, err := shell.Run(ctx, "docker", "info")
	if err != nil {
		fmt.Println("   Docker is not running or not installed. Skipping.")
		return nil
	}

	// Get dangling images
	output, err := shell.Run(ctx, "docker", "images", "-f", "dangling=true", "-q")
	if err != nil {
		return fmt.Errorf("failed to list dangling images: %w", err)
	}

	danglingImages := strings.Split(strings.TrimSpace(output), "\n")
	imageCount := 0
	for _, img := range danglingImages {
		if img != "" {
			imageCount++
		}
	}

	// Get stopped containers
	output, err = shell.Run(ctx, "docker", "ps", "-a", "-f", "status=exited", "-q")
	if err != nil {
		return fmt.Errorf("failed to list stopped containers: %w", err)
	}

	stoppedContainers := strings.Split(strings.TrimSpace(output), "\n")
	containerCount := 0
	for _, c := range stoppedContainers {
		if c != "" {
			containerCount++
		}
	}

	fmt.Printf("   Dangling images:     %d\n", imageCount)
	fmt.Printf("   Stopped containers:  %d\n", containerCount)

	if imageCount == 0 && containerCount == 0 {
		fmt.Println("   ✓ Nothing to clean")
		return nil
	}

	if dryRun {
		fmt.Println("   [Dry run - no changes made]")
		return nil
	}

	confirm, err := promptConfirmation("   Clean Docker resources?")
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println("   Skipped")
		return nil
	}

	// Clean up
	if containerCount > 0 {
		_, err = shell.Run(ctx, "docker", "container", "prune", "-f")
		if err != nil {
			fmt.Printf("   Warning: failed to prune containers: %v\n", err)
		}
	}

	if imageCount > 0 {
		_, err = shell.Run(ctx, "docker", "image", "prune", "-f")
		if err != nil {
			fmt.Printf("   Warning: failed to prune images: %v\n", err)
		}
	}

	fmt.Println("   ✓ Docker cleanup complete")
	return nil
}

// cleanTerraform removes .terraform directories and lock files
func cleanTerraform(ctx context.Context, dryRun bool) error {
	fmt.Println("\n🏗️  Terraform Cleanup")
	fmt.Println(strings.Repeat("-", 40))

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var terraformDirs []string
	var lockFiles []string

	// Find .terraform directories and lock files
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories (except .terraform)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".terraform" {
			return filepath.SkipDir
		}

		if info.IsDir() && info.Name() == ".terraform" {
			terraformDirs = append(terraformDirs, path)
			return filepath.SkipDir
		}

		if info.Name() == ".terraform.lock.hcl" {
			lockFiles = append(lockFiles, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan for terraform files: %w", err)
	}

	fmt.Printf("   .terraform directories: %d\n", len(terraformDirs))
	fmt.Printf("   .terraform.lock.hcl:    %d\n", len(lockFiles))

	if len(terraformDirs) == 0 && len(lockFiles) == 0 {
		fmt.Println("   ✓ Nothing to clean")
		return nil
	}

	if dryRun {
		fmt.Println("   [Dry run - no changes made]")
		if len(terraformDirs) > 0 {
			fmt.Println("   Would remove:")
			for _, dir := range terraformDirs {
				relPath, _ := filepath.Rel(cwd, dir)
				fmt.Printf("     - %s\n", relPath)
			}
		}
		return nil
	}

	confirm, err := promptConfirmation("   Remove .terraform directories?")
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println("   Skipped")
		return nil
	}

	// Remove directories
	for _, dir := range terraformDirs {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("   Warning: failed to remove %s: %v\n", dir, err)
		}
	}

	fmt.Printf("   ✓ Removed %d .terraform directories\n", len(terraformDirs))
	return nil
}

// cleanGo removes Go build cache and test cache
func cleanGo(ctx context.Context, dryRun bool) error {
	fmt.Println("\n🐹 Go Cleanup")
	fmt.Println(strings.Repeat("-", 40))

	// Check if Go is available
	_, err := shell.Run(ctx, "go", "version")
	if err != nil {
		fmt.Println("   Go is not installed. Skipping.")
		return nil
	}

	// Get cache sizes
	buildCache, _ := shell.Run(ctx, "go", "env", "GOCACHE")
	buildCache = strings.TrimSpace(buildCache)

	modCache, _ := shell.Run(ctx, "go", "env", "GOMODCACHE")
	modCache = strings.TrimSpace(modCache)

	fmt.Printf("   Build cache: %s\n", buildCache)
	fmt.Printf("   Module cache: %s\n", modCache)

	if dryRun {
		fmt.Println("   [Dry run - no changes made]")
		return nil
	}

	confirm, err := promptConfirmation("   Clean Go build cache?")
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println("   Skipped")
		return nil
	}

	// Clean build cache (not module cache - that's usually shared)
	_, err = shell.Run(ctx, "go", "clean", "-cache")
	if err != nil {
		return fmt.Errorf("failed to clean Go cache: %w", err)
	}

	fmt.Println("   ✓ Go build cache cleaned")
	return nil
}

// cleanAll runs all cleanup operations
func cleanAll(ctx context.Context, dryRun bool) error {
	fmt.Println("🧹 Running all cleanup operations")
	fmt.Println(strings.Repeat("=", 60))

	if err := cleanDocker(ctx, dryRun); err != nil {
		fmt.Printf("Warning: Docker cleanup error: %v\n", err)
	}

	if err := cleanTerraform(ctx, dryRun); err != nil {
		fmt.Printf("Warning: Terraform cleanup error: %v\n", err)
	}

	if err := cleanGo(ctx, dryRun); err != nil {
		fmt.Printf("Warning: Go cleanup error: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ Cleanup complete!")
	return nil
}

// NewCleanCmd creates the clean command
func NewCleanCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "clean",
		Usage: "Clean up development cruft (Docker, Terraform, Go caches)",
		Flags: []ufcli.Flag{
			&ufcli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Show what would be cleaned without making changes",
			},
		},
		Subcommands: []*ufcli.Command{
			{
				Name:  "docker",
				Usage: "Clean dangling Docker images and stopped containers",
				Flags: []ufcli.Flag{
					&ufcli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Show what would be cleaned without making changes",
					},
				},
				Action: func(c *ufcli.Context) error {
					return cleanDocker(c.Context, c.Bool("dry-run"))
				},
			},
			{
				Name:  "terraform",
				Usage: "Clean .terraform directories in current tree",
				Flags: []ufcli.Flag{
					&ufcli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Show what would be cleaned without making changes",
					},
				},
				Action: func(c *ufcli.Context) error {
					return cleanTerraform(c.Context, c.Bool("dry-run"))
				},
			},
			{
				Name:  "go",
				Usage: "Clean Go build cache",
				Flags: []ufcli.Flag{
					&ufcli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Show what would be cleaned without making changes",
					},
				},
				Action: func(c *ufcli.Context) error {
					return cleanGo(c.Context, c.Bool("dry-run"))
				},
			},
			{
				Name:  "all",
				Usage: "Run all cleanup operations",
				Flags: []ufcli.Flag{
					&ufcli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Show what would be cleaned without making changes",
					},
				},
				Action: func(c *ufcli.Context) error {
					return cleanAll(c.Context, c.Bool("dry-run"))
				},
			},
		},
		Action: func(c *ufcli.Context) error {
			// Default action: run all cleanup with interactive prompts
			return cleanAll(c.Context, c.Bool("dry-run"))
		},
	}
}

