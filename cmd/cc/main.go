package main

import (
	"context"
	"fmt"
	"os"

	"github.com/christopher.carver/cc/internal/clean"
	"github.com/christopher.carver/cc/internal/env"
	"github.com/christopher.carver/cc/internal/explain"
	"github.com/christopher.carver/cc/internal/git"
	"github.com/christopher.carver/cc/internal/terraform"
	"github.com/christopher.carver/cc/internal/uuid"
	"github.com/christopher.carver/cc/internal/whoami"
	ufcli "github.com/urfave/cli/v2"
)

// Version information - update these when releasing
const (
	Version   = "1.2.0"
	BuildDate = "2025-12-30"
)

func main() {
	ctx := context.Background()

	app := &ufcli.App{
		Name:    "cc",
		Version: Version,
		Usage:   "Development and SRE-based CLI tool - turning cc commands into shortcuts for git and Terraform interaction",
		Commands: []*ufcli.Command{
			git.NewGitCmd(),
			terraform.NewTerraformCmd(),
			explain.NewExplainCmd(),
			env.NewEnvCmd(),
			clean.NewCleanCmd(),
			whoami.NewWhoamiCmd(),
			uuid.NewUUIDCmd(),
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
