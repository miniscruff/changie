package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

const (
	configEnvVar        = "CHANGIE_CONFIG_PATH"
	timeFormat   string = "20060102-150405"

	CreateFileMode os.FileMode = 0644
	CreateDirMode  os.FileMode = 0755

	AutoLevel  = "auto"
	MajorLevel = "major"
	MinorLevel = "minor"
	PatchLevel = "patch"
	NoneLevel  = "none"
	EmptyLevel = ""
)

var ConfigPaths []string = []string{
	".changie.yaml",
	".changie.yml",
}

var ErrConfigNotFound = errors.New("no changie config found")

// GetVersions will return, in semver sorted order, all released versions
type GetVersions func(Config) ([]*semver.Version, error)

// Config handles configuration for a project.
//
// Custom configuration path:
//
// By default Changie will try and load ".changie.yaml" and ".changie.yml" before running commands.
// If you want to change this path set the environment variable `CHANGIE_CONFIG_PATH` to the desired file.
//
// `export CHANGIE_CONFIG_PATH=./tools/changie.yaml`
//
// [Template Cache](#templatecache-type) handles all the template executions.
type Config struct {
	// Which configuration version we are using. Version 1 is used if none is provided.
	Version int `yaml:"version"`

	// Directory for change files, header file and unreleased files.
	// Relative to project root.
	// example: yaml
	// changesDir: .changes
	RootDir string `yaml:"rootDir" default:".changes"`

	// Prefix of environment variables to load for templates.
	// The prefix is removed from resulting key map.
	// example: yaml
	// # if we have an environment variable like so:
	// # export CHANGIE_PROJECT=changie
	// # we can use that in our templates if we set the prefix
	// envPrefix: "CHANGIE_"
	// versionFormat: "New release for {{.Env.PROJECT}}"
	EnvPrefix string `yaml:"envPrefix,omitempty"`

	// TODO: docs
	Fragment     FragmentConfig     `yaml:"fragment"`
	Project      ProjectConfig      `yaml:"project"`
	Component    ComponentConfig    `yaml:"component"`
	Kind         KindConfig         `yaml:"kind"`
	Change       ChangeConfig       `yaml:"change"`
	ReleaseNotes ReleaseNotesConfig `yaml:"releaseNotes"`
	Changelog    ChangelogConfig    `yaml:"changelog"`

	cachedEnvVars map[string]string
}

func (c *Config) KindFromKeyOrLabel(keyOrLabel string) *KindOptions {
	for _, kc := range c.Kind.Options {
		if kc.Prompt.Equals(keyOrLabel) {
			return &kc
		}
	}

	return nil
}

func (c *Config) KindHeader(keyOrLabel string) string {
	for _, kc := range c.Kind.Options {
		if kc.Prompt.Format != "" && kc.Prompt.Equals(keyOrLabel) {
			return kc.Prompt.Format
		}
	}

	return c.Kind.Prompt.Format
}

func (c *Config) ChangeFormatForKind(keyOrLabel string) string {
	for _, kc := range c.Kind.Options {
		if kc.Change.Prompt.Format != "" && kc.Prompt.Equals(keyOrLabel) {
			return kc.Change.Prompt.Format
		}
	}

	return c.Change.Prompt.Format
}

func (c *Config) EnvVars() map[string]string {
	if c.cachedEnvVars == nil {
		c.cachedEnvVars = LoadEnvVars(c, os.Environ())
	}

	return c.cachedEnvVars
}

// Save will save the config as a yaml file to the default path
func (c *Config) Save() error {
	bs, _ := yaml.Marshal(&c)
	return os.WriteFile(ConfigPaths[0], bs, CreateFileMode)
}

func (c *Config) ProjectByName(labelOrKey string) (*ProjectOptions, error) {
	if len(c.Project.Options) == 0 {
		return &ProjectOptions{}, nil
	}

	if len(labelOrKey) == 0 {
		return nil, errProjectRequired
	}

	for _, pc := range c.Project.Options {
		if pc.Label == labelOrKey || pc.Key == labelOrKey {
			return &pc, nil
		}
	}

	return nil, errProjectNotFound
}

