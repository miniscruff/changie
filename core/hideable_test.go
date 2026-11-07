package core

import (
	"strings"
	"testing"

	"charm.land/huh/v2"

	"github.com/miniscruff/changie/then"
)

// hideTestForm builds a form of three inputs where the middle one is hidden
// whenever hidden is true.
func hideTestForm(hidden *bool) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().Key("first").Title("First Title"),
		hideWhen(
			huh.NewInput().Key("middle").Title("Middle Title"),
			func() bool { return *hidden },
		),
		huh.NewInput().Key("last").Title("Last Title"),
	))
}

func TestHideableFieldShownWhenNotHidden(t *testing.T) {
	hidden := false
	d := newFormDriver(t, hideTestForm(&hidden))

	then.Contains(t, "Middle Title", d.view())

	d.next()
	then.Equals(t, "middle", d.focused())

	d.next()
	then.Equals(t, "last", d.focused())

	d.prev()
	then.Equals(t, "middle", d.focused())
}

func TestHideableFieldSkippedWhenHidden(t *testing.T) {
	hidden := true
	d := newFormDriver(t, hideTestForm(&hidden))

	then.False(t, strings.Contains(d.view(), "Middle Title"))

	d.next()
	then.Equals(t, "last", d.focused())

	d.prev()
	then.Equals(t, "first", d.focused())
}

func TestHideableFieldUpdatesWithOtherValues(t *testing.T) {
	hidden := true
	d := newFormDriver(t, hideTestForm(&hidden))

	then.False(t, strings.Contains(d.view(), "Middle Title"))

	// in a real form the hide value changes as the result of a key press,
	// such as selecting a kind, so we toggle it before the next key
	hidden = false

	d.next()
	then.Equals(t, "middle", d.focused())
	then.Contains(t, "Middle Title", d.view())
}

func TestHideableFieldKeepsValue(t *testing.T) {
	var value string

	hidden := false
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Key("first"),
		hideWhen(
			huh.NewInput().Key("middle").Value(&value),
			func() bool { return hidden },
		),
	))
	d := newFormDriver(t, form)

	d.next()
	d.typeText("abc")

	then.Equals(t, "abc", value)
	then.Contains(t, "abc", d.view())
}

func TestHideableFieldCompletesForm(t *testing.T) {
	hidden := true
	d := newFormDriver(t, hideTestForm(&hidden))

	d.next()
	then.False(t, d.completed())

	d.enter()
	then.True(t, d.completed())
}
