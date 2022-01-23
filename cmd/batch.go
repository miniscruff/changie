package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
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

// batchConfig organizes all the configuration for our batch pipeline
type batchConfig struct {
	afs               afero.Afero
	versionHeaderPath string
	version           string
	keepFragments     bool
	dryRun            bool
}

var (
	// prefix cobra flags with underscore
	// use batchConfig instead
	_versionHeaderPath = ""
	_keepFragments     = false
	_batchDryRun       = false
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
	batchCmd.Flags().StringVar(&_versionHeaderPath, "headerPath", "", "Path to version header file")
	batchCmd.Flags().BoolVarP(
		&_keepFragments,
		"keep", "k",
		false,
		"Keep change fragments instead of deleting them",
	)
	batchCmd.Flags().BoolVarP(
		&_batchDryRun,
		"dry-run", "d",
		false,
		"Print batched changes instead of writing to disk, does not delete fragments",
	)
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	out, err := batchPipeline(batchConfig{
		afs:               afs,
		version:           args[0],
		versionHeaderPath: _versionHeaderPath,
		keepFragments:     _keepFragments,
		dryRun:            _batchDryRun,
	})

	if err != nil {
		return err
	}

	if out != "" {
		_, err = cmd.OutOrStdout().Write([]byte(out))
	}

	return err
}

func batchPipeline(batchConfig batchConfig) (string, error) {
	config, err := core.LoadConfig(batchConfig.afs.ReadFile)
	if err != nil {
		return "", err
	}

	ver, err := core.GetNextVersion(batchConfig.afs.ReadDir, config, batchConfig.version)
	if err != nil {
		return "", err
	}

	allChanges, err := getChanges(batchConfig.afs, config)
	if err != nil {
		return "", err
	}

	headerFilePath := ""
	headerContents := ""

	if batchConfig.versionHeaderPath != "" {
		headerFilePath = batchConfig.versionHeaderPath
	} else if config.VersionHeaderPath != "" {
		headerFilePath = config.VersionHeaderPath
	}

	if headerFilePath != "" {
		fullPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, headerFilePath)
		headerBytes, readErr := batchConfig.afs.ReadFile(fullPath)

		if readErr != nil && !os.IsNotExist(readErr) {
			return "", readErr
		}

		headerContents = string(headerBytes)
	}

	var builder strings.Builder
	err = batchNewVersion(&builder, config, batchData{
		Version: ver,
		Changes: allChanges,
		Header:  headerContents,
	})

	if err != nil {
		return "", err
	}

	// return our builder in dry run
	if batchConfig.dryRun {
		return builder.String(), nil
	}

	if !batchConfig.keepFragments {
		err = deleteUnreleased(batchConfig.afs, config, headerFilePath)
		if err != nil {
			return "", err
		}
	}

	// write to file
	versionPath := filepath.Join(config.ChangesDir, ver.Original()+"."+config.VersionExt)
	versionFile, err := batchConfig.afs.Fs.Create(versionPath)

	if err != nil {
		return "", err
	}
	defer versionFile.Close()

	_, err = versionFile.Write([]byte(builder.String()))

	return "", err
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

func batchNewVersion(writer io.Writer, config core.Config, data batchData) error {
	templateCache := core.NewTemplateCache()

	err := templateCache.Execute(config.VersionFormat, writer, map[string]interface{}{
		"Version": data.Version.Original(),
		"Time":    time.Now(),
	})
	if err != nil {
		return err
	}

	if data.Header != "" {
		// write string is not err checked as we already write to the same
		// file in the templates
		_, _ = writer.Write([]byte("\n"))
		_, _ = writer.Write([]byte(data.Header))
	}

	lastComponent := ""
	lastKind := ""

	for _, change := range data.Changes {
		if config.ComponentFormat != "" && lastComponent != change.Component {
			lastComponent = change.Component
			lastKind = ""
			_, _ = writer.Write([]byte("\n"))

			err = templateCache.Execute(config.ComponentFormat, writer, map[string]string{
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

			err = templateCache.Execute(kindHeader, writer, map[string]string{
				"Kind": change.Kind,
			})
			if err != nil {
				return err
			}
		}

		_, _ = writer.Write([]byte("\n"))
		changeFormat := changeFormatFromKind(config, lastKind)
		err = templateCache.Execute(changeFormat, writer, change)

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
