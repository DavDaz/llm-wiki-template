package pages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty wiki dir", func(t *testing.T) {
		dir := t.TempDir()
		got, skipped, err := List(dir)
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.Equal(t, 0, skipped)
	})

	t.Run("skips malformed and no frontmatter", func(t *testing.T) {
		dir := t.TempDir()

		writeFile(t, filepath.Join(dir, "a.md"), "---\ntitulo: A\ntipo: referencia\nstatus: borrador\nactualizado: 2026-01-01\n---\n# A\n")
		writeFile(t, filepath.Join(dir, "b.md"), "---\ntitulo: B\nstatus: vigente\n---\n# B\n")
		writeFile(t, filepath.Join(dir, "bad1.md"), "# no frontmatter\n")
		writeFile(t, filepath.Join(dir, "bad2.md"), "---\nnot yaml line\nstatus: borrador\n---\n")

		got, skipped, err := List(dir)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, 2, skipped)
		assert.Equal(t, "A", got[0].Title)
		assert.Equal(t, Status("vigente"), got[1].Status)
		assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), got[0].UpdatedAt)
	})

	t.Run("non recursive scan", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		require.NoError(t, os.MkdirAll(sub, 0o755))

		writeFile(t, filepath.Join(dir, "root.md"), "---\ntitulo: Root\nstatus: borrador\n---\n")
		writeFile(t, filepath.Join(sub, "nested.md"), "---\ntitulo: Nested\nstatus: borrador\n---\n")

		got, skipped, err := List(dir)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "Root", got[0].Title)
		assert.Equal(t, 0, skipped)
	})

	t.Run("ignores reserved wiki files", func(t *testing.T) {
		dir := t.TempDir()

		writeFile(t, filepath.Join(dir, "index.md"), "# index\n")
		writeFile(t, filepath.Join(dir, "log.md"), "# log\n")
		writeFile(t, filepath.Join(dir, "draft.md"), "---\ntitulo: Draft\nstatus: borrador\n---\n")
		writeFile(t, filepath.Join(dir, "bad.md"), "# malformed\n")

		got, skipped, err := List(dir)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "Draft", got[0].Title)
		assert.Equal(t, 1, skipped)
	})

	t.Run("missing dir returns error", func(t *testing.T) {
		_, _, err := List(filepath.Join(t.TempDir(), "missing"))
		require.Error(t, err)
	})
}

func TestFilterByStatus(t *testing.T) {
	input := []Page{
		{Title: "1", Status: StatusVigente},
		{Title: "2", Status: StatusBorrador},
		{Title: "3", Status: StatusBorrador},
		{Title: "4", Status: StatusDeprecado},
	}

	filtered := FilterByStatus(input, StatusBorrador)
	require.Len(t, filtered, 2)
	assert.Equal(t, "2", filtered[0].Title)
	assert.Equal(t, "3", filtered[1].Title)

	empty := FilterByStatus(nil, StatusBorrador)
	assert.NotNil(t, empty)
	assert.Empty(t, empty)
}

func TestSetStatus(t *testing.T) {
	t.Run("happy path LF", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "page.md")
		original := "---\ntitulo: A\nstatus: borrador\n---\n# body\n"
		writeFile(t, path, original)

		require.NoError(t, SetStatus(path, StatusVigente))

		got := readFile(t, path)
		assert.Contains(t, got, "status: vigente\n")
		assert.Contains(t, got, "# body\n")
		assert.NoFileExists(t, path+".tmp")
	})

	t.Run("preserves CRLF", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "page.md")
		original := "---\r\ntitulo: A\r\nstatus: borrador\r\n---\r\nbody\r\n"
		writeFile(t, path, original)

		require.NoError(t, SetStatus(path, StatusVigente))
		got := readFile(t, path)
		assert.Contains(t, got, "status: vigente\r\n")
		assert.NotContains(t, strings.ReplaceAll(got, "\r\n", ""), "\n")
	})

	t.Run("preserves BOM", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "page.md")
		original := string([]byte{0xEF, 0xBB, 0xBF}) + "---\ntitulo: A\nstatus: borrador\n---\n"
		writeFile(t, path, original)

		require.NoError(t, SetStatus(path, StatusVigente))
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(data), 3)
		assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, data[:3])
	})

	t.Run("no frontmatter", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "page.md")
		writeFile(t, path, "# no frontmatter\n")

		err := SetStatus(path, StatusVigente)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "frontmatter")
	})

	t.Run("frontmatter without status", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "page.md")
		writeFile(t, path, "---\ntitulo: A\n---\n")

		err := SetStatus(path, StatusVigente)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status")
	})

	t.Run("unknown status", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "page.md")
		writeFile(t, path, "---\nstatus: borrador\n---\n")

		err := SetStatus(path, Status("publicado"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown status")
	})

	t.Run("preserves no trailing newline", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "page.md")
		writeFile(t, path, "---\nstatus: borrador\n---\nbody")

		require.NoError(t, SetStatus(path, StatusDeprecado))
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.NotEqual(t, byte('\n'), data[len(data)-1])
		assert.Contains(t, string(data), "status: deprecado")
	})

	t.Run("rename failure keeps original", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "page.md")
		original := "---\nstatus: borrador\n---\n"
		writeFile(t, path, original)

		require.NoError(t, os.Chmod(dir, 0o555))
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

		err := SetStatus(path, StatusVigente)
		require.Error(t, err)
		assert.Equal(t, original, readFile(t, path))
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}
