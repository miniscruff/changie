package cmd

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"strconv"
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
	config, err := LoadConfig(fs)
	if err != nil {
		return err
	}

	kindPrompt := promptui.Select{
		Label: "Kind",
		Items: config.Kinds,
	}

	_, kind, err := kindPrompt.Run()
	if err != nil {
		return err
	}

	bodyPrompt := promptui.Prompt{
		Label: "Body",
	}

	body, err := bodyPrompt.Run()
	if err != nil {
		return err
	}

	customs := make(map[string]string, 0)

	for name, custom := range config.CustomChoices {
		label := name
		if custom.Label != "" {
			label = custom.Label
		}

		if custom.Type == CustomString {
			stringPrompt := promptui.Prompt{
				Label: label,
			}
			customs[name], err = stringPrompt.Run()
			if err != nil {
				return err
			}
		} else if custom.Type == CustomInt {
			intPrompt := promptui.Prompt{
				Label: label,
				Validate: func(input string) error {
					value, err := strconv.ParseInt(input, 10, 64)
					if err != nil {
						return errors.New("Invalid number")
					}
					if custom.MinInt != nil && value < *custom.MinInt {
						return errors.New(fmt.Sprintf("%v is below minimum of %v", value, custom.MinInt))
					}
					if custom.MaxInt != nil && value > *custom.MaxInt {
						return errors.New(fmt.Sprintf("%v is above maximum of %v", value, custom.MaxInt))
					}
					return nil
				},
			}
			customs[name], err = intPrompt.Run()
			if err != nil {
				return err
			}
		} else if custom.Type == CustomEnum {
			enumPrompt := promptui.Select{
				Label: label,
				Items: custom.EnumOptions,
			}
			_, customs[name], err = enumPrompt.Run()
			if err != nil {
				return err
			}
		}
	}

	change := Change{
		Kind:   kind,
		Body:   body,
		Custom: customs,
	}

	afs := afero.Afero{Fs: fs}
	return change.SaveUnreleased(afs, time.Now, config)
}
