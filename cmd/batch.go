package cmd

import (
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// batchData organizes all the prerequisite data we need to batch a version
type batchData struct {
	Version *semver.Version
	Changes []Change
	Header  string
}

var (
	versionHeaderPath = ""
)

var batchCmd = &cobra.Command{
	Use:   "batch version|major|minor|patch",
	Short: "Batch unreleased changes into a single changelog",
	Long: `Merges all unreleased changes into one version changelog.

Can use major, minor or patch as version to use the latest release and bump.

The new version changelog can then be modified with extra descriptions,
context or with custom tweaks before merging into the main file.
Line breaks are added before each formatted line except the first, if you wish to
add more line breaks include them in your format configurations.

Changes are sorted in the following order:
* Components if enabled, in order specified by config.components
* Kinds if enabled, in order specified by config.kinds
* Timestamp newest first`,
	Args: cobra.ExactArgs(1),
	RunE: runBatch,
}

func init() {
	batchCmd.Flags().StringVar(&versionHeaderPath, "headerPath", "", "Path to version header file")
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return batchPipeline(afs, args[0])
}

func batchPipeline(afs afero.Afero, version string) error {
	config, err := LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	ver, err := getNextVersion(afs.ReadDir, config, version)
	if err != nil {
		return err
	}

	allChanges, err := getChanges(afs, config)
	if err != nil {
		return err
	}

	headerFilePath := ""
	headerContents := ""

	if versionHeaderPath != "" {
		headerFilePath = versionHeaderPath
	} else if config.VersionHeaderPath != "" {
		headerFilePath = config.VersionHeaderPath
	}

	if headerFilePath != "" {
		fullPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, headerFilePath)
		headerBytes, readErr := afs.ReadFile(fullPath)

		if readErr != nil && !os.IsNotExist(readErr) {
			return readErr
		}

		headerContents = string(headerBytes)
	}

	err = batchNewVersion(afs.Fs, config, batchData{
		Version: ver,
		Changes: allChanges,
		Header:  headerContents,
	})
	if err != nil {
		return err
	}

	err = deleteUnreleased(afs, config, headerFilePath)
	if err != nil {
		return err
	}

	return nil
}

func getChanges(afs afero.Afero, config Config) ([]Change, error) {
	var changes []Change

	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)

	// read all markdown files from changes/unreleased
	fileInfos, err := afs.ReadDir(unreleasedPath)
	if err != nil {
		return []Change{}, err
	}

	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(config.ChangesDir, config.UnreleasedDir, file.Name())

		c, err := LoadChange(path, afs.ReadFile)
		if err != nil {
			return changes, err
		}

		changes = append(changes, c)
	}

	SortByConfig(config).Sort(changes)

	return changes, nil
}

func loadTemplates(config Config) (
	versionTemplate *template.Template,
	componentTemplate *template.Template,
	kindTemplate *template.Template,
	changeTemplate *template.Template,
	err error,
) {
	versionTemplate, err = template.New("version").Parse(config.VersionFormat)
	if err != nil {
		return
	}

	componentTemplate, err = template.New("component").Parse(config.ComponentFormat)
	if err != nil {
		return
	}

	kindTemplate, err = template.New("kind").Parse(config.KindFormat)
	if err != nil {
		return
	}

	changeTemplate, err = template.New("change").Parse(config.ChangeFormat)
	if err != nil {
		return
	}

	return
}

func batchNewVersion(fs afero.Fs, config Config, data batchData) error {
	versionPath := filepath.Join(config.ChangesDir, data.Version.Original()+"."+config.VersionExt)

	versionFile, err := fs.Create(versionPath)
	if err != nil {
		return err
	}
	defer versionFile.Close()

	versionTemplate, componentTemplate, kindTemplate, changeTemplate, err := loadTemplates(config)
	if err != nil {
		return err
	}

	err = versionTemplate.Execute(versionFile, map[string]interface{}{
		"Version": data.Version.Original(),
		"Time":    time.Now(),
	})
	if err != nil {
		return err
	}

	if data.Header != "" {
		// write string is not err checked as we already write to the same
		// file in the templates
		_, _ = versionFile.WriteString("\n")
		_, _ = versionFile.WriteString(data.Header)
	}

	lastComponent := ""
	lastKind := ""

	for _, change := range data.Changes {
		if config.ComponentFormat != "" && lastComponent != change.Component {
			lastComponent = change.Component
			lastKind = ""
			_, _ = versionFile.WriteString("\n")

			err = componentTemplate.Execute(versionFile, map[string]string{
				"Component": change.Component,
			})
			if err != nil {
				return err
			}
		}

		if config.KindFormat != "" && lastKind != change.Kind {
			lastKind = change.Kind
			_, _ = versionFile.WriteString("\n")

			err = kindTemplate.Execute(versionFile, map[string]string{
				"Kind": change.Kind,
			})
			if err != nil {
				return err
			}
		}

		_, _ = versionFile.WriteString("\n")
		err = changeTemplate.Execute(versionFile, change)

		if err != nil {
			return err
		}
	}

	return nil
}

func deleteUnreleased(afs afero.Afero, config Config, headerPath string) error {
	fileInfos, _ := afs.ReadDir(filepath.Join(config.ChangesDir, config.UnreleasedDir))
	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" && file.Name() != headerPath {
			continue
		}

		path := filepath.Join(config.ChangesDir, config.UnreleasedDir, file.Name())

		err := afs.Remove(path)
		if err != nil {
			return err
		}
	}

	return nil
}
