package cmd

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	godoc "go/doc"

	"github.com/miniscruff/changie/then"
)

func TestFilePrependerCreatesFrontmatter(t *testing.T) {
	prepender := filePrepender("my_command.go")
	then.Contains(t, `title: "my command"`, prepender)
	then.Contains(t, `description:`, prepender)
}

func TestLinkHandlerLinksToOtherPages(t *testing.T) {
	link := linkHandler("My_Command.go")
	then.Equals(t, "my_command.md", link)
}

func TestCanGetFiles(t *testing.T) {
	// move our wd to project root instead of cmd dir
	wd, _ := os.Getwd()
	t.Chdir(filepath.Join(wd, ".."))

	cliDocsPath := filepath.Join("docs", "cli")
	configDocsPath := filepath.Join("docs", "config", "index.md")

	// ignore any bad error trying to remove the file, it may not exist and that is ok
	_ = os.Remove(configDocsPath)

	// create cli docs directory if it does not already exist
	err := os.MkdirAll(cliDocsPath, os.ModePerm)
	then.Nil(t, err)

	cliDocsFiles, err := os.ReadDir(cliDocsPath)
	then.Nil(t, err)

	// remove all our generated cli docs before running gen
	for _, fp := range cliDocsFiles {
		if fp.Name() == "_index.md" || fp.Name() == ".gitignore" {
			continue
		}

		err = os.Remove(filepath.Join(cliDocsPath, fp.Name()))
		then.Nil(t, err)
	}

	// run the gen command
	cmd := RootCmd()
	cmd.SetArgs([]string{"gen"})
	then.Nil(t, cmd.Execute())

	// test a few files exist
	then.FileExists(t, cliDocsPath, "changie.md")
	then.FileExists(t, cliDocsPath, "changie_init.md")
	then.FileExists(t, cliDocsPath, "changie_batch.md")
	then.FileExists(t, cliDocsPath, "changie_merge.md")
	then.FileExists(t, configDocsPath)
}

func TestErrorGenBadWriteTypesWriter(t *testing.T) {
	w := then.NewErrWriter()
	parent := TypeProps{
		Name: "Parent",
	}

	w.Raised(t, writeType(w, parent))
}

func TestErrorGenBadWriteTypesWriterWithFields(t *testing.T) {
	w := then.NewErrWriter()
	parent := TypeProps{
		Name: "Config",
		Fields: []FieldProps{{
			Key: "Anything",
		}},
	}

	w.Raised(t, writeType(w, parent))
}

func TestErrorGenWriteFieldBadWriter(t *testing.T) {
	w := then.NewErrWriter()
	parent := TypeProps{
		Name: "Parent",
	}
	field := FieldProps{
		Key: "Key",
	}

	w.Raised(t, writeField(w, parent, field))
}

func TestPanicBuildFieldBadFieldType(t *testing.T) {
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
		Type: &ast.Ellipsis{}, // this is an unknown type for our system
	}

	fset := token.NewFileSet()
	_ = fset.AddFile("file.go", -1, 1000)

	then.Panic(t, func() {
		_ = buildField(fset, field, corePackages, &queue)
	})
}

