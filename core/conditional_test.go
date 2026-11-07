package core

import (
	"strings"
	"testing"

	"charm.land/huh/v2"

	"github.com/miniscruff/changie/then"
)

// conditionalTestForm builds a form of three inputs where the middle one is
// only shown while shown is true.
func conditionalTestForm(shown *bool) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().Key("first").Title("First Title"),
		showWhen(
			huh.NewInput().Key("middle").Title("Middle Title"),
			func() bool { return *shown },
		),
		huh.NewInput().Key("last").Title("Last Title"),
	))
}

func TestConditionalFieldShown(t *testing.T) {
	shown := true
	d := newFormDriver(t, conditionalTestForm(&shown))

	then.Contains(t, "Middle Title", d.view())

	d.next()
	then.Equals(t, "middle", d.focused())

	d.next()
	then.Equals(t, "last", d.focused())

	d.prev()
	then.Equals(t, "middle", d.focused())
}

func TestConditionalFieldSkippedWhenNotShown(t *testing.T) {
	shown := false
	d := newFormDriver(t, conditionalTestForm(&shown))

	then.False(t, strings.Contains(d.view(), "Middle Title"))

	d.next()
	then.Equals(t, "last", d.focused())

	d.prev()
	then.Equals(t, "first", d.focused())
}

func TestConditionalFieldUpdatesWithOtherValues(t *testing.T) {
	shown := false
	d := newFormDriver(t, conditionalTestForm(&shown))

	then.False(t, strings.Contains(d.view(), "Middle Title"))

	// in a real form the show value changes as the result of a key press,
	// such as selecting a kind, so we toggle it before the next key
	shown = true

	d.next()
	then.Equals(t, "middle", d.focused())
	then.Contains(t, "Middle Title", d.view())
}

func TestConditionalFieldKeepsValue(t *testing.T) {
	var value string

	shown := true
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Key("first"),
		showWhen(
			huh.NewInput().Key("middle").Value(&value),
			func() bool { return shown },
		),
	))
	d := newFormDriver(t, form)

	d.next()
	d.typeText("abc")

	then.Equals(t, "abc", value)
	then.Contains(t, "abc", d.view())
}

func TestConditionalFieldCompletesForm(t *testing.T) {
	shown := false
	d := newFormDriver(t, conditionalTestForm(&shown))

	d.next()
	then.False(t, d.completed())

	d.enter()
	then.True(t, d.completed())
}
