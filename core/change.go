package core

import (
	"errors"
	"io"
	"sort"
	"time"

	"github.com/manifoldco/promptui"

	"gopkg.in/yaml.v2"

	"github.com/miniscruff/changie/shared"
)

type ChangesConfigSorter struct {
	changes []Change
	config  Config
}

func SortByConfig(config Config) *ChangesConfigSorter {
	return &ChangesConfigSorter{
		config: config,
	}
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (s *ChangesConfigSorter) Sort(changes []Change) {
	s.changes = changes
	sort.Sort(s)
}

// Len is part of sort.Interface.
func (s *ChangesConfigSorter) Len() int {
	return len(s.changes)
}

// Swap is part of sort.Interface.
func (s *ChangesConfigSorter) Swap(i, j int) {
	s.changes[i], s.changes[j] = s.changes[j], s.changes[i]
}

// Less will compare two Change values with the config settings.
// * Components, if enabled, are sorted by index in config
// * Kind, if enabled, are sorted by index in config
// * Time sorted newest first
func (s *ChangesConfigSorter) Less(i, j int) bool {
	a, b := &s.changes[i], &s.changes[j]

	// Start by sorting by component index
	if len(s.config.Components) > 0 && a.Component != b.Component {
		for _, c := range s.config.Components {
			if a.Component == c {
				return true
			} else if b.Component == c {
				return false
			}
		}
	}

	// Then sort by kind index
	if len(s.config.Kinds) > 0 && a.Kind != b.Kind {
		for _, k := range s.config.Kinds {
			if a.Kind == k.Label {
				return true
			} else if b.Kind == k.Label {
				return false
			}
		}
	}

	// Finish sort by oldest first
	return b.Time.After(a.Time)
}

// Change represents an atomic change to a project.
type Change struct {
	// Component of our change, if one was provided.
	Component string `yaml:",omitempty" default:""`
	// Kind of our change, if one was provided.
	Kind string `yaml:",omitempty" default:""`
	// Body message of our change, if one was provided.
	Body string `yaml:",omitempty" default:""`
	// When our change was made.
	Time time.Time `yaml:"" required:"true"`
	// Custom values corresponding to our options where each key-value pair is the key of the custom option
	// and value the one provided in the change.
	// example: yaml
	// custom:
	// - key: Issue
	//   type: int
	// changeFormat: "{{.Body}} from #{{.Custom.Issue}}"
	Custom map[string]string `yaml:",omitempty" default:"nil"`
}

// SaveUnreleased will save an unreleased change to the unreleased directory
func (change Change) Write(writer io.Writer) error {
	bs, _ := yaml.Marshal(&change)
	_, err := writer.Write(bs)

	return err
}

// AskPrompts will ask the user prompts based on the configuration
// updating the change as prompts are answered.
func (change *Change) AskPrompts(config Config, stdinReader io.ReadCloser) error {
	p := Prompter{
		change:      change,
		config:      config,
		stdinReader: stdinReader,
		kind:        nil,
		err:         nil,
	}

	return p.Prompt()
}

type Prompter struct {
	change      *Change
	config      Config
	stdinReader io.ReadCloser
	kind        *KindConfig
	err         error
}

func (p *Prompter) Prompt() error {
	p.promptForComponent()
	p.promptForKind()
	p.parseKind()
	p.promptForBody()
	p.promptForUserChoices()

	return p.err
}

func (p *Prompter) promptForComponent() {
	if len(p.config.Components) == 0 {
		return
	}

	compPrompt := promptui.Select{
		Label: "Component",
		Items: p.config.Components,
		Stdin: p.stdinReader,
	}

	_, p.change.Component, p.err = compPrompt.Run()
}

func (p *Prompter) promptForKind() {
	if p.err != nil {
		return
	}

	if len(p.config.Kinds) == 0 || len(p.change.Kind) > 0 {
		return
	}

	kindPrompt := promptui.Select{
		Label: "Kind",
		Items: p.config.Kinds,
		Stdin: p.stdinReader,
	}

	_, p.change.Kind, p.err = kindPrompt.Run()
}

func (p *Prompter) parseKind() {
	if p.err != nil {
		return
	}

	if len(p.change.Kind) == 0 {
		return
	}

	for i := range p.config.Kinds {
		kindConfig := &p.config.Kinds[i]

		if kindConfig.Label == p.change.Kind {
			p.kind = kindConfig
			return
		}
	}

	p.err = errors.New("invalid kind")
}

func (p *Prompter) promptForBody() {
	if p.err != nil {
		return
	}

	if (p.kind == nil || !p.kind.SkipBody) && len(p.change.Body) == 0 {
		bodyPrompt := p.config.Body.CreatePrompt(p.stdinReader)
		p.change.Body, p.err = bodyPrompt.Run()
	}
}

func (p *Prompter) promptForUserChoices() {
	if p.err != nil {
		return
	}

	p.change.Custom = make(map[string]string)
	userChoices := make([]Custom, 0)

	if p.kind == nil || !p.kind.SkipGlobalChoices {
		userChoices = append(userChoices, p.config.CustomChoices...)
	}

	if p.kind != nil {
		userChoices = append(userChoices, p.kind.AdditionalChoices...)
	}

	for _, custom := range userChoices {
		var prompt Prompt
		prompt, p.err = custom.CreatePrompt(p.stdinReader)

		if p.err != nil {
			return
		}

		p.change.Custom[custom.Key], p.err = prompt.Run()

		if p.err != nil {
			return
		}
	}
}

// LoadChange will load a change from file path
func LoadChange(path string, rf shared.ReadFiler) (Change, error) {
	var c Change

	bs, err := rf(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
