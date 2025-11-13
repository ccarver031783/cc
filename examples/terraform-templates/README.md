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

### Using Default Location

The default location for template files is `~/.cc/terraform-templates/`.

**Scaffold a single provider directory:**

```bash
cc tf init-dir ./my-terraform --provider aws
```

**Scaffold a multi-provider resource:**

```bash
cc tf new my-resource --providers aws,azure
```

### Using Custom Config Path

**Single provider with custom config file:**

```bash
cc tf init-dir ./my-terraform --provider aws --config /path/to/aws.yaml
```

**Multi-provider with custom config directory:**

```bash
cc tf new my-resource --providers aws,azure --config /path/to/templates/
```

## Template File Structure

Each YAML file supports the following sections:

### Provider Configuration

The `provider` section configures the cloud provider block:

- **`region`**: Default AWS region / Azure location / GCP region
- **`profile`**: AWS profile name (AWS only)
- **`credentials`**: Path to credentials file
- **`tags`**: Common tags to apply to resources
- **`extra`**: Provider-specific settings
  - Azure: `subscription_id`, `tenant_id`
  - GCP: `project`, `zone`

### Variables

The `variables` section defines variables to include in `input.tf`:

- **`name`**: Variable name
- **`description`**: Variable description
- **`type`**: Terraform type (`string`, `number`, `bool`, `map(string)`, etc.)
- **`default`**: Default value (optional)

### Outputs

The `outputs` section defines outputs to include in `output.tf`:

- **`name`**: Output name
- **`description`**: Output description
- **`value`**: Output value expression

### Data Sources

The `data` section defines data sources to include in `data.tf`:

- **`type`**: Data source type (e.g., `aws_vpc`, `azurerm_resource_group`)
- **`name`**: Data source name
- **`config`**: Key-value pairs for data source configuration
- **`comment`**: Optional comment above the data source

### Terraform Variables

The `tfvars` section defines key-value pairs to include in `terraform.tfvars`.

## Examples

See the example files for complete examples:

- [`aws.yaml`](aws.yaml) - AWS provider template
- [`azure.yaml`](azure.yaml) - Azure provider template
- [`gcp.yaml`](gcp.yaml) - GCP provider template
