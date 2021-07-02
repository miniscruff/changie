package core

import (
	"bytes"
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
	Path    string
	Find    string
	Replace string
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

	transformer := replace.Regexp(regexp.MustCompile(r.Find), buf.Bytes())

	newData, _, _ := transform.Bytes(transformer, fileData)

	err = writeFile(r.Path, newData, 0644)
	if err != nil {
		return err
	}

	return nil
}
