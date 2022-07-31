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
	dryRun        bool
	timeNow       shared.TimeNow
	stdinReader   io.ReadCloser
	templateCache *core.TemplateCache
}

var (
	_newDryRun = false
	_kind      = ""
	_body      = ""
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
	newCmd.Flags().StringVarP(&_kind, "kind", "k", "", "Set the change kind")
	newCmd.Flags().StringVarP(&_body, "body", "b", "", "Set the change body")
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
		dryRun:        _newDryRun,
	})
}

func newPipeline(newConfig newConfig) error {
	config, err := core.LoadConfig(newConfig.afs.ReadFile)
	if err != nil {
		return err
	}

	var change core.Change

	change.Kind = _kind
	change.Body = _body

	err = change.AskPrompts(config, newConfig.stdinReader)
	if err != nil {
		return err
	}

	change.Time = newConfig.timeNow()

	var writer io.Writer

	if newConfig.dryRun {
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
