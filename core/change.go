package core

import (
	"errors"
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

var (
	errInvalidKind                        = errors.New("invalid kind")
	errInvalidComponent                   = errors.New("invalid component")
	errKindDoesNotAcceptBody              = errors.New("kind does not accept a body")
	errKindProvidedWhenNotConfigured      = errors.New("kind provided but not supported")
	errComponentProvidedWhenNotConfigured = errors.New("component provided but not supported")
)

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

type PromptContext struct {
	config      Config
	stdinReader io.ReadCloser
	kind        *KindConfig
}

// AskPrompts will ask the user prompts based on the configuration
// updating the change as prompts are answered.
func (change *Change) AskPrompts(config Config, stdinReader io.ReadCloser) error {
	ctx := PromptContext{
		config:      config,
		stdinReader: stdinReader,
		kind:        nil,
	}

	err := change.validateArguments(config)
	if err != nil {
		return err
	}

	err = change.promptForComponent(&ctx)
	if err != nil {
		return err
	}

	err = change.promptForKind(&ctx)
	if err != nil {
		return err
	}

	err = change.promptForBody(&ctx)
	if err != nil {
		return err
	}

	return change.promptForUserChoices(&ctx)
}

// validateArguments will check the initial state of a change against the config
// and return an error if anything is invalid
func (change *Change) validateArguments(config Config) error {
	if len(config.Components) == 0 && len(change.Component) > 0 {
		return errComponentProvidedWhenNotConfigured
	}

	if len(config.Kinds) == 0 && len(change.Kind) > 0 {
		return errKindProvidedWhenNotConfigured
	}

	return nil
}

func (change *Change) promptForComponent(ctx *PromptContext) error {
	if len(ctx.config.Components) == 0 {
		return nil
	}

	if len(change.Component) == 0 {
		compPrompt := promptui.Select{
			Label: "Component",
			Items: ctx.config.Components,
			Stdin: ctx.stdinReader,
		}

		var err error

		_, change.Component, err = compPrompt.Run()
		if err != nil {
			return err
		}
	}

	for _, comp := range ctx.config.Components {
		if comp == change.Component {
			return nil
		}
	}

	return fmt.Errorf("%w: %s", errInvalidComponent, change.Component)
}

// promptForKind will prompt for a kind if not provided
// and validate and the kind exists and save the kind config for later use
func (change *Change) promptForKind(ctx *PromptContext) error {
	if len(ctx.config.Kinds) == 0 {
		return nil
	}

	if len(change.Kind) == 0 {
		kindPrompt := promptui.Select{
			Label: "Kind",
			Items: ctx.config.Kinds,
			Stdin: ctx.stdinReader,
		}

		var err error

		_, change.Kind, err = kindPrompt.Run()
		if err != nil {
			return err
		}
	}

	for i := range ctx.config.Kinds {
		kindConfig := &ctx.config.Kinds[i]

		if kindConfig.Label == change.Kind {
			ctx.kind = kindConfig
			return nil
		}
	}

	return fmt.Errorf("%w: %s", errInvalidKind, change.Kind)
}

func (change *Change) promptForBody(ctx *PromptContext) error {
	if ctx.expectsNoBody() && len(change.Body) > 0 {
		return fmt.Errorf("%w: %s", errKindDoesNotAcceptBody, change.Kind)
	}

	if ctx.expectsBody() && len(change.Body) == 0 {
		bodyPrompt := ctx.config.Body.CreatePrompt(ctx.stdinReader)

		var err error
		change.Body, err = bodyPrompt.Run()

		return err
	}

	if ctx.expectsBody() && len(change.Body) > 0 {
		return ctx.config.Body.Validate(change.Body)
	}

	return nil
}

func (ctx *PromptContext) expectsNoBody() bool {
	return ctx.kind != nil && ctx.kind.SkipBody
}

func (ctx *PromptContext) expectsBody() bool {
	return ctx.kind == nil || !ctx.kind.SkipBody
}

func (change *Change) promptForUserChoices(ctx *PromptContext) error {
	change.Custom = make(map[string]string)
	userChoices := make([]Custom, 0)

	if ctx.kind == nil || !ctx.kind.SkipGlobalChoices {
		userChoices = append(userChoices, ctx.config.CustomChoices...)
	}

	if ctx.kind != nil {
		userChoices = append(userChoices, ctx.kind.AdditionalChoices...)
	}

	for _, custom := range userChoices {
		prompt, err := custom.CreatePrompt(ctx.stdinReader)

		if err != nil {
			return err
		}

		change.Custom[custom.Key], err = prompt.Run()

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
