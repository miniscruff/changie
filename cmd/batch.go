package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

type BatchPipeliner interface {
	WriteTemplate(
		writer io.Writer,
		template string,
		beforeNewlines int,
		afterNewlines int,
		templateData interface{},
	) error
	WriteFile(
		writer io.Writer,
		config core.Config,
		beforeNewlines int,
		afterNewlines int,
		relativePath string,
		templateData interface{},
	) error
	WriteChanges(writer io.Writer, config core.Config, changes []core.Change) error
	ClearUnreleased(
		config core.Config,
		moveDir string,
		searchPaths []string,
		otherFiles ...string,
	) error
}

type standardBatchPipeline struct {
	templateCache *core.TemplateCache
}

var (
	batchDryRunOut io.Writer = os.Stdout
	// deprecated but still supported until 2.0
	oldHeaderPathFlag              = ""
	versionHeaderPathFlag          = ""
	versionFooterPathFlag          = ""
	keepFragmentsFlag              = false
	removePrereleasesFlag          = false
	moveDirFlag                    = ""
	batchIncludeDirs      []string = nil
	batchDryRunFlag                = false
	batchPrereleaseFlag   []string = nil
	batchMetaFlag         []string = nil
	batchForce                     = false
	errVersionExists               = errors.New("version already exists")
)

