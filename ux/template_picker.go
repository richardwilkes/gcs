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
	"fmt"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/side"
)

// The picker processing is held in a variable so that tests, which have no way to respond to the dialogs it presents, can
// substitute their own.
var promptForPickers = processPickers

// processPickers presents the template picker dialog for each row of the parts that has one, replacing the rows with
// the resulting choices. It returns false if the user canceled one of them, in which case the parts must be discarded.
// The operation describes what the dialogs are part of and may be empty (see newOperationLabel). promptChoices says
// whether the modifier prompt follows, so the rows are costed as it will see them.
func processPickers(op promptOperation, parts *applyParts, promptChoices bool) bool {
	return parts.all(func(part applyPartOps) bool { return part.resolvePickers(op, promptChoices) })
}

// showPicker puts up the dialog for the choice container, returning how it was closed. The boxes checked in it are
// recorded in the session. depth is how many of these dialogs it is stacked on.
func (s *pickerSession[T]) showPicker(row T, depth int) int {
	dialog, _ := s.newPickerDialog(row, depth)
	if dialog == nil {
		return unison.ModalResponseCancel
	}
	return dialog.RunModal()
}

// newPickerDialog returns the dialog showPicker puts up, or nil if it couldn't be made, and what brings it up to date
// with the session.
func (s *pickerSession[T]) newPickerDialog(row T, depth int) (dialog *unison.Dialog, refresh func()) {
	tp := templatePicker(row)
	headers := pickerRowDetailHeaders(row)
	// A column for the pencil, and one for the choose button when the modifier prompt follows or an option is or holds a
	// choice.
	chooseColumn := s.prompted || slices.ContainsFunc(gurps.TemplateChoiceOptions(row), func(child T) bool {
		return gurps.IsTemplateChoiceContainer(child) || len(s.nestedChoices(child)) != 0
	})
	buttonColumns := 1
	if chooseColumn {
		buttonColumns++
	}
	list := &pickerList{
		panel:        unison.NewPanel(),
		columns:      1 + len(headers) + buttonColumns,
		pt:           tp.Type,
		depth:        depth,
		chooseColumn: chooseColumn,
	}
	list.panel.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)))
	list.panel.SetLayout(&unison.FlexLayout{
		Columns:  list.columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	if len(headers) != 0 {
		list.panel.AddChild(unison.NewPanel())
		for _, header := range headers {
			label := unison.NewLabel()
			label.Font = fonts.FieldSecondary
			label.SetTitle(header)
			label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
			list.panel.AddChild(label)
		}
		for range buttonColumns {
			list.panel.AddChild(unison.NewPanel())
		}
	}
	list.claim(0, nil)

	progress, updateProgress := newPickerStatePill(unison.StdVSpacing * 2)
	progress.Side = side.Right
	scroll := unison.NewScrollPanel()
	hint := newWrappingLabel()
	hint.font = fonts.FieldSecondary
	// The hint wraps to the list's width rather than widening the dialog.
	hint.SetSizer(func(size geom.Size) (minSize, prefSize, maxSize geom.Size) {
		if size.Width <= 0 {
			_, pref, _ := scroll.Sizes(geom.Size{})
			size.Width = pref.Width
		}
		return hint.sizes(size)
	})
	hint.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill})
	refresh = func() {
		for _, update := range list.updates {
			update()
		}
		state := s.state(row)
		updateProgress(state, s.pillText(row))
		progress.Tooltip = newWrappedTooltip(s.pillTip(row))
		t := s.hint(row)
		hint.setText(t.text, pickerTextInk(t, pickerStateInks[pickerOK]))
		// The dialog is first sized with the text above in place; it widens if later text needs more room.
		if dialog != nil {
			dialog.Button(unison.ModalResponseOK).SetEnabled(state == pickerOK)
			growWindowToFit(dialog.Window())
		}
	}
	list.refresh = refresh
	s.addPickerRows(list, row.NodeChildren(), nil, 0)
	list.sync()
	refresh()

	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list.panel, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HSpan:  2,
		HGrab:  true,
		VGrab:  true,
	})

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	if opLabel := newOperationLabel(s.op); opLabel != nil {
		opLabel.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		panel.AddChild(opLabel)
	}
	label := unison.NewLabel()
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	label.SetTitle(row.String())
	panel.AddChild(label)
	// A choice nested within another names the containers above it, to tie it back to the answer that brought it up.
	if location := rowLocation(row); location != "" {
		label = newTruncatedLabel(location, maxContextLineLength, fonts.FieldSecondary)
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		panel.AddChild(label)
	}
	if notesCapable, hasNotes := any(row).(interface{ Notes() string }); hasNotes {
		if notes := notesCapable.Notes(); notes != "" {
			label = unison.NewLabel()
			label.Font = fonts.FieldSecondary
			label.SetTitle(notes)
			label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
			panel.AddChild(label)
		}
	}
	label = unison.NewLabel()
	label.SetTitle(tp.StringWithUnits(pickerWeightUnits(row)))
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2}))
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
	})
	panel.AddChild(label)
	progress.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	panel.AddChild(progress)
	panel.AddChild(scroll)
	panel.AddChild(hint)

	buttons := []*unison.DialogButtonInfo{
		unison.NewCancelButtonInfo(),
		{Title: i18n.Text("Override"), ResponseCode: unison.ModalResponseUserBase},
		unison.NewOKButtonInfo(),
	}
	if depth > 0 {
		buttons = slices.Insert(buttons, 0, &unison.DialogButtonInfo{
			Title:        i18n.Text("Clear Selections"),
			ResponseCode: unison.ModalResponseUserBase + 1,
		})
	}
	var err error
	if dialog, err = newPromptDialog(s.op.at(promptstep.Choice), nil, nil, panel, buttons...); err != nil {
		errs.Log(err)
		return nil, nil
	}
	overrideTip := i18n.Text("Accept the checked options whether or not they satisfy the choice, leaving any choices still to be made below them for when the template is applied")
	dialog.Button(unison.ModalResponseUserBase).Tooltip = newWrappedTooltip(overrideTip)
	if depth > 0 {
		// Clearing leaves the dialog up.
		dialog.Button(unison.ModalResponseUserBase + 1).ClickCallback = func() {
			s.clear(row)
			refresh()
		}
		// It opens centered over the dialog it is stacked on, so shift it to show that it is.
		wnd := dialog.Window()
		frame := wnd.FrameRect()
		frame.Point = frame.Point.Add(geom.NewPoint(2*unison.StdHSpacing, 3*unison.StdVSpacing))
		wnd.SetFrameRect(frame)
		wnd.EnsureOnDisplay()
	}
	refresh()
	return dialog, refresh
}

