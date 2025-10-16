package framework

import (
	"fmt"
	"sync"
)

var (
	frameworks = make(map[string]HookFramework)
	mu         sync.RWMutex
)

// RegisterFramework registers a hook framework implementation
func RegisterFramework(name string, framework HookFramework) {
	mu.Lock()
	defer mu.Unlock()
	frameworks[name] = framework
}

// GetFramework retrieves a registered framework by name
func GetFramework(name string) (HookFramework, error) {
	mu.RLock()
	defer mu.RUnlock()

	framework, ok := frameworks[name]
	if !ok {
		return nil, fmt.Errorf("framework %q not registered", name)
	}

	return framework, nil
}

// ListFrameworks returns a list of all registered framework names
func ListFrameworks() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(frameworks))
	for name := range frameworks {
		names = append(names, name)
	}

	return names
}
