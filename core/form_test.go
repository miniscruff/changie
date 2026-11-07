package core

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/miniscruff/changie/then"
)

func TestFormProjects(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Projects: []ProjectConfig{
				{Label: "Client", Key: "client"},
				{Label: "Server", Key: "server"},
				{Key: "shared"},
			},
		},
	}
	d := newFormDriver(t, p.newForm())

	then.Contains(t, "Projects", d.view())
	then.Contains(t, "Client", d.view())
	// projects without a label display their key
	then.Contains(t, "shared", d.view())

	// select client and shared
	d.press(' ')
	d.down()
	d.down()
	d.press(' ')
	d.enter()

	then.Equals(t, "body", d.focused())
	then.SliceEquals(t, []string{"client", "shared"}, p.Projects)
}

func TestFormProjectsRequired(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Projects: []ProjectConfig{
				{Label: "Client", Key: "client"},
			},
		},
	}
	d := newFormDriver(t, p.newForm())

	d.enter()

	then.Equals(t, "projects", d.focused())
	then.Contains(t, errProjectRequired.Error(), d.view())
}

func TestFormProjectsFilter(t *testing.T) {
	projects := make([]ProjectConfig, 100)
	for i := range projects {
		projects[i] = ProjectConfig{Key: fmt.Sprintf("project-%03d", i)}
	}

	p := &Prompts{
		Config: &Config{
			Projects: projects,
		},
	}
	d := newFormDriver(t, p.newForm())

	// long lists scroll instead of showing every project
	then.False(t, strings.Contains(d.view(), "project-099"))

	d.press('/')
	d.typeText("099")
	d.enter() // accept the filter
	d.press(' ')
	d.enter()

	then.Equals(t, "body", d.focused())
	then.SliceEquals(t, []string{"project-099"}, p.Projects)
}

func TestFormNoProjects(t *testing.T) {
	p := &Prompts{
		Config: &Config{},
	}

	then.SliceLen(t, 0, p.projectFields())
}

func TestFormComponent(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Components: []string{"cli", "tests", "utils"},
		},
	}
	d := newFormDriver(t, p.newForm())

	then.Contains(t, "Component", d.view())
	then.Contains(t, "utils", d.view())

	d.down()
	d.enter()

	then.Equals(t, "body", d.focused())
	then.Equals(t, "tests", p.Component)
}

func TestFormComponentFilter(t *testing.T) {
	components := make([]string, 50)
	for i := range components {
		components[i] = fmt.Sprintf("component-%02d", i)
	}

	p := &Prompts{
		Config: &Config{
			Components: components,
		},
	}
	d := newFormDriver(t, p.newForm())

	then.False(t, strings.Contains(d.view(), "component-49"))

	d.press('/')
	d.typeText("49")
	d.enter() // accept the filter

	then.Equals(t, "component-49", p.Component)
}

func TestFormProjectsThenComponent(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Projects: []ProjectConfig{
				{Label: "Client", Key: "client"},
			},
			Components: []string{"cli", "tests"},
		},
	}
	d := newFormDriver(t, p.newForm())

	d.press(' ')
	d.next()
	then.Equals(t, "component", d.focused())
	// previous answers stay visible
	then.Contains(t, "Client", d.view())

	d.prev()
	then.Equals(t, "projects", d.focused())
}

func TestFormNoComponents(t *testing.T) {
	p := &Prompts{
		Config: &Config{},
	}

	then.SliceLen(t, 0, p.componentFields())
}

func TestFormKind(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Added", Key: "added"},
				{Label: "Removed"},
			},
		},
	}
	d := newFormDriver(t, p.newForm())

	then.Contains(t, "Kind", d.view())
	then.Contains(t, "Added", d.view())

	d.down()
	d.enter()

	then.Equals(t, "body", d.focused())
	// kinds without a key use the label
	then.Equals(t, "Removed", p.Kind)
}

func TestFormSelectedKindFollowsCursor(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Added", Key: "added"},
				{Label: "Removed", Key: "removed", SkipBody: true},
			},
		},
	}
	d := newFormDriver(t, p.newForm())

	then.Equals(t, "added", p.selectedKind().Key)

	d.down()
	then.False(t, d.completed())
	then.Equals(t, "removed", p.selectedKind().Key)
	then.True(t, p.selectedKind().SkipBody)
}

func TestFormNoKinds(t *testing.T) {
	p := &Prompts{
		Config: &Config{},
	}

	then.SliceLen(t, 0, p.kindFields())
	then.Equals(t, nil, p.selectedKind())
}

func TestFormBody(t *testing.T) {
	p := &Prompts{
		Config: &Config{},
	}
	d := newFormDriver(t, p.newForm())

	then.Contains(t, "Body", d.view())

	d.typeText("body stuff")
	d.enter()

	then.True(t, d.completed())
	then.Equals(t, "body stuff", p.Body)
}

func TestFormBodyBlock(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Body: BodyConfig{UseBlock: true},
		},
	}
	d := newFormDriver(t, p.newForm())

	d.typeText("first line")
	d.send(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt}) // new line
	d.typeText("second line")
	d.enter()

	then.True(t, d.completed())
	then.Equals(t, "first line\nsecond line", p.Body)
}

func TestFormBodyLength(t *testing.T) {
	minLength := int64(5)
	p := &Prompts{
		Config: &Config{
			Body: BodyConfig{MinLength: &minLength},
		},
	}
	d := newFormDriver(t, p.newForm())

	d.typeText("abc")
	d.enter()

	then.False(t, d.completed())
	then.Contains(t, errInputTooShort.Error(), d.view())

	d.typeText("de")
	d.enter()

	then.True(t, d.completed())
	then.Equals(t, "abcde", p.Body)
}

func TestFormBodyHiddenBySkipBody(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Added", Key: "added"},
				{Label: "Dependencies", Key: "deps", SkipBody: true},
			},
		},
	}
	d := newFormDriver(t, p.newForm())

	then.Contains(t, "Body", d.view())

	d.down()
	then.False(t, strings.Contains(d.view(), "Body"))

	// enter on the kind is the last visible field so it submits the form
	d.enter()
	then.True(t, d.completed())
}

func TestFormBodyShownForKindWithBody(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Dependencies", Key: "deps", SkipBody: true},
				{Label: "Added", Key: "added"},
			},
		},
	}
	d := newFormDriver(t, p.newForm())

	d.down()
	d.next()
	then.Equals(t, "body", d.focused())

	d.prev()
	then.Equals(t, "kind", d.focused())
}

func TestFormBodyWithEditor(t *testing.T) {
	p := &Prompts{
		Config:     &Config{},
		BodyEditor: true,
	}

	then.SliceLen(t, 0, p.bodyFields())
}
