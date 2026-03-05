# eks_upgrade – Sanitized / Open-Source Ready Copy

This directory is the **generic, sanitized** version of `cc/eksupgrade/`. It is intended for future open-source release and is **not** wired into the cc CLI (which uses `cc/eksupgrade`).

## Workflow: Syncing Changes

When you refine `cc/eksupgrade/` (the internal, org-specific version):

1. **Copy** the updated logic from `eksupgrade/` into `eks_upgrade/`.
2. **Apply sanitization** as you copy:
   - `StencilClusters` → `ExternallyManagedClusters`
   - `IsStencilCluster` → `IsExternallyManaged`
   - `StencilManaged` → `ExternallyManaged`
   - `AddonsForToolingDev` → `AddonsForSpecialClusters` (map)
   - Hardcoded cluster names → config-driven
   - `terraform-modules__udemy/eks` → `cfg.Target.EKSModulePattern`
   - `txtOwnerId` → `cfg.Steps.Step3.ExternalDNSConfigMarkers`
3. **Update** `eks-upgrade-config.example.yaml` if the config schema changes.
4. **Verify** with `go build ./cc/eks_upgrade/...`

## Key Differences from eksupgrade

| eksupgrade (internal) | eks_upgrade (sanitized) |
|-----------------------|-------------------------|
| `ParseClusterDir(infraRoot, clusterPath)` | `ParseClusterDir(cfg, infraRoot, clusterPath)` |
| Hardcoded module pattern | `target.eks_module_pattern` in config |
| Hardcoded external_dns marker | `steps.step3.external_dns_config_markers` |
| `StencilClusters` / `StencilManaged` | `ExternallyManagedClusters` / `ExternallyManaged` |
| `AddonsForToolingDev` (single cluster) | `AddonsForSpecialClusters` (map) |

## Using This Package

To wire this into a CLI or use as a library:

```go
import "github.com/christopher.carver/cc/eks_upgrade"
```

The config YAML schema uses the generic field names above; see `eks-upgrade-config.example.yaml`.
