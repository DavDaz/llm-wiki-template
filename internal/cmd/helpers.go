package cmd

import (
	"fmt"
	"os"

	"github.com/DavDaz/llm-wiki-template/internal/manifest"
)

// loadManifestFromCwd loads wiki.toml from the current working directory.
// Returns the manifest and the wiki root path.
func loadManifestFromCwd() (*manifest.Manifest, string, error) {
	wikiRoot, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("get working dir: %w", err)
	}
	m, err := manifest.Load(wikiRoot)
	if err != nil {
		return nil, "", fmt.Errorf("no wiki found in current directory (wiki.toml not found): %w", err)
	}
	return m, wikiRoot, nil
}

// setToolEnabled toggles a tool on/off in the manifest by name.
func setToolEnabled(m *manifest.Manifest, name string, enabled bool) {
	switch name {
	case "claude-code":
		m.Tools.ClaudeCode = enabled
	case "opencode":
		m.Tools.OpenCode = enabled
	case "pi":
		m.Tools.Pi = enabled
	}
}
