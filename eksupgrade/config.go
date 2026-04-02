package eksupgrade

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the EKS upgrade configuration loaded from YAML.
type Config struct {
	Target          TargetConfig `yaml:"target"`
	Steps           StepsConfig  `yaml:"steps"`
	StencilClusters []string     `yaml:"stencil_clusters"`
	Scope           []string     `yaml:"scope"`
}

// StepsConfig defines what changes belong to each step.
type StepsConfig struct {
	Step1 StepDef `yaml:"step1"`
	Step2 StepDef `yaml:"step2"`
	Step3 StepDef `yaml:"step3"`
}

// StepDef defines a single upgrade step.
type StepDef struct {
	Description         string                       `yaml:"description"`
	Addons              []string                     `yaml:"addons"`
	ClusterVersion      bool                         `yaml:"cluster_version"`
	EKSModuleVersion    bool                         `yaml:"eks_module_version"`
	AWSProviderVersion  bool                         `yaml:"aws_provider_version"`
	ClusterOverrides    map[string]map[string]string `yaml:"cluster_overrides"`
	AddonOverrides      map[string]string            `yaml:"addon_overrides"` // Step 1: apply to ALL clusters
	AddonsForToolingDev []string                     `yaml:"addons_for_tooling_dev"`
	EKSTFNote                           string                   `yaml:"eks_tf_note"`
	EksTFBlocks                         map[string]EksTFBlockDef `yaml:"eks_tf_blocks"`
	LocalTFBottlerocket                 *LocalTFBottlerocketDef  `yaml:"local_tf_bottlerocket"` // Step 2: platform -> ami_type
	ExternalDNSExcludeIstioGatewayClusters []string               `yaml:"external_dns_exclude_istio_gateway_clusters"` // Cluster name prefixes (e.g. "labs-") that should use sources without istio-gateway
	EbsCsiSnapshotter512MiClusters         []string               `yaml:"ebs_csi_snapshotter_512mi_clusters"`         // Cluster name prefixes (e.g. "labs-") that need csi-snapshotter 512Mi
}

// LocalTFBottlerocketDef describes the local.tf Bottlerocket change for step 2.
type LocalTFBottlerocketDef struct {
	Find    string `yaml:"find"`    // e.g. platform = "bottlerocket"
	Replace string `yaml:"replace"` // e.g. ami_type = "BOTTLEROCKET_x86_64"
}

// EksTFBlockDef holds the HCL content for a block to add to eks.tf.
type EksTFBlockDef struct {
	Content string `yaml:"content"`
}

// TargetConfig holds the target versions for an upgrade.
type TargetConfig struct {
	ClusterVersion     string            `yaml:"cluster_version"`
	EKSModuleVersion   string            `yaml:"eks_module_version"`
	AWSProviderVersion string            `yaml:"aws_provider_version"`
	EKSAddonsVersions  map[string]string `yaml:"eks_addons_versions"`
}

// LoadConfig reads and parses the config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// IsStencilCluster returns true if the cluster name is in the Stencil-managed list.
func (c *Config) IsStencilCluster(clusterName string) bool {
	for _, name := range c.StencilClusters {
		if name == clusterName {
			return true
		}
	}
	return false
}

// InScope returns true if the cluster is in scope (or scope is empty = all).
func (c *Config) InScope(clusterName string) bool {
	if len(c.Scope) == 0 {
		return true
	}
	for _, name := range c.Scope {
		if name == clusterName {
			return true
		}
	}
	return false
}

// GetStepAddons returns the addons for a given step and cluster.
// tooling-dev-useast1-green has special handling for step 1 (coredns override) and step 3 (adds coredns).
func (c *Config) GetStepAddons(step int, clusterName string) []string {
	switch step {
	case 1:
		return c.Steps.Step1.Addons
	case 2:
		return nil // step 2 has no addons
	case 3:
		if clusterName == "tooling-dev-useast1-green" && len(c.Steps.Step3.AddonsForToolingDev) > 0 {
			return c.Steps.Step3.AddonsForToolingDev
		}
		return c.Steps.Step3.Addons
	default:
		return nil
	}
}

// ShouldExcludeIstioGatewayFromExternalDNS returns true if the cluster should use
// external_dns sources without istio-gateway (e.g. labs clusters that don't have Istio).
func (c *Config) ShouldExcludeIstioGatewayFromExternalDNS(clusterName string) bool {
	for _, prefix := range c.Steps.Step3.ExternalDNSExcludeIstioGatewayClusters {
		if len(prefix) > 0 && len(clusterName) >= len(prefix) && clusterName[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// ShouldAddEbsCsiSnapshotter512Mi returns true if the cluster needs the ebs-csi snapshotter 512Mi memory limit.
func (c *Config) ShouldAddEbsCsiSnapshotter512Mi(clusterName string) bool {
	for _, prefix := range c.Steps.Step3.EbsCsiSnapshotter512MiClusters {
		if len(prefix) > 0 && len(clusterName) >= len(prefix) && clusterName[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// GetExternalDNSConfigBlockContent returns the eks.tf block content for external_dns_config,
// choosing the no-istio variant for clusters that don't have Istio (e.g. labs).
func (c *Config) GetExternalDNSConfigBlockContent(clusterName string) string {
	if c.ShouldExcludeIstioGatewayFromExternalDNS(clusterName) {
		if block, ok := c.Steps.Step3.EksTFBlocks["external_dns_config_no_istio"]; ok && block.Content != "" {
			return block.Content
		}
	}
	if block, ok := c.Steps.Step3.EksTFBlocks["external_dns_config"]; ok {
		return block.Content
	}
	return ""
}

// GetStep1AddonVersion returns the target version for an addon in step 1.
// Checks addon_overrides (all clusters) first, then cluster_overrides (per-cluster), then target.
func (c *Config) GetStep1AddonVersion(clusterName, addon string) string {
	if v, ok := c.Steps.Step1.AddonOverrides[addon]; ok {
		return v
	}
	if overrides, ok := c.Steps.Step1.ClusterOverrides[clusterName]; ok {
		if v, ok := overrides[addon]; ok {
			return v
		}
	}
	return c.Target.EKSAddonsVersions[addon]
}
