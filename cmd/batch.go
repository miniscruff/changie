package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

// batchData organizes all the prerequisite data we need to batch a version
type batchData struct {
	Version *semver.Version
	Changes []core.Change
	Header  string
}

var (
	versionHeaderPath = ""
	keepFragments     = false
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
	batchCmd.Flags().BoolVarP(
		&keepFragments,
		"keep", "k",
		false,
		"Keep change fragments instead of deleting them",
	)
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return batchPipeline(afs, args[0])
}

func batchPipeline(afs afero.Afero, version string) error {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	ver, err := core.GetNextVersion(afs.ReadDir, config, version)
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

	if !keepFragments {
		err = deleteUnreleased(afs, config, headerFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func getChanges(afs afero.Afero, config core.Config) ([]core.Change, error) {
	var changes []core.Change

	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)

	// read all markdown files from changes/unreleased
	fileInfos, err := afs.ReadDir(unreleasedPath)
	if err != nil {
		return []core.Change{}, err
	}

	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(config.ChangesDir, config.UnreleasedDir, file.Name())

		c, err := core.LoadChange(path, afs.ReadFile)
		if err != nil {
			return changes, err
		}

		changes = append(changes, c)
	}

	core.SortByConfig(config).Sort(changes)

	return changes, nil
}

func headerForKind(config core.Config, label string) string {
	for _, kindConfig := range config.Kinds {
		if kindConfig.Header != "" && kindConfig.Label == label {
			return kindConfig.Header
		}
	}

	return config.KindFormat
}

func changeFormatFromKind(config core.Config, label string) string {
	for _, kindConfig := range config.Kinds {
		if kindConfig.ChangeFormat != "" && kindConfig.Label == label {
			return kindConfig.ChangeFormat
		}
	}

	return config.ChangeFormat
}

func batchNewVersion(fs afero.Fs, config core.Config, data batchData) error {
	templateCache := core.NewTemplateCache()
	versionPath := filepath.Join(config.ChangesDir, data.Version.Original()+"."+config.VersionExt)

	versionFile, err := fs.Create(versionPath)
	if err != nil {
		return err
	}
	defer versionFile.Close()

	err = templateCache.Execute(config.VersionFormat, versionFile, map[string]interface{}{
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

			err = templateCache.Execute(config.ComponentFormat, versionFile, map[string]string{
				"Component": change.Component,
			})
			if err != nil {
				return err
			}
		}

		if config.KindFormat != "" && lastKind != change.Kind {
			lastKind = change.Kind
			kindHeader := headerForKind(config, change.Kind)

			_, _ = versionFile.WriteString("\n")

			err = templateCache.Execute(kindHeader, versionFile, map[string]string{
				"Kind": change.Kind,
			})
			if err != nil {
				return err
			}
		}

		_, _ = versionFile.WriteString("\n")
		changeFormat := changeFormatFromKind(config, lastKind)
		err = templateCache.Execute(changeFormat, versionFile, change)

		if err != nil {
			return err
		}
	}

	return nil
}

func deleteUnreleased(afs afero.Afero, config core.Config, headerPath string) error {
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
