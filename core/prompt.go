package core

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"slices"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/multichoose"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	errInvalidKind                        = errors.New("invalid kind")
	errInvalidComponent                   = errors.New("invalid component")
	errKindDoesNotAcceptBody              = errors.New("kind does not accept a body")
	errKindProvidedWhenNotConfigured      = errors.New("kind provided but not supported")
	errComponentProvidedWhenNotConfigured = errors.New("component provided but not supported")
	errCustomProvidedNotConfigured        = errors.New("custom value provided but not configured")
	errProjectNotFound                    = errors.New("project not found")
	errProjectRequired                    = errors.New("project missing but required")

	// prompt disabled
	errProjectMissingPromptDisabled   = errors.New("project missing and prompt is disabled")
	errComponentMissingPromptDisabled = errors.New("component missing and prompt is disabled")
	errKindMissingPromptDisabled      = errors.New("kind missing and prompt is disabled")
	errBodyMissingPromptDisabled      = errors.New("body missing and prompt is disabled")
	errCustomMissingPromptDisabled    = errors.New("custom missing and prompt is disabled")
)

type Prompts struct {
	Cfg              *Config
	StdinReader      io.Reader
	KindOption       *KindOptions
	BodyEditor       bool
	EditorCmdBuilder func(string) (EditorRunner, error)
	TimeNow          TimeNow

	// Enabled checks to make sure our terminal supports prompts
	Enabled bool

	// Values can be submitted from the environment or shell arguments.
	Projects  []string
	Component string
	Kind      string
	Body      string
	Customs   map[string]string
}

