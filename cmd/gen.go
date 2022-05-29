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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
summary: "Help for using the '%s' command"
---
`

type CoreTypes map[string]*godoc.Type

type TypeProps struct {
	Name   string
	Doc    string
	Fields []FieldProps
}

type FieldProps struct {
	Name           string
	Key            string
	TypeName       string
	Doc            string
	ExampleLang    string
	ExampleContent string
	TemplateType   string
	IsCustomType   bool
	Required       bool
	Slice          bool
}

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate documentation",
	Long:  `Generate markdown documentation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := genConfigDocs()
		if err != nil {
			return err
		}

		return doc.GenMarkdownTreeCustom(rootCmd, "website/content/cli", filePrepender, linkHandler)
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(genCmd)
}

func filePrepender(filename string) string {
	now := time.Now().Format(time.RFC3339)
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	title := strings.ReplaceAll(base, "_", " ")
	url := "/cli/" + strings.ToLower(base) + "/"

	return fmt.Sprintf(fmTemplate, now, title, base, url, title)
}

func linkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "/cli/" + strings.ToLower(base) + "/"
}

func genConfigDocs() error {
	file, err := os.Create("website/content/config/_index.md")
	if err != nil {
		return err
	}

	defer file.Close()
	var writer io.Writer = file

	writeConfigFrontMatter(writer, time.Now)

	corePackages, err := getCorePackages("core")
	if err != nil {
		return err
	}

	allTypeProps, err := buildUniqueTypes(writer, corePackages, "Config")
	if err != nil {
		return err
	}

	// TODO: build template funcs and types

	// grab our root config and store it so we can sort the rest
	// but keep Config at the top
	rootConfigType := allTypeProps[0]
	allTypeProps = append(allTypeProps[1:])

	sort.Slice(
		allTypeProps,
		func(i, j int) bool {
			return allTypeProps[i].Name < allTypeProps[j].Name
		},
	)

	// put the now sorted items back after our root
	allTypeProps = append([]TypeProps{rootConfigType}, allTypeProps...)

	for _, typeProps := range allTypeProps {
		err = writeType(writer, typeProps)
		if err != nil {
			return err
		}
	}

	return err
}

func buildUniqueTypes(
	writer io.Writer,
	coreTypes CoreTypes,
	packageName string,
) ([]TypeProps, error) {
	typeQueue := []string{packageName}
	completed := make(map[string]struct{}, 0)
	allTypeProps := make([]TypeProps, 0)

	for {
		if len(typeQueue) == 0 {
			break
		}

		typeName := typeQueue[0]
		typeQueue = append(typeQueue[1:])

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

		fmt.Println("building type:", typeName)

		typeProps, err := buildType(docType, coreTypes, &typeQueue)
		if err != nil {
			return allTypeProps, err
		}

		allTypeProps = append(allTypeProps, typeProps)
	}

	return allTypeProps, nil
}

func getCorePackages(packageName string) (CoreTypes, error) {
	corePackages := make(CoreTypes)
	packagePath := fmt.Sprintf("./%v", packageName)

	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)
	if err != nil {
		return corePackages, err
	}

	corePackage := packages[packageName]
	p := godoc.New(corePackage, "./", 0)

	for _, t := range p.Types {
		corePackages[t.Name] = t
	}
	return corePackages, nil
}

func writeConfigFrontMatter(writer io.Writer, nower func() time.Time) {
	writer.Write([]byte(fmt.Sprintf(`---
title: "Configuration"
date: %v
layout: single
singlePage: true
---
`, nower())))
}

func buildType(docType *godoc.Type, coreTypes CoreTypes, queue *[]string) (TypeProps, error) {
	typeProps := TypeProps{
		Name: docType.Name,
		Doc:  docType.Doc,
	}

	fieldProps := make([]FieldProps, 0)

	for _, spec := range docType.Decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		for _, field := range structType.Fields.List {
			newField, err := buildField(field, coreTypes, queue)
			if err != nil {
				return typeProps, err
			}

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

	return typeProps, nil
}

func buildField(field *ast.Field, coreTypes CoreTypes, queue *[]string) (FieldProps, error) {
	props := FieldProps{
		Name:  field.Names[0].Name,
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
		rootType := fieldType.Key.(*ast.Ident)
		*queue = append(*queue, rootType.Name)
		props.TypeName = rootType.Name
	default:
		fmt.Printf("unknown field type: %T for field: %v\n", fieldType, field.Names[0])
		return props, fmt.Errorf("unknown field type: %T for field: '%v'\n", fieldType, field.Names[0])
	}

	before, after, found := strings.Cut(props.Doc, "example:")
	if found {
		props.Doc = before
		lang, content, _ := strings.Cut(strings.Trim(after, " "), "\n")
		props.ExampleLang = lang
		props.ExampleContent = content
	}

	_, isCoreType := coreTypes[props.TypeName]
	props.IsCustomType = isCoreType
	props.Key = strings.ToLower(props.Name)

	if field.Tag != nil {
		var tags reflect.StructTag = reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

		if yaml, ok := tags.Lookup("yaml"); ok {
			// yaml can include more than a name, but its the first word separated by comma
			key, _, _ := strings.Cut(yaml, ",")
			if key != "" {
				props.Key = key
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

	return props, nil
}

func writeType(writer io.Writer, typeProps TypeProps) error {
	// Do not write our root Config type header
	if typeProps.Name != "Config" {
		_, err := writer.Write([]byte(fmt.Sprintf(
			"## %s {#%s-type}\n%s\n\n",
			typeProps.Name,
			strings.ToLower(typeProps.Name),
			typeProps.Doc,
		)))
		if err != nil {
			return err
		}
	}

	for _, f := range typeProps.Fields {
		err := writeField(writer, typeProps, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeField(writer io.Writer, parent TypeProps, field FieldProps) error {
	typePrefix := ""
	if field.Slice {
		typePrefix = "[]"
	}

	writer.Write([]byte(fmt.Sprintf(
		"### %s {#%s-%s}\n",
		field.Key,
		strings.ToLower(parent.Name),
		strings.ToLower(field.Key),
	)))

	if field.IsCustomType {
		writer.Write([]byte(fmt.Sprintf(
			"type: [%s%s](#%s-type)",
			typePrefix,
			field.TypeName,
			strings.ToLower(field.TypeName),
		)))
	} else {
		writer.Write([]byte(fmt.Sprintf("type: `%s%s`", typePrefix, field.TypeName)))
	}

	writer.Write([]byte(" | "))

	// TODO: templateData from struct tag
	if field.Required {
		writer.Write([]byte("required"))
	} else {
		writer.Write([]byte("optional"))
	}

	if field.TemplateType != "" {
		writer.Write([]byte(" | "))
		writer.Write([]byte(fmt.Sprintf(
			"template type: [%s](#%s-type)",
			field.TemplateType,
			strings.ToLower(field.TemplateType),
		)))
	}

	writer.Write([]byte("\n\n"))
	writer.Write([]byte(field.Doc))
	writer.Write([]byte("\n\n"))

	if field.ExampleContent != "" {
		writer.Write([]byte(fmt.Sprintf(`{{< expand "Example" "%v" >}}`, field.ExampleLang)))
		writer.Write([]byte(field.ExampleContent))
		writer.Write([]byte("{{< /expand >}}"))
		writer.Write([]byte("\n"))
	}

	return nil
}
