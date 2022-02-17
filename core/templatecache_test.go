package core

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template Cache", func() {
	var (
		cache   *TemplateCache
		builder *strings.Builder
	)

	BeforeEach(func() {
		cache = NewTemplateCache()
		builder = &strings.Builder{}
	})

	It("returns error on bad template", func() {
		_, err := cache.Load("{{...___..__}}")
		Expect(err).NotTo(BeNil())
	})

	It("caches a template for reuse", func() {
		first, err := cache.Load("### {{.Kind}}")
		Expect(err).To(BeNil())
		second, err := cache.Load("### {{.Kind}}")
		Expect(err).To(BeNil())
		// use == to compare pointers
		Expect(first == second).To(BeTrue())
	})

	It("returns error on bad execute", func() {
		err := cache.Execute("bad template {{...__.}}", builder, struct{}{})

		Expect(err).NotTo(BeNil())
		Expect(builder.String()).To(BeEmpty())
	})

	It("can execute template directly", func() {
		data := map[string]string{
			"A": "apple",
		}

		err := cache.Execute("A: {{.A}}", builder, data)

		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("A: apple"))
	})

	It("can use sprig functions", func() {
		err := cache.Execute("{{ \"hello!\" | upper | repeat 5 }}", builder, nil)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("HELLO!HELLO!HELLO!HELLO!HELLO!"))
	})

	It("can count strings", func() {
		data := BatchData{
			Changes: []Change{
				{Kind: "added", Body: "A"},
				{Kind: "added", Body: "B"},
				{Kind: "removed", Body: "C"},
			},
		}
		err := cache.Execute("{{ kinds .Changes | count \"added\" }} kinds", builder, data)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("2 kinds"))
	})

	It("can get kinds", func() {
		data := BatchData{
			Changes: []Change{
				{Kind: "added", Body: "A"},
				{Kind: "added", Body: "B"},
				{Kind: "removed", Body: "C"},
			},
		}
		err := cache.Execute("{{ kinds .Changes }} kinds", builder, data)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("[added added removed] kinds"))
	})

	It("can get components", func() {
		data := BatchData{
			Changes: []Change{
				{Component: "ui", Body: "A"},
				{Component: "ui", Body: "B"},
				{Component: "backend", Body: "C"},
			},
		}
		err := cache.Execute("{{ components .Changes }} components", builder, data)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("[ui ui backend] components"))
	})

	It("can get bodies", func() {
		data := BatchData{
			Changes: []Change{
				{Component: "ui", Body: "A"},
				{Component: "ui", Body: "B"},
				{Component: "backend", Body: "C"},
			},
		}
		err := cache.Execute("{{ bodies .Changes }} bodies", builder, data)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("[A B C] bodies"))
	})

	It("can get times", func() {
		data := BatchData{
			Changes: []Change{
				{Time: time.Date(2022, time.January, 1, 9, 30, 0, 0, time.UTC)},
				{Time: time.Date(2022, time.February, 1, 10, 30, 0, 0, time.UTC)},
				{Time: time.Date(2022, time.March, 1, 11, 30, 0, 0, time.UTC)},
			},
		}
		err := cache.Execute("{{ times .Changes }} times", builder, data)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal(
			fmt.Sprintf(
				"[%v %v %v] times",
				"2022-01-01 09:30:00 +0000 UTC",
				"2022-02-01 10:30:00 +0000 UTC",
				"2022-03-01 11:30:00 +0000 UTC",
			),
		))
	})

	It("can get customs", func() {
		data := BatchData{
			Changes: []Change{
				{Custom: map[string]string{"Author": "miniscruff"}},
				{Custom: map[string]string{"Author": "rocktavious"}},
				{Custom: map[string]string{"Author": "emmyoop"}},
			},
		}
		err := cache.Execute("{{ customs .Changes \"Author\" }} authors", builder, data)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("[miniscruff rocktavious emmyoop] authors"))
	})

	It("returns error on missing key", func() {
		data := BatchData{
			Changes: []Change{
				{Custom: map[string]string{"Author": "miniscruff"}},
			},
		}
		err := cache.Execute("{{ customs .Changes \"MissingKey\" }}", builder, data)
		Expect(err).NotTo(BeNil())
		Expect(builder.String()).To(BeEmpty())
	})
})
