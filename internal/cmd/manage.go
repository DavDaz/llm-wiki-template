package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-generator/internal/tui/dashboard"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Open the TUI dashboard to manage the current wiki",
	RunE:  runManage,
}

func init() {
	rootCmd.AddCommand(manageCmd)
}

func runManage(_ *cobra.Command, _ []string) error {
	m, wikiRoot, err := loadManifestFromCwd()
	if err != nil {
		return err
	}
	d := dashboard.New(m, wikiRoot)
	p := tea.NewProgram(d, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