func TestBuildType(t *testing.T) {
	for _, tc := range []struct {
		name          string
		docType       *godoc.Type
		output        TypeProps
		expectedQueue []string
	}{
		{
			name: "ExampleAndMethod",
			docType: &godoc.Type{
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
			output: TypeProps{
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
		},
		{
			name: "TypeWithMultipleSpecs",
			docType: &godoc.Type{
				Name:    "MyType",
				Doc:     "Type docs",
				Methods: []*godoc.Func{},
				Decl: &ast.GenDecl{
					TokPos: 1,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name:    "RandomInterface",
								NamePos: 10,
							},
							Type: &ast.InterfaceType{},
						},
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name:    "ParentType",
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
			output: TypeProps{
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
						File:           "d.go",
						Line:           1,
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
						File:           "c.go",
						Line:           1,
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
			expectedQueue: []string{"string", "int"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			corePackages := make(CoreTypes)
			// include one package so we can check is custom type
			corePackages["MyOptionalType"] = nil
			queue := make([]string, 0)
			fset := token.NewFileSet()
			_ = fset.AddFile("a.go", -1, 100)
			_ = fset.AddFile("b.go", -1, 100)
			_ = fset.AddFile("c.go", -1, 100)
			_ = fset.AddFile("d.go", -1, 100)

			props := buildType(fset, tc.docType, corePackages, &queue)
			// annoying hack to deeply compare output
			then.True(t, reflect.DeepEqual(tc.output, props))
			then.SliceEquals(t, tc.expectedQueue, queue)
		})
	}
}

func TestBuildMethod(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method *godoc.Func
		output FieldProps
	}{
		{
			name: "Method",
			method: &godoc.Func{
				Name: "MyFunc",
				Decl: &ast.FuncDecl{
					Type: &ast.FuncType{
						Func: 1,
					},
				},
				Doc: "Func docs",
			},
			output: FieldProps{
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
		},
		{
			name: "MethodWithExample",
			method: &godoc.Func{
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
			output: FieldProps{
				Name:           "MyFuncWithExample",
				Key:            "myfuncwithexample",
				TypeName:       "",
				File:           "file.go",
				Line:           1,
				Doc:            "Some func description\n",
				ExampleLang:    "yaml",
				ExampleContent: "doSomething: '{{.Body}}'",
				TemplateType:   "",
				IsCustomType:   false,
				Required:       false,
				Slice:          false,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			_ = fset.AddFile("file.go", -1, 1000)
			props := buildMethod(fset, tc.method)
			then.Equals(t, tc.output, props)
		})
	}
}

func TestGenBuildField(t *testing.T) {
	for _, tc := range []struct {
		name          string
		field         *ast.Field
		output        FieldProps
		expectedQueue []string
	}{
		{
			name: "Identifier",
			field: &ast.Field{
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
			output: FieldProps{
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
			expectedQueue: []string{"MyType"},
		},
		{
			name: "RequiredSliceWithExample",
			field: &ast.Field{
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
			output: FieldProps{
				Name:           "MySliceVar",
				Key:            "mySliceVar",
				TypeName:       "MySliceType",
				File:           "file.go",
				Line:           1,
				Doc:            "This is a slice with an example\n",
				ExampleLang:    "yaml",
				ExampleContent: "mySliceVar: [1, 2, 3]",
				TemplateType:   "",
				IsCustomType:   false,
				Required:       true,
				Slice:          true,
			},
			expectedQueue: []string{"MySliceType"},
		},
		{
			name: "PointerWithTemplateType",
			field: &ast.Field{
				Names: []*ast.Ident{
					{Name: "Optional", NamePos: 1},
				},
				Type: &ast.StarExpr{
					X: &ast.Ident{Name: "MyOptionalType"},
				},
				Tag: &ast.BasicLit{Value: `yaml:"" templateType:"CustomData"`},
			},
			output: FieldProps{
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
			expectedQueue: []string{"MyOptionalType", "CustomData"},
		},
		{
			name: "SelectorExpressions",
			field: &ast.Field{
				Names: []*ast.Ident{
					{Name: "MyVarName", NamePos: 1},
				},
				Type: &ast.SelectorExpr{
					Sel: &ast.Ident{Name: "MySelectorType"},
				},
			},
			output: FieldProps{
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
			expectedQueue: []string{"MySelectorType"},
		},
		{
			name: "MapType",
			field: &ast.Field{
				Names: []*ast.Ident{
					{Name: "MyMap", NamePos: 1},
				},
				Type: &ast.MapType{
					Key:   &ast.Ident{Name: "string"},
					Value: &ast.Ident{Name: "string"},
				},
			},
			output: FieldProps{
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
			expectedQueue: []string{"string", "string"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			corePackages := make(CoreTypes)
			// include one package so we can check is custom type
			corePackages["MyOptionalType"] = nil
			queue := make([]string, 0)
			fset := token.NewFileSet()
			_ = fset.AddFile("file.go", -1, 1000)

			props := buildField(fset, tc.field, corePackages, &queue)
			then.Equals(t, tc.output, props)
			then.SliceEquals(t, tc.expectedQueue, queue)
		})
	}
}
