package explain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ufcli "github.com/urfave/cli/v2"
)

// NewExplainCmd creates the explain command
func NewExplainCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "explain",
		Usage: "Explain Terraform modules using AI",
		Subcommands: []*ufcli.Command{
			{
				Name:      "tf",
				Usage:     "Explain a Terraform module or directory",
				ArgsUsage: "[path]",
				Flags: []ufcli.Flag{
					&ufcli.BoolFlag{
						Name:    "local",
						Aliases: []string{"l"},
						Usage:   "Force use of local Ollama (skip Claude API)",
					},
				},
				Action: func(c *ufcli.Context) error {
					path := c.Args().First()
					if path == "" {
						path = "."
					}

					safePath, err := validatePath(path)
					if err != nil {
						return err
					}

					forceLocal := c.Bool("local")
					return explainTerraform(c.Context, safePath, forceLocal)
				},
			},
		},
	}
}

// validatePath ensures the path exists and is safe to read
func validatePath(path string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}

	// Must be a directory
	if !info.IsDir() {
		return "", fmt.Errorf("path must be a directory")
	}

	return absPath, nil
}

// explainTerraform reads Terraform files and generates an explanation
func explainTerraform(ctx context.Context, path string, forceLocal bool) error {
	fmt.Printf("Analyzing Terraform module at: %s\n\n", path)

	// Read common Terraform files
	files := []string{"main.tf", "variables.tf", "outputs.tf", "versions.tf", "README.md"}
	var content []string
	var foundFiles []string

	for _, file := range files {
		fullPath := filepath.Join(path, file)
		data, err := os.ReadFile(fullPath)
		if err == nil {
			content = append(content, fmt.Sprintf("=== %s ===\n%s", file, string(data)))
			foundFiles = append(foundFiles, file)
		}
	}

	if len(content) == 0 {
		return fmt.Errorf("no Terraform files found in %s", path)
	}

	fmt.Printf("Found files: %s\n\n", strings.Join(foundFiles, ", "))

	// Build the prompt
	prompt := buildPrompt(strings.Join(content, "\n\n"))

	// Get explanation from AI
	fmt.Println("Generating explanation...")
	explanation, err := callAI(ctx, prompt, forceLocal)
	if err != nil {
		return fmt.Errorf("failed to generate explanation: %w", err)
	}

	// Display results
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("EXPLANATION")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(explanation)
	fmt.Println(strings.Repeat("=", 80))

	return nil
}

// buildPrompt creates the AI prompt from Terraform files
func buildPrompt(moduleText string) string {
	return fmt.Sprintf(`You are a Terraform expert. Analyze the following Terraform module files and provide a clear, concise explanation.

Your explanation should include:
1. **Purpose**: What does this module do?
2. **Resources**: What resources does it create or manage?
3. **Key Variables**: What are the important inputs and what do they control?
4. **Outputs**: What values does it expose for other modules?
5. **Dependencies**: Does it depend on or integrate with other infrastructure?
6. **Use Case**: When would you use this module?

Be specific but concise. Focus on what matters to a developer trying to understand this code.

Module Files:
%s

Provide your explanation in markdown format.`, moduleText)
}

// callAI sends the prompt to an AI service (Claude or Ollama)
func callAI(ctx context.Context, prompt string, forceLocal bool) (string, error) {
	// Try Claude API first (unless forced to use local)
	if !forceLocal {
		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			fmt.Println("Using Claude API...")
			response, err := callClaude(ctx, prompt, apiKey)
			if err == nil {
				return response, nil
			}
			fmt.Printf("⚠️  Claude API failed: %v\n", err)
			fmt.Println("Falling back to local Ollama...")
		}
	}

	// Fallback to Ollama
	fmt.Println("Using local Ollama...")
	return callOllama(ctx, prompt)
}
