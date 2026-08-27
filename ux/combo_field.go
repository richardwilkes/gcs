// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// comboFieldButtonGap is the horizontal spacing between the text field and the dropdown-toggle button. It's
// deliberately zero: the two should touch, reading as one continuous surface rather than two controls with a seam
// between them.
const comboFieldButtonGap = 0

// comboFieldButtonPadding is the uniform padding on all four sides between the dropdown-toggle button's edge and
// its glyph, giving the button a comfortably larger click target than the glyph alone.
const comboFieldButtonPadding = 6

// NewComboField creates an editable combo box: a text field with a dropdown button that shows a list of options,
// laid out as a wrapper panel containing both -- the wrapper owns the single visible border (swapped between
// focused/unfocused as usual, driven by the field's real focus events -- see unison.InstallFocusBorders), while the
// field and button inside it are both borderless, so the pair reads as one seamless control rather than two boxes
// glued together.
//
// Every value this widget deals in -- each element of options, initial, and every value passed to changed -- is a
// *string, not a string, and the distinction is deliberate and load-bearing: nil means "not set" (no value at all),
// while a non-nil pointer, even one pointing at "", means a real, present value ("empty"). This widget never
// conflates the two, and options is how the caller controls both: including a nil element offers a "not set" entry
// in the dropdown; including a pointer to "" offers a distinct "empty" entry; everything else is a literal choice.
// Both a nil and an empty-string element are optional -- omit either (or both) and this widget simply won't offer
// that state at all, which is how a future caller enforcing "this nameable may not be left unset" would work: by
// leaving nil out of options, not by passing a flag to this widget. options is deduped case-insensitively by
// displayed text before use (so e.g. two elements that both display as "Fire" collapse to one, and a nil deduplicates
// against another nil the same way), preserving the order the surviving elements first appeared in.
//
// Both "not set" and "empty" are rendered as bracketed placeholder text (e.g. "«not set»") rather than true field
// content: that placeholder is atomic. Moving the cursor into it does nothing special, but the moment the user types
// a character or presses Backspace/Delete while it's showing, the whole placeholder is replaced or cleared in one
// action -- never edited into character by character.
//
// The dropdown is a plain unison.Menu (via unison.DefaultMenuFactory), built fresh and shown each time it's opened
// -- by clicking the field or button, or pressing Down while the field has focus. It always lists every option,
// unfiltered by whatever's currently in the field; the item matching the field's current displayed text (if any) is
// what starts pre-highlighted. Once open, all of the usual Menu behavior applies as-is (arrow-key navigation, Enter
// to choose, Escape or an outside click to dismiss), and clicking into the field to keep typing simply dismisses it
// like any other outside click.
//
// If editable is true, the field accepts typed input. If editable is false, the field accepts no typed input at
// all -- clicking it or the dropdown button just opens the full option list for picking.
//
// If options offers no "empty" entry (no element pointing at ""), typing the field down to a literal blank ("" --
// not the "not set" placeholder, which is never blank text) is treated as invalid, not as a value: the field is
// marked invalid (see unison.Field.Invalid, wired via ValidateCallback) and changed is not called, so nothing
// overwrites the last valid value until the user either types real text or explicitly picks something from the
// dropdown. This validation only applies to edits made through this widget -- initial is never checked against it,
// and neither is options itself beyond the dedup pass above. A file saved before a marker's rules changed (e.g. an
// "empty" entry that used to be offered no longer is) can perfectly well hold a now-invalid stored value; that
// value is displayed and preserved exactly as loaded until the user actively changes it, never silently rejected or
// coerced on load.
//
// changed, if non-nil, is called with the field's new value whenever it changes -- whether by typing (editable mode
// only) or by picking an item from the dropdown. It is not called for initial or reset (the returned func) changes,
// matching the convention that a caller-driven change to displayed state, as opposed to a user action, doesn't get
// reported back to that same caller.
//
// minWidth, if greater than zero, is a floor under the *widget's* (wrapper's) overall width, not the text field's.
// The field's own auto-calculated width (sized to fit the widest option and placeholder text) plus the button's
// fixed width plus the border/spacing between them gives the widget's natural minimum; the widget's actual minimum
// is max(minWidth, that natural minimum). Any width minWidth adds beyond the natural minimum goes entirely to the
// text field -- the button's width never varies. The dropdown, when shown, is opened against the widget's full
// width, so it's never narrower either (Menu widens further than that to fit its content, but never narrower). Pass
// 0 to just use the natural, option-driven width.
//
// The returned func resets the field's displayed value back to "not set" (nil) without invoking changed and
// without opening the dropdown, e.g. for a "clear this value" action.
func NewComboField(options []*string, editable bool, minWidth float32, initial *string, changed func(value *string)) (widget unison.Paneler, reset func()) {
	notSetDisplay := i18n.Text("«not set»")
	emptyDisplay := i18n.Text("«empty»")
	displayFor := func(value *string) string {
		switch {
		case value == nil:
			return notSetDisplay
		case *value == "":
			return emptyDisplay
		default:
			return *value
		}
	}

	seen := make(map[string]bool, len(options))
	deduped := make([]*string, 0, len(options))
	for _, one := range options {
		key := strings.ToLower(displayFor(one))
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, one)
	}
	options = deduped
	emptyAllowed := slices.ContainsFunc(options, func(one *string) bool { return one != nil && *one == "" })

	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: comboFieldButtonGap, VAlign: align.Middle})
	wrapper.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})

	field := unison.NewField()
	unison.UninstallFocusBorders(field, field) // the wrapper owns the border instead -- see below
	widthCandidates := make([]string, len(options))
	for i, one := range options {
		widthCandidates[i] = displayFor(one)
	}
	field.SetMinimumTextWidthUsing(widthCandidates...)
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})

	button := unison.NewButton()
	button.HideBase = true
	// Uniform on all four sides, so the glyph sits centered in the button's own space; generous enough that the
	// button's overall hit area is a comfortable click target despite the glyph itself being small.
	button.DrawableOnlyHMargin = comboFieldButtonPadding
	button.DrawableOnlyVMargin = comboFieldButtonPadding
	button.Font = field.Font
	button.Drawable = dropdownGlyph{}
	button.SetFocusable(false)
	button.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor { return unison.ArrowCursor() }
	button.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle, VAlign: align.Middle})
	// HideBase suppresses Button's own background/border so it doesn't look like a separate pill-shaped control;
	// paint the field's own background ink behind the glyph instead, so the button's face reads as a continuation
	// of the field's surface rather than a distinct element -- the wrapper's border is what visually contains both.
	defaultButtonDraw := button.DrawCallback
	button.DrawCallback = func(canvas *unison.Canvas, rect geom.Rect) {
		canvas.DrawRect(rect, field.EditableInk.Paint(canvas, rect, paintstyle.Fill))
		defaultButtonDraw(canvas, rect)
	}

	// minWidth is a floor under the whole widget's width, not the field's. Any of it beyond the natural
	// (option-fit-field + button + spacing + border) minimum goes entirely into the field; the button never varies.
	_, buttonPref, _ := button.Sizes(geom.Size{})
	borderInsets := unison.NewDefaultFieldBorder(false).Insets() // same for focused and unfocused, by contract
	naturalMin := borderInsets.Width() + field.MinimumTextWidth + comboFieldButtonGap + buttonPref.Width
	widgetMinWidth := max(minWidth, naturalMin)
	if widgetMinWidth > naturalMin {
		field.MinimumTextWidth += widgetMinWidth - naturalMin
	}

	unison.InstallFocusBorders(field, wrapper, unison.NewDefaultFieldBorder(true), unison.NewDefaultFieldBorder(false))

	updating := false
	setDisplay := func(value *string) {
		updating = true
		field.SetText(displayFor(value))
		updating = false
	}

	isShowingPlaceholder := func() bool {
		t := field.Text()
		return t == notSetDisplay || t == emptyDisplay
	}

	selectValue := func(value *string) {
		// Run this action deferred so it's not happening inside the popup menu callback
		// The task should trigger on the very next UI tick, after the native popup has been torn down
		unison.InvokeTask(func() {
			setDisplay(value)
			field.RequestFocus()
			if changed != nil {
				changed(value)
			}
		})
	}

	openMenu := func() {
		field.RequestFocus()
		list := options
		currentText := field.Text()
		initialIndex := 0
		fac := unison.DefaultMenuFactory()
		m := fac.NewMenu(unison.PopupMenuTemporaryBaseID, "", nil)
		defer m.Dispose()
		for i, c := range list {
			display := displayFor(c)
			if display == currentText {
				initialIndex = i
			}
			value := c
			m.InsertItem(-1, fac.NewItem(unison.PopupMenuTemporaryBaseID+i+1, display, unison.KeyBinding{}, nil,
				func(_ unison.MenuItem) { selectValue(value) }))
		}
		m.Popup(wrapper.RectToRoot(wrapper.ContentRect(true)), initialIndex)
	}

	field.ModifiedCallback = func(_, after *unison.FieldState) {
		if updating {
			return
		}
		if !emptyAllowed && after.Text == "" {
			// "" isn't a legal value here -- options offers no "empty" entry. ValidateCallback below marks the
			// field invalid; changed is deliberately not called, so a value that was never really chosen doesn't
			// get silently written -- whatever was last valid stays in place until the user either types real text
			// or explicitly picks an entry from the dropdown.
			return
		}
		value := after.Text
		if changed != nil {
			changed(&value)
		}
	}

	field.ValidateCallback = func() bool {
		// Mirrors the ModifiedCallback rule above: blank text is only valid when options offers an "empty" entry.
		// This never misfires on the "not set"/"empty" placeholder text, since neither is actually blank.
		return emptyAllowed || field.Text() != ""
	}

	defaultKeyDown := field.KeyDownCallback
	field.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		switch keyCode {
		case unison.KeyDown:
			openMenu()
			return true
		case unison.KeyBackspace, unison.KeyDelete:
			if !editable {
				return true
			}
			if isShowingPlaceholder() {
				// The whole placeholder goes in one action -- never partially erased. Left unguarded (unlike
				// setDisplay) so the normal ModifiedCallback fires and reports the resulting blank value.
				field.SetText("")
				return true
			}
		}
		if !editable {
			return true
		}
		return defaultKeyDown(keyCode, mods, repeat)
	}

	if editable {
		defaultRuneTyped := field.RuneTypedCallback
		field.RuneTypedCallback = func(ch rune) bool {
			if isShowingPlaceholder() {
				// The placeholder is replaced wholesale by whatever gets typed, not edited into.
				setDisplay(new(string))
			}
			return defaultRuneTyped(ch)
		}
	} else {
		field.RuneTypedCallback = func(_ rune) bool { return true }
		field.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool {
			openMenu()
			return true
		}
	}

	button.ClickCallback = openMenu
	wrapper.AddChild(field)
	wrapper.AddChild(button)

	setDisplay(initial)

	return wrapper, func() { setDisplay(nil) }
}

// dropdownGlyph is a small hand-drawn "expand" triangle used as a unison.Button Drawable for the combo field's
// dropdown-toggle button, matching the glyph unison.PopupMenu draws for the same purpose.
type dropdownGlyph struct{}

func (dropdownGlyph) LogicalSize() geom.Size {
	return geom.NewSize(8, 5)
}

func (dropdownGlyph) DrawInRect(canvas *unison.Canvas, rect geom.Rect, _ *unison.SamplingOptions, paint *unison.Paint) {
	path := unison.NewPath()
	path.MoveTo(geom.NewPoint(rect.X, rect.Y))
	path.LineTo(geom.NewPoint(rect.Right(), rect.Y))
	path.LineTo(geom.NewPoint(rect.X+rect.Width/2, rect.Bottom()))
	path.Close()
	canvas.DrawPath(path, paint)
}