// BuildChanges will ask the user prompts based on the configuration
// returning all changes as prompts are answered.
// A change relates to a single project and single change, so if the change
// affects multiple projects we will return multiple changes.
func (p *Prompts) BuildChanges() ([]*Change, error) {
	err := p.validateArguments()
	if err != nil {
		return nil, err
	}

	err = p.projects()
	if err != nil {
		return nil, err
	}

	err = p.component()
	if err != nil {
		return nil, err
	}

	err = p.kind()
	if err != nil {
		return nil, err
	}

	err = p.body()
	if err != nil {
		return nil, err
	}

	err = p.userChoices()
	if err != nil {
		return nil, err
	}

	envs := p.Cfg.EnvVars()

	// If we don't have projects enabled, just create a single change.
	if len(p.Projects) == 0 {
		change := &Change{
			Component: p.Component,
			Kind:      p.Kind,
			Body:      p.Body,
			Time:      p.TimeNow(),
			Custom:    p.Customs,
			Env:       envs,
		}

		err = change.PostProcess(p.Cfg, p.KindOption)
		if err != nil {
			return nil, err
		}

		return []*Change{change}, nil
	}

	changes := make([]*Change, len(p.Projects))
	for i := range changes {
		// Err is already validated when getting the project above.
		projConfig, _ := p.Cfg.ProjectByName(p.Projects[i])
		changes[i] = &Change{
			Project:   projConfig.Key,
			Component: p.Component,
			Kind:      p.Kind,
			Body:      p.Body,
			Time:      p.TimeNow(),
			Custom:    p.Customs,
			Env:       envs,
		}

		err = changes[i].PostProcess(p.Cfg, p.KindOption)
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

// validateArguments will check the initial state of a change against the config
// and return an error if anything is invalid
func (p *Prompts) validateArguments() error {
	if len(p.Cfg.Component.Options) == 0 && len(p.Component) > 0 {
		return errComponentProvidedWhenNotConfigured
	}

	if len(p.Cfg.Kind.Options) == 0 && len(p.Kind) > 0 {
		return errKindProvidedWhenNotConfigured
	}

	configuredCustoms := make([]Custom, 0)

	if len(p.Cfg.Kind.Options) > 0 && len(p.Kind) > 0 {
		kc := p.Cfg.KindFromKeyOrLabel(p.Kind)
		if kc == nil {
			return fmt.Errorf("%w: %s", errInvalidKind, p.Kind)
		}

		// TODO: support this feature
		// configuredCustoms = append(configuredCustoms, kc.AdditionalChoices...)
		// TODO: support this feature
		/*
			if !kc.SkipGlobalChoices {
				configuredCustoms = append(configuredCustoms, p.Cfg.CustomChoices...)
			}
		*/
	} else {
		// todo: support this feature
		// configuredCustoms = append(configuredCustoms, p.Cfg.CustomChoices...)
	}

	// make sure no custom values are assigned that do not exist
	foundCustoms := map[string]struct{}{}

	for key, value := range p.Customs {
		for _, choice := range configuredCustoms {
			if choice.Key == key {
				foundCustoms[key] = struct{}{}

				err := choice.Validate(value)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	for key := range p.Customs {
		_, ok := foundCustoms[key]
		if !ok {
			return fmt.Errorf("%w: %s", errCustomProvidedNotConfigured, key)
		}
	}

	return nil
}

func (p *Prompts) projects() error {
	if len(p.Cfg.Project.Options) == 0 {
		return nil
	}

	if len(p.Projects) == 0 {
		if !p.Enabled {
			return errProjectMissingPromptDisabled
		}

		var err error

		projs, err := prompt.New().Ask("Projects").
			MultiChoose(
				p.Cfg.ProjectLabels(),
				multichoose.WithHelp(true),
				multichoose.WithTeaProgramOpts(tea.WithInput(p.StdinReader)),
			)
		if err != nil {
			return err
		}

		if len(projs) == 0 {
			return errProjectRequired
		}

		p.Projects = projs
	}

	// Quickly validate the project exists and lines up before moving on.
	for _, proj := range p.Projects {
		_, err := p.Cfg.ProjectByName(proj)
		if err != nil {
			return fmt.Errorf("%w: %s", err, proj)
		}
	}

	return nil
}

func (p *Prompts) component() error {
	if len(p.Cfg.Component.Options) == 0 {
		return nil
	}

	if len(p.Component) == 0 {
		if !p.Enabled {
			return errComponentMissingPromptDisabled
		}

		var err error

		comp, err := Custom{
			Type:        CustomEnum,
			Label:       "Component",
			EnumOptions: p.Cfg.Component.Names(),
		}.AskPrompt(p.StdinReader)
		if err != nil {
			return err
		}

		p.Component = comp
	}

	if !slices.Contains(p.Cfg.Component.Names(), p.Component) {
		return fmt.Errorf("%w: %s", errInvalidComponent, p.Component)
	}

	return nil
}

func (p *Prompts) kind() error {
	if len(p.Cfg.Kind.Options) == 0 {
		return nil
	}

	if len(p.Kind) == 0 {
		if !p.Enabled {
			return errKindMissingPromptDisabled
		}

		kindLabels := make([]string, len(p.Cfg.Kind.Options))
		for i, kc := range p.Cfg.Kind.Options {
			kindLabels[i] = kc.Prompt.KeyOrLabel()
		}

		kind, err := Custom{
			Type:        CustomEnum,
			Label:       "Kind",
			EnumOptions: kindLabels,
		}.AskPrompt(p.StdinReader)
		if err != nil {
			return err
		}

		p.Kind = kind
	}

	for i := range p.Cfg.Kind.Options {
		kindOption := &p.Cfg.Kind.Options[i]

		if kindOption.Prompt.Equals(p.Kind) {
			p.KindOption = kindOption
			p.Kind = kindOption.Prompt.KeyOrLabel()

			return nil
		}
	}

	return fmt.Errorf("%w: %s", errInvalidKind, p.Kind)
}

func (p *Prompts) body() error {
	if p.expectsNoBody() && len(p.Body) > 0 {
		return fmt.Errorf("%w: %s", errKindDoesNotAcceptBody, p.Kind)
	}

	if p.expectsBody() && len(p.Body) == 0 {
		if !p.Enabled {
			return errBodyMissingPromptDisabled
		}

		if p.BodyEditor {
			file, err := createTempFile(runtime.GOOS, p.Cfg.ReleaseNotes.Extension)
			if err != nil {
				return err
			}

			runner, err := p.EditorCmdBuilder(file)
			if err != nil {
				return err
			}

			p.Body, err = getBodyTextWithEditor(runner, file)

			return err
		} else {
			bodyCustom := p.Cfg.Change.CreateCustom()

			body, err := bodyCustom.AskPrompt(p.StdinReader)
			if err != nil {
				return err
			}

			p.Body = body

			return err
		}
	}

	if p.expectsBody() && len(p.Body) > 0 {
		return p.Cfg.Change.Validate(p.Body)
	}

	return nil
}

func (p *Prompts) expectsNoBody() bool {
	return p.KindOption != nil
	// TODO: SUPPORT
	// && p.KindOption.SkipBody
}

func (p *Prompts) expectsBody() bool {
	return p.KindOption == nil
	// TODO: SUPPORT
	// || !p.KindOption.SkipBody
}

func (p *Prompts) userChoices() error {
	userChoices := make([]Custom, 0)

	// TODO: support
	/*
	if p.KindOption == nil || !p.KindOption.SkipGlobalChoices {
		userChoices = append(userChoices, p.Cfg.CustomChoices...)
	}
	*/

	// TODO: support
	/*
	if p.KindOption != nil {
		userChoices = append(userChoices, p.KindOption.AdditionalChoices...)
	}
	*/

	// custom map may be nil, which is fine if we have no choices
	// otherwise we need to initialize it
	if len(userChoices) > 0 && p.Customs == nil {
		p.Customs = make(map[string]string)
	}

	for _, custom := range userChoices {
		// skip already provided values
		if p.Customs[custom.Key] != "" {
			continue
		}

		if !p.Enabled {
			return fmt.Errorf("%w: custom key '%s'", errCustomMissingPromptDisabled, custom.Key)
		}

		var err error

		p.Customs[custom.Key], err = custom.AskPrompt(p.StdinReader)
		if err != nil {
			return err
		}
	}

	return nil
}
