/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ignoreCmd represents the ignore command
var ignoreCmd = &cobra.Command{
	Use:   "ignore [language]",
	Short: "Generate or update a .gitignore file for a specific language or framework",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Println("ignore called", args[0])
		} else {
			fmt.Println("ignore called")
		}
	},
}

func init() {
	rootCmd.AddCommand(ignoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ignoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ignoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
