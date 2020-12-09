package cmd

import (
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

// No config options yet, coming later
/*
var (
	changesDir    string
	unreleasedDir string
	headerPath    string
	outputPath    string
)
*/

func init() {
	rootCmd.AddCommand(initCmd)

	/*
		initCmd.Flags().StringVarP(&changesDir, "dir", "d", "changes", "where to store intermediary change files")
		initCmd.Flags().StringVarP(&unreleasedDir, "unreleased", "u", "unreleased", "where to store unreleased changes")
		initCmd.Flags().StringVar(&headerPath, "header", "header.tpl.md", "header for output changelog")
		initCmd.Flags().StringVarP(&outputPath, "output", "o", "CHANGELOG.md", "where to output our changelog")
	*/
}

func runInit(cmd *cobra.Command, args []string) error {
	// TODO: Add testing once config loading is in place otherwise it will
	// not be idempotent
	fs := afero.NewOsFs()
	config := ProjectConfig{}
	p := NewProject(fs, config)
	return p.Init()
}
