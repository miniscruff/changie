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
	afs         afero.Afero
	timeNow     shared.TimeNow
	stdinReader io.ReadCloser
	dryRun      bool
}

var (
	_newDryRun = false
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
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	out, err := newPipeline(newConfig{
		afs:         afs,
		timeNow:     time.Now,
		stdinReader: os.Stdin,
		dryRun:      _newDryRun,
	})

	if err != nil {
		return err
	}

	if out != "" {
		_, err = cmd.OutOrStdout().Write([]byte(out))
	}

	return err
}

func newPipeline(newConfig newConfig) (string, error) {
	var builder strings.Builder

	config, err := core.LoadConfig(newConfig.afs.ReadFile)
	if err != nil {
		return "", err
	}

	var change core.Change

	err = core.AskPrompts(&change, config, newConfig.stdinReader)
	if err != nil {
		return "", err
	}

	change.Time = newConfig.timeNow()
	_ = change.Write(&builder)

	if newConfig.dryRun {
		return builder.String(), nil
	}

	newFile, err := newConfig.afs.Fs.Create(change.OutputPath(config))
	if err != nil {
		return "", err
	}
	defer newFile.Close()

	_, err = newFile.Write([]byte(builder.String()))

	return "", err
}
