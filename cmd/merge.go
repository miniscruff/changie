package cmd

import (
	"path/filepath"

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

	changeFile, err := creator(config.ChangelogPath)
	if err != nil {
		return err
	}
	defer changeFile.Close()

	allVersions, err := getAllVersions(afs.ReadDir, config)
	if err != nil {
		return err
	}

	err = appendFile(afs.Open, changeFile, filepath.Join(config.ChangesDir, config.HeaderPath))
	if err != nil {
		return err
	}

	if len(allVersions) == 0 {
		return nil
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

	replaceData := ReplaceData{
		Version:         allVersions[0].Original(),
		VersionNoPrefix: allVersions[0].String(),
	}

	for _, rep := range config.Replacements {
		err = rep.Execute(afs.ReadFile, afs.WriteFile, replaceData)
		if err != nil {
			return err
		}
	}

	return nil
}
