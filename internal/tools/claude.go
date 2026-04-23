package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
	"github.com/DavDaz/llm-wiki-generator/internal/templates"
)

const claudeSkillsDir = ".claude/skills"
const claudeCommandsLegacyDir = ".claude/commands" // migrated away from this format

const claudeSkillsTree = "└── .claude/skills/\n    ├── wiki-ingest/SKILL.md\n    ├── wiki-query/SKILL.md\n    └── wiki-lint/SKILL.md"

// ClaudeTool implements ToolSupport for Claude Code.
type ClaudeTool struct{}

func (ClaudeTool) Name() string { return "claude-code" }

func (ClaudeTool) SharedFiles() []string { return nil } // CLAUDE.md is exclusive

func (ClaudeTool) IsInstalled(wikiRoot string) bool {
	_, err := os.Stat(filepath.Join(wikiRoot, "CLAUDE.md"))
	return err == nil
}

func (ClaudeTool) Install(wikiRoot string, m *manifest.Manifest) error {
	// Remove legacy commands dir if present (migration from old format).
	legacyDir := filepath.Join(wikiRoot, claudeCommandsLegacyDir)
	if _, err := os.Stat(legacyDir); err == nil {
		if err := os.RemoveAll(legacyDir); err != nil {
			return fmt.Errorf("remove legacy commands dir: %w", err)
		}
	}

	content, err := renderSchema(m, claudeSkillsDir, claudeSkillsTree, "CLAUDE.md")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(wikiRoot, "CLAUDE.md"), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write CLAUDE.md: %w", err)
	}

	return copySkills(filepath.Join(wikiRoot, claudeSkillsDir))
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
func renderSchema(m *manifest.Manifest, commandsDir, commandsTree, instructionsFile string) (string, error) {
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
		CommandsTree:     commandsTree,
		InstructionsFile: instructionsFile,
	}
	return templates.RenderSchema(data)
}

// copyCommandFiles copies the three wiki command files as flat .md files into destDir.
// Used by OpenCode and Pi which follow a flat commands directory convention.
func copyCommandFiles(destDir string) error {
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

// copySkills installs the three wiki commands as Claude Code skills (new format).
// Each command becomes .claude/skills/<name>/SKILL.md.
func copySkills(skillsDir string) error {
	for _, name := range []string{"wiki-ingest", "wiki-query", "wiki-lint"} {
		skillDir := filepath.Join(skillsDir, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", skillDir, err)
		}
		data, err := templates.ReadFile("commands/" + name + ".md")
		if err != nil {
			return fmt.Errorf("read embedded command %s: %w", name, err)
		}
		dest := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
	}
	return nil
}
