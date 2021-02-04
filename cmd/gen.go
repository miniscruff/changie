package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate documentation",
	Long:  `Generate markdown documentation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doc.GenMarkdownTreeCustom(rootCmd, "website/content/cli", filePrepender, linkHandler)
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(genCmd)
}

const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
summary: "Help for using the '%s' command"
---
`

func filePrepender(filename string) string {
	now := time.Now().Format(time.RFC3339)
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	title := strings.ReplaceAll(base, "_", " ")
	url := "/cli/" + strings.ToLower(base) + "/"

	return fmt.Sprintf(fmTemplate, now, title, base, url, title)
}

func linkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "/cli/" + strings.ToLower(base) + "/"
}
