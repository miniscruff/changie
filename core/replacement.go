package core

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	"github.com/icholy/replace"
	"golang.org/x/text/transform"

	"github.com/miniscruff/changie/shared"
)

// Template data used for replacing version values.
type ReplaceData struct {
	// Version of the release, will include "v" prefix if used
	Version string
	// Version of the release without the "v" prefix if used
	VersionNoPrefix string
	// Major value of the version
	Major int
	// Minor value of the version
	Minor int
	// Patch value of the version
	Patch int
	// Prerelease value of the version
	Prerelease string
	// Metadata value of the version
	Metadata string
}

// Replacement handles the finding and replacing values when merging the changelog.
// This can be used to keep version strings in-sync when preparing a release without having to
// manually update them.
// This works similar to the find and replace from IDE tools but also includes the file path of the
// file.
// example: yaml
// # NodeJS package.json
// replacements:
// - path: package.json
//   find: '  "version": ".*",'
//   replace: '  "version": "{{.VersionNoPrefix}}",'
type Replacement struct {
	// Path of the file to find and replace in.
	Path string `yaml:"path" required:"true"`
	// Regular expression to search for in the file.
	Find string `yaml:"find" required:"true"`
	// Template string to replace the line with.
	Replace string `yaml:"replace" required:"true" templateType:"ReplaceData"`
	// Optional regular expression mode flags.
	// Defaults to the m flag for multiline such that ^ and $ will match the start and end of each line
	// and not just the start and end of the string.
	//
	// For more details on regular expression flags in Go view the
	// [regexp/syntax](https://pkg.go.dev/regexp/syntax).
	Flags string `yaml:"flags,omitempty" default:"m"`
}

func (r Replacement) Execute(
	readFile shared.ReadFiler,
	writeFile shared.WriteFiler,
	data ReplaceData,
) error {
	templ, err := template.New("replacement").Parse(r.Replace)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	//nolint:errcheck
	templ.Execute(&buf, data)

	fileData, err := readFile(r.Path)
	if err != nil {
		return err
	}

	flags := r.Flags
	if flags == "" {
		flags = "m"
	}

	regexString := fmt.Sprintf("(?%s)%s", flags, r.Find)
	transformer := replace.Regexp(regexp.MustCompile(regexString), buf.Bytes())
	newData, _, _ := transform.Bytes(transformer, fileData)

	err = writeFile(r.Path, newData, CreateFileMode)
	if err != nil {
		return err
	}

	return nil
}
