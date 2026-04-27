package dashboard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/require"
)

func TestRootMenuRendersHeader(t *testing.T) {
	root := NewRoot(writeWikiFixture(t)).(*rootModel)
	view := root.View()
	require.Contains(t, view, "llm-wiki — manage")
}

func TestRootMenuOptionsExactOrderAndLabels(t *testing.T) {
	require.Equal(t, []rootMenuOption{
		{label: "Tools backends", value: "tools"},
		{label: "Drafts (status: borrador)", value: "drafts"},
		{label: "Exit", value: "exit"},
	}, rootMenuOptions)
}

func TestRootEscAndQQuit(t *testing.T) {
	root := NewRoot(writeWikiFixture(t)).(*rootModel)

	_, cmd := root.Update(keyMsg("esc"))
	require.NotNil(t, cmd)
	require.IsType(t, tea.QuitMsg{}, runCmd(cmd))

	root = NewRoot(writeWikiFixture(t)).(*rootModel)
	_, cmd = root.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	require.IsType(t, tea.QuitMsg{}, runCmd(cmd))
}

func TestRootNavigationRoundTrip(t *testing.T) {
	root := NewRoot(writeWikiFixture(t)).(*rootModel)
	_ = root.Init()

	root = updateAndDrain(root, keyMsg("enter")).(*rootModel)
	require.Equal(t, stateSubView, root.state)

	_, cmd := root.Update(keyMsg("esc"))
	require.NotNil(t, cmd)
	msg := runCmd(cmd)
	_, ok := msg.(BackToRootMsg)
	require.True(t, ok)

	root = updateAndDrain(root, msg).(*rootModel)
	require.Equal(t, stateMenu, root.state)

	root = updateAndDrain(root, keyMsg("down")).(*rootModel)
	root = updateAndDrain(root, keyMsg("enter")).(*rootModel)
	require.Equal(t, stateSubView, root.state)

	_, cmd = root.Update(keyMsg("esc"))
	require.NotNil(t, cmd)
	msg = runCmd(cmd)
	_, ok = msg.(BackToRootMsg)
	require.True(t, ok)
	root = updateAndDrain(root, msg).(*rootModel)
	require.Equal(t, stateMenu, root.state)

	root.vals.choice = "exit"
	root.form = newRootForm(root.vals)
	root.form.State = huh.StateCompleted
	_, cmd = root.Update(nil)
	require.NotNil(t, cmd)
	require.IsType(t, tea.QuitMsg{}, runCmd(cmd))
}

func TestRootFormRebuiltOnBack(t *testing.T) {
	root := NewRoot(writeWikiFixture(t)).(*rootModel)
	originalForm := root.form

	root.state = stateSubView
	root.active = NewPagesView(root.wikiDir, "borrador")

	root = updateAndDrain(root, BackToRootMsg{}).(*rootModel)
	require.Equal(t, stateMenu, root.state)
	require.NotNil(t, root.form)
	require.NotSame(t, originalForm, root.form)
}

func TestRootCtrlCAlwaysQuits(t *testing.T) {
	root := NewRoot(writeWikiFixture(t)).(*rootModel)
	root.state = stateSubView
	root.active = NewPagesView(root.wikiDir, "borrador")

	_, cmd := root.Update(keyMsg("ctrl+c"))
	require.NotNil(t, cmd)
	require.IsType(t, tea.QuitMsg{}, runCmd(cmd))
}
