package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadManifestFromCwdOutsideWikiReturnsClearError(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWD))
	})

	require.NoError(t, os.Chdir(t.TempDir()))

	_, _, err = loadManifestFromCwd()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no wiki found in current directory")
	require.Contains(t, err.Error(), "wiki.toml not found")
}
