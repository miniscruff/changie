package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

var errVersionExists = errors.New("version already exists")

type Batch struct {
	*cobra.Command

	// CLI args
	OldHeaderPath     string // deprecated but still supported until 2.0
	VersionHeaderPath string
	VersionFooterPath string
	KeepFragments     bool
	RemovePrereleases bool
	Project           string
	MoveDir           string
	IncludeDirs       []string
	DryRun            bool
	Prerelease        []string
	Meta              []string
	Force             bool

	// Dependencies
	ReadFile      shared.ReadFiler
	WriteFile     shared.WriteFiler
	TimeNow       core.TimeNow
	TemplateCache *core.TemplateCache

	// Computed values
	config  *core.Config // current configuration
	writer  io.Writer    // writer we are batching to
	version string       // the version we are bumping to
}

func NewBatch(
	readFile shared.ReadFiler,
	writeFile shared.WriteFiler,
	timeNow core.TimeNow,
	templateCache *core.TemplateCache,
) *Batch {
	b := &Batch{
		ReadFile:      readFile,
		WriteFile:     writeFile,
		TimeNow:       timeNow,
		TemplateCache: templateCache,
	}

	cmd := &cobra.Command{
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
		RunE: b.Run,
	}

	cmd.Flags().StringVar(
		&b.VersionHeaderPath,
		"header-path", "",
		"Path to version header file in unreleased directory",
	)
	cmd.Flags().StringVar(
		&b.OldHeaderPath,
		"headerPath", "",
		"Path to version header file in unreleased directory",
	)

	_ = cmd.Flags().MarkDeprecated("headerPath", "use --header-path instead")

	cmd.Flags().StringVar(
		&b.VersionFooterPath,
		"footer-path", "",
		"Path to version footer file in unreleased directory",
	)
	cmd.Flags().StringVar(
		&b.MoveDir,
		"move-dir", "",
		"Path to move unreleased changes",
	)
	cmd.Flags().StringSliceVarP(
		&b.IncludeDirs,
		"include", "i",
		nil,
		"Include extra directories to search for change files, relative to change directory",
	)
	cmd.Flags().BoolVarP(
		&b.KeepFragments,
		"keep", "k",
		false,
		"Keep change fragments instead of deleting them",
	)
	cmd.Flags().BoolVar(
		&b.RemovePrereleases,
		"remove-prereleases",
		false,
		"Remove existing prerelease versions",
	)
	cmd.Flags().BoolVarP(
		&b.DryRun,
		"dry-run", "d",
		false,
		"Print batched changes instead of writing to disk, does not delete fragments",
	)
	cmd.Flags().StringSliceVarP(
		&b.Prerelease,
		"prerelease", "p",
		nil,
		"Prerelease values to append to version",
	)
	cmd.Flags().StringSliceVarP(
		&b.Meta,
		"metadata", "m",
		nil,
		"Metadata values to append to version",
	)
	cmd.Flags().BoolVarP(
		&b.Force,
		"force", "f",
		false,
		"Force a new version file even if one already exists",
	)
	cmd.Flags().StringVarP(
		&b.Project,
		"project", "j",
		"",
		"Specify which project version we are batching",
	)

	b.Command = cmd

	return b
}

func (b *Batch) getBatchData() (*core.BatchData, error) {
	previousVersion, err := core.GetLatestVersion(b.config, false, b.Project)
	if err != nil {
		return nil, err
	}

	allChanges, err := core.GetChanges(b.config, b.IncludeDirs, b.ReadFile, b.Project)
	if err != nil {
		return nil, err
	}

	currentVersion, err := core.GetNextVersion(
		b.config,
		b.version,
		b.Prerelease,
		b.Meta,
		allChanges,
		b.Project,
	)
	if err != nil {
		return nil, err
	}

	return &core.BatchData{
		Time:            b.TimeNow(),
		Version:         currentVersion.Original(),
		VersionNoPrefix: currentVersion.String(),
		PreviousVersion: previousVersion.Original(),
		Major:           int(currentVersion.Major()),
		Minor:           int(currentVersion.Minor()),
		Patch:           int(currentVersion.Patch()),
		Prerelease:      currentVersion.Prerelease(),
		Metadata:        currentVersion.Metadata(),
		Changes:         allChanges,
		Env:             b.config.EnvVars(),
	}, nil
}

