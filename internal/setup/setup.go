package setup

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/christopher.carver/cc/internal/shell"
	ufcli "github.com/urfave/cli/v2"
)

// PackageType represents whether a package is a formula or cask
type PackageType string

const (
	PackageTypeFormula PackageType = "formula"
	PackageTypeCask    PackageType = "cask"
	PackageTypeUnknown PackageType = "unknown"
)

// PackageInfo holds information about a Homebrew package
type PackageInfo struct {
	Name              string
	DisplayName       string
	Type              PackageType
	Installed         bool
	ManuallyInstalled bool
	ManualPath        string
	CurrentVersion    string
	LatestVersion     string
	NeedsUpgrade      bool
	NeedsInstall      bool
	NeedsMigration    bool
}

// RequiredPackages is the list of packages that should be checked
var RequiredPackages = []struct {
	Name        string
	DisplayName string
}{
	// {"cursor", "Cursor"}, // Not available via Homebrew
	{"go", "GoLang"},
	{"sequel-ace", "SequelAce"},
	{"utm", "UTM"},
}

// checkHomebrewInstalled checks if Homebrew is installed
func checkHomebrewInstalled(ctx context.Context) (bool, error) {
	_, err := shell.Run(ctx, "brew", "--version")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// installHomebrew installs Homebrew if it's not already installed
func installHomebrew(ctx context.Context) error {
	fmt.Println("Homebrew is not installed. Installing Homebrew...")
	fmt.Println("This may take a few minutes...")
	fmt.Println()

	// Use the official Homebrew installation script
	// Run the installation script interactively so user can see progress
	// The script URL is passed as an argument to curl
	err := shell.RunInteractive(ctx, "/bin/bash", "-c", "curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash")
	if err != nil {
		return fmt.Errorf("failed to install Homebrew: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Homebrew installed successfully")
	fmt.Println("Note: You may need to add Homebrew to your PATH. Follow the instructions above.")
	return nil
}

// detectPackageType checks if a package exists as a formula or cask
func detectPackageType(ctx context.Context, packageName string) (PackageType, error) {
	// Check if it's a formula
	_, err := shell.Run(ctx, "brew", "info", "--formula", packageName)
	if err == nil {
		return PackageTypeFormula, nil
	}

	// Check if it's a cask
	_, err = shell.Run(ctx, "brew", "info", "--cask", packageName)
	if err == nil {
		return PackageTypeCask, nil
	}

	return PackageTypeUnknown, fmt.Errorf("package %s not found as formula or cask", packageName)
}

// checkPackageInstalled checks if a package is installed
func checkPackageInstalled(ctx context.Context, packageName string, pkgType PackageType) (bool, error) {
	var cmd string
	var args []string

	if pkgType == PackageTypeCask {
		cmd = "brew"
		args = []string{"list", "--cask", packageName}
	} else {
		cmd = "brew"
		args = []string{"list", "--formula", packageName}
	}

	output, err := shell.Run(ctx, cmd, args...)
	if err != nil {
		return false, nil
	}

	return strings.Contains(output, packageName), nil
}

// checkManualInstallation checks if a package is installed outside of Homebrew
func checkManualInstallation(ctx context.Context, packageName string, pkgType PackageType) (bool, string) {
	if pkgType == PackageTypeCask {
		// For casks, check if the app exists in /Applications
		// Common app names to check - convert package name to title case
		appNames := []string{
			packageName,
		}

		// Convert to title case manually (strings.Title is deprecated)
		titleCase := strings.ToUpper(packageName[:1]) + strings.ToLower(packageName[1:])
		appNames = append(appNames, titleCase)

		// Handle hyphenated names like "sequel-ace" -> "Sequel Ace"
		if strings.Contains(packageName, "-") {
			parts := strings.Split(packageName, "-")
			var titleParts []string
			for _, part := range parts {
				titleParts = append(titleParts, strings.ToUpper(part[:1])+strings.ToLower(part[1:]))
			}
			appNames = append(appNames, strings.Join(titleParts, " "))
			appNames = append(appNames, strings.Join(titleParts, ""))
		}

		for _, appName := range appNames {
			appPath := filepath.Join("/Applications", appName+".app")
			if _, err := os.Stat(appPath); err == nil {
				return true, appPath
			}
		}

		// Also check for common variations
		if packageName == "sequel-ace" {
			if _, err := os.Stat("/Applications/Sequel Ace.app"); err == nil {
				return true, "/Applications/Sequel Ace.app"
			}
		}
	} else {
		// For formulas, check if the binary exists in PATH
		// For Go, check if 'go' command exists
		if packageName == "go" {
			if path, err := exec.LookPath("go"); err == nil {
				// Check if it's not in Homebrew's typical paths
				if !strings.Contains(path, "/opt/homebrew") && !strings.Contains(path, "/usr/local") {
					return true, path
				}
				// Even if in homebrew paths, verify it's actually managed by homebrew
				installed, _ := checkPackageInstalled(ctx, packageName, pkgType)
				if !installed {
					return true, path
				}
			}
		} else {
			// For other formulas, check if binary exists
			if path, err := exec.LookPath(packageName); err == nil {
				// Check if it's not in Homebrew's typical paths
				if !strings.Contains(path, "/opt/homebrew") && !strings.Contains(path, "/usr/local") {
					return true, path
				}
				// Verify it's actually managed by homebrew
				installed, _ := checkPackageInstalled(ctx, packageName, pkgType)
				if !installed {
					return true, path
				}
			}
		}
	}

	return false, ""
}

// removeManualInstallation attempts to remove a manually installed package
func removeManualInstallation(ctx context.Context, packageName string, pkgType PackageType, manualPath string) error {
	if pkgType == PackageTypeCask {
		// For GUI apps, remove the .app bundle
		fmt.Printf("  Removing manually installed app: %s\n", manualPath)
		if err := os.RemoveAll(manualPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", manualPath, err)
		}
		// Also try to remove from Applications folder if it's a symlink or alias
		return nil
	} else {
		// For command-line tools, we can't automatically remove them safely
		// as they might be installed system-wide or via other package managers
		fmt.Printf("  Warning: Manual installation detected at %s\n", manualPath)
		fmt.Printf("  Homebrew will install alongside the manual version and take precedence via PATH\n")
		return fmt.Errorf("manual removal required for %s", packageName)
	}
}

// installPackageAlongside installs a package via Homebrew even if a manual installation exists
// This allows Homebrew to "assume control" by ensuring its version takes precedence in PATH
func installPackageAlongside(ctx context.Context, packageName string, pkgType PackageType, manualPath string) error {
	fmt.Printf("  Installing %s via Homebrew (manual installation at %s will remain)\n", packageName, manualPath)

	// Install via Homebrew
	if err := installPackage(ctx, packageName, pkgType); err != nil {
		return err
	}

	// Verify Homebrew's installation takes precedence
	if pkgType == PackageTypeFormula {
		homebrewPaths := []string{"/opt/homebrew/bin", "/usr/local/bin"}
		fmt.Printf("\n  ✓ %s installed via Homebrew\n", packageName)
		fmt.Printf("  Ensure Homebrew's bin directory comes first in your PATH:\n")
		for _, path := range homebrewPaths {
			fmt.Printf("    export PATH=\"%s:$PATH\"\n", path)
		}
		fmt.Printf("\n  Add this to your shell profile (~/.zshrc or ~/.bash_profile) if needed\n")

		// Try to detect which binary will be used
		output, err := shell.Run(ctx, "which", packageName)
		if err == nil {
			output = strings.TrimSpace(output)
			isHomebrew := false
			for _, path := range homebrewPaths {
				if strings.HasPrefix(output, path) {
					isHomebrew = true
					break
				}
			}
			if isHomebrew {
				fmt.Printf("  ✓ Homebrew version is now active: %s\n", output)
			} else {
				fmt.Printf("  ⚠ Manual version still active: %s\n", output)
				fmt.Printf("    You may need to restart your terminal or update your PATH\n")
			}
		}
	}

	return nil
}

// getPackageVersion extracts the installed version from brew info output
func getPackageVersion(ctx context.Context, packageName string, pkgType PackageType) (string, error) {
	var args []string
	if pkgType == PackageTypeCask {
		args = []string{"info", "--cask", packageName}
	} else {
		args = []string{"info", "--formula", packageName}
	}

	output, err := shell.Run(ctx, "brew", args...)
	if err != nil {
		return "", err
	}

	// Parse version from brew info output
	// Format: "package_name: version" or "package_name version"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, packageName+":") || strings.HasPrefix(line, packageName+" ") {
			// Extract version - look for pattern like "1.2.3" or "v1.2.3"
			re := regexp.MustCompile(`(\d+\.\d+\.\d+[^\s]*)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				return matches[1], nil
			}
		}
	}

	// Try to find version in installed packages list
	if pkgType == PackageTypeCask {
		output, err = shell.Run(ctx, "brew", "list", "--cask", "--versions", packageName)
	} else {
		output, err = shell.Run(ctx, "brew", "list", "--formula", "--versions", packageName)
	}

	if err == nil && output != "" {
		parts := strings.Fields(output)
		if len(parts) > 1 {
			return parts[len(parts)-1], nil
		}
	}

	return "unknown", nil
}

// getLatestVersion gets the latest available version from brew info
func getLatestVersion(ctx context.Context, packageName string, pkgType PackageType) (string, error) {
	var args []string
	if pkgType == PackageTypeCask {
		args = []string{"info", "--cask", packageName}
	} else {
		args = []string{"info", "--formula", packageName}
	}

	output, err := shell.Run(ctx, "brew", args...)
	if err != nil {
		return "", err
	}

	// Parse version from brew info output
	// Look for version pattern in the output
	re := regexp.MustCompile(`(\d+\.\d+\.\d+[^\s]*)`)
	matches := re.FindAllString(output, -1)
	if len(matches) > 0 {
		return matches[0], nil
	}

	return "unknown", nil
}

// compareVersions compares two version strings (simple string comparison for now)
func compareVersions(current, latest string) bool {
	if current == "unknown" || latest == "unknown" {
		return false
	}
	return current != latest
}

// upgradePackage upgrades a package with progress indication
func upgradePackage(ctx context.Context, packageName string, pkgType PackageType) error {
	var args []string
	if pkgType == PackageTypeCask {
		args = []string{"upgrade", "--cask", packageName}
	} else {
		args = []string{"upgrade", packageName}
	}

	fmt.Printf("  [%s] Starting upgrade...\n", packageName)

	// Run upgrade with output streaming for progress
	// This will show Homebrew's native progress output
	err := shell.RunInteractive(ctx, "brew", args...)
	if err != nil {
		return fmt.Errorf("failed to upgrade %s: %w", packageName, err)
	}

	fmt.Printf("  [%s] ✓ Upgrade completed\n", packageName)
	return nil
}

// installPackage installs a package with progress indication
func installPackage(ctx context.Context, packageName string, pkgType PackageType) error {
	var args []string
	if pkgType == PackageTypeCask {
		args = []string{"install", "--cask", packageName}
	} else {
		args = []string{"install", packageName}
	}

	fmt.Printf("  [%s] Starting installation...\n", packageName)

	// Run install with output streaming for progress
	// This will show Homebrew's native progress output
	err := shell.RunInteractive(ctx, "brew", args...)
	if err != nil {
		return fmt.Errorf("failed to install %s: %w", packageName, err)
	}

	fmt.Printf("  [%s] ✓ Installation completed\n", packageName)
	return nil
}

// printSummaryTable prints a formatted table of package statuses
func printSummaryTable(packages []PackageInfo) {
	fmt.Println("\n" + strings.Repeat("=", 90))
	fmt.Printf("%-20s %-15s %-15s %-25s %-15s\n", "Package", "Current", "Latest", "Status", "Action")
	fmt.Println(strings.Repeat("-", 90))

	for _, pkg := range packages {
		status := "Not Installed"
		action := "Install"

		if pkg.ManuallyInstalled {
			status = "Manual Install"
			action = "Migrate"
		} else if pkg.Installed {
			if pkg.NeedsUpgrade {
				status = "Update Available"
				action = "Upgrade"
			} else {
				status = "Up to Date"
				action = "None"
			}
		}

		currentVer := pkg.CurrentVersion
		if currentVer == "" {
			currentVer = "-"
		}

		latestVer := pkg.LatestVersion
		if latestVer == "" {
			latestVer = "-"
		}

		fmt.Printf("%-20s %-15s %-15s %-25s %-15s\n",
			pkg.DisplayName, currentVer, latestVer, status, action)
	}
	fmt.Println(strings.Repeat("=", 90) + "\n")
}

// promptConfirmation asks the user for confirmation
func promptConfirmation(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message + " (y/n): ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// NewSetupCmd creates the setup command
func NewSetupCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "setup",
		Usage: "Check and upgrade required Homebrew packages",
		Action: func(c *ufcli.Context) error {
			ctx := c.Context

			fmt.Println("Checking Homebrew installation...")

			// Check if Homebrew is installed
			installed, err := checkHomebrewInstalled(ctx)
			if err != nil {
				return fmt.Errorf("error checking Homebrew: %w", err)
			}

			if !installed {
				confirm, err := promptConfirmation("Homebrew is not installed. Would you like to install it now?")
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}
				if !confirm {
					return fmt.Errorf("Homebrew installation cancelled")
				}

				if err := installHomebrew(ctx); err != nil {
					return err
				}
			} else {
				fmt.Println("✓ Homebrew is installed")
			}

			fmt.Println("\nChecking required packages...")

			// Collect package information
			var packageInfos []PackageInfo
			var packagesToUpgrade []PackageInfo

			for _, reqPkg := range RequiredPackages {
				pkgInfo := PackageInfo{
					Name:        reqPkg.Name,
					DisplayName: reqPkg.DisplayName,
				}

				// Detect package type
				pkgType, err := detectPackageType(ctx, reqPkg.Name)
				if err != nil {
					fmt.Printf("⚠ Warning: Could not detect type for %s: %v\n", reqPkg.DisplayName, err)
					pkgInfo.Type = PackageTypeUnknown
					packageInfos = append(packageInfos, pkgInfo)
					continue
				}
				pkgInfo.Type = pkgType

				// Check if installed via Homebrew
				installed, err := checkPackageInstalled(ctx, reqPkg.Name, pkgType)
				if err != nil {
					fmt.Printf("⚠ Warning: Error checking installation for %s: %v\n", reqPkg.DisplayName, err)
				}
				pkgInfo.Installed = installed

				// Check if manually installed (outside Homebrew)
				if !installed {
					manuallyInstalled, manualPath := checkManualInstallation(ctx, reqPkg.Name, pkgType)
					if manuallyInstalled {
						pkgInfo.ManuallyInstalled = true
						pkgInfo.ManualPath = manualPath
						pkgInfo.NeedsMigration = true
					}
				}

				if installed {
					// Get current version
					currentVer, err := getPackageVersion(ctx, reqPkg.Name, pkgType)
					if err != nil {
						fmt.Printf("⚠ Warning: Could not get version for %s: %v\n", reqPkg.DisplayName, err)
						currentVer = "unknown"
					}
					pkgInfo.CurrentVersion = currentVer

					// Get latest version
					latestVer, err := getLatestVersion(ctx, reqPkg.Name, pkgType)
					if err != nil {
						fmt.Printf("⚠ Warning: Could not get latest version for %s: %v\n", reqPkg.DisplayName, err)
						latestVer = "unknown"
					}
					pkgInfo.LatestVersion = latestVer

					// Check if upgrade is needed
					if compareVersions(currentVer, latestVer) {
						pkgInfo.NeedsUpgrade = true
						packagesToUpgrade = append(packagesToUpgrade, pkgInfo)
					}
				} else {
					pkgInfo.NeedsInstall = true
					// Get latest version for display
					latestVer, err := getLatestVersion(ctx, reqPkg.Name, pkgType)
					if err == nil {
						pkgInfo.LatestVersion = latestVer
					}
				}

				packageInfos = append(packageInfos, pkgInfo)
			}

			// Display summary table
			printSummaryTable(packageInfos)

			// If there are packages to upgrade, ask for confirmation
			if len(packagesToUpgrade) > 0 {
				fmt.Printf("Found %d package(s) that need upgrading:\n", len(packagesToUpgrade))
				for _, pkg := range packagesToUpgrade {
					fmt.Printf("  - %s (%s -> %s)\n", pkg.DisplayName, pkg.CurrentVersion, pkg.LatestVersion)
				}
				fmt.Println()

				confirm, err := promptConfirmation("Would you like to upgrade these packages?")
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}

				if confirm {
					fmt.Println("\nUpgrading packages...")
					for _, pkg := range packagesToUpgrade {
						if err := upgradePackage(ctx, pkg.Name, pkg.Type); err != nil {
							fmt.Printf("✗ Error upgrading %s: %v\n", pkg.DisplayName, err)
						}
						// Small delay to make progress visible
						time.Sleep(500 * time.Millisecond)
					}
					fmt.Println("\n✓ Package upgrades completed")
				} else {
					fmt.Println("Upgrade cancelled")
				}
			} else {
				fmt.Println("All installed packages are up to date!")
			}

			// Check for packages that need migration (manually installed)
			var packagesToMigrate []PackageInfo
			for _, pkg := range packageInfos {
				if pkg.NeedsMigration {
					packagesToMigrate = append(packagesToMigrate, pkg)
				}
			}

			if len(packagesToMigrate) > 0 {
				fmt.Printf("\nFound %d package(s) installed outside of Homebrew:\n", len(packagesToMigrate))
				for _, pkg := range packagesToMigrate {
					fmt.Printf("  - %s (found at: %s)\n", pkg.DisplayName, pkg.ManualPath)
				}
				fmt.Println()
				fmt.Println("These packages can be migrated to Homebrew management for easier updates.")
				fmt.Println()

				confirm, err := promptConfirmation("Would you like to migrate these packages to Homebrew?")
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}

				if confirm {
					fmt.Println("\nMigrating packages to Homebrew...")
					for _, pkg := range packagesToMigrate {
						fmt.Printf("\n[%s] Starting migration...\n", pkg.DisplayName)

						// Handle migration based on package type
						if pkg.Type == PackageTypeFormula {
							// For command-line tools, install alongside and let Homebrew take precedence
							fmt.Printf("  Installing via Homebrew alongside manual installation\n")
							if err := installPackageAlongside(ctx, pkg.Name, pkg.Type, pkg.ManualPath); err != nil {
								fmt.Printf("✗ Error installing %s via Homebrew: %v\n", pkg.DisplayName, err)
								fmt.Printf("  You may need to manually reinstall %s\n", pkg.DisplayName)
							} else {
								fmt.Printf("  ✓ %s now managed by Homebrew\n", pkg.DisplayName)
							}
						} else {
							// For GUI apps (casks), remove the manual installation first
							if err := removeManualInstallation(ctx, pkg.Name, pkg.Type, pkg.ManualPath); err != nil {
								fmt.Printf("✗ Error removing manual installation: %v\n", err)
								continue
							}

							// Small delay
							time.Sleep(500 * time.Millisecond)

							// Install via Homebrew
							if err := installPackage(ctx, pkg.Name, pkg.Type); err != nil {
								fmt.Printf("✗ Error installing %s via Homebrew: %v\n", pkg.DisplayName, err)
								fmt.Printf("  You may need to manually reinstall %s\n", pkg.DisplayName)
							} else {
								fmt.Printf("  ✓ %s successfully migrated to Homebrew\n", pkg.DisplayName)
							}
						}
					}
					fmt.Println("\n✓ Package migrations completed")
				} else {
					fmt.Println("Migration cancelled")
				}
			}

			// Check for packages that need installation
			var packagesToInstall []PackageInfo
			for _, pkg := range packageInfos {
				if pkg.NeedsInstall && !pkg.NeedsMigration {
					packagesToInstall = append(packagesToInstall, pkg)
				}
			}

			if len(packagesToInstall) > 0 {
				fmt.Printf("\nFound %d package(s) that are not installed:\n", len(packagesToInstall))
				for _, pkg := range packagesToInstall {
					fmt.Printf("  - %s\n", pkg.DisplayName)
				}
				fmt.Println()

				confirm, err := promptConfirmation("Would you like to install these packages?")
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}

				if confirm {
					fmt.Println("\nInstalling packages...")
					for _, pkg := range packagesToInstall {
						if err := installPackage(ctx, pkg.Name, pkg.Type); err != nil {
							fmt.Printf("✗ Error installing %s: %v\n", pkg.DisplayName, err)
						}
						// Small delay to make progress visible
						time.Sleep(500 * time.Millisecond)
					}
					fmt.Println("\n✓ Package installations completed")
				} else {
					fmt.Println("Installation cancelled")
					fmt.Println("\nTo install these packages manually, run:")
					for _, pkg := range packagesToInstall {
						if pkg.Type == PackageTypeCask {
							fmt.Printf("  brew install --cask %s\n", pkg.Name)
						} else {
							fmt.Printf("  brew install %s\n", pkg.Name)
						}
					}
				}
			}

			return nil
		},
	}
}
