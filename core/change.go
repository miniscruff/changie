package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

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
	errCustomProvidedNotConfigured        = errors.New("custom value provided but not configured")
	errProjectNotFound                    = errors.New("project not found")
	errProjectRequired                    = errors.New("project missing but required")
)

func SortByConfig(config *Config) *ChangesConfigSorter {
	return &ChangesConfigSorter{
		config: *config,
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
	// Project of our change, if one was provided.
	Project string `yaml:",omitempty" default:""`
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
	// Env vars configured by the system.
	// This is not written in change fragments but instead loaded by the system and accessible for templates.
	// For example if you want to use an env var in [change format](#config-changeformat) you can,
	// but env vars configured when executing `changie new` will not be saved.
	// See [envPrefix](#config-envprefix) for configuration.
	Env map[string]string `yaml:"-" default:"nil"`
	// Filename the change was saved to.
	Filename string `yaml:"-"`
}

// WriteTo will write a change to the writer as YAML
func (change Change) WriteTo(writer io.Writer) (int64, error) {
	bs, _ := yaml.Marshal(&change)
	n, err := writer.Write(bs)

	return int64(n), err
}

type PromptContext struct {
	Config           *Config
	StdinReader      io.Reader
	Kind             *KindConfig
	BodyEditor       bool
	CreateFiler      shared.CreateFiler
	EditorCmdBuilder func(string) (EditorRunner, error)
}

// AskPrompts will ask the user prompts based on the configuration
// updating the change as prompts are answered.
func (change *Change) AskPrompts(ctx PromptContext) error {
	err := change.validateArguments(ctx.Config)
	if err != nil {
		return err
	}

	err = change.promptForProject(&ctx)
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

	err = change.promptForUserChoices(&ctx)
	if err != nil {
		return err
	}

	return change.postProcess(&ctx)
}

// validateArguments will check the initial state of a change against the config
// and return an error if anything is invalid
func (change *Change) validateArguments(config *Config) error {
	if len(config.Components) == 0 && len(change.Component) > 0 {
		return errComponentProvidedWhenNotConfigured
	}

	if len(config.Kinds) == 0 && len(change.Kind) > 0 {
		return errKindProvidedWhenNotConfigured
	}

	// make sure no custom values are assigned that do not exist
	foundCustoms := map[string]struct{}{}

	for key, value := range change.Custom {
		for _, choice := range config.CustomChoices {
			if choice.Key == key {
				foundCustoms[key] = struct{}{}

				err := choice.Validate(value)
				if err != nil {
					return err
				}
			}
		}
	}

	for key := range change.Custom {
		_, ok := foundCustoms[key]
		if !ok {
			return fmt.Errorf("%w: %s", errCustomProvidedNotConfigured, key)
		}
	}

	return nil
}

func (change *Change) promptForProject(ctx *PromptContext) error {
	if len(ctx.Config.Projects) == 0 {
		return nil
	}

	if len(change.Project) == 0 {
		fmt.Println(ProjectsWarning)

		var err error

		change.Project, err = Custom{
			Type:        CustomEnum,
			Label:       "Project",
			EnumOptions: ctx.Config.ProjectLabels(),
		}.AskPrompt(ctx.StdinReader)
		if err != nil {
			return err
		}
	}

	pc, err := ctx.Config.Project(change.Project)
	if err != nil {
		return fmt.Errorf("%w: %s", err, change.Project)
	}

	change.Project = pc.Key

	return nil
}

func (change *Change) promptForComponent(ctx *PromptContext) error {
	if len(ctx.Config.Components) == 0 {
		return nil
	}

	if len(change.Component) == 0 {
		var err error

		change.Component, err = Custom{
			Type:        CustomEnum,
			Label:       "Component",
			EnumOptions: ctx.Config.Components,
		}.AskPrompt(ctx.StdinReader)
		if err != nil {
			return err
		}
	}

	for _, comp := range ctx.Config.Components {
		if comp == change.Component {
			return nil
		}
	}

	return fmt.Errorf("%w: %s", errInvalidComponent, change.Component)
}

// promptForKind will prompt for a kind if not provided
// and validate and the kind exists and save the kind config for later use
func (change *Change) promptForKind(ctx *PromptContext) error {
	if len(ctx.Config.Kinds) == 0 {
		return nil
	}

	if len(change.Kind) == 0 {
		kindLabels := make([]string, len(ctx.Config.Kinds))
		for i, kc := range ctx.Config.Kinds {
			kindLabels[i] = kc.Label
		}

		var err error
		change.Kind, err = Custom{
			Type:        CustomEnum,
			Label:       "Kind",
			EnumOptions: kindLabels,
		}.AskPrompt(ctx.StdinReader)

		if err != nil {
			return err
		}
	}

	for i := range ctx.Config.Kinds {
		kindConfig := &ctx.Config.Kinds[i]

		if kindConfig.Label == change.Kind {
			ctx.Kind = kindConfig
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
		if ctx.BodyEditor {
			file, err := createTempFile(ctx.CreateFiler, runtime.GOOS, ctx.Config.VersionExt)
			if err != nil {
				return err
			}

			runner, err := ctx.EditorCmdBuilder(file)
			if err != nil {
				return err
			}

			change.Body, err = getBodyTextWithEditor(runner, file, os.ReadFile)

			return err
		} else {
			bodyCustom := ctx.Config.Body.CreateCustom()

			var err error
			change.Body, err = bodyCustom.AskPrompt(ctx.StdinReader)

			return err
		}
	}

	if ctx.expectsBody() && len(change.Body) > 0 {
		return ctx.Config.Body.Validate(change.Body)
	}

	return nil
}

func (ctx *PromptContext) expectsNoBody() bool {
	return ctx.Kind != nil && ctx.Kind.SkipBody
}

func (ctx *PromptContext) expectsBody() bool {
	return ctx.Kind == nil || !ctx.Kind.SkipBody
}

func (change *Change) promptForUserChoices(ctx *PromptContext) error {
	userChoices := make([]Custom, 0)

	if ctx.Kind == nil || !ctx.Kind.SkipGlobalChoices {
		userChoices = append(userChoices, ctx.Config.CustomChoices...)
	}

	if ctx.Kind != nil {
		userChoices = append(userChoices, ctx.Kind.AdditionalChoices...)
	}

	// custom map may be nil, which is fine if we have no choices
	// otherwise we need to initialize it
	if len(userChoices) > 0 && change.Custom == nil {
		change.Custom = make(map[string]string)
	}

	for _, custom := range userChoices {
		// skip already provided values
		if change.Custom[custom.Key] != "" {
			continue
		}

		var err error

		change.Custom[custom.Key], err = custom.AskPrompt(ctx.StdinReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func (change *Change) postProcess(ctx *PromptContext) error {
	postConfigs := make([]PostProcessConfig, 0)

	if ctx.Kind == nil || !ctx.Kind.SkipGlobalPost {
		postConfigs = append(postConfigs, ctx.Config.Post...)
	}

	if ctx.Kind != nil {
		postConfigs = append(postConfigs, ctx.Kind.Post...)
	}

	if len(postConfigs) == 0 {
		return nil
	}

	if change.Custom == nil {
		change.Custom = make(map[string]string)
	}

	templateCache := NewTemplateCache()

	for _, post := range postConfigs {
		writer := strings.Builder{}

		err := templateCache.Execute(post.Value, &writer, change)
		if err != nil {
			return err
		}

		change.Custom[post.Key] = writer.String()
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

	c.Filename = path

	return c, nil
}
