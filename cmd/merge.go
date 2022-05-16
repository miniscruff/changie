package cmd

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

var _mergeDryRun = false

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

	out, err := mergePipeline(afs, afs.Create, _mergeDryRun)
	if err != nil {
		return err
	}

	if out != "" {
		_, err = cmd.OutOrStdout().Write([]byte(out))
	}

	return err
}

func mergePipeline(afs afero.Afero, creator shared.CreateFiler, dryRun bool) (string, error) {
	var builder strings.Builder

	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return "", err
	}

	allVersions, err := core.GetAllVersions(afs.ReadDir, config, false)
	if err != nil {
		return "", err
	}

	headerFile, err := afs.Open(filepath.Join(config.ChangesDir, config.HeaderPath))
	if err != nil {
		return "", err
	}

	defer headerFile.Close()

	_, _ = io.Copy(&builder, headerFile)

	for _, version := range allVersions {
		_, _ = builder.WriteString("\n\n")
		versionPath := filepath.Join(config.ChangesDir, version.Original()+"."+config.VersionExt)

		err = core.AppendFile(afs.Open, &builder, versionPath)
		if err != nil {
			return "", err
		}
	}

	if dryRun {
		return builder.String(), nil
	}

	changeFile, err := creator(config.ChangelogPath)
	if err != nil {
		return "", err
	}
	defer changeFile.Close()

	_, err = changeFile.Write([]byte(builder.String()))
	if err != nil {
		return "", err
	}

	if len(allVersions) == 0 {
		return "", nil
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
			return "", err
		}
	}

	return "", nil
}
