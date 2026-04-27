package dashboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-generator/internal/pages"
)

func optionValues(opts []huh.Option[string]) []string {
	values := make([]string, 0, len(opts))
	for _, opt := range opts {
		values = append(values, opt.Value)
	}
	return values
}

func TestStatusTargets(t *testing.T) {
	tests := []struct {
		name     string
		current  pages.Status
		expected []string
	}{
		{name: "from borrador", current: pages.StatusBorrador, expected: []string{"vigente", "deprecado", "cancel"}},
		{name: "from vigente", current: pages.StatusVigente, expected: []string{"borrador", "deprecado", "cancel"}},
		{name: "from deprecado", current: pages.StatusDeprecado, expected: []string{"borrador", "vigente", "cancel"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, optionValues(statusTargets(tt.current)))
			assert.Equal(t, tt.expected[0], defaultStatusTarget(tt.current))
		})
	}
}

func TestPagesViewFilteredAndSkippedFooter(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")

	m := NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel)
	require.Len(t, m.list, 1)
	assert.Equal(t, "Draft", m.list[0].Title)

	view := m.View()
	assert.Contains(t, view, "Draft")
	assert.Contains(t, view, "1 skipped (malformed)")
}

func TestPagesViewEmptyState(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")

	m := NewPagesView(wikiDir, pages.StatusDeprecado).(*pagesModel)
	require.Empty(t, m.list)

	view := m.View()
	assert.Contains(t, view, "No pages with status deprecado found")
	assert.Contains(t, view, "[esc] back")

	_, cmd := m.Update(keyMsg("enter"))
	assert.Nil(t, cmd)
	assert.False(t, m.picking)
}

func TestPagesViewTitleReflectsFilter(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")

	assert.Contains(t, NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel).View(), "Pages (status: borrador)")
	assert.Contains(t, NewPagesView(wikiDir, pages.StatusVigente).(*pagesModel).View(), "Pages (status: vigente)")
	assert.Contains(t, NewPagesView(wikiDir, pages.StatusDeprecado).(*pagesModel).View(), "Pages (status: deprecado)")
}

func TestPagesCursorNavigationBounds(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")

	writePage(t, filepath.Join(wikiDir, "draft2.md"), "Draft 2", "borrador")
	m := NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel)
	require.Len(t, m.list, 2)

	m.cursor = 0
	_, _ = m.Update(keyMsg("up"))
	assert.Equal(t, 0, m.cursor)

	_, _ = m.Update(keyMsg("down"))
	assert.Equal(t, 1, m.cursor)

	_, _ = m.Update(keyMsg("down"))
	assert.Equal(t, 1, m.cursor)
}

func TestPagesEnterOpensPicker(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")
	m := NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel)

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	assert.True(t, m.picking)
	assert.NotNil(t, m.pickForm)
}

func TestPagesPickVigenteRemovesRow(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")
	m := NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel)
	require.Len(t, m.list, 1)

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	require.True(t, m.picking)

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel) // default = vigente
	assert.False(t, m.picking)
	assert.Len(t, m.list, 0)

	b, err := os.ReadFile(filepath.Join(wikiDir, "draft.md"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "status: vigente")
}

func TestPagesPickDeprecadoRemovesRow(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")
	m := NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel)

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	require.True(t, m.picking)
	m = updateAndDrain(m, keyMsg("down")).(*pagesModel)
	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)

	assert.False(t, m.picking)
	assert.Len(t, m.list, 0)
	b, err := os.ReadFile(filepath.Join(wikiDir, "draft.md"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "status: deprecado")
}

func TestPagesPickFromVigenteOffersValidTargets(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")
	m := NewPagesView(wikiDir, pages.StatusVigente).(*pagesModel)
	require.Len(t, m.list, 1)

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	require.True(t, m.picking)

	picker := m.View()
	assert.Contains(t, picker, "borrador")
	assert.Contains(t, picker, "deprecado")
	assert.Contains(t, picker, "cancel")
	assert.NotContains(t, picker, "\n  vigente\n")

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	assert.False(t, m.picking)
	assert.Len(t, m.list, 0)

	b, err := os.ReadFile(filepath.Join(wikiDir, "vigente.md"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "status: borrador")
}

func TestPagesPickFromDeprecadoOffersValidTargets(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")
	writeFixturePage(t, wikiDir, "deprecado.md", "Legacy", "deprecado")

	m := NewPagesView(wikiDir, pages.StatusDeprecado).(*pagesModel)
	require.Len(t, m.list, 1)

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	require.True(t, m.picking)

	picker := m.View()
	assert.Contains(t, picker, "borrador")
	assert.Contains(t, picker, "vigente")
	assert.Contains(t, picker, "cancel")
	assert.NotContains(t, picker, "\n  deprecado\n")

	m = updateAndDrain(m, keyMsg("down")).(*pagesModel)
	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	assert.False(t, m.picking)
	assert.Len(t, m.list, 0)

	b, err := os.ReadFile(filepath.Join(wikiDir, "deprecado.md"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "status: vigente")
}

func TestPagesPickCancelNoChange(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")
	m := NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel)

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	m = updateAndDrain(m, keyMsg("down")).(*pagesModel)
	m = updateAndDrain(m, keyMsg("down")).(*pagesModel)
	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)

	assert.False(t, m.picking)
	assert.Len(t, m.list, 1)
	b, err := os.ReadFile(filepath.Join(wikiDir, "draft.md"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "status: borrador")
}

func TestPagesPickErrorKeepsRow(t *testing.T) {
	root := writeWikiFixture(t)
	wikiDir := filepath.Join(root, "wiki")
	m := NewPagesView(wikiDir, pages.StatusBorrador).(*pagesModel)

	require.NoError(t, os.Chmod(wikiDir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(wikiDir, 0o755) })

	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)
	m = updateAndDrain(m, keyMsg("enter")).(*pagesModel)

	assert.False(t, m.picking)
	assert.Len(t, m.list, 1)
	assert.NotEmpty(t, m.errMsg)
}

func TestPagesEscEmitsBack(t *testing.T) {
	m := NewPagesView(filepath.Join(writeWikiFixture(t), "wiki"), pages.StatusBorrador).(*pagesModel)
	_, cmd := m.Update(keyMsg("esc"))
	require.NotNil(t, cmd)
	_, ok := runCmd(cmd).(BackToRootMsg)
	require.True(t, ok)
}

func TestPagesCtrlCDoesNotQuitFromSubview(t *testing.T) {
	m := NewPagesView(filepath.Join(writeWikiFixture(t), "wiki"), pages.StatusBorrador).(*pagesModel)
	_, cmd := m.Update(keyMsg("ctrl+c"))
	require.Nil(t, cmd)
}

func writePage(t *testing.T, path, title, status string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte("---\ntitulo: "+title+"\nstatus: "+status+"\nactualizado: 2026-01-01\n---\n"), 0o644))
}
