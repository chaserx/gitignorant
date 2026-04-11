/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chaserx/gitignorant/internal/gitignore"
	tmpl "github.com/chaserx/gitignorant/internal/template"
	"github.com/spf13/cobra"
)

const gitignorePath = ".gitignore"

var ignoreCmd = &cobra.Command{
	Use:     "ignore [languages...]",
	Short:   "Generate or update a .gitignore file for specific languages or frameworks",
	Example: "  ig ignore go\n  ig ignore go python node",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"i"},
	RunE:    runIgnore,
}

func init() {
	rootCmd.AddCommand(ignoreCmd)
}

func runIgnore(cmd *cobra.Command, args []string) error {
	// 1. Load templates from embedded filesystem.
	templates, err := tmpl.LoadAll(TemplateFS)
	if err != nil {
		return fmt.Errorf("gitignore templates not found: %w", err)
	}

	// 2. Resolve arguments against available templates.
	matched, unmatched := tmpl.Resolve(templates, args)

	// 3. Handle unmatched arguments.
	if len(unmatched) > 0 {
		fmt.Fprintf(os.Stderr, "No templates found for: %s\n", strings.Join(unmatched, ", "))
		if len(matched) == 0 {
			return fmt.Errorf("no matching templates found")
		}
		names := templateNames(matched)
		if !confirm(fmt.Sprintf("Continue with %s? (y/n) ", strings.Join(names, ", "))) {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// 4. Read existing .gitignore.
	existing, err := gitignore.Read(gitignorePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", gitignorePath, err)
	}

	// 5. Compute new lines for each template, accumulating existing lines.
	var allNewLines []string
	accumulated := existing

	for _, t := range matched {
		newLines := gitignore.NewLines(accumulated, t.Content)
		if newLines == nil {
			fmt.Printf("%s patterns are already present, nothing to add.\n", t.Name)
			continue
		}
		formatted := gitignore.Format(t.Name, newLines)
		allNewLines = append(allNewLines, formatted...)
		accumulated = append(accumulated, formatted...)
	}

	// 6. Check if there is anything to add.
	if len(allNewLines) == 0 {
		fmt.Println("All patterns are already present. Nothing to do.")
		return nil
	}

	// 7. Preview lines to be added.
	fmt.Println("The following lines will be added to .gitignore:")
	fmt.Println(strings.Repeat("-", 40))
	for _, line := range allNewLines {
		fmt.Println(line)
	}
	fmt.Println(strings.Repeat("-", 40))

	// 8-9. Prompt for permission.
	fileExists := gitignore.Exists(gitignorePath)
	var prompt string
	if fileExists {
		prompt = "Append the above to .gitignore? (y/n) "
	} else {
		prompt = "No .gitignore found. Create one with the above content? (y/n) "
	}
	if !confirm(prompt) {
		fmt.Println("Aborted.")
		return nil
	}

	// 10. Write.
	if err := gitignore.Write(gitignorePath, allNewLines); err != nil {
		return fmt.Errorf("writing %s: %w", gitignorePath, err)
	}

	// 11. Success.
	if fileExists {
		fmt.Println("Updated .gitignore successfully.")
	} else {
		fmt.Println("Created .gitignore successfully.")
	}
	return nil
}

func templateNames(templates []tmpl.Template) []string {
	names := make([]string, len(templates))
	for i, t := range templates {
		names[i] = t.Name
	}
	return names
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}
