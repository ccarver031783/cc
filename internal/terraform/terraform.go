package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/christopher.carver/cc/internal/shell"

	ufcli "github.com/urfave/cli/v2"
)

// ============================================================================
// Helper Functions
// ============================================================================

// validatePath ensures the path is safe and resolves to a valid location.
// It prevents path traversal attacks while allowing complex directory structures.
// Returns an absolute path if valid, or an error if the path is unsafe or invalid.
func validatePath(path string) (string, error) {
	if path == "" {
		// Empty path means current directory - that's fine
		return ".", nil
	}

	// Clean the path to resolve any ".." or "." components
	// This converts paths like "../../etc/passwd" to their resolved form
	cleanedPath := filepath.Clean(path)

	// Get the absolute path to prevent any remaining traversal
	absPath, err := filepath.Abs(cleanedPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Get current working directory to ensure path is within it
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if the absolute path is within the current working directory
	// This prevents accessing files outside your project
	relPath, err := filepath.Rel(cwd, absPath)
	if err != nil {
		return "", fmt.Errorf("path outside working directory: %w", err)
	}

	// Check for path traversal attempts (paths starting with "..")
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path traversal detected: %s", path)
	}
	return absPath, nil
}

// ============================================================================
// Main Command
// ============================================================================

// NewTerraformCmd creates the main terraform command with all subcommands.
// This is the entry point for all terraform-related operations.
func NewTerraformCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "terraform",
		Usage: "Terraform operations and shortcuts",
		Subcommands: []*ufcli.Command{
			// Basic Terraform Commands (Core Workflow)
			NewTerraformInitCmd(),
			NewTerraformFormatCmd(),
			NewTerraformValidateCmd(),
			NewTerraformPlanCmd(),
			NewTerraformApplyCmd(),
			NewTerraformDestroyCmd(),
			// Security & Validation Commands
			NewTerraformScanCmd(),
			NewTerraformCheckCmd(),
			// State & Information Commands
			NewTerraformStateListCmd(),
			NewTerraformOutputCmd(),
			NewTerraformShowCmd(),
			NewTerraformTestCmd(),
			NewTerraformProviderCmd(),
			NewTerraformWorkspaceCmd(),
			NewTerraformGraphCmd(),
		},
	}
}

// ============================================================================
// Basic Terraform Commands (Core Workflow)
// ============================================================================

// NewTerraformInitCmd creates the init command.
// Initializes a Terraform working directory by downloading providers and modules.
// This is typically the first command run in a new Terraform project.
func NewTerraformInitCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "init",
		Usage: "Initialize Terraform working directory",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "init", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformFormatCmd creates the fmt command.
// Formats Terraform configuration files to a canonical format and style.
// This ensures consistent code style across the project.
func NewTerraformFormatCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "fmt",
		Usage: "Format Terraform configuration files",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "fmt", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformValidateCmd creates the validate command.
// Validates Terraform configuration files for syntax errors and internal consistency.
// Does not check against external APIs or verify resource existence.
func NewTerraformValidateCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "validate",
		Usage: "Validate Terraform configuration syntax",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "validate", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformPlanCmd creates the plan command.
// Creates an execution plan showing what actions Terraform will take to reach
// the desired state. This is a dry-run that doesn't make any changes.
func NewTerraformPlanCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "plan",
		Usage: "Generate and show an execution plan",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "plan", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformApplyCmd creates the apply command.
// Applies the changes required to reach the desired state of the configuration.
// This command modifies real infrastructure and should be used with caution.
func NewTerraformApplyCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "apply",
		Usage: "Apply Terraform changes to infrastructure",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "apply", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformDestroyCmd creates the destroy command.
// Destroys all resources managed by the Terraform configuration.
// This is a destructive operation that permanently removes infrastructure.
func NewTerraformDestroyCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "destroy",
		Usage: "Destroy Terraform-managed infrastructure",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "destroy", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// ============================================================================
// Security & Validation Commands
// ============================================================================

