package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

var (
	errVersionExists       = errors.New("version already exists")
	errNoChangesNotAllowed = errors.New("no changes found and allow no changes disabled")
)

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
	AllowNoChanges    bool

	// Dependencies
	TimeNow       core.TimeNow
	TemplateCache *core.TemplateCache

	// Computed values
	cfg     *core.Config // current configuration
	writer  io.Writer    // writer we are batching to
	version string       // the version we are bumping to
}

func NewBatch(
	timeNow core.TimeNow,
	templateCache *core.TemplateCache,
) *Batch {
	b := &Batch{
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
	cmd.Flags().BoolVar(
		&b.AllowNoChanges,
		"allow-no-changes",
		true,
		"Allow batching no change fragments into an empty release note",
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
	previousVersion, err := b.cfg.LatestVersion(false, b.Project)
	if err != nil {
		return nil, err
	}

	allChanges, err := b.cfg.Changes(b.IncludeDirs, b.Project)
	if err != nil {
		return nil, err
	}

	if !b.AllowNoChanges && len(allChanges) == 0 {
		return nil, errNoChangesNotAllowed
	}

	currentVersion, err := b.cfg.NextVersion(
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
		Major:           int(currentVersion.Major()), //nolint:gosec
		Minor:           int(currentVersion.Minor()), //nolint:gosec
		Patch:           int(currentVersion.Patch()), //nolint:gosec
		Prerelease:      currentVersion.Prerelease(),
		Metadata:        currentVersion.Metadata(),
		Changes:         allChanges,
		Env:             b.cfg.EnvVars(),
	}, nil
}

//nolint:gocyclo
func (b *Batch) Run(cmd *cobra.Command, args []string) (err error) {
	// save our version for later use
	b.version = args[0]

	b.cfg, err = core.LoadConfig()
	if err != nil {
		return err
	}

	if len(b.cfg.Project.Options) > 0 {
		var pc *core.ProjectOptions

		pc, err = b.cfg.ProjectByName(b.Project)
		if err != nil {
			return err
		}

		b.Project = pc.Key

		err = os.MkdirAll(filepath.Join(b.cfg.RootDir, b.Project), core.CreateDirMode)
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
		var versionFileName string

		versionFileName, err = b.TemplateCache.ExecuteString(b.cfg.ReleaseNotes.Version.FilePath, data)
		if err != nil {
			return err
		}

		versionFilePath := filepath.Join(b.cfg.RootDir, b.Project, versionFileName)

		if !b.Force {
			if exists, existErr := core.FileExists(versionFilePath); exists || existErr != nil {
				return fmt.Errorf("%w: %v", errVersionExists, versionFilePath)
			}
		}

		versionFile, createErr := os.Create(versionFilePath)
		if createErr != nil {
			return fmt.Errorf("creating release notes file %q: %w", versionFilePath, createErr)
		}

		defer func() {
			if err != nil {
				removeErr := os.Remove(versionFilePath)
				if removeErr != nil {
					err = fmt.Errorf("batching error: %w, removing new file error: %w", err, removeErr)
				}
			}
		}()

		defer versionFile.Close()

		b.writer = versionFile
	}

	err = b.WriteTemplate(
		b.cfg.ReleaseNotes.Version.Format,
		b.cfg.ReleaseNotes.Version.Newlines.Before,
		b.cfg.ReleaseNotes.Version.Newlines.After,
		data,
	)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{
		b.VersionHeaderPath,
		b.OldHeaderPath,
		b.cfg.ReleaseNotes.Header.FilePath,
	} {
		err = b.WriteTemplateFile(
			relativePath,
			b.cfg.ReleaseNotes.Version.Newlines.Before+1,
			b.cfg.ReleaseNotes.Version.Newlines.After,
			data,
		)
		if err != nil {
			return err
		}
	}

	err = b.WriteTemplate(
		b.cfg.ReleaseNotes.Header.Format,
		b.cfg.ReleaseNotes.Header.Newlines.Before+1,
		b.cfg.ReleaseNotes.Header.Newlines.After,
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
		b.cfg.ReleaseNotes.Footer.Format,
		b.cfg.ReleaseNotes.Footer.Newlines.Before+1,
		b.cfg.ReleaseNotes.Footer.Newlines.After,
		data,
	)
	if err != nil {
		return err
	}

	for _, relativePath := range []string{
		b.VersionFooterPath,
		b.cfg.ReleaseNotes.Footer.FilePath,
	} {
		err = b.WriteTemplateFile(
			relativePath,
			b.cfg.ReleaseNotes.Footer.Newlines.Before+1,
			b.cfg.ReleaseNotes.Footer.Newlines.After,
			data,
		)
		if err != nil {
			return err
		}
	}

	_ = core.WriteNewlines(b.writer, b.cfg.ReleaseNotes.Version.Newlines.After)
	_ = core.WriteNewlines(b.writer, b.cfg.ReleaseNotes.Newlines.After)

	if !b.DryRun && !b.KeepFragments {
		err = b.ClearUnreleased(
			data.Changes,
			b.VersionHeaderPath,
			b.cfg.ReleaseNotes.Header.FilePath,
			b.VersionFooterPath,
			b.cfg.ReleaseNotes.Footer.FilePath,
		)
		if err != nil {
			return err
		}
	}

	if !b.DryRun && b.RemovePrereleases {
		// only chance we fail is already checked above
		allVers, _ := b.cfg.AllVersions(false, b.Project)

		for _, v := range allVers {
			if v.Prerelease() == "" {
				continue
			}

			err = os.Remove(filepath.Join(
				b.cfg.RootDir,
				b.Project,
				v.Original()+"."+b.cfg.ReleaseNotes.Extension,
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

	fullPath := filepath.Join(b.cfg.RootDir, b.cfg.Fragment.Dir, relativePath)

	var fileBytes []byte

	fileBytes, readErr := os.ReadFile(fullPath)
	if readErr != nil && !errors.Is(readErr, fs.ErrNotExist) {
		return readErr
	}

	if errors.Is(readErr, fs.ErrNotExist) {
		return nil
	}

	return b.WriteTemplate(string(fileBytes), beforeNewlines, afterNewlines, templateData)
}

func (b *Batch) WriteChanges(changes []core.Change) error {
	lastComponent := ""
	lastKind := ""

	for _, change := range changes {
		if b.cfg.Component.Prompt.Format != "" && lastComponent != change.Component {
			lastComponent = change.Component
			lastKind = ""

			err := b.WriteTemplate(
				b.cfg.Component.Prompt.Format,
				b.cfg.Component.Newlines.Before+1,
				b.cfg.Component.Newlines.After,
				core.ComponentData{
					Component: lastComponent,
					Env:       b.cfg.EnvVars(),
				},
			)
			if err != nil {
				return err
			}
		}

		if lastKind != change.Kind {
			lastKind = change.Kind
			newKind := b.cfg.KindFromKeyOrLabel(change.Kind)
			format := b.cfg.KindHeader(change.Kind)

			// TODO: new config supports per kind newlines
			// we need to pick the proper value between
			// this kinds newlines and the global newlines
			// but for now, just use globals

			err := b.WriteTemplate(
				format,
				b.cfg.Kind.Newlines.Before+1,
				b.cfg.Kind.Newlines.After,
				core.KindData{
					// TODO: should this be key?
					Kind: newKind.Prompt.Label,
					Env:  b.cfg.EnvVars(),
				},
			)
			if err != nil {
				return err
			}
		}

		changeFormat := b.cfg.ChangeFormatForKind(lastKind)

		err := b.WriteTemplate(
			changeFormat,
			b.cfg.Change.Newlines.Before+1,
			b.cfg.Change.Newlines.After,
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
		err = os.MkdirAll(filepath.Join(b.cfg.RootDir, b.MoveDir), core.CreateDirMode)
		if err != nil {
			return err
		}
	}

	for _, p := range otherFiles {
		if p == "" {
			continue
		}

		fullPath := filepath.Join(b.cfg.RootDir, b.cfg.Fragment.Dir, p)

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
				filepath.Join(b.cfg.RootDir, b.MoveDir, filepath.Base(f)),
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
		fullInclude := filepath.Join(b.cfg.RootDir, include)

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
