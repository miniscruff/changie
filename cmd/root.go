package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

// rootCmd represents the base command when called without any subcommands
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changie",
		Short: "changie handles conflict-free changelog management",
		Long: `Changie keeps your changes organized and attached to your code.

Changie is aimed at seemlessly integrating into your release process while also
being easy to use for developers and your release team.`,
	}

	templateCache := core.NewTemplateCache()

	batch := NewBatch(
		os.ReadFile,
		os.ReadDir,
		os.Rename,
		os.WriteFile,
		os.MkdirAll,
		time.Now,
		templateCache,
	)

	merge := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		templateCache,
	)

	cmd.AddCommand(batch.Command)
	cmd.AddCommand(NewGen().Command)
	cmd.AddCommand(NewInit(os.MkdirAll, os.WriteFile).Command)
	cmd.AddCommand(NewLatest(os.ReadFile, os.ReadDir).Command)
	cmd.AddCommand(merge.Command)
	cmd.AddCommand(NewNew(os.ReadFile, time.Now, os.MkdirAll, templateCache).Command)
	cmd.AddCommand(NewNext(os.ReadDir, os.ReadFile).Command)

	return cmd
}