// NewTerraformScanCmd creates the scan command.
// Runs security scanning tools (tfsec or tflint) on changed Terraform files.
// Only scans files that have been modified between HEAD and origin/main,
// making it efficient for large repositories. Falls back to HEAD~1 if origin/main
// is not available. Filters results to only .tf files.
func NewTerraformScanCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "scan",
		Usage: "Run tfsec or tflint on changed Terraform files",
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:    "tool",
				Aliases: []string{"t"},
				Usage:   "Security tool to use: tfsec or tflint",
				Value:   "tfsec",
			},
		},
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			tool := c.String("tool")
			// Validate tool flag to prevent command injection
			if tool != "tfsec" && tool != "tflint" {
				return fmt.Errorf("tool must be either 'tfsec' or 'tflint', got %s", tool)
			}

			// Step 1: Get changed files from git
			// Attempts to get files changed between HEAD and origin/main
			output, err := shell.Run(ctx, "git", "diff", "--name-only", "HEAD", "origin/main")
			if err != nil {
				// Fallback: try comparing with HEAD~1 (previous commit) if origin/main doesn't exist
				output, err = shell.Run(ctx, "git", "diff", "--name-only", "HEAD~1", "HEAD")
				if err != nil {
					return fmt.Errorf("failed to get changed files: %w", err)
				}
			}

			// Step 2: Parse the output - split by newlines to get individual files
			changedFiles := strings.Split(output, "\n")

			// Step 3: Filter for only .tf files
			var tfFiles []string
			for _, file := range changedFiles {
				file = strings.TrimSpace(file)
				if file == "" {
					continue // Skip empty lines
				}
				if strings.HasSuffix(file, ".tf") {
					tfFiles = append(tfFiles, file)
				}
			}

			// Step 4: If no .tf files changed, exit early
			if len(tfFiles) == 0 {
				fmt.Println("No Terraform files changed")
				return nil
			}

			// Step 5: Run the tool on the changed .tf files
			// Join the files with spaces to pass as arguments
			filesArg := strings.Join(tfFiles, " ")

			// Run the tool (tfsec or tflint) on the files
			_, err = shell.Run(ctx, tool, filesArg)
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			return nil
		},
	}
}

// NewTerraformCheckCmd creates the check command.
// Runs a comprehensive pre-push workflow: formats files, validates syntax,
// and runs both tflint and tfsec security scans. This is designed to be used
// as a pre-push hook to ensure code quality before committing changes.
// Stops at the first failure to provide fast feedback.
func NewTerraformCheckCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "check",
		Usage: "Run fmt, validate, and security scans (pre-push workflow)",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}

			// Step 1: Format files
			_, err = shell.Run(ctx, "terraform", "fmt", safePath)
			if err != nil {
				return err
			}

			// Step 2: Validate syntax
			_, err = shell.Run(ctx, "terraform", "validate", safePath)
			if err != nil {
				return err
			}

			// Step 3: Run tflint (code quality linter)
			_, err = shell.Run(ctx, "tflint", safePath)
			if err != nil {
				return err
			}

			// Step 4: Run tfsec (security scanner)
			_, err = shell.Run(ctx, "tfsec", safePath)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

// ============================================================================
// State & Information Commands
// ============================================================================

// NewTerraformStateListCmd creates the state-list command.
// Lists all resources currently tracked in the Terraform state.
// Useful for auditing what infrastructure Terraform is managing.
func NewTerraformStateListCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "state-list",
		Usage: "List resources in Terraform state",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			_, err := shell.Run(ctx, "terraform", "state", "list")
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformOutputCmd creates the output command.
// Shows the values of output variables defined in the Terraform configuration.
// Outputs are typically used to expose important values like resource IDs or endpoints.
func NewTerraformOutputCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "output",
		Usage: "Show Terraform output values",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			_, err := shell.Run(ctx, "terraform", "output")
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformShowCmd creates the show command.
// Displays human-readable output from a state file or plan file.
// Useful for inspecting the current or planned state of infrastructure.
func NewTerraformShowCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "show",
		Usage: "Show Terraform state or plan in human-readable format",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "show", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformTestCmd creates the test command.
// Runs Terraform tests defined in the configuration.
// Tests verify that Terraform configurations behave as expected.
func NewTerraformTestCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "test",
		Usage: "Run Terraform tests",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			path := c.String("path")
			safePath, err := validatePath(path)
			if err != nil {
				return err
			}
			_, err = shell.Run(ctx, "terraform", "test", safePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformProviderCmd creates the providers command.
// Lists all providers required by the current configuration.
// Shows which providers Terraform will download during init.
func NewTerraformProviderCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "providers",
		Usage: "List Terraform providers",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			_, err := shell.Run(ctx, "terraform", "providers")
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformWorkspaceCmd creates the workspace command.
// Manages Terraform workspaces, which allow multiple state files for the same configuration.
// Workspaces enable managing multiple environments (dev, staging, prod) with one config.
func NewTerraformWorkspaceCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "workspace",
		Usage: "Manage Terraform workspaces",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			_, err := shell.Run(ctx, "terraform", "workspace")
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewTerraformGraphCmd creates the graph command.
// Generates a visual representation of the Terraform dependency graph.
// Output can be piped to GraphViz tools for visualization.
func NewTerraformGraphCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "graph",
		Usage: "Generate a GraphViz graph of Terraform dependencies",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context
			operation := c.String("operation")
			_, err := shell.Run(ctx, "terraform", "graph", operation)
			if err != nil {
				return err
			}
			return nil
		},
	}
}
