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
	Name           string
	Doc            string
	ExampleLang    string
	ExampleContent string
	Fields         []FieldProps
}

type FieldProps struct {
	Name             string
	Key              string
	TypeName         string
	MapKeyTypeName   string
	MapValueTypeName string
	Doc              string
	ExampleLang      string
	ExampleContent   string
	TemplateType     string
	IsCustomType     bool
	Required         bool
	Slice            bool
}

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate documentation",
	Long:  `Generate markdown documentation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Create(filepath.Join("docs", "content", "config", "_index.md"))
		if err != nil {
			return fmt.Errorf("unable to create or open config index: %w", err)
		}

		defer file.Close()

		corePackages := getCorePackages("core")

		writeConfigFrontMatter(file, time.Now)
		// only way this can fail is a bad writer, but we assume the file is good above
		// maybe not the best approach but for code generation that is only used internally
		// it should be fine.
		_ = genConfigDocs(file, corePackages)

		return doc.GenMarkdownTreeCustom(rootCmd, "docs/content/cli", filePrepender, linkHandler)
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

func genConfigDocs(writer io.Writer, corePackages CoreTypes) error {
	allTypeProps := buildUniqueTypes(writer, corePackages, "Config", "TemplateCache")

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

	return nil
}

func buildUniqueTypes(
	writer io.Writer,
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

		typeProps := buildType(docType, coreTypes, &typeQueue)
		allTypeProps = append(allTypeProps, typeProps)
	}

	return allTypeProps
}

func getCorePackages(packageName string) CoreTypes {
	corePackages := make(CoreTypes)
	packagePath := fmt.Sprintf("./%v", packageName)

	fset := token.NewFileSet()
	packages, _ := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)

	corePackage := packages[packageName]
	p := godoc.New(corePackage, "./", 0)

	for _, t := range p.Types {
		corePackages[t.Name] = t
	}

	return corePackages
}

func writeConfigFrontMatter(writer io.Writer, nower func() time.Time) {
	_, _ = writer.Write([]byte(fmt.Sprintf(`---
title: "Configuration"
date: %v
layout: single
singlePage: true
---
`, nower())))
}

func buildType(docType *godoc.Type, coreTypes CoreTypes, queue *[]string) TypeProps {
	typeProps := TypeProps{
		Name: docType.Name,
		Doc:  docType.Doc,
	}

	fieldProps := make([]FieldProps, 0)

	// if any method has an example, we want to list it on the documentation,
	// kind of an arbitrary rule but it works for now.
	for _, method := range docType.Methods {
		if strings.Contains(method.Doc, "example:") {
			newField := buildMethod(method)
			fieldProps = append(fieldProps, newField)
		}
	}

	before, after, found := strings.Cut(typeProps.Doc, "example:")
	if found {
		typeProps.Doc = before
		lang, content, _ := strings.Cut(strings.Trim(after, " "), "\n")
		typeProps.ExampleLang = lang
		typeProps.ExampleContent = content
	}

	for _, spec := range docType.Decl.Specs {
		typeSpec, _ := spec.(*ast.TypeSpec)
		structType, ok := typeSpec.Type.(*ast.StructType)

		if !ok {
			continue
		}

		for _, field := range structType.Fields.List {
			newField := buildField(field, coreTypes, queue)
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

func buildMethod(method *godoc.Func) FieldProps {
	props := FieldProps{
		Name:         method.Name,
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
		props.ExampleContent = content
	}

	return props
}

func buildField(field *ast.Field, coreTypes CoreTypes, queue *[]string) FieldProps {
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
		props.ExampleContent = content
	}

	_, isCoreType := coreTypes[props.TypeName]
	props.IsCustomType = isCoreType
	props.Key = props.Name

	if field.Tag != nil {
		tags := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

		if yaml, ok := tags.Lookup("yaml"); ok {
			// yaml can include more than a name, but its the first word separated by comma
			key, _, _ := strings.Cut(yaml, ",")
			if key != "" {
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

	return props
}

func writeType(writer io.Writer, typeProps TypeProps) error {
	// Do not write our root Config type header
	if typeProps.Name != "Config" {
		_, err := writer.Write([]byte(fmt.Sprintf(
			"## %s {#%s-type}\n%s\n",
			typeProps.Name,
			strings.ToLower(typeProps.Name),
			typeProps.Doc,
		)))
		if err != nil {
			return err
		}
	}

	if typeProps.ExampleContent != "" {
		_, _ = writer.Write([]byte(fmt.Sprintf(`
{{< expand "Example" "%v" >}}
%v
{{< /expand >}}
`,
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

	_, err := writer.Write([]byte(fmt.Sprintf(
		"### %s {#%s-%s}\n",
		field.Key,
		strings.ToLower(parent.Name),
		strings.ToLower(field.Key),
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
		_, _ = writer.Write([]byte(fmt.Sprintf(`
{{< expand "Example" "%v" >}}
%v
{{< /expand >}}
`,
			field.ExampleLang,
			strings.Trim(field.ExampleContent, "\n"),
		)))
	}

	_, _ = writer.Write([]byte("\n"))

	return nil
}
