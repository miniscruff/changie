package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

type Merge struct {
	*cobra.Command

	// cli args
	DryRun           bool
	UnreleasedHeader string

	// dependencies
	TemplateCache *core.TemplateCache
}

func NewMerge(templateCache *core.TemplateCache) *Merge {
	m := &Merge{
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

func (m *Merge) Run(cmd *cobra.Command, args []string) error {
	cfg, err := core.LoadConfig()
	if err != nil {
		return err
	}

	// If we have projects, merge all of them.
	if len(cfg.Projects) > 0 {
		for _, pc := range cfg.Projects {
			err = m.mergeProject(cfg, pc.Key, pc.ChangelogPath, pc.Replacements)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return m.mergeProject(cfg, "", cfg.ChangelogPath, cfg.Replacements)
}

func (m *Merge) mergeProject(
	cfg *core.Config,
	project, changelogPath string,
	replacements []core.Replacement,
) error {
	var writer io.Writer
	if m.DryRun {
		writer = m.OutOrStdout()
	} else {
		dirErr := os.MkdirAll(filepath.Dir(changelogPath), core.CreateDirMode)
		if dirErr != nil {
			return fmt.Errorf("creating changelog file directory: %w", dirErr)
		}

		changeFile, changeErr := os.Create(changelogPath)
		if changeErr != nil {
			return fmt.Errorf("creating changelog file: %w", changeErr)
		}

		defer changeFile.Close()

		writer = changeFile
	}

	allVersions, err := cfg.AllVersions(false, project)
	if err != nil {
		return fmt.Errorf("finding release notes: %w", err)
	}

	if cfg.HeaderPath != "" {
		err = core.AppendFile(writer, filepath.Join(cfg.ChangesDir, cfg.HeaderPath))
		if err != nil {
			return err
		}

		_ = core.WriteNewlines(writer, cfg.Newlines.AfterChangelogHeader)
	}

	if m.UnreleasedHeader != "" {
		var unrelErr error

		allChanges, unrelErr := cfg.Changes(nil, project)
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
				cfg:           cfg,
				writer:        writer,
				TemplateCache: m.TemplateCache,
			}

			unrelErr = b.WriteChanges(allChanges)
			if unrelErr != nil {
				return unrelErr
			}

			_ = core.WriteNewlines(b.writer, b.cfg.Newlines.EndOfVersion)
			_ = core.WriteNewlines(writer, cfg.Newlines.AfterChangelogVersion)
		}
	}

	for _, version := range allVersions {
		_ = core.WriteNewlines(writer, cfg.Newlines.BeforeChangelogVersion)
		versionPath := filepath.Join(cfg.ChangesDir, project, version.Original()+"."+cfg.VersionExt)

		err = core.AppendFile(writer, versionPath)
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
		Major:           int(version.Major()), //nolint:gosec
		Minor:           int(version.Minor()), //nolint:gosec
		Patch:           int(version.Patch()), //nolint:gosec
		Prerelease:      version.Prerelease(),
		Metadata:        version.Metadata(),
	}

	for _, rep := range replacements {
		err = rep.Execute(replaceData)
		if err != nil {
			return err
		}
	}

	return nil
}
