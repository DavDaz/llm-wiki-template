package tools_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-template/internal/manifest"
	"github.com/DavDaz/llm-wiki-template/internal/tools"
)

func baseManifest() *manifest.Manifest {
	m := manifest.New("Test Wiki", "test-wiki", "es")
	m.Domain.PrimaryEntities = []string{"usuario", "rol"}
	m.Domain.PageTypes = []string{"proceso", "referencia"}
	return m
}

// ─── Claude Code ─────────────────────────────────────────────────────────────

func TestClaudeTool_InstallUninstall(t *testing.T) {
	dir := t.TempDir()
	m := baseManifest()
	m.Tools.ClaudeCode = true

	tool := tools.ClaudeTool{}
	assert.False(t, tool.IsInstalled(dir))

	require.NoError(t, tool.Install(dir, m))
	assert.True(t, tool.IsInstalled(dir))

	// Verify expected files exist.
	assertFileExists(t, filepath.Join(dir, "CLAUDE.md"))
	assertFileExists(t, filepath.Join(dir, ".claude/commands/wiki-ingest.md"))
	assertFileExists(t, filepath.Join(dir, ".claude/commands/wiki-query.md"))
	assertFileExists(t, filepath.Join(dir, ".claude/commands/wiki-lint.md"))

	// CLAUDE.md must contain the wiki name.
	assertFileContains(t, filepath.Join(dir, "CLAUDE.md"), "Test Wiki")

	// Idempotent install.
	require.NoError(t, tool.Install(dir, m))

	// Uninstall.
	m.Tools.ClaudeCode = false
	require.NoError(t, tool.Uninstall(dir, m))
	assert.False(t, tool.IsInstalled(dir))
	assertFileAbsent(t, filepath.Join(dir, "CLAUDE.md"))
	assertFileAbsent(t, filepath.Join(dir, ".claude"))
}

// ─── OpenCode ────────────────────────────────────────────────────────────────

func TestOpenCodeTool_InstallUninstall(t *testing.T) {
	dir := t.TempDir()
	m := baseManifest()
	m.Tools.OpenCode = true

	tool := tools.OpenCodeTool{}
	require.NoError(t, tool.Install(dir, m))

	assertFileExists(t, filepath.Join(dir, "AGENTS.md"))
	assertFileExists(t, filepath.Join(dir, ".opencode/commands/wiki-ingest.md"))

	m.Tools.OpenCode = false
	require.NoError(t, tool.Uninstall(dir, m))
	assertFileAbsent(t, filepath.Join(dir, ".opencode"))
	assertFileAbsent(t, filepath.Join(dir, "AGENTS.md"))
}

// ─── Pi ──────────────────────────────────────────────────────────────────────

func TestPiTool_InstallUninstall(t *testing.T) {
	dir := t.TempDir()
	m := baseManifest()
	m.Tools.Pi = true

	tool := tools.PiTool{}
	require.NoError(t, tool.Install(dir, m))

	assertFileExists(t, filepath.Join(dir, "AGENTS.md"))
	assertFileExists(t, filepath.Join(dir, ".pi/prompts/wiki-ingest.md"))

	m.Tools.Pi = false
	require.NoError(t, tool.Uninstall(dir, m))
	assertFileAbsent(t, filepath.Join(dir, ".pi"))
	assertFileAbsent(t, filepath.Join(dir, "AGENTS.md"))
}

// ─── AGENTS.md shared file gotcha ────────────────────────────────────────────

func TestSharedAgentsMd_UninstallOpenCode_PiStillEnabled(t *testing.T) {
	dir := t.TempDir()
	m := baseManifest()
	m.Tools.OpenCode = true
	m.Tools.Pi = true

	require.NoError(t, tools.OpenCodeTool{}.Install(dir, m))
	require.NoError(t, tools.PiTool{}.Install(dir, m))

	assertFileExists(t, filepath.Join(dir, "AGENTS.md"))

	// Disable OpenCode — Pi still enabled, AGENTS.md must survive.
	m.Tools.OpenCode = false
	require.NoError(t, tools.OpenCodeTool{}.Uninstall(dir, m))

	assertFileAbsent(t, filepath.Join(dir, ".opencode"))
	assertFileExists(t, filepath.Join(dir, "AGENTS.md"), "AGENTS.md must persist while Pi is still enabled")
	assertFileExists(t, filepath.Join(dir, ".pi/prompts/wiki-ingest.md"))
}

func TestSharedAgentsMd_UninstallPi_OpenCodeStillEnabled(t *testing.T) {
	dir := t.TempDir()
	m := baseManifest()
	m.Tools.OpenCode = true
	m.Tools.Pi = true

	require.NoError(t, tools.OpenCodeTool{}.Install(dir, m))
	require.NoError(t, tools.PiTool{}.Install(dir, m))

	// Disable Pi — OpenCode still enabled, AGENTS.md must survive.
	m.Tools.Pi = false
	require.NoError(t, tools.PiTool{}.Uninstall(dir, m))

	assertFileAbsent(t, filepath.Join(dir, ".pi"))
	assertFileExists(t, filepath.Join(dir, "AGENTS.md"), "AGENTS.md must persist while OpenCode is still enabled")
	assertFileExists(t, filepath.Join(dir, ".opencode/commands/wiki-ingest.md"))
}

func TestSharedAgentsMd_UninstallBoth_AgentsMdRemoved(t *testing.T) {
	dir := t.TempDir()
	m := baseManifest()
	m.Tools.OpenCode = true
	m.Tools.Pi = true

	require.NoError(t, tools.OpenCodeTool{}.Install(dir, m))
	require.NoError(t, tools.PiTool{}.Install(dir, m))

	m.Tools.OpenCode = false
	require.NoError(t, tools.OpenCodeTool{}.Uninstall(dir, m))

	m.Tools.Pi = false
	require.NoError(t, tools.PiTool{}.Uninstall(dir, m))

	assertFileAbsent(t, filepath.Join(dir, "AGENTS.md"))
	assertFileAbsent(t, filepath.Join(dir, ".opencode"))
	assertFileAbsent(t, filepath.Join(dir, ".pi"))
}

// ─── Registry ────────────────────────────────────────────────────────────────

func TestRegistry_Get(t *testing.T) {
	for _, name := range []string{"claude-code", "opencode", "pi"} {
		t.Run(name, func(t *testing.T) {
			tool, err := tools.Get(name)
			require.NoError(t, err)
			assert.Equal(t, name, tool.Name())
		})
	}
}

func TestRegistry_Get_Unknown(t *testing.T) {
	_, err := tools.Get("vim")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func assertFileExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, append([]interface{}{path + " should exist"}, msgAndArgs...)...)
}

func assertFileAbsent(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), append([]interface{}{path + " should not exist"}, msgAndArgs...)...)
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), substr)
}
