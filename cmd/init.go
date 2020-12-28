package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
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
	changesDir    string
	changelogPath string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&changesDir, "dir", "d", "changes", "directory for all changes")
	initCmd.Flags().StringVarP(
		&changelogPath,
		"output",
		"o",
		"CHANGELOG.md",
		"file path to output our changelog",
	)
}

func runInit(cmd *cobra.Command, args []string) error {
	config := Config{
		ChangesDir:    changesDir,
		UnreleasedDir: "unreleased",
		HeaderPath:    "header.tpl.md",
		ChangelogPath: changelogPath,
		VersionExt:    "md",
		VersionFormat: "## {{.Version}} - {{.Time.Format \"2006-01-02\"}}",
		KindFormat:    "### {{.Kind}}",
		ChangeFormat:  "* {{.Body}}",
		Kinds: []string{
			"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security",
		},
	}

	afs := afero.Afero{Fs: afero.NewOsFs()}

	return initPipeline(afs.MkdirAll, afs.WriteFile, config)
}

func initPipeline(mkdir MkdirAller, wf WriteFiler, config Config) error {
	var err error

	headerPath := filepath.Join(config.ChangesDir, config.HeaderPath)
	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)
	keepPath := filepath.Join(unreleasedPath, ".gitkeep")

	err = config.Save(wf)
	if err != nil {
		return err
	}

	err = mkdir(unreleasedPath, 0644)
	if err != nil {
		return err
	}

	err = wf(keepPath, []byte{}, os.ModePerm)
	if err != nil {
		return err
	}

	err = wf(headerPath, []byte(defaultHeader), os.ModePerm)
	if err != nil {
		return err
	}

	err = wf(config.ChangelogPath, []byte(defaultChangelog), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
