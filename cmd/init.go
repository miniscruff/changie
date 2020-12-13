package cmd

import (
	"fmt"
	"github.com/miniscruff/changie/project"
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

Values will also be saved in a changie config at .changie.yaml`,
	RunE: runInit,
}

var (
	changesDir    string
	changelogPath string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&changesDir, "dir", "d", "changes", "directory for all changes")
	initCmd.Flags().StringVarP(&changelogPath, "output", "o", "CHANGELOG.md", "file path to output our changelog")
}

func runInit(cmd *cobra.Command, args []string) error {
	var err error

	fs := afero.NewOsFs()
	config := project.Config{
		ChangesDir:    changesDir,
		UnreleasedDir: "unreleased",
		HeaderPath:    "header.tpl.md",
		ChangelogPath: changelogPath,
		VersionFormat: "## :rocket: {{Version}}",
		KindFormat:    "### {{Kind}}",
		ChangeFormat:  "* {{.Body}}",
		Kinds: []string{
			"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security",
		},
	}

	err = config.Save(fs)
	if err != nil {
		return err
	}

	err = fs.MkdirAll(fmt.Sprintf("%s/%s", config.ChangesDir, config.UnreleasedDir), 644)
	if err != nil {
		return err
	}

	keepFile, err := fs.Create(fmt.Sprintf("%s/%s/.gitkeep", config.ChangesDir, config.UnreleasedDir))
	if err != nil {
		return err
	}
	defer keepFile.Close()

	headerFile, err := fs.Create(fmt.Sprintf("%s/%s", config.ChangesDir, config.HeaderPath))
	if err != nil {
		return err
	}
	defer headerFile.Close()

	_, err = headerFile.WriteString(defaultHeader)
	if err != nil {
		return err
	}

	outputFile, err := fs.Create(config.ChangelogPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(defaultChangelog)
	if err != nil {
		return err
	}

	return nil
}
