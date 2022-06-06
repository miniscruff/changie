package cmd

import (
	// "os"
	// "path/filepath"

	// "github.com/miniscruff/changie/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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
		// delete expected gen files
		// run gen
		// expect they exist again

		/*
		docsPath := filepath.Join(tempDir, "website", "content", "cli")
		Expect(os.MkdirAll(docsPath, core.CreateDirMode)).To(Succeed())
		rootCmd.SetArgs([]string{"gen"})
		Expect(Execute("")).To(Succeed())

		// test a few command files exist
		Expect(filepath.Join(docsPath, "changie.md")).To(BeARegularFile())
		Expect(filepath.Join(docsPath, "changie_init.md")).To(BeARegularFile())
		Expect(filepath.Join(docsPath, "changie_batch.md")).To(BeARegularFile())
		Expect(filepath.Join(docsPath, "changie_merge.md")).To(BeARegularFile())
		*/
	})
})
