package eksupgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ufcli "github.com/urfave/cli/v2"
)

// NewEKSUpgradeCmd creates the eks upgrade command.
func NewEKSUpgradeCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "eks",
		Usage: "EKS upgrade helper - discover clusters and diff against target config",
		Subcommands: []*ufcli.Command{
			NewEKSUpgradeDiffCmd(),
		},
	}
}

// NewEKSUpgradeDiffCmd creates the diff subcommand.
func NewEKSUpgradeDiffCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "upgrade",
		Usage: "Discover EKS clusters and diff current state vs target config",
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to eks-upgrade-config.yaml",
				Value:   "eks-upgrade-config.yaml",
			},
			&ufcli.StringFlag{
				Name:    "infra-path",
				Aliases: []string{"i"},
				Usage:   "Path to infrastructure-terraform repo",
				Value:   "../infrastructure-terraform",
			},
			&ufcli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Output as JSON (for AI/automation)",
			},
			&ufcli.IntFlag{
				Name:    "step",
				Aliases: []string{"s"},
				Usage:   "Run only step 1, 2, or 3 (omit for full diff)",
				Value:   0,
			},
		},
		Action: runDiff,
	}
}

func runDiff(c *ufcli.Context) error {
	configPath := c.String("config")
	infraPath := c.String("infra-path")
	outputJSON := c.Bool("json")
	step := c.Int("step")

	// Resolve paths relative to cwd
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	configPath = resolvePath(cwd, configPath)
	infraPath = resolvePath(cwd, infraPath)

	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	clusters, err := DiscoverClusters(infraPath)
	if err != nil {
		return fmt.Errorf("discover clusters: %w", err)
	}

	var results []*DiffResult
	for _, clusterPath := range clusters {
		state, err := ParseClusterDir(infraPath, clusterPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skip %s: %v\n", clusterPath, err)
			continue
		}
		if !cfg.InScope(state.ClusterName) {
			continue
		}
		var dr *DiffResult
		if step >= 1 && step <= 3 {
			dr = ComputeDiffForStep(cfg, state, step)
		} else {
			dr = ComputeDiff(cfg, state)
		}
		results = append(results, dr)
	}

	if outputJSON {
		out, err := DiffToJSON(results)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	}

	// Human-readable output
	header := fmt.Sprintf("EKS Upgrade Diff (target: %s)", cfg.Target.ClusterVersion)
	if step >= 1 && step <= 3 {
		header = fmt.Sprintf("EKS Upgrade Step %d: %s", step, GetStepDescription(cfg, step))
	}
	fmt.Printf("%s\n", header)
	fmt.Printf("Config: %s | Infra: %s\n\n", configPath, infraPath)
	for _, dr := range results {
		fmt.Println(FormatDiff(dr))
		// Step 2: output local.tf Bottlerocket change for clusters that need it
		if step == 2 && dr.LocalTFBottlerocket && cfg.Steps.Step2.LocalTFBottlerocket != nil {
			fmt.Println("  --- In local.tf self_managed_node_groups, replace ---")
			fmt.Printf("    %s\n", cfg.Steps.Step2.LocalTFBottlerocket.Find)
			fmt.Println("  --- with ---")
			fmt.Printf("    %s\n", cfg.Steps.Step2.LocalTFBottlerocket.Replace)
			fmt.Println()
		}
		// Step 3: output eks.tf block content for clusters that need it
		if step == 3 && len(dr.EksTFBlocksNeeded) > 0 && len(cfg.Steps.Step3.EksTFBlocks) > 0 {
			fmt.Println("  --- Add these blocks to kubernetes_addons module in eks.tf ---")
			for _, blockName := range dr.EksTFBlocksNeeded {
				var content string
				if blockName == "external_dns_config" {
					content = cfg.GetExternalDNSConfigBlockContent(dr.ClusterName)
				} else if block, ok := cfg.Steps.Step3.EksTFBlocks[blockName]; ok {
					content = block.Content
				}
				if content != "" {
					fmt.Printf("  %s:\n%s\n", blockName, indentBlock(content))
				}
			}
			fmt.Println()
		} else {
			fmt.Println()
		}
	}
	if step == 3 && cfg.Steps.Step3.EKSTFNote != "" {
		fmt.Printf("Note: %s\n", cfg.Steps.Step3.EKSTFNote)
	}
	return nil
}

func resolvePath(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(base, p)
}

// indentBlock adds 4 spaces to each line for display.
func indentBlock(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "    " + line
	}
	return strings.Join(lines, "\n")
}
