# agent-hook-vault-radar

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
┌───────────────────────────────────────────────────────┐
│               hook-vault-radar CLI                    │
├───────────────────────────────────────────────────────┤
│                                                       │
│  stdin (JSON) → Framework Parser → Content Extractor  │
│                         ↓                             │
│                Vault Radar Scanner                    │
│                         ↓                             │
│                 Decision Engine                       │
│                         ↓                             │
│          Response Formatter → stdout (JSON)           │
│                                                       │
└───────────────────────────────────────────────────────┘
```

### Components

- **Framework Layer** (`internal/framework/`): Abstracts different hook frameworks (Claude, etc.)
- **Scanner Layer** (`internal/scanner/`): Interfaces with security scanning tools (Vault Radar)
- **Decision Engine** (`internal/decision/`): Applies policies and makes blocking decisions
- **Processor** (`internal/processor/`): Orchestrates the complete flow

## Installation

### From Source

#### Prerequisites

- Go 1.25.2 or later

#### Build from Source

```bash
go build -o hook-vault-radar
```

### Install

```bash
# Install to ~/.local/bin (or your preferred location)
cp hook-vault-radar ~/.agent-hooks/vault-radar/hook-vault-radar
chmod +x ~/.agent-hooks/vault-radar/hook-vault-radar
```

## Configuration

Configuration is loaded from `~/.agent-hooks/vault-radar/config.yaml` (or current directory). All settings have sensible defaults.

### Environment Variables (.env File)

The application supports loading environment variables from `.env` files (in the same directory as the executable) for HCP credentials and configuration:

```bash
# Copy the example file
cp .env.example .env

# Edit with your HCP credentials
# Required for vault-radar CLI:
HCP_PROJECT_ID=your-project-id
HCP_CLIENT_ID=your-client-id
HCP_CLIENT_SECRET=your-client-secret
```

**.env File Locations** (all existing files are loaded; later files override earlier ones):
1. `~/.agent-hooks/vault-radar/.env.local` - Local overrides (config directory) - **highest precedence**
2. `./.env.local` - Local overrides (current directory)
3. `~/.agent-hooks/vault-radar/.env` - Config directory
4. `./.env` - Current directory - **lowest precedence**

**Note**: `.env` files are gitignored to prevent accidental commit of secrets.

### YAML Configuration

```yaml
framework: "claude"  # Hook framework to use (currently supported: claude)

vault_radar:
  command: "vault-radar"
  scan_command: "scan file"
  timeout_seconds: 30
  extra_args: []  # Additional vault-radar flags (--disable-ui is always included automatically)

logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json or text
  log_file: ""   # Optional: path to log file (empty = stderr only, supports ~ for home dir)

decision:
  block_on_findings: true
  severity_threshold: "medium" # critical, high, medium, low
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

### Severity Threshold Configuration

The `decision.severity_threshold` setting controls which security findings trigger blocking behavior. It acts as a **minimum severity level** - findings at the threshold level or higher will cause blocking when `block_on_findings` is `true`.

**Severity Levels** (from lowest to highest):
- `low` (level 1) - Minor security concerns
- `medium` / `info` (level 2) - Moderate security risks; Vault Radar uses "info" for many real secrets like AWS keys (default)
- `high` (level 3) - Serious security issues
- `critical` (level 4) - Critical security vulnerabilities

**How It Works**:

The threshold filters findings using a "greater than or equal to" comparison:

| Threshold | Blocks on | Ignores |
|-----------|-----------|---------|
| `critical` | Critical findings only | High, Medium, Info, Low |
| `high` | Critical + High findings | Medium, Info, Low |
| `medium` | Critical + High + Medium + Info findings | Low |
| `low` | All findings | None |

**Example**:

```yaml
decision:
  block_on_findings: true
  severity_threshold: "medium"  # Block on MEDIUM, INFO, HIGH, and CRITICAL findings
```

If Vault Radar detects:
- 1 CRITICAL finding → **Blocks** ✓
- 2 HIGH findings → **Blocks** ✓
- 3 MEDIUM findings → **Blocks** ✓
- 3 INFO findings → **Blocks** ✓
- 1 LOW finding → **Allows** (below threshold)

**Note**: If `block_on_findings` is `false`, findings are still reported but never block execution, regardless of severity threshold.

## Usage

### Claude Code Integration

Add to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
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

## Hook Flow Example (Claude Framework - UserPromptSubmit)

