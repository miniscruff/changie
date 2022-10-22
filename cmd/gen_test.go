package cmd

import (
	"errors"
	"fmt"
	"go/ast"
	godoc "go/doc"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

type ErrorWriter struct {
	err error
}

func (ew *ErrorWriter) Write(p []byte) (int, error) {
	return 0, ew.err
}

var errMock = errors.New("mock error")

var _ = Describe("Gen", func() {
	It("file prepender creates frontmatter", func() {
		prepender := filePrepender("my_command.go")
		Expect(prepender).To(ContainSubstring(`date:`))
		Expect(prepender).To(ContainSubstring(`title: "my command"`))
		Expect(prepender).To(ContainSubstring(`slug: my_command`))
		Expect(prepender).To(ContainSubstring(`url: /cli/my_command/`))
		Expect(prepender).To(ContainSubstring(`summary:`))
	})

	It("link handler links to other pages", func() {
		link := linkHandler("my_command.go")
		Expect(link).To(Equal("/cli/my_command/"))
	})

	It("can gen files", func() {
		// move our wd to project root instead of cmd dir
		wd, _ := os.Getwd()
		Expect(os.Chdir(filepath.Join(wd, ".."))).To(Succeed())
		defer func() {
			Expect(os.Chdir(wd)).To(Succeed())
		}()

		contentPath := filepath.Join("docs", "content")
		cliDocsPath := filepath.Join(contentPath, "cli")
		configDocsPath := filepath.Join(contentPath, "config", "_index.md")

		// ignore any bad error trying to remove the file, it may not exist and that is ok
		_ = os.Remove(configDocsPath)

		cliDocsFiles, err := os.ReadDir(cliDocsPath)
		Expect(err).To(BeNil())

		// remove all our generated cli docs before running gen
		for _, fp := range cliDocsFiles {
			if fp.Name() == "_index.md" || fp.Name() == ".gitignore" {
				continue
			}

			Expect(os.Remove(filepath.Join(cliDocsPath, fp.Name()))).To(Succeed())
		}

		// run the gen command
		rootCmd.SetArgs([]string{"gen"})
		Expect(Execute("")).To(Succeed())

		// test a few files exist
		Expect(filepath.Join(cliDocsPath, "changie.md")).To(BeARegularFile())
		Expect(filepath.Join(cliDocsPath, "changie_init.md")).To(BeARegularFile())
		Expect(filepath.Join(cliDocsPath, "changie_batch.md")).To(BeARegularFile())
		Expect(filepath.Join(cliDocsPath, "changie_merge.md")).To(BeARegularFile())
		Expect(configDocsPath).To(BeARegularFile())
	})

	It("will fail to gen if content path is missing", func() {
		startDir, err := os.Getwd()
		Expect(err).To(BeNil())

		tempDir, err := os.MkdirTemp("", "gen-test-"+CurrentGinkgoTestDescription().TestText)
		Expect(err).To(BeNil())

		err = os.Chdir(tempDir)
		Expect(err).To(BeNil())

		defer func() {
			err := os.Chdir(startDir)
			Expect(err).To(BeNil())
		}()

		rootCmd.SetArgs([]string{"gen"})
		Expect(Execute("")).NotTo(Succeed())
	})

	It("write type returns error on bad writer", func() {
		badWriter := &ErrorWriter{
			err: errMock,
		}

		parent := TypeProps{
			Name: "Parent",
		}

		err := writeType(badWriter, parent)
		Expect(err).NotTo(BeNil())
	})

	It("write type returns error on bad writer for types", func() {
		badWriter := &ErrorWriter{
			err: errMock,
		}

		// special case if our Config type has a bad writer
		// our type may fail to write
		parent := TypeProps{
			Name: "Config",
			Fields: []FieldProps{{
				Key: "Anything",
			}},
		}

		err := writeType(badWriter, parent)
		Expect(err).NotTo(BeNil())
	})

	It("write field returns error on bad writer", func() {
		badWriter := &ErrorWriter{
			err: errMock,
		}

		parent := TypeProps{
			Name: "Parent",
		}

		field := FieldProps{
			Key: "Key",
		}

		err := writeField(badWriter, parent, field)
		Expect(err).NotTo(BeNil())
	})

	It("build field panics with a bad field type", func() {
		corePackages := make(CoreTypes)
		queue := make([]string, 0)
		field := &ast.Field{
			Names: []*ast.Ident{
				{Name: "MyVar", NamePos: 1},
			},
			Doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "This is our var docs"},
				},
			},
			Type: &ast.Ellipsis{},
		}

		fset := token.NewFileSet()
		_ = fset.AddFile("file.go", -1, 1000)

		Expect(func() {
			_ = buildField(fset, field, corePackages, &queue)
		}).To(Panic())
	})
})

