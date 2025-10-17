# agent-hook-vault-radar

A Go program that integrates AI agent hook frameworks with HashiCorp Vault Radar for secret scanning. It bridges the gap between interactive AI coding assistants and security scanning tools to prevent accidental exposure of sensitive information.

## Overview

`hook-vault-radar` reads hook data from stdin (as JSON), scans the content using Vault Radar CLI, makes intelligent decisions based on findings, and outputs responses (as JSON) to stdout. This allows AI agent frameworks like Claude Code to automatically scan user prompts and code for secrets before processing.

## Features

- **Extensible Framework Support**: Plugin-based architecture supports multiple hook frameworks
- **Claude Code Integration**: Built-in support for Claude Code's UserPromptSubmit hook
- **Vault Radar Integration**: Leverages HashiCorp Vault Radar CLI for enterprise-grade secret detection
- **Configurable Policies**: Customizable severity thresholds and blocking behavior
- **Remediation System**: Automatic actions when secrets detected (logging, webhooks, etc.) - opt-in feature
- **Concurrent Strategy Execution**: Parallel remediation for optimal performance
- **File-Only Logging**: JSON or text logging to file (avoids interfering with hook framework IO)
- **Single Binary**: Self-contained executable requiring only vault-radar CLI

## Architecture

```
┌────────────────────────────────────────────────────────┐
│                hook-vault-radar CLI                    │
├────────────────────────────────────────────────────────┤
│                                                        │
│  stdin (JSON) → Framework Parser → Content Extractor   │
│                         ↓                              │
│                Vault Radar Scanner                     │
│                         ↓                              │
│                 Decision Engine                        │
│                         ↓                              │
│              Remediation Engine (optional)             │
│                         ↓                              │
│          Response Formatter → stdout (JSON)            │
│                                                        │
└────────────────────────────────────────────────────────┘
```

### Components

- **Framework Layer** (`internal/framework/`): Abstracts different hook frameworks (Claude, etc.)
- **Scanner Layer** (`internal/scanner/`): Interfaces with security scanning tools (Vault Radar)
- **Decision Engine** (`internal/decision/`): Applies policies and makes blocking decisions
- **Remediation Engine** (`internal/remediation/`): Executes automatic actions when secrets detected (logging, webhooks, etc.)
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

### Custom Configuration File

You can specify a custom configuration file using the `--config` flag:

```bash
# Use a custom config file (absolute path)
cat input.json | ./hook-vault-radar --framework claude --config /path/to/custom-config.yaml

# Use a custom config file (relative path)
cat input.json | ./hook-vault-radar --framework claude --config ./configs/dev.yaml

# Use a custom config file (with ~ expansion)
cat input.json | ./hook-vault-radar --framework claude --config ~/my-configs/prod.yaml
```

**Default behavior** (without `--config` flag):
- Searches for `config.yaml` in:
  1. `~/.agent-hooks/vault-radar/`
  2. Current directory (`.`)
- Gracefully handles missing config files (uses defaults)

**With `--config` flag**:
- Uses the specified file directly
- Returns an error if the file doesn't exist
- Useful for:
  - Development vs production environments
  - Testing different configurations
  - Multi-tenant scenarios
  - CI/CD pipelines with environment-specific configs

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

**Note**: The hook framework (e.g., `claude`) is specified via the `--framework` CLI flag, not in the configuration file.