// newPickerStatePill returns a label drawn as a pill in the look of a picker state, top points below the top of its
// border, and the function that sets its state and title. It isn't drawn until that is called.
func newPickerStatePill(top float32) (pill *unison.Label, update func(state pickerState, title string)) {
	pill = unison.NewLabel()
	pill.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: top, Left: unison.StdHSpacing, Right: unison.StdHSpacing}))
	var background unison.Color
	pill.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		if pill.Drawable == nil {
			return
		}
		r := pill.ContentRect(true)
		r.Y += top
		r.Height -= top
		gc.DrawRoundedRect(r, geom.NewUniformSize(8), background.Paint(gc, r, paintstyle.Fill))
		pill.DefaultDraw(gc, r)
	}
	update = func(state pickerState, title string) {
		var img *unison.SVG
		switch state {
		case pickerOK:
			img, background = unison.CheckmarkSVG, unison.Green
		case pickerOpen:
			img, background = unison.CircledQuestionSVG, unison.RGB(138, 83, 0)
		case pickerWarning:
			img, background = unison.TriangleExclamationSVG, unison.RGB(240, 196, 25)
		default:
			img, background = svg.Not, unison.ThemeError.GetColor()
		}
		size := max(pill.Font.Baseline()-2, 6)
		pill.Drawable = &unison.DrawableSVG{SVG: img, Size: geom.NewSize(size, size)}
		pill.OnBackgroundInk = background.On()
		pill.SetTitle(title)
		pill.MarkForLayoutRecursivelyUpward()
		pill.MarkForRedraw()
	}
	return pill, update
}

