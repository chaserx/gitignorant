/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"unicode/utf8"

	tmpl "github.com/chaserx/gitignorant/internal/template"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List available .gitignore templates",
	Example: "  ig list\n  ig list | grep -i java",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// 1. Load templates from embedded filesystem.
	templates, err := tmpl.LoadAll(TemplateFS)
	if err != nil {
		return fmt.Errorf("gitignore templates not found: %w", err)
	}

	// 2. Dedup to the priority winner per name and sort alphabetically.
	cat := tmpl.Catalog(templates)

	// 3. Render differently for a terminal vs. a pipe/redirect.
	for _, line := range renderList(cat, isTerminal(os.Stdout)) {
		fmt.Println(line)
	}
	return nil
}

// renderList formats the catalog into output lines. At a terminal it aligns
// tokens into a column and appends the source tag; when piped it emits bare
// tokens so the output composes cleanly with grep/fzf and feeds back into
// `ig ignore`.
func renderList(cat []tmpl.CatalogEntry, tty bool) []string {
	lines := make([]string, 0, len(cat))
	if !tty {
		for _, e := range cat {
			lines = append(lines, e.Token)
		}
		return lines
	}

	width := 0
	for _, e := range cat {
		if w := utf8.RuneCountInString(e.Token); w > width {
			width = w
		}
	}
	for _, e := range cat {
		lines = append(lines, fmt.Sprintf("%-*s  (%s)", width, e.Token, e.Source))
	}
	return lines
}

// isTerminal reports whether f is attached to an interactive terminal.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
