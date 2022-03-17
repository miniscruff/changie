package cmd

import (
	"io"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

var (
	nextPrereleaseFlag []string = nil
	nextMetaFlag       []string = nil
)

var nextCmd = &cobra.Command{
	Use:   "next major|minor|patch",
	Short: "Next echos the next version based on semantic versioning",
	Long: `Next increments version based on semantic versioning.
Check latest version and increment part (major, minor, patch).
Echo the next release version number to be used by CI tools or other commands like batch.`,
	ValidArgs: []string{"major", "minor", "patch"},
	Args:      cobra.ExactValidArgs(1),
	RunE:      runNext,
}

func init() {
	nextCmd.Flags().StringSliceVarP(
		&nextPrereleaseFlag,
		"prerelease", "p",
		nil,
		"Prerelease values to append to version",
	)
	nextCmd.Flags().StringSliceVarP(
		&nextMetaFlag,
		"metadata", "m",
		nil,
		"Metadata values to append to version",
	)

	rootCmd.AddCommand(nextCmd)
}

func runNext(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return nextPipeline(
		afs,
		cmd.OutOrStdout(),
		strings.ToLower(args[0]),
		nextPrereleaseFlag,
		nextMetaFlag,
	)
}

func nextPipeline(afs afero.Afero, writer io.Writer, part string, prerelease, meta []string) error {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	next, err := core.GetNextVersion(afs.ReadDir, config, part, prerelease, meta)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(next.Original()))

	return err
}