// pickerList is the picker dialog's list of options, with what its rows need and the cells they are made of.
type pickerList struct {
	panel        *unison.Panel
	refresh      func()
	updates      []func()
	cells        []*unison.Panel
	sections     []*pickerSection
	columns      int
	depth        int
	pt           picker.Type
	chooseColumn bool
}

// pickerSection is a run of the list's rows under a disclosure, shown when it and every section it sits in is open.
type pickerSection struct {
	parent *pickerSection
	open   bool
}

func (sec *pickerSection) shown() bool {
	for ; sec != nil; sec = sec.parent {
		if !sec.open {
			return false
		}
	}
	return true
}

// claim records the cells added to the list since it held from of them as being in the section.
func (l *pickerList) claim(from int, sec *pickerSection) {
	for _, cell := range l.panel.Children()[from:] {
		l.cells = append(l.cells, cell)
		l.sections = append(l.sections, sec)
	}
}

// sync takes the cells of hidden sections out of the list and puts those of shown ones back in their places. Whole rows
// come and go, so the rest keep their columns.
func (l *pickerList) sync() {
	i := 0
	for j, cell := range l.cells {
		switch {
		case !l.sections[j].shown():
			if cell.Parent() == l.panel {
				cell.RemoveFromParent()
			}
			continue
		case cell.Parent() != l.panel:
			l.panel.AddChildAtIndex(cell, i)
		}
		i++
	}
	l.panel.MarkForLayoutRecursivelyUpward()
	l.panel.MarkForRedraw()
}

// addPickerRows adds a row for each of the children to the list, in the section, indent levels in. An organizing group
// becomes a header with its options beneath it, one level further in. A container picked as a unit may be opened to
// show what it holds.
func (s *pickerSession[T]) addPickerRows(list *pickerList, children []T, sec *pickerSection, indent int) {
	for _, child := range children {
		from := len(list.panel.Children())
		if gurps.IsOrganizingGroup(child) {
			inner := addPickerHeader(list, child, sec, indent)
			list.claim(from, sec)
			s.addPickerRows(list, child.NodeChildren(), inner, indent+1)
			continue
		}
		var unit *pickerSection
		var disclosure *unison.Button
		if child.HasChildren() && !gurps.IsTemplateChoiceContainer(child) {
			unit = &pickerSection{parent: sec}
			disclosure = newPickerDisclosure(list, unit, child.String(), unison.DefaultCheckBoxTheme.Font)
		}
		list.updates = append(list.updates, s.addPickerRow(list.panel, child, disclosure, list.pt, indent, list.depth,
			list.chooseColumn, list.refresh))
		list.claim(from, sec)
		if unit != nil {
			addPickerInfoRows(list, child.NodeChildren(), unit, indent+1)
		}
	}
}

// addPickerHeader adds the header of an organizing group to the list and returns the section within the given one
// that its options go in, which starts open.
func addPickerHeader[T gurps.Node[T]](list *pickerList, group T, within *pickerSection, indent int) *pickerSection {
	sec := &pickerSection{parent: within, open: true}
	name := unison.NewLabel()
	name.SetTitle(group.String())
	name.Accessibility.Role = role.Heading
	name.Accessibility.Level = indent + 1
	addPickerNameRow(list, group, indent, newPickerDisclosure(list, sec, group.String(), name.Font), name)
	return sec
}

// addPickerInfoRows adds a row naming each of the rows, and those within them, to the list, in the section, indent
// levels in. They tell what a container picked as a unit holds, so there is nothing in them to pick or change.
func addPickerInfoRows[T gurps.Node[T]](list *pickerList, rows []T, sec *pickerSection, indent int) {
	for _, row := range rows {
		from := len(list.panel.Children())
		name := unison.NewLabel()
		name.SetTitle(row.String())
		addPickerNameRow(list, row, indent, name)
		list.claim(from, sec)
		addPickerInfoRows(list, row.NodeChildren(), sec, indent+1)
	}
}

