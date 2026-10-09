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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// Rebuildable defines the methods a rebuildable panel should provide.
type Rebuildable interface {
	unison.Paneler
	fmt.Stringer
	Rebuild(full bool)
}

// Owned defines the methods a value owned by a Rebuildable should have.
type Owned interface {
	Owner() Rebuildable
}

// FindOwner walks up the panel's lineage and returns the first owner that is a T, or the zero value if there is none.
func FindOwner[T Rebuildable](panel *unison.Panel) T {
	for panel != nil {
		if owned, ok := any(panel.Self).(Owned); ok && !xreflect.IsNil(owned) && !xreflect.IsNil(owned.Owner()) {
			if owner, ok2 := owned.Owner().AsPanel().Self.(T); ok2 {
				return owner
			}
			panel = owned.Owner().AsPanel()
		} else {
			panel = panel.Parent()
		}
	}
	var zero T
	return zero
}

// HasOwner reports whether the panel has an owner that is a T.
func HasOwner[T Rebuildable](panel *unison.Panel) bool {
	return !xreflect.IsNil(FindOwner[T](panel))
}

// Targeted defines the methods a value with a node target should have.
type Targeted[N gurps.Node[N]] interface {
	Target() N
}

// FindTarget walks up the panel's lineage and returns the target of the first Targeted[T] found, or the zero value if
// there is none.
func FindTarget[T gurps.Node[T]](panel *unison.Panel) T {
	for panel != nil {
		if targeted, ok := any(panel.Self).(Targeted[T]); ok {
			return targeted.Target()
		}
		panel = panel.Parent()
	}
	var zero T
	return zero
}

// Syncer is implemented by objects that can update their UI state from their model.
type Syncer interface {
	Sync()
}

// DeepSync calls Sync on the panel and each of its descendants that is a Syncer, depth-first.
func DeepSync(panel unison.Paneler) {
	p := panel.AsPanel()
	for _, child := range p.Children() {
		DeepSync(child)
	}
	if syncer, ok := p.Self.(Syncer); ok {
		syncer.Sync()
	}
}

// ModifiableRoot marks the root of a modifiable tree of components, typically a Dockable.
type ModifiableRoot interface {
	MarkModified(src unison.Paneler)
}

// MarkModified discards the global resolve cache, then calls MarkModified on the nearest ModifiableRoot at or above the
// panel, if there is one.
func MarkModified(panel unison.Paneler) {
	gurps.DiscardGlobalResolveCache()
	p := panel.AsPanel()
	for p != nil {
		if modifiable, ok := p.Self.(ModifiableRoot); ok {
			modifiable.MarkModified(panel)
			break
		}
		p = p.Parent()
	}
}

// modificationTimestampBumper is implemented by those owners that record when their data was last changed. Marking such
// an owner as modified bumps that timestamp, but rebuilding it doesn't, so anything that rebuilds in place of marking
// as modified has to ask for the bump itself (see rebuildAsModified). Not every owner has one -- a template's
// modification time is whatever the file system says it is -- hence the optional interface.
type modificationTimestampBumper interface {
	bumpModificationTimestamp()
}

var (
	_ modificationTimestampBumper = &Sheet{}
	_ modificationTimestampBumper = &LootSheet{}
)

// rebuildAsModified reports an edit to its owner by rebuilding the owner instead of marking it as modified. For every
// kind of owner a rebuild does everything marking as modified does -- recalculating the entity, re-syncing every table,
// refreshing the search results and restoring the focus and scroll position -- except bump the modification timestamp,
// which is done here, before the rebuild, since the panel showing it only picks up the new value when synced. A rebuild
// is needed when an edit changes more than the rows it touched: a sheet shows its melee weapon, ranged weapon, reaction
// and conditional modifier lists only while they have something in them, and a table's columns, which come from what
// its rows use, can only change by building a new table. Doing both would repeat the whole update, which on a sheet
// with hundreds of rows is the entire cost of the edit. A nil owner, typed or otherwise, is ignored.
func rebuildAsModified(owner Rebuildable, full bool) {
	if xreflect.IsNil(owner) {
		return
	}
	if bumper, ok := owner.(modificationTimestampBumper); ok {
		bumper.bumpModificationTimestamp()
	}
	owner.Rebuild(full)
}

func addNameLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Name"), "", fieldData)
}

func addSpecializationLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Required Specialization"), "", fieldData)
}

func addOptionalSpecializationLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Optional Specialization"), "", fieldData)
}

func addPageRefLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Page Reference"), gurps.PageRefTooltip(), fieldData)
}

func addPageRefHighlightLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Page Highlight"),
		i18n.Text(`A snippet of text to highlight on the page when opening the page reference; only needed if the default behavior isn't highlighting the expected area`),
		fieldData)
}

func addNotesLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndScriptField(parent, nil, "", i18n.Text("Notes"),
		i18n.Text("These notes may have scripts embedded in them by wrapping each script in <script>your script goes here</script> tags."),
		func() string { return *fieldData },
		func(value string) {
			*fieldData = value
			parent.MarkForLayoutAndRedraw()
			MarkModified(parent)
		}, true)
}

func addVTTNotesLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndMultiLineStringField(parent, i18n.Text("VTT Notes"),
		i18n.Text("Any notes for VTT use; see the instructions for your VTT to determine if/how these can be used"),
		fieldData)
}

func addUserDescLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndMultiLineStringField(parent, i18n.Text("User Description"),
		i18n.Text("Additional notes for your own reference. These only exist in character sheets and will be removed if transferred to a data list or template"),
		fieldData)
}

func addTechLevelRequired(parent *unison.Panel, fieldData **string, ownerIsSheet bool) {
	tl := i18n.Text("Tech Level")
	var field *StringField
	wrapper, label := addFlowWrapper(parent, tl, 2)
	field = NewStringField(nil, "", tl, func() string {
		if *fieldData == nil {
			return ""
		}
		return **fieldData
	}, func(value string) {
		if *fieldData == nil {
			return
		}
		**fieldData = value
		MarkModified(parent)
	})
	tip := gurps.TechLevelInfo()
	if !ownerIsSheet {
		tip = xstrings.Wrap("", i18n.Text("Leave field blank to auto-populate with the character's TL when added to a character sheet."), 60) + "\n\n" + tip
	}
	field.Tooltip = newWrappedTooltip(tip)
	if *fieldData == nil {
		field.SetEnabled(false)
	}
	field.SetMinimumTextWidthUsing("12^")
	field.Accessibility.LabeledBy = label
	wrapper.AddChild(field)
	parent = wrapper
	last := *fieldData
	required := last != nil
	parent.AddChild(NewCheckBox(nil, "", i18n.Text("Required"),
		func() check.Enum { return check.FromBool(required) },
		func(state check.Enum) {
			if required = state == check.On; required {
				if last == nil {
					var data string
					last = &data
				}
				*fieldData = last
				if field != nil {
					field.SetEnabled(true)
				}
			} else {
				last = *fieldData
				*fieldData = nil
				if field != nil {
					field.SetEnabled(false)
				}
			}
		}))
}

func addAttributeChoicePopup(parent *unison.Panel, entity *gurps.Entity, prefix string, fieldData *string, flags gurps.AttributeFlags) *unison.PopupMenu[*gurps.AttributeChoice] {
	choices, current := gurps.AttributeChoices(entity, prefix, flags, *fieldData)
	popup := addPopup(parent, choices, &current)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[*gurps.AttributeChoice]) {
		if choice, ok := p.Selected(); ok {
			*fieldData = choice.Key
			MarkModified(parent)
		}
	}
	return popup
}

func addDifficultyLabelAndFields(parent *unison.Panel, entity *gurps.Entity, attrDiff *gurps.AttributeDifficulty) {
	wrapper, label := addFlowWrapper(parent, i18n.Text("Difficulty"), 3)
	addAttributeChoicePopup(wrapper, entity, "", &attrDiff.Attribute, gurps.TenFlag).Accessibility.LabeledBy = label
	wrapper.AddChild(NewFieldTrailingLabel("/", false))
	addDifficultyLevelPopup(wrapper, &attrDiff.Difficulty)
}

