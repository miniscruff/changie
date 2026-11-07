package core

import (
	"io"

	"charm.land/huh/v2"

	tea "charm.land/bubbletea/v2"
)

// hideableField wraps a huh field so it can be hidden based on the answers
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
type hideableField struct {
	huh.Field
	hide func() bool
}

// hideWhen wraps field so that it is hidden whenever hide returns true.
// hide is called on every navigation and render so it should be cheap.
func hideWhen(field huh.Field, hide func() bool) huh.Field {
	return &hideableField{
		Field: field,
		hide:  hide,
	}
}

func (h *hideableField) Skip() bool {
	return h.hide() || h.Field.Skip()
}

func (h *hideableField) View() string {
	if h.hide() {
		return ""
	}

	return h.Field.View()
}

func (h *hideableField) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	model, cmd := h.Field.Update(msg)
	h.Field = model.(huh.Field)

	return h, cmd
}

func (h *hideableField) RunAccessible(w io.Writer, r io.Reader) error {
	if h.hide() {
		return nil
	}

	return h.Field.RunAccessible(w, r)
}

// fieldVisible returns whether field is currently shown.
func fieldVisible(field huh.Field) bool {
	h, ok := field.(*hideableField)

	return !ok || !h.hide()
}

// resizeOnChange wraps a field whose value decides which other fields are
// hidden, such as the kind. huh sizes the form once, from the fields visible
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
