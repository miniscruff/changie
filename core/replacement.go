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

type ReplaceData struct {
	Version         string
	VersionNoPrefix string
}

type Replacement struct {
	Path    string `yaml:"path"`
	Find    string `yaml:"find"`
	Replace string `yaml:"replace"`
	Flags   string `yaml:"flags,omitempty"`
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

	err = writeFile(r.Path, newData, 0644)
	if err != nil {
		return err
	}

	return nil
}
