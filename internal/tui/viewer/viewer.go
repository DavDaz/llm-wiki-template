// Package viewer provides a scrollable TUI pager for markdown content.
package viewer

import (
	"fmt"

	"charm.land/glamour/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/DavDaz/llm-wiki-generator/internal/tui/styles"
)

// Model is a scrollable markdown viewer.
type Model struct {
	viewport viewport.Model
	ready    bool
	content  string
}

// New creates a viewer Model with the given markdown content.
// The content is rendered with glamour before display.
func New(markdown string) (Model, error) {
	rendered, err := glamour.NewTermRenderer(
		glamour.WithEnvironmentConfig(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return Model{}, fmt.Errorf("create renderer: %w", err)
	}
	out, err := rendered.Render(markdown)
	if err != nil {
		return Model{}, fmt.Errorf("render markdown: %w", err)
	}
	return Model{content: out}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := 2
		footerHeight := 2
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return ""
	}

	header := styles.Title.Render("llm-wiki — Guide")
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
		fmt.Sprintf("  %d%%   ↑/↓ scroll · PgUp/PgDn · q quit",
			int(m.viewport.ScrollPercent()*100)),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}
