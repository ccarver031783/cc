# cc CLI Roadmap

This document tracks planned features for the cc CLI tool.

## Planned Features

### Phase 1 - General Use Tools

| Feature | Command | Description | Status |
|---------|---------|-------------|--------|
| **Who Am I** | `cc whoami` | Display current git user, AWS identity, k8s context in one command | 🔜 Planned Jan 2026 |

### Phase 2 - Additional Core Features

| Feature | Command | Description | Status |
|---------|---------|-------------|--------|
| **Ship** | `cc git ship` | Safe commit workflow: verify branch, show changes, prompt for message, push | 🔜 Planned |
| **Cost Estimation** | `cc cost` | Terraform cost estimation via [Infracost](https://www.infracost.io/) integration | 🔜 Planned |
| **Kubernetes** | `cc k8s` | Context switching, log tailing, pod exec, restart deployments | 🔜 Planned |
| **Logs** | `cc logs` | Unified log viewer for k8s pods, CloudWatch, local files | 🔜 Planned |
| **ArgoCD** | `cc argocd` | Sync status, app health, quick sync triggers | 🔜 Planned |

### Deprioritized

| Feature | Command | Description | Reason |
|---------|---------|-------------|--------|
| **Secrets** | `cc secrets` | AWS Secrets Manager lookup/list | Scale concerns ( some teams manage hundreds of secrets across multiple cloud providers and accounts) |

---

## Feature Details

### Phase 1 Features:

### `cc whoami`
Display identity information across tools:
```bash
cc whoami
# Output:
# Git:     christopher.carver <email>
# AWS:     arn:aws:iam::123456789:user/christopher.carver (account: prod)
# K8s:     main-prod-useast1 (namespace: default)
```
### Phase 2 Features:

### `cc cost`
Terraform cost estimation:
```bash
cc cost             # Estimate cost for current directory
cc cost diff        # Show cost difference vs main branch
cc cost --format json  # Output as JSON for CI
```

### `cc git ship`
Safe add, commit, and push workflow:
```bash
cc git ship
# 1. Verifies not on main/master branch
# 2. Shows staged changes
# 3. Prompts for the commit message for all commits
# 4. Confirms changes before push
# 5. Execution summary: git add . && git commit -m "Message" && git push
```

### `cc k8s`
Kubernetes shortcuts:
```bash
cc k8s context      # Show/switch context
cc k8s logs <pod>   # Tail pod logs
cc k8s exec <pod>   # Exec into pod
cc k8s restart <deployment>  # Restart deployment
```

### `cc logs`
Unified log viewing:
```bash
cc logs k8s <pod>           # Tail k8s pod logs
cc logs cloudwatch <group>  # Tail CloudWatch log group
cc logs file <path>         # Tail local file
```

### `cc argocd`
ArgoCD operations:
```bash
cc argocd status            # Show all app statuses
cc argocd sync <app>        # Trigger sync
cc argocd diff <app>        # Show pending changes
```

---

## Implementation Priority

1. **Phase 1** - General Use Tools: `whoami`
2. **Phase 2** - Core SRE Tools: `k8s`, `argocd`, `cost`, `logs`, `git ship`

---

## Contributing

Want to help implement a feature? Check the [CHANGELOG.md](CHANGELOG.md) for recent changes and the [README.md](README.md) for development setup.

