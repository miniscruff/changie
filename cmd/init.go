package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

var errConfigExists = errors.New("changie config already exists")

type Init struct {
	*cobra.Command

	// CLI args
	ChangesDir    string
	ChangelogPath string
	Force         bool
}

func NewInit() *Init {
	i := &Init{}

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
	// still gotta fix this lmao

	cfg := core.Config{
		////ChangesDir:    i.ChangesDir,
		//UnreleasedDir: "unreleased",
		//HeaderPath:    "header.tpl.md",
		//ChangelogPath: i.ChangelogPath,
		//VersionExt:    "md",
		//VersionFormat: "## {{.Version}} - {{.Time.Format \"2006-01-02\"}}",
		//KindFormat:    "### {{.Kind}}",
		//ChangeFormat:  "* {{.Body}}",
		//Kinds: []core.KindConfig{
		//{
		//Label:     "Added",
		//AutoLevel: core.MinorLevel,
		//},
		//{
		//Label:     "Changed",
		//AutoLevel: core.MajorLevel,
		//},
		//{
		//Label:     "Deprecated",
		//AutoLevel: core.MinorLevel,
		//},
		//{
		//Label:     "Removed",
		//AutoLevel: core.MajorLevel,
		//},
		//{
		//Label:     "Fixed",
		//AutoLevel: core.PatchLevel,
		//},
		//{
		//Label:     "Security",
		//AutoLevel: core.PatchLevel,
		//},
		//},
		//Newlines: core.NewlinesConfig{
		//AfterChangelogHeader:   1,
		//BeforeChangelogVersion: 1,
		//EndOfVersion:           1,
		//},
		// EnvPrefix: "CHANGIE_",
	}

	headerPath := filepath.Join(cfg.RootDir, cfg.Changelog.Header.FilePath)
	unreleasedPath := filepath.Join(cfg.RootDir, cfg.Fragment.Dir)
	keepPath := filepath.Join(unreleasedPath, ".gitkeep")

	if !i.Force {
		exists, existErr := cfg.Exists()
		if exists || existErr != nil {
			return errConfigExists
		}
	}

	err := cfg.Save()
	if err != nil {
		return err
	}

	err = os.MkdirAll(unreleasedPath, core.CreateDirMode)
	if err != nil {
		return err
	}

	err = os.WriteFile(keepPath, []byte{}, core.CreateFileMode)
	if err != nil {
		return err
	}

	err = os.WriteFile(headerPath, []byte(defaultHeader), core.CreateFileMode)
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.Changelog.Output, []byte(defaultChangelog), core.CreateFileMode)
	if err != nil {
		return err
	}

	return nil
}
