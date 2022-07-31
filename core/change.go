package core

import (
	"fmt"
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

type ErrInvalidKind struct {
	message string
}

func (e ErrInvalidKind) Error() string {
	return fmt.Sprintf("ErrInvalidKind: %s", e.message)
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
	}

	return p.Prompt()
}

type Prompter struct {
	change      *Change
	config      Config
	stdinReader io.ReadCloser
	kind        *KindConfig
}

func (p *Prompter) Prompt() error {
	err := p.promptForComponent()

	if err != nil {
		return err
	}

	err = p.promptForKind()

	if err != nil {
		return err
	}

	err = p.parseKind()

	if err != nil {
		return err
	}

	err = p.promptForBody()

	if err != nil {
		return err
	}

	return p.promptForUserChoices()
}

func (p *Prompter) promptForComponent() error {
	if len(p.config.Components) == 0 {
		return nil
	}

	compPrompt := promptui.Select{
		Label: "Component",
		Items: p.config.Components,
		Stdin: p.stdinReader,
	}

	var err error
	_, p.change.Component, err = compPrompt.Run()

	return err
}

func (p *Prompter) promptForKind() error {
	if len(p.config.Kinds) == 0 || len(p.change.Kind) > 0 {
		return nil
	}

	kindPrompt := promptui.Select{
		Label: "Kind",
		Items: p.config.Kinds,
		Stdin: p.stdinReader,
	}

	var err error
	_, p.change.Kind, err = kindPrompt.Run()

	return err
}

func (p *Prompter) parseKind() error {
	if len(p.change.Kind) == 0 {
		return nil
	}

	for i := range p.config.Kinds {
		kindConfig := &p.config.Kinds[i]

		if kindConfig.Label == p.change.Kind {
			p.kind = kindConfig
			return nil
		}
	}

	return ErrInvalidKind{p.change.Kind}
}

func (p *Prompter) promptForBody() error {
	if (p.kind == nil || !p.kind.SkipBody) && len(p.change.Body) == 0 {
		bodyPrompt := p.config.Body.CreatePrompt(p.stdinReader)

		var err error
		p.change.Body, err = bodyPrompt.Run()

		return err
	}

	return nil
}

func (p *Prompter) promptForUserChoices() error {
	p.change.Custom = make(map[string]string)
	userChoices := make([]Custom, 0)

	if p.kind == nil || !p.kind.SkipGlobalChoices {
		userChoices = append(userChoices, p.config.CustomChoices...)
	}

	if p.kind != nil {
		userChoices = append(userChoices, p.kind.AdditionalChoices...)
	}

	for _, custom := range userChoices {
		prompt, err := custom.CreatePrompt(p.stdinReader)

		if err != nil {
			return err
		}

		p.change.Custom[custom.Key], err = prompt.Run()

		if err != nil {
			return err
		}
	}

	return nil
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
