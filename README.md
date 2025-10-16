# hook-vault-radar

A Go program that integrates AI agent hook frameworks with HashiCorp Vault Radar for secret scanning. It bridges the gap between interactive AI coding assistants and security scanning tools to prevent accidental exposure of sensitive information.

## Overview

`hook-vault-radar` reads hook data from stdin (as JSON), scans the content using Vault Radar CLI, makes intelligent decisions based on findings, and outputs responses (as JSON) to stdout. This allows AI agent frameworks like Claude Code to automatically scan user prompts and code for secrets before processing.

## Features

- **Extensible Framework Support**: Plugin-based architecture supports multiple hook frameworks
- **Claude Code Integration**: Built-in support for Claude Code's UserPromptSubmit hook
- **Vault Radar Integration**: Leverages HashiCorp Vault Radar CLI for enterprise-grade secret detection
- **Configurable Policies**: Customizable severity thresholds and blocking behavior
- **Structured Logging**: JSON or text logging to stderr (stdout reserved for hook responses)
- **Zero Dependencies Runtime**: Single binary with no external dependencies except vault-radar CLI

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    hook-vault-radar CLI                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  stdin (JSON) → Framework Parser → Content Extractor        │
│                      ↓                                       │
│                 Vault Radar Scanner                          │
│                      ↓                                       │
│                 Decision Engine                              │
│                      ↓                                       │
│              Response Formatter → stdout (JSON)              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Components

- **Framework Layer** (`internal/framework/`): Abstracts different hook frameworks (Claude, etc.)
- **Scanner Layer** (`internal/scanner/`): Interfaces with security scanning tools (Vault Radar)
- **Decision Engine** (`internal/decision/`): Applies policies and makes blocking decisions
- **Processor** (`internal/processor/`): Orchestrates the complete flow

## Installation

### Prerequisites

- Go 1.21 or later
- HashiCorp Vault Radar CLI installed and configured

### Build from Source

```bash
go build -o hook-vault-radar
```

### Install

```bash
# Install to ~/.local/bin (or your preferred location)
cp hook-vault-radar ~/.local/bin/
chmod +x ~/.local/bin/hook-vault-radar
```

## Configuration

Configuration is loaded from `~/.hook-vault-radar/config.yaml` (or current directory). All settings have sensible defaults.

### Environment Variables (.env File)

The application supports loading environment variables from `.env` files for HCP credentials and configuration:

```bash
# Copy the example file
cp .env.example .env

# Edit with your HCP credentials
# Required for vault-radar CLI:
HCP_PROJECT_ID=your-project-id
HCP_CLIENT_ID=your-client-id
HCP_CLIENT_SECRET=your-client-secret
```

**.env File Locations** (checked in order):
1. `./.env` - Current directory
2. `~/.hook-vault-radar/.env` - Config directory
3. `./.env.local` - Local overrides (current directory)
4. `~/.hook-vault-radar/.env.local` - Local overrides (config directory)

**Note**: `.env` files are gitignored to prevent accidental commit of secrets.

### YAML Configuration

```yaml
vault_radar:
  command: "vault-radar"
  scan_command: "scan folder"
  timeout_seconds: 30
  extra_args: []

logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json or text

decision:
  block_on_findings: true
  severity_threshold: "high" # critical, high, medium, low
```

### Environment Variable Overrides

Override YAML configuration via environment variables with the `HOOK_VAULT_RADAR_` prefix:

```bash
export HOOK_VAULT_RADAR_LOGGING_LEVEL=debug
export HOOK_VAULT_RADAR_DECISION_SEVERITY_THRESHOLD=medium
```

**Configuration Precedence** (lowest to highest):
1. Default values
2. `.env` files
3. YAML config file (`config.yaml`)
4. Environment variables (`HOOK_VAULT_RADAR_*`)
5. Command-line flags

## Usage

### Claude Code Integration

Add to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "hook-vault-radar --framework claude",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
```

### Command Line

The `--framework` flag is required to specify which hook framework you're using:

```bash
# Process hook input from stdin (framework flag is required)
cat testdata/claude/userpromptsubmit.json | ./hook-vault-radar --framework claude

# With debug logging
cat testdata/claude/userpromptsubmit.json | ./hook-vault-radar --framework claude --log-level debug

# View help
./hook-vault-radar --help
```

## Hook Flow Example

### Input (stdin)
```json
{
  "session_id": "abc123",
  "transcript_path": "/tmp/transcript.jsonl",
  "cwd": "/Users/dev/project",
  "hook_event_name": "UserPromptSubmit",
  "prompt": "Configure AWS with key AKIAIOSFODNN7EXAMPLE"
}
```

### Output (stdout) - When Secrets Found
```json
{
  "decision": "block",
  "reason": "Vault Radar detected 1 security finding:\n\n1. [HIGH] aws_access_key: AWS Access Key detected\n\nPlease remove or redact sensitive information before proceeding.",
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit"
  },
  "continue": false,
  "suppressOutput": false,
  "systemMessage": "Vault Radar detected 1 security finding..."
}
```

### Output (stdout) - Clean Content
```json
{
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit"
  },
  "continue": true,
  "suppressOutput": false
}
```

## Development

### Project Structure

```
agent-hook-vault-radar/
├── main.go                            # Entry point
├── cmd/
│   └── root.go                        # Cobra root command
├── internal/
│   ├── config/                        # Configuration management
│   ├── framework/                     # Hook framework abstractions
│   │   └── claude/                    # Claude Code implementation
│   ├── scanner/                       # Scanner interface + implementations
│   ├── decision/                      # Decision engine and policies
│   └── processor/                     # Main orchestration logic
├── pkg/
│   └── types/                         # Shared type definitions
└── testdata/                          # Test fixtures
```

### Testing

```bash
# Build
go build

# Test with sample input
cat testdata/claude/userpromptsubmit.json | ./hook-vault-radar --log-level debug

# Test with clean input (no secrets)
cat testdata/claude/userpromptsubmit_clean.json | ./hook-vault-radar
```

### Adding New Hook Frameworks

1. Create a new directory under `internal/framework/`
2. Implement the `HookFramework` interface
3. Implement `HookHandler` interface for each hook type
4. Register the framework in `processor.go`

### Adding New Scanners

1. Implement the `Scanner` interface in `internal/scanner/`
2. Update the processor to use the new scanner

## Exit Codes

- `0`: Success (content scanned, no blocking)
- `1`: Error (unexpected failure)
- `2`: Blocking decision (secrets found and blocked)

## Logging

All logs go to stderr in JSON format (configurable to text). Stdout is reserved exclusively for hook framework communication.

Example log entry:
```json
{"time":"2025-10-15T10:30:00Z","level":"INFO","msg":"scan completed","has_findings":true,"finding_count":1,"duration":"1.2s"}
```

## Contributing

This project follows the Golang Standards Project Layout and Go Style Guide. Contributions welcome.

## License

MIT License

## Security Considerations

- Vault Radar CLI must be properly configured with valid credentials
- Temporary files are created in secure temp directories and cleaned up
- All sensitive content is written to files with 0600 permissions
- Logs may contain information about findings (not the secrets themselves)

## Future Enhancements

- Additional hook framework support (GitHub Actions, Jenkins, etc.)
- Alternative scanner support (TruffleHog, git-secrets)
- Custom policy rules and severity mapping
- Webhook notifications on findings
- Caching layer for scan results
- Metrics and telemetry
