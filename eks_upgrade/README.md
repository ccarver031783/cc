# eks_upgrade

A Go package that helps plan and execute EKS cluster upgrades when your infrastructure lives in Terraform. It discovers EKS cluster directories, parses their current state (cluster version, addons, module versions), and diffs them against a target configuration—organized into steps so you can apply changes incrementally across many clusters.

## What It Does

1. **Discovers** EKS cluster directories under an infrastructure repo (looks for `terraform.tfvars` + `eks.tf` under `aws/`).
2. **Parses** each cluster’s current state from `terraform.tfvars`, `eks.tf`, `local.tf`, and `versions.tf` (cluster version, EKS addons, EKS module version, AWS provider version, Bottlerocket config, etc.).
3. **Diffs** current state vs. a YAML config that defines target versions and step definitions.
4. **Outputs** human-readable diffs or JSON for automation, scoped by upgrade step (1, 2, or 3).

## Step-Based Workflow

Upgrades are split into steps so you can apply changes in controlled phases:

- **Step 1**: First-wave addons (e.g. coredns, vpc-cni, metrics-server, external-dns) with optional overrides.
- **Step 2**: EKS control plane version, EKS Terraform module version, AWS provider version, and Bottlerocket `platform` → `ami_type` migration.
- **Step 3**: Remaining addons plus `eks.tf` configuration blocks (e.g. `old_cluster`, `external_dns_config`, `metrics_server_addons_config`).

You run the diff per step, apply the changes, then move to the next step.

## Usage

This package provides CLI commands via [urfave/cli](https://github.com/urfave/cli). Wire it into your CLI:

```go
import "github.com/christopher.carver/cc/eks_upgrade"

app.Commands = append(app.Commands, eksupgrade.NewEKSUpgradeCmd())
```

Then:

```bash
cc eks upgrade                    # Full diff (all steps)
cc eks upgrade --step 1           # Step 1 only
cc eks upgrade --step 2           # Step 2 only
cc eks upgrade --step 3           # Step 3 only
cc eks upgrade --config my.yaml   # Custom config path
cc eks upgrade --infra-path /path/to/terraform
cc eks upgrade --json             # JSON output for automation
```

## Configuration

Copy `eks-upgrade-config.example.yaml` to `eks-upgrade-config.yaml` and customize:

| Section | Purpose |
|---------|---------|
| `target` | Target versions (cluster, EKS module, AWS provider, addons). `eks_module_pattern` identifies your EKS Terraform module source. |
| `externally_managed_clusters` | Clusters managed by external tooling (e.g. GitOps)—shown as `[Externally managed]` in output. |
| `scope` | Limit to specific cluster names (empty = all discovered). |
| `steps` | Per-step definitions: which addons, which config blocks, Bottlerocket find/replace, etc. |

See the example file for the full schema.

## Library Usage

You can use the package programmatically:

```go
cfg, _ := eksupgrade.LoadConfig("eks-upgrade-config.yaml")
clusters, _ := eksupgrade.DiscoverClusters("/path/to/infra")
for _, path := range clusters {
    state, _ := eksupgrade.ParseClusterDir(cfg, "/path/to/infra", path)
    if cfg.InScope(state.ClusterName) {
        dr := eksupgrade.ComputeDiffForStep(cfg, state, 1)
        // ...
    }
}
```

## Relationship to `cc/eksupgrade`

This directory (`eks_upgrade`) is the **sanitized, generic** version intended for open-source or reuse outside a single organization. The cc CLI uses `cc/eksupgrade`, which contains org-specific logic and naming.

- **`cc/eksupgrade`**: Internal version, wired into the cc CLI, org-specific.
- **`cc/eks_upgrade`**: Sanitized copy with generic config fields, no hardcoded org references.

When you change `eksupgrade`, copy the logic into `eks_upgrade` and apply the sanitization rules (see `SYNC_NOTES.md`).

## Requirements

- Go 1.21+
- Terraform layout: cluster dirs with `terraform.tfvars`, `eks.tf`, `versions.tf` (and optionally `local.tf`)
- Paths under `aws/` in the infra repo
