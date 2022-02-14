package cmd

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

type batchData struct {
	Time            time.Time
	Version         *semver.Version
	PreviousVersion *semver.Version
	Changes         []core.Change
}

type BatchPipeliner interface {
	// afs afero.Afero should be part of struct
	GetChanges(config core.Config) ([]core.Change, error)
	WriteTemplate(writer io.Writer, template string, templateData *batchData) error
	WriteFile(writer io.Writer, config core.Config, relativePath string, templateData *batchData) error
	WriteChanges(writer io.Writer, config core.Config, changes []core.Change) error
	DeleteUnreleased(config core.Config, otherFiles ...string) error
}

type standardBatchPipeline struct {
	afs           afero.Afero
	templateCache *core.TemplateCache
}

var (
	// deprecated but still supported until 2.0
	oldHeaderPathFlag               = ""
	versionHeaderPathFlag           = ""
	versionFooterPathFlag           = ""
	keepFragmentsFlag               = false
	batchDryRunFlag                 = false
	batchDryRunOut        io.Writer = os.Stdout
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
		"header-path", "",
		"Path to version header file in unreleased directory",
	)
	batchCmd.Flags().StringVar(
		&versionFooterPathFlag,
		"footer-path", "",
		"Path to version footer file in unreleased directory",
	)
	batchCmd.Flags().StringVar(
		&oldHeaderPathFlag,
		"headerPath", "",
		"Deprecated use --header-path instead",
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

func runBatch(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return batchPipeline(&standardBatchPipeline{afs: afs, templateCache: core.NewTemplateCache()}, afs, args[0])
}

func getBatchData(config core.Config, afs afero.Afero, version string, batcher BatchPipeliner) (*batchData, error) {
	previousVersion, err := core.GetLatestVersion(afs.ReadDir, config)
	if err != nil {
		return nil, err
	}

	currentVersion, err := core.GetNextVersion(afs.ReadDir, config, version)
	if err != nil {
		return nil, err
	}

	allChanges, err := batcher.GetChanges(config)
	if err != nil {
		return nil, err
	}

	return &batchData{
		Time:            time.Now(),
		Version:         currentVersion,
		PreviousVersion: previousVersion,
		Changes:         allChanges,
	}, nil
}

func batchPipeline(batcher BatchPipeliner, afs afero.Afero, version string) error {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	data, err := getBatchData(config, afs, version, batcher)
	if err != nil {
		return err
	}

	var writer io.Writer
	if batchDryRunFlag {
		writer = batchDryRunOut
	} else {
		versionPath := filepath.Join(config.ChangesDir, data.Version.Original()+"."+config.VersionExt)

		versionFile, createErr := afs.Fs.Create(versionPath)
		if createErr != nil {
			return createErr
		}

		defer versionFile.Close()
		writer = versionFile
	}

	// start writing our version
	err = batcher.WriteTemplate(writer, config.VersionFormat, data)
	if err != nil {
		return err
	}

	for _, unrelFile := range []string{
		versionHeaderPathFlag,
		oldHeaderPathFlag,
		config.VersionHeaderPath,
	} {
		err = batcher.WriteFile(writer, config, unrelFile, data)
		if err != nil {
			return err
		}
	}

	err = batcher.WriteChanges(writer, config, data.Changes)
	if err != nil {
		return err
	}

	for _, unrelFile := range []string{versionFooterPathFlag, config.VersionFooterPath} {
		err = batcher.WriteFile(writer, config, unrelFile, data)
		if err != nil {
			return err
		}
	}

	if !batchDryRunFlag && !keepFragmentsFlag {
		err = batcher.DeleteUnreleased(
			config,
			versionHeaderPathFlag,
			config.VersionHeaderPath,
			versionFooterPathFlag,
			config.VersionFooterPath,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *standardBatchPipeline) GetChanges(config core.Config) ([]core.Change, error) {
	var changes []core.Change

	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)

	// read all markdown files from changes/unreleased
	fileInfos, err := b.afs.ReadDir(unreleasedPath)
	if err != nil {
		return []core.Change{}, err
	}

	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(unreleasedPath, file.Name())

		c, err := core.LoadChange(path, b.afs.ReadFile)
		if err != nil {
			return changes, err
		}

		changes = append(changes, c)
	}

	core.SortByConfig(config).Sort(changes)

	return changes, nil
}

func (b *standardBatchPipeline) WriteTemplate(writer io.Writer, template string, templateData *batchData) error {
	return b.templateCache.Execute(template, writer, templateData)
}

func (b *standardBatchPipeline) WriteFile(
	writer io.Writer,
	config core.Config,
	relativePath string,
	templateData *batchData,
) error {
	if relativePath == "" {
		return nil
	}

	fullPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, relativePath)

	fileBytes, readErr := b.afs.ReadFile(fullPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}

	if os.IsNotExist(readErr) {
		return nil
	}

	// pre-write a newline to create a gap
	_, err := writer.Write([]byte("\n"))
	if err != nil {
		return err
	}

	return b.WriteTemplate(writer, string(fileBytes), templateData)
}

func (b *standardBatchPipeline) WriteChanges(
	writer io.Writer,
	config core.Config,
	changes []core.Change,
) error {
	lastComponent := ""
	lastKind := ""

	for _, change := range changes {
		if config.ComponentFormat != "" && lastComponent != change.Component {
			lastComponent = change.Component
			lastKind = ""
			_, _ = writer.Write([]byte("\n"))

			err := b.templateCache.Execute(config.ComponentFormat, writer, map[string]string{
				"Component": change.Component,
			})
			if err != nil {
				return err
			}
		}

		if config.KindFormat != "" && lastKind != change.Kind {
			lastKind = change.Kind
			kindHeader := config.KindHeader(change.Kind)

			_, _ = writer.Write([]byte("\n"))

			err := b.templateCache.Execute(kindHeader, writer, map[string]string{
				"Kind": change.Kind,
			})
			if err != nil {
				return err
			}
		}

		_, _ = writer.Write([]byte("\n"))
		changeFormat := config.ChangeFormatForKind(lastKind)

		err := b.templateCache.Execute(changeFormat, writer, change)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *standardBatchPipeline) DeleteUnreleased(config core.Config, otherFiles ...string) error {
	var filesToDelete []string

	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)

	for _, p := range otherFiles {
		if p == "" {
			continue
		}

		fullPath := filepath.Join(unreleasedPath, p)

		// make sure the file exists first
		_, err := b.afs.Stat(fullPath)
		if err == nil {
			filesToDelete = append(filesToDelete, filepath.Join(unreleasedPath, p))
		}
	}

	fileInfos, _ := b.afs.ReadDir(unreleasedPath)
	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		filesToDelete = append(filesToDelete, filepath.Join(unreleasedPath, file.Name()))
	}

	for _, f := range filesToDelete {
		err := b.afs.Remove(f)
		if err != nil {
			return err
		}
	}

	return nil
}
