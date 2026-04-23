// Package dashboard provides a TUI for managing an existing wiki.
package dashboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DavDaz/llm-wiki-generator/internal/generator"
	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
	"github.com/DavDaz/llm-wiki-generator/internal/tools"
	"github.com/DavDaz/llm-wiki-generator/internal/tui/styles"
)

type toolEntry struct {
	name      string
	installed bool
	enabled   bool
}

// Model is the Bubbletea model for the wiki management dashboard.
type Model struct {
	manifest  *manifest.Manifest
	wikiRoot  string
	toolItems []toolEntry
	cursor    int
	migrated  bool
	errMsg    string
}

// New creates a new dashboard Model for the given wiki.
func New(m *manifest.Manifest, wikiRoot string) Model {
	d := Model{
		manifest: m,
		wikiRoot: wikiRoot,
	}
	d.toolItems = buildToolItems(m, wikiRoot)
	return d
}

func buildToolItems(m *manifest.Manifest, wikiRoot string) []toolEntry {
	all := tools.All()
	items := make([]toolEntry, len(all))
	for i, t := range all {
		items[i] = toolEntry{
			name:      t.Name(),
			installed: t.IsInstalled(wikiRoot),
			enabled:   isEnabled(m, t.Name()),
		}
	}
	return items
}

func isEnabled(m *manifest.Manifest, name string) bool {
	switch name {
	case "claude-code":
		return m.Tools.ClaudeCode
	case "opencode":
		return m.Tools.OpenCode
	case "pi":
		return m.Tools.Pi
	}
	return false
}

func (d Model) Init() tea.Cmd {
	return nil
}

func (d Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return d, tea.Quit

		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}

		case "down", "j":
			if d.cursor < len(d.toolItems)-1 {
				d.cursor++
			}

		case " ":
			d.toggleTool()

		case "m":
			d.errMsg = ""
			d.migrated = false
			if err := d.migrate(); err != nil {
				d.errMsg = err.Error()
			} else {
				d.migrated = true
			}
		}
	}
	return d, nil
}

func (d *Model) toggleTool() {
	item := &d.toolItems[d.cursor]
	newEnabled := !item.enabled

	// prevent disabling all tools
	if !newEnabled && d.countEnabled() <= 1 {
		d.errMsg = "at least one tool must remain enabled"
		return
	}
	d.errMsg = ""

	item.enabled = newEnabled
	d.applyToolsToManifest()
	if err := d.manifest.Save(d.wikiRoot); err != nil {
		d.errMsg = err.Error()
	}
}

func (d *Model) countEnabled() int {
	n := 0
	for _, t := range d.toolItems {
		if t.enabled {
			n++
		}
	}
	return n
}

func (d *Model) applyToolsToManifest() {
	for _, t := range d.toolItems {
		switch t.name {
		case "claude-code":
			d.manifest.Tools.ClaudeCode = t.enabled
		case "opencode":
			d.manifest.Tools.OpenCode = t.enabled
		case "pi":
			d.manifest.Tools.Pi = t.enabled
		}
	}
}

func (d *Model) migrate() error {
	d.applyToolsToManifest()
	if err := d.manifest.Save(d.wikiRoot); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}
	if err := generator.Migrate(d.wikiRoot, d.manifest); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	d.toolItems = buildToolItems(d.manifest, d.wikiRoot)
	return nil
}

func (d Model) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("llm-wiki — Dashboard"))
	b.WriteString("\n\n")

	// wiki info
	info := fmt.Sprintf(
		"  Name:     %s\n  Slug:     %s\n  Language: %s\n  Version:  %s\n  Created:  %s",
		d.manifest.Wiki.Name,
		d.manifest.Wiki.Slug,
		d.manifest.Wiki.Language,
		d.manifest.Wiki.TemplateVersion,
		d.manifest.Wiki.CreatedAt,
	)
	b.WriteString(styles.Box.Render(info))
	b.WriteString("\n\n")

	// tools section
	b.WriteString(styles.Bold.Render("  Tools"))
	b.WriteString("\n")
	for i, t := range d.toolItems {
		cursor := "  "
		if i == d.cursor {
			cursor = styles.Primary.Render("> ")
		}

		icon := "○"
		iconStyle := styles.Muted
		if t.enabled && t.installed {
			icon = "●"
			iconStyle = styles.Success
		} else if t.enabled && !t.installed {
			icon = "◎"
			iconStyle = styles.Warning
		}

		name := t.name
		if i == d.cursor {
			name = styles.Bold.Render(name)
		}

		b.WriteString(cursor + iconStyle.Render(icon) + " " + name + "\n")
	}

	// legend
	b.WriteString("\n")
	b.WriteString(styles.Muted.Render(
		"  ● installed & enabled  ◎ enabled (run migrate)  ○ disabled",
	))
	b.WriteString("\n")

	// status message
	if d.errMsg != "" {
		b.WriteString("\n" + styles.Warning.Render("  ✗ "+d.errMsg) + "\n")
	}
	if d.migrated {
		b.WriteString("\n" + styles.Success.Render("  ✓ Migration complete") + "\n")
	}

	// keybinds
	b.WriteString(styles.KeyHint.Render(
		"\n  [↑/↓] navigate  [space] toggle  [m] migrate  [q] quit",
	))
	b.WriteString("\n")

	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}
