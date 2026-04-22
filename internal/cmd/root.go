// Package cmd defines all CLI subcommands for llm-wiki.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-template/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "llm-wiki",
	Short: "Manage LLM wikis — create, migrate tools, edit domain config",
	Long: `llm-wiki is a CLI + TUI for creating and managing LLM wikis.

It replaces the legacy setup.sh with a versioned, distributable binary
that supports Claude Code, OpenCode, and Pi as AI tool backends.`,
	// Default: no-arg invocation will open TUI (wired in Phase 2).
	SilenceUsage: true,
}

// Execute is the entry point called by main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the llm-wiki version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version.Version)
	},
}
