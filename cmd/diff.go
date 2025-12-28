package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
)

type Diff struct {
	*cobra.Command

	// CLI args
	SkipPrereleases bool
	Project         string
}

func NewDiff() *Diff {
	diff := &Diff{}

	cmd := &cobra.Command{
		Use:   "diff N|>,>=,<,<=version|versiona - versionb|start...end",
		Short: "diff outputs the release notes between versions.",
		Long: `diff outputs to stdout the release notes between two versions.

If the argument is a number, we simply output that many previous versions.
If the argument is two versions split between triple dots similar to how GitHub
supports as a comparison then we output the release notes between those two.

Finally one last option is any of the constraints options from the semver package:
https://github.com/Masterminds/semver#checking-version-constraints.

Between versions we also add an amount of newlines specified by the AfterChangelogVersion value.`,
		Example: `v1.20.0...v1.21.1`,
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    diff.Run,
	}
	cmd.Flags().BoolVarP(
		&diff.SkipPrereleases,
		"skip-prereleases", "",
		false,
		"Excludes prereleases to determine the diff",
	)
	cmd.Flags().StringVarP(
		&diff.Project,
		"project", "j",
		"",
		"Specify which project we are interested in",
	)

	diff.Command = cmd

	return diff
}

func (d *Diff) Run(cmd *cobra.Command, args []string) error {
	writer := cmd.OutOrStdout()
	versionRange := args[0]

	cfg, err := core.LoadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Projects) > 0 {
		var pc *core.ProjectConfig

		pc, err = cfg.Project(d.Project)
		if err != nil {
			return err
		}

		d.Project = pc.Key
	}

	vers, err := cfg.AllVersions(d.SkipPrereleases, d.Project)
	if err != nil {
		return err
	}

	firstOutput := true

	versionCount, err := strconv.ParseInt(versionRange, 10, 64)
	if err == nil {
		for i := range min(int(versionCount), len(vers)) {
			if !firstOutput {
				err := core.WriteNewlines(writer, cfg.Newlines.AfterChangelogVersion)
				if err != nil {
					return err
				}
			}

			firstOutput = false

			versionPath := filepath.Join(
				cfg.ChangesDir,
				d.Project,
				vers[i].Original()+"."+cfg.VersionExt,
			)

			err = core.AppendFile(writer, versionPath)
			if err != nil {
				return err
			}
		}

		return nil
	}

	before, after, found := strings.Cut(versionRange, "...")
	if found {
		versionRange = before + " - " + after
	}

	con, err := semver.NewConstraint(versionRange)
	if err != nil {
		return fmt.Errorf("version range: %w", err)
	}

	for _, ver := range vers {
		if !con.Check(ver) {
			continue
		}

		if !firstOutput {
			err := core.WriteNewlines(writer, cfg.Newlines.AfterChangelogVersion)
			if err != nil {
				return err
			}
		}

		firstOutput = false

		versionPath := filepath.Join(
			cfg.ChangesDir,
			d.Project,
			ver.Original()+"."+cfg.VersionExt,
		)

		err = core.AppendFile(writer, versionPath)
		if err != nil {
			return err
		}
	}

	return err
}
