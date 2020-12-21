package cmd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"path/filepath"
	"sort"
	"strings"
)

// latestCmd represents the latest command
var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Echos the latest release version number",
	Long:  `Echo the latest release version number to be used by CI tools.`,
	RunE:  runLatest,
}

var removePrefix string = ""

func init() {
	rootCmd.AddCommand(latestCmd)

	latestCmd.Flags().StringVarP(&removePrefix, "remove-prefix", "r", "", "Remove specified prefix from output")
}

func runLatest(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	config, err := LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	allVersions := make([]*semver.Version, 0)

	fileInfos, err := afs.ReadDir(config.ChangesDir)
	if err != nil {
		return err
	}

	for _, file := range fileInfos {
		if file.Name() == config.HeaderPath || file.IsDir() {
			continue
		}

		versionString := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		v, err := semver.NewVersion(versionString)
		if err != nil {
			continue
		}

		allVersions = append(allVersions, v)
	}

	sort.Sort(sort.Reverse(semver.Collection(allVersions)))

	if removePrefix != "" {
		fmt.Println(strings.TrimPrefix(allVersions[0].Original(), removePrefix))
	} else {
		fmt.Println(allVersions[0].Original())
	}
	return nil
}
