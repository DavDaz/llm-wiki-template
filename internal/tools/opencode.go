package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavDaz/llm-wiki-template/internal/manifest"
)

const opencodeCommandsDir = ".opencode/commands"

// OpenCodeTool implements ToolSupport for OpenCode.
type OpenCodeTool struct{}

func (OpenCodeTool) Name() string { return "opencode" }

// SharedFiles declares AGENTS.md as shared with Pi — both write to the same file.
func (OpenCodeTool) SharedFiles() []string { return []string{"AGENTS.md"} }

func (OpenCodeTool) IsInstalled(wikiRoot string) bool {
	_, err := os.Stat(filepath.Join(wikiRoot, opencodeCommandsDir))
	return err == nil
}

func (OpenCodeTool) Install(wikiRoot string, m *manifest.Manifest) error {
	cmdsDir := filepath.Join(wikiRoot, opencodeCommandsDir)
	if err := os.MkdirAll(cmdsDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", cmdsDir, err)
	}

	content, err := renderSchema(m, opencodeCommandsDir, "AGENTS.md")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(wikiRoot, "AGENTS.md"), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write AGENTS.md: %w", err)
	}

	return copyCommands(cmdsDir)
}

func (OpenCodeTool) Uninstall(wikiRoot string, m *manifest.Manifest) error {
	if err := os.RemoveAll(filepath.Join(wikiRoot, ".opencode")); err != nil {
		return fmt.Errorf("remove .opencode: %w", err)
	}
	// Only remove AGENTS.md if Pi is also disabled — they share the file.
	if !m.Tools.Pi {
		if err := os.Remove(filepath.Join(wikiRoot, "AGENTS.md")); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove AGENTS.md: %w", err)
		}
	}
	return nil
}