```yaml
vault_radar:
  command: "vault-radar"
  scan_command: "scan file"
  timeout_seconds: 30
  extra_args: []  # Additional vault-radar flags (--disable-ui is always included automatically)

logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json or text
  log_file: "~/.agent-hooks/vault-radar/logs/hook.log"  # Required for logging (empty = disabled)

decision:
  block_on_findings: true
  severity_threshold: "medium" # critical, high, medium, low

remediation:
  enabled: false  # Opt-in feature (default: false)
  timeout_seconds: 10
  protocols: []  # See Remediation System section for configuration
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

## Remediation System

The remediation subsystem enables automatic actions when secrets are detected. This is an **opt-in feature** (disabled by default) that executes configured strategies concurrently when security findings match specific triggers.

**Key Features**:
- **Informational Only**: Remediation results (✓/✗) are displayed in user messages but never affect the security blocking decision
- **Concurrent Execution**: Strategies run in parallel using goroutines for optimal performance
- **Configurable Triggers**: Execute on blocking, all findings, by severity, or by finding type
- **Strategy System**: Pluggable architecture for different remediation actions

### Configuration

Enable remediation in `~/.agent-hooks/vault-radar/config.yaml`:

```yaml
remediation:
  enabled: true  # Opt-in (default: false)
  timeout_seconds: 10  # Overall timeout for all strategies

  protocols:
    - name: "log-blocked-secrets"
      triggers:
        on_block: true  # Execute when blocking occurs
        on_findings: false  # Don't execute for non-blocking findings
        severity_threshold: "medium"  # Minimum severity to trigger
        # finding_types: ["aws_*", "github_*"]  # Optional: filter by type patterns
      strategies:
        - type: "log"
          config:
            log_file: "~/.agent-hooks/vault-radar/logs/findings.log"
            format: "json"  # or "text"
```

### Available Strategies

#### Log Strategy (Implemented)
Writes finding details to a file in JSON or text format.

**Configuration**:
```yaml
- type: "log"
  config:
    log_file: "~/.agent-hooks/vault-radar/logs/findings.log"  # Required
    format: "json"  # "json" or "text"
```

**JSON Format Output**:
```json
{
  "timestamp": "2025-10-17T10:30:00Z",
  "framework": "claude",
  "session_id": "abc123",
  "blocked": true,
  "finding_count": 2,
  "findings": [
    {
      "severity": "info",
      "type": "aws_access_key_id",
      "location": "scan-content.txt",
      "description": "AWS access key ID"
    }
  ]
}
```

**Text Format Output**:
```
[2025-10-17 10:30:00] Framework: claude | Session: abc123 | Findings: 2 | Blocked: true
  - [INFO] aws_access_key_id: AWS access key ID (scan-content.txt)
  - [MEDIUM] aws_secret_key: AWS secret key (scan-content.txt)
```

#### Planned Strategies (Phase 3)
- **Webhook**: Send HTTP POST notifications to external systems
- **Vault KVv2**: Store metadata in HashiCorp Vault
- **Slack**: Send alerts to Slack channels

### Triggers

Control when remediation protocols execute:

| Trigger | Description | Example |
|---------|-------------|---------|
| `on_block` | Execute when blocking occurs | `true` |
| `on_findings` | Execute whenever findings exist (even if not blocking) | `false` |
| `severity_threshold` | Minimum severity to trigger | `"medium"` |
| `finding_types` | Filter by finding type patterns (supports wildcards) | `["aws_*", "github_*"]` |

**Multiple triggers are AND-ed together** - all conditions must be true for the protocol to execute.

### User Message Enrichment

When remediation executes, results are appended to the user-facing message:

```
Vault Radar detected 2 security findings:

1. [INFO] aws_access_key_id: AWS access key ID (scan-content.txt)
2. [MEDIUM] aws_secret_key: AWS secret key (scan-content.txt)

Remediation actions taken (1 strategy, 12ms total):
  ✓ Logged 2 findings to findings.log (12ms)

