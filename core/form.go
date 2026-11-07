package core

import (
	"slices"

	"charm.land/huh/v2"
)

// maxOptionsHeight limits how tall select fields grow so long option lists,
// such as hundreds of projects, scroll instead of overflowing the terminal.
const maxOptionsHeight = 12

// newForm builds a single huh form asking for every configured value that
// was not already provided by flags or environment variables.
// If there is nothing to ask the form is nil.
// All fields are kept in one group so previous answers stay visible and
// shift+tab can be used to go back.
func (p *Prompts) newForm() (*huh.Form, error) {
	var fields []huh.Field

	fields = append(fields, p.projectFields()...)
	fields = append(fields, p.componentFields()...)
	fields = append(fields, p.kindFields()...)
	fields = append(fields, p.bodyFields()...)

	customFields, err := p.customFields()
	if err != nil {
		return nil, err
	}

	fields = append(fields, customFields...)

	// huh panics on an empty group and never completes if every field is
	// hidden, which can happen when the kind is provided
	if !slices.ContainsFunc(fields, fieldVisible) {
		return nil, nil
	}

	return huh.NewForm(huh.NewGroup(fields...)), nil
}

func (p *Prompts) projectFields() []huh.Field {
	if len(p.Config.Projects) == 0 || len(p.Projects) > 0 {
		return nil
	}

	options := make([]huh.Option[string], len(p.Config.Projects))
	for i, label := range p.Config.ProjectLabels() {
		options[i] = huh.NewOption(label, p.Config.Projects[i].Key)
	}

	return []huh.Field{
		huh.NewMultiSelect[string]().
			Key("projects").
			Title("Projects").
			Options(options...).
			Height(min(len(options)+1, maxOptionsHeight)).
			Validate(func(projects []string) error {
				if len(projects) == 0 {
					return errProjectRequired
				}

				return nil
			}).
			Value(&p.Projects),
	}
}

func (p *Prompts) componentFields() []huh.Field {
	if len(p.Config.Components) == 0 || len(p.Component) > 0 {
		return nil
	}

	return []huh.Field{
		huh.NewSelect[string]().
			Key("component").
			Title("Component").
			Options(huh.NewOptions(p.Config.Components...)...).
			Height(min(len(p.Config.Components)+1, maxOptionsHeight)).
			Value(&p.Component),
	}
}

func (p *Prompts) kindFields() []huh.Field {
	if len(p.Config.Kinds) == 0 || len(p.Kind) > 0 {
		return nil
	}

	options := make([]huh.Option[string], len(p.Config.Kinds))
	for i := range p.Config.Kinds {
		kc := &p.Config.Kinds[i]
		options[i] = huh.NewOption(kc.Label, kc.KeyOrLabel())
	}

	return []huh.Field{
		huh.NewSelect[string]().
			Key("kind").
			Title("Kind").
			Options(options...).
			Height(min(len(options)+1, maxOptionsHeight)).
			Value(&p.Kind),
	}
}

// selectedKind returns the config of the currently selected kind, or nil if
// kinds are not configured. This is either the kind provided by flag or the
// kind under the cursor. The kind field updates the selection as the
// cursor moves, so fields depending on the kind can use this to hide
// themselves while the form is running.
func (p *Prompts) selectedKind() *KindConfig {
	return p.Config.KindFromKeyOrLabel(p.Kind)
}

func (p *Prompts) bodyFields() []huh.Field {
	if len(p.Body) > 0 {
		return nil
	}

	// the body is written with the external editor after the form instead
	if p.BodyEditor {
		return nil
	}

	custom := p.Config.Body.CreateCustom()

	var field huh.Field

	if p.Config.Body.UseBlock {
		text := huh.NewText().
			Key("body").
			Title(custom.DisplayLabel()).
			Validate(custom.validateString).
			Value(&p.Body)

		if p.Config.VersionExt != "" {
			text.EditorExtension(p.Config.VersionExt)
		}

		field = text
	} else {
		field = huh.NewInput().
			Key("body").
			Title(custom.DisplayLabel()).
			Validate(custom.validateString).
			Value(&p.Body)
	}

	return []huh.Field{hideWhen(field, p.skipsBody)}
}

// skipsBody returns whether the selected kind does not accept a body.
func (p *Prompts) skipsBody() bool {
	kc := p.selectedKind()

	return kc != nil && kc.SkipBody
}

// formCustom is a custom choice asked in the form along with its answer.
// Each field keeps its own answer as kinds may share custom keys.
type formCustom struct {
	custom Custom
	value  string
	hidden func() bool
}

// customFields creates fields for the global custom choices, hidden when the
// selected kind skips them, followed by every kind's additional choices,
// only shown while that kind is selected.
func (p *Prompts) customFields() ([]huh.Field, error) {
	p.customAnswers = nil

	for _, custom := range p.Config.CustomChoices {
		if p.customProvided(custom) {
			continue
		}

		p.customAnswers = append(p.customAnswers, &formCustom{
			custom: custom,
			hidden: p.skipsGlobalChoices,
		})
	}

	for i := range p.Config.Kinds {
		kindKey := p.Config.Kinds[i].KeyOrLabel()
		notSelected := func() bool {
			kc := p.selectedKind()

			return kc == nil || kc.KeyOrLabel() != kindKey
		}

		for _, custom := range p.Config.Kinds[i].AdditionalChoices {
			if p.customProvided(custom) {
				continue
			}

			p.customAnswers = append(p.customAnswers, &formCustom{
				custom: custom,
				hidden: notSelected,
			})
		}
	}

	fields := make([]huh.Field, len(p.customAnswers))

	for i, answer := range p.customAnswers {
		field, err := answer.custom.formField(&answer.value)
		if err != nil {
			return nil, err
		}

		fields[i] = hideWhen(field, answer.hidden)
	}

	return fields, nil
}

// customProvided returns whether the custom value was provided by flag.
func (p *Prompts) customProvided(custom Custom) bool {
	return p.Customs[custom.Key] != ""
}

// skipsGlobalChoices returns whether the selected kind skips global choices.
func (p *Prompts) skipsGlobalChoices() bool {
	kc := p.selectedKind()

	return kc != nil && kc.SkipGlobalChoices
}

// formCustoms returns the custom values answered in the form, ignoring
// fields hidden by the selected kind.
func (p *Prompts) formCustoms() map[string]string {
	customs := make(map[string]string)

	for _, answer := range p.customAnswers {
		if answer.hidden() {
			continue
		}

		customs[answer.custom.Key] = answer.value
	}

	return customs
}
