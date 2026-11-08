package core

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"slices"

	"github.com/charmbracelet/x/term"
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

// newPromptsWarning is shown before the huh form while the legacy prompts are
// still available, so users know how to switch back if they hit a problem.
const newPromptsWarning = "changie is using new prompts, if you run into any issues please report them at " +
	"https://github.com/miniscruff/changie/issues and set CHANGIE_LEGACY_PROMPTS=on to use the previous prompts, " +
	"or CHANGIE_LEGACY_PROMPTS=off to hide this warning."

type Prompts struct {
	Config           *Config
	StdinReader      io.Reader
	KindConfig       *KindConfig
	BodyEditor       bool
	EditorCmdBuilder func(string) (EditorRunner, error)
	TimeNow          TimeNow

	// Enabled checks to make sure our terminal supports prompts
	Enabled bool
	// LegacyPrompts asks for missing values using the cqroot prompts, one at
	// a time, instead of the huh form.
	LegacyPrompts bool
	// WarningWriter, if set, is where the new prompts warning is written
	// before the huh form is shown.
	WarningWriter io.Writer

	// Values can be submitted from the environment or shell arguments.
	Projects  []string
	Component string
	Kind      string
	Body      string
	Customs   map[string]string

	// customAnswers are the custom choices asked in the form.
	customAnswers []*formCustom
	// bodyProvided is whether the body was provided before the form ran.
	bodyProvided bool
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

	if p.LegacyPrompts {
		err = p.askWithCqroot()
	} else {
		err = p.askWithHuh()
	}

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
	for i := range changes {
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

	for _, proj := range p.Projects {
		_, err := p.Config.Project(proj)
		if err != nil {
			return fmt.Errorf("%w: %s", err, proj)
		}
	}

	if len(p.Component) > 0 && !slices.Contains(p.Config.Components, p.Component) {
		return fmt.Errorf("%w: %s", errInvalidComponent, p.Component)
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

// askWithHuh asks for every missing value in a single huh form, returning an
// error for missing values when prompts are disabled.
func (p *Prompts) askWithHuh() error {
	p.bodyProvided = len(p.Body) > 0

	var err error

	if p.Enabled {
		err = p.runForm()
	} else {
		err = p.requireProvided()
	}

	if err != nil {
		return err
	}

	return p.resolveValues()
}

// requireProvided returns an error if any value we would prompt for was not
// provided, used when prompts are disabled.
// Optional custom values may be missing.
func (p *Prompts) requireProvided() error {
	if len(p.Config.Projects) > 0 && len(p.Projects) == 0 {
		return errProjectMissingPromptDisabled
	}

	if len(p.Config.Components) > 0 && len(p.Component) == 0 {
		return errComponentMissingPromptDisabled
	}

	if len(p.Config.Kinds) > 0 && len(p.Kind) == 0 {
		return errKindMissingPromptDisabled
	}

	kc := p.selectedKind()

	if (kc == nil || !kc.SkipBody) && len(p.Body) == 0 {
		return errBodyMissingPromptDisabled
	}

	choices := make([]Custom, 0)

	if kc == nil || !kc.SkipGlobalChoices {
		choices = append(choices, p.Config.CustomChoices...)
	}

	if kc != nil {
		choices = append(choices, kc.AdditionalChoices...)
	}

	for _, custom := range choices {
		if !custom.Optional && !p.customProvided(custom) {
			return fmt.Errorf("%w: custom key '%s'", errCustomMissingPromptDisabled, custom.Key)
		}
	}

	return nil
}

// runForm asks for any missing values using a huh form.
func (p *Prompts) runForm() error {
	form, err := p.newForm()
	if err != nil {
		return err
	}

	if form == nil {
		return nil
	}

	if p.WarningWriter != nil {
		_, _ = fmt.Fprintln(p.WarningWriter, newPromptsWarning)
	}

	err = form.WithInput(formInput(p.StdinReader)).Run()
	if err != nil {
		return err
	}

	p.applyFormAnswers()

	return nil
}

// applyFormAnswers tidies the values answered in the form and adds the
// custom answers to the provided custom values.
func (p *Prompts) applyFormAnswers() {
	if !p.bodyProvided {
		// a body typed before switching to a kind without a body is not used
		if !p.acceptsBody() {
			p.Body = ""
		}

		// block bodies may have been written with an editor using ctrl+e
		p.Body = cleanText(p.Body)
	}

	answers := p.formCustoms()
	if len(answers) > 0 && p.Customs == nil {
		p.Customs = make(map[string]string)
	}

	maps.Copy(p.Customs, answers)
}

// resolveValues resolves the kind config and validates the body, which may
// still be written using an editor.
func (p *Prompts) resolveValues() error {
	if len(p.Config.Kinds) > 0 {
		kc := p.selectedKind()
		if kc == nil {
			return fmt.Errorf("%w: %s", errInvalidKind, p.Kind)
		}

		p.KindConfig = kc
		p.Kind = kc.KeyOrLabel()
	}

	return p.resolveBody()
}

func (p *Prompts) resolveBody() error {
	if !p.acceptsBody() {
		if len(p.Body) > 0 {
			return fmt.Errorf("%w: %s", errKindDoesNotAcceptBody, p.Kind)
		}

		return nil
	}

	if len(p.Body) > 0 {
		return p.Config.Body.Validate(p.Body)
	}

	if !p.Enabled || !p.BodyEditor {
		return nil
	}

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
}

// formInput returns the input for the huh form.
// bubbletea reads files with a cancel reader, which it can close while its
// read loop is still running when the form ends, a data race caught by tests
// run with -race. Files that are not terminals, such as piped input, do not
// need raw mode or canceling as only one form is run per process, so they are
// read as a plain reader instead.
func formInput(input io.Reader) io.Reader {
	file, ok := input.(*os.File)
	if !ok || term.IsTerminal(file.Fd()) {
		return input
	}

	return struct{ io.Reader }{file}
}
