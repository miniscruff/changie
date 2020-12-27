package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new change file",
	Long: `Creates a new change file.
Change files are processed when preparing a new release and will create a changelog for the new version.
Each version is combined together for the overall project changelog.`,
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}
	return newPipeline(afs, os.Stdin)
}

func newPipeline(afs afero.Afero, stdinReader io.ReadCloser) error {
	config, err := LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	kindPrompt := promptui.Select{
		Label: "Kind",
		Items: config.Kinds,
		Stdin: stdinReader,
	}

	_, kind, err := kindPrompt.Run()
	if err != nil {
		return err
	}

	bodyPrompt := promptui.Prompt{
		Label: "Body",
		Stdin: stdinReader,
	}

	body, err := bodyPrompt.Run()
	if err != nil {
		return err
	}

	customs := make(map[string]string, 0)

	for name, custom := range config.CustomChoices {
		prompt, err := custom.CreatePrompt(name, stdinReader)
		if err != nil {
			return err
		}

		customs[name], err = prompt.Run()
		if err != nil {
			return err
		}
	}

	change := Change{
		Kind:   kind,
		Body:   body,
		Custom: customs,
	}

	return change.SaveUnreleased(afs.WriteFile, time.Now, config)
}
