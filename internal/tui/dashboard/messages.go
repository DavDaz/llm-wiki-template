package dashboard

import tea "github.com/charmbracelet/bubbletea"

// BackToRootMsg is emitted by sub-views to return to the root menu.
type BackToRootMsg struct{}

// BackToRoot emits BackToRootMsg.
func BackToRoot() tea.Cmd {
	return func() tea.Msg { return BackToRootMsg{} }
}
