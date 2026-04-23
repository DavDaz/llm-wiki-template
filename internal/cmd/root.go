// Package cmd defines all CLI subcommands for llm-wiki.
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-generator/internal/tui/dashboard"
	"github.com/DavDaz/llm-wiki-generator/internal/tui/wizard"
	"github.com/DavDaz/llm-wiki-generator/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "llm-wiki",
	Short: "Manage LLM wikis — create, migrate tools, edit domain config",
	Long: `llm-wiki is a CLI + TUI for creating and managing LLM wikis.

It replaces the legacy setup.sh with a versioned, distributable binary
that supports Claude Code, OpenCode, and Pi as AI tool backends.`,
	SilenceUsage: true,
	RunE:         runRoot,
}

// Execute is the entry point called by main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the llm-wiki version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version.Version)
	},
}

// runRoot opens the dashboard TUI if inside a wiki, otherwise the init wizard.
func runRoot(_ *cobra.Command, _ []string) error {
	m, wikiRoot, err := loadManifestFromCwd()
	if err == nil {
		d := dashboard.New(m, wikiRoot)
		p := tea.NewProgram(d, tea.WithAltScreen())
		_, runErr := p.Run()
		return runErr
	}

	parentDir, _ := os.Getwd()
	wiz := wizard.New(parentDir)
	p := tea.NewProgram(wiz, tea.WithAltScreen())
	final, runErr := p.Run()
	if runErr != nil {
		return runErr
	}
	if wm, ok := final.(wizard.Model); ok {
		r := wm.GetResult()
		if !r.Aborted {
			fmt.Printf("✓ Wiki created: %s\n", r.WikiRoot)
		}
	}
	return nil
}
