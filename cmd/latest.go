package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

// latestCmd represents the latest command
var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Echos the latest release version number",
	Long:  `Echo the latest release version number to be used by CI tools.`,
	RunE:  runLatest,
}

var (
	removePrefix          bool = false
	latestSkipPrereleases bool = false
)

func init() {
	rootCmd.AddCommand(latestCmd)

	latestCmd.Flags().BoolVarP(
		&removePrefix,
		"remove-prefix", "r",
		false,
		"Remove 'v' prefix before echoing",
	)
	latestCmd.Flags().BoolVarP(
		&latestSkipPrereleases,
		"skip-prereleases", "",
		false,
		"Excludes prereleases to determine the latest version.",
	)
}

func runLatest(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	result, err := latestPipeline(afs, latestSkipPrereleases)
	if err != nil {
		return err
	}

	_, err = cmd.OutOrStdout().Write([]byte(result))

	return err
}

func latestPipeline(afs afero.Afero, skipPrereleases bool) (string, error) {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return "", err
	}

	ver, err := core.GetLatestVersion(afs.ReadDir, config, skipPrereleases)
	if err != nil {
		return "", err
	}

	if removePrefix {
		return fmt.Sprintln(strings.TrimPrefix(ver.Original(), "v")), nil
	}

	return fmt.Sprintln(ver.Original()), nil
}
