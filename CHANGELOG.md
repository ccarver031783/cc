# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

_No unreleased changes_

---

## [1.0.0] - 2025-12-30

### Added
- `.mise.toml` configuration for dependency management via [mise](https://mise.jdx.dev/)
- Mise tasks for common operations: `build`, `test`, `install`, `fmt`, `lint`, `run`, `clean`
- `golangci-lint` as a managed dependency for code linting
- `CHANGELOG.md` for tracking version history
- GitHub Actions workflow for automatic version tagging on push to main
- `--version` flag to display current version

### Changed
- Dependency management moved from Homebrew to mise
- Updated README with mise-based installation and usage instructions
- Tool dependencies now project-scoped rather than system-wide

### Removed
- `cc setup` command - replaced by `mise install`
- Homebrew dependency for CLI tool management

### Deprecated
- `internal/setup/setup.go` - kept for historical reference, no longer used

---

## [0.1.0] - 2025-12-01

### Added
- Initial release of cc CLI tool
- `cc git` commands: branch, rebase, clean, status
- `cc terraform` commands: fmt, scan, validate, pre-push, init-dir, new
- `cc explain tf` command with Claude API and Ollama support
- `cc setup` command for Homebrew package management
- Pre-push git hook installation
- Terraform module scaffolding templates (AWS, Azure, GCP)

---

## Version Summary

| Version | Date | Highlights |
|---------|------|------------|
| 1.0.0 | 2025-12-30 | Mise integration, removed Homebrew dependency |
| 0.1.0 | 2025-12-01 | Initial release with git, terraform, explain commands |