var _ = DescribeTable(
	"build type",
	func(docType *godoc.Type, output TypeProps, expectedQueue ...string) {
		corePackages := make(CoreTypes)
		// include one package so we can check is custom type
		corePackages["MyOptionalType"] = nil
		queue := make([]string, 0)
		fset := token.NewFileSet()
		_ = fset.AddFile("a.go", -1, 100)
		_ = fset.AddFile("b.go", -1, 100)
		_ = fset.AddFile("c.go", -1, 100)
		_ = fset.AddFile("d.go", -1, 100)

		props := buildType(fset, docType, corePackages, &queue)
		Expect(props).To(Equal(output))
		Expect(queue).To(Equal(expectedQueue))
	},
	Entry(
		"type with example and a method",
		&godoc.Type{
			Name: "MyType",
			Doc: `Type docs
example: yaml
here is an example`,
			Methods: []*godoc.Func{{
				Name: "MyFuncWithExample",
				Decl: &ast.FuncDecl{
					Type: &ast.FuncType{
						Func: 250,
					},
				},
				Doc: `Some func description
example: yaml
has example`,
			}},
			Decl: &ast.GenDecl{
				TokPos: 1,
				Specs:  []ast.Spec{},
			},
		},
		TypeProps{
			Name:           "MyType",
			Doc:            "Type docs\n",
			File:           "a.go",
			Line:           1,
			ExampleLang:    "yaml",
			ExampleContent: "here is an example",
			Fields: []FieldProps{
				{
					Name:           "MyFuncWithExample",
					Key:            "myfuncwithexample",
					File:           "c.go",
					Line:           1,
					TypeName:       "",
					Doc:            "Some func description\n",
					ExampleLang:    "yaml",
					ExampleContent: "has example",
					TemplateType:   "",
					IsCustomType:   false,
					Required:       false,
					Slice:          false,
				},
			},
		},
	),
	Entry(
		"type with multiple specs",
		&godoc.Type{
			Name:    "MyType",
			Doc:     "Type docs",
			Methods: []*godoc.Func{},
			Decl: &ast.GenDecl{
				TokPos: 1,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: &ast.Ident{
							Name: "RandomInterface",
							NamePos: 10,
						},
						Type: &ast.InterfaceType{},
					},
					&ast.TypeSpec{
						Name: &ast.Ident{
							Name: "ParentType",
							NamePos: 30,
						},
						Type: &ast.StructType{
							Struct: 150,
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{
											{Name: "OtherVar", NamePos: 250},
										},
										Doc: &ast.CommentGroup{
											List: []*ast.Comment{
												{Text: "othervar docs"},
											},
										},
										Type: &ast.Ident{
											Name: "string",
										},
									},
									{
										Names: []*ast.Ident{
											{Name: "MyVar", NamePos: 350},
										},
										Doc: &ast.CommentGroup{
											List: []*ast.Comment{
												{Text: "myvar docs"},
											},
										},
										Type: &ast.Ident{
											Name: "int",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		TypeProps{
			Name:           "MyType",
			Doc:            "Type docs",
			File:           "a.go",
			Line:           1,
			ExampleLang:    "",
			ExampleContent: "",
			Fields: []FieldProps{
				{
					Name:           "MyVar",
					Key:            "MyVar",
					TypeName:       "int",
					Doc:            "myvar docs\n",
					File: "d.go",
					Line: 1,
					ExampleLang:    "",
					ExampleContent: "",
					TemplateType:   "",
					IsCustomType:   false,
					Required:       false,
					Slice:          false,
				},
				{
					Name:           "OtherVar",
					Key:            "OtherVar",
					TypeName:       "string",
					File: "c.go",
					Line: 1,
					Doc:            "othervar docs\n",
					ExampleLang:    "",
					ExampleContent: "",
					TemplateType:   "",
					IsCustomType:   false,
					Required:       false,
					Slice:          false,
				},
			},
		},
		"string",
		"int",
	),
)

var _ = DescribeTable(
	"build method",
	func(method *godoc.Func, output FieldProps) {
		fset := token.NewFileSet()
		_ = fset.AddFile("file.go", -1, 1000)
		props := buildMethod(fset, method)
		Expect(props).To(Equal(output))
	},
	Entry(
		"method",
		&godoc.Func{
			Name: "MyFunc",
			Decl: &ast.FuncDecl{
				Type: &ast.FuncType{
					Func: 1,
				},
			},
			Doc: "Func docs",
		},
		FieldProps{
			Name:           "MyFunc",
			Key:            "myfunc",
			File:           "file.go",
			Line:           1,
			TypeName:       "",
			Doc:            "Func docs",
			ExampleLang:    "",
			ExampleContent: "",
			TemplateType:   "",
			IsCustomType:   false,
			Required:       false,
			Slice:          false,
		},
	),
	Entry(
		"method with example",
		&godoc.Func{
			Name: "MyFuncWithExample",
			Decl: &ast.FuncDecl{
				Type: &ast.FuncType{
					Func: 1,
				},
			},
			Doc: `Some func description
example: yaml
doSomething: '{{.Body}}'
`,
		},
		FieldProps{
			Name:           "MyFuncWithExample",
			Key:            "myfuncwithexample",
			TypeName:       "",
			File:           "file.go",
			Line:           1,
			Doc:            "Some func description\n",
			ExampleLang:    "yaml",
			ExampleContent: "doSomething: '{{.Body}}'\n",
			TemplateType:   "",
			IsCustomType:   false,
			Required:       false,
			Slice:          false,
		},
	),
)

var _ = DescribeTable(
	"build field",
	func(field *ast.Field, output FieldProps, expectedQueue ...string) {
		corePackages := make(CoreTypes)
		// include one package so we can check is custom type
		corePackages["MyOptionalType"] = nil
		queue := make([]string, 0)
		fset := token.NewFileSet()
		_ = fset.AddFile("file.go", -1, 1000)

		props := buildField(fset, field, corePackages, &queue)
		Expect(props).To(Equal(output))
		Expect(queue).To(Equal(expectedQueue))
	},
	Entry(
		"identifier",
		&ast.Field{
			Names: []*ast.Ident{
				{Name: "MyVar", NamePos: 1},
			},
			Doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "This is our var docs"},
				},
			},
			Type: &ast.Ident{
				Name: "MyType",
			},
		},
		FieldProps{
			Name:           "MyVar",
			Key:            "MyVar",
			TypeName:       "MyType",
			File:           "file.go",
			Line:           1,
			Doc:            "This is our var docs\n",
			ExampleLang:    "",
			ExampleContent: "",
			TemplateType:   "",
			IsCustomType:   false,
			Required:       false,
			Slice:          false,
		},
		"MyType",
	),
	Entry(
		"required slice with an example",
		&ast.Field{
			Names: []*ast.Ident{
				{Name: "MySliceVar", NamePos: 1},
			},
			Doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "This is a slice with an example"},
					{Text: "example: yaml"},
					{Text: "mySliceVar: [1, 2, 3]"},
				},
			},
			Type: &ast.ArrayType{
				Elt: &ast.Ident{Name: "MySliceType"},
			},
			Tag: &ast.BasicLit{Value: `required:"true" yaml:"mySliceVar"`},
		},
		FieldProps{
			Name:           "MySliceVar",
			Key:            "mySliceVar",
			TypeName:       "MySliceType",
			File:           "file.go",
			Line:           1,
			Doc:            "This is a slice with an example\n",
			ExampleLang:    "yaml",
			ExampleContent: "mySliceVar: [1, 2, 3]\n",
			TemplateType:   "",
			IsCustomType:   false,
			Required:       true,
			Slice:          true,
		},
		"MySliceType",
	),
	Entry(
		"pointer with template type",
		&ast.Field{
			Names: []*ast.Ident{
				{Name: "Optional", NamePos: 1},
			},
			Type: &ast.StarExpr{
				X: &ast.Ident{Name: "MyOptionalType"},
			},
			Tag: &ast.BasicLit{Value: `yaml:"" templateType:"CustomData"`},
		},
		FieldProps{
			Name:           "Optional",
			Key:            "optional",
			TypeName:       "MyOptionalType",
			File:           "file.go",
			Line:           1,
			Doc:            "",
			ExampleLang:    "",
			ExampleContent: "",
			TemplateType:   "CustomData",
			IsCustomType:   true,
			Required:       false,
			Slice:          false,
		},
		"MyOptionalType",
		"CustomData",
	),
	Entry(
		"selector expressions",
		&ast.Field{
			Names: []*ast.Ident{
				{Name: "MyVarName", NamePos: 1},
			},
			Type: &ast.SelectorExpr{
				Sel: &ast.Ident{Name: "MySelectorType"},
			},
		},
		FieldProps{
			Name:           "MyVarName",
			Key:            "MyVarName",
			TypeName:       "MySelectorType",
			File:           "file.go",
			Line:           1,
			Doc:            "",
			ExampleLang:    "",
			ExampleContent: "",
			TemplateType:   "",
			IsCustomType:   false,
			Required:       false,
			Slice:          false,
		},
		"MySelectorType",
	),
	Entry(
		"map type",
		&ast.Field{
			Names: []*ast.Ident{
				{Name: "MyMap", NamePos: 1},
			},
			Type: &ast.MapType{
				Key:   &ast.Ident{Name: "string"},
				Value: &ast.Ident{Name: "string"},
			},
		},
		FieldProps{
			Name:             "MyMap",
			Key:              "MyMap",
			File:             "file.go",
			Line:             1,
			TypeName:         "",
			MapKeyTypeName:   "string",
			MapValueTypeName: "string",
			Doc:              "",
			ExampleLang:      "",
			ExampleContent:   "",
			TemplateType:     "",
			IsCustomType:     false,
			Required:         false,
			Slice:            false,
		},
		"string",
		"string",
	),
)

