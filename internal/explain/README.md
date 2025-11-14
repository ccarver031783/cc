# Explain Command

AI-powered explanations for Terraform modules using Claude API or local Ollama.

## Features

- 🤖 **Dual AI Support**: Claude API (cloud) with Ollama fallback (local)
- 🔒 **Smart Fallback**: Automatically falls back to Ollama if Claude fails
- 📁 **Directory Analysis**: Reads all common Terraform files in a module
- 🎯 **Structured Output**: Clear, formatted explanations in markdown

## Setup

### Option 1: Claude API (Recommended)

1. Get an API key from [Anthropic Console](https://console.anthropic.com/)
2. Set environment variable:
   ```bash
   export ANTHROPIC_API_KEY=your_api_key_here
   ```
3. Add to your `~/.zshrc` or `~/.bashrc` to persist

### Option 2: Local Ollama (No API Key Required)

1. Install Ollama: https://ollama.ai/download
2. Pull a model:
   ```bash
   ollama pull llama3.2
   ```
3. Start Ollama (in a separate terminal):
   ```bash
   ollama serve
   ```

## Usage

### Basic Usage (tries Claude, falls back to Ollama)
```bash
cc explain tf /path/to/terraform/module
```

### Explain current directory
```bash
cc explain tf .
```

### Force local Ollama (skip Claude)
```bash
cc explain tf /path/to/module --local
```

## Examples

```bash
# Explain an S3 module
cc explain tf ./terraform-modules/aws/s3

# Explain current Terraform directory
cd infrastructure-terraform/prod-useast1
cc explain tf .

# Use only local Ollama
cc explain tf ./my-module --local
```

## Output Format

The command will show:
1. **Purpose**: What the module does
2. **Resources**: Infrastructure components created
3. **Key Variables**: Important inputs
4. **Outputs**: Values exposed to other modules
5. **Dependencies**: Related infrastructure
6. **Use Case**: When to use this module

## Files Analyzed

The tool looks for these common Terraform files:
- `main.tf` - Main resource definitions
- `variables.tf` - Input variables
- `outputs.tf` - Output values
- `versions.tf` - Provider versions
- `README.md` - Existing documentation

## Troubleshooting

### "ollama is not running"
Start Ollama in a separate terminal:
```bash
ollama serve
```

### "claude api error"
Check your API key:
```bash
echo $ANTHROPIC_API_KEY
```

### Model not found (Ollama)
Pull the default model:
```bash
ollama pull llama3.2
```

## Cost Considerations

- **Claude API**: ~$0.003 per module explanation (varies by module size)
- **Ollama**: Free, unlimited, but requires local compute power

## Tips

- Use `--local` flag when offline or to save API credits
- Claude provides higher quality explanations
- Ollama is great for quick, offline analysis
- Add `ANTHROPIC_API_KEY` to your shell profile for persistence


