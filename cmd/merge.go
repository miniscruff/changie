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
	DryRun           bool
	UnreleasedHeader string

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
	cmd.Flags().StringVarP(
		&m.UnreleasedHeader,
		"include-unreleased", "u",
		"",
		"Include unreleased changes with this value as the header",
	)

	m.Command = cmd

	return m
}

//nolint:gocyclo
func (m *Merge) Run(cmd *cobra.Command, args []string) error {
	cfg, err := core.LoadConfig(m.ReadFile)
	if err != nil {
		return err
	}

	// If we have projects, merge all of them.
	if len(cfg.Projects) > 0 {
		for _, pc := range cfg.Projects {
			err = m.mergeProject(cmd, cfg, pc.Key, pc.ChangelogPath)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return m.mergeProject(cmd, cfg, "", cfg.ChangelogPath)
}

func (m *Merge) mergeProject(cmd *cobra.Command, cfg *core.Config, project, changelogPath string) error {
	var writer io.Writer
	if m.DryRun {
		writer = m.Command.OutOrStdout()
	} else {
		changeFile, changeErr := m.CreateFile(changelogPath)
		if changeErr != nil {
			return changeErr
		}
		defer changeFile.Close()
		writer = changeFile
	}

	allVersions, err := core.GetAllVersions(m.ReadDir, cfg, false, project)
	if err != nil {
		return err
	}

	if cfg.HeaderPath != "" {
		err = core.AppendFile(m.OpenFile, writer, filepath.Join(cfg.ChangesDir, cfg.HeaderPath))
		if err != nil {
			return err
		}

		_ = core.WriteNewlines(writer, cfg.Newlines.AfterChangelogHeader)
	}

	if m.UnreleasedHeader != "" {
		var unrelErr error

		allChanges, unrelErr := core.GetChanges(
			cfg,
			nil,
			m.ReadDir,
			m.ReadFile,
			"",
		)
		if unrelErr != nil {
			return unrelErr
		}

		// Make sure we have any changes before writing the unreleased content.
		if len(allChanges) > 0 {
			_ = core.WriteNewlines(writer, cfg.Newlines.BeforeChangelogVersion)
			_ = core.WriteNewlines(writer, cfg.Newlines.BeforeVersion)
			_, _ = writer.Write([]byte(m.UnreleasedHeader))
			_ = core.WriteNewlines(writer, cfg.Newlines.AfterVersion)

			// create a fake batch to write the changes
			b := &Batch{
				config:        cfg,
				writer:        writer,
				TemplateCache: m.TemplateCache,
			}

			unrelErr = b.WriteChanges(allChanges)
			if unrelErr != nil {
				return unrelErr
			}

			_ = core.WriteNewlines(b.writer, b.config.Newlines.EndOfVersion)
			_ = core.WriteNewlines(writer, cfg.Newlines.AfterChangelogVersion)
		}
	}

	for _, version := range allVersions {
		_ = core.WriteNewlines(writer, cfg.Newlines.BeforeChangelogVersion)
		versionPath := filepath.Join(cfg.ChangesDir, project, version.Original()+"."+cfg.VersionExt)

		err = core.AppendFile(m.OpenFile, writer, versionPath)
		if err != nil {
			return err
		}

		_ = core.WriteNewlines(writer, cfg.Newlines.AfterChangelogVersion)
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

	for _, rep := range cfg.Replacements {
		err = rep.Execute(m.ReadFile, m.WriteFile, replaceData)
		if err != nil {
			return err
		}
	}

	return nil
}
