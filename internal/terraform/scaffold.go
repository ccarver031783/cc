package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ufcli "github.com/urfave/cli/v2"
)

// ============================================================================
// Scaffolding Helper Functions
// ============================================================================

// scaffoldTerraformDir creates a new Terraform directory with standard files.
// This is a helper function used by both init-dir and tf new commands.
// configPath can be empty to use default location or a specific path to a template config file.
func scaffoldTerraformDir(dirPath, provider, tfVersion, providerVersion, configPath string) error {
	// Validate provider
	if provider != "aws" && provider != "azure" && provider != "gcp" {
		return fmt.Errorf("provider must be 'aws', 'azure', or 'gcp', got: %s", provider)
	}

	// Validate and clean the directory path
	safePath, err := validatePath(dirPath)
	if err != nil {
		return fmt.Errorf("invalid directory path: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(safePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Load template configuration
	templateConfig, err := LoadProviderTemplate(configPath, provider)
	if err != nil {
		return fmt.Errorf("failed to load template config: %w", err)
	}

	// Create version.tf
	versionContent := fmt.Sprintf(`terraform {
  required_version = "%s"
  required_providers {
`, tfVersion)
	if provider == "aws" {
		versionContent += fmt.Sprintf(`    aws = {
      source  = "hashicorp/aws"
      version = "%s"
    }
`, providerVersion)
	} else if provider == "azure" {
		versionContent += fmt.Sprintf(`    azurerm = {
      source  = "hashicorp/azurerm"
      version = "%s"
    }
`, providerVersion)
	} else if provider == "gcp" {
		versionContent += fmt.Sprintf(`    google = {
      source  = "hashicorp/google"
      version = "%s"
    }
`, providerVersion)
	}
	versionContent += `  }
}
`

	// Create main.tf using template
	var providerName string
	if provider == "aws" {
		providerName = "aws"
	} else if provider == "azure" {
		providerName = "azurerm"
	} else if provider == "gcp" {
		providerName = "google"
	}
	mainContent := templateConfig.GenerateProviderBlock(providerName)

	// Create input.tf using template
	inputContent := templateConfig.GenerateVariablesBlock()

	// Create output.tf using template
	outputContent := templateConfig.GenerateOutputsBlock()

	// Create data.tf using template
	dataContent := templateConfig.GenerateDataBlock(provider)

	// Create terraform.tfvars using template
	tfvarsContent := templateConfig.GenerateTfvarsBlock()

	// Write all files
	files := map[string]string{
		"version.tf":       versionContent,
		"main.tf":          mainContent,
		"input.tf":         inputContent,
		"output.tf":        outputContent,
		"data.tf":          dataContent,
		"terraform.tfvars": tfvarsContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(safePath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	fmt.Printf("✓ Successfully scaffolded Terraform directory: %s\n", safePath)
	fmt.Printf("  Created files: version.tf, main.tf, input.tf, output.tf, data.tf, terraform.tfvars\n")
	return nil
}

// scaffoldMultiProviderResource creates a base directory structure for a new resource
// that supports multiple cloud providers (AWS, Azure, GCP). Creates a main directory
// with provider-specific subdirectories, each containing a complete Terraform structure.
// providers should be a slice of provider names (e.g., []string{"aws", "azure"}).
// configBasePath is the base directory for template config files (empty uses default).
func scaffoldMultiProviderResource(basePath, resourceName, tfVersion string, providers []string, awsVersion, azureVersion, gcpVersion, configBasePath string) error {
	// Validate providers
	validProviders := map[string]bool{"aws": true, "azure": true, "gcp": true}
	providerVersions := map[string]string{
		"aws":   awsVersion,
		"azure": azureVersion,
		"gcp":   gcpVersion,
	}

	if len(providers) == 0 {
		return fmt.Errorf("at least one provider must be specified")
	}

	for _, p := range providers {
		if !validProviders[p] {
			return fmt.Errorf("invalid provider: %s (must be one of: aws, azure, gcp)", p)
		}
	}

	// Validate and clean the base path
	safePath, err := validatePath(basePath)
	if err != nil {
		return fmt.Errorf("invalid base path: %w", err)
	}

	// Create the resource directory
	resourceDir := filepath.Join(safePath, resourceName)
	if err := os.MkdirAll(resourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create resource directory: %w", err)
	}

	// Build provider list for README
	providerList := make([]string, len(providers))
	for i, p := range providers {
		providerList[i] = strings.ToUpper(p)
	}

	// Build structure tree for README
	var structureTree strings.Builder
	for i, p := range providers {
		prefix := "├──"
		indent := "│   "
		if i == len(providers)-1 {
			prefix = "└──"
			indent = "    "
		}
		structureTree.WriteString(fmt.Sprintf("%s %s/\n", prefix, p))
		structureTree.WriteString(fmt.Sprintf("%s├── version.tf\n", indent))
		structureTree.WriteString(fmt.Sprintf("%s├── main.tf\n", indent))
		structureTree.WriteString(fmt.Sprintf("%s├── input.tf\n", indent))
		structureTree.WriteString(fmt.Sprintf("%s├── output.tf\n", indent))
		structureTree.WriteString(fmt.Sprintf("%s├── data.tf\n", indent))
		structureTree.WriteString(fmt.Sprintf("%s└── terraform.tfvars\n", indent))
	}

	// Create a README.md in the root resource directory
	readmeContent := fmt.Sprintf(`# %s

Multi-cloud Terraform resource for %s.

This resource supports the following cloud providers:
%s

Each provider has its own subdirectory with provider-specific Terraform configurations.

## Structure

\`\`\`
%s/
%s\`\`\`

## Usage

Navigate to the provider-specific directory and initialize Terraform:

\`\`\`bash
cd %s/%s
terraform init
terraform plan
terraform apply
\`\`\`
`, resourceName, resourceName, "- "+strings.Join(providerList, "\n- "), resourceName, structureTree.String(), resourceName, providers[0])

	readmePath := filepath.Join(resourceDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	// Create provider-specific directories
	var createdDirs []string
	for _, p := range providers {
		providerDir := filepath.Join(resourceDir, p)
		// Build config path for this provider
		var configPath string
		if configBasePath != "" {
			configPath = filepath.Join(configBasePath, fmt.Sprintf("%s.yaml", p))
		}
		if err := scaffoldTerraformDir(providerDir, p, tfVersion, providerVersions[p], configPath); err != nil {
			return fmt.Errorf("failed to scaffold %s directory: %w", p, err)
		}
		createdDirs = append(createdDirs, p+"/")
	}

	fmt.Printf("✓ Successfully created multi-provider resource structure: %s\n", resourceDir)
	fmt.Printf("  Created directories: %s\n", strings.Join(createdDirs, ", "))
	fmt.Printf("  Each directory contains: version.tf, main.tf, input.tf, output.tf, data.tf, terraform.tfvars\n")
	return nil
}

// ============================================================================
// Scaffolding Commands
// ============================================================================

// NewTerraformInitDirCmd creates the init-dir command.
// Scaffolds a new Terraform directory with standard file structure:
// - main.tf: Provider configuration
// - input.tf: Variable declarations
// - output.tf: Output declarations (empty template)
// - data.tf: Data source declarations (empty template)
// - version.tf: Terraform and provider version constraints
// - terraform.tfvars: Variable assignments (empty template)
// This ensures consistency across all Terraform directories in the repository.
func NewTerraformInitDirCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:      "init-dir",
		Usage:     "Scaffold a new Terraform directory with standard files",
		ArgsUsage: "<directory-path>",
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "Cloud provider: aws, azure, gcp (default: aws)",
				Value:   "aws",
			},
			&ufcli.StringFlag{
				Name:  "terraform-version",
				Usage: "Terraform version constraint (default: >= 1.4.0)",
				Value: ">= 1.4.0",
			},
			&ufcli.StringFlag{
				Name:  "aws-version",
				Usage: "AWS provider version constraint (default: >= 5.0)",
				Value: ">= 5.0",
			},
			&ufcli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to provider template config file (default: ~/.cc/terraform-templates/{provider}.yaml)",
				Value:   "",
			},
		},
		Action: func(c *ufcli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("directory path is required")
			}

			dirPath := c.Args().First()
			provider := c.String("provider")
			tfVersion := c.String("terraform-version")
			awsVersion := c.String("aws-version")
			configPath := c.String("config")

			// Determine provider version based on provider type
			var providerVersion string
			switch provider {
			case "aws":
				providerVersion = awsVersion
			case "azure":
				providerVersion = ">= 3.0" // Default for Azure
			case "gcp":
				providerVersion = ">= 4.0" // Default for GCP
			default:
				providerVersion = awsVersion // Fallback
			}

			return scaffoldTerraformDir(dirPath, provider, tfVersion, providerVersion, configPath)
		},
	}
}

