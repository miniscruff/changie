package core

import (
	"fmt"
	"strings"
	"testing"

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

	then.True(t, d.completed())
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

	then.False(t, d.completed())
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

	then.True(t, d.completed())
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

	then.True(t, d.completed())
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
	d.enter()

	then.True(t, d.completed())
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
