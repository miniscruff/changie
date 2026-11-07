package core

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// promptsLegacy sets LegacyPrompts for the shared prompt tests, which run with
// the huh form by default and again with the cqroot prompts by
// TestPromptsWithLegacy.
// Remove this along with cqroot once the huh form is the only option.
var promptsLegacy = false

func TestPromptsWithLegacy(t *testing.T) {
	promptsLegacy = true

	t.Cleanup(func() { promptsLegacy = false })

	for _, test := range []func(*testing.T){
		TestAskPromptsForBody,
		TestAskPromptsForBodyWithProject,
		TestAskPromptsForBodyWithProjectErrBadProject,
		TestAskPromptsFailIfDisabled,
		TestAskPromptsForBodyWithProjectErrBadInput,
		TestAskComponentKindBody,
		TestBodyKindPostProcess,
		TestBodyCustom,
		TestBodyCustomWithExistingCustomValue,
		TestSkippedBodyGlobalChoicesKindWithAdditional,
		TestBodyAndPostProcess,
		TestBodyAndPostProcessSkipGlobalPost,
		TestErrorInvalidPost,
		TestErrorInvalidPostWithProjects,
		TestErrorInvalidCustomType,
		TestErrorInvalidComponent,
		TestErrorInvalidKind,
		TestErrorFaultBody,
		TestErrorInvalidCustomValue,
		TestErrorBadKindLabel,
		TestErrorBadComponentInput,
		TestErrorComponentGivenWithNoConfiguration,
		TestErrorKindGivenWithNoConfiguration,
		TestErrorCustomGivenWithNoConfiguration,
		TestErrorCustomGivenDoesNotPassValidation,
		TestErrorBodyGivenWithNoConfiguration,
		TestErrorBodyGivenDoesNotPassValidation,
		TestSkipPromptForComponentIfSet,
		TestSkipPromptForCustomIfOptionalAndDisabled,
		TestSkipPromptForPromptsWithCustomPromptsInKindConfig,
		TestSkipPromptForKindIfSet,
		TestSkipPromptForBodyIfSet,
		TestGetBodyTxtWithEditor,
		TestGetBodyTxtWithEditorUnableToCreateCmd,
		TestAskPromptsKeepsProvidedCustoms,
		TestAskPromptsNothingToAsk,
		TestAskPromptsProjectWithoutKey,
		TestAskPromptsKindWithoutKey,
	} {
		name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
		t.Run(name[strings.LastIndex(name, ".")+1:], test)
	}
}
