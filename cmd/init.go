package cmd

import (
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

var errConfigExists = errors.New("changie config already exists")

type Init struct {
	*cobra.Command

	// CLI args
	ChangesDir    string
	ChangelogPath string
	Force         bool

	// dependencies
	MkdirAll  shared.MkdirAller
	WriteFile shared.WriteFiler
}

func NewInit(
	mkdirAll shared.MkdirAller,
	writeFile shared.WriteFiler,
) *Init {
	i := &Init{
		MkdirAll:  mkdirAll,
		WriteFile: writeFile,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new changie skeleton",
		Long: `Initialize a few changie specifics.

* Folder to place all change files
* Subfolder to place all unreleased changes
* Header file to place on top of the changelog
* Output file when generating a changelog
* Unreleased folder includes a .gitkeep file

Values will also be saved in a changie config at .changie.yaml.
Default values follow keep a changelog and semver specs but are customizable.`,
		Args: cobra.NoArgs,
		RunE: i.Run,
	}
	cmd.Flags().StringVarP(&i.ChangesDir, "dir", "d", ".changes", "directory for all changes")
	cmd.Flags().StringVarP(
		&i.ChangelogPath,
		"output",
		"o",
		"CHANGELOG.md",
		"file path to output our changelog",
	)
	cmd.Flags().BoolVarP(&i.Force, "force", "f", false, "force initialize even if config already exist")

	i.Command = cmd

	return i
}

func (i *Init) Run(cmd *cobra.Command, args []string) error {
	config := core.Config{
		ChangesDir:    i.ChangesDir,
		EnvPrefix:     "CHANGIE_",
		UnreleasedDir: "unreleased",
		HeaderPath:    "header.tpl.md",
		ChangelogPath: i.ChangelogPath,
		VersionExt:    "md",
		VersionFormat: "## {{.Version}} - {{.Time.Format \"2006-01-02\"}}",
		KindFormat:    "### {{.Kind}}",
		ChangeFormat:  "* {{.Body}}",
		/*
			Kinds: []core.KindConfig{
				{
					Label:     "Added",
					AutoLevel: core.MinorLevel,
				},
				{
					Label:     "Changed",
					AutoLevel: core.MajorLevel,
				},
				{
					Label:     "Deprecated",
					AutoLevel: core.MinorLevel,
				},
				{
					Label:     "Removed",
					AutoLevel: core.MajorLevel,
				},
				{
					Label:     "Fixed",
					AutoLevel: core.PatchLevel,
				},
				{
					Label:     "Security",
					AutoLevel: core.PatchLevel,
				},
			},
			Newlines: core.NewlinesConfig{
				AfterChangelogHeader:   1,
				BeforeChangelogVersion: 1,
				EndOfVersion:           1,
			},
		*/
	}

	headerPath := filepath.Join(config.ChangesDir, config.HeaderPath)
	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)
	keepPath := filepath.Join(unreleasedPath, ".gitkeep")

	if !i.Force {
		exists, existErr := config.Exists()
		if exists || existErr != nil {
			return errConfigExists
		}
	}

	err := config.Save(i.WriteFile)
	if err != nil {
		return err
	}

	err = i.MkdirAll(unreleasedPath, core.CreateDirMode)
	if err != nil {
		return err
	}

	err = i.WriteFile(keepPath, []byte{}, core.CreateFileMode)
	if err != nil {
		return err
	}

	err = i.WriteFile(headerPath, []byte(defaultHeader), core.CreateFileMode)
	if err != nil {
		return err
	}

	err = i.WriteFile(config.ChangelogPath, []byte(defaultChangelog), core.CreateFileMode)
	if err != nil {
		return err
	}

	return nil
}
