package cmd

import (
	"fmt"
	"go/ast"
	godoc "go/doc"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/miniscruff/changie/core"
)

const fmTemplate = `---
title: "%s"
description: "Help for using the '%s' command"
---
`

type CoreTypes map[string]*godoc.Type

type TypeProps struct {
	Name           string
	Doc            string
	File           string
	Line           int
	ExampleLang    string
	ExampleContent string
	Fields         []FieldProps
}

type FieldProps struct {
	Name             string
	TypeName         string
	Doc              string
	File             string
	Line             int
	Key              string
	MapKeyTypeName   string
	MapValueTypeName string
	ExampleLang      string
	ExampleContent   string
	TemplateType     string
	IsCustomType     bool
	Required         bool
	Slice            bool
}

type Gen struct {
	*cobra.Command
}

func NewGen() *Gen {
	g := &Gen{}

	cmd := &cobra.Command{
		Use:    "gen",
		Short:  "Generate documentation",
		Long:   `Generate markdown documentation.`,
		RunE:   g.Run,
		Args:   cobra.NoArgs,
		Hidden: true,
	}

	g.Command = cmd

	return g
}

func (g *Gen) Run(cmd *cobra.Command, args []string) error {
	err := os.MkdirAll(filepath.Join("docs", "config"), core.CreateDirMode)
	if err != nil {
		return fmt.Errorf("creating docs/config directory: %w", err)
	}

	file, err := os.Create(filepath.Join("docs", "config", "index.md"))
	if err != nil {
		return fmt.Errorf("creating or opening config index: %w", err)
	}

	defer file.Close()

	fset, corePackages := getCorePackages("core")

	writeConfigFrontMatter(file)
	genConfigDocs(fset, file, corePackages)

	// This auto generates a h4 which is added to our table of contents.
	cmd.Root().DisableAutoGenTag = true

	return doc.GenMarkdownTreeCustom(cmd.Root(), "docs/cli", filePrepender, linkHandler)
}

func filePrepender(filename string) string {
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	title := strings.ReplaceAll(base, "_", " ")

	return fmt.Sprintf(fmTemplate, title, title)
}

func linkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return strings.ToLower(base) + ".md"
}

func genConfigDocs(fset *token.FileSet, writer io.Writer, corePackages CoreTypes) {
	allTypeProps := buildUniqueTypes(fset, corePackages, "Config", "TemplateCache")

	// grab our root config and store it so we can sort the rest
	// but keep Config at the top
	rootConfigType := allTypeProps[0]
	allTypeProps = allTypeProps[1:]

	sort.Slice(
		allTypeProps,
		func(i, j int) bool {
			return allTypeProps[i].Name < allTypeProps[j].Name
		},
	)

	// put the now sorted items back after our root
	allTypeProps = append([]TypeProps{rootConfigType}, allTypeProps...)

	for _, typeProps := range allTypeProps {
		// we would have already written by this point, so we should be fine
		// not ideal but should work
		_ = writeType(writer, typeProps)
	}
}

func buildUniqueTypes(
	fset *token.FileSet,
	coreTypes CoreTypes,
	packageName ...string,
) []TypeProps {
	typeQueue := packageName
	completed := make(map[string]struct{}, 0)
	allTypeProps := make([]TypeProps, 0)

	for {
		if len(typeQueue) == 0 {
			break
		}

		typeName := typeQueue[0]
		typeQueue = typeQueue[1:]

		docType, found := coreTypes[typeName]
		if !found {
			continue
		}

		// don't write objects more than once
		_, seen := completed[typeName]
		if seen {
			continue
		}

		completed[typeName] = struct{}{}

		typeProps := buildType(fset, docType, coreTypes, &typeQueue)
		allTypeProps = append(allTypeProps, typeProps)
	}

	return allTypeProps
}

func getCorePackages(packageName string) (*token.FileSet, CoreTypes) {
	corePackages := make(CoreTypes)
	packagePath := fmt.Sprintf("./%v", packageName)

	fset := token.NewFileSet()
	packages, _ := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)

	corePackage := packages[packageName]
	p := godoc.New(corePackage, "./", 0)

	for _, t := range p.Types {
		corePackages[t.Name] = t
	}

	return fset, corePackages
}

func writeConfigFrontMatter(writer io.Writer) {
	_, _ = writer.Write([]byte(`---
title: "Configuration"
hide:
  - navigation
---
`))
}

