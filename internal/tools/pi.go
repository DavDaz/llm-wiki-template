package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
)

const piPromptsDir = ".pi/prompts"
const piPromptsTree = "└── .pi/prompts/\n    ├── wiki-ingest.md\n    ├── wiki-query.md\n    └── wiki-lint.md"

// PiTool implements ToolSupport for Pi.
type PiTool struct{}

func (PiTool) Name() string { return "pi" }

// SharedFiles declares AGENTS.md as shared with OpenCode.
func (PiTool) SharedFiles() []string { return []string{"AGENTS.md"} }

func (PiTool) IsInstalled(wikiRoot string) bool {
	_, err := os.Stat(filepath.Join(wikiRoot, piPromptsDir))
	return err == nil
}

func (PiTool) Install(wikiRoot string, m *manifest.Manifest) error {
	promptsDir := filepath.Join(wikiRoot, piPromptsDir)
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", promptsDir, err)
	}

	content, err := renderSchema(m, piPromptsDir, piPromptsTree, "AGENTS.md")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(wikiRoot, "AGENTS.md"), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write AGENTS.md: %w", err)
	}

	return copyCommandFiles(promptsDir)
}

func (PiTool) Uninstall(wikiRoot string, m *manifest.Manifest) error {
	if err := os.RemoveAll(filepath.Join(wikiRoot, ".pi")); err != nil {
		return fmt.Errorf("remove .pi: %w", err)
	}
	// Only remove AGENTS.md if OpenCode is also disabled — they share the file.
	if !m.Tools.OpenCode {
		if err := os.Remove(filepath.Join(wikiRoot, "AGENTS.md")); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove AGENTS.md: %w", err)
		}
	}
	return nil
}
