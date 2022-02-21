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

type BatchPipeliner interface {
	// afs afero.Afero should be part of struct
	GetChanges(config core.Config) ([]core.Change, error)
	WriteTemplate(writer io.Writer, template string, pregap bool, templateData interface{}) error
	WriteFile(
		writer io.Writer,
		config core.Config,
		relativePath string,
		templateData interface{},
	) error
	WriteChanges(writer io.Writer, config core.Config, changes []core.Change) error
	ClearUnreleased(config core.Config, moveDir string, otherFiles ...string) error
}

type standardBatchPipeline struct {
	afs           afero.Afero
	templateCache *core.TemplateCache
}

var (
	batchDryRunOut io.Writer = os.Stdout
	// deprecated but still supported until 2.0
	oldHeaderPathFlag              = ""
	versionHeaderPathFlag          = ""
	versionFooterPathFlag          = ""
	keepFragmentsFlag              = false
	moveDirFlag                    = ""
	batchDryRunFlag                = false
	batchPrereleaseFlag   []string = nil
	batchMetaFlag         []string = nil
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
		&oldHeaderPathFlag,
		"headerPath", "",
		"Path to version header file in unreleased directory",
	)

	_ = batchCmd.Flags().MarkDeprecated("headerPath", "use --header-path instead")

	batchCmd.Flags().StringVar(
		&versionFooterPathFlag,
		"footer-path", "",
		"Path to version footer file in unreleased directory",
	)

	batchCmd.Flags().StringVar(
		&moveDirFlag,
		"move-dir", "",
		"Path to move unreleased changes",
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
	batchCmd.Flags().StringSliceVarP(
		&batchPrereleaseFlag,
		"prerelease", "p",
		nil,
		"Prerelease values to append to version",
	)
	batchCmd.Flags().StringSliceVarP(
		&batchMetaFlag,
		"metadata", "m",
		nil,
		"Metadata values to append to version",
	)
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return batchPipeline(
		&standardBatchPipeline{
			afs:           afs,
			templateCache: core.NewTemplateCache(),
		},
		afs,
		args[0],
	)
}

func getBatchData(
	config core.Config,
	afs afero.Afero,
	version string,
	batcher BatchPipeliner,
) (*core.BatchData, error) {
	previousVersion, err := core.GetLatestVersion(afs.ReadDir, config)
	if err != nil {
		return nil, err
	}

	currentVersion, err := core.GetNextVersion(
		afs.ReadDir,
		config,
		version,
		batchPrereleaseFlag,
		batchMetaFlag,
	)
	if err != nil {
		return nil, err
	}

	allChanges, err := batcher.GetChanges(config)
	if err != nil {
		return nil, err
	}

	return &core.BatchData{
		Time:            time.Now(),
		Version:         currentVersion.Original(),
		PreviousVersion: previousVersion.Original(),
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
		versionPath := filepath.Join(config.ChangesDir, data.Version+"."+config.VersionExt)

		versionFile, createErr := afs.Fs.Create(versionPath)
		if createErr != nil {
			return createErr
		}

		defer versionFile.Close()
		writer = versionFile
	}

	err = batcher.WriteTemplate(writer, config.VersionFormat, false, data)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{
		versionHeaderPathFlag,
		oldHeaderPathFlag,
		config.VersionHeaderPath,
	} {
		err = batcher.WriteFile(writer, config, relativePath, data)
		if err != nil {
			return err
		}
	}

	err = batcher.WriteTemplate(writer, config.HeaderFormat, true, data)
	if err != nil {
		return err
	}

	err = batcher.WriteChanges(writer, config, data.Changes)
	if err != nil {
		return err
	}

	err = batcher.WriteTemplate(writer, config.FooterFormat, true, data)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{versionFooterPathFlag, config.VersionFooterPath} {
		err = batcher.WriteFile(writer, config, relativePath, data)
		if err != nil {
			return err
		}
	}

	if !batchDryRunFlag && !keepFragmentsFlag {
		err = batcher.ClearUnreleased(
			config,
			moveDirFlag,
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

func (b *standardBatchPipeline) WriteTemplate(
	writer io.Writer,
	template string,
	pregap bool,
	templateData interface{},
) error {
	if template == "" {
		return nil
	}

	if pregap {
		_, err := writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	return b.templateCache.Execute(template, writer, templateData)
}

func (b *standardBatchPipeline) WriteFile(
	writer io.Writer,
	config core.Config,
	relativePath string,
	templateData interface{},
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

	return b.WriteTemplate(writer, string(fileBytes), true, templateData)
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

			err := b.WriteTemplate(writer, config.ComponentFormat, true, map[string]string{
				"Component": lastComponent,
			})
			if err != nil {
				return err
			}
		}

		if config.KindFormat != "" && lastKind != change.Kind {
			lastKind = change.Kind
			kindHeader := config.KindHeader(change.Kind)

			err := b.WriteTemplate(writer, kindHeader, true, map[string]string{
				"Kind": lastKind,
			})
			if err != nil {
				return err
			}
		}

		changeFormat := config.ChangeFormatForKind(lastKind)

		err := b.WriteTemplate(writer, changeFormat, true, change)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *standardBatchPipeline) ClearUnreleased(
	config core.Config,
	moveDir string,
	otherFiles ...string,
) error {
	var (
		filesToMove []string
		err         error
	)

	if moveDir != "" {
		err = b.afs.MkdirAll(filepath.Join(config.ChangesDir, moveDir), core.CreateDirMode)
		if err != nil {
			return err
		}
	}

	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)

	for _, p := range otherFiles {
		if p == "" {
			continue
		}

		fullPath := filepath.Join(unreleasedPath, p)

		// make sure the file exists first
		_, err = b.afs.Stat(fullPath)
		if err == nil {
			filesToMove = append(filesToMove, p)
		}
	}

	fileInfos, _ := b.afs.ReadDir(unreleasedPath)
	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		filesToMove = append(filesToMove, file.Name())
	}

	for _, f := range filesToMove {
		if moveDir != "" {
			err = b.afs.Rename(
				filepath.Join(unreleasedPath, f),
				filepath.Join(config.ChangesDir, moveDir, f),
			)
			if err != nil {
				return err
			}
		} else {
			fullPath := filepath.Join(unreleasedPath, f)
			err = b.afs.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
