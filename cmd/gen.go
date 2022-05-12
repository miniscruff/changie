package cmd

import (
	"fmt"
	"go/ast"
	godoc "go/doc"
	"go/parser"
	"go/token"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"
	"os"

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

type FieldProps struct {
	Name       string
	ParentType string
	TypeName   string
	Doc        string
	Example    string
	Required   bool
	Slice      bool
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

func genConfigDocs() error {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, "./core", nil, parser.ParseComments)
	if err != nil {
		return err
	}

	corePackage := packages["core"]
	p := godoc.New(corePackage, "./", 0)

	corePackages := make(map[string]*godoc.Type)
	for _, t := range p.Types {
		corePackages[t.Name] = t
	}

	file, err := os.Create("docs/content/config/_index.md")
	if err != nil {
		return err
	}

	defer file.Close()

	file.WriteString(fmt.Sprintf(`---
title: "Configuration"
date: %v
layout: single
singlePage: true
---
`, time.Now()))

	// temp write to a string and print it for testing
	typeQueue := make([]string, 0)
	typeQueue = append(typeQueue, "Config")
	completed := make(map[string]struct{}, 0)

	for {
		fmt.Printf("Queue: %v\n", typeQueue)
		if len(typeQueue) == 0 {
			break
		}

		typeName := typeQueue[0]
		typeQueue = append(typeQueue[1:])

		docType, found := corePackages[typeName]
		if !found {
			continue
		}

		// don't write objects more than once
		_, seen := completed[typeName]
		if seen {
			continue
		}

		completed[typeName] = struct{}{}

		fmt.Printf("Writing: %v\n", docType.Name)
		err = writeType(file, docType, &typeQueue)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeType(writer io.Writer, docType *godoc.Type, queue *[]string) error {
	_, err := writer.Write([]byte(fmt.Sprintf("## %s\n%s\n\n", docType.Name, docType.Doc)))
	if err != nil {
		return err
	}

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
			err = writeField(writer, docType.Name, field, queue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeField(writer io.Writer, parentName string, field *ast.Field, queue *[]string) error {
	props := FieldProps{
		Name:       field.Names[0].Name,
		ParentType: parentName,
		Doc:        field.Doc.Text(),
		Slice:      false,
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
	default:
		fmt.Printf("unknown field type: %T for field: %v\n", fieldType, field.Names[0])
		return fmt.Errorf("unknown field type: %T for field: '%v'\n", fieldType, field.Names[0])
	}

	before, after, found := strings.Cut(props.Doc, "example:")
	if found {
		props.Doc = before
		props.Example = after
	}

	typePrefix := ""
	if props.Slice {
		typePrefix = "[]"
	}

	writer.Write([]byte(fmt.Sprintf(
		"### %s%s\n",
		typePrefix,
		props.Name,
	)))
	writer.Write([]byte(fmt.Sprintf("type: `%s`\n", props.TypeName)))

	// TODO: default from struct tag
	// TODO: required/optional from struct tag
	if props.Required {
		writer.Write([]byte("required\n"))
	} else {
		writer.Write([]byte("optional\n"))
	}

	writer.Write([]byte("\n"))
	writer.Write([]byte(props.Doc))
	writer.Write([]byte("\n"))

	if props.Example != "" {
		//writer.Write([]byte(`{{< expand "Example" >}}`))
		writer.Write([]byte(props.Example))
		writer.Write([]byte("\n"))
		//writer.Write([]byte("{{< /expand >}}"))
		writer.Write([]byte("\n"))
	}

	return nil
}

/* psuedo code
go through all types in core package and cache them in a map
go through the config type:
	write type comment
	for each field:
		write field
		if field type is custom:
			add link to that type
			document that type after the current one ( queue? )

## <Type>
<type comment>

{{ range of alphabatized fields }}
### <field name>:<type> #type.fieldName
type: <type>
default: <default> if not empty
optional/required

<description>

(expand)
<example>
(/expand)

{{ end range }}

field comments:

type Config struct {
	// Generic description here anything goes
	// can use as many lines as necessary
	// Example:
	// Everything else is placed in the example expand
	Value string `required:"true/false" default="non-empty default"`
}
*/
