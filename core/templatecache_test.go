package core

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template Cache", func() {
	It("returns error on bad template", func() {
		cache := NewTemplateCache()
		_, err := cache.Load("{{...___..__}}")
		Expect(err).NotTo(BeNil())
	})

	It("caches a template for reuse", func() {
		cache := NewTemplateCache()
		first, err := cache.Load("### {{.Kind}}")
		Expect(err).To(BeNil())
		second, err := cache.Load("### {{.Kind}}")
		Expect(err).To(BeNil())
		// use == to compare pointers
		Expect(first == second).To(BeTrue())
	})

	It("returns error on bad execute", func() {
		cache := NewTemplateCache()
		builder := &strings.Builder{}

		err := cache.Execute("bad template {{...__.}}", builder, struct{}{})

		Expect(err).NotTo(BeNil())
		Expect(builder.String()).To(BeEmpty())
	})

	It("can execute template directly", func() {
		cache := NewTemplateCache()
		builder := &strings.Builder{}
		data := map[string]string{
			"A": "apple",
		}

		err := cache.Execute("A: {{.A}}", builder, data)

		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("A: apple"))
	})
})