// addPickerNameRow adds a row to the list, indent levels in, holding the panels followed by the row's page reference in
// the list's first column, so the reference lines up with those of the options, and nothing in the rest.
func addPickerNameRow[T gurps.Node[T]](list *pickerList, row T, indent int, panels ...unison.Paneler) {
	if link := newPickerPageLink(pickerRowPageRef(row)); link != nil {
		panels = append(panels, link)
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  len(panels),
		HSpacing: unison.StdHSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: pickerIndent(indent)}))
	for _, one := range panels {
		panel.AddChild(one)
	}
	list.panel.AddChild(panel)
	for range list.columns - 1 {
		list.panel.AddChild(unison.NewPanel())
	}
}

// newPickerDisclosure returns the chevron that shows or hides the section of the list, named for what it holds. Space
// or Return toggles it.
func newPickerDisclosure(list *pickerList, sec *pickerSection, title string, font unison.Font) *unison.Button {
	button := unison.NewButton()
	button.HideBase = true
	button.HMargin = 0
	button.VMargin = 0
	// The button is nothing but its chevron, so its tooltip is also what a screen reader calls it.
	button.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("Show or hide %s"), title))
	size := max(font.Baseline()-2, 6)
	chevron := &unison.DrawableSVG{
		SVG:  unison.CircledChevronRightSVG,
		Size: geom.NewSize(size, size),
	}
	turn := func() {
		chevron.RotationDegrees = 0
		if sec.open {
			chevron.RotationDegrees = 90
		}
	}
	turn()
	button.Drawable = chevron
	button.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	button.ClickCallback = func() {
		sec.open = !sec.open
		turn()
		list.sync()
		if wnd := button.Window(); wnd != nil {
			growWindowToFit(wnd)
		}
	}
	button.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if (keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter) && mods&mod.NonSticky == 0 {
			button.Click()
			return true
		}
		return button.DefaultKeyDown(keyCode, mods, repeat)
	}
	addAccessibilityCallback(button, func(node *accessibility.Node) {
		node.Expandable = true
		node.Expanded = sec.open
	})
	return button
}

// pickerIndent returns how far in a row of the picker's list is set at the indent level.
func pickerIndent(indent int) float32 {
	return float32(indent) * 2 * unison.StdHSpacing
}

// pickerRowPageRef returns the page reference of the row and the text to highlight on its page, if it has one.
func pickerRowPageRef[T gurps.Node[T]](row T) (pageRef, highlight string) {
	switch actual := any(row).(type) {
	case *gurps.Trait:
		return actual.PageRef, actual.PageRefHighlight
	case *gurps.Skill:
		return actual.PageRef, actual.PageRefHighlight
	case *gurps.Spell:
		return actual.PageRef, actual.PageRefHighlight
	case *gurps.Equipment:
		return actual.PageRef, actual.PageRefHighlight
	}
	return "", ""
}

// newPickerPageLink returns the link to the first of the page references, set to line up along the right edge, or nil
// if there is none.
func newPickerPageLink(pageRef, highlight string) *unison.Label {
	pageRefs := ExtractPageReferences(pageRef)
	if len(pageRefs) == 0 {
		return nil
	}
	var tooltip string
	var icon *unison.DrawableSVG
	title, img := convertLinksForPageRef(pageRefs[0])
	if img != nil {
		title = ""
		height := unison.DefaultLinkTheme.Font.Baseline()
		icon = &unison.DrawableSVG{
			SVG:  img,
			Size: geom.NewSize(height, height).Ceil(),
		}
		tooltip = pageRefs[0]
	}
	link := newLink(title, tooltip, "", &unison.DefaultLinkTheme, func(_ unison.Paneler, _ string) {
		OpenPageReference(pageRefs[0], highlight, nil)
	})
	link.VAlign = align.Start
	link.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		HGrab:  true,
	})
	if icon != nil {
		link.Drawable = icon
	}
	if tooltip != "" {
		link.Tooltip = newWrappedTooltip(tooltip)
	}
	return link
}

