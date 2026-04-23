// Package wizard provides a TUI form for creating a new wiki interactively.
package wizard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/DavDaz/llm-wiki-template/internal/generator"
	"github.com/DavDaz/llm-wiki-template/internal/tui/styles"
)

// Result holds the outcome of a completed wizard run.
type Result struct {
	WikiRoot string
	Aborted  bool
}

// formValues holds all mutable form field values behind a pointer so they
// survive Bubbletea's copy semantics when the Model is passed by value.
type formValues struct {
	name        string
	slug        string
	language    string
	tools       []string
	entities    string
	pageTypes   string
	conventions string
	confirmed   bool
}

// Model is the Bubbletea model for the init wizard.
type Model struct {
	form      *huh.Form
	values    *formValues // pointer — valid across Bubbletea copies
	result    Result
	done      bool
	errMsg    string
	parentDir string
}

// New creates a new wizard Model. parentDir is the directory where the wiki
// will be created (empty string = current directory).
func New(parentDir string) Model {
	v := &formValues{
		language:  "es",
		pageTypes: "proceso, referencia, entidad, politica",
	}
	return Model{
		parentDir: parentDir,
		values:    v,
		form:      buildForm(v),
	}
}

func buildForm(v *formValues) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Wiki name").
				Description("Human-readable name for your wiki").
				Placeholder("Legal Wiki").
				Value(&v.name),

			huh.NewInput().
				Title("Slug").
				Description("Kebab-case identifier (e.g. legal-wiki)").
				Placeholder("legal-wiki").
				Value(&v.slug),

			huh.NewSelect[string]().
				Title("Language").
				Options(
					huh.NewOption("Spanish (es)", "es"),
					huh.NewOption("English (en)", "en"),
				).
				Value(&v.language),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("AI tool backends").
				Description("Select at least one").
				Options(
					huh.NewOption("Claude Code", "claude-code"),
					huh.NewOption("OpenCode", "opencode"),
					huh.NewOption("Pi", "pi"),
				).
				Value(&v.tools),

			huh.NewInput().
				Title("Primary entities").
				Description("Comma-separated domain entities (optional)").
				Placeholder("usuario, rol, permiso").
				Value(&v.entities),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Page types").
				Description("Comma-separated content types for this wiki").
				Value(&v.pageTypes),

			huh.NewInput().
				Title("Domain conventions").
				Description("Comma-separated rules the AI must enforce (optional)").
				Placeholder("Citar fuente en fuentes:, No abreviar nombres").
				Value(&v.conventions),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create wiki?").
				Description("This will generate the wiki directory structure.").
				Affirmative("Create").
				Negative("Cancel").
				Value(&v.confirmed),
		),
	).WithTheme(huh.ThemeCatppuccin())
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+c" {
		m.result = Result{Aborted: true}
		m.done = true
		return m, tea.Quit
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		if !m.values.confirmed {
			m.result = Result{Aborted: true}
			m.done = true
			return m, tea.Quit
		}
		wikiRoot, err := create(m.values, m.parentDir)
		if err != nil {
			m.errMsg = err.Error()
			m.values.confirmed = false
			m.form = buildForm(m.values)
			return m, m.form.Init()
		}
		m.result = Result{WikiRoot: wikiRoot}
		m.done = true
		return m, tea.Quit
	}
	if m.form.State == huh.StateAborted {
		m.result = Result{Aborted: true}
		m.done = true
		return m, tea.Quit
	}

	return m, cmd
}

func (m Model) View() string {
	if m.done {
		if m.result.Aborted {
			return styles.Muted.Render("Cancelled.") + "\n"
		}
		return styles.Success.Render("✓ Wiki created: "+m.result.WikiRoot) + "\n"
	}

	header := styles.Title.Render("llm-wiki — New Wiki")
	hint := styles.KeyHint.Render("ctrl+c to cancel")

	parts := []string{header, m.form.View()}
	if m.errMsg != "" {
		parts = append(parts, styles.Warning.Render("✗ "+m.errMsg))
	}
	parts = append(parts, hint)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// GetResult returns the wizard result after the program has exited.
func (m Model) GetResult() Result {
	return m.result
}

func create(v *formValues, parentDir string) (string, error) {
	claude, opencode, pi := resolveTools(v.tools)
	pageTypes := parseCSV(v.pageTypes)
	if len(pageTypes) == 0 {
		pageTypes = []string{"proceso", "referencia", "entidad", "politica"}
	}
	cfg := generator.InitConfig{
		ParentDir:       parentDir,
		Name:            v.name,
		Slug:            v.slug,
		Language:        v.language,
		ClaudeCode:      claude,
		OpenCode:        opencode,
		Pi:              pi,
		PrimaryEntities: parseCSV(v.entities),
		PageTypes:       pageTypes,
		Conventions:     parseCSV(v.conventions),
	}
	return generator.Init(cfg)
}

func resolveTools(names []string) (claude, opencode, pi bool) {
	for _, n := range names {
		switch strings.TrimSpace(strings.ToLower(n)) {
		case "claude-code":
			claude = true
		case "opencode":
			opencode = true
		case "pi":
			pi = true
		}
	}
	return
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