// addDifficultyLevelPopup adds the popup for the level half of a difficulty, naming it for a screen reader, since the
// "/" label that separates it from the attribute half would otherwise be taken as its name.
func addDifficultyLevelPopup(parent *unison.Panel, level *difficulty.Level) *unison.PopupMenu[difficulty.Level] {
	popup := addPopup(parent, difficulty.Levels, level)
	popup.Accessibility.Name = i18n.Text("Difficulty Level")
	return popup
}

func addTagsLabelAndField(parent *unison.Panel, fieldData *[]string) {
	addLabelAndListField(parent, i18n.Text("Tags"), i18n.Text("tags"), fieldData)
}

func addLabelAndListField(parent *unison.Panel, labelText, pluralForTooltip string, fieldData *[]string) {
	get, set := pointerAccessors(parent, fieldData)
	addMultiLineStringFieldWith(parent, labelText, i18n.Text("Separate multiple %s with commas", pluralForTooltip),
		func() string { return gurps.CombineTags(get()) },
		func(value string) { set(gurps.ExtractTags(value)) })
}

func addLabelAndStringField(parent *unison.Panel, labelText, tooltip string, fieldData *string) *StringField {
	addLabel(parent, labelText, tooltip)
	return addStringField(parent, labelText, tooltip, fieldData)
}

func newMarkdownTooltip(text, workingDir string) *unison.Panel {
	tip := unison.NewTooltipBase()
	tip.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	m := unison.NewMarkdown(false)
	m.StripBottomEmptyMargin = true
	if workingDir != "" {
		m.ClientData()[WorkingDirKey] = workingDir
	}
	adjustMarkdownThemeForPage(m, unison.DefaultTooltipTheme.Label.Font)
	m.OnBackgroundInk = unison.ThemeOnTooltip
	m.SetContent(markdownHardLineBreaks(text), 500)
	tip.AddChild(m)
	return tip
}

// markdownHardLineBreaks converts single newlines into Markdown hard line breaks so that multi-line tooltip content
// renders one line per newline; the renderer otherwise treats a single newline as a soft break and collapses it into a
// space. Blank lines and other block constructs are unaffected, since a hard break before a blank line is ignored.
func markdownHardLineBreaks(text string) string {
	return strings.ReplaceAll(text, "\n", "  \n")
}

func newWrappedTooltip(tooltip string) *unison.Panel {
	return unison.NewTooltipWithText(wrapTextForTooltip(tooltip))
}

func newWrappedTooltipWithSecondaryText(primary, secondary string) *unison.Panel {
	return unison.NewTooltipWithSecondaryText(wrapTextForTooltip(primary), wrapTextForTooltip(secondary))
}

func wrapTextForTooltip(tooltip string) string {
	return strings.ReplaceAll(xstrings.Wrap("", strings.ReplaceAll(tooltip, " ", "␣"), 80), "␣", " ")
}

// pointerAccessors returns accessors for the value the pointer refers to. The setter stores the value and then marks
// the parent modified.
func pointerAccessors[T any](parent *unison.Panel, fieldData *T) (get func() T, set func(T)) {
	return func() T { return *fieldData },
		func(value T) {
			*fieldData = value
			MarkModified(parent)
		}
}

// baseTooltipSetter is implemented by fields that temporarily replace their tooltip with another one, such as
// NumericField, which shows an explanation of why its content is invalid while that is the case.
type baseTooltipSetter interface {
	SetBaseTooltip(tip *unison.Panel)
}

// setFieldTooltip installs a tooltip on a field, as its base tooltip when the field has one, so that a temporary
// replacement -- such as the message a NumericField shows while its content is invalid -- isn't clobbered.
func setFieldTooltip(field unison.Paneler, tip *unison.Panel) {
	if setter, ok := field.(baseTooltipSetter); ok {
		setter.SetBaseTooltip(tip)
		return
	}
	field.AsPanel().Tooltip = tip
}