// addPickerRow adds the row to the dialog's list, indent levels in and led by the disclosure if not nil, returning what
// brings it up to date with the session. refresh is called after anything in the row changes. depth is that of the
// dialog, and chooseColumn says whether it has a column for the choose button.
func (s *pickerSession[T]) addPickerRow(parent *unison.Panel, row T, disclosure *unison.Button, pt picker.Type, indent, depth int, chooseColumn bool, refresh func()) (update func()) {
	op, prompted := s.op, s.prompted
	wrapper := unison.NewPanel()
	columns := 3
	if disclosure != nil {
		columns++
		wrapper.AddChild(disclosure)
	}
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
	})
	// The wrapper fills its column so that the page reference it ends with lines up along the right edge.
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	if indent > 0 {
		wrapper.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: pickerIndent(indent)}))
	}
	parent.AddChild(wrapper)
	checkBox := unison.NewCheckBox()
	checkBox.ClickCallback = func() {
		s.chosen[row] = checkBox.State == check.On
		refresh()
	}
	wrapper.AddChild(checkBox)
	name, detail, cost := unison.NewLabel(), unison.NewLabel(), unison.NewLabel()
	text := unison.NewPanel()
	text.SetLayout(&unison.FlexLayout{Columns: 3})
	for _, label := range []*unison.Label{name, detail, cost} {
		label.Font = checkBox.Font
		// Clicking the text clicks the box, as when the text was its title.
		label.MouseDownCallback = func(geom.Point, int, int, mod.Modifiers) bool { return true }
		label.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
			if where.In(label.ContentRect(false)) {
				checkBox.Click()
			}
			return true
		}
		text.AddChild(label)
	}
	name.SetTitle(row.String())
	wrapper.AddChild(text)
	var onClick func()
	var editTooltip string
	details := make([]*unison.Label, 0, len(pickerRowDetailHeaders(row)))
	switch actual := any(row).(type) {
	case *gurps.Trait:
		if actual.IsLeveled() {
			onClick = func() { pickerRowLevelEditor(op, actual, refresh) }
			editTooltip = i18n.Text("Edit level")
		}
	case *gurps.Skill:
		if !actual.Container() {
			onClick = func() { pickerRowPointEditor(op, actual, refresh) }
			editTooltip = i18n.Text("Edit points")
		}
	case *gurps.Spell:
		if !actual.Container() {
			onClick = func() { pickerRowPointEditor(op, actual, refresh) }
			editTooltip = i18n.Text("Edit points")
		}
	case *gurps.Equipment:
		// A choice made by value or weight may take more than one of an option, so its quantity may be set while
		// picking, if it has one of its own to set.
		if (pt == picker.Value || pt == picker.Weight) && !actual.IsGroup() {
			onClick = func() { pickerRowQuantityEditor(op, actual, refresh) }
			editTooltip = i18n.Text("Edit quantity")
		}
	}
	if link := newPickerPageLink(pickerRowPageRef(row)); link != nil {
		wrapper.AddChild(link)
	}
	for range pickerRowDetailHeaders(row) {
		label := unison.NewLabel()
		label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
		parent.AddChild(label)
		details = append(details, label)
	}
	if onClick == nil {
		label := unison.NewLabel()
		label.SetTitle(" ")
		parent.AddChild(label)
	} else {
		button := NewSVGButtonForFont(svg.Edit, checkBox.Font, -2)
		button.Tooltip = newWrappedTooltip(editTooltip)
		button.ClickCallback = onClick
		parent.AddChild(button)
	}
	var choose *unison.Button
	if chooseColumn {
		isChoice := gurps.IsTemplateChoiceContainer(row)
		if isChoice || (prompted && len(s.modTargets(row)) != 0) || len(s.nestedChoices(row)) != 0 {
			choose = NewSVGButtonForFont(svg.Settings, checkBox.Font, -2)
			choose.ClickCallback = func() {
				if isChoice {
					s.choosePicks(row, depth)
				} else {
					s.chooseWithin(row, depth)
				}
				refresh()
			}
			parent.AddChild(choose)
		} else {
			parent.AddChild(unison.NewPanel())
		}
	}
	return func() {
		checkBox.State = check.FromBool(s.chosen[row])
		setPickerText(detail, s.detail(row), unison.ThemeOnSurface)
		if pt == picker.Points || pt == picker.Count {
			setPickerText(cost, s.cost(row, picker.Points), unison.ThemeOnSurface)
		}
		checkBox.Accessibility.Name = name.String() + detail.String() + cost.String()
		if eqp, ok := any(row).(*gurps.Equipment); ok && len(details) == 3 {
			details[0].SetTitle(pickerRowQuantity(eqp))
			setPickerText(details[1], s.cost(row, picker.Value), unison.ThemeOnSurface)
			setPickerText(details[2], s.cost(row, picker.Weight), unison.ThemeOnSurface)
		}
		if choose != nil {
			s.updateChooseButton(choose, row)
		}
		checkBox.MarkForLayoutRecursivelyUpward()
		checkBox.MarkForRedraw()
	}
}

