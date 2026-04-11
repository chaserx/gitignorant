/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/chaserx/gitignorant/cmd"
)

func main() {
	cmd.TemplateFS = gitignoreFS
	cmd.Execute()
}