var _ = DescribeTable(
	"gen write type",
	func(typeProps TypeProps, output string) {
		format.TruncatedDiff = false
		defer func() {
			format.TruncatedDiff = true
		}()

		writer := strings.Builder{}

		err := writeType(&writer, typeProps)
		Expect(err).To(BeNil())
		Expect(writer.String()).To(Equal(output))
	},
	Entry(
		"basic type with field",
		TypeProps{
			Name: "Minimum",
			File: "core/mini.go",
			Line: 22,
			Doc:  "type description",
			Fields: []FieldProps{{
				Name: "MyField",
				Key:  "myField",
				File: "core/field.go",
				Line: 13,
				Doc:  "field description",
			}},
		},
		`## Minimum {{< source name="Minimum" file="core/mini.go" line="22" >}} {#minimum-type}
type description
### myField {{< source name="MyField" file="core/field.go" line="13" >}} {#minimum-myfield}


field description

---
`),
	Entry(
		"skips config",
		TypeProps{
			Name: "Config",
			Doc:  "skipped",
		},
		`
---
`),
	Entry(
		"type with example",
		TypeProps{
			Name:           "Minimum",
			File:           "core/config.go",
			Line:           15,
			Doc:            "type description",
			ExampleLang:    "yaml",
			ExampleContent: "type: '{{.Thing}}'",
		},
		`## Minimum {{< source name="Minimum" file="core/config.go" line="15" >}} {#minimum-type}
type description

{{< expand "Example" "yaml" >}}
type: '{{.Thing}}'
{{< /expand >}}

---
`),
)

