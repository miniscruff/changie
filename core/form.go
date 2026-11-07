package core

import (
	"charm.land/huh/v2"
)

// maxOptionsHeight limits how tall select fields grow so long option lists,
// such as hundreds of projects, scroll instead of overflowing the terminal.
const maxOptionsHeight = 12

// newForm builds a single huh form asking for every configured value.
// All fields are kept in one group so previous answers stay visible and
// shift+tab can be used to go back.
func (p *Prompts) newForm() *huh.Form {
	var fields []huh.Field

	fields = append(fields, p.projectFields()...)
	fields = append(fields, p.componentFields()...)
	fields = append(fields, p.kindFields()...)

	return huh.NewForm(huh.NewGroup(fields...))
}

func (p *Prompts) projectFields() []huh.Field {
	if len(p.Config.Projects) == 0 {
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
	if len(p.Config.Components) == 0 {
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
	if len(p.Config.Kinds) == 0 {
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
// kinds are not configured. The kind field updates the selection as the
// cursor moves, so fields depending on the kind can use this to hide
// themselves while the form is running.
func (p *Prompts) selectedKind() *KindConfig {
	return p.Config.KindFromKeyOrLabel(p.Kind)
}
