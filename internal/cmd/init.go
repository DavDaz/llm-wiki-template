package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-generator/internal/generator"
	"github.com/DavDaz/llm-wiki-generator/internal/tui/wizard"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new wiki (TUI wizard or headless with flags)",
	Example: `  llm-wiki init
  llm-wiki init --name "Legal Wiki" --slug legal-wiki --tools claude-code,opencode
  llm-wiki init --name "RENAB" --slug renab --tools all --entities "usuario,rol,permiso"`,
	RunE: runInit,
}

var initFlags struct {
	name        string
	slug        string
	language    string
	tools       string
	entities    string
	pageTypes   string
	conventions string
	parentDir   string
}

func init() {
	f := initCmd.Flags()
	f.StringVar(&initFlags.name, "name", "", "Wiki name")
	f.StringVar(&initFlags.slug, "slug", "", "Wiki slug — kebab-case identifier")
	f.StringVar(&initFlags.language, "lang", "es", "Language code (default: es)")
	f.StringVar(&initFlags.tools, "tools", "claude-code", "Comma-separated tools: claude-code,opencode,pi or 'all'")
	f.StringVar(&initFlags.entities, "entities", "", "Comma-separated primary entities (e.g. usuario,rol)")
	f.StringVar(&initFlags.pageTypes, "page-types", "proceso,referencia,entidad,politica", "Comma-separated page types")
	f.StringVar(&initFlags.conventions, "conventions", "", "Comma-separated domain conventions")
	f.StringVar(&initFlags.parentDir, "dir", "", "Parent directory for the new wiki (default: current dir)")

	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, _ []string) error {
	// no required flags provided → open TUI wizard
	if initFlags.name == "" && initFlags.slug == "" {
		parentDir := initFlags.parentDir
		if parentDir == "" {
			parentDir, _ = os.Getwd()
		}
		m := wizard.New(parentDir)
		p := tea.NewProgram(m, tea.WithAltScreen())
		final, err := p.Run()
		if err != nil {
			return err
		}
		if wm, ok := final.(wizard.Model); ok {
			r := wm.GetResult()
			if !r.Aborted {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Wiki created: %s\n", r.WikiRoot)
			}
		}
		return nil
	}

	// headless mode — both --name and --slug required
	if initFlags.name == "" || initFlags.slug == "" {
		return fmt.Errorf("--name and --slug are both required in headless mode")
	}

	toolNames := parseCSV(initFlags.tools)
	claude, opencode, pi := resolveTools(toolNames)
	if !claude && !opencode && !pi {
		return fmt.Errorf("no valid tools specified — use: claude-code, opencode, pi, or all")
	}

	cfg := generator.InitConfig{
		ParentDir:       initFlags.parentDir,
		Name:            initFlags.name,
		Slug:            initFlags.slug,
		Language:        initFlags.language,
		ClaudeCode:      claude,
		OpenCode:        opencode,
		Pi:              pi,
		PrimaryEntities: parseCSV(initFlags.entities),
		PageTypes:       parseCSV(initFlags.pageTypes),
		Conventions:     parseCSV(initFlags.conventions),
	}

	wikiRoot, err := generator.Init(cfg)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Wiki created: %s\n", wikiRoot)
	return nil
}

// resolveTools maps tool name strings to booleans, handling the "all" shortcut.
func resolveTools(names []string) (claude, opencode, pi bool) {
	for _, n := range names {
		switch strings.TrimSpace(strings.ToLower(n)) {
		case "all":
			return true, true, true
		case "claude-code", "claude":
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
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
