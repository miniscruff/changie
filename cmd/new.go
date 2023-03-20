package cmd

import (
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

type New struct {
	*cobra.Command

	// cli args
	DryRun    bool
	Component string
	Kind      string
	Body      string
	Custom    []string

	// dependencies
	ReadFile      shared.ReadFiler
	CreateFile    shared.CreateFilerOS
	TimeNow       shared.TimeNow
	TemplateCache *core.TemplateCache
}

func NewNew(
	readFile shared.ReadFiler,
	createFile shared.CreateFilerOS,
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
		RunE: n.Run,
	}

	cmd.Flags().BoolVarP(
		&n.DryRun,
		"dry-run", "d",
		false,
		"Print new fragment instead of writing to disk",
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
		Component: n.Component,
		Kind:      n.Kind,
		Body:      n.Body,
		Custom:    customValues,
	}

	err = change.AskPrompts(config, n.InOrStdin())
	if err != nil {
		return err
	}

	change.Time = n.TimeNow()

	var writer io.Writer

	if n.DryRun {
		writer = n.OutOrStdout()
	} else {
		var outputPath strings.Builder

		outputPath.WriteString(config.ChangesDir + "/" + config.UnreleasedDir + "/")

		err = n.TemplateCache.Execute(config.FragmentFileFormat, &outputPath, change)
		if err != nil {
			return err
		}

		outputPath.WriteString(".yaml")

		newFile, err := n.CreateFile(outputPath.String())
		if err != nil {
			return err
		}
		defer newFile.Close()

		writer = newFile
	}

	return change.Write(writer)
}
