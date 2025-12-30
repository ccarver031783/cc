# cc CLI Roadmap

This document tracks planned features for the cc CLI tool.

## Planned Features

### Phase 1 - Essential Tools

| Feature | Command | Description | Status |
|---------|---------|-------------|--------|
| **UUID** | `cc uuid` | Generate UUID v4 | 🔜 Planned |
| **Base64** | `cc b64 <encode/decode>` | Base64 encode/decode (k8s secrets, tokens) | 🔜 Planned |
| **Epoch** | `cc epoch [timestamp]` | Convert epoch ↔ human readable timestamps | 🔜 Planned |
| **IP** | `cc ip` | Show internal and external IP addresses | 🔜 Planned |
| **JWT** | `cc jwt <token>` | Decode and display JWT payload | 🔜 Planned |
| **Ports** | `cc ports` | Show what's listening on common dev ports | 🔜 Planned |

### Phase 2 - Core SRE & Git Features

| Feature | Command | Description | Status |
|---------|---------|-------------|--------|
| **Ship** | `cc git ship` | Safe commit workflow: verify branch, show changes, prompt for message, push | 🔜 Planned |
| **Recent Branches** | `cc git recent` | Show your recent branches (last 5-10) | 🔜 Planned |
| **PR** | `cc pr` | Quick `gh pr create` with smart defaults | 🔜 Planned |
| **SSL Check** | `cc ssl <domain>` | Check SSL certificate expiry date | 🔜 Planned |
| **Open** | `cc open` | Open contextual URLs (GitHub repo, PR, Datadog) based on current dir | 🔜 Planned |
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

### Phase 1 - Essential Tools:

### `cc uuid`
Generate UUIDs:
```bash
cc uuid           # Generate single UUID v4
cc uuid -n 5      # Generate 5 UUIDs
```

### `cc b64`
Base64 encode/decode:
```bash
cc b64 encode "hello world"      # Output: aGVsbG8gd29ybGQ=
cc b64 decode "aGVsbG8gd29ybGQ=" # Output: hello world
cat secret.yaml | cc b64 decode  # Pipe support
```

### `cc epoch`
Timestamp conversion:
```bash
cc epoch                    # Current time as epoch
cc epoch 1703980800         # Convert epoch to human readable
cc epoch "2024-12-31 00:00" # Convert human readable to epoch
```

### `cc ip`
Show IP addresses:
```bash
cc ip
# Internal: 192.168.1.100
# External: 203.0.113.50
```

### `cc jwt`
Decode JWT tokens:
```bash
cc jwt eyJhbGciOiJIUzI1NiIs...
# Header:  {"alg": "HS256", "typ": "JWT"}
# Payload: {"sub": "1234", "name": "John", "exp": 1703980800}
# Expires: 2024-12-31 00:00:00 (in 2 days)
```

### `cc ports`
Show listening ports:
```bash
cc ports
# :3000  node (pid 1234)
# :5432  postgres (pid 5678)
# :8080  go (pid 9012)
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

### `cc git recent`
Show recent branches:
```bash
cc git recent
# 1. feature/add-whoami (2 hours ago)
# 2. fix/login-bug (yesterday)
# 3. feature/dashboard (3 days ago)
```

### `cc pr`
Quick PR creation:
```bash
cc pr                       # Create PR with smart defaults
cc pr -d                    # Create draft PR
cc pr --base develop        # Specify base branch
```

### `cc ssl`
Check SSL certificate:
```bash
cc ssl example.com
# Issuer:  Let's Encrypt
# Expires: 2024-03-15 (in 45 days)
# Status:  ✓ Valid
```

### `cc open`
Open contextual URLs:
```bash
cc open              # Open GitHub repo for current directory
cc open pr           # Open current branch's PR
cc open actions      # Open GitHub Actions
cc open datadog      # Open Datadog dashboard (if configured)
```

---

## Implementation Priority

1. **Phase 1** - Essential Tools: `uuid`, `b64`, `epoch`, `ip`, `jwt`, `ports`
2. **Phase 2** - Core SRE & Git: `git ship`, `git recent`, `pr`, `ssl`, `open`, `k8s`, `argocd`, `cost`, `logs`

## Completed

| Version | Command | Description |
|---------|---------|-------------|
| v1.1.0 | `cc whoami` | Git user, AWS identity, K8s context |
| v1.0.0 | `cc env` | Environment context display |
| v1.0.0 | `cc clean` | Docker, Terraform, Go cache cleanup |

---

## Contributing

Want to help implement a feature? Check the [CHANGELOG.md](CHANGELOG.md) for recent changes and the [README.md](README.md) for development setup.

