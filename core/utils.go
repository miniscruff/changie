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

func GetAllVersions(readDir shared.ReadDirer, config Config) ([]*semver.Version, error) {
	allVersions := make([]*semver.Version, 0)

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

		allVersions = append(allVersions, v)
	}

	sort.Sort(sort.Reverse(semver.Collection(allVersions)))

	return allVersions, nil
}

func GetLatestVersion(readDir shared.ReadDirer, config Config) (*semver.Version, error) {
	allVersions, err := GetAllVersions(readDir, config)
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
	readDir shared.ReadDirer, config Config, partOrVersion string,
) (*semver.Version, error) {
	// if part or version is a valid version, then return it
	newVer, newErr := semver.NewVersion(partOrVersion)
	if newErr == nil {
		return newVer, nil
	}

	// otherwise use a bump type command
	ver, err := GetLatestVersion(readDir, config)
	if err != nil {
		return nil, err
	}

	var next semver.Version

	switch partOrVersion {
	case "major":
		next = ver.IncMajor()
	case "minor":
		next = ver.IncMinor()
	case "patch":
		next = ver.IncPatch()
	default:
		return nil, ErrBadVersionOrPart
	}

	return &next, nil
}