func (c *Config) ProjectLabels() []string {
	projectLabels := make([]string, len(c.Project.Options))

	for i, pc := range c.Project.Options {
		if len(pc.Label) > 0 {
			projectLabels[i] = pc.Label
		} else {
			projectLabels[i] = pc.Key
		}
	}

	return projectLabels
}

// Exists returns whether or not a config already exists
func (c *Config) Exists() (bool, error) {
	for _, p := range ConfigPaths {
		if exists, err := FileExists(p); exists || err != nil {
			return exists, err
		}
	}

	return false, nil
}

func (c *Config) Changes(
	searchPaths []string,
	projectKey string,
) ([]Change, error) {
	var changes []Change

	changeFiles, err := c.ChangeFiles(searchPaths)
	if err != nil {
		return changes, err
	}

	for _, cf := range changeFiles {
		ch, err := LoadChange(cf)
		if err != nil {
			return changes, err
		}

		if len(projectKey) > 0 && ch.Project != projectKey {
			continue
		}

		ch.Env = c.EnvVars()

		kc := c.KindFromKeyOrLabel(ch.KindKey)
		if kc != nil {
			ch.KindLabel = kc.Prompt.Label
		} else if len(c.Kind.Options) > 0 {
			return nil, fmt.Errorf("%w: '%s'", ErrKindNotFound, ch.KindKey)
		}

		changes = append(changes, ch)
	}

	sort.Slice(changes, ChangeLess(c, changes))

	return changes, nil
}

func (c *Config) AllVersions(
	skipPrereleases bool,
	projectKey string,
) ([]*semver.Version, error) {
	allVersions := make([]*semver.Version, 0)

	versionsPath := c.RootDir
	if len(projectKey) > 0 {
		versionsPath = filepath.Join(versionsPath, projectKey)
	}

	fileInfos, err := os.ReadDir(versionsPath)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return allVersions, nil
	}

	if err != nil {
		return allVersions, fmt.Errorf("reading files from '%s': %w", versionsPath, err)
	}

	for _, file := range fileInfos {
		if file.Name() == c.ReleaseNotes.Header.FilePath || file.IsDir() {
			continue
		}

		versionString := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		v, err := semver.NewVersion(versionString)
		if err != nil {
			continue
		}

		if skipPrereleases && v.Prerelease() != "" {
			continue
		}

		allVersions = append(allVersions, v)
	}

	sort.Sort(sort.Reverse(semver.Collection(allVersions)))

	return allVersions, nil
}

func (c *Config) LatestVersion(
	skipPrereleases bool,
	projectKey string,
) (*semver.Version, error) {
	allVersions, err := c.AllVersions(skipPrereleases, projectKey)
	if err != nil {
		return nil, err
	}

	// if no versions exist default to v0.0.0
	if len(allVersions) == 0 {
		return semver.MustParse("v0.0.0"), nil
	}

	return allVersions[0], nil
}

func (c *Config) HighestAutoLevel(allChanges []Change) (string, error) {
	if len(allChanges) == 0 {
		return "", ErrNoChangesFoundForAuto
	}

	highestLevel := EmptyLevel
	err := ErrNoChangesFoundForAuto

	for _, change := range allChanges {
		for _, kc := range c.Kind.Options {
			if !kc.Prompt.Equals(change.Kind) {
				continue
			}

			switch kc.AutoLevel {
			case MajorLevel:
				// major is the highest one, so we can just return it
				return MajorLevel, nil
			case PatchLevel:
				if highestLevel == EmptyLevel {
					highestLevel = PatchLevel
					err = nil
				}
			case MinorLevel:
				// bump to minor if we have a minor level
				if highestLevel == PatchLevel || highestLevel == EmptyLevel {
					highestLevel = MinorLevel
					err = nil
				}
			case EmptyLevel:
				return EmptyLevel, ErrMissingAutoLevel
			}
		}
	}

	return highestLevel, err
}

