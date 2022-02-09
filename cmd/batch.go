package cmd

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

var (
	versionHeaderPathFlag = ""
	keepFragmentsFlag     = false
	batchDryRunFlag       = false
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
	batchCmd.Flags().StringVar(
		&versionHeaderPathFlag,
		"headerPath", "",
		"Path to version header file in unreleased directory",
	)
	batchCmd.Flags().BoolVarP(
		&keepFragmentsFlag,
		"keep", "k",
		false,
		"Keep change fragments instead of deleting them",
	)
	batchCmd.Flags().BoolVarP(
		&batchDryRunFlag,
		"dry-run", "d",
		false,
		"Print batched changes instead of writing to disk, does not delete fragments",
	)
	rootCmd.AddCommand(batchCmd)
}

/*
	3. Write header template
	5. Write footer file
	6. Write footer template
*/

func runBatch(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return batchPipeline(afs, args[0])
}

func batchPipeline(afs afero.Afero, version string) error {
	templateCache := core.NewTemplateCache()

	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	previousVersion, err := core.GetLatestVersion(afs.ReadDir, config)
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

	var writer io.Writer
	if batchDryRunFlag {
		writer = os.Stdout
	} else {
		versionPath := filepath.Join(config.ChangesDir, ver.Original()+"."+config.VersionExt)
		versionFile, err := afs.Fs.Create(versionPath)

		if err != nil {
			return err
		}
		defer versionFile.Close()
		writer = versionFile
	}

	// start writing our version
	err = templateCache.Execute(config.VersionFormat, writer, map[string]interface{}{
		"Version":         ver.Original(),
		"PreviousVersion": previousVersion.Original(),
		"Time":            time.Now(),
	})
	if err != nil {
		return err
	}

	err = writeUnreleasedFile(writer, afs, config, versionHeaderPathFlag)
	if err != nil {
		return err
	}

	err = writeUnreleasedFile(writer, afs, config, config.VersionHeaderPath)
	if err != nil {
		return err
	}

	err = writeChanges(writer, config, templateCache, allChanges)
	if err != nil {
		return err
	}

	if !batchDryRunFlag && !keepFragmentsFlag {
		err = deleteUnreleased(
			afs,
			config,
			versionHeaderPathFlag,
			config.VersionHeaderPath,
			// footer paths
		)
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

		path := filepath.Join(unreleasedPath, file.Name())

		c, err := core.LoadChange(path, afs.ReadFile)
		if err != nil {
			return changes, err
		}

		changes = append(changes, c)
	}

	core.SortByConfig(config).Sort(changes)

	return changes, nil
}

func writeUnreleasedFile(writer io.Writer, afs afero.Afero, config core.Config, relativePath string) error {
	if relativePath == "" {
		return nil
	}

	fullPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, relativePath)
	headerBytes, readErr := afs.ReadFile(fullPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}

	if os.IsNotExist(readErr) {
		return nil
	}

	_, err := writer.Write(headerBytes)
	return err
}

func writeChanges(writer io.Writer, config core.Config, templateCache *core.TemplateCache, changes []core.Change) error {
	lastComponent := ""
	lastKind := ""

	for _, change := range changes {
		if config.ComponentFormat != "" && lastComponent != change.Component {
			lastComponent = change.Component
			lastKind = ""
			_, _ = writer.Write([]byte("\n"))

			err := templateCache.Execute(config.ComponentFormat, writer, map[string]string{
				"Component": change.Component,
			})
			if err != nil {
				return err
			}
		}

		if config.KindFormat != "" && lastKind != change.Kind {
			lastKind = change.Kind
			kindHeader := headerForKind(config, change.Kind)

			_, _ = writer.Write([]byte("\n"))

			err := templateCache.Execute(kindHeader, writer, map[string]string{
				"Kind": change.Kind,
			})
			if err != nil {
				return err
			}
		}

		_, _ = writer.Write([]byte("\n"))
		changeFormat := changeFormatFromKind(config, lastKind)

		err := templateCache.Execute(changeFormat, writer, change)
		if err != nil {
			return err
		}
	}

	return nil
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

func deleteUnreleased(afs afero.Afero, config core.Config, otherFiles ...string) error {
	var filesToDelete []string
	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)

	for _, p := range otherFiles {
		if p == "" {
			continue
		}
		fullPath := filepath.Join(unreleasedPath, p)

		// make sure the file exists first
		_, err := afs.Stat(fullPath)
		if err == nil {
			filesToDelete = append(filesToDelete, filepath.Join(unreleasedPath, p))
		}
	}

	fileInfos, _ := afs.ReadDir(unreleasedPath)
	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		filesToDelete = append(filesToDelete, filepath.Join(unreleasedPath, file.Name()))
	}

	for _, f := range filesToDelete {
		err := afs.Remove(f)
		if err != nil {
			return err
		}
	}

	return nil
}
