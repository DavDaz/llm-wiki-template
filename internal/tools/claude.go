package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavDaz/llm-wiki-template/internal/manifest"
	"github.com/DavDaz/llm-wiki-template/internal/templates"
)

const claudeCommandsDir = ".claude/commands"

// ClaudeTool implements ToolSupport for Claude Code.
type ClaudeTool struct{}

func (ClaudeTool) Name() string { return "claude-code" }

func (ClaudeTool) SharedFiles() []string { return nil } // CLAUDE.md is exclusive

func (ClaudeTool) IsInstalled(wikiRoot string) bool {
	_, err := os.Stat(filepath.Join(wikiRoot, "CLAUDE.md"))
	return err == nil
}

func (ClaudeTool) Install(wikiRoot string, m *manifest.Manifest) error {
	cmdsDir := filepath.Join(wikiRoot, claudeCommandsDir)
	if err := os.MkdirAll(cmdsDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", cmdsDir, err)
	}

	content, err := renderSchema(m, claudeCommandsDir, "CLAUDE.md")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(wikiRoot, "CLAUDE.md"), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write CLAUDE.md: %w", err)
	}

	return copyCommands(cmdsDir)
}

func (ClaudeTool) Uninstall(wikiRoot string, _ *manifest.Manifest) error {
	if err := os.RemoveAll(filepath.Join(wikiRoot, ".claude")); err != nil {
		return fmt.Errorf("remove .claude: %w", err)
	}
	if err := os.Remove(filepath.Join(wikiRoot, "CLAUDE.md")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove CLAUDE.md: %w", err)
	}
	return nil
}

// renderSchema builds the schema/instructions file content for a given tool.
func renderSchema(m *manifest.Manifest, commandsDir, instructionsFile string) (string, error) {
	conventions := make([]string, len(m.Domain.Conventions))
	for i, c := range m.Domain.Conventions {
		conventions[i] = c.Rule
	}

	data := templates.SchemaData{
		WikiName:         m.Wiki.Name,
		WikiSlug:         m.Wiki.Slug,
		Language:         m.Wiki.Language,
		CreatedDate:      m.Wiki.CreatedAt,
		PrimaryEntities:  m.Domain.PrimaryEntities,
		PageTypes:        m.Domain.PageTypes,
		Conventions:      conventions,
		CommandsDir:      commandsDir,
		InstructionsFile: instructionsFile,
	}
	return templates.RenderSchema(data)
}

// copyCommands copies the three wiki skill files into destDir.
func copyCommands(destDir string) error {
	for _, name := range []string{"wiki-ingest.md", "wiki-query.md", "wiki-lint.md"} {
		data, err := templates.ReadFile("commands/" + name)
		if err != nil {
			return fmt.Errorf("read embedded command %s: %w", name, err)
		}
		dest := filepath.Join(destDir, name)
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
	}
	return nil
}