// installField gives the field the tooltip, if there is one, and adds it to the parent.
func installField[F unison.Paneler](parent *unison.Panel, field F, tooltip string) F {
	if tooltip != "" {
		setFieldTooltip(field, newWrappedTooltip(tooltip))
	}
	parent.AddChild(field)
	return field
}

// installSizedField is installField for a text field whose minimum width should come from the prototype text, when
// there is one.
func installSizedField(parent *unison.Panel, field *StringField, tooltip, prototype string) *StringField {
	if prototype != "" {
		field.SetMinimumTextWidthUsing(prototype)
	}
	return installField(parent, field, tooltip)
}

// addLabelAndTargetedStringField adds a label and a single-line field that edits the value the accessors reach and
// records its undo through the target manager. The label's text is also the undo title, and the field's minimum width
// comes from the prototype text.
func addLabelAndTargetedStringField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip, prototype string, get func() string, set func(string)) *StringField {
	addLabel(parent, labelText, "")
	return installSizedField(parent, NewStringField(targetMgr, targetKey, labelText, get, set), tooltip, prototype)
}

// addLabelAndTargetedMultiLineStringField is addLabelAndTargetedStringField for a field that may hold several lines.
func addLabelAndTargetedMultiLineStringField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip, prototype string, get func() string, set func(string)) *StringField {
	addLabel(parent, labelText, "")
	return installSizedField(parent, NewMultiLineStringField(targetMgr, targetKey, labelText, get, set), tooltip,
		prototype)
}

// addLabelAndTargetedIntegerField adds a label and an integer field that edits the value the accessors reach and
// records its undo through the target manager. The label's text is also the undo title.
func addLabelAndTargetedIntegerField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, get func() int, set func(int), minValue, maxValue int, forceSign bool) *IntegerField {
	addLabel(parent, labelText, "")
	return installField(parent, NewIntegerField(targetMgr, targetKey, labelText, get, set, minValue, maxValue, forceSign,
		false), tooltip)
}

// addLabelAndTargetedPopup adds a label and a popup that edits the value the accessors reach and records its undo
// through the target manager. The label's text is also the undo title.
func addLabelAndTargetedPopup[T comparable](parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, get func() T, set func(T), items ...T) *Popup[T] {
	addLabel(parent, labelText, "")
	return installField(parent, NewPopup(targetMgr, targetKey, labelText, get, set, items...), tooltip)
}

// addLabelAndScriptField adds a label and a script field, as addScriptField does, whose undo title is the label's
// text.
func addLabelAndScriptField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, get func() string, set func(string), includeMarkdownButton bool) *StringField {
	label := addLabel(parent, labelText, "")
	field := addScriptField(parent, targetMgr, targetKey, labelText, tooltip, get, set, includeMarkdownButton)
	// The field shares a wrapper with its help buttons, so the label is not its sibling and has to be pointed at.
	field.Accessibility.LabeledBy = label
	return field
}

func addStringField(parent *unison.Panel, labelText, tooltip string, fieldData *string) *StringField {
	get, set := pointerAccessors(parent, fieldData)
	return installField(parent, NewStringField(nil, "", labelText, get, set), tooltip)
}

func addLabelAndMultiLineStringField(parent *unison.Panel, labelText, tooltip string, fieldData *string) {
	get, set := pointerAccessors(parent, fieldData)
	addMultiLineStringFieldWith(parent, labelText, tooltip, get, set)
}

// addMultiLineStringFieldWith adds a label and a multi-line field that edits the value the accessors reach. Since the
// field's height follows its text, the parent is laid out again after each change.
func addMultiLineStringFieldWith(parent *unison.Panel, labelText, tooltip string, get func() string, set func(string)) *StringField {
	addLabel(parent, labelText, tooltip)
	field := NewMultiLineStringField(nil, "", labelText, get, func(value string) {
		set(value)
		parent.MarkForLayoutAndRedraw()
	})
	field.AutoScroll = false
	return installField(parent, field, tooltip)
}

func addLabelAndIntegerField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *int, minValue, maxValue int) *IntegerField {
	addLabel(parent, labelText, tooltip)
	return addIntegerField(parent, targetMgr, targetKey, labelText, tooltip, fieldData, minValue, maxValue)
}

