package cmd

import (
	"errors"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
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
	RunE: runInit,
}

var (
	changesDir      string
	changelogPath   string
	initForce       bool
	errConfigExists = errors.New("changie config already exists")
)

func init() {
	initCmd.Flags().StringVarP(&changesDir, "dir", "d", ".changes", "directory for all changes")
	initCmd.Flags().StringVarP(
		&changelogPath,
		"output",
		"o",
		"CHANGELOG.md",
		"file path to output our changelog",
	)
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "force initialize even if config already exist")
}

func runInit(cmd *cobra.Command, args []string) error {
	config := core.Config{
		ChangesDir:    changesDir,
		UnreleasedDir: "unreleased",
		HeaderPath:    "header.tpl.md",
		ChangelogPath: changelogPath,
		VersionExt:    "md",
		VersionFormat: "## {{.Version}} - {{.Time.Format \"2006-01-02\"}}",
		KindFormat:    "### {{.Kind}}",
		ChangeFormat:  "* {{.Body}}",
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
		EnvPrefix: "CHANGIE_",
	}

	afs := afero.Afero{Fs: afero.NewOsFs()}

	return initPipeline(afs.MkdirAll, afs.WriteFile, afs.Exists, config)
}

func initPipeline(mkdir shared.MkdirAller, wf shared.WriteFiler, fe shared.FileExister, config core.Config) error {
	var err error

	headerPath := filepath.Join(config.ChangesDir, config.HeaderPath)
	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)
	keepPath := filepath.Join(unreleasedPath, ".gitkeep")

	if !initForce {
		exists, existErr := config.Exists(fe)
		if exists || existErr != nil {
			return errConfigExists
		}
	}

	err = config.Save(wf)
	if err != nil {
		return err
	}

	err = mkdir(unreleasedPath, core.CreateDirMode)
	if err != nil {
		return err
	}

	err = wf(keepPath, []byte{}, core.CreateFileMode)
	if err != nil {
		return err
	}

	err = wf(headerPath, []byte(defaultHeader), core.CreateFileMode)
	if err != nil {
		return err
	}

	err = wf(config.ChangelogPath, []byte(defaultChangelog), core.CreateFileMode)
	if err != nil {
		return err
	}

	return nil
}