//nolint:gocyclo
func (b *Batch) Run(cmd *cobra.Command, args []string) error {
	var err error

	// save our version for later use
	b.version = args[0]

	b.config, err = core.LoadConfig(b.ReadFile)
	if err != nil {
		return err
	}

	if len(b.config.Projects) > 0 {
		var pc *core.ProjectConfig

		pc, err = b.config.Project(b.Project)
		if err != nil {
			return err
		}

		b.Project = pc.Key

		err = os.MkdirAll(filepath.Join(b.config.ChangesDir, b.Project), core.CreateDirMode)
		if err != nil {
			return err
		}
	}

	data, err := b.getBatchData()
	if err != nil {
		return err
	}

	if b.DryRun {
		b.writer = cmd.OutOrStdout()
	} else {
		versionPath := filepath.Join(b.config.ChangesDir, b.Project, data.Version+"."+b.config.VersionExt)

		if !b.Force {
			if exists, existErr := core.FileExists(versionPath); exists || existErr != nil {
				return fmt.Errorf("%w: %v", errVersionExists, versionPath)
			}
		}

		versionFile, createErr := os.Create(versionPath)
		if createErr != nil {
			return createErr
		}

		defer versionFile.Close()
		b.writer = versionFile
	}

	err = b.WriteTemplate(
		b.config.VersionFormat,
		b.config.Newlines.BeforeVersion,
		b.config.Newlines.AfterVersion,
		data,
	)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{
		b.VersionHeaderPath,
		b.OldHeaderPath,
		b.config.VersionHeaderPath,
	} {
		err = b.WriteTemplateFile(
			relativePath,
			b.config.Newlines.BeforeHeaderFile+1,
			b.config.Newlines.AfterHeaderFile,
			data,
		)
		if err != nil {
			return err
		}
	}

	err = b.WriteTemplate(
		b.config.HeaderFormat,
		b.config.Newlines.BeforeHeaderTemplate+1,
		b.config.Newlines.AfterHeaderTemplate,
		data,
	)
	if err != nil {
		return err
	}

	err = b.WriteChanges(data.Changes)
	if err != nil {
		return err
	}

	err = b.WriteTemplate(
		b.config.FooterFormat,
		b.config.Newlines.BeforeFooterTemplate+1,
		b.config.Newlines.AfterFooterTemplate,
		data,
	)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{b.VersionFooterPath, b.config.VersionFooterPath} {
		err = b.WriteTemplateFile(
			relativePath,
			b.config.Newlines.BeforeFooterFile+1,
			b.config.Newlines.AfterFooterFile,
			data,
		)
		if err != nil {
			return err
		}
	}

	_ = core.WriteNewlines(b.writer, b.config.Newlines.EndOfVersion)

	if !b.DryRun && !b.KeepFragments {
		err = b.ClearUnreleased(
			data.Changes,
			b.VersionHeaderPath,
			b.config.VersionHeaderPath,
			b.VersionFooterPath,
			b.config.VersionFooterPath,
		)
		if err != nil {
			return err
		}
	}

	if !b.DryRun && b.RemovePrereleases {
		// only chance we fail is already checked above
		allVers, _ := core.GetAllVersions(b.config, false, b.Project)

		for _, v := range allVers {
			if v.Prerelease() == "" {
				continue
			}

			err = os.Remove(filepath.Join(
				b.config.ChangesDir,
				b.Project,
				v.Original()+"."+b.config.VersionExt,
			))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Batch) WriteTemplate(
	template string,
	beforeNewlines int,
	afterNewlines int,
	templateData any,
) error {
	if template == "" {
		return nil
	}

	err := core.WriteNewlines(b.writer, beforeNewlines)
	if err != nil {
		return err
	}

	if err := b.TemplateCache.Execute(template, b.writer, templateData); err != nil {
		return err
	}

	_ = core.WriteNewlines(b.writer, afterNewlines)

	return nil
}

func (b *Batch) WriteTemplateFile(
	relativePath string,
	beforeNewlines int,
	afterNewlines int,
	templateData any,
) error {
	if relativePath == "" {
		return nil
	}

	fullPath := filepath.Join(b.config.ChangesDir, b.config.UnreleasedDir, relativePath)

	var fileBytes []byte

	fileBytes, readErr := b.ReadFile(fullPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}

	if os.IsNotExist(readErr) {
		return nil
	}

	return b.WriteTemplate(string(fileBytes), beforeNewlines, afterNewlines, templateData)
}

func (b *Batch) WriteChanges(changes []core.Change) error {
	lastComponent := ""
	lastKind := ""

	for _, change := range changes {
		if b.config.ComponentFormat != "" && lastComponent != change.Component {
			lastComponent = change.Component
			lastKind = ""

			err := b.WriteTemplate(
				b.config.ComponentFormat,
				b.config.Newlines.BeforeComponent+1,
				b.config.Newlines.AfterComponent,
				core.ComponentData{
					Component: lastComponent,
					Env:       b.config.EnvVars(),
				},
			)
			if err != nil {
				return err
			}
		}

		if b.config.KindFormat != "" && lastKind != change.Kind {
			lastKind = change.Kind
			kindHeader := b.config.KindHeader(change.Kind)

			err := b.WriteTemplate(
				kindHeader,
				b.config.Newlines.BeforeKind+1,
				b.config.Newlines.AfterKind,
				core.KindData{
					Kind: lastKind,
					Env:  b.config.EnvVars(),
				},
			)
			if err != nil {
				return err
			}
		}

		changeFormat := b.config.ChangeFormatForKind(lastKind)

		err := b.WriteTemplate(
			changeFormat,
			b.config.Newlines.BeforeChange+1,
			b.config.Newlines.AfterChange,
			change,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Batch) ClearUnreleased(changes []core.Change, otherFiles ...string) error {
	var (
		filesToMove []string
		err         error
	)

	if b.MoveDir != "" {
		err = os.MkdirAll(filepath.Join(b.config.ChangesDir, b.MoveDir), core.CreateDirMode)
		if err != nil {
			return err
		}
	}

	for _, p := range otherFiles {
		if p == "" {
			continue
		}

		fullPath := filepath.Join(b.config.ChangesDir, b.config.UnreleasedDir, p)

		if exists, existErr := core.FileExists(fullPath); exists && existErr == nil {
			filesToMove = append(filesToMove, fullPath)
		}
	}

	for _, ch := range changes {
		filesToMove = append(filesToMove, ch.Filename)
	}

	for _, f := range filesToMove {
		if b.MoveDir != "" {
			err = os.Rename(
				f,
				filepath.Join(b.config.ChangesDir, b.MoveDir, filepath.Base(f)),
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

	for _, include := range b.IncludeDirs {
		fullInclude := filepath.Join(b.config.ChangesDir, include)

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
