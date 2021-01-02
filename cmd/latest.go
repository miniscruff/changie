package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// latestCmd represents the latest command
var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Echos the latest release version number",
	Long:  `Echo the latest release version number to be used by CI tools.`,
	RunE:  runLatest,
}

var removePrefix bool = false

func init() {
	rootCmd.AddCommand(latestCmd)

	latestCmd.Flags().BoolVarP(
		&removePrefix,
		"remove-prefix",
		"r",
		false,
		"Remove 'v' prefix before echoing",
	)
}

func runLatest(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	result, err := latestPipeline(afs)
	if err != nil {
		return err
	}

	_, err = cmd.OutOrStdout().Write([]byte(result))

	return err
}

func latestPipeline(afs afero.Afero) (string, error) {
	config, err := LoadConfig(afs.ReadFile)
	if err != nil {
		return "", err
	}

	allVersions, err := getAllVersions(afs.ReadDir, config)
	if err != nil {
		return "", err
	}

	// if no versions exist default to v0.0.0
	if len(allVersions) == 0 {
		return "v0.0.0\n", nil
	}

	if removePrefix {
		return fmt.Sprintln(strings.TrimPrefix(allVersions[0].Original(), "v")), nil
	}

	return fmt.Sprintln(allVersions[0].Original()), nil
}
