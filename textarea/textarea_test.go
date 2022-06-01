package textarea

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	textarea := newTextArea()
	view := textarea.View()

	if !strings.Contains(view, ">") {
		t.Log(view)
		t.Error("Text area did not render the prompt")
	}

	if !strings.Contains(view, "World!") {
		t.Log(view)
		t.Error("Text area did not render the placeholder")
	}
}

func TestInput(t *testing.T) {
	textarea := newTextArea()

	input := "foo"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	if !strings.Contains(view, input) {
		t.Log(view)
		t.Error("Text area did not render the input")
	}

	if textarea.col != len(input) {
		t.Log(view)
		t.Error("Text area did not move the cursor to the correct position")
	}
}

func TestSoftWrap(t *testing.T) {
	textarea := newTextArea()
	textarea.Width = 5
	textarea.Height = 5
	textarea.CharLimit = 60

	textarea, _ = textarea.Update(initialBlinkMsg{})

	input := "foo bar baz"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	for _, word := range strings.Split(input, " ") {
		if !strings.Contains(view, word) {
			t.Log(view)
			t.Error("Text area did not render the input")
		}
	}

	// Due to the word wrapping, each word will be on a new line and the
	// text area will look like this:
	//
	// > foo
	// > bar
	// > baz█
	//
	// However, due to soft-wrapping the column will still be at the end of the line.
	if textarea.row != 0 || textarea.col != len(input) {
		t.Log(view)
		t.Error("Text area did not move the cursor to the correct position")
	}
}

func TestCharLimit(t *testing.T) {
	textarea := newTextArea()

	// First input (foo bar) should be accepted as it will fall within the
	// CharLimit. Second input (baz) should not appear in the input.
	input := []string{"foo bar", "baz"}
	textarea.CharLimit = len(input[0])

	for _, k := range []rune(strings.Join(input, " ")) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()
	if strings.Contains(view, input[1]) {
		t.Log(view)
		t.Error("Text area should not include input past the character limit")
	}
}

func TestVerticalScrolling(t *testing.T) {
	textarea := newTextArea()

	textarea.Height = 1
	textarea.Width = 20
	textarea.CharLimit = 100

	textarea, _ = textarea.Update(initialBlinkMsg{})

	input := "This is a really long line that should wrap around the text area."

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	// The view should contain the first "line" of the input.
	if !strings.Contains(view, "This is a really") {
		t.Log(view)
		t.Error("Text area did not render the input")
	}

	// But we should be able to scroll to see the next line.
	// Let's scroll down for each line to view the full input.
	lines := []string{
		"long line that",
		"should wrap around",
		"the text area.",
	}
	for _, line := range lines {
		textarea.viewport.LineDown(1)
		view = textarea.View()
		if !strings.Contains(view, line) {
			t.Log(view)
			t.Error("Text area did not render the correct scrolled input")
		}
	}
}

func TestWordWrapOverflowing(t *testing.T) {
	// An interesting edge case is when the user enters many words that fill up
	// the text area and then goes back up and inserts a few words which causes
	// a cascading wrap and causes an overflow of the last line.
	//
	// In this case, we should not let the user insert more words if, after the
	// entire wrap is complete, the last line is overflowing.
	textarea := newTextArea()

	textarea.Height = 3
	textarea.Width = 20
	textarea.CharLimit = 500

	textarea, _ = textarea.Update(initialBlinkMsg{})

	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	// We have essentially filled the text area with input.
	// Let's see if we can cause wrapping to overflow the last line.
	textarea.row = 0
	textarea.col = 0

	input = "Testing"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	lastLineWidth := len(textarea.value[len(textarea.value)-1])
	if lastLineWidth > textarea.Width {
		t.Log(lastLineWidth)
		t.Log(textarea.View())
		t.Fail()
	}
}

func TestValueSoftWrap(t *testing.T) {
	textarea := newTextArea()
	textarea.Width = 16
	textarea.Height = 10
	textarea.CharLimit = 500

	textarea, _ = textarea.Update(initialBlinkMsg{})

	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	value := textarea.Value()
	if value != input {
		t.Log(value)
		t.Log(input)
		t.Fatal("The text area does not have the correct value")
	}
}

func newTextArea() Model {
	textarea := New()

	textarea.Prompt = "> "
	textarea.Placeholder = "Hello, World!"

	textarea.Focus()

	textarea, _ = textarea.Update(initialBlinkMsg{})

	return textarea
}

func keyPress(key rune) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}, Alt: false}
}
