package cmd

import (
	"errors"
	"fmt"
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
		Skip("temp skip for easy test coverage finding")

		// move our wd to project root instead of cmd dir
		wd, _ := os.Getwd()
		Expect(os.Chdir(filepath.Join(wd, ".."))).To(Succeed())
		defer func() {
			Expect(os.Chdir(wd)).To(Succeed())
		}()

		contentPath := filepath.Join("website", "content")
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
		Skip("temp skip for easy test coverage finding")

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
})

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
			Doc:  "type description",
			Fields: []FieldProps{{
				Name: "MyField",
				Key:  "myField",
				Doc:  "field description",
			}},
		},
		`## Minimum {#minimum-type}
type description
### myField {#minimum-myfield}


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
			Doc:            "type description",
			ExampleLang:    "yaml",
			ExampleContent: "type: '{{.Thing}}'",
		},
		`## Minimum {#minimum-type}
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
			Doc:  "field description",
		},
		`### minimum {#parent-minimum}


field description
`),
	Entry(
		"custom type and required",
		FieldProps{
			Name:         "Minimum",
			Key:          "minimum",
			Doc:          "field description",
			TypeName:     "Animal",
			Required:     true,
			IsCustomType: true,
		},
		`### minimum {#parent-minimum}
type: [Animal](#animal-type) | required

field description
`),
	Entry(
		"with type name optional and slice",
		FieldProps{
			Name:     "Minimum",
			Key:      "minimum",
			TypeName: "string",
			Required: false,
			Slice:    true,
			Doc:      "field description",
		},
		fmt.Sprintf(`### minimum {#parent-minimum}
type: %v | optional

field description
`, "`[]string`")),
	Entry(
		"with template type",
		FieldProps{
			Name:         "Minimum",
			Key:          "minimum",
			Doc:          "field description",
			TypeName:     "string",
			IsCustomType: false,
			Required:     false,
			TemplateType: "Kitchen",
		},
		fmt.Sprintf(`### minimum {#parent-minimum}
type: %v | optional | template type: [Kitchen](#kitchen-type)

field description
`, "`string`")),
	Entry(
		"with example",
		FieldProps{
			Name:           "Minimum",
			Key:            "minimum",
			Doc:            "field description",
			ExampleLang:    "yaml",
			ExampleContent: "format: '{{.Body}}'",
		},
		`### minimum {#parent-minimum}


field description
{{< expand "Example" "yaml" >}}
format: '{{.Body}}'
{{< /expand >}}

`),
)
