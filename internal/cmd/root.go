// Package cmd defines all CLI subcommands for llm-wiki.
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-generator/internal/tui/dashboard"
	"github.com/DavDaz/llm-wiki-generator/internal/tui/launcher"
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

var runProgram = func(model tea.Model, opts ...tea.ProgramOption) (tea.Model, error) {
	p := tea.NewProgram(model, opts...)
	return p.Run()
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

// runRoot opens the dashboard if inside a wiki, otherwise shows the launcher menu.
func runRoot(_ *cobra.Command, _ []string) error {
	m, wikiRoot, err := loadManifestFromCwd()
	if err == nil {
		d := dashboard.NewTools(m, wikiRoot)
		_, runErr := runProgram(d, tea.WithAltScreen())
		return runErr
	}

	// Outside a wiki — loop so guide can return to the launcher.
	for {
		l := launcher.New()
		final, runErr := runProgram(l, tea.WithAltScreen())
		if runErr != nil {
			return runErr
		}

		lm, ok := final.(launcher.Model)
		if !ok {
			return nil
		}

		switch lm.Result() {
		case launcher.ActionNew:
			return runInitWizard()
		case launcher.ActionGuide:
			if err := runGuide(nil, nil); err != nil {
				return err
			}
			// guide closed → loop back to launcher
		default:
			return nil
		}
	}
}

// runInitWizard launches the wizard TUI and prints the result path.
func runInitWizard() error {
	parentDir, _ := os.Getwd()
	wiz := wizard.New(parentDir)
	final, err := runProgram(wiz, tea.WithAltScreen())
	if err != nil {
		return err
	}
	if wm, ok := final.(wizard.Model); ok {
		r := wm.GetResult()
		if !r.Aborted {
			fmt.Printf("✓ Wiki created: %s\n", r.WikiRoot)
		}
	}
	return nil
}
