package cmd

import (
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

type Merge struct {
	*cobra.Command

	// cli args
	DryRun bool

	// dependencies
	ReadFile      shared.ReadFiler
	WriteFile     shared.WriteFiler
	ReadDir       shared.ReadDirer
	OpenFile      shared.OpenFiler
	CreateFile    shared.CreateFiler
	TemplateCache *core.TemplateCache
}

func NewMerge(
	readFile shared.ReadFiler,
	writeFile shared.WriteFiler,
	readDir shared.ReadDirer,
	openFile shared.OpenFiler,
	createFile shared.CreateFiler,
	templateCache *core.TemplateCache,
) *Merge {
	m := &Merge{
		ReadFile:      readFile,
		WriteFile:     writeFile,
		ReadDir:       readDir,
		OpenFile:      openFile,
		CreateFile:    createFile,
		TemplateCache: templateCache,
	}

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge all versions into one changelog",
		Long: `Merge all version files into one changelog file and run any replacement commands.

Note that a newline is added between each version file.`,
        Args: cobra.NoArgs,
		RunE: m.Run,
	}

	cmd.Flags().BoolVarP(
		&m.DryRun,
		"dry-run", "d",
		false,
		"Print merged changelog instead of writing to disk, will not run replacements",
	)

	m.Command = cmd
	return m
}

func (m *Merge) Run(cmd *cobra.Command, args []string) error {
	config, err := core.LoadConfig(m.ReadFile)
	if err != nil {
		return err
	}

	var writer io.Writer
	if m.DryRun {
		writer = m.Command.OutOrStdout()
	} else {
		changeFile, changeErr := m.CreateFile(config.ChangelogPath)
		if changeErr != nil {
			return changeErr
		}
		defer changeFile.Close()
		writer = changeFile
	}

	allVersions, err := core.GetAllVersions(m.ReadDir, config, false)
	if err != nil {
		return err
	}

	if config.HeaderPath != "" {
		err = core.AppendFile(m.OpenFile, writer, filepath.Join(config.ChangesDir, config.HeaderPath))
		if err != nil {
			return err
		}

		_ = core.WriteNewlines(writer, config.Newlines.AfterChangelogHeader)
	}

	for _, version := range allVersions {
		_ = core.WriteNewlines(writer, config.Newlines.BeforeChangelogVersion)
		versionPath := filepath.Join(config.ChangesDir, version.Original()+"."+config.VersionExt)

		err = core.AppendFile(m.OpenFile, writer, versionPath)
		if err != nil {
			return err
		}

		_ = core.WriteNewlines(writer, config.Newlines.AfterChangelogVersion)
	}

	if len(allVersions) == 0 {
		return nil
	}

	if m.DryRun {
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
		err = rep.Execute(m.ReadFile, m.WriteFile, replaceData)
		if err != nil {
			return err
		}
	}

	return nil
}
