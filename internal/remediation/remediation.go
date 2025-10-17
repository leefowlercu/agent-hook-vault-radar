package remediation

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// RemediationStrategy defines the interface that all remediation strategies must implement
type RemediationStrategy interface {
	// Execute performs the remediation action and returns the result
	Execute(ctx context.Context, input types.RemediationInput) types.RemediationResult

	// GetType returns the type identifier for this strategy (e.g., "log", "webhook")
	GetType() string

	// Validate checks if the strategy configuration is valid
	Validate() error
}

// Engine orchestrates the execution of remediation protocols
type Engine struct {
	cfg      *config.Config
	logger   *slog.Logger
	registry *Registry
}

// NewEngine creates a new remediation engine
func NewEngine(cfg *config.Config, logger *slog.Logger) *Engine {
	return &Engine{
		cfg:      cfg,
		logger:   logger,
		registry: NewRegistry(),
	}
}

// RegisterStrategy registers a strategy with the engine
func (e *Engine) RegisterStrategy(strategy RemediationStrategy) error {
	return e.registry.RegisterStrategy(strategy)
}

// Execute runs the appropriate remediation protocol based on the decision and findings
func (e *Engine) Execute(ctx context.Context, input types.RemediationInput) types.RemediationResults {
	// Check if remediation is enabled
	if !e.cfg.Remediation.Enabled {
		e.logger.Debug("remediation disabled, skipping")
		return types.RemediationResults{Executed: false}
	}

	// Find the first protocol whose triggers match
	var protocol *Protocol
	for _, protocolCfg := range e.cfg.Remediation.Protocols {
		p := NewProtocol(protocolCfg)
		if p.ShouldExecute(input) {
			protocol = p
			e.logger.Info("matched remediation protocol", "protocol", p.Name)
			break
		}
	}

	if protocol == nil {
		e.logger.Debug("no remediation protocol matched triggers")
		return types.RemediationResults{Executed: false}
	}

	// Execute the protocol
	return e.executeProtocol(ctx, protocol, input)
}

// executeProtocol executes a single protocol with concurrent strategy execution
func (e *Engine) executeProtocol(ctx context.Context, protocol *Protocol, input types.RemediationInput) types.RemediationResults {
	startTime := time.Now()

	// Apply timeout if configured
	timeout := time.Duration(e.cfg.Remediation.TimeoutSeconds) * time.Second
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	strategies := protocol.Strategies
	if len(strategies) == 0 {
		e.logger.Warn("protocol has no strategies", "protocol", protocol.Name)
		return types.RemediationResults{
			Executed:     true,
			Results:      []types.RemediationResult{},
			ProtocolName: protocol.Name,
		}
	}

	// Channel to collect results
	resultChan := make(chan types.RemediationResult, len(strategies))
	var wg sync.WaitGroup

	// Launch all strategies concurrently
	for _, strategyCfg := range strategies {
		strategy, err := e.registry.GetStrategy(strategyCfg.Type)
		if err != nil {
			e.logger.Warn("unknown strategy type", "type", strategyCfg.Type, "error", err)
			// Add a failed result for unknown strategy
			resultChan <- types.RemediationResult{
				StrategyType: strategyCfg.Type,
				Success:      false,
				Message:      fmt.Sprintf("Unknown strategy type: %s", strategyCfg.Type),
				Error:        err,
			}
			continue
		}

		wg.Add(1)
		go e.executeStrategy(ctx, &wg, strategy, input, resultChan)
	}

	// Wait for all strategies to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := e.collectResults(resultChan)
	totalDuration := time.Since(startTime)

	e.logger.Info("remediation protocol completed",
		"protocol", protocol.Name,
		"strategies", len(results),
		"duration", totalDuration)

	return types.RemediationResults{
		Executed:      true,
		Results:       results,
		TotalDuration: totalDuration,
		ProtocolName:  protocol.Name,
	}
}

// executeStrategy runs a single strategy in a goroutine with panic recovery
func (e *Engine) executeStrategy(ctx context.Context, wg *sync.WaitGroup, strategy RemediationStrategy, input types.RemediationInput, resultChan chan<- types.RemediationResult) {
	defer wg.Done()

	// Recover from panics to prevent bringing down the entire remediation
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("strategy panicked", "type", strategy.GetType(), "panic", r)
			resultChan <- types.RemediationResult{
				StrategyType: strategy.GetType(),
				Success:      false,
				Message:      "Strategy panicked during execution",
				Error:        fmt.Errorf("panic: %v", r),
			}
		}
	}()

	strategyType := strategy.GetType()
	e.logger.Debug("executing strategy", "type", strategyType)

	startTime := time.Now()
	result := strategy.Execute(ctx, input)
	result.Duration = time.Since(startTime)
	result.StrategyType = strategyType

	e.logger.Debug("strategy completed",
		"type", strategyType,
		"success", result.Success,
		"duration", result.Duration)

	// Try to send result, but don't block if context is cancelled
	select {
	case resultChan <- result:
	case <-ctx.Done():
		e.logger.Warn("context cancelled while sending result", "type", strategyType)
	}
}

// collectResults gathers all results from the channel
func (e *Engine) collectResults(resultChan <-chan types.RemediationResult) []types.RemediationResult {
	results := make([]types.RemediationResult, 0)
	for result := range resultChan {
		results = append(results, result)
	}
	return results
}
