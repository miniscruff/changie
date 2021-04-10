package cmd

import (
	"errors"
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
	Version        string
	ChangesPerKind map[string][]Change
	Header         string
}

var (
	errNotSemanticVersion = errors.New("version string is not a valid semantic version")
	headerPath            = ""
)

var batchCmd = &cobra.Command{
	Use:   "batch version",
	Short: "Batch unreleased changes into a single changelog",
	Long: `Merges all unreleased changes into one version changelog.

The new version changelog can then be modified with extra descriptions,
context or with custom tweaks before merging into the main file.`,
	Args: cobra.ExactArgs(1),
	RunE: runBatch,
}

func init() {
	batchCmd.Flags().StringVar(&headerPath, "headerPath", "", "Path to version header file")
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return batchPipeline(afs, args[0])
}

func batchPipeline(afs afero.Afero, ver string) error {
	_, err := semver.NewVersion(ver)
	if err != nil {
		return errNotSemanticVersion
	}

	config, err := LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	changesPerKind, err := getChanges(afs, config)
	if err != nil {
		return err
	}

	headerFilePath := ""
	if headerPath != "" {
		headerFilePath = headerPath
	} else if config.VersionHeaderPath != "" {
		headerFilePath = config.VersionHeaderPath
	}

	headerContents := ""
	if headerFilePath != "" {
		fullPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, headerFilePath)
		headerBytes, err := afs.ReadFile(fullPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		headerContents = string(headerBytes)
	}

	err = batchNewVersion(afs.Fs, config, batchData{
		Version:        ver,
		ChangesPerKind: changesPerKind,
		Header:         headerContents,
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

func getChanges(afs afero.Afero, config Config) (map[string][]Change, error) {
	changesPerKind := make(map[string][]Change)
	unreleasedPath := filepath.Join(config.ChangesDir, config.UnreleasedDir)

	// read all markdown files from changes/unreleased
	fileInfos, err := afs.ReadDir(unreleasedPath)
	if err != nil {
		return changesPerKind, err
	}

	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(config.ChangesDir, config.UnreleasedDir, file.Name())

		c, err := LoadChange(path, afs.ReadFile)
		if err != nil {
			return changesPerKind, err
		}

		if _, ok := changesPerKind[c.Kind]; !ok {
			changesPerKind[c.Kind] = make([]Change, 0)
		}

		changesPerKind[c.Kind] = append(changesPerKind[c.Kind], c)
	}

	return changesPerKind, nil
}

func batchNewVersion(fs afero.Fs, config Config, data batchData) error {
	versionPath := filepath.Join(config.ChangesDir, data.Version+"."+config.VersionExt)

	versionFile, err := fs.Create(versionPath)
	if err != nil {
		return err
	}
	defer versionFile.Close()

	vTempl, err := template.New("version").Parse(config.VersionFormat)
	if err != nil {
		return err
	}

	kTempl, err := template.New("kind").Parse(config.KindFormat)
	if err != nil {
		return err
	}

	cTempl, err := template.New("change").Parse(config.ChangeFormat)
	if err != nil {
		return err
	}

	err = vTempl.Execute(versionFile, map[string]interface{}{
		"Version": data.Version,
		"Time":    time.Now(),
	})
	if err != nil {
		return err
	}

	if data.Header != "" {
		versionFile.WriteString("\n")
		versionFile.WriteString("\n")
		versionFile.WriteString(data.Header)
	}

	for _, kind := range config.Kinds {
		changes, ok := data.ChangesPerKind[kind]
		if !ok {
			continue
		}

		// just check the first write string
		_, err = versionFile.WriteString("\n")
		if err != nil {
			return err
		}

		_, _ = versionFile.WriteString("\n")

		err = kTempl.Execute(versionFile, map[string]string{
			"Kind": kind,
		})
		if err != nil {
			return err
		}

		for _, change := range changes {
			_, _ = versionFile.WriteString("\n")
			err = cTempl.Execute(versionFile, change)

			if err != nil {
				return err
			}
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