**Note**: This example shows the input/output format for the Claude framework's `UserPromptSubmit` hook. Other hooks in the Claude framework (such as future tool-specific hooks) may have different input/output structures. Hook frameworks other than Claude will have entirely different formats.

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
  "reason": "\nVault Radar detected 1 security finding:\n\n1. [INFO] aws_access_key_id: AWS access key ID (scan-content.txt)\n\nPlease remove or redact sensitive information before proceeding.",
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit"
  },
  "continue": true,
  "suppressOutput": false,
  "systemMessage": "Vault Radar detected 1 security finding:\n\n1. [INFO] aws_access_key_id: AWS access key ID (scan-content.txt)\n\nPlease remove or redact sensitive information before proceeding."
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
├── main.go                              # Entry point
├── go.mod                               # Go module definition
├── go.sum                               # Go module checksums
├── Makefile                             # Build automation
├── README.md                            # Project documentation
├── .gitignore                           # Git ignore rules
├── cmd/                                 # CLI commands
│   ├── root.go                          # Cobra root command
│   └── version.go                       # Version subcommand
├── internal/                            # Internal packages
│   ├── config/                          # Configuration management
│   │   ├── config.go                    # Viper config initialization
│   │   ├── constants.go                 # Default configuration values
│   │   └── types.go                     # Configuration type definitions
│   ├── framework/                       # Hook framework abstractions
│   │   ├── framework.go                 # Framework and handler interfaces
│   │   ├── registry.go                  # Framework registration system
│   │   └── claude/                      # Claude Code implementation
│   │       ├── claude.go                # Claude framework implementation
│   │       ├── types.go                 # Claude-specific types
│   │       └── userpromptsubmit.go      # UserPromptSubmit handler
│   ├── scanner/                         # Scanner interface + implementations
│   │   ├── scanner.go                   # Scanner interface definition
│   │   └── vaultradar.go                # Vault Radar CLI wrapper
│   ├── decision/                        # Decision engine and policies
│   │   └── decision.go                  # Policy-based decision making
│   └── processor/                       # Main orchestration logic
│       └── processor.go                 # Hook processing orchestration
├── pkg/                                 # Public packages
│   └── types/                           # Shared type definitions
│       └── types.go                     # Common types used across packages
└── testdata/                            # Test fixtures
    └── claude/                          # Claude framework test data
        ├── userpromptsubmit.json        # Test with secrets
        └── userpromptsubmit_clean.json  # Test without secrets
```

### Testing

```bash
# Build
go build

# Test with sample input
cat testdata/claude/userpromptsubmit.json | ./hook-vault-radar --framework claude --log-level debug

# Test with clean input (no secrets)
cat testdata/claude/userpromptsubmit_clean.json | ./hook-vault-radar --framework claude
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

Exit code behavior is framework-specific:

**Claude Code (current implementation):**
- `0`: All responses (blocking controlled by JSON `"continue"` and `"decision"` fields)
- `1`: Error (unexpected failure)

When secrets are detected, the hook exits 0 and includes `"continue": true` with `"decision": "block"` in the JSON response to signal blocking. The user-facing message is provided in the `"reason"` and `"systemMessage"` fields.

## Logging

By default, all logs go to stderr in JSON format (configurable to text). Stdout is reserved exclusively for hook framework communication.

Example log entry:
```json
{"time":"2025-10-15T10:30:00Z","level":"INFO","msg":"scan completed","has_findings":true,"finding_count":1,"duration":"1.2s"}
```

### Debugging Hooks with File Logging

When running as a hook in an AI agent framework (like Claude Code), stderr output may not be easily accessible. To debug hook execution, enable file-based logging:

**Configuration** (`~/.agent-hooks/vault-radar/config.yaml`):
```yaml
logging:
  level: "debug"  # Enable debug logging
  format: "json"
  log_file: "~/.agent-hooks/vault-radar/logs/hook.log"  # Log to file
```

**Or use environment variable**:
```bash
export HOOK_VAULT_RADAR_LOGGING_LOG_FILE="~/.agent-hooks/vault-radar/logs/hook.log"
export HOOK_VAULT_RADAR_LOGGING_LEVEL="debug"
```

**Monitor logs in real-time**:
```bash
tail -f ~/.agent-hooks/vault-radar/logs/hook.log
```

When file logging is enabled, logs are written to BOTH stderr and the specified file. The debug level includes:
- Raw stdin input received from the hook framework
- Detailed parsing and scanning information
- Complete decision-making process
- Any errors encountered during execution

## Security Considerations

- Vault Radar CLI must be properly configured with valid credentials
- Temporary files are created in secure temp directories and cleaned up
- All sensitive content is written to files with 0600 permissions
- Logs may contain information about findings (not the secrets themselves)

## Future Enhancements

- Additional hook framework support (OpenAI Codex, Gemini CLI, AWS Strands SDK, etc.)
- Custom policy rules and severity mapping

