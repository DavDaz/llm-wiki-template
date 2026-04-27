package dashboard

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
)

func TestToolsEscEmitsBack(t *testing.T) {
	m := manifest.New("Wiki", "wiki", "es")
	d := New(m, t.TempDir())

	_, cmd := d.Update(keyMsg("esc"))
	require.NotNil(t, cmd)
	_, ok := runCmd(cmd).(BackToRootMsg)
	require.True(t, ok)

	_, cmd = d.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	_, ok = runCmd(cmd).(BackToRootMsg)
	require.True(t, ok)
}

func TestToolsCtrlCDoesNotQuitFromSubview(t *testing.T) {
	m := manifest.New("Wiki", "wiki", "es")
	d := New(m, t.TempDir())

	_, cmd := d.Update(keyMsg("ctrl+c"))
	require.Nil(t, cmd)
}
