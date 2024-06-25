package core

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/multichoose"
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
)

type Prompts struct {
	Config           *Config
	StdinReader      io.Reader
	KindConfig       *KindConfig
	BodyEditor       bool
	EditorCmdBuilder func(string) (EditorRunner, error)
	TimeNow          TimeNow

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

	envs := p.Config.EnvVars()

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

		err = change.PostProcess(p.Config, p.KindConfig)
		if err != nil {
			return nil, err
		}

		return []*Change{change}, nil
	}

	changes := make([]*Change, len(p.Projects))
	for i := 0; i < len(changes); i++ {
		// Err is already validated when getting the project above.
		projConfig, _ := p.Config.Project(p.Projects[i])
		changes[i] = &Change{
			Project:   projConfig.Key,
			Component: p.Component,
			Kind:      p.Kind,
			Body:      p.Body,
			Time:      p.TimeNow(),
			Custom:    p.Customs,
			Env:       envs,
		}

		err = changes[i].PostProcess(p.Config, p.KindConfig)
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

// validateArguments will check the initial state of a change against the config
// and return an error if anything is invalid
func (p *Prompts) validateArguments() error {
	if len(p.Config.Components) == 0 && len(p.Component) > 0 {
		return errComponentProvidedWhenNotConfigured
	}

	if len(p.Config.Kinds) == 0 && len(p.Kind) > 0 {
		return errKindProvidedWhenNotConfigured
	}

	configuredCustoms := make([]Custom, 0)

	if len(p.Config.Kinds) > 0 && len(p.Kind) > 0 {
		kc := p.Config.KindFromKeyOrLabel(p.Kind)
		if kc == nil {
			return fmt.Errorf("%w: %s", errInvalidKind, p.Kind)
		}

		configuredCustoms = append(configuredCustoms, kc.AdditionalChoices...)
		if !kc.SkipGlobalChoices {
			configuredCustoms = append(configuredCustoms, p.Config.CustomChoices...)
		}
	} else {
		configuredCustoms = append(configuredCustoms, p.Config.CustomChoices...)
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
	if len(p.Config.Projects) == 0 {
		return nil
	}

	if len(p.Projects) == 0 {
		var err error

		projs, err := prompt.New().Ask("Projects").
			MultiChoose(
				p.Config.ProjectLabels(),
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
		_, err := p.Config.Project(proj)
		if err != nil {
			return fmt.Errorf("%w: %s", err, proj)
		}
	}

	return nil
}

func (p *Prompts) component() error {
	if len(p.Config.Components) == 0 {
		return nil
	}

	if len(p.Component) == 0 {
		var err error

		comp, err := Custom{
			Type:        CustomEnum,
			Label:       "Component",
			EnumOptions: p.Config.Components,
		}.AskPrompt(p.StdinReader)
		if err != nil {
			return err
		}

		p.Component = comp
	}

	for _, comp := range p.Config.Components {
		if comp == p.Component {
			return nil
		}
	}

	return fmt.Errorf("%w: %s", errInvalidComponent, p.Component)
}

func (p *Prompts) kind() error {
	if len(p.Config.Kinds) == 0 {
		return nil
	}

	if len(p.Kind) == 0 {
		kindLabels := make([]string, len(p.Config.Kinds))
		for i, kc := range p.Config.Kinds {
			kindLabels[i] = kc.Label
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

	for i := range p.Config.Kinds {
		kindConfig := &p.Config.Kinds[i]

		if kindConfig.Label == p.Kind || kindConfig.Key == p.Kind {
			p.KindConfig = kindConfig
			p.Kind = kindConfig.KeyOrLabel()

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
		if p.BodyEditor {
			file, err := createTempFile(runtime.GOOS, p.Config.VersionExt)
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
			bodyCustom := p.Config.Body.CreateCustom()

			body, err := bodyCustom.AskPrompt(p.StdinReader)
			if err != nil {
				return err
			}

			p.Body = body

			return err
		}
	}

	if p.expectsBody() && len(p.Body) > 0 {
		return p.Config.Body.Validate(p.Body)
	}

	return nil
}

func (p *Prompts) expectsNoBody() bool {
	return p.KindConfig != nil && p.KindConfig.SkipBody
}

func (p *Prompts) expectsBody() bool {
	return p.KindConfig == nil || !p.KindConfig.SkipBody
}

func (p *Prompts) userChoices() error {
	userChoices := make([]Custom, 0)

	if p.KindConfig == nil || !p.KindConfig.SkipGlobalChoices {
		userChoices = append(userChoices, p.Config.CustomChoices...)
	}

	if p.KindConfig != nil {
		userChoices = append(userChoices, p.KindConfig.AdditionalChoices...)
	}

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

		var err error

		p.Customs[custom.Key], err = custom.AskPrompt(p.StdinReader)
		if err != nil {
			return err
		}
	}

	return nil
}
