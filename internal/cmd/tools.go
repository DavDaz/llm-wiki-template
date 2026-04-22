package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-template/internal/generator"
	"github.com/DavDaz/llm-wiki-template/internal/tools"
)

var addToolCmd = &cobra.Command{
	Use:     "add-tool <name>",
	Short:   "Enable a tool backend for this wiki (claude-code | opencode | pi)",
	Args:    cobra.ExactArgs(1),
	RunE:    runAddTool,
	Example: "  llm-wiki add-tool opencode",
}

var removeToolCmd = &cobra.Command{
	Use:     "remove-tool <name>",
	Short:   "Disable a tool backend for this wiki",
	Args:    cobra.ExactArgs(1),
	RunE:    runRemoveTool,
	Example: "  llm-wiki remove-tool pi",
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply manifest changes to the filesystem (re-render instruction files, sync tools)",
	RunE:  runMigrate,
}

func init() {
	rootCmd.AddCommand(addToolCmd)
	rootCmd.AddCommand(removeToolCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runAddTool(cmd *cobra.Command, args []string) error {
	m, wikiRoot, err := loadManifestFromCwd()
	if err != nil {
		return err
	}

	toolName := args[0]
	if _, err := tools.Get(toolName); err != nil {
		return err
	}

	setToolEnabled(m, toolName, true)

	if err := m.Save(wikiRoot); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}
	if err := generator.Migrate(wikiRoot, m); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ %s enabled\n", toolName)
	return nil
}

func runRemoveTool(cmd *cobra.Command, args []string) error {
	m, wikiRoot, err := loadManifestFromCwd()
	if err != nil {
		return err
	}

	toolName := args[0]
	if _, err := tools.Get(toolName); err != nil {
		return err
	}

	setToolEnabled(m, toolName, false)

	if err := m.Validate(); err != nil {
		return fmt.Errorf("cannot remove tool: %w", err)
	}
	if err := m.Save(wikiRoot); err != nil {
		return err
	}
	if err := generator.Migrate(wikiRoot, m); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ %s disabled\n", toolName)
	return nil
}

func runMigrate(cmd *cobra.Command, _ []string) error {
	m, wikiRoot, err := loadManifestFromCwd()
	if err != nil {
		return err
	}
	if err := generator.Migrate(wikiRoot, m); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "✓ Migration complete")
	return nil
}