var batchCmd = &cobra.Command{
	Use:   "batch version|major|minor|patch|auto",
	Short: "Batch unreleased changes into a single changelog",
	Long: `Merges all unreleased changes into one version changelog.

Batch takes one argument for the next version to use, below are possible options.
* A specific semantic version value, with optional prefix
* Major, minor or patch to bump one level by one
* Auto which will automatically bump based on what changes were found

The new version changelog can then be modified with extra descriptions,
context or with custom tweaks before merging into the main file.
Line breaks are added before each formatted line except the first, if you wish to
add more line breaks include them in your format configurations.

Changes are sorted in the following order:
* Components if enabled, in order specified by config.components
* Kinds if enabled, in order specified by config.kinds
* Timestamp oldest first`,
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
	batchCmd.Flags().StringSliceVarP(
		&batchIncludeDirs,
		"include", "i",
		nil,
		"Include extra directories to search for change files, relative to change directory",
	)
	batchCmd.Flags().BoolVarP(
		&keepFragmentsFlag,
		"keep", "k",
		false,
		"Keep change fragments instead of deleting them",
	)
	batchCmd.Flags().BoolVar(
		&removePrereleasesFlag,
		"remove-prereleases",
		false,
		"Remove existing prerelease versions",
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
	batchCmd.Flags().BoolVarP(
		&batchForce,
		"force", "f",
		false,
		"Force a new version file even if one already exists",
	)
}

func runBatch(cmd *cobra.Command, args []string) error {
	return batchPipeline(
		&standardBatchPipeline{
			templateCache: core.NewTemplateCache(),
		},
		args[0],
	)
}

func getBatchData(
	config core.Config,
	version string,
	changePaths []string,
) (*core.BatchData, error) {
	previousVersion, err := core.GetLatestVersion(os.ReadDir, config, false)
	if err != nil {
		return nil, err
	}

	allChanges, err := core.GetChanges(config, changePaths, os.ReadDir, os.ReadFile)
	if err != nil {
		return nil, err
	}

	currentVersion, err := core.GetNextVersion(
		os.ReadDir,
		config,
		version,
		batchPrereleaseFlag,
		batchMetaFlag,
		allChanges,
	)
	if err != nil {
		return nil, err
	}

	return &core.BatchData{
		Time:            time.Now(),
		Version:         currentVersion.Original(),
		PreviousVersion: previousVersion.Original(),
		Major:           int(currentVersion.Major()),
		Minor:           int(currentVersion.Minor()),
		Patch:           int(currentVersion.Patch()),
		Prerelease:      currentVersion.Prerelease(),
		Metadata:        currentVersion.Metadata(),
		Changes:         allChanges,
		Env:             config.EnvVars(),
	}, nil
}

//nolint:funlen,gocyclo
func batchPipeline(batcher BatchPipeliner, version string) error {
	config, err := core.LoadConfig(os.ReadFile)
	if err != nil {
		return err
	}

	data, err := getBatchData(config, version, batchIncludeDirs)
	if err != nil {
		return err
	}

	var writer io.Writer
	if batchDryRunFlag {
		writer = batchDryRunOut
	} else {
		versionPath := filepath.Join(config.ChangesDir, data.Version+"."+config.VersionExt)

		if !batchForce {
			_, statErr := os.Stat(versionPath)
			if !errors.Is(statErr, fs.ErrNotExist) {
				return fmt.Errorf("%w: %v", errVersionExists, versionPath)
			}
		}

		versionFile, createErr := os.Create(versionPath)
		if createErr != nil {
			return createErr
		}

		defer versionFile.Close()
		writer = versionFile
	}

	err = batcher.WriteTemplate(
		writer,
		config.VersionFormat,
		config.Newlines.BeforeVersion,
		config.Newlines.AfterVersion,
		data,
	)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{
		versionHeaderPathFlag,
		oldHeaderPathFlag,
		config.VersionHeaderPath,
	} {
		err = batcher.WriteFile(
			writer,
			config,
			config.Newlines.BeforeHeaderFile+1,
			config.Newlines.AfterHeaderFile,
			relativePath,
			data,
		)
		if err != nil {
			return err
		}
	}

	err = batcher.WriteTemplate(
		writer,
		config.HeaderFormat,
		config.Newlines.BeforeHeaderTemplate+1,
		config.Newlines.AfterHeaderTemplate,
		data,
	)
	if err != nil {
		return err
	}

	err = batcher.WriteChanges(writer, config, data.Changes)
	if err != nil {
		return err
	}

	err = batcher.WriteTemplate(
		writer,
		config.FooterFormat,
		config.Newlines.BeforeFooterTemplate+1,
		config.Newlines.AfterFooterTemplate,
		data,
	)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{versionFooterPathFlag, config.VersionFooterPath} {
		err = batcher.WriteFile(
			writer,
			config,
			config.Newlines.BeforeFooterFile+1,
			config.Newlines.AfterFooterFile,
			relativePath,
			data,
		)
		if err != nil {
			return err
		}
	}

	_ = core.WriteNewlines(writer, config.Newlines.EndOfVersion)

	if !batchDryRunFlag && !keepFragmentsFlag {
		err = batcher.ClearUnreleased(
			config,
			moveDirFlag,
			batchIncludeDirs,
			versionHeaderPathFlag,
			config.VersionHeaderPath,
			versionFooterPathFlag,
			config.VersionFooterPath,
		)
		if err != nil {
			return err
		}
	}

	if !batchDryRunFlag && removePrereleasesFlag {
		// only chance we fail is already checked above
		allVers, _ := core.GetAllVersions(os.ReadDir, config, false)

		for _, v := range allVers {
			if v.Prerelease() == "" {
				continue
			}

			err = os.Remove(filepath.Join(
				config.ChangesDir,
				v.Original()+"."+config.VersionExt,
			))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/*
func (b *standardBatchPipeline) GetChanges(
	config core.Config,
	searchPaths []string,
) ([]core.Change, error) {
	var changes []core.Change

	changeFiles, err := core.FindChangeFiles(config, b.afs.ReadDir, searchPaths)
	if err != nil {
		return changes, err
	}

	for _, cf := range changeFiles {
		c, err := core.LoadChange(cf, b.afs.ReadFile)
		if err != nil {
			return changes, err
		}

		c.Env = config.EnvVars()
		changes = append(changes, c)
	}

	core.SortByConfig(config).Sort(changes)

	return changes, nil
}
*/

func (b *standardBatchPipeline) WriteTemplate(
	writer io.Writer,
	template string,
	beforeNewlines int,
	afterNewlines int,
	templateData interface{},
) error {
	if template == "" {
		return nil
	}

	err := core.WriteNewlines(writer, beforeNewlines)
	if err != nil {
		return err
	}

	if err := b.templateCache.Execute(template, writer, templateData); err != nil {
		return err
	}

	_ = core.WriteNewlines(writer, afterNewlines)

	return nil
}

func (b *standardBatchPipeline) WriteFile(
	writer io.Writer,
	config core.Config,
	beforeNewlines int,
	afterNewlines int,
	relativePath string,
	templateData interface{},
) error {
	if relativePath == "" {
		return nil
	}

	// fullPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, relativePath)

    var fileBytes []byte
	// fileBytes, readErr := b.afs.ReadFile(fullPath)
	// if readErr != nil && !os.IsNotExist(readErr) {
		// return readErr
	// }

	// if os.IsNotExist(readErr) {
		// return nil
	// }

	return b.WriteTemplate(writer, string(fileBytes), beforeNewlines, afterNewlines, templateData)
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

			err := b.WriteTemplate(
				writer,
				config.ComponentFormat,
				config.Newlines.BeforeComponent+1,
				config.Newlines.AfterComponent,
				core.ComponentData{
					Component: lastComponent,
					Env:       config.EnvVars(),
				},
			)
			if err != nil {
				return err
			}
		}

		if config.KindFormat != "" && lastKind != change.Kind {
			lastKind = change.Kind
			kindHeader := config.KindHeader(change.Kind)

			err := b.WriteTemplate(
				writer,
				kindHeader,
				config.Newlines.BeforeKind+1,
				config.Newlines.AfterKind,
				core.KindData{
					Kind: lastKind,
					Env:  config.EnvVars(),
				},
			)
			if err != nil {
				return err
			}
		}

		changeFormat := config.ChangeFormatForKind(lastKind)

		err := b.WriteTemplate(
			writer,
			changeFormat,
			config.Newlines.BeforeChange+1,
			config.Newlines.AfterChange,
			change,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *standardBatchPipeline) ClearUnreleased(
	config core.Config,
	moveDir string,
	includeDirs []string,
	otherFiles ...string,
) error {
	var (
		filesToMove []string
		err         error
	)

	if moveDir != "" {
        // TODO: Use b.MkdirAll
		err = os.MkdirAll(filepath.Join(config.ChangesDir, moveDir), core.CreateDirMode)
		if err != nil {
			return err
		}
	}

	for _, p := range otherFiles {
		if p == "" {
			continue
		}

		fullPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, p)

		// make sure the file exists first
        // TODO: use b.Stat
		_, err = os.Stat(fullPath)
		if err == nil {
			filesToMove = append(filesToMove, fullPath)
		}
	}

    // TODO: use b.ReadDir
	filePaths, err := core.FindChangeFiles(config, os.ReadDir, includeDirs)
	if err != nil {
		return err
	}

	filesToMove = append(filesToMove, filePaths...)

	for _, f := range filesToMove {
		if moveDir != "" {
            // TODO: do not use os directly
			err = os.Rename(
				f,
				filepath.Join(config.ChangesDir, moveDir, filepath.Base(f)),
			)
			if err != nil {
				return err
			}
		} else {
			err = os.Remove(f)
			if err != nil {
				return err
			}
		}
	}

	for _, include := range includeDirs {
		fullInclude := filepath.Join(config.ChangesDir, include)

        // TODO: do not use os
		files, _ := os.ReadDir(fullInclude)
		if len(files) == 0 {
			err = os.RemoveAll(fullInclude)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
