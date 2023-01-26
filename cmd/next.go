package cmd

import (
	"io"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

var (
	nextIncludeDirs    []string = nil
	nextPrereleaseFlag []string = nil
	nextMetaFlag       []string = nil
)

var nextCmd = &cobra.Command{
	Use:   "next major|minor|patch",
	Short: "Next echos the next version based on semantic versioning",
	Long: `Next increments version based on semantic versioning.
Check latest version and increment part (major, minor, patch).
Echo the next release version number to be used by CI tools or other commands like batch.`,
	ValidArgs: []string{"major", "minor", "patch", "auto"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:      runNext,
}

func init() {
	nextCmd.Flags().StringSliceVarP(
		&nextIncludeDirs,
		"include", "i",
		nil,
		"Include extra directories to search for change files, relative to change directory",
	)
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
		nextIncludeDirs,
	)
}

func nextPipeline(
	afs afero.Afero,
	writer io.Writer,
	part string,
	prerelease, meta, includeDirs []string) error {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

    var changes []core.Change
    // only worry about loading changes, if we are in auto mode
	if part == core.AutoLevel {
		changes, err = core.GetChanges(config, includeDirs, afs.ReadDir, afs.ReadFile)
		if err != nil {
			return err
		}
	}

	next, err := core.GetNextVersion(afs.ReadDir, config, part, prerelease, meta, changes)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(next.Original()))

	return err
}