func addIntegerField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *int, minValue, maxValue int) *IntegerField {
	get, set := pointerAccessors(parent, fieldData)
	return installField(parent, NewIntegerField(targetMgr, targetKey, labelText, get, set, minValue, maxValue, false, false),
		tooltip)
}

// addLabel adds a leading label for a field and returns it, for a caller whose field will not be the label's sibling
// to point the field at with Accessibility.LabeledBy.
func addLabel(parent *unison.Panel, labelText, tooltip string) *unison.Label {
	return installField(parent, NewFieldLeadingLabel(labelText, false), tooltip)
}

func addLabelAndDecimalField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *fxp.Int, minValue, maxValue fxp.Int) *DecimalField {
	addLabel(parent, labelText, tooltip)
	return addDecimalField(parent, targetMgr, targetKey, labelText, tooltip, fieldData, minValue, maxValue, false)
}

func addDecimalField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *fxp.Int, minValue, maxValue fxp.Int, forceSign bool) *DecimalField {
	get, set := pointerAccessors(parent, fieldData)
	return installField(parent, NewDecimalField(targetMgr, targetKey, labelText, get, set, minValue, maxValue, forceSign,
		false), tooltip)
}

func addCheckBox(parent *unison.Panel, labelText string, fieldData *bool) *CheckBox {
	checkBox := NewCheckBox(nil, "", labelText,
		func() check.Enum { return check.FromBool(*fieldData) },
		func(state check.Enum) { *fieldData = state == check.On })
	parent.AddChild(checkBox)
	return checkBox
}

// addSwitchedOnCheckBox adds the "Switched On" checkbox used by the editors of items that can hold switchable features,
// preceded by an empty panel to keep it in the field column of a two-column layout.
func addSwitchedOnCheckBox(parent *unison.Panel, fieldData *bool) *CheckBox {
	parent.AddChild(unison.NewPanel())
	checkBox := addCheckBox(parent, i18n.Text("Switched On"), fieldData)
	checkBox.Tooltip = newWrappedTooltip(gurps.SwitchedOnTooltip())
	return checkBox
}

func addInvertedCheckBox(parent *unison.Panel, labelText string, fieldData *bool) *CheckBox {
	checkBox := NewCheckBox(nil, "", labelText,
		func() check.Enum { return check.FromBool(!*fieldData) },
		func(state check.Enum) { *fieldData = state == check.Off })
	parent.AddChild(checkBox)
	return checkBox
}

// addFlowWrapper adds a leading label and, beside it, a wrapper laid out in the given number of columns for the
// controls the label describes. It returns both: the label is not a sibling of anything placed in the wrapper, so the
// control it names should set it as its Accessibility.LabeledBy (see labelControl), or be named some other way when
// the label text is empty.
func addFlowWrapper(parent *unison.Panel, labelText string, count int) (wrapper *unison.Panel, label *unison.Label) {
	label = NewFieldLeadingLabel(labelText, false)
	parent.AddChild(label)
	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  count,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		VAlign:   align.Middle,
	})
	parent.AddChild(wrapper)
	return wrapper, label
}

// addFillWrapper is addFlowWrapper for a wrapper that fills the width it is given.
func addFillWrapper(parent *unison.Panel, labelText string, count int) (wrapper *unison.Panel, label *unison.Label) {
	wrapper, label = addFlowWrapper(parent, labelText, count)
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	return wrapper, label
}

// labelControl points a control at the label that names it, for a control that is not the label's sibling -- one placed
// in the wrapper addFlowWrapper adds beside the label, say. A label with no text names nothing and is left out of it.
func labelControl[C unison.Paneler](control C, label *unison.Label) C {
	if label != nil && label.String() != "" {
		control.AsPanel().Accessibility.LabeledBy = label
	}
	return control
}

func addLabelAndPopup[T comparable](parent *unison.Panel, labelText, tooltip string, choices []T, fieldData *T) *unison.PopupMenu[T] {
	addLabel(parent, labelText, tooltip)
	return addPopup(parent, choices, fieldData)
}

