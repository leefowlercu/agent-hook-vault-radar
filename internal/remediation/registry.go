package remediation

import (
	"fmt"
	"sync"
)

// Registry manages available remediation strategies
type Registry struct {
	strategies map[string]RemediationStrategy
	mu         sync.RWMutex
}

// NewRegistry creates a new strategy registry
func NewRegistry() *Registry {
	return &Registry{
		strategies: make(map[string]RemediationStrategy),
	}
}

// RegisterStrategy adds a strategy to the registry
func (r *Registry) RegisterStrategy(strategy RemediationStrategy) error {
	if strategy == nil {
		return fmt.Errorf("strategy cannot be nil")
	}

	strategyType := strategy.GetType()
	if strategyType == "" {
		return fmt.Errorf("strategy type cannot be empty")
	}

	// Validate the strategy configuration
	if err := strategy.Validate(); err != nil {
		return fmt.Errorf("strategy validation failed: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.strategies[strategyType]; exists {
		return fmt.Errorf("strategy type %q is already registered", strategyType)
	}

	r.strategies[strategyType] = strategy
	return nil
}

// GetStrategy retrieves a strategy by type
func (r *Registry) GetStrategy(strategyType string) (RemediationStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	strategy, exists := r.strategies[strategyType]
	if !exists {
		return nil, fmt.Errorf("strategy type %q not found", strategyType)
	}

	return strategy, nil
}

// ListStrategies returns all registered strategy types
func (r *Registry) ListStrategies() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.strategies))
	for strategyType := range r.strategies {
		types = append(types, strategyType)
	}

	return types
}

// HasStrategy checks if a strategy type is registered
func (r *Registry) HasStrategy(strategyType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.strategies[strategyType]
	return exists
}

// UnregisterStrategy removes a strategy from the registry
func (r *Registry) UnregisterStrategy(strategyType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.strategies[strategyType]; !exists {
		return fmt.Errorf("strategy type %q not found", strategyType)
	}

	delete(r.strategies, strategyType)
	return nil
}
