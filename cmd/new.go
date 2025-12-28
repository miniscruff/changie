package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

type New struct {
	*cobra.Command

	// cli args
	DryRun      bool
	Projects    []string
	Component   string
	Kind        string
	Body        string
	BodyEditor  bool
	Custom      []string
	Interactive bool

	// dependencies
	TimeNow       core.TimeNow
	TemplateCache *core.TemplateCache
}

func NewNew(
	timeNow core.TimeNow,
	templateCache *core.TemplateCache,
) *New {
	n := &New{
		TimeNow:       timeNow,
		TemplateCache: templateCache,
	}

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new change file",
		Long: `Creates a new change file.
Change files are processed when batching a new release.
Each version is merged together for the overall project changelog.

Prompts are disabled and this command will fail if any values
are not defined or valid and if any of the following are true:

Custom prompt values can also be passed in via an environment variable.
Use the following format: "${env var prefix}_CUSTOM_${custom key}=value".
Example: "CHANGIE_CUSTOM_Author=miniscruff"

1. CI env var is true
2. --interactive=false
`,
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
		"Set the change projects without a prompt",
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
	cmd.Flags().BoolVarP(
		&n.Interactive,
		"interactive",
		"i",
		true,
		"Set missing values with prompts",
	)

	n.Command = cmd

	return n
}

func (n *New) Run(cmd *cobra.Command, args []string) error {
	config, err := core.LoadConfig()
	if err != nil {
		return err
	}

	customValues, err := core.CustomMapFromStrings(n.Custom)
	if err != nil {
		return err
	}

	changieEnvs := config.EnvVars()
	for key, value := range changieEnvs {
		key, found := strings.CutPrefix(key, "CUSTOM_")
		if found {
			customValues[key] = value
		}
	}

	prompts := &core.Prompts{
		StdinReader:      n.InOrStdin(),
		BodyEditor:       n.BodyEditor,
		Projects:         n.Projects,
		Component:        n.Component,
		Kind:             n.Kind,
		Body:             n.Body,
		TimeNow:          n.TimeNow,
		Cfg:           config,
		Customs:          customValues,
		EditorCmdBuilder: core.BuildCommand,
		Enabled:          n.parsePromptEnabled(),
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

			fileErr = os.MkdirAll(filepath.Dir(outputPath), core.CreateDirMode)
			if fileErr != nil {
				return fileErr
			}

			newFile, fileErr := os.Create(outputPath)
			if fileErr != nil {
				return fileErr
			}

			defer newFile.Close()

			writer = newFile
		}

		_, err = change.WriteTo(writer)
		if err != nil {
			return err
		}
	}

	return err
}

func (n *New) parsePromptEnabled() bool {
	return n.Interactive && strings.ToLower(os.Getenv("CI")) != "true"
}
