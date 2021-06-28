package cmd

import (
	"fmt"
	"strings"

	"github.com/miniscruff/changie/core"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next major|minor|patch",
	Short: "Next echos the next version based on semantic versioning",
	Long: `Next increments version based on semantic versioning.
Check latest version and increment part (major, minor, patch).
Echo the next release version number to be used by CI tools or other commands like batch.`,
	Args: cobra.ExactArgs(1),
	RunE: runNext,
}

func init() {
	rootCmd.AddCommand(nextCmd)
}

func runNext(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	result, err := nextPipeline(afs, strings.ToLower(args[0]))
	if err != nil {
		return err
	}

	_, err = cmd.OutOrStdout().Write([]byte(result))

	return err
}

func nextPipeline(afs afero.Afero, part string) (string, error) {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return "", err
	}

	next, err := core.GetNextVersion(afs.ReadDir, config, part)
	if err != nil {
		return "", err
	}

	return fmt.Sprintln(next.Original()), nil
}
