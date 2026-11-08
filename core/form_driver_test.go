package core

import (
	"io"
	"reflect"
	"regexp"
	"testing"
	"time"

	"charm.land/huh/v2"

	tea "charm.land/bubbletea/v2"
)

// formDriver drives a huh form in tests by sending key messages straight to
// Update, without a terminal. Commands returned by the form are run and their
// messages fed back in, the same way a bubbletea program would.
type formDriver struct {
	t    *testing.T
	form *huh.Form
}

// cmdTimeout bounds how long a command may run. Timer commands, such as
// cursor blinks, never return in time and are dropped, which is fine as they
// only affect rendering.
const cmdTimeout = 20 * time.Millisecond

// maxFormMessages guards against a feedback loop of commands never settling.
const maxFormMessages = 1000

var ansiEscapes = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

func newFormDriver(t *testing.T, form *huh.Form) *formDriver {
	t.Helper()

	d := &formDriver{
		t: t,
		form: form.
			WithInput(nil).
			WithOutput(io.Discard).
			WithAccessible(false),
	}

	d.run(d.form.Init())
	d.send(tea.WindowSizeMsg{Width: 80, Height: 40})

	return d
}

func (d *formDriver) send(msgs ...tea.Msg) {
	d.t.Helper()

	queue := msgs
	for count := 0; len(queue) > 0; count++ {
		if count > maxFormMessages {
			d.t.Fatal("form did not settle after too many messages")
		}

		msg := queue[0]
		queue = queue[1:]

		model, cmd := d.form.Update(msg)
		d.form = model.(*huh.Form)

		queue = append(queue, runCmd(cmd)...)
	}
}

func (d *formDriver) run(cmd tea.Cmd) {
	d.t.Helper()
	d.send(runCmd(cmd)...)
}

// runCmd runs cmd and returns the resulting messages, expanding batches and
// sequences into their individual messages.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	result := make(chan tea.Msg, 1)

	go func() {
		result <- cmd()
	}()

	var msg tea.Msg
	select {
	case msg = <-result:
	case <-time.After(cmdTimeout):
		return nil
	}

	if msg == nil {
		return nil
	}

	// Quitting is tracked by the form state instead.
	if _, ok := msg.(tea.QuitMsg); ok {
		return nil
	}

	// Both batch and sequence messages are slices of commands, but the
	// sequence type is not exported so we check the kind instead.
	value := reflect.ValueOf(msg)
	if value.Kind() == reflect.Slice && value.Type().Elem() == reflect.TypeFor[tea.Cmd]() {
		var msgs []tea.Msg
		for i := range value.Len() {
			msgs = append(msgs, runCmd(value.Index(i).Interface().(tea.Cmd))...)
		}

		return msgs
	}

	return []tea.Msg{msg}
}

func (d *formDriver) typeText(text string) {
	d.t.Helper()

	for _, r := range text {
		d.send(tea.KeyPressMsg{Text: string(r), Code: r})
	}
}

func (d *formDriver) press(code rune) {
	d.t.Helper()
	d.send(tea.KeyPressMsg{Code: code})
}

func (d *formDriver) enter() {
	d.t.Helper()
	d.press(tea.KeyEnter)
}

func (d *formDriver) down() {
	d.t.Helper()
	d.press(tea.KeyDown)
}

func (d *formDriver) next() {
	d.t.Helper()
	d.press(tea.KeyTab)
}

func (d *formDriver) prev() {
	d.t.Helper()
	d.send(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
}

func (d *formDriver) focused() string {
	return d.form.GetFocusedField().GetKey()
}

// view returns the rendered form without color or style escape codes.
func (d *formDriver) view() string {
	return ansiEscapes.ReplaceAllString(d.form.View(), "")
}

func (d *formDriver) completed() bool {
	return d.form.State == huh.StateCompleted
}