func addPopup[T comparable](parent *unison.Panel, choices []T, fieldData *T) *unison.PopupMenu[T] {
	if fieldData != nil && len(choices) > 0 && !slices.Contains(choices, *fieldData) {
		*fieldData = choices[0]
	}
	popup := newPopupMenu(choices, *fieldData, func(item T) {
		*fieldData = item
		MarkModified(parent)
	})
	parent.AddChild(popup)
	return popup
}

// newPopupMenu creates a popup menu offering the items, with current selected, that hands each choice the user makes to
// onSelect. Selecting current does not call onSelect.
func newPopupMenu[T comparable](items []T, current T, onSelect func(T)) *unison.PopupMenu[T] {
	popup := unison.NewPopupMenu[T]()
	popup.AddItem(items...)
	installPopupSelection(popup, current, onSelect)
	return popup
}

// installPopupSelection selects current in the popup, then installs a selection callback that hands each choice the
// user makes to onSelect. The selection is made before the callback is installed, since selecting an item calls it.
func installPopupSelection[T comparable](popup *unison.PopupMenu[T], current T, onSelect func(T)) {
	popup.Select(current)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[T]) {
		if item, ok := p.Selected(); ok {
			onSelect(item)
		}
	}
}

// AdjustFieldBlank disables the field and paints over it in its background ink when blank, so that it shows nothing,
// and restores it otherwise.
func AdjustFieldBlank(field unison.Paneler, blank bool) {
	panel := field.AsPanel()
	panel.SetEnabled(!blank)
	if blank {
		panel.DrawOverCallback = func(gc *unison.Canvas, _ geom.Rect) {
			var ink unison.Ink
			if f, ok := panel.Self.(*unison.Field); ok {
				ink = f.BackgroundInk
			} else {
				ink = unison.DefaultFieldTheme.BackgroundInk
			}
			r := panel.ContentRect(false)
			gc.DrawRect(r, ink.Paint(gc, r, paintstyle.Fill))
		}
	} else {
		panel.DrawOverCallback = nil
	}
}

// AdjustPopupBlank disables the popup and paints over it in its background ink when blank, so that it shows nothing,
// and restores it otherwise.
func AdjustPopupBlank[T comparable](popup *unison.PopupMenu[T], blank bool) {
	popup.SetEnabled(!blank)
	if blank {
		popup.DrawOverCallback = func(gc *unison.Canvas, _ geom.Rect) {
			unison.DrawRoundedRectBase(gc, popup.ContentRect(false), popup.CornerRadius, 1, popup.BackgroundInk, popup.EdgeInk)
		}
	} else {
		popup.DrawOverCallback = nil
	}
}

// criteriaTitles returns what a criteria's two controls are called, from the subject they qualify: the comparison
// popup's accessible name, and the qualifier field's undo title, which also serves as its accessible name. The
// controls sit in a row that reads as a sentence, with nothing before either that could name it.
func criteriaTitles(subject string) (comparisonName, qualifierTitle string) {
	return i18n.Text("%s Comparison", subject), i18n.Text("%s Qualifier", subject)
}

// newCriteriaPanel adds a two-column panel to the parent for a criteria's comparison popup and qualifier field,
// spanning hSpan of the parent's columns and growing to fill them. When includeEmptyFiller is true, an empty panel is
// added ahead of it to occupy the parent's first column.
func newCriteriaPanel(parent *unison.Panel, hSpan int, includeEmptyFiller bool) *unison.Panel {
	if includeEmptyFiller {
		parent.AddChild(unison.NewPanel())
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		VAlign:   align.Middle,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hSpan,
		HAlign: align.Fill,
		HGrab:  true,
	})
	parent.AddChild(panel)
	return panel
}

// newComparisonPopup creates the popup menu for a criteria's comparison, named for a screen reader as given and
// offering the choices in order with the one at selectedIndex chosen. No selection callback is installed, since
// installing one first would have it called by the initial selection; the caller adds its own afterwards.
func newComparisonPopup(name string, choices []string, selectedIndex int) *unison.PopupMenu[string] {
	popup := unison.NewPopupMenu[string]()
	popup.Accessibility.Name = name
	popup.AddItem(choices...)
	popup.SelectIndex(selectedIndex)
	return popup
}

