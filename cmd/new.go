package cmd

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/miniscruff/changie/project"
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runNew(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	config, err := project.LoadConfig(fs)
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
		if custom.Type == project.CustomString {
			stringPrompt := promptui.Prompt{
				Label: name,
			}
			customs[name], err = stringPrompt.Run()
			if err != nil {
				return err
			}
		} else if custom.Type == project.CustomInt {
			intPrompt := promptui.Prompt{
				Label: name,
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
		} else if custom.Type == project.CustomEnum {
			enumPrompt := promptui.Select{
				Label: name,
				Items: custom.EnumOptions,
			}
			_, customs[name], err = enumPrompt.Run()
			if err != nil {
				return err
			}
		}
	}

	change := project.Change{
		Kind:   kind,
		Body:   body,
		Time:   time.Now(),
		Custom: customs,
	}

	return change.SaveUnreleased(fs, config)
}
