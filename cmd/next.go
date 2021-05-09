package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	errNotSupportedPart = errors.New("part string is not a supported next version increment")
)

var nextCmd = &cobra.Command{
	Use:   "next part",
	Short: "Next echos the next version based on semantic versioning",
	Long: `Next increment version based on semantic versioning.
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
	ver, err := latestPipeline(afs)
	if err != nil {
		return "", err
	}

	sver, _ := semver.NewVersion(strings.Trim(ver, "\n"))

	var next semver.Version

	switch part {
	case "major":
		next = sver.IncMajor()
	case "minor":
		next = sver.IncMinor()
	case "patch":
		next = sver.IncPatch()
	default:
		err = errNotSupportedPart
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintln(next.String()), nil
}