// addStringCriteriaPanel adds a text criteria's comparison popup and qualifier field, titled for the subject they
// qualify; see criteriaTitles.
func addStringCriteriaPanel(parent *unison.Panel, prefix, notPrefix, subject string, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	panel := newCriteriaPanel(parent, hSpan, includeEmptyFiller)
	var criteriaField *StringField
	comparisonName, undoTitle := criteriaTitles(subject)
	popup := newComparisonPopup(comparisonName, criteria.PrefixedStringComparisonChoices(prefix, notPrefix),
		int(strCriteria.Compare.EnsureValid()))
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		strCriteria.Compare = criteria.StringComparisons[p.SelectedIndex()]
		AdjustFieldBlank(criteriaField, strCriteria.IsZero())
		MarkModified(panel)
	}
	panel.AddChild(popup)
	criteriaField = addStringField(panel, undoTitle, "", &strCriteria.Qualifier)
	AdjustFieldBlank(criteriaField, strCriteria.IsZero())
	return popup, criteriaField
}

func addScriptField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, undoTitle, tooltip string, get func() string, set func(string), includeMarkdownButton bool) *StringField {
	var list []unison.Paneler
	field := NewMultiLineStringField(targetMgr, targetKey, undoTitle, get, set)
	field.AutoScroll = false
	field.SetMinimumTextWidthUsing("floor($basic_speed)")
	field.Tooltip = newWrappedTooltip(tooltip)
	list = append(list, field)
	scriptHelpButton := unison.NewSVGButton(svg.Script)
	scriptHelpButton.ClickCallback = func() { HandleLink(nil, "md:User%20Guide/Scripting%20Guide") }
	scriptHelpButton.Tooltip = newWrappedTooltip(i18n.Text("Scripting Guide"))
	list = append(list, scriptHelpButton)
	if includeMarkdownButton {
		list = append(list, NewMarkdownGuideButton())
	}
	parent.AddChild(WrapWithSpan(1, list...))
	return field
}

// WrapWithSpan wraps the children in a single panel that asks to fill span columns of its parent's layout.
func WrapWithSpan(span int, children ...unison.Paneler) *unison.Panel {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(children),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  span,
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	for _, child := range children {
		wrapper.AddChild(child)
	}
	return wrapper
}

// NewSVGButtonForFont creates a new SVG button with the given font and a size adjustment.
func NewSVGButtonForFont(svgData *unison.SVG, font unison.Font, sizeAdjust float32) *unison.Button {
	b := unison.NewButton()
	b.ButtonTheme = unison.DefaultButtonTheme
	b.Font = font
	b.DrawableOnlyVMargin = 1
	b.DrawableOnlyHMargin = 1
	b.HideBase = true
	baseline := font.Baseline() + sizeAdjust
	b.Drawable = &unison.DrawableSVG{
		SVG:  svgData,
		Size: geom.NewSize(baseline, baseline).Ceil(),
	}
	return b
}

// NewMarkdownGuideButton creates a button that links to the markdown guide.
func NewMarkdownGuideButton() *unison.Button {
	button := unison.NewSVGButton(svg.MarkdownFile)
	button.ClickCallback = func() { HandleLink(nil, "md:User%20Guide/Markdown%20Guide") }
	button.Tooltip = newWrappedTooltip(i18n.Text("Markdown Guide"))
	return button
}

