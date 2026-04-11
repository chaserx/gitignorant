/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"io/fs"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// TemplateFS holds the embedded gitignore template filesystem.
// Set by main before calling Execute.
var TemplateFS fs.FS

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ig",
	Short: "Ignorant is a CLI tool to generate .gitignore files for various languages and frameworks",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// err := rootCmd.Execute()
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gitignorant.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
