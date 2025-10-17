package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/internal/decision"
	"github.com/leefowlercu/agent-hook-vault-radar/internal/framework"
	"github.com/leefowlercu/agent-hook-vault-radar/internal/framework/claude"
	"github.com/leefowlercu/agent-hook-vault-radar/internal/remediation"
	"github.com/leefowlercu/agent-hook-vault-radar/internal/remediation/strategies"
	"github.com/leefowlercu/agent-hook-vault-radar/internal/scanner"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// Processor orchestrates the entire hook processing flow
type Processor struct {
	cfg               *config.Config
	logger            *slog.Logger
	scanner           scanner.Scanner
	decisionEngine    *decision.Engine
	remediationEngine *remediation.Engine
}

// NewProcessor creates a new processor instance
func NewProcessor(cfg *config.Config, logger *slog.Logger) *Processor {
	// Create remediation engine
	remediationEngine := remediation.NewEngine(cfg, logger)

	// Register strategies
	// Note: Strategy configs come from cfg.Remediation.Protocols[].Strategies
	// We register strategy types here, and they'll be instantiated with config at execution time
	registerRemediationStrategies(remediationEngine, cfg, logger)

	return &Processor{
		cfg:               cfg,
		logger:            logger,
		scanner:           scanner.NewVaultRadarScanner(cfg, logger),
		decisionEngine:    decision.NewEngine(cfg),
		remediationEngine: remediationEngine,
	}
}

// registerRemediationStrategies registers all available remediation strategies
func registerRemediationStrategies(engine *remediation.Engine, cfg *config.Config, logger *slog.Logger) {
	// Iterate through all protocol strategies and register them
	for _, protocol := range cfg.Remediation.Protocols {
		for _, strategyCfg := range protocol.Strategies {
			switch strategyCfg.Type {
			case "log":
				logStrategy, err := strategies.NewLogStrategy(strategyCfg)
				if err != nil {
					logger.Warn("failed to create log strategy", "error", err)
					continue
				}
				if err := engine.RegisterStrategy(logStrategy); err != nil {
					logger.Warn("failed to register log strategy", "error", err)
				}
			default:
				logger.Warn("unknown strategy type", "type", strategyCfg.Type)
			}
		}
	}
}

// Process is the main entry point that reads from stdin and writes to stdout
func Process(stdin io.Reader, stdout io.Writer, frameworkName string) error {
	// Load configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration; %w", err)
	}

	// Setup logger
	logger := setupLogger(cfg)

	// Create processor
	proc := NewProcessor(cfg, logger)

	// Process the hook
	ctx := context.Background()
	return proc.ProcessHook(ctx, stdin, stdout, frameworkName)
}

// ProcessHook processes a single hook invocation
func (p *Processor) ProcessHook(ctx context.Context, stdin io.Reader, stdout io.Writer, frameworkName string) error {
	p.logger.Info("processing hook request", "framework", frameworkName)

	// Register frameworks
	framework.RegisterFramework("claude", claude.NewFramework())

	// Get the specified framework
	fw, err := framework.GetFramework(frameworkName)
	if err != nil {
		available := framework.ListFrameworks()
		return fmt.Errorf("failed to get framework %q; available frameworks: %v", frameworkName, available)
	}

	// Read stdin into buffer so we can still parse it
	rawInput, err := io.ReadAll(stdin)
	if err != nil {
		p.logger.Error("failed to read stdin", "error", err)
		return fmt.Errorf("failed to read stdin; %w", err)
	}

	// Parse input from the buffer
	hookInput, err := fw.ParseInput(bytes.NewReader(rawInput))
	if err != nil {
		p.logger.Error("failed to parse input", "error", err)
		return fmt.Errorf("failed to parse input; %w", err)
	}

	p.logger.Info("parsed hook input",
		"framework", hookInput.Framework,
		"hook_type", hookInput.HookType)

	// Get the appropriate handler - use a type-safe approach
	var handler framework.HookHandler

	// Type switch for framework-specific handling
	switch f := fw.(type) {
	case *claude.Framework:
		handler, err = f.GetHandler(hookInput)
		if err != nil {
			p.logger.Error("failed to get handler", "error", err)
			return fmt.Errorf("failed to get handler; %w", err)
		}
	default:
		return fmt.Errorf("unsupported framework type: %T", fw)
	}

	p.logger.Debug("using handler", "type", handler.GetType())

	// Extract content to scan
	content, err := handler.ExtractContent(ctx, hookInput)
	if err != nil {
		p.logger.Error("failed to extract content", "error", err)
		return fmt.Errorf("failed to extract content; %w", err)
	}

	p.logger.Debug("extracted content",
		"type", content.Type,
		"length", len(content.Content))

	// Scan content
	scanResults, err := p.scanner.Scan(ctx, content)
	if err != nil {
		p.logger.Error("scan failed", "error", err)
		// Continue with error in results
	}

	p.logger.Info("scan completed",
		"has_findings", scanResults.HasFindings,
		"finding_count", len(scanResults.Findings),
		"duration", scanResults.ScanDuration)

	// Make decision using the decision engine (framework-agnostic)
	finalDecision, err := p.decisionEngine.Evaluate(ctx, scanResults)
	if err != nil {
		p.logger.Error("failed to make decision", "error", err)
		return fmt.Errorf("failed to make decision; %w", err)
	}

	p.logger.Info("decision made",
		"block", finalDecision.Block)

	// Execute remediation if enabled
	remediationInput := types.RemediationInput{
		ScanResults: scanResults,
		HookInput:   hookInput,
		Decision:    finalDecision,
		Timestamp:   time.Now(),
		Framework:   frameworkName,
	}

	remediationResults := p.remediationEngine.Execute(ctx, remediationInput)

	// Enrich decision message with remediation results
	if remediationResults.Executed {
		p.logger.Info("remediation executed",
			"protocol", remediationResults.ProtocolName,
			"strategies", len(remediationResults.Results),
			"duration", remediationResults.TotalDuration)

		decision.EnrichWithRemediation(&finalDecision, remediationResults)
	}

	// Format output
	output, err := fw.FormatOutput(finalDecision, hookInput)
	if err != nil {
		p.logger.Error("failed to format output", "error", err)
		return fmt.Errorf("failed to format output; %w", err)
	}

	// Write output to stdout
	if _, err := stdout.Write(output); err != nil {
		p.logger.Error("failed to write output", "error", err)
		return fmt.Errorf("failed to write output; %w", err)
	}

	// Add newline for cleaner output
	if _, err := stdout.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline; %w", err)
	}

	p.logger.Info("hook processing completed successfully")

	// Get exit code from framework (framework determines exit code semantics)
	exitCode := fw.GetExitCode(finalDecision)
	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// setupLogger creates and configures the logger based on configuration
// Logs are written to file only (not stderr) to avoid interfering with hook framework IO
func setupLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Determine output writer - file only, no stderr
	var output io.Writer

	if cfg.Logging.LogFile != "" {
		logFile, err := openLogFile(cfg.Logging.LogFile)
		if err != nil {
			// Critical error during startup - write to stderr and use discard
			fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", cfg.Logging.LogFile, err)
			output = io.Discard
		} else {
			output = logFile
		}
	} else {
		// No log file configured - disable logging
		output = io.Discard
	}

	var handler slog.Handler
	if cfg.Logging.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}

// openLogFile opens or creates a log file for writing
func openLogFile(path string) (*os.File, error) {
	// Expand ~ to home directory if present
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory; %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory; %w", err)
	}

	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file; %w", err)
	}

	return file, nil
}
