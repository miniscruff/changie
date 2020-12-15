package cmd

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/miniscruff/changie/project"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

var batchCmd = &cobra.Command{
	Use:   "batch version",
	Short: "Batch unreleased changes into a single changelog",
	Long: `Merges all unreleased changes into one version changelog.

The new version changelog can then be modified with extra descriptions,
context or with custom tweaks before merging into the main file.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runBatch,
}

func init() {
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
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
	err = batchNewVersion(fs, config, ver, changesPerKind)
	if err != nil {
		return err
	}

	err = deleteUnreleased(fs, config)
	if err != nil {
		return err
	}

	return nil
}

func batchNewVersion(fs afero.Fs, config project.Config, version string, changesPerKind map[string][]project.Change) error {
	versionFile, err := fs.Create(fmt.Sprintf("%s/%s.%s", config.ChangesDir, version, config.VersionExt))
	if err != nil {
		return err
	}
	defer versionFile.Close()

	vTempl, _ := template.New("version").Parse(config.VersionFormat)
	kTempl, _ := template.New("kind").Parse(config.KindFormat)
	cTempl, _ := template.New("change").Parse(config.ChangeFormat)

	err = vTempl.Execute(versionFile, map[string]interface{}{
		"Version": version,
		"Time":    time.Now(),
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