func (c *Config) NextVersion(
	partOrVersion string,
	prerelease, meta []string,
	allChanges []Change,
	projectKey string,
) (*semver.Version, error) {
	var (
		err  error
		next *semver.Version
		ver  semver.Version
	)

	// if part or version is a valid version, then return it
	next, err = semver.NewVersion(partOrVersion)
	if err != nil {
		if !ValidBumpLevel(partOrVersion) {
			return nil, ErrBadVersionOrPart
		}

		// otherwise use a bump type command
		next, err = c.LatestVersion(false, projectKey)
		if err != nil {
			return nil, err
		}

		if partOrVersion == AutoLevel {
			partOrVersion, err = c.HighestAutoLevel(allChanges)
			if err != nil {
				return nil, err
			}
		}

		switch partOrVersion {
		case MajorLevel:
			ver = next.IncMajor()
		case MinorLevel:
			ver = next.IncMinor()
		case PatchLevel:
			ver = next.IncPatch()
		}

		next = &ver
	}

	if len(prerelease) > 0 {
		ver, err = next.SetPrerelease(strings.Join(prerelease, "."))
		if err != nil {
			return nil, err
		}

		next = &ver
	}

	if len(meta) > 0 {
		ver, err = next.SetMetadata(strings.Join(meta, "."))
		if err != nil {
			return nil, err
		}

		next = &ver
	}

	return next, nil
}

func (c *Config) ChangeFiles(searchPaths []string) ([]string, error) {
	var yamlFiles []string

	// add the unreleased path to any search paths included
	searchPaths = append(searchPaths, c.Fragment.Dir)

	// read all yaml files from our search paths
	for _, searchPath := range searchPaths {
		rootPath := filepath.Join(c.RootDir, searchPath)

		fileInfos, err := os.ReadDir(rootPath)
		if err != nil {
			return yamlFiles, err
		}

		for _, file := range fileInfos {
			if filepath.Ext(file.Name()) != ".yaml" {
				continue
			}

			path := filepath.Join(rootPath, file.Name())
			yamlFiles = append(yamlFiles, path)
		}
	}

	return yamlFiles, nil
}

// findConfigUpwards will recursively look up until we find a changie config file
func findConfigUpwards() ([]byte, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		for _, path := range ConfigPaths {
			bs, err := os.ReadFile(filepath.Join(currDir, path))
			if err == nil || !errors.Is(err, fs.ErrNotExist) {
				return bs, nil
			}
		}

		err := os.Chdir("..")
		if err != nil {
			return nil, err
		}

		lastDir := currDir

		currDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}

		// Keep going up, until going up is unchanged.
		if lastDir == currDir {
			return nil, ErrConfigNotFound
		}
	}
}

// LoadConfig will load the config from the default path
func LoadConfig() (*Config, error) {
	var (
		c   Config
		bs  []byte
		err error
	)

	customPath := os.Getenv(configEnvVar)
	if customPath != "" {
		bs, err = os.ReadFile(customPath)
	} else {
		bs, err = findConfigUpwards()
	}

	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	c.bumpV0()
	c.bumpV1()

	return &c, nil
}

func (c *Config) bumpV0() {
	// load v0 breaking changes into v1
	/*
		if c.FragmentFileFormat == "" {
			if len(c.Projects) > 0 {
				c.FragmentFileFormat += "{{.Project}}-"
			}

			if len(c.Components) > 0 {
				c.FragmentFileFormat += "{{.Component}}-"
			}

			if len(c.Kinds) > 0 {
				c.FragmentFileFormat += "{{.Kind}}-"
			}

			c.FragmentFileFormat += fmt.Sprintf("{{.Time.Format \"%v\"}}", timeFormat)
		}

		if c.VersionFileFormat == "" {
			c.VersionFileFormat = "{{.Version}}." + c.VersionExt
		}
	*/
}

func (c *Config) bumpV1() {
}
