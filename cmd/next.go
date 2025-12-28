package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

type Next struct {
	*cobra.Command

	// cli args
	IncludeDirs []string
	Prerelease  []string
	Meta        []string
	Project     string
}

func NewNext() *Next {
	next := &Next{}

	cmd := &cobra.Command{
		Use:   "next major|minor|patch|auto",
		Short: "Next echos the next version based on semantic versioning",
		Long: `Next increments version based on semantic versioning.
Check latest version and increment part (major, minor, patch).
If auto is used, it will try and find the next version based on what kinds of changes are
currently unreleased.
Echo the next release version number to be used by CI tools or other commands like batch.`,
		ValidArgs: []string{"major", "minor", "patch", "auto"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE:      next.Run,
	}

	cmd.Flags().StringSliceVarP(
		&next.IncludeDirs,
		"include", "i",
		nil,
		"Include extra directories to search for change files, relative to change directory",
	)
	cmd.Flags().StringSliceVarP(
		&next.Prerelease,
		"prerelease", "p",
		nil,
		"Prerelease values to append to version",
	)
	cmd.Flags().StringSliceVarP(
		&next.Meta,
		"metadata", "m",
		nil,
		"Metadata values to append to version",
	)
	cmd.Flags().StringVarP(
		&next.Project,
		"project", "j",
		"",
		"Specify which project we are interested in",
	)

	next.Command = cmd

	return next
}

func (n *Next) Run(cmd *cobra.Command, args []string) error {
	writer := cmd.OutOrStdout()
	part := strings.ToLower(args[0])
	projPrefix := ""

	cfg, err := core.LoadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Projects) > 0 {
		var pc *core.ProjectConfig

		pc, err = cfg.Project(n.Project)
		if err != nil {
			return err
		}

		n.Project = pc.Key
		projPrefix = pc.Key + cfg.ProjectsVersionSeparator
	}

	var changes []core.Change
	// only worry about loading changes, if we are in auto mode
	if part == core.AutoLevel {
		changes, err = cfg.Changes(n.IncludeDirs, n.Project)
		if err != nil {
			return err
		}
	}

	next, err := cfg.NextVersion(part, n.Prerelease, n.Meta, changes, n.Project)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(projPrefix + next.Original()))

	return err
}
