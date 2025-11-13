package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProviderTemplateConfig holds the template configuration for a provider
type ProviderTemplateConfig struct {
	Provider struct {
		Region      string            `yaml:"region"`
		Profile     string            `yaml:"profile"`
		Credentials string            `yaml:"credentials"`
		Tags        map[string]string `yaml:"tags"`
		Extra       map[string]string `yaml:"extra"`
	} `yaml:"provider"`
	Variables []VariableTemplate `yaml:"variables"`
	Outputs   []OutputTemplate   `yaml:"outputs"`
	Data      []DataTemplate     `yaml:"data"`
	Tfvars    map[string]string   `yaml:"tfvars"`
}

// VariableTemplate defines a variable template
type VariableTemplate struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Default     string `yaml:"default,omitempty"`
}

// OutputTemplate defines an output template
type OutputTemplate struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Value       string `yaml:"value"`
}

// DataTemplate defines a data source template
type DataTemplate struct {
	Type    string            `yaml:"type"`
	Name    string            `yaml:"name"`
	Config  map[string]string `yaml:"config"`
	Comment string            `yaml:"comment,omitempty"`
}

// LoadProviderTemplate loads a provider template configuration from a YAML file
func LoadProviderTemplate(configPath, provider string) (*ProviderTemplateConfig, error) {
	// If no config path provided, use default location
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".cc", "terraform-templates", fmt.Sprintf("%s.yaml", provider))
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default empty config if file doesn't exist
		return &ProviderTemplateConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProviderTemplateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GenerateProviderBlock generates the provider configuration block from template
func (c *ProviderTemplateConfig) GenerateProviderBlock(providerName string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf(`provider "%s" {`, providerName))

	if c.Provider.Region != "" {
		lines = append(lines, fmt.Sprintf(`  region = "%s"`, c.Provider.Region))
	}
	if c.Provider.Profile != "" {
		lines = append(lines, fmt.Sprintf(`  profile = "%s"`, c.Provider.Profile))
	}
	if c.Provider.Credentials != "" {
		lines = append(lines, fmt.Sprintf(`  shared_credentials_files = ["%s"]`, c.Provider.Credentials))
	}

	// Add provider-specific configurations
	if providerName == "azurerm" {
		if c.Provider.Region != "" {
			lines = append(lines, fmt.Sprintf(`  features {}`))
		}
		if subscriptionID, ok := c.Provider.Extra["subscription_id"]; ok {
			lines = append(lines, fmt.Sprintf(`  subscription_id = "%s"`, subscriptionID))
		}
		if tenantID, ok := c.Provider.Extra["tenant_id"]; ok {
			lines = append(lines, fmt.Sprintf(`  tenant_id = "%s"`, tenantID))
		}
	}

	if providerName == "google" {
		if project, ok := c.Provider.Extra["project"]; ok {
			lines = append(lines, fmt.Sprintf(`  project = "%s"`, project))
		}
		if region, ok := c.Provider.Extra["region"]; ok {
			lines = append(lines, fmt.Sprintf(`  region = "%s"`, region))
		}
		if zone, ok := c.Provider.Extra["zone"]; ok {
			lines = append(lines, fmt.Sprintf(`  zone = "%s"`, zone))
		}
	}

	lines = append(lines, `}`)
	return strings.Join(lines, "\n")
}

// GenerateVariablesBlock generates the variables.tf content from template
func (c *ProviderTemplateConfig) GenerateVariablesBlock() string {
	if len(c.Variables) == 0 {
		return `# Variable declarations
# Example:
# variable "example" {
#   description = "Example variable"
#   type        = string
# }
`
	}

	var lines []string
	lines = append(lines, `# Variable declarations`)
	for _, v := range c.Variables {
		lines = append(lines, ``)
		lines = append(lines, fmt.Sprintf(`variable "%s" {`, v.Name))
		if v.Description != "" {
			lines = append(lines, fmt.Sprintf(`  description = %q`, v.Description))
		}
		if v.Type != "" {
			lines = append(lines, fmt.Sprintf(`  type        = %s`, v.Type))
		}
		if v.Default != "" {
			lines = append(lines, fmt.Sprintf(`  default     = %s`, v.Default))
		}
		lines = append(lines, `}`)
	}

	return strings.Join(lines, "\n")
}

// GenerateOutputsBlock generates the outputs.tf content from template
func (c *ProviderTemplateConfig) GenerateOutputsBlock() string {
	if len(c.Outputs) == 0 {
		return `# Output declarations
# Example:
# output "example" {
#   description = "Example output"
#   value       = var.example
# }
`
	}

	var lines []string
	lines = append(lines, `# Output declarations`)
	for _, o := range c.Outputs {
		lines = append(lines, ``)
		lines = append(lines, fmt.Sprintf(`output "%s" {`, o.Name))
		if o.Description != "" {
			lines = append(lines, fmt.Sprintf(`  description = %q`, o.Description))
		}
		if o.Value != "" {
			lines = append(lines, fmt.Sprintf(`  value       = %s`, o.Value))
		}
		lines = append(lines, `}`)
	}

	return strings.Join(lines, "\n")
}

// GenerateDataBlock generates the data.tf content from template
func (c *ProviderTemplateConfig) GenerateDataBlock(provider string) string {
	if len(c.Data) == 0 {
		// Provider-specific default examples
		switch provider {
		case "aws":
			return `# Data source declarations
# Example:
# data "aws_vpc" "example" {
#   filter {
#     name   = "tag:Name"
#     values = ["example-vpc"]
#   }
# }
`
		case "azure":
			return `# Data source declarations
# Example:
# data "azurerm_resource_group" "example" {
#   name = "example-rg"
# }
`
		case "gcp":
			return `# Data source declarations
# Example:
# data "google_project" "example" {
#   project_id = "example-project"
# }
`
		default:
			return `# Data source declarations
`
		}
	}

	var lines []string
	lines = append(lines, `# Data source declarations`)
	for _, d := range c.Data {
		if d.Comment != "" {
			lines = append(lines, fmt.Sprintf(`# %s`, d.Comment))
		}
		lines = append(lines, ``)
		lines = append(lines, fmt.Sprintf(`data "%s" "%s" {`, d.Type, d.Name))
		for key, value := range d.Config {
			lines = append(lines, fmt.Sprintf(`  %s = %s`, key, value))
		}
		lines = append(lines, `}`)
	}

	return strings.Join(lines, "\n")
}

// GenerateTfvarsBlock generates the terraform.tfvars content from template
func (c *ProviderTemplateConfig) GenerateTfvarsBlock() string {
	if len(c.Tfvars) == 0 {
		return `# Variable assignments
# Example:
# example = "value"
`
	}

	var lines []string
	lines = append(lines, `# Variable assignments`)
	for key, value := range c.Tfvars {
		lines = append(lines, fmt.Sprintf(`%s = %q`, key, value))
	}

	return strings.Join(lines, "\n")
}