// updateChooseButton shows whether the row still has choices to make, and whether it was answered.
func (s *pickerSession[T]) updateChooseButton(button *unison.Button, row T) {
	var name, tip string
	var open bool
	if gurps.IsTemplateChoiceContainer(row) {
		name, open = i18n.Text("Choose from %s"), !s.resolved(row)
		tip = i18n.Text("Its picks are still to be made or miss its rule; choose them now to fix its cost")
		if s.pickerAnswered[row] {
			name = i18n.Text("Change the picks in %s")
		}
	} else {
		name, open = i18n.Text("Choose modifiers for %s"), !s.resolved(row)
		tip = i18n.Text("Its cost depends on modifier choices still to be made; choose them now to fix it")
		switch {
		case len(s.nestedChoices(row)) != 0:
			name = i18n.Text("Choose within %s")
			tip = i18n.Text("Its cost depends on choices within it still to be made; choose them now to fix it")
		case s.modsAnswered[row]:
			name = i18n.Text("Change the modifiers for %s")
		}
	}
	name = fmt.Sprintf(name, row.String())
	button.Accessibility.Name = name
	if open {
		button.OnBackgroundInk = unison.ThemeWarning
		button.Tooltip = newWrappedTooltip(tip)
	} else {
		button.OnBackgroundInk = unison.DefaultButtonTheme.OnBackgroundInk
		button.Tooltip = newWrappedTooltip(name)
	}
	button.MarkForRedraw()
}

// pickerRowDetailHeaders returns the headings of the columns of details shown for each option of the picker container,
// if its options have any. Only equipment has them: an option's quantity, and the value and weight of all of it.
func pickerRowDetailHeaders[T gurps.Node[T]](container T) []string {
	if _, ok := any(container).(*gurps.Equipment); ok {
		return []string{i18n.Text("Qty"), i18n.Text("Value"), i18n.Text("Weight")}
	}
	return nil
}

// pickerRowQuantity returns the quantity shown for an option, if it has one of its own.
func pickerRowQuantity(eqp *gurps.Equipment) string {
	if !eqp.IsGroup() {
		return eqp.Quantity.Comma()
	}
	return ""
}

// pickerStateInks holds the color of text telling of each state. A warning's is a dark yellow, as the pill's is too
// light to read as text.
var pickerStateInks = [...]unison.Ink{
	pickerOK:      unison.Green,
	pickerOpen:    unison.ThemeWarning,
	pickerWarning: &unison.ThemeColor{Light: unison.RGB(122, 92, 0), Dark: unison.RGB(240, 196, 25)},
	pickerError:   unison.ThemeError,
}

// setPickerText shows the text on the label with its tooltip, colored by its state, or plain when that is OK.
func setPickerText(label *unison.Label, t pickerText, plain unison.Ink) {
	label.OnBackgroundInk = pickerTextInk(t, plain)
	label.SetTitle(t.text)
	label.Tooltip = nil
	if t.tip != "" {
		label.Tooltip = newWrappedTooltip(t.tip)
	}
}