// NewTerraformNewCmd creates the new command.
// Creates a base directory structure for a new resource that supports multiple
// cloud providers (AWS, Azure, GCP). Each provider gets its own subdirectory
// with a complete Terraform file structure, allowing for provider-specific
// implementations while maintaining a consistent overall structure.
func NewTerraformNewCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:      "new",
		Usage:     "Create a new multi-provider resource directory structure",
		ArgsUsage: "<resource-name>",
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Base path where the resource directory will be created (default: current directory)",
				Value:   ".",
			},
			&ufcli.StringSliceFlag{
				Name:    "providers",
				Aliases: []string{"provider"},
				Usage:   "Cloud providers to create (aws, azure, gcp). Can be specified multiple times or comma-separated. Default: all providers",
				Value:   nil,
			},
			&ufcli.StringFlag{
				Name:  "terraform-version",
				Usage: "Terraform version constraint (default: >= 1.4.0)",
				Value: ">= 1.4.0",
			},
			&ufcli.StringFlag{
				Name:  "aws-version",
				Usage: "AWS provider version constraint (default: >= 5.0)",
				Value: ">= 5.0",
			},
			&ufcli.StringFlag{
				Name:  "azure-version",
				Usage: "Azure provider version constraint (default: >= 3.0)",
				Value: ">= 3.0",
			},
			&ufcli.StringFlag{
				Name:  "gcp-version",
				Usage: "GCP provider version constraint (default: >= 4.0)",
				Value: ">= 4.0",
			},
			&ufcli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Base path to provider template config files directory (default: ~/.cc/terraform-templates/)",
				Value:   "",
			},
		},
		Action: func(c *ufcli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("resource name is required")
			}

			resourceName := c.Args().First()
			basePath := c.String("path")
			tfVersion := c.String("terraform-version")
			awsVersion := c.String("aws-version")
			azureVersion := c.String("azure-version")
			gcpVersion := c.String("gcp-version")
			configBasePath := c.String("config")

			// Parse providers flag
			providerSlice := c.StringSlice("providers")
			var providers []string

			if len(providerSlice) == 0 {
				// Default: create all providers
				providers = []string{"aws", "azure", "gcp"}
			} else {
				// Parse comma-separated values and collect unique providers
				providerMap := make(map[string]bool)
				for _, p := range providerSlice {
					// Handle comma-separated values
					parts := strings.Split(p, ",")
					for _, part := range parts {
						part = strings.TrimSpace(strings.ToLower(part))
						if part != "" {
							providerMap[part] = true
						}
					}
				}
				// Convert map to slice
				for p := range providerMap {
					providers = append(providers, p)
				}
			}

			return scaffoldMultiProviderResource(basePath, resourceName, tfVersion, providers, awsVersion, azureVersion, gcpVersion, configBasePath)
		},
	}
}

