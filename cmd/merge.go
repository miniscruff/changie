package cmd

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
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

func mergePipeline(afs afero.Afero, creator CreateFiler) error {
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

		var v *semver.Version

		v, err = semver.NewVersion(versionString)
		if err != nil {
			continue
		}

		allVersions = append(allVersions, v)
	}

	sort.Sort(sort.Reverse(semver.Collection(allVersions)))

	changeFile, err := creator(config.ChangelogPath)
	if err != nil {
		return err
	}
	defer changeFile.Close()

	err = appendFile(afs.Open, changeFile, filepath.Join(config.ChangesDir, config.HeaderPath))
	if err != nil {
		return err
	}

	for _, version := range allVersions {
		_, err = changeFile.WriteString("\n\n")
		if err != nil {
			return err
		}

		versionPath := filepath.Join(config.ChangesDir, version.Original()+"."+config.VersionExt)

		err = appendFile(afs.Open, changeFile, versionPath)
		if err != nil {
			return err
		}
	}

	return nil
}
