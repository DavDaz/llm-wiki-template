package dashboard

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
	"github.com/DavDaz/llm-wiki-generator/internal/pages"
	"github.com/DavDaz/llm-wiki-generator/internal/tui/styles"
)

type rootState int

const (
	stateMenu rootState = iota
	stateSubView
)

type rootValues struct {
	choice string
}

type rootMenuOption struct {
	label string
	value string
}

var rootMenuOptions = []rootMenuOption{
	{label: "Tools backends", value: "tools"},
	{label: "Drafts (status: borrador)", value: "drafts"},
	{label: "Exit", value: "exit"},
}

type rootModel struct {
	wikiRoot string
	wikiDir  string
	state    rootState
	form     *huh.Form
	vals     *rootValues
	active   tea.Model
	errMsg   string
}

// NewRoot creates the root manage menu model.
func NewRoot(wikiRoot string) tea.Model {
	v := &rootValues{choice: "tools"}
	return &rootModel{
		wikiRoot: wikiRoot,
		wikiDir:  filepath.Join(wikiRoot, "wiki"),
		state:    stateMenu,
		form:     newRootForm(v),
		vals:     v,
	}
}

func newRootForm(v *rootValues) *huh.Form {
	options := make([]huh.Option[string], 0, len(rootMenuOptions))
	for _, opt := range rootMenuOptions {
		options = append(options, huh.NewOption(opt.label, opt.value))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("llm-wiki — manage").
				Options(options...).
				Value(&v.choice),
		),
	).WithTheme(huh.ThemeCatppuccin())
}

func (m *rootModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if _, ok := msg.(BackToRootMsg); ok {
		m.form = newRootForm(m.vals)
		m.state = stateMenu
		m.active = nil
		return m, m.form.Init()
	}

	switch m.state {
	case stateMenu:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "esc", "q":
				return m, tea.Quit
			}
		}

		f, cmd := m.form.Update(msg)
		if updated, ok := f.(*huh.Form); ok {
			m.form = updated
		}

		if m.form.State == huh.StateCompleted {
			switch m.vals.choice {
			case "tools":
				mf, err := manifest.Load(m.wikiRoot)
				if err != nil {
					m.errMsg = err.Error()
					m.form = newRootForm(m.vals)
					return m, m.form.Init()
				}
				m.active = NewTools(mf, m.wikiRoot)
				m.state = stateSubView
				return m, m.active.Init()
			case "drafts":
				m.active = NewPagesView(m.wikiDir, pages.StatusBorrador)
				m.state = stateSubView
				return m, m.active.Init()
			case "exit":
				return m, tea.Quit
			}
		}

		if m.form.State == huh.StateAborted {
			return m, tea.Quit
		}

		return m, cmd

	case stateSubView:
		sub, cmd := m.active.Update(msg)
		m.active = sub
		return m, cmd
	}

	return m, nil
}

func (m *rootModel) View() string {
	if m.state == stateSubView && m.active != nil {
		return m.active.View()
	}

	header := styles.Title.Render("llm-wiki — manage") + "\n"
	header += styles.Muted.Render(fmt.Sprintf("wiki: %s", m.wikiRoot)) + "\n\n"
	if m.errMsg != "" {
		header += styles.Warning.Render("  ✗ "+m.errMsg) + "\n\n"
	}
	return header + m.form.View()
}
