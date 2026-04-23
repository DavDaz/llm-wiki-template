package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-generator/internal/templates"
	"github.com/DavDaz/llm-wiki-generator/internal/tui/viewer"
)

var guideCmd = &cobra.Command{
	Use:   "guide",
	Short: "Browse the conceptual guide with keyboard navigation",
	RunE:  runGuide,
}

func init() {
	rootCmd.AddCommand(guideCmd)
}

func runGuide(_ *cobra.Command, _ []string) error {
	raw, err := templates.ReadFile("GUIDE.md")
	if err != nil {
		return fmt.Errorf("read guide: %w", err)
	}

	m, err := viewer.New(string(raw))
	if err != nil {
		// fallback: print raw if TUI fails (e.g. no TTY)
		fmt.Fprintln(os.Stdout, string(raw))
		return nil
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