Please remove or redact sensitive information before proceeding.
```

Success indicators:
- ✓ (U+2713) - Strategy executed successfully
- ✗ (U+2717) - Strategy failed (error details included)

### Template Variables

Future strategies (webhook, vault) will support template variables:

- `.Date` - YYYY-MM-DD format (e.g., "2025-10-17")
- `.Time` - HH:MM:SS format (e.g., "10:30:45")
- `.Timestamp` - Unix timestamp
- `.Type` - Finding type (e.g., "aws_access_key_id")
- `.Severity` - Finding severity
- `.SessionID` - Hook session ID
- `.Framework` - Framework name (e.g., "claude")
- `.Location` - Where the secret was found
- `.Count` - Number of findings detected

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
            "command": "/<path>/<to>/hook-vault-radar --framework claude",
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

# With custom config file
cat testdata/claude/userpromptsubmit.json | ./hook-vault-radar --framework claude --config /path/to/custom-config.yaml

# With custom config and debug logging
cat testdata/claude/userpromptsubmit.json | ./hook-vault-radar --framework claude --config ~/my-configs/dev.yaml --log-level debug

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
│   │   ├── decision.go                  # Policy-based decision making
│   │   └── decision_test.go             # Decision engine tests
│   ├── remediation/                     # Remediation subsystem (opt-in)
│   │   ├── remediation.go               # Engine with concurrent execution
│   │   ├── protocol.go                  # Protocol and trigger logic
│   │   ├── registry.go                  # Strategy registry
│   │   └── strategies/                  # Strategy implementations
│   │       ├── log.go                   # Log strategy
│   │       └── log_test.go              # Log strategy tests
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

#### Unit Tests

Run unit tests with coverage:

```bash
# Run all tests with race detection and coverage
make test

# Run tests and display coverage report
make test-coverage

# Run tests directly with go test
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Current test coverage**:
- `internal/decision/` - Decision engine and message enrichment
- `internal/remediation/strategies/` - Log strategy (13 test cases)

#### Integration Tests

Test with sample fixtures:

```bash
# Run all integration tests (clean + secrets)
make test-integration

# Quick test with clean input (no secrets)
make run-test

# Quick test with secret-containing input
make run-test-secret

# Manual testing
cat testdata/claude/userpromptsubmit.json | ./hook-vault-radar --framework claude --log-level debug
cat testdata/claude/userpromptsubmit_clean.json | ./hook-vault-radar --framework claude
```

#### Development Commands

```bash
# Build the binary
make build

# Install to ~/.agent-hooks/vault-radar/
make install

# Format code
make fmt

# Run static analysis
make vet

# Run linter (requires golangci-lint)
make lint

# Run all checks (fmt, vet, lint, test)
make check

# Clean build artifacts
make clean

# Build release binaries for multiple platforms
make release

# Show version information
make version

# Show help
make help
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

All logs are written to a log file only (not stderr) to avoid interfering with hook framework IO expectations. Stdout is reserved exclusively for hook framework communication (JSON responses).

**Default log location**: `~/.agent-hooks/vault-radar/logs/hook.log`

Example log entry:
```json
{"time":"2025-10-15T10:30:00Z","level":"INFO","msg":"scan completed","has_findings":true,"finding_count":1,"duration":"1.2s"}
```

### Configuring Logging

**Configuration** (`~/.agent-hooks/vault-radar/config.yaml`):
```yaml
logging:
  level: "info"   # Logging level: debug, info, warn, error
  format: "json"  # Format: json or text
  log_file: "~/.agent-hooks/vault-radar/logs/hook.log"  # Required for logging
```

**Or use environment variables**:
```bash
export HOOK_VAULT_RADAR_LOGGING_LOG_FILE="~/.agent-hooks/vault-radar/logs/hook.log"
export HOOK_VAULT_RADAR_LOGGING_LEVEL="debug"
export HOOK_VAULT_RADAR_LOGGING_FORMAT="json"
```

**Monitor logs in real-time**:
```bash
tail -f ~/.agent-hooks/vault-radar/logs/hook.log
```

**Important**: If no log file is configured (empty string), logging is disabled entirely. The debug level includes:
- Detailed parsing and scanning information
- Complete decision-making process
- Remediation execution details
- Any errors encountered during execution

## Security Considerations

- Vault Radar CLI must be properly configured with valid credentials
- Temporary files are created in secure temp directories and cleaned up
- All sensitive content is written to files with 0600 permissions
- Logs may contain information about findings (not the secrets themselves)

## Future Enhancements

- Additional hook framework support (OpenAI Codex, Gemini CLI, AWS Strands SDK, etc.)
- Custom policy rules and severity mapping

