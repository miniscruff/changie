package cmd

import (
	"os"
	"time"

	"github.com/spf13/afero"
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

	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}
	templateCache := core.NewTemplateCache()

	cmd.AddCommand(batchCmd)
	cmd.AddCommand(genCmd)
	cmd.AddCommand(initCmd)
	cmd.AddCommand(latestCmd)
	cmd.AddCommand(NewMerge(afs.ReadFile, afs.WriteFile, afs.ReadDir, afs.Open, os.Create, templateCache).Command)
	cmd.AddCommand(NewNew(afs.ReadFile, os.Create, time.Now, templateCache).Command)
	cmd.AddCommand(NewNext(afs.ReadDir, afs.ReadFile).Command)

	return cmd
}
