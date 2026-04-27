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

	writeFixturePage(t, wikiDir, "draft.md", "Draft", "borrador")
	writeFixturePage(t, wikiDir, "vigente.md", "Live", "vigente")
	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, "bad.md"), []byte("# invalid"), 0o644))

	return root
}

func writeFixturePage(t *testing.T, wikiDir, fileName, title, status string) {
	t.Helper()
	buf := []byte("---\ntitulo: " + title + "\ntipo: referencia\nstatus: " + status + "\nactualizado: 2026-01-01\n---\n# " + title + "\n")
	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, fileName), buf, 0o644))
}
