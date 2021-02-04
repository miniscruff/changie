package cmd

import (
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
})
