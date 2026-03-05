package eksupgrade

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ClusterState holds the current versions parsed from a cluster directory.
type ClusterState struct {
	ClusterName        string
	Path               string
	ClusterVersion     string
	EKSModuleVersion   string
	AWSProviderVersion string
	Addons             map[string]string
	// EKS add-on block status (for step 3)
	HasOldCluster                bool // old_cluster = false in kubernetes_addons
	HasNewExternalDNSConfig       bool // external_dns_config with configuration_values
	HasMetricsServerAddonsConfig  bool // metrics_server_addons_config block
	// local.tf Bottlerocket config (for step 2)
	HasPlatformBottlerocket bool // platform = "bottlerocket" - needs change to ami_type
}

var (
	clusterVersionRe     = regexp.MustCompile(`cluster_version\s*=\s*"([^"]+)"`)
	addonVersionRe       = regexp.MustCompile(`(\w+(?:-\w+)*)\s*=\s*\{\s*version\s*=\s*"([^"]+)"`)
)

// ParseClusterDir reads terraform.tfvars, eks.tf, and versions.tf from a cluster directory.
// cfg is used for configurable patterns (EKS module source, external_dns markers).
func ParseClusterDir(cfg *Config, infraRoot, clusterPath string) (*ClusterState, error) {
	base := filepath.Join(infraRoot, clusterPath)
	clusterName := filepath.Base(filepath.Dir(clusterPath))

	state := &ClusterState{
		ClusterName: clusterName,
		Path:        clusterPath,
		Addons:      make(map[string]string),
	}

	// Parse terraform.tfvars
	tfvarsPath := filepath.Join(base, "terraform.tfvars")
	tfvars, err := os.ReadFile(tfvarsPath)
	if err != nil {
		return nil, err
	}
	tfvarsStr := string(tfvars)

	if m := clusterVersionRe.FindStringSubmatch(tfvarsStr); len(m) > 1 {
		state.ClusterVersion = m[1]
	}

	// Parse addons from eks_addons_versions block
	if idx := strings.Index(tfvarsStr, "eks_addons_versions"); idx >= 0 {
		block := tfvarsStr[idx:]
		if end := strings.Index(block, "}"); end > 0 {
			block = block[:end]
		}
		for _, m := range addonVersionRe.FindAllStringSubmatch(block, -1) {
			if len(m) > 2 {
				state.Addons[m[1]] = m[2]
			}
		}
	}

	// Parse eks.tf for EKS module version and step 3 block status
	eksPath := filepath.Join(base, "eks.tf")
	eksContent, err := os.ReadFile(eksPath)
	if err == nil {
		eksStr := string(eksContent)
		// EKS module version - use configurable pattern
		modulePattern := cfg.Target.EKSModulePattern
		if modulePattern == "" {
			modulePattern = "eks" // minimal fallback
		}
		if strings.Contains(eksStr, modulePattern) {
			eksModuleRe := regexp.MustCompile(`module "eks_cluster"[\s\S]*?version\s*=\s*"([^"]+)"`)
			if m := eksModuleRe.FindStringSubmatch(eksStr); len(m) > 1 {
				state.EKSModuleVersion = m[1]
			}
		}
		// Step 3: detect kubernetes_addons block status
		state.HasOldCluster = strings.Contains(eksStr, "old_cluster")
		// New external_dns_config: must have block + at least one marker (default: txtOwnerId)
		markers := cfg.Steps.Step3.ExternalDNSConfigMarkers
		if len(markers) == 0 {
			markers = []string{"txtOwnerId"}
		}
		state.HasNewExternalDNSConfig = strings.Contains(eksStr, "external_dns_config")
		if state.HasNewExternalDNSConfig {
			found := false
			for _, m := range markers {
				if strings.Contains(eksStr, m) {
					found = true
					break
				}
			}
			state.HasNewExternalDNSConfig = found
		}
		state.HasMetricsServerAddonsConfig = strings.Contains(eksStr, "metrics_server_addons_config")
	}

	// Parse local.tf for Bottlerocket config (step 2)
	localPath := filepath.Join(base, "local.tf")
	localContent, err := os.ReadFile(localPath)
	if err == nil {
		localStr := string(localContent)
		state.HasPlatformBottlerocket = strings.Contains(localStr, `platform`) &&
			strings.Contains(strings.ToLower(localStr), `bottlerocket`)
	}

	// Parse versions.tf for AWS provider
	versionsPath := filepath.Join(base, "versions.tf")
	versionsContent, err := os.ReadFile(versionsPath)
	if err == nil {
		versionsStr := string(versionsContent)
		awsBlockRe := regexp.MustCompile(`aws\s*=\s*\{[\s\S]*?version\s*=\s*"([^"]+)"`)
		if m := awsBlockRe.FindStringSubmatch(versionsStr); len(m) > 1 {
			state.AWSProviderVersion = strings.TrimSpace(m[1])
		}
	}

	return state, nil
}
