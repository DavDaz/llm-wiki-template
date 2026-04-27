package cmd

import (
	"fmt"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
)

func TestRunRootInsideWikiKeepsNoArgToolsDashboardRouting(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWD))
	})

	wikiRoot := t.TempDir()
	m := manifest.New("Test Wiki", "test-wiki", "es")
	require.NoError(t, m.Save(wikiRoot))
	require.NoError(t, os.Chdir(wikiRoot))

	originalRunProgram := runProgram
	t.Cleanup(func() {
		runProgram = originalRunProgram
	})

	called := 0
	modelType := ""
	runProgram = func(model tea.Model, _ ...tea.ProgramOption) (tea.Model, error) {
		called++
		modelType = fmt.Sprintf("%T", model)
		return model, nil
	}

	err = runRoot(nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, called)
	require.Equal(t, "*dashboard.Model", modelType)
}
