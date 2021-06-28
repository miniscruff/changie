package cmd

import (
	"io"
	"os"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new change file",
	Long: `Creates a new change file.
Change files are processed when batching a new release.
Each version is merged together for the overall project changelog.`,
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return newPipeline(afs, time.Now, os.Stdin)
}

func newPipeline(afs afero.Afero, tn shared.TimeNow, stdinReader io.ReadCloser) error {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	var change core.Change

	if len(config.Components) > 0 {
		compPrompt := promptui.Select{
			Label: "Component",
			Items: config.Components,
			Stdin: stdinReader,
		}

		_, comp, compPromptErr := compPrompt.Run()
		if compPromptErr != nil {
			return compPromptErr
		}

		change.Component = comp
	}

	if len(config.Kinds) > 0 {
		kindPrompt := promptui.Select{
			Label: "Kind",
			Items: config.Kinds,
			Stdin: stdinReader,
		}

		_, kind, kindPromptErr := kindPrompt.Run()
		if kindPromptErr != nil {
			return kindPromptErr
		}

		change.Kind = kind
	}

	bodyPrompt := promptui.Prompt{
		Label: "Body",
		Stdin: stdinReader,
	}

	body, err := bodyPrompt.Run()
	if err != nil {
		return err
	}

	change.Body = body

	customs := make(map[string]string)

	for _, custom := range config.CustomChoices {
		prompt, err := custom.CreatePrompt(stdinReader)
		if err != nil {
			return err
		}

		customs[custom.Key], err = prompt.Run()
		if err != nil {
			return err
		}
	}

	change.Custom = customs
	change.Time = tn()

	return change.SaveUnreleased(afs.WriteFile, config)
}
