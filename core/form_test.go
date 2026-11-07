package core

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/miniscruff/changie/then"
)

func newPromptsDriver(t *testing.T, p *Prompts) *formDriver {
	t.Helper()

	form, err := p.newForm()
	then.Nil(t, err)

	return newFormDriver(t, form)
}

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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
	d := newPromptsDriver(t, p)

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

func TestFormGlobalCustoms(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Added", Key: "added"},
				{Label: "Dependencies", Key: "deps", SkipBody: true, SkipGlobalChoices: true},
			},
			CustomChoices: []Custom{
				{Key: "Issue", Type: CustomInt},
			},
		},
	}
	d := newPromptsDriver(t, p)

	then.Contains(t, "Issue", d.view())

	d.next() // body
	d.next() // issue
	then.Equals(t, "Issue", d.focused())

	d.typeText("12")
	d.enter()

	then.True(t, d.completed())
	then.MapEquals(t, map[string]string{"Issue": "12"}, p.formCustoms())
}

func TestFormGlobalCustomsSkipped(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Added", Key: "added"},
				{Label: "Dependencies", Key: "deps", SkipBody: true, SkipGlobalChoices: true},
			},
			CustomChoices: []Custom{
				{Key: "Issue", Type: CustomInt},
			},
		},
	}
	d := newPromptsDriver(t, p)

	d.down()
	then.False(t, strings.Contains(d.view(), "Issue"))

	d.enter()

	then.True(t, d.completed())
	then.MapLen(t, 0, p.formCustoms())
}

func TestFormKindAdditionalChoices(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{
					Label: "Added",
					Key:   "added",
					AdditionalChoices: []Custom{
						{Key: "Feature", Label: "Feature Flag", Type: CustomString},
					},
				},
				{
					Label: "Security",
					Key:   "security",
					AdditionalChoices: []Custom{
						{Key: "CVE", Label: "CVE Number", Type: CustomString},
					},
				},
			},
		},
	}
	d := newPromptsDriver(t, p)

	then.Contains(t, "Feature Flag", d.view())
	then.False(t, strings.Contains(d.view(), "CVE Number"))

	d.down()
	then.False(t, strings.Contains(d.view(), "Feature Flag"))
	then.Contains(t, "CVE Number", d.view())

	d.next() // body
	d.next() // cve
	then.Equals(t, "CVE", d.focused())

	d.typeText("2026-0001")
	d.enter()

	then.True(t, d.completed())
	then.MapEquals(t, map[string]string{"CVE": "2026-0001"}, p.formCustoms())
}

func TestFormKindAdditionalChoicesSharedKey(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{
					Label:             "Added",
					Key:               "added",
					AdditionalChoices: []Custom{{Key: "Ticket", Type: CustomString}},
				},
				{
					Label:             "Fixed",
					Key:               "fixed",
					AdditionalChoices: []Custom{{Key: "Ticket", Type: CustomString}},
				},
			},
		},
	}
	d := newPromptsDriver(t, p)

	// answer the added ticket, then go back and switch to fixed
	d.next()
	d.next()
	d.typeText("ADD-1")
	d.prev()
	d.prev()
	then.Equals(t, "kind", d.focused())

	d.down()
	d.next()
	d.next()
	d.typeText("FIX-2")
	d.enter()

	then.True(t, d.completed())
	then.MapEquals(t, map[string]string{"Ticket": "FIX-2"}, p.formCustoms())
}

func TestFormCustomsInvalidType(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			CustomChoices: []Custom{
				{Key: "Bad", Type: "bad"},
			},
		},
	}

	_, err := p.newForm()
	then.Err(t, errInvalidPromptType, err)
}

func TestFormSkipsProvidedValues(t *testing.T) {
	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
		},
		Components: []string{"cli", "tests"},
		Kinds: []KindConfig{
			{Label: "Added", Key: "added"},
		},
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
			{Key: "Author", Type: CustomString},
		},
	}

	p := &Prompts{
		Config:    config,
		Projects:  []string{"client"},
		Component: "cli",
		Kind:      "added",
		Body:      "body from flag",
		Customs:   map[string]string{"Issue": "15"},
	}

	then.SliceLen(t, 0, p.projectFields())
	then.SliceLen(t, 0, p.componentFields())
	then.SliceLen(t, 0, p.kindFields())
	then.SliceLen(t, 0, p.bodyFields())

	d := newPromptsDriver(t, p)

	// only the custom value not provided is asked
	then.Equals(t, "Author", d.focused())

	d.typeText("miniscruff")
	d.enter()

	then.True(t, d.completed())
	then.MapEquals(t, map[string]string{"Author": "miniscruff"}, p.formCustoms())
}

func TestFormProvidedKindHidesFields(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Added", Key: "added"},
				{
					Label:             "Dependencies",
					Key:               "deps",
					SkipBody:          true,
					SkipGlobalChoices: true,
					AdditionalChoices: []Custom{{Key: "Package", Type: CustomString}},
				},
			},
			CustomChoices: []Custom{
				{Key: "Issue", Type: CustomInt},
			},
		},
		Kind: "deps",
	}
	d := newPromptsDriver(t, p)

	then.Equals(t, "Package", d.focused())
	then.False(t, strings.Contains(d.view(), "Body"))
	then.False(t, strings.Contains(d.view(), "Issue"))

	d.typeText("huh")
	d.enter()

	then.True(t, d.completed())
	then.MapEquals(t, map[string]string{"Package": "huh"}, p.formCustoms())
}

func TestFormNothingToAsk(t *testing.T) {
	p := &Prompts{
		Config: &Config{},
		Body:   "body from flag",
	}

	form, err := p.newForm()
	then.Nil(t, err)
	then.Equals(t, nil, form)
}

func TestFormNothingVisibleToAsk(t *testing.T) {
	p := &Prompts{
		Config: &Config{
			Kinds: []KindConfig{
				{Label: "Dependencies", Key: "deps", SkipBody: true},
			},
		},
		Kind: "deps",
	}

	form, err := p.newForm()
	then.Nil(t, err)
	then.Equals(t, nil, form)
}
