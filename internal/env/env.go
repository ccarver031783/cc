package env

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/christopher.carver/cc/internal/shell"
	ufcli "github.com/urfave/cli/v2"
)

// EnvironmentInfo holds information about the current environment
type EnvironmentInfo struct {
	AWSProfile   string
	AWSAccountID string
	AWSUserARN   string
	K8sContext   string
	K8sNamespace string
	GitBranch    string
	GitRemote    string
}

// getAWSInfo retrieves current AWS identity information
func getAWSInfo(ctx context.Context) (profile, accountID, userARN string) {
	// Get AWS profile from environment
	profile = os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = os.Getenv("AWS_DEFAULT_PROFILE")
	}
	if profile == "" {
		profile = "(default)"
	}

	// Get AWS caller identity
	output, err := shell.Run(ctx, "aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	if err == nil {
		accountID = strings.TrimSpace(output)
	} else {
		accountID = "(not authenticated)"
	}

	output, err = shell.Run(ctx, "aws", "sts", "get-caller-identity", "--query", "Arn", "--output", "text")
	if err == nil {
		userARN = strings.TrimSpace(output)
	} else {
		userARN = "(not authenticated)"
	}

	return profile, accountID, userARN
}

// getK8sInfo retrieves current Kubernetes context information
func getK8sInfo(ctx context.Context) (k8sContext, namespace string) {
	output, err := shell.Run(ctx, "kubectl", "config", "current-context")
	if err == nil {
		k8sContext = strings.TrimSpace(output)
	} else {
		k8sContext = "(no context set)"
	}

	output, err = shell.Run(ctx, "kubectl", "config", "view", "--minify", "--output", "jsonpath={..namespace}")
	if err == nil && strings.TrimSpace(output) != "" {
		namespace = strings.TrimSpace(output)
	} else {
		namespace = "default"
	}

	return k8sContext, namespace
}

// getGitInfo retrieves current git information
func getGitInfo(ctx context.Context) (branch, remote string) {
	output, err := shell.Run(ctx, "git", "branch", "--show-current")
	if err == nil {
		branch = strings.TrimSpace(output)
	} else {
		branch = "(not in a git repo)"
	}

	output, err = shell.Run(ctx, "git", "remote", "get-url", "origin")
	if err == nil {
		remote = strings.TrimSpace(output)
	} else {
		remote = "(no remote)"
	}

	return branch, remote
}

// showCurrentEnv displays the current environment information
func showCurrentEnv(ctx context.Context) error {
	fmt.Println("Current Environment")
	fmt.Println(strings.Repeat("=", 60))

	// AWS Information
	fmt.Println("\n📦 AWS")
	profile, accountID, userARN := getAWSInfo(ctx)
	fmt.Printf("   Profile:    %s\n", profile)
	fmt.Printf("   Account:    %s\n", accountID)
	fmt.Printf("   Identity:   %s\n", userARN)

	// Kubernetes Information
	fmt.Println("\n☸️  Kubernetes")
	k8sContext, namespace := getK8sInfo(ctx)
	fmt.Printf("   Context:    %s\n", k8sContext)
	fmt.Printf("   Namespace:  %s\n", namespace)

	// Git Information
	fmt.Println("\n🔀 Git")
	branch, remote := getGitInfo(ctx)
	fmt.Printf("   Branch:     %s\n", branch)
	fmt.Printf("   Remote:     %s\n", remote)

	fmt.Println()
	return nil
}

// listK8sContexts lists all available Kubernetes contexts
func listK8sContexts(ctx context.Context) error {
	fmt.Println("Available Kubernetes Contexts")
	fmt.Println(strings.Repeat("=", 60))

	output, err := shell.Run(ctx, "kubectl", "config", "get-contexts", "--output", "name")
	if err != nil {
		return fmt.Errorf("failed to list contexts: %w", err)
	}

	currentCtx, _ := shell.Run(ctx, "kubectl", "config", "current-context")
	currentCtx = strings.TrimSpace(currentCtx)

	contexts := strings.Split(strings.TrimSpace(output), "\n")
	for _, c := range contexts {
		if c == currentCtx {
			fmt.Printf("  * %s (current)\n", c)
		} else {
			fmt.Printf("    %s\n", c)
		}
	}

	return nil
}

// switchK8sContext switches to a different Kubernetes context
func switchK8sContext(ctx context.Context, targetContext string) error {
	fmt.Printf("Switching Kubernetes context to: %s\n", targetContext)

	_, err := shell.Run(ctx, "kubectl", "config", "use-context", targetContext)
	if err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	fmt.Printf("✓ Switched to context: %s\n", targetContext)
	return nil
}

// NewEnvCmd creates the env command
func NewEnvCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "env",
		Usage: "Display and manage environment context (AWS, K8s, Git)",
		Subcommands: []*ufcli.Command{
			{
				Name:  "show",
				Usage: "Show current environment (default action)",
				Action: func(c *ufcli.Context) error {
					return showCurrentEnv(c.Context)
				},
			},
			{
				Name:  "list",
				Usage: "List available Kubernetes contexts",
				Action: func(c *ufcli.Context) error {
					return listK8sContexts(c.Context)
				},
			},
			{
				Name:      "switch",
				Usage:     "Switch Kubernetes context",
				ArgsUsage: "<context-name>",
				Action: func(c *ufcli.Context) error {
					if c.NArg() < 1 {
						return fmt.Errorf("context name required. Usage: cc env switch <context-name>")
					}
					return switchK8sContext(c.Context, c.Args().First())
				},
			},
		},
		Action: func(c *ufcli.Context) error {
			// Default action: show current environment
			return showCurrentEnv(c.Context)
		},
	}
}

