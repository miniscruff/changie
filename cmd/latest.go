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
	Project         string

	// dependencies
	ReadFile shared.ReadFiler
}

func NewLatest(readFile shared.ReadFiler) *Latest {
	l := &Latest{
		ReadFile: readFile,
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
	cmd.Flags().StringVarP(
		&l.Project,
		"project", "j",
		"",
		"Specify which project we are interested in",
	)

	l.Command = cmd

	return l
}

func (l *Latest) Run(cmd *cobra.Command, args []string) error {
	projPrefix := ""

	config, err := core.LoadConfig(l.ReadFile)
	if err != nil {
		return err
	}

	if len(config.Projects) > 0 {
		var pc *core.ProjectConfig

		pc, err = config.Project(l.Project)
		if err != nil {
			return err
		}

		l.Project = pc.Key
		projPrefix = pc.Key + config.ProjectsVersionSeparator
	}

	ver, err := core.GetLatestVersion(config, l.SkipPrereleases, l.Project)
	if err != nil {
		return err
	}

	latestVer := ver.Original()
	if l.RemovePrefix {
		latestVer = strings.TrimPrefix(latestVer, "v")
	}

	_, err = l.OutOrStdout().Write([]byte(projPrefix + latestVer))

	return err
}
