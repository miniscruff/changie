package core

import (
	"errors"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/miniscruff/changie/shared"
)

var ErrBadVersionOrPart = errors.New("part string is not a supported version or version increment")

func AppendFile(opener shared.OpenFiler, rootFile io.Writer, path string) error {
	otherFile, err := opener(path)
	if err != nil {
		return err
	}
	defer otherFile.Close()

	// Copy doesn't seem to break in any test I have
	// so ignoring this err return value
	_, _ = io.Copy(rootFile, otherFile)

	return nil
}

func GetAllVersions(readDir shared.ReadDirer, config Config, opts ...Option) ([]*semver.Version, error) {
	allVersions := make([]*semver.Version, 0)
	o := optionsFromDefault(opts...)

	fileInfos, err := readDir(config.ChangesDir)
	if err != nil {
		return allVersions, err
	}

	for _, file := range fileInfos {
		if file.Name() == config.HeaderPath || file.IsDir() {
			continue
		}

		versionString := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		v, err := semver.NewVersion(versionString)
		if err != nil {
			continue
		}

		if o.skipPrereleases && v.Prerelease() != "" {
			continue
		}

		allVersions = append(allVersions, v)
	}

	sort.Sort(sort.Reverse(semver.Collection(allVersions)))

	return allVersions, nil
}

func GetLatestVersion(readDir shared.ReadDirer, config Config,
	opts ...Option,
) (*semver.Version, error) {
	allVersions, err := GetAllVersions(readDir, config, opts...)
	if err != nil {
		return nil, err
	}

	// if no versions exist default to v0.0.0
	if len(allVersions) == 0 {
		return semver.MustParse("v0.0.0"), nil
	}

	return allVersions[0], nil
}

func GetNextVersion(
	readDir shared.ReadDirer, config Config, partOrVersion string, prerelease, meta []string,
	opts ...Option,
) (*semver.Version, error) {
	var (
		err  error
		next *semver.Version
		ver  semver.Version
	)
	o := optionsFromDefault(opts...)

	// if part or version is a valid version, then return it
	next, err = semver.NewVersion(partOrVersion)
	if err != nil {
		// otherwise use a bump type command
		next, err = GetLatestVersion(readDir, config)
		if err != nil {
			return nil, err
		}

		switch partOrVersion {
		case "major":
			ver = next.IncMajor()
		case "minor":
			ver = next.IncMinor()
		case "patch":
			if o.forcePatch {
				// force patch increment of the prerelease (by temporarily clearing the prerelease)
				pre := next.Prerelease()
				meta := next.Metadata()
				ver, _ = next.SetPrerelease("")
				ver, _ = ver.SetMetadata("")

				ver = ver.IncPatch()

				ver, _ = ver.SetPrerelease(pre)
				ver, _ = ver.SetMetadata(meta)

				break
			}

			// otherwise, applies default IncPatch semantics from the semver package
			ver = next.IncPatch()
		default:
			return nil, ErrBadVersionOrPart
		}

		next = &ver
	}

	ver, err = next.SetPrerelease(strings.Join(prerelease, "."))
	if err != nil {
		return nil, err
	}

	ver, err = ver.SetMetadata(strings.Join(meta, "."))
	if err != nil {
		return nil, err
	}

	return &ver, nil
}

func FindChangeFiles(
	config Config,
	readDir shared.ReadDirer,
	searchPaths []string,
) ([]string, error) {
	var yamlFiles []string

	// add the unreleased path to any search paths included
	searchPaths = append(searchPaths, config.UnreleasedDir)

	// read all yaml files from our search paths
	for _, searchPath := range searchPaths {
		rootPath := filepath.Join(config.ChangesDir, searchPath)
		fileInfos, err := readDir(rootPath)

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
