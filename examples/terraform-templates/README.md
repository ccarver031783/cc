# Terraform Template Configuration

This directory contains example template configuration files for customizing Terraform scaffolding.

## Setup

1. Copy the template files to your home directory:
   ```bash
   mkdir -p ~/.cc/terraform-templates
   cp aws.yaml ~/.cc/terraform-templates/
   cp azure.yaml ~/.cc/terraform-templates/
   cp gcp.yaml ~/.cc/terraform-templates/
   ```

2. Edit the template files to match your organization's defaults:
   - Update provider settings (region, credentials, etc.)
   - Add common variables your team uses
   - Configure default outputs
   - Set up data sources you frequently reference

## Usage

### Using Default Location (~/.cc/terraform-templates/)

```bash
# Scaffold with default templates
cc tf init-dir ./my-terraform --provider aws

# Scaffold multi-provider resource
cc tf new my-resource --providers aws,azure
```

### Using Custom Config Path

```bash
# Single provider with custom config
cc tf init-dir ./my-terraform --provider aws --config /path/to/aws.yaml

# Multi-provider with custom config directory
cc tf new my-resource --providers aws,azure --config /path/to/templates/
```

## Template File Structure

Each YAML file supports the following sections:

### `provider`
- `region`: Default AWS region / Azure location / GCP region
- `profile`: AWS profile name (AWS only)
- `credentials`: Path to credentials file
- `tags`: Common tags to apply
- `extra`: Provider-specific settings (subscription_id, tenant_id for Azure; project, zone for GCP)

### `variables`
List of variables to include in `input.tf`:
- `name`: Variable name
- `description`: Variable description
- `type`: Terraform type (string, number, bool, map(string), etc.)
- `default`: Default value (optional)

### `outputs`
List of outputs to include in `output.tf`:
- `name`: Output name
- `description`: Output description
- `value`: Output value expression

### `data`
List of data sources to include in `data.tf`:
- `type`: Data source type (e.g., `aws_vpc`, `azurerm_resource_group`)
- `name`: Data source name
- `config`: Key-value pairs for data source configuration
- `comment`: Optional comment above the data source

### `tfvars`
Key-value pairs to include in `terraform.tfvars`

## Examples

See the example files (`aws.yaml`, `azure.yaml`, `gcp.yaml`) for complete examples.

