package eksupgrade

// ComputeDiffForStep returns a DiffResult filtered to only include changes for the given step.
func ComputeDiffForStep(cfg *Config, state *ClusterState, step int) *DiffResult {
	full := ComputeDiff(cfg, state)
	filtered := &DiffResult{
		ClusterName:    full.ClusterName,
		Path:           full.Path,
		StencilManaged: full.StencilManaged,
		Addons:         make(map[string]VersionDiff),
	}

	switch step {
	case 1:
		addons := cfg.GetStepAddons(1, state.ClusterName)
		for _, addon := range addons {
			targetVer := cfg.GetStep1AddonVersion(state.ClusterName, addon)
			current := state.Addons[addon]
			if current != targetVer {
				filtered.NeedsUpgrade = true
				filtered.Addons[addon] = VersionDiff{Current: current, Target: targetVer}
			}
		}
	case 2:
		if cfg.Steps.Step2.ClusterVersion && full.ClusterVersion.Target != "" {
			filtered.NeedsUpgrade = true
			filtered.ClusterVersion = full.ClusterVersion
		}
		if cfg.Steps.Step2.EKSModuleVersion && full.EKSModuleVersion.Target != "" {
			filtered.NeedsUpgrade = true
			filtered.EKSModuleVersion = full.EKSModuleVersion
		}
		if cfg.Steps.Step2.AWSProviderVersion && full.AWSProvider.Target != "" {
			filtered.NeedsUpgrade = true
			filtered.AWSProvider = full.AWSProvider
		}
		if state.HasPlatformBottlerocket {
			filtered.NeedsUpgrade = true
			filtered.LocalTFBottlerocket = true
		}
	case 3:
		addons := cfg.GetStepAddons(3, state.ClusterName)
		for _, addon := range addons {
			targetVer := cfg.Target.EKSAddonsVersions[addon]
			current := state.Addons[addon]
			if current != targetVer {
				filtered.NeedsUpgrade = true
				filtered.Addons[addon] = VersionDiff{Current: current, Target: targetVer}
			}
		}
		// Step 3: determine which eks.tf blocks this cluster needs
		if len(cfg.Steps.Step3.EksTFBlocks) > 0 {
			var blocksNeeded []string
			if !state.HasOldCluster {
				if _, ok := cfg.Steps.Step3.EksTFBlocks["old_cluster"]; ok {
					blocksNeeded = append(blocksNeeded, "old_cluster")
				}
			}
			if !state.HasNewExternalDNSConfig {
				if _, ok := cfg.Steps.Step3.EksTFBlocks["external_dns_config"]; ok {
					blocksNeeded = append(blocksNeeded, "external_dns_config")
				}
			}
			if !state.HasMetricsServerAddonsConfig {
				if _, ok := cfg.Steps.Step3.EksTFBlocks["metrics_server_addons_config"]; ok {
					blocksNeeded = append(blocksNeeded, "metrics_server_addons_config")
				}
			}
			if !state.HasEbsCsiSnapshotter512Mi && cfg.ShouldAddEbsCsiSnapshotter512Mi(state.ClusterName) {
				if _, ok := cfg.Steps.Step3.EksTFBlocks["ebs_csi_snapshotter_512mi"]; ok {
					blocksNeeded = append(blocksNeeded, "ebs_csi_snapshotter_512mi")
				}
			}
			if len(blocksNeeded) > 0 {
				filtered.NeedsUpgrade = true
				filtered.EksTFBlocksNeeded = blocksNeeded
			}
		}
	default:
		return full
	}

	return filtered
}

// GetStepDescription returns the description for a step.
func GetStepDescription(cfg *Config, step int) string {
	switch step {
	case 1:
		return cfg.Steps.Step1.Description
	case 2:
		return cfg.Steps.Step2.Description
	case 3:
		return cfg.Steps.Step3.Description
	default:
		return ""
	}
}
