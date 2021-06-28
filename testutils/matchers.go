package testutils

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/spf13/afero"
)

func HaveContents(afs afero.Afero, content string) types.GomegaMatcher {
	return &haveContentMatcher{
		afs:     afs,
		content: content,
	}
}

type haveContentMatcher struct {
	afs      afero.Afero
	content  string
	expected string
}

func (matcher *haveContentMatcher) Match(actual interface{}) (success bool, err error) {
	pathStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("beAnEmptyFile matcher expects a string")
	}

	bs, err := matcher.afs.ReadFile(pathStr)
	if err != nil {
		return false, err
	}

	matcher.expected = string(bs)

	return matcher.expected == matcher.content, nil
}

func (matcher *haveContentMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"Expected '%s' to have contents\n\t'%s' but was \n\t'%s'",
		actual,
		matcher.content,
		matcher.expected,
	)
}

func (matcher *haveContentMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"Expected '%s' to not have contents\n\t'%s' but was \n\t'%s'",
		actual,
		matcher.content,
		matcher.expected,
	)
}

func BeAnEmptyFile(afs afero.Afero) types.GomegaMatcher {
	return &beAnEmptyFileMatcher{
		afs: afs,
	}
}

type beAnEmptyFileMatcher struct {
	afs afero.Afero
}

func (matcher *beAnEmptyFileMatcher) Match(actual interface{}) (success bool, err error) {
	pathStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("beAnEmptyFile matcher expects a string")
	}

	bs, err := matcher.afs.ReadFile(pathStr)
	if err != nil {
		return false, err
	}

	return len(bs) == 0, nil
}

func (*beAnEmptyFileMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected '%s' to be an empty file", actual)
}

func (*beAnEmptyFileMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected not to be '%s' to be an empty file", actual)
}

func BeADir(afs afero.Afero) types.GomegaMatcher {
	return &beADirMatcher{
		afs: afs,
	}
}

type beADirMatcher struct {
	afs afero.Afero
}

func (matcher *beADirMatcher) Match(actual interface{}) (success bool, err error) {
	pathStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("beADir matcher expects a string")
	}

	return matcher.afs.DirExists(pathStr)
}

func (*beADirMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected '%s' to be a directory", actual)
}

func (*beADirMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected not to be '%s' to be a directory", actual)
}