// pickerTextInk returns the color of the text, by its state, or plain when that is OK.
func pickerTextInk(t pickerText, plain unison.Ink) unison.Ink {
	if t.state != pickerOK {
		return pickerStateInks[t.state]
	}
	return plain
}

// growWindowToFit widens the window, as far as its display allows, when its content has come to want more width than it
// has. Its height is left alone, as the list scrolls, so the list keeps its place.
func growWindowToFit(wnd *unison.Window) {
	_, pref, _ := wnd.Content().Sizes(geom.Size{})
	r := wnd.ContentRect()
	width := pref.Width
	if d := wnd.Display(); d != nil {
		width = min(width, d.Usable.Width-(wnd.FrameRect().Width-r.Width))
	}
	if width > r.Width {
		r.Width = width
		wnd.SetContentRect(r)
		wnd.EnsureOnDisplay()
	}
}

// pickerWeightUnits returns the units the picker dialog shows weights in: those of the sheet the row being picked from
// is headed for, since it is already owned by that sheet while the dialog is shown, or the default ones otherwise.
func pickerWeightUnits[T gurps.Node[T]](row T) fxp.WeightUnit {
	var entity *gurps.Entity
	if !xreflect.IsNil(row) {
		entity = gurps.EntityFromNode(row)
	}
	return gurps.SheetSettingsFor(entity).DefaultWeightUnits
}

// formatPickerTotal renders the running total of the options picked, as a value or a weight when the choice is made by
// one.
func formatPickerTotal[T gurps.Node[T]](row T, pt picker.Type, total gurps.NumericRange) string {
	switch pt {
	case picker.Value:
		return "$" + gurps.FormatValueRange(total, fxp.Int.Comma)
	case picker.Weight:
		return gurps.FormatWeightRange(total, pickerWeightUnits(row).Format)
	default:
		return total.Comma()
	}
}

// pickerRowQuantityEditor asks for a new quantity of an option of a choice made by value or weight.
func pickerRowQuantityEditor(op promptOperation, eqp *gurps.Equipment, callback func()) {
	quantity := eqp.Quantity
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("%s Quantity"), eqp.String()))
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return quantity },
		func(value fxp.Int) { quantity = value }, fxp.One, fxp.Max-1, false, false))
	dialog, err := newPromptDialog(op.at(promptstep.Quantity), nil, nil, panel, unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	eqp.Quantity = quantity
	callback()
}

// pointsText returns the points, as in "5 points" or "1 point".
func pointsText(points gurps.NumericRange) string {
	if value, settled := points.Settled(); settled && value == fxp.One {
		return points.Comma() + " " + i18n.Text("point")
	}
	return points.Comma() + " " + i18n.Text("points")
}

func pickerRowLevelEditor(op promptOperation, trait *gurps.Trait, callback func()) {
	levels := trait.Levels
	maximum := trait.ResolvedMaxLevels()
	fieldMax := fxp.MaxBasePoints
	if maximum > 0 {
		fieldMax = maximum
		levels = min(levels, maximum)
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	if maximum > 0 {
		label.SetTitle(fmt.Sprintf(i18n.Text("%s Level (max %s)"), trait.Kind(), maximum.String()))
	} else {
		label.SetTitle(fmt.Sprintf(i18n.Text("%s Level"), trait.Kind()))
	}
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return levels },
		func(value fxp.Int) { levels = value }, 0, fieldMax, false, false))
	dialog, err := newPromptDialog(op.at(promptstep.Level), nil, nil, panel, unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	trait.Levels = levels
	callback()
}

type pickerRowPointEditorTypes[T gurps.Node[T]] interface {
	gurps.Node[T]
	gurps.RawPointsAdjuster
}

func pickerRowPointEditor[T pickerRowPointEditorTypes[T]](op promptOperation, node T, callback func()) {
	points := node.RawPoints()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("%s Points"), node.Kind()))
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return points },
		func(value fxp.Int) { points = value }, 0, fxp.MaxBasePoints, false, false))
	dialog, err := newPromptDialog(op.at(promptstep.Points), nil, nil, panel, unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	node.SetRawPoints(points)
	callback()
}
