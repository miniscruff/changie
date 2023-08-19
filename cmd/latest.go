package cmd

import (
	"errors"
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
	config, err := core.LoadConfig(l.ReadFile)
	if err != nil {
		return err
	}

	// TODO: move this to a util func
	if len(config.Projects) > 0 {
		if len(l.Project) == 0 {
			return errors.New("missing project label or key")
		}

		// make sure our passed in project is the key not the label
		for _, pc := range config.Projects {
			if l.Project == pc.Label {
				l.Project = pc.Key
			}

			if l.Project == pc.Key {
				break
			}
		}
	}

	ver, err := core.GetLatestVersion(l.ReadDir, config, l.SkipPrereleases, l.Project)
	if err != nil {
		return err
	}

	latestVer := ver.Original()
	if l.RemovePrefix {
		latestVer = strings.TrimPrefix(latestVer, "v")
	}

	projPrefix := ""
	if len(l.Project) > 0 {
		projPrefix = l.Project + config.ProjectsVersionSeparator
	}

	_, err = l.OutOrStdout().Write([]byte(projPrefix + latestVer))

	return err
}
