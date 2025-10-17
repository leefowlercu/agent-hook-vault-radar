# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.0.0] - 2025-10-17

### Added
- Remediation engine subsystem with concurrent strategy execution
- Protocol system with configurable triggers (on_block, on_findings, severity_threshold, finding_types)
- Log remediation strategy supporting JSON and text output formats
- Decision engine message enrichment with remediation results
- Comprehensive test suite for remediation strategies (13 test cases for log strategy)
- Unit tests for decision engine and message formatting (decision_test.go)
- Template variable system for future remediation strategies
- Success/failure indicators (✓/✗) in user-facing messages
- Configuration example file (config.yaml.example)

### Changed
- Configuration structure to include remediation settings with protocols and strategies
- README documentation significantly expanded with remediation system documentation
- Processor architecture to integrate remediation engine
- Decision engine to support message enrichment with remediation results

## [2.0.0] - 2025-10-16

### Changed
- **BREAKING**: Decision engine responsibilities refactored to be framework-agnostic
- **BREAKING**: Framework handler architecture updated to clarify separation of concerns
- Framework interface to remove framework-specific logic from processor layer

### Removed
- Framework-specific decision logic from processor (moved to framework implementations)

## [1.1.0] - 2025-10-15

### Added
- File-based logging system supporting JSON and text formats
- Configurable log levels (debug, info, warn, error)
- Log file path configuration with tilde (~) expansion
- Automatic log directory creation
- File-only logging to avoid interfering with hook framework IO

### Changed
- Logging configuration structure to include log_file, level, and format options
- Default logging behavior to use file output instead of stderr
- README with comprehensive logging documentation

## [1.0.0] - 2025-10-15

### Added
- Core hook framework architecture with plugin-based design
- Claude Code framework integration
- UserPromptSubmit hook handler for Claude Code
- Vault Radar CLI scanner integration
- Decision engine with configurable severity thresholds
- Configuration management using Viper
- Support for environment variable configuration (.env files)
- Command-line interface using Cobra
- Multi-level configuration precedence (defaults, .env, YAML, env vars, CLI flags)
- Severity threshold filtering (critical, high, medium/info, low)
- Test fixtures for Claude Code hooks
- Makefile with build, test, and release automation
- Cross-platform release builds (darwin, linux, windows for amd64/arm64)
- Comprehensive README documentation

[unreleased]: https://github.com/leefowlercu/agent-hook-vault-radar/compare/v3.0.0...HEAD
[3.0.0]: https://github.com/leefowlercu/agent-hook-vault-radar/compare/v2.0.0...v3.0.0
[2.0.0]: https://github.com/leefowlercu/agent-hook-vault-radar/compare/v1.1.0...v2.0.0
[1.1.0]: https://github.com/leefowlercu/agent-hook-vault-radar/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/leefowlercu/agent-hook-vault-radar/releases/tag/v1.0.0
