package dashboard

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
)

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "q", "j", "k", "m":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func runCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

func updateAndDrain(m tea.Model, msg tea.Msg) tea.Model {
	next, cmd := m.Update(msg)
	for cmd != nil {
		followUp := cmd()
		if followUp == nil {
			break
		}
		next, cmd = next.Update(followUp)
	}
	return next
}

func writeWikiFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	m := manifest.New("Test Wiki", "test-wiki", "es")
	m.Tools.ClaudeCode = true
	m.Tools.OpenCode = false
	m.Tools.Pi = false
	require.NoError(t, m.Save(root))

	wikiDir := filepath.Join(root, "wiki")
	require.NoError(t, os.MkdirAll(wikiDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, "draft.md"), []byte("---\ntitulo: Draft\ntipo: referencia\nstatus: borrador\nactualizado: 2026-01-01\n---\n# draft\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, "vigente.md"), []byte("---\ntitulo: Live\ntipo: referencia\nstatus: vigente\nactualizado: 2026-01-01\n---\n# live\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, "bad.md"), []byte("# invalid"), 0o644))

	return root
}
