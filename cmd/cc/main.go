package main

import (
	"context"
	"fmt"
	"os"

	"github.com/christopher.carver/cc/internal/explain"
	"github.com/christopher.carver/cc/internal/git"
	"github.com/christopher.carver/cc/internal/setup"
	"github.com/christopher.carver/cc/internal/terraform"
	ufcli "github.com/urfave/cli/v2"
)

func main() {
	ctx := context.Background()

	app := &ufcli.App{
		Name:  "cc",
		Usage: "Development and SRE-based CLI tooling - turning cc commands into shortcuts for git and terraform interaction ",
		Commands: []*ufcli.Command{
			setup.NewSetupCmd(),
			git.NewGitCmd(),
			terraform.NewTerraformCmd(),
			explain.NewExplainCmd(),
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
