package cmd

import (
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
	cmd.SilenceUsage = true

	templateCache := core.NewTemplateCache()

	batch := NewBatch(time.Now, templateCache)
	merge := NewMerge(templateCache)

	cmd.AddCommand(batch.Command)
	cmd.AddCommand(NewGen().Command)
	cmd.AddCommand(NewInit().Command)
	cmd.AddCommand(NewLatest().Command)
	cmd.AddCommand(merge.Command)
	cmd.AddCommand(NewNew(time.Now, templateCache).Command)
	cmd.AddCommand(NewNext().Command)
	cmd.AddCommand(NewDiff().Command)

	return cmd
}
