package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

type New struct {
	*cobra.Command

	// cli args
	DryRun     bool
	Projects   []string
	Component  string
	Kind       string
	Body       string
	BodyEditor bool
	Custom     []string

	// dependencies
	ReadFile      shared.ReadFiler
	CreateFile    shared.CreateFiler
	TimeNow       shared.TimeNow
	MkdirAll      shared.MkdirAller
	TemplateCache *core.TemplateCache
}

func NewNew(
	readFile shared.ReadFiler,
	createFile shared.CreateFiler,
	timeNow shared.TimeNow,
	mkdirAll shared.MkdirAller,
	templateCache *core.TemplateCache,
) *New {
	n := &New{
		ReadFile:      readFile,
		CreateFile:    createFile,
		TimeNow:       timeNow,
		MkdirAll:      mkdirAll,
		TemplateCache: templateCache,
	}

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new change file",
		Long: `Creates a new change file.
Change files are processed when batching a new release.
Each version is merged together for the overall project changelog.`,
		Args: cobra.NoArgs,
		RunE: n.Run,
	}

	cmd.Flags().BoolVarP(
		&n.DryRun,
		"dry-run", "d",
		false,
		"Print new fragment instead of writing to disk",
	)
	cmd.Flags().StringSliceVarP(
		&n.Projects,
		"projects", "j",
		[]string{},
		"(Preview) Set the change projects without a prompt",
	)
	cmd.Flags().StringVarP(
		&n.Component,
		"component", "c",
		"",
		"Set the change component without a prompt",
	)
	cmd.Flags().StringVarP(
		&n.Kind,
		"kind", "k",
		"",
		"Set the change kind without a prompt",
	)
	cmd.Flags().StringVarP(
		&n.Body,
		"body", "b",
		"",
		"Set the change body without a prompt",
	)
	cmd.Flags().BoolVarP(
		&n.BodyEditor,
		"editor", "e",
		false,
		"Edit body message using your text editor defined by 'EDITOR' env variable",
	)
	cmd.Flags().StringSliceVarP(
		&n.Custom,
		"custom", "m",
		nil,
		"Set custom values without a prompt",
	)

	n.Command = cmd

	return n
}

func (n *New) Run(cmd *cobra.Command, args []string) error {
	config, err := core.LoadConfig(n.ReadFile)
	if err != nil {
		return err
	}

	customValues, err := core.CustomMapFromStrings(n.Custom)
	if err != nil {
		return err
	}

	prompts := &core.Prompts{
		StdinReader:      n.InOrStdin(),
		BodyEditor:       n.BodyEditor,
		Projects:         n.Projects,
		Component:        n.Component,
		Kind:             n.Kind,
		Body:             n.Body,
		TimeNow:          n.TimeNow,
		Config:           config,
		Customs:          customValues,
		CreateFiler:      os.Create,
		EditorCmdBuilder: core.BuildCommand,
	}

	changes, err := prompts.BuildChanges()
	if err != nil {
		return err
	}

	for _, change := range changes {
		var writer io.Writer

		if n.DryRun {
			writer = n.OutOrStdout()
		} else {
			var fragmentWriter strings.Builder
			fileErr := n.TemplateCache.Execute(config.FragmentFileFormat, &fragmentWriter, change)
			if fileErr != nil {
				return fileErr
			}
			fragmentWriter.WriteString(".yaml")
			outputFilename := fragmentWriter.String()

			// Sanatize the filename to remove invalid characters such as slashes
			replacer := strings.NewReplacer("/", "-", "\\", "-")
			outputFilename = replacer.Replace(outputFilename)

			outputPath := filepath.Join(config.ChangesDir, config.UnreleasedDir, outputFilename)
			fileErr = n.MkdirAll(filepath.Dir(outputPath), core.CreateDirMode)
			if fileErr != nil {
				return fileErr
			}

			newFile, fileErr := n.CreateFile(outputPath)
			if fileErr != nil {
				return fileErr
			}

			defer newFile.Close()

			writer = newFile
		}

		_, err = change.WriteTo(writer)
	}

	return err
}
