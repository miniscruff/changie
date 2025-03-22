package core

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/charmbracelet/huh"
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

	fields := make([]huh.Field, 0)

	if len(p.Config.Projects) > 0 && len(p.Projects) == 0 {
		projectOptions := make([]huh.Option[string], len(p.Config.Projects))
		for i, proj := range p.Config.Projects {
			projectOptions[i] = huh.NewOption(proj.Label, proj.Key)
		}

		fields = append(fields, huh.NewMultiSelect[string]().
			Title("Projects").
			Options(projectOptions...).
			//TODO validation
			Value(&p.Projects),
		)
	}

	if len(p.Config.Components) > 0 && len(p.Component) == 0 {
		componentOptions := make([]huh.Option[string], len(p.Config.Components))
		for i, comp := range p.Config.Components {
			componentOptions[i] = huh.NewOption(comp, comp)
		}
		fields = append(fields, huh.NewSelect[string]().
			Title("Component").
			Options(componentOptions...).
			//TODO validation
			Value(&p.Component),
		)
	}

	if len(p.Config.Kinds) > 0 && len(p.Kind) == 0 {
		kindOptions := make([]huh.Option[string], len(p.Config.Kinds))
		for i, kindConfig := range p.Config.Kinds {
			kindOptions[i] = huh.NewOption(kindConfig.Label, kindConfig.Key)
		}

		fields = append(fields, huh.NewSelect[string]().
			Title("Kind").
			Options(kindOptions...).
			Validate(func(value string) error {
				kc := p.Config.KindFromKeyOrLabel(p.Kind)
				if kc == nil {
					return errInvalidKind
				}

				p.KindConfig = kc
				return nil
			}).
			Value(&p.Kind),
		)
	}

	fields = append(fields, huh.NewText().
		Title("Body").
		//TODO validation
		WithHideFunc(func() bool {
			kc := p.Config.KindFromKeyOrLabel(p.Kind)
			if kc == nil {
				return false
			}
			return kc.SkipBody
		}, &p.Kind).
		Value(&p.Body),
	)

	for _, choice := range p.Config.CustomChoices {
		valFunc := func() bool {
			kc := p.Config.KindFromKeyOrLabel(p.Kind)
			if kc == nil {
				return true
			}
			return kc.SkipGlobalChoices
		}

		var f huh.Field
		switch choice.Type {
		case CustomString:
			var s string
			f = huh.NewText().
				Title(choice.DisplayLabel()).
				//TODO validation
				WithHideFunc(valFunc, &p.Kind).
				Validate(func(value string) error {
					if len(value) == 0 {
						delete(p.Customs, choice.Key)
					} else {
						p.Customs[choice.Key] = value
					}

					return nil
				}).
				Value(&s)
		// case CustomBlock:
		// return c.askBlock(stdinReader)
		case CustomInt:
			var s string
			f = huh.NewText().
				Title(choice.DisplayLabel()).
				WithHideFunc(valFunc, &p.Kind).
				Validate(func(value string) error {
					//TODO validation
					if len(value) == 0 {
						delete(p.Customs, choice.Key)
					} else {
						p.Customs[choice.Key] = value
					}

					return nil
				}).
				Value(&s)
			// case CustomEnum:
			// return c.askEnum(stdinReader)
		}
		fields = append(fields, f)
	}

	for _, kc := range p.Config.Kinds {
		for _, choice := range kc.AdditionalChoices {
			valFunc := func() bool {
				configKind := p.Config.KindFromKeyOrLabel(p.Kind)
				if configKind == nil {
					return true
				}
				return configKind.Key != kc.Key
			}

			var f huh.Field
			switch choice.Type {
			case CustomString:
				var s string
				f = huh.NewText().
					Title(choice.DisplayLabel()).
					WithHideFunc(valFunc, &p.Kind).
					Validate(func(value string) error {
						if len(value) == 0 {
							delete(p.Customs, choice.Key)
						} else {
							p.Customs[choice.Key] = value
						}

						return nil
					}).
					Value(&s)
			// case CustomBlock:
			// return c.askBlock(stdinReader)
			case CustomInt:
				var s string
				f = huh.NewText().
					Title(choice.DisplayLabel()).
					WithHideFunc(valFunc, &p.Kind).
					Validate(func(value string) error {
						//TODO validation
						if len(value) == 0 {
							delete(p.Customs, choice.Key)
						} else {
							p.Customs[choice.Key] = value
						}

						return nil
					}).
					Value(&s)
				// case CustomEnum:
				// return c.askEnum(stdinReader)
			}
			fields = append(fields, f)
		}
	}

	/*
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
	*/

	form := huh.NewForm(huh.NewGroup(fields...))

	err = form.Run()
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
