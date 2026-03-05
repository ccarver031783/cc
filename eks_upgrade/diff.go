package eksupgrade

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DiffResult represents the changes needed for a single cluster.
type DiffResult struct {
	ClusterName         string                 `json:"cluster_name"`
	Path                string                 `json:"path"`
	ExternallyManaged   bool                   `json:"externally_managed"`
	NeedsUpgrade        bool                   `json:"needs_upgrade"`
	ClusterVersion      VersionDiff            `json:"cluster_version,omitempty"`
	EKSModuleVersion    VersionDiff            `json:"eks_module_version,omitempty"`
	AWSProvider         VersionDiff            `json:"aws_provider,omitempty"`
	Addons              map[string]VersionDiff `json:"addons,omitempty"`
	EksTFBlocksNeeded   []string               `json:"eks_tf_blocks_needed,omitempty"`
	LocalTFBottlerocket bool                   `json:"local_tf_bottlerocket,omitempty"`
}

// VersionDiff holds current and target versions.
type VersionDiff struct {
	Current string `json:"current"`
	Target  string `json:"target"`
}

// ComputeDiff compares cluster state against target config and returns changes needed.
func ComputeDiff(cfg *Config, state *ClusterState) *DiffResult {
	dr := &DiffResult{
		ClusterName:       state.ClusterName,
		Path:              state.Path,
		ExternallyManaged: cfg.IsExternallyManaged(state.ClusterName),
		Addons:            make(map[string]VersionDiff),
	}

	if state.ClusterVersion != cfg.Target.ClusterVersion {
		dr.NeedsUpgrade = true
		dr.ClusterVersion = VersionDiff{Current: state.ClusterVersion, Target: cfg.Target.ClusterVersion}
	}

	if state.EKSModuleVersion != "" && state.EKSModuleVersion != cfg.Target.EKSModuleVersion {
		dr.NeedsUpgrade = true
		dr.EKSModuleVersion = VersionDiff{Current: state.EKSModuleVersion, Target: cfg.Target.EKSModuleVersion}
	}

	if state.AWSProviderVersion != "" && !providerVersionMatches(state.AWSProviderVersion, cfg.Target.AWSProviderVersion) {
		dr.NeedsUpgrade = true
		dr.AWSProvider = VersionDiff{Current: state.AWSProviderVersion, Target: cfg.Target.AWSProviderVersion}
	}

	for addon, targetVer := range cfg.Target.EKSAddonsVersions {
		current := state.Addons[addon]
		if current != targetVer {
			dr.NeedsUpgrade = true
			dr.Addons[addon] = VersionDiff{Current: current, Target: targetVer}
		}
	}

	return dr
}

func providerVersionMatches(current, target string) bool {
	if current == target {
		return true
	}
	c := extractVersionPrefix(current)
	t := extractVersionPrefix(target)
	return c == t
}

func extractVersionPrefix(v string) string {
	v = strings.TrimPrefix(v, "~> ")
	v = strings.TrimPrefix(v, ">= ")
	v = strings.TrimPrefix(v, "> ")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

// FormatDiff returns a human-readable summary of the diff.
func FormatDiff(dr *DiffResult) string {
	if !dr.NeedsUpgrade {
		return fmt.Sprintf("%s: up to date", dr.ClusterName)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s (%s)\n", dr.ClusterName, dr.Path))
	if dr.ExternallyManaged {
		b.WriteString("  [Externally managed]\n")
	}
	if dr.ClusterVersion.Target != "" {
		b.WriteString(fmt.Sprintf("  cluster_version: %s -> %s\n", dr.ClusterVersion.Current, dr.ClusterVersion.Target))
	}
	if dr.EKSModuleVersion.Target != "" {
		b.WriteString(fmt.Sprintf("  eks_module: %s -> %s\n", dr.EKSModuleVersion.Current, dr.EKSModuleVersion.Target))
	}
	if dr.AWSProvider.Target != "" {
		b.WriteString(fmt.Sprintf("  aws_provider: %s -> %s\n", dr.AWSProvider.Current, dr.AWSProvider.Target))
	}
	for addon, vd := range dr.Addons {
		if vd.Target != "" {
			b.WriteString(fmt.Sprintf("  addon %s: %s -> %s\n", addon, vd.Current, vd.Target))
		}
	}
	if len(dr.EksTFBlocksNeeded) > 0 {
		b.WriteString(fmt.Sprintf("  eks.tf blocks needed: %v\n", dr.EksTFBlocksNeeded))
	}
	if dr.LocalTFBottlerocket {
		b.WriteString("  local.tf: replace platform = \"bottlerocket\" with ami_type = \"BOTTLEROCKET_x86_64\"\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// DiffToJSON outputs the full diff as JSON for AI/automation consumption.
func DiffToJSON(results []*DiffResult) (string, error) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
