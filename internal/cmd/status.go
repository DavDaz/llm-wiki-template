package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-template/internal/manifest"
	"github.com/DavDaz/llm-wiki-template/internal/tools"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the current wiki status from wiki.toml",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, _ []string) error {
	m, wikiRoot, err := loadManifestFromCwd()
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Wiki:             %s\n", m.Wiki.Name)
	fmt.Fprintf(w, "  Slug:             %s\n", m.Wiki.Slug)
	fmt.Fprintf(w, "  Language:         %s\n", m.Wiki.Language)
	fmt.Fprintf(w, "  Template version: %s\n", m.Wiki.TemplateVersion)
	fmt.Fprintf(w, "  Created:          %s\n", m.Wiki.CreatedAt)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Tools:\n")
	for _, tool := range tools.All() {
		installed := tool.IsInstalled(wikiRoot)
		enabledInManifest := isToolEnabled(m, tool.Name())
		icon := "○"
		if installed && enabledInManifest {
			icon = "●"
		} else if installed != enabledInManifest {
			icon = "!"
		}
		fmt.Fprintf(w, "    %s %s\n", icon, tool.Name())
	}
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Entities: %s\n", strings.Join(m.Domain.PrimaryEntities, ", "))
	fmt.Fprintf(w, "  Page types: %s\n", strings.Join(m.Domain.PageTypes, ", "))
	fmt.Fprintln(w, "")
	return nil
}

func isToolEnabled(m *manifest.Manifest, name string) bool {
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
