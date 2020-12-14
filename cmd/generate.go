package cmd

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/miniscruff/changie/project"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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
	_, err := semver.NewVersion(ver)
	if err != nil {
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
	fileInfos, err := ioutil.ReadDir(filepath.Join(config.ChangesDir, config.UnreleasedDir))
	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(config.ChangesDir, config.UnreleasedDir, file.Name())
		c, err := project.LoadChange(path, afs)
		if err != nil {
			return err
		}

		if _, ok := changesPerKind[c.Kind]; !ok {
			changesPerKind[c.Kind] = make([]project.Change, 0)
		}

		changesPerKind[c.Kind] = append(changesPerKind[c.Kind], c)
	}

	// merge them into one {version}.md file in changes/
	err = mergeNewVersion(fs, config, ver, changesPerKind)
	if err != nil {
		return err
	}

	// delete all markdown files in unreleased ( keeping .gitkeep )
	// defer deleteChangeFiles()
	err = mergeChangelog(fs, config)
	if err != nil {
		return err
	}

	err = deleteUnreleased(fs, config)
	if err != nil {
		return err
	}

	return nil
}

func mergeNewVersion(fs afero.Fs, config project.Config, version string, changesPerKind map[string][]project.Change) error {
	versionFile, err := fs.Create(fmt.Sprintf("%s/%s.yaml", config.ChangesDir, version))
	if err != nil {
		return err
	}
	defer versionFile.Close()

	vTempl, _ := template.New("version").Parse(config.VersionFormat)
	kTempl, _ := template.New("kind").Parse(config.KindFormat)
	cTempl, _ := template.New("change").Parse(config.ChangeFormat)

	err = vTempl.Execute(versionFile, map[string]string{
		"Version": version,
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
	return nil
}

func mergeChangelog(fs afero.Fs, config project.Config) error {
	allVersions := make([]*semver.Version, 0)

	fileInfos, err := ioutil.ReadDir(config.ChangesDir)
	if err != nil {
		return err
	}

	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}
		versionString := strings.TrimSuffix(file.Name(), ".yaml")
		v, err := semver.NewVersion(versionString)
		if err != nil {
			continue
		}

		allVersions = append(allVersions, v)
	}
	fmt.Println(allVersions)
	sort.Sort(sort.Reverse(semver.Collection(allVersions)))

	changeFile, err := fs.Create(config.ChangelogPath)
	if err != nil {
		return err
	}
	defer changeFile.Close()

	appendFile(fs, changeFile, filepath.Join(config.ChangesDir, config.HeaderPath))

	for _, version := range allVersions {
		changeFile.WriteString("\n")
		changeFile.WriteString("\n")
		appendFile(fs, changeFile, filepath.Join(config.ChangesDir, version.Original()+".yaml"))
	}

	return nil
}

func appendFile(fs afero.Fs, rootFile afero.File, path string) error {
	otherFile, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer otherFile.Close()

	_, err = io.Copy(rootFile, otherFile)
	if err != nil {
		return err
	}

	return nil
}

func deleteUnreleased(fs afero.Fs, config project.Config) error {
	fileInfos, err := ioutil.ReadDir(filepath.Join(config.ChangesDir, config.UnreleasedDir))
	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(config.ChangesDir, config.UnreleasedDir, file.Name())
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}