// newApplyCancelButtons adds the Apply Changes and Discard Changes buttons to the toolbar and returns them, disabled
// until there is something to apply. Clicking apply calls the given function and, when it reports success, closes the
// editor without its usual prompt, which is also all that cancel does. showKeys adds the keyboard shortcuts the editors
// bind to the buttons' tooltips.
func newApplyCancelButtons(toolbar *unison.Panel, showKeys bool, apply func() bool, closeWithoutPrompt func()) (applyButton, cancelButton *unison.Button) {
	applyText := i18n.Text("Apply Changes")
	cancelText := i18n.Text("Discard Changes")
	applyButton = unison.NewSVGButton(unison.CheckmarkSVG)
	cancelButton = unison.NewSVGButton(svg.Not)
	if showKeys {
		applyButton.Tooltip = newWrappedTooltipWithSecondaryText(applyText, i18n.Text("%v%v or %v%v",
			mod.OSMenuCommand(), unison.KeyReturn, mod.OSMenuCommand(), unison.KeyNumPadEnter))
		cancelButton.Tooltip = newWrappedTooltipWithSecondaryText(cancelText, unison.KeyEscape.String())
	} else {
		applyButton.Tooltip = newWrappedTooltip(applyText)
		cancelButton.Tooltip = newWrappedTooltip(cancelText)
	}
	applyButton.SetEnabled(false)
	applyButton.ClickCallback = func() {
		if apply() {
			closeWithoutPrompt()
		}
	}
	toolbar.AddChild(applyButton)
	cancelButton.SetEnabled(false)
	cancelButton.ClickCallback = closeWithoutPrompt
	toolbar.AddChild(cancelButton)
	return applyButton, cancelButton
}

// menuEntry is one item of a menu that showMenu builds. One with no action is a heading, shown disabled after a
// separator unless it comes first; with no label as well, it is just the separator. Disabled grays out one with an
// action.
type menuEntry struct {
	Label    string
	Act      func()
	Disabled bool
}

// showMenu pops up a menu of the entries below the anchor.
func showMenu(anchor *unison.Panel, entries []menuEntry) {
	// A zero width lets the menu size to its items rather than stretch to a wide anchor.
	where := anchor.RectToRoot(anchor.ContentRect(true))
	where.Width = 0
	newEntriesMenu(entries).Popup(where, 0)
}

// newEntriesMenu returns a popup menu of the entries.
func newEntriesMenu(entries []menuEntry) unison.Menu {
	f := unison.DefaultMenuFactory()
	m := f.NewMenu(unison.PopupMenuTemporaryBaseID|unison.ContextMenuIDFlag, "", nil)
	for i, entry := range entries {
		id := unison.PopupMenuTemporaryBaseID + i + 1
		if entry.Act != nil {
			var validator func(unison.MenuItem) bool
			if entry.Disabled {
				validator = func(unison.MenuItem) bool { return false }
			}
			m.InsertItem(-1, f.NewItem(id, entry.Label, unison.KeyBinding{}, validator,
				func(unison.MenuItem) { entry.Act() }))
			continue
		}
		if i != 0 {
			m.InsertSeparator(-1, false)
		}
		if entry.Label != "" {
			m.InsertItem(-1, f.NewItem(id, entry.Label, unison.KeyBinding{}, func(unison.MenuItem) bool { return false }, nil))
		}
	}
	return m
}

// compactCornerRadius is the corner radius of compact controls and of the boxes drawn around them.
const compactCornerRadius = 4

// faintInk returns the ink at 30% opacity.
func faintInk(ink unison.Ink) unison.Ink {
	return &unison.ColorFilteredInk{OriginalInk: ink, ColorFilter: unison.Alpha30Filter()}
}

// hbox lays out the panel's children in a row that fills the width, returning the panel.
func hbox(panel *unison.Panel, spacing float32) *unison.Panel {
	panel.SetLayout(&unison.FlexLayout{Columns: len(panel.Children()), HSpacing: spacing})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return panel
}

// putOnLine has the child, an icon of the size, sit at the top of its row, centered on a first line of the height
// rather than on the whole row, whose text may wrap.
func putOnLine(child *unison.Panel, height, size float32) {
	var insets geom.Insets
	if border := child.Border(); border != nil {
		insets = border.Insets()
	}
	insets.Top = max((height-size)/2, 0)
	// Room below that makes the whole height whole, which sizers would otherwise round up, pushing the icon down.
	insets.Bottom = xmath.Ceil(insets.Top+size) - insets.Top - size
	child.SetBorder(unison.NewEmptyBorder(insets))
	child.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
}
