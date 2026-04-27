package dashboard

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/DavDaz/llm-wiki-generator/internal/pages"
	"github.com/DavDaz/llm-wiki-generator/internal/tui/styles"
)

type pickValues struct {
	choice string
}

type pagesModel struct {
	wikiDir  string
	filter   pages.Status
	list     []pages.Page
	skipped  int
	cursor   int
	picking  bool
	pickForm *huh.Form
	pickVals *pickValues
	errMsg   string
}

// NewPagesView creates a pages list view filtered by status.
func NewPagesView(wikiDir string, filter pages.Status) tea.Model {
	all, skipped, err := pages.List(wikiDir)
	m := &pagesModel{
		wikiDir: wikiDir,
		filter:  filter,
		skipped: skipped,
		list:    make([]pages.Page, 0),
	}
	if err != nil {
		m.errMsg = err.Error()
		return m
	}
	m.list = pages.FilterByStatus(all, filter)
	return m
}

func newPickForm(v *pickValues) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Change status to:").
				Options(
					huh.NewOption("vigente", "vigente"),
					huh.NewOption("deprecado", "deprecado"),
					huh.NewOption("cancel", "cancel"),
				).
				Value(&v.choice),
		),
	).WithTheme(huh.ThemeCatppuccin())
}

func (m *pagesModel) Init() tea.Cmd {
	return nil
}

func (m *pagesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.picking {
		f, cmd := m.pickForm.Update(msg)
		if updated, ok := f.(*huh.Form); ok {
			m.pickForm = updated
		}

		if m.pickForm.State == huh.StateCompleted {
			choice := m.pickVals.choice
			m.picking = false
			m.pickForm = nil

			if choice == "cancel" {
				return m, nil
			}

			if m.cursor < 0 || m.cursor >= len(m.list) {
				return m, nil
			}

			target := m.list[m.cursor]
			if err := pages.SetStatus(target.Path, pages.Status(choice)); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}

			m.list = append(m.list[:m.cursor], m.list[m.cursor+1:]...)
			if m.cursor >= len(m.list) && m.cursor > 0 {
				m.cursor--
			}

			return m, nil
		}

		if m.pickForm.State == huh.StateAborted {
			m.picking = false
			m.pickForm = nil
			return m, nil
		}

		return m, cmd
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "q":
			return m, BackToRoot()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.errMsg = ""
		case "down", "j":
			if m.cursor < len(m.list)-1 {
				m.cursor++
			}
			m.errMsg = ""
		case "enter":
			if len(m.list) == 0 {
				return m, nil
			}
			m.pickVals = &pickValues{choice: "vigente"}
			m.pickForm = newPickForm(m.pickVals)
			m.picking = true
			m.errMsg = ""
			return m, m.pickForm.Init()
		}
	}

	return m, nil
}

func (m *pagesModel) View() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("Drafts (status: " + string(m.filter) + ")"))
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(styles.Warning.Render("  ✗ " + m.errMsg))
		b.WriteString("\n\n")
	}

	if m.picking && m.pickForm != nil {
		if len(m.list) > 0 && m.cursor >= 0 && m.cursor < len(m.list) {
			b.WriteString(styles.Bold.Render("  " + m.list[m.cursor].Title))
			b.WriteString("\n")
		}
		b.WriteString(m.pickForm.View())
		return b.String()
	}

	if len(m.list) == 0 {
		b.WriteString("  No drafts found.\n")
		b.WriteString(styles.KeyHint.Render("  [esc] back"))
		b.WriteString("\n")
	} else {
		for i, p := range m.list {
			cursor := "  "
			if i == m.cursor {
				cursor = styles.Primary.Render("> ")
			}
			relPath, err := filepath.Rel(m.wikiDir, p.Path)
			if err != nil {
				relPath = p.Path
			}
			b.WriteString(fmt.Sprintf("%s%s | %s | %s | %s\n", cursor, p.Title, p.Type, formatUpdatedAt(p.UpdatedAt), relPath))
		}
	}

	if m.skipped > 0 {
		b.WriteString(styles.Muted.Render(fmt.Sprintf("\n%d skipped (malformed)", m.skipped)))
		b.WriteString("\n")
	}

	b.WriteString(styles.KeyHint.Render("\n  [↑/↓] nav  [enter] change status  [esc] back  [ctrl+c] quit"))
	b.WriteString("\n")
	return b.String()
}

func formatUpdatedAt(ts time.Time) string {
	if ts.IsZero() {
		return "-"
	}
	return ts.Format("2006-01-02")
}
