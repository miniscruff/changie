package cmd

import (
	"errors"
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
	Project          string

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
	cmd.Flags().StringVarP(
		&m.Project,
		"project", "j",
		"",
		"Specify which project we are merging",
	)

	m.Command = cmd

	return m
}

//nolint:gocyclo
func (m *Merge) Run(cmd *cobra.Command, args []string) error {
	config, err := core.LoadConfig(m.ReadFile)
	if err != nil {
		return err
	}

    changelogPath := config.ChangelogPath

	if len(config.Projects) > 0 {
		if len(m.Project) == 0 {
			return errors.New("missing project label or key")
		}

		// make sure our passed in project is the key not the label
		for _, pc := range config.Projects {
			if m.Project == pc.Label {
				m.Project = pc.Key
			}

			if m.Project == pc.Key {
                changelogPath = pc.ChangelogPath
				break
			}
		}

        // todo: make sure our passed in project actually exists
	}

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

	allVersions, err := core.GetAllVersions(m.ReadDir, config, false, m.Project)
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

	if m.UnreleasedHeader != "" {
		var unrelErr error

		allChanges, unrelErr := core.GetChanges(
			config,
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
			_ = core.WriteNewlines(writer, config.Newlines.BeforeChangelogVersion)
			_ = core.WriteNewlines(writer, config.Newlines.BeforeVersion)
			_, _ = writer.Write([]byte(m.UnreleasedHeader))
			_ = core.WriteNewlines(writer, config.Newlines.AfterVersion)

			// create a fake batch to write the changes
			b := &Batch{
				config:        config,
				writer:        writer,
				TemplateCache: m.TemplateCache,
			}

			unrelErr = b.WriteChanges(allChanges)
			if unrelErr != nil {
				return unrelErr
			}

			_ = core.WriteNewlines(b.writer, b.config.Newlines.EndOfVersion)
			_ = core.WriteNewlines(writer, config.Newlines.AfterChangelogVersion)
		}
	}

	for _, version := range allVersions {
		_ = core.WriteNewlines(writer, config.Newlines.BeforeChangelogVersion)
		versionPath := filepath.Join(config.ChangesDir, m.Project, version.Original()+"."+config.VersionExt)

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
