package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

type Latest struct {
	*cobra.Command

	// CLI args
	RemovePrefix    bool
	SkipPrereleases bool

	// dependencies
	ReadFile shared.ReadFiler
	ReadDir  shared.ReadDirer
}

func NewLatest(readFile shared.ReadFiler, readDir shared.ReadDirer) *Latest {
	l := &Latest{
		ReadFile: readFile,
		ReadDir:  readDir,
	}

	cmd := &cobra.Command{
		Use:   "latest",
		Short: "Echos the latest release version number",
		Long:  `Echo the latest release version number to be used by CI tools.`,
		Args:  cobra.NoArgs,
		RunE:  l.Run,
	}
	cmd.Flags().BoolVarP(
		&l.RemovePrefix,
		"remove-prefix", "r",
		false,
		"Remove 'v' prefix before echoing",
	)
	cmd.Flags().BoolVarP(
		&l.SkipPrereleases,
		"skip-prereleases", "",
		false,
		"Excludes prereleases to determine the latest version.",
	)

	l.Command = cmd

	return l
}

func (l *Latest) Run(cmd *cobra.Command, args []string) error {
	config, err := core.LoadConfig(l.ReadFile)
	if err != nil {
		return err
	}

	ver, err := core.GetLatestVersion(l.ReadDir, config, l.SkipPrereleases, "")
	if err != nil {
		return err
	}

	latestVer := ver.Original()
	if l.RemovePrefix {
		latestVer = strings.TrimPrefix(latestVer, "v")
	}

	_, err = l.OutOrStdout().Write([]byte(latestVer))

	return err
}
