package cmd

import (
	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"path/filepath"
	"sort"
	"strings"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge all versions into one changelog",
	RunE:  runMerge,
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}
	return mergePipeline(afs, afs.Create)
}

func mergePipeline(afs afero.Afero, creater CreateFiler) error {
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

	changeFile, err := creater(config.ChangelogPath)
	if err != nil {
		return err
	}
	defer changeFile.Close()

	appendFile(afs.Open, changeFile, filepath.Join(config.ChangesDir, config.HeaderPath))

	for _, version := range allVersions {
		changeFile.WriteString("\n")
		changeFile.WriteString("\n")
		appendFile(afs.Open, changeFile, filepath.Join(config.ChangesDir, version.Original()+"."+config.VersionExt))
	}

	return nil
}
