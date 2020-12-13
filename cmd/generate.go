package cmd

import (
	"errors"
	"fmt"
	"github.com/miniscruff/changie/project"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [version]",
	Short: "Generate a new version changelog and a new changelog.md",
	Long: `Merges all unreleased changes into one changelog and adds
it to the main changelog file.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("Missing required version")
	}

	ver := args[0]
	if !semver.IsValid(ver) {
		return errors.New("Invalid version, must be a semantic version")
	}

	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	config, err := project.LoadConfig(fs)
	if err != nil {
		return err
	}

	changesPerKind := make(map[string][]project.Change, 0)

	// read all markdown files from changes/unreleased
	absDir := fmt.Sprintf("%s/%s", config.ChangesDir, config.UnreleasedDir)
	err = filepath.Walk(absDir, func(path string, file os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		c, err := project.LoadChange(path, afs)
		if err != nil {
			return err
		}

		if _, ok := changesPerKind[c.Kind]; !ok {
			changesPerKind[c.Kind] = make([]project.Change, 0)
		}

		changesPerKind[c.Kind] = append(changesPerKind[c.Kind], c)

		return nil
	})
	if err != nil {
		return err
	}

	// delete all markdown files in unreleased ( keeping .gitkeep )

	// merge them into one version.md file in changes/
	versionFile, err := fs.Create(fmt.Sprintf("%s/%s.yaml", config.ChangesDir, ver))
	if err != nil {
		return err
	}
	defer versionFile.Close()

	vTempl, _ := template.New("version").Parse(config.VersionFormat)
	kTempl, _ := template.New("kind").Parse(config.KindFormat)
	cTempl, _ := template.New("change").Parse(config.ChangeFormat)

	err = vTempl.Execute(versionFile, map[string]string{
		"Version": ver,
	})
	if err != nil {
		return err
	}

	for kind, changes := range changesPerKind {
		_, err = versionFile.WriteString("\n")
		_, err = versionFile.WriteString("\n")
		if err != nil {
			return err
		}

		err = kTempl.Execute(versionFile, map[string]string{
			"Kind": kind,
		})
		if err != nil {
			return err
		}

		for _, change := range changes {
			_, err = versionFile.WriteString("\n")
			if err != nil {
				return err
			}

			err = cTempl.Execute(versionFile, change)
			if err != nil {
				return err
			}
		}
	}

	// merge all version files + header file in changes/ into changelog.md
	/*
		changelogFile, err := fs.Create(config.ChangelogPath)
		if err != nil {
			return err
		}
		defer changelogFile.Close()
	*/
	return nil
}