var _ = DescribeTable(
	"gen write field",
	func(field FieldProps, output string) {
		format.TruncatedDiff = false
		writer := strings.Builder{}
		parent := TypeProps{
			Name: "Parent",
		}

		err := writeField(&writer, parent, field)
		Expect(err).To(BeNil())
		Expect(writer.String()).To(Equal(output))
	},
	Entry(
		"bare minimum",
		FieldProps{
			Name: "Minimum",
			Key:  "minimum",
			File: "mini.go",
			Line: 29,
			Doc:  "field description",
		},
		`### minimum {{< source name="Minimum" file="mini.go" line="29" >}} {#parent-minimum}


field description
`),
	Entry(
		"custom type and required",
		FieldProps{
			Name:         "Minimum",
			Key:          "minimum",
			File:         "mini.go",
			Line:         28,
			Doc:          "field description",
			TypeName:     "Animal",
			Required:     true,
			IsCustomType: true,
		},
		`### minimum {{< source name="Minimum" file="mini.go" line="28" >}} {#parent-minimum}
type: [Animal](#animal-type) | required

field description
`),
	Entry(
		"with type name, optional and slice",
		FieldProps{
			Name:     "Minimum",
			Key:      "minimum",
			TypeName: "string",
			File:     "mini.go",
			Line:     27,
			Required: false,
			Slice:    true,
			Doc:      "field description",
		},
		fmt.Sprintf(`### minimum {{< source name="Minimum" file="mini.go" line="27" >}} {#parent-minimum}
type: %v | optional

field description
`, "`[]string`")),
	Entry(
		"map key and value",
		FieldProps{
			Name:             "Minimum",
			Key:              "minimum",
			File:             "mini.go",
			Line:             26,
			Doc:              "field description",
			MapKeyTypeName:   "string",
			MapValueTypeName: "string",
		},
		fmt.Sprintf(`### minimum {{< source name="Minimum" file="mini.go" line="26" >}} {#parent-minimum}
type: map [ %v ] %v | optional

field description
`, "`string`", "`string`")),
	Entry(
		"with template type",
		FieldProps{
			Name:         "Minimum",
			Key:          "minimum",
			Doc:          "field description",
			File:         "mini.go",
			Line:         25,
			TypeName:     "string",
			IsCustomType: false,
			Required:     false,
			TemplateType: "Kitchen",
		},
		fmt.Sprintf(`### minimum {{< source name="Minimum" file="mini.go" line="25" >}} {#parent-minimum}
type: %v | optional | template type: [Kitchen](#kitchen-type)

field description
`, "`string`")),
	Entry(
		"with example",
		FieldProps{
			Name:           "Minimum",
			Key:            "minimum",
			File:           "mini.go",
			Line:           24,
			Doc:            "field description",
			ExampleLang:    "yaml",
			ExampleContent: "format: '{{.Body}}'",
		},
		`### minimum {{< source name="Minimum" file="mini.go" line="24" >}} {#parent-minimum}


field description
{{< expand "Example" "yaml" >}}
format: '{{.Body}}'
{{< /expand >}}

`),
)
