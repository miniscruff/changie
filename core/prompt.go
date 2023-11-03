package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/miniscruff/changie/shared"
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

type PromptChange struct {
	Projects  []string
	Component string
	Kind      string
	Body      string
	Time      time.Time
	Custom    map[string]string
	Env       map[string]string
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
func (change *PromptChange) AskPrompts(ctx PromptContext) error {
	err := change.validateArguments(ctx.Config)
	if err != nil {
		return err
	}

	err = change.promptForProjects(&ctx)
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
func (change *PromptChange) validateArguments(config *Config) error {
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

func (change *PromptChange) promptForProjects(ctx *PromptContext) error {
	if len(ctx.Config.Projects) == 0 {
		return nil
	}

	if len(change.Projects) == 0 {
		fmt.Println(ProjectsWarning)

		var err error

		change.Projects, err = Custom{
			Type:        CustomEnum,
			Label:       "Projects",
			EnumOptions: ctx.Config.ProjectLabels(),
		}.AskMultiChoosePrompt(ctx.StdinReader)
		if err != nil {
			return err
		}
	}

	for i, p := range change.Projects {
		pc, err := ctx.Config.Project(p)
		if err != nil {
			return fmt.Errorf("%w: %s", err, p)
		}

		change.Projects[i] = pc.Key
	}

	return nil
}

func (change *PromptChange) promptForComponent(ctx *PromptContext) error {
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
func (change *PromptChange) promptForKind(ctx *PromptContext) error {
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

func (change *PromptChange) promptForBody(ctx *PromptContext) error {
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

func (change *PromptChange) promptForUserChoices(ctx *PromptContext) error {
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

func (change *PromptChange) postProcess(ctx *PromptContext) error {
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
