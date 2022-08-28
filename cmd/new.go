package cmd

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

type newConfig struct {
	afs           afero.Afero
	cmdOut        io.Writer
	timeNow       shared.TimeNow
	stdinReader   io.ReadCloser
	templateCache *core.TemplateCache
}

var (
	_newDryRun          = false
	_component          = ""
	_kind               = ""
	_body               = ""
	_custom    []string = nil
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
	newCmd.Flags().BoolVarP(
		&_newDryRun,
		"dry-run", "d",
		false,
		"Print new fragment instead of writing to disk",
	)
	newCmd.Flags().StringVarP(
		&_component,
		"component", "c",
		"",
		"Set the change component without a prompt",
	)
	newCmd.Flags().StringVarP(
		&_kind,
		"kind", "k",
		"",
		"Set the change kind without a prompt",
	)
	newCmd.Flags().StringVarP(
		&_body,
		"body", "b",
		"",
		"Set the change body without a prompt",
	)
	newCmd.Flags().StringSliceVarP(
		&_custom,
		"custom", "m",
		nil,
		"Set custom values without a prompt",
	)
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return newPipeline(newConfig{
		afs:           afs,
		timeNow:       time.Now,
		stdinReader:   os.Stdin,
		templateCache: core.NewTemplateCache(),
		cmdOut:        cmd.OutOrStdout(),
	})
}

func newPipeline(newConfig newConfig) error {
	config, err := core.LoadConfig(newConfig.afs.ReadFile)
	if err != nil {
		return err
	}

	customValues, err := core.CustomMapFromStrings(_custom)
	if err != nil {
		return err
	}

	change := core.Change{
		Component: _component,
		Kind:      _kind,
		Body:      _body,
		Custom:    customValues,
	}

	err = change.AskPrompts(config, newConfig.stdinReader)
	if err != nil {
		return err
	}

	change.Time = newConfig.timeNow()

	var writer io.Writer

	if _newDryRun {
		writer = newConfig.cmdOut
	} else {
		var outputPath strings.Builder

		outputPath.WriteString(config.ChangesDir + "/" + config.UnreleasedDir + "/")

		err = newConfig.templateCache.Execute(config.FragmentFileFormat, &outputPath, change)
		if err != nil {
			return err
		}

		outputPath.WriteString(".yaml")

		newFile, err := newConfig.afs.Fs.Create(outputPath.String())
		if err != nil {
			return err
		}
		defer newFile.Close()

		writer = newFile
	}

	return change.Write(writer)
}
