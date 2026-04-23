package cmd

import (
	"fmt"
	"os"

	"charm.land/glamour/v2"
	"github.com/spf13/cobra"

	"github.com/DavDaz/llm-wiki-generator/internal/templates"
)

var guideCmd = &cobra.Command{
	Use:   "guide",
	Short: "Show the conceptual guide — what each field means and why it matters",
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

	out, err := glamour.RenderWithEnvironmentConfig(string(raw))
	if err != nil {
		// fallback: print raw if glamour fails (e.g. no TTY)
		fmt.Fprintln(os.Stdout, string(raw))
		return nil
	}

	fmt.Print(out)
	return nil
}
