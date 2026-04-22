package tools

import "fmt"

var registry = map[string]ToolSupport{
	"claude-code": ClaudeTool{},
	"opencode":    OpenCodeTool{},
	"pi":          PiTool{},
}

// All returns all registered tool implementations.
func All() []ToolSupport {
	return []ToolSupport{ClaudeTool{}, OpenCodeTool{}, PiTool{}}
}

// Get returns the ToolSupport for the given name, or an error if unknown.
func Get(name string) (ToolSupport, error) {
	t, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool %q (valid: claude-code, opencode, pi)", name)
	}
	return t, nil
}
