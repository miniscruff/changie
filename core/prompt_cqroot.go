package core

import (
	"fmt"
	"runtime"
	"slices"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/multichoose"

	tea "github.com/charmbracelet/bubbletea"
)

// The cqroot prompts are kept behind CHANGIE_LEGACY_PROMPTS while the huh
// form is tested by more users, after which this file can be removed along
// with cqroot. Provided values are validated before asking, in
// validateArguments.

// askWithCqroot asks for each missing value one prompt at a time, returning an
// error for missing values when prompts are disabled.
func (p *Prompts) askWithCqroot() error {
	err := p.projects()
	if err != nil {
		return err
	}

	err = p.component()
	if err != nil {
		return err
	}

	err = p.kind()
	if err != nil {
		return err
	}

	err = p.body()
	if err != nil {
		return err
	}

	return p.userChoices()
}

func (p *Prompts) projects() error {
	if len(p.Config.Projects) == 0 {
		return nil
	}

	if len(p.Projects) == 0 {
		if !p.Enabled {
			return errProjectMissingPromptDisabled
		}

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

	return nil
}

func (p *Prompts) component() error {
	if len(p.Config.Components) == 0 {
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
			EnumOptions: p.Config.Components,
		}.AskPrompt(p.StdinReader)
		if err != nil {
			return err
		}

		p.Component = comp
	}

	return nil
}

func (p *Prompts) kind() error {
	if len(p.Config.Kinds) == 0 {
		return nil
	}

	if len(p.Kind) == 0 {
		if !p.Enabled {
			return errKindMissingPromptDisabled
		}

		kindLabels := make([]string, len(p.Config.Kinds))
		for i, kc := range p.Config.Kinds {
			kindLabels[i] = kc.Label
		}

		label, err := Custom{
			Type:        CustomEnum,
			Label:       "Kind",
			EnumOptions: kindLabels,
		}.AskPrompt(p.StdinReader)
		if err != nil {
			return err
		}

		// kinds are looked up by key, so use the key of the selected label
		i := slices.IndexFunc(p.Config.Kinds, func(kc KindConfig) bool {
			return kc.Label == label
		})
		p.Kind = p.Config.Kinds[i].KeyOrLabel()
	}

	// provided kinds are already validated
	p.KindConfig = p.selectedKind()

	return nil
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

		if !p.Enabled {
			if custom.Optional {
				continue
			}

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