func buildType(fset *token.FileSet, docType *godoc.Type, coreTypes CoreTypes, queue *[]string) TypeProps {
	tokPos := docType.Decl.TokPos
	typeFile := fset.File(tokPos)

	typeProps := TypeProps{
		Name: docType.Name,
		Doc:  docType.Doc,
		File: typeFile.Name(),
		Line: typeFile.Position(tokPos).Line,
	}

	fieldProps := make([]FieldProps, 0)

	// if any method has an example, we want to list it on the documentation,
	// kind of an arbitrary rule but it works for now.
	for _, method := range docType.Methods {
		if strings.Contains(method.Doc, "example:") {
			newField := buildMethod(fset, method)
			fieldProps = append(fieldProps, newField)
		}
	}

	before, after, found := strings.Cut(typeProps.Doc, "example:")
	if found {
		typeProps.Doc = before
		lang, content, _ := strings.Cut(strings.Trim(after, " "), "\n")
		typeProps.ExampleLang = lang
		typeProps.ExampleContent = formatExample(content)
	}

	for _, spec := range docType.Decl.Specs {
		typeSpec, _ := spec.(*ast.TypeSpec)
		structType, ok := typeSpec.Type.(*ast.StructType)

		if !ok {
			continue
		}

		for _, field := range structType.Fields.List {
			newField := buildField(fset, field, coreTypes, queue)
			fieldProps = append(fieldProps, newField)
		}
	}

	sort.Slice(
		fieldProps,
		func(i, j int) bool {
			return fieldProps[i].Name < fieldProps[j].Name
		},
	)

	typeProps.Fields = fieldProps

	return typeProps
}

func buildMethod(fset *token.FileSet, method *godoc.Func) FieldProps {
	tokPos := method.Decl.Pos()
	tokenFile := fset.File(tokPos)

	props := FieldProps{
		Name:         method.Name,
		File:         tokenFile.Name(),
		Line:         tokenFile.Line(tokPos),
		Doc:          method.Doc,
		Key:          strings.ToLower(method.Name),
		IsCustomType: false,
		Slice:        false,
	}

	before, after, found := strings.Cut(props.Doc, "example:")
	if found {
		props.Doc = before
		lang, content, _ := strings.Cut(strings.Trim(after, " "), "\n")
		props.ExampleLang = lang
		props.ExampleContent = formatExample(content)
	}

	return props
}

func buildField(fset *token.FileSet, field *ast.Field, coreTypes CoreTypes, queue *[]string) FieldProps {
	tokPos := field.Pos()
	tokenFile := fset.File(tokPos)

	props := FieldProps{
		Name:  field.Names[0].Name,
		File:  tokenFile.Name(),
		Line:  tokenFile.Line(tokPos),
		Doc:   field.Doc.Text(),
		Slice: false,
	}

	switch fieldType := field.Type.(type) {
	case *ast.Ident:
		*queue = append(*queue, fieldType.Name)
		props.TypeName = fieldType.Name
	case *ast.ArrayType:
		rootType := fieldType.Elt.(*ast.Ident)
		*queue = append(*queue, rootType.Name)
		props.TypeName = rootType.Name
		props.Slice = true
	case *ast.StarExpr:
		rootType := fieldType.X.(*ast.Ident)
		*queue = append(*queue, rootType.Name)
		props.TypeName = rootType.Name
		props.Required = false
	case *ast.SelectorExpr:
		rootType := fieldType.Sel
		*queue = append(*queue, rootType.Name)
		props.TypeName = rootType.Name
	case *ast.MapType:
		mapKeyType := fieldType.Key.(*ast.Ident)
		*queue = append(*queue, mapKeyType.Name)

		mapValueType := fieldType.Value.(*ast.Ident)
		*queue = append(*queue, mapValueType.Name)

		// reset type name to empty as we need to use map key type and value type
		props.TypeName = ""
		props.MapKeyTypeName = mapKeyType.Name
		props.MapValueTypeName = mapValueType.Name
	default:
		panic(fmt.Errorf("unknown field type: %T for field: '%v'", fieldType, field.Names[0]))
	}

	before, after, found := strings.Cut(props.Doc, "example:")
	if found {
		props.Doc = before
		lang, content, _ := strings.Cut(strings.Trim(after, " "), "\n")
		props.ExampleLang = lang
		props.ExampleContent = formatExample(content)
	}

	_, isCoreType := coreTypes[props.TypeName]
	props.IsCustomType = isCoreType
	props.Key = props.Name
	parseFieldStructTags(field, &props, queue)

	return props
}

