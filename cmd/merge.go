package cmd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

var (
	mergeDryRunOut io.Writer = os.Stdout
	_mergeDryRun             = false
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge all versions into one changelog",
	Long: `Merge all version files into one changelog file and run any replacement commands.

Note that a newline is added between each version file.`,
	RunE: runMerge,
}

func init() {
	mergeCmd.Flags().BoolVarP(
		&_mergeDryRun,
		"dry-run", "d",
		false,
		"Print merged changelog instead of writing to disk, will not run replacements",
	)
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return mergePipeline(afs)
}

func mergePipeline(afs afero.Afero) error {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	var writer io.Writer
	if _mergeDryRun {
		writer = mergeDryRunOut
	} else {
		changeFile, changeErr := afs.Create(config.ChangelogPath)
		if changeErr != nil {
			return changeErr
		}
		defer changeFile.Close()
		writer = changeFile
	}

	allVersions, err := core.GetAllVersions(afs.ReadDir, config, false)
	if err != nil {
		return err
	}

	if config.HeaderPath != "" {
		headerFile, headerErr := afs.Open(filepath.Join(config.ChangesDir, config.HeaderPath))
		if headerErr != nil {
			return headerErr
		}

		defer headerFile.Close()

		_, _ = io.Copy(writer, headerFile)
		_ = core.WriteNewlines(writer, config.Newlines.AfterChangelogHeader)
	}

	for _, version := range allVersions {
		_ = core.WriteNewlines(writer, config.Newlines.BeforeChangelogVersion)
		versionPath := filepath.Join(config.ChangesDir, version.Original()+"."+config.VersionExt)

		err = core.AppendFile(afs.Open, writer, versionPath)
		if err != nil {
			return err
		}

		_ = core.WriteNewlines(writer, config.Newlines.AfterChangelogVersion)
	}

	if len(allVersions) == 0 {
		return nil
	}

	if _mergeDryRun {
		return nil
	}

	version := allVersions[0]
	replaceData := core.ReplaceData{
		Version:         version.Original(),
		VersionNoPrefix: version.String(),
		Major:           int(version.Major()),
		Minor:           int(version.Minor()),
		Patch:           int(version.Patch()),
		Prerelease:      version.Prerelease(),
		Metadata:        version.Metadata(),
	}

	for _, rep := range config.Replacements {
		err = rep.Execute(afs.ReadFile, afs.WriteFile, replaceData)
		if err != nil {
			return err
		}
	}

	return nil
}
