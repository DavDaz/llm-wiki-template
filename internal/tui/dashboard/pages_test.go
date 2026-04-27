package dashboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-generator/internal/pages"
)

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
	assert.Contains(t, view, "No drafts found")
	assert.Contains(t, view, "[esc] back")

	_, cmd := m.Update(keyMsg("enter"))
	assert.Nil(t, cmd)
	assert.False(t, m.picking)
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
