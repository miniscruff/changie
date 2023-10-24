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
	Project    string
	Component  string
	Kind       string
	Body       string
	BodyEditor bool
	Custom     []string

	// dependencies
	ReadFile      shared.ReadFiler
	CreateFile    shared.CreateFiler
	TimeNow       shared.TimeNow
	TemplateCache *core.TemplateCache
}

func NewNew(
	readFile shared.ReadFiler,
	createFile shared.CreateFiler,
	timeNow shared.TimeNow,
	templateCache *core.TemplateCache,
) *New {
	n := &New{
		ReadFile:      readFile,
		CreateFile:    createFile,
		TimeNow:       timeNow,
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
	cmd.Flags().StringVarP(
		&n.Project,
		"project", "j",
		"",
		"(Preview) Set the change project key without a prompt",
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

	change := core.Change{
		Project:   n.Project,
		Component: n.Component,
		Kind:      n.Kind,
		Body:      n.Body,
		Custom:    customValues,
		Env:       config.EnvVars(),
	}

	err = change.AskPrompts(core.PromptContext{
		Config:           config,
		StdinReader:      n.InOrStdin(),
		BodyEditor:       n.BodyEditor,
		CreateFiler:      os.Create,
		EditorCmdBuilder: core.BuildCommand,
	})
	if err != nil {
		return err
	}

	change.Time = n.TimeNow()

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
		newFile, fileErr := n.CreateFile(outputPath)
		if fileErr != nil {
			return fileErr
		}

		defer newFile.Close()

		writer = newFile
	}

	_, err = change.WriteTo(writer)

	return err
}
