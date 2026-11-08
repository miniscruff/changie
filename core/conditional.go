package core

import (
	"io"

	"charm.land/huh/v2"

	tea "charm.land/bubbletea/v2"
)

// conditionalField wraps a huh field so it is only shown based on the answers
// to other fields, such as only asking for a body when the selected kind
// accepts one.
//
// huh only supports hiding whole groups, and groups are shown one at a time,
// which would hide previous answers. Instead we keep every field in a single
// group and hide individual fields by:
//   - reporting Skip so tab and shift+tab move past the field
//   - rendering nothing so the field is not displayed
//   - not asking the field in accessible mode, which asks each field in order
//     instead of using Skip
//
// This relies on huh's group calling Skip and View on the fields it stores and
// storing the model returned by Update, so the tests for this type also act as
// a check that a huh upgrade has not changed that behavior.
type conditionalField struct {
	huh.Field
	show func() bool
}

// showWhen wraps field so that it is only shown while show returns true.
// show is called on every navigation and render so it should be cheap.
func showWhen(field huh.Field, show func() bool) huh.Field {
	return &conditionalField{
		Field: field,
		show:  show,
	}
}

func (c *conditionalField) Skip() bool {
	return !c.show() || c.Field.Skip()
}

func (c *conditionalField) View() string {
	if !c.show() {
		return ""
	}

	return c.Field.View()
}

func (c *conditionalField) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	model, cmd := c.Field.Update(msg)
	c.Field = model.(huh.Field)

	return c, cmd
}

func (c *conditionalField) RunAccessible(w io.Writer, r io.Reader) error {
	if !c.show() {
		return nil
	}

	return c.Field.RunAccessible(w, r)
}

// fieldVisible returns whether field is currently shown.
func fieldVisible(field huh.Field) bool {
	c, ok := field.(*conditionalField)

	return !ok || c.show()
}

// resizeOnChange wraps a field whose value decides which other fields are
// shown, such as the kind. huh sizes the form once, from the fields visible
// when the terminal size is known, so fields shown by a later change would be
// cut off. Whenever the value changes the last terminal size is sent again so
// huh recalculates the height for the fields now visible.
type resizeOnChange struct {
	huh.Field
	size *tea.WindowSizeMsg
}

func resizeWhenChanged(field huh.Field) huh.Field {
	return &resizeOnChange{Field: field}
}

func (r *resizeOnChange) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		r.size = &size
	}

	before := r.GetValue()
	model, cmd := r.Field.Update(msg)
	r.Field = model.(huh.Field)

	if r.size == nil || r.GetValue() == before {
		return r, cmd
	}

	size := *r.size

	return r, tea.Batch(cmd, func() tea.Msg { return size })
}