func parseFieldStructTags(field *ast.Field, props *FieldProps, queue *[]string) {
	if field.Tag == nil {
		return
	}

	tags := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

	if yaml, ok := tags.Lookup("yaml"); ok {
		// yaml can include more than a name, but its the first word separated by comma
		key, _, _ := strings.Cut(yaml, ",")
		if key != "" && key != "-" {
			props.Key = key
		} else {
			props.Key = strings.ToLower(props.Key)
		}
	}

	if isRequired, ok := tags.Lookup("required"); ok {
		props.Required = isRequired == "true"
	}

	if templateType, ok := tags.Lookup("templateType"); ok {
		props.TemplateType = templateType
		*queue = append(*queue, templateType)
	}
}

func writeType(writer io.Writer, typeProps TypeProps) error {
	// Do not write our root Config type header
	if typeProps.Name != "Config" {
		anchor := strings.ToLower(typeProps.Name) + "-type"
		_, err := writer.Write([]byte(fmt.Sprintf(
			"## %s %v\n%s\n",
			typeProps.Name,
			buildSourceAnchorLink(anchor, typeProps.File, typeProps.Line),
			typeProps.Doc,
		)))

		if err != nil {
			return err
		}
	}

	if typeProps.ExampleContent != "" {
		_, _ = writer.Write([]byte(fmt.Sprintf(
			"??? Example\n    ```%v\n    %v\n    ```\n",
			typeProps.ExampleLang,
			strings.Trim(typeProps.ExampleContent, "\n"),
		)))
	}

	for _, f := range typeProps.Fields {
		err := writeField(writer, typeProps, f)
		if err != nil {
			return err
		}
	}

	_, _ = writer.Write([]byte("\n---\n"))

	return nil
}

func writeField(writer io.Writer, parent TypeProps, field FieldProps) error {
	typePrefix := ""
	if field.Slice {
		typePrefix = "[]"
	}

	anchor := strings.ToLower(parent.Name) + "-" + strings.ToLower(field.Key)

	_, err := writer.Write([]byte(fmt.Sprintf(
		"### %s %v\n",
		field.Key,
		buildSourceAnchorLink(anchor, field.File, field.Line),
	)))
	if err != nil {
		return err
	}

	switch {
	case field.IsCustomType:
		_, _ = writer.Write([]byte(fmt.Sprintf(
			"type: [%s%s](#%s-type)",
			typePrefix,
			field.TypeName,
			strings.ToLower(field.TypeName),
		)))
	case field.TypeName != "":
		_, _ = writer.Write([]byte(fmt.Sprintf("type: `%s%s`", typePrefix, field.TypeName)))
	case field.MapKeyTypeName != "":
		// currently no option of having a map of custom types or slices
		// but that is not used right now
		_, _ = writer.Write([]byte(fmt.Sprintf(
			"type: map [ `%s` ] `%s`",
			field.MapKeyTypeName,
			field.MapValueTypeName,
		)))
	}

	if field.TypeName != "" || field.MapKeyTypeName != "" {
		_, _ = writer.Write([]byte(" | "))

		if field.Required {
			_, _ = writer.Write([]byte("required"))
		} else {
			_, _ = writer.Write([]byte("optional"))
		}
	}

	if field.TemplateType != "" {
		_, _ = writer.Write([]byte(" | "))
		_, _ = writer.Write([]byte(fmt.Sprintf(
			"template type: [%s](#%s-type)",
			field.TemplateType,
			strings.ToLower(field.TemplateType),
		)))
	}

	_, _ = writer.Write([]byte("\n\n"))
	_, _ = writer.Write([]byte(field.Doc))

	if field.ExampleContent != "" {
		_, _ = writer.Write([]byte(fmt.Sprintf(
			"??? Example\n    ```%v\n    %v\n    ```\n",
			field.ExampleLang,
			strings.Trim(field.ExampleContent, "\n"),
		)))
	}

	_, _ = writer.Write([]byte("\n"))

	return nil
}

func buildSourceAnchorLink(anchor, fileName string, line int) string {
	file := strings.ReplaceAll(fileName, "\\", "/")
	url := fmt.Sprintf("https://github.com/miniscruff/changie/blob/<< current_version >>/%v#L%v", file, line)

	return "[:octicons-code-24:](" + url + ") [:octicons-link-24:](#" + anchor + ") {: #" + anchor + "}"
}

func formatExample(content string) string {
	return strings.ReplaceAll(strings.Trim(content, "\n\t "), "\n", "\n    ")
}
