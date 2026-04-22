package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-template/internal/generator"
	"github.com/DavDaz/llm-wiki-template/internal/manifest"
)

func defaultCfg(dir string) generator.InitConfig {
	return generator.InitConfig{
		ParentDir:       dir,
		Name:            "Legal Wiki",
		Slug:            "legal-wiki",
		Language:        "es",
		ClaudeCode:      true,
		PrimaryEntities: []string{"usuario", "rol"},
		PageTypes:       []string{"proceso", "referencia", "entidad"},
		Conventions:     []string{"Citar fuente en fuentes:"},
	}
}

func TestInit_ClaudeCodeOnly(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultCfg(dir)
	cfg.ClaudeCode = true
	cfg.OpenCode = false
	cfg.Pi = false

	wikiRoot, err := generator.Init(cfg)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "legal-wiki-wiki"), wikiRoot)

	// Core structure.
	assertExists(t, wikiRoot)
	assertExists(t, filepath.Join(wikiRoot, "raw"))
	assertExists(t, filepath.Join(wikiRoot, "wiki", "index.md"))
	assertExists(t, filepath.Join(wikiRoot, "wiki", "log.md"))
	assertExists(t, filepath.Join(wikiRoot, "wiki.toml"))
	assertExists(t, filepath.Join(wikiRoot, ".gitignore"))

	// Claude-only files.
	assertExists(t, filepath.Join(wikiRoot, "CLAUDE.md"))
	assertExists(t, filepath.Join(wikiRoot, ".claude", "commands", "wiki-ingest.md"))
	assertExists(t, filepath.Join(wikiRoot, ".claude", "commands", "wiki-query.md"))
	assertExists(t, filepath.Join(wikiRoot, ".claude", "commands", "wiki-lint.md"))

	// No other tool files.
	assertAbsent(t, filepath.Join(wikiRoot, "AGENTS.md"))
	assertAbsent(t, filepath.Join(wikiRoot, ".opencode"))
	assertAbsent(t, filepath.Join(wikiRoot, ".pi"))

	// Manifest is valid and reflects config.
	m, err := manifest.Load(wikiRoot)
	require.NoError(t, err)
	assert.Equal(t, "Legal Wiki", m.Wiki.Name)
	assert.Equal(t, "legal-wiki", m.Wiki.Slug)
	assert.True(t, m.Tools.ClaudeCode)
	assert.False(t, m.Tools.OpenCode)
	assert.False(t, m.Tools.Pi)
	assert.Equal(t, []string{"usuario", "rol"}, m.Domain.PrimaryEntities)

	// CLAUDE.md must contain the wiki name and slug.
	claudeContent := readFile(t, filepath.Join(wikiRoot, "CLAUDE.md"))
	assert.Contains(t, claudeContent, "Legal Wiki")
	assert.Contains(t, claudeContent, "legal-wiki")
	assert.Contains(t, claudeContent, "usuario")

	// log.md must mention the domain.
	logContent := readFile(t, filepath.Join(wikiRoot, "wiki", "log.md"))
	assert.Contains(t, logContent, "legal-wiki")

	// index.md must have the table header.
	indexContent := readFile(t, filepath.Join(wikiRoot, "wiki", "index.md"))
	assert.Contains(t, indexContent, "| Página |")
}

func TestInit_AllTools(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultCfg(dir)
	cfg.ClaudeCode = true
	cfg.OpenCode = true
	cfg.Pi = true

	_, err := generator.Init(cfg)
	require.NoError(t, err)

	wikiRoot := filepath.Join(dir, "legal-wiki-wiki")
	assertExists(t, filepath.Join(wikiRoot, "CLAUDE.md"))
	assertExists(t, filepath.Join(wikiRoot, "AGENTS.md"))
	assertExists(t, filepath.Join(wikiRoot, ".claude", "commands", "wiki-ingest.md"))
	assertExists(t, filepath.Join(wikiRoot, ".opencode", "commands", "wiki-ingest.md"))
	assertExists(t, filepath.Join(wikiRoot, ".pi", "prompts", "wiki-ingest.md"))
}

func TestInit_DuplicateDirError(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultCfg(dir)

	_, err := generator.Init(cfg)
	require.NoError(t, err)

	// Second init with same slug should fail.
	_, err = generator.Init(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInit_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultCfg(dir)
	cfg.ClaudeCode = false // no tools → validation error
	cfg.OpenCode = false
	cfg.Pi = false

	_, err := generator.Init(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one tool")
}

func TestMigrate_AddTool(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultCfg(dir)
	cfg.ClaudeCode = true

	wikiRoot, err := generator.Init(cfg)
	require.NoError(t, err)
	assertAbsent(t, filepath.Join(wikiRoot, "AGENTS.md"))

	// Enable OpenCode via manifest + migrate.
	m, err := manifest.Load(wikiRoot)
	require.NoError(t, err)
	m.Tools.OpenCode = true
	require.NoError(t, m.Save(wikiRoot))
	require.NoError(t, generator.Migrate(wikiRoot, m))

	assertExists(t, filepath.Join(wikiRoot, "AGENTS.md"))
	assertExists(t, filepath.Join(wikiRoot, ".opencode", "commands", "wiki-ingest.md"))
}

func TestMigrate_RemoveTool(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultCfg(dir)
	cfg.ClaudeCode = true
	cfg.OpenCode = true

	wikiRoot, err := generator.Init(cfg)
	require.NoError(t, err)

	m, err := manifest.Load(wikiRoot)
	require.NoError(t, err)
	m.Tools.OpenCode = false
	require.NoError(t, m.Save(wikiRoot))
	require.NoError(t, generator.Migrate(wikiRoot, m))

	assertAbsent(t, filepath.Join(wikiRoot, ".opencode"))
	assertAbsent(t, filepath.Join(wikiRoot, "AGENTS.md")) // no Pi either
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func assertExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, path+" should exist")
}

func assertAbsent(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), path+" should not exist")
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}
