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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
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

// Owned defines the methods a value owned by a Rebuildable should have
type Owned interface {
	Owner() Rebuildable
}

// FindOwner follow the lineage of a panel up locate a parent that satisfies `Owned`
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

// HasOwner follow the lineage of a panel up to determine if parent satisfies `Owned`
func HasOwner[T Rebuildable](panel *unison.Panel) bool {
	return !xreflect.IsNil(FindOwner[T](panel))
}

// Targeted defines the methods a value with a node target should have
type Targeted[N gurps.Node[N]] interface {
	Target() N
}

// FindTarget follow the lineage of a panel up locate a parent that satisfies `Targeted`
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

// Syncer should be called to sync an object's UI state to its model.
type Syncer interface {
	Sync()
}

// DeepSync does a depth-first traversal of the panel and all of its descendents and calls Sync() on any Syncer objects
// it finds.
func DeepSync(panel unison.Paneler) {
	p := panel.AsPanel()
	for _, child := range p.Children() {
		DeepSync(child)
	}
	if syncer, ok := p.Self.(Syncer); ok {
		syncer.Sync()
	}
}

// ModifiableRoot marks the root of a modifable tree of components, typically a Dockable.
type ModifiableRoot interface {
	MarkModified(src unison.Paneler)
}

// MarkModified looks for a ModifiableRoot, starting at the panel. If found, it then called MarkModified() on it.
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
// modification time is whatever the file system says it is -- which is why this is an optional interface rather than
// part of Rebuildable.
type modificationTimestampBumper interface {
	bumpModificationTimestamp()
}

var (
	_ modificationTimestampBumper = &Sheet{}
	_ modificationTimestampBumper = &LootSheet{}
)

// rebuildAsModified reports an edit to its owner by rebuilding the owner, in place of marking it as modified. A
// rebuild is a superset of marking as modified for every kind of owner -- it recalculates the entity, re-syncs every
// table, refreshes the search results and restores the focus and scroll position -- and is what an edit needs when it
// changes more of what the owner shows than the rows it touched: a sheet only carries the melee weapons, ranged
// weapons, reactions and conditional modifiers lists on the page while there is something to put in them, the weapon
// lists drop the columns nothing in them uses, and the switch column comes and goes with the presence of switchable
// features, and a set of columns can only change by building a new table. Doing both would repeat the whole update,
// and on a sheet holding hundreds of rows that update is the entire cost of the edit. The one thing a rebuild leaves
// out is bumping the owner's modification timestamp, so that is done here, and before the rebuild, since the panel
// showing the timestamp only picks up the new value when it is synced. A nil owner, typed or otherwise, is ignored.
func rebuildAsModified(owner Rebuildable, full bool) {
	if xreflect.IsNil(owner) {
		return
	}
	if bumper, ok := owner.(modificationTimestampBumper); ok {
		bumper.bumpModificationTimestamp()
	}
	owner.Rebuild(full)
}

func addSourceFields(parent *unison.Panel, source *gurps.SourcedID) {
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("ID"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(string(source.TID))
	}))
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("Source ID"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(string(source.Source.TID))
	}))
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("Source Library"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(source.Source.Library)
	}))
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("Source Path"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(source.Source.Path)
	}))
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
	labelText := i18n.Text("Notes")
	addLabel(parent, labelText, "")
	addScriptField(parent, nil, "", labelText,
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
	wrapper := addFlowWrapper(parent, tl, 2)
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
	wrapper := addFlowWrapper(parent, i18n.Text("Difficulty"), 3)
	addAttributeChoicePopup(wrapper, entity, "", &attrDiff.Attribute, gurps.TenFlag)
	wrapper.AddChild(NewFieldTrailingLabel("/", false))
	addPopup(wrapper, difficulty.Levels, &attrDiff.Difficulty)
}

func addTagsLabelAndField(parent *unison.Panel, fieldData *[]string) {
	addLabelAndListField(parent, i18n.Text("Tags"), i18n.Text("tags"), fieldData)
}

func addLabelAndListField(parent *unison.Panel, labelText, pluralForTooltip string, fieldData *[]string) {
	get, set := pointerAccessors(parent, fieldData)
	addMultiLineStringFieldWith(parent, labelText, fmt.Sprintf(i18n.Text("Separate multiple %s with commas"), pluralForTooltip),
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

// markdownHardLineBreaks converts the single newlines in the given text into Markdown hard line breaks so that
// multi-line tooltip content renders one line per newline. Without this, the Markdown renderer treats a single newline
// as a soft break, collapsing it into a space and putting everything on one line. Blank lines (paragraph breaks) and
// other block constructs are preserved, since a trailing hard break before a blank line is ignored by the renderer.
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

// pointerAccessors returns the accessors for a field that edits the value the pointer refers to. The setter stores
// the value and then marks the parent modified.
func pointerAccessors[T any](parent *unison.Panel, fieldData *T) (get func() T, set func(T)) {
	return func() T { return *fieldData },
		func(value T) {
			*fieldData = value
			MarkModified(parent)
		}
}

// installField gives the field the tooltip, if there is one, and adds it to the parent.
func installField[F unison.Paneler](parent *unison.Panel, field F, tooltip string) F {
	if tooltip != "" {
		field.AsPanel().Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(field)
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

func addLabel(parent *unison.Panel, labelText, tooltip string) {
	installField(parent, NewFieldLeadingLabel(labelText, false), tooltip)
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

func addWeightField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, entity *gurps.Entity, fieldData *fxp.Weight, noMinWidth bool) *WeightField {
	get, set := pointerAccessors(parent, fieldData)
	return installField(parent, NewWeightField(targetMgr, targetKey, labelText, entity, get, set, 0, fxp.Weight(fxp.Max),
		noMinWidth), tooltip)
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

func addFlowWrapper(parent *unison.Panel, labelText string, count int) *unison.Panel {
	parent.AddChild(NewFieldLeadingLabel(labelText, false))
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  count,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		VAlign:   align.Middle,
	})
	parent.AddChild(wrapper)
	return wrapper
}

func addFillWrapper(parent *unison.Panel, labelText string, count int) *unison.Panel {
	wrapper := addFlowWrapper(parent, labelText, count)
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	return wrapper
}

func addLabelAndPopup[T comparable](parent *unison.Panel, labelText, tooltip string, choices []T, fieldData *T) *unison.PopupMenu[T] {
	addLabel(parent, labelText, tooltip)
	return addPopup(parent, choices, fieldData)
}

func addPopup[T comparable](parent *unison.Panel, choices []T, fieldData *T) *unison.PopupMenu[T] {
	// Ensure that the passed field value is in the list of choices
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

func addBoolPopup(parent *unison.Panel, trueChoice, falseChoice string, fieldData *bool) *unison.PopupMenu[string] {
	popup := unison.NewPopupMenu[string]()
	popup.AddItem(trueChoice)
	popup.AddItem(falseChoice)
	if *fieldData {
		popup.SelectIndex(0)
	} else {
		popup.SelectIndex(1)
	}
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		*fieldData = p.SelectedIndex() == 0
		MarkModified(parent)
	}
	parent.AddChild(popup)
	return popup
}

func addHasPopup(parent *unison.Panel, has *bool) {
	addBoolPopup(parent, i18n.Text("has"), i18n.Text("doesn't have"), has)
}

func adjustFieldBlank(field unison.Paneler, blank bool) {
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

func adjustPopupBlank[T comparable](popup *unison.PopupMenu[T], blank bool) {
	popup.SetEnabled(!blank)
	if blank {
		popup.DrawOverCallback = func(gc *unison.Canvas, _ geom.Rect) {
			unison.DrawRoundedRectBase(gc, popup.ContentRect(false), popup.CornerRadius, 1, popup.BackgroundInk, popup.EdgeInk)
		}
	} else {
		popup.DrawOverCallback = nil
	}
}

func addNameCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	prefix := i18n.Text("whose name")
	return addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Name Qualifier"), strCriteria, hSpan,
		includeEmptyFiller)
}

func addSpecializationCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	prefix := i18n.Text("and whose specialization")
	return addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Specialization Qualifier"), strCriteria, hSpan,
		includeEmptyFiller)
}

func addUsageCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	prefix := i18n.Text("and whose usage")
	return addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Usage Qualifier"), strCriteria, hSpan,
		includeEmptyFiller)
}

func addTagCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	popup, field := addStringCriteriaPanel(parent, i18n.Text("and at least one tag"), i18n.Text("and all tags"),
		i18n.Text("Tag Qualifier"), strCriteria, hSpan, includeEmptyFiller)
	field.Tooltip = newWrappedTooltip(i18n.Text(`Separate multiple tags with commas to match any one of them, e.g. "Sword, Axe"`))
	return popup, field
}

func addNotesCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	prefix := i18n.Text("and whose notes")
	return addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Notes Qualifier"), strCriteria, hSpan,
		includeEmptyFiller)
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

// newComparisonPopup creates the popup menu for a criteria's comparison, offering the choices in order with the one at
// selectedIndex chosen. No selection callback is installed, since installing one first would have it called by the
// initial selection; the caller adds its own afterwards.
func newComparisonPopup(choices []string, selectedIndex int) *unison.PopupMenu[string] {
	popup := unison.NewPopupMenu[string]()
	popup.AddItem(choices...)
	popup.SelectIndex(selectedIndex)
	return popup
}

func addStringCriteriaPanel(parent *unison.Panel, prefix, notPrefix, undoTitle string, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	panel := newCriteriaPanel(parent, hSpan, includeEmptyFiller)
	var criteriaField *StringField
	popup := newComparisonPopup(criteria.PrefixedStringComparisonChoices(prefix, notPrefix),
		int(strCriteria.Compare.EnsureValid()))
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		strCriteria.Compare = criteria.StringComparisons[p.SelectedIndex()]
		adjustFieldBlank(criteriaField, strCriteria.IsZero())
		MarkModified(panel)
	}
	panel.AddChild(popup)
	criteriaField = addStringField(panel, undoTitle, "", &strCriteria.Qualifier)
	adjustFieldBlank(criteriaField, strCriteria.IsZero())
	return popup, criteriaField
}

func addLevelCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey string, numCriteria *criteria.Number, hSpan int, includeEmptyFiller bool) {
	addNumericCriteriaPanel(parent, targetMgr, targetKey, i18n.Text("and whose level"), i18n.Text("Level Qualifier"),
		numCriteria, 0, fxp.Thousand, hSpan, false, includeEmptyFiller)
}

func addNumericCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey, prefix, undoTitle string, numCriteria *criteria.Number, minValue, maxValue fxp.Int, hSpan int, integerOnly, includeEmptyFiller bool) (popup *unison.PopupMenu[string], field unison.Paneler) {
	panel := newCriteriaPanel(parent, hSpan, includeEmptyFiller)
	popup = newComparisonPopup(criteria.PrefixedNumericComparisonChoices(prefix), int(numCriteria.Compare.EnsureValid()))
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		numCriteria.Compare = criteria.NumericComparisons[p.SelectedIndex()]
		adjustFieldBlank(field, numCriteria.Compare == criteria.AnyNumber)
		MarkModified(panel)
	}
	panel.AddChild(popup)
	if integerOnly {
		field = NewIntegerField(targetMgr, targetKey, undoTitle,
			func() int { return numCriteria.Qualifier.AsInteger[int]() },
			func(value int) {
				numCriteria.Qualifier = fxp.FromInteger(value)
				MarkModified(panel)
			}, minValue.AsInteger[int](), maxValue.AsInteger[int](), false, false)
		panel.AddChild(field)
	} else {
		field = addDecimalField(panel, targetMgr, targetKey, undoTitle, "", &numCriteria.Qualifier, minValue, maxValue, false)
	}
	adjustFieldBlank(field, numCriteria.Compare == criteria.AnyNumber)
	return popup, field
}

// addWeightCriteriaPanel adds a weight criteria's comparison popup and qualifier field directly to the parent, which is
// expected to lay them out itself.
func addWeightCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey string, entity *gurps.Entity, weightCriteria *criteria.Weight) (popup *unison.PopupMenu[string], field *WeightField) {
	popup = newComparisonPopup(criteria.PrefixedNumericComparisonChoices(i18n.Text("which")),
		int(weightCriteria.Compare.EnsureValid()))
	parent.AddChild(popup)
	field = addWeightField(parent, targetMgr, targetKey, i18n.Text("Weight Qualifier"), "", entity,
		&weightCriteria.Qualifier, false)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		weightCriteria.Compare = criteria.NumericComparisons[p.SelectedIndex()]
		adjustFieldBlank(field, weightCriteria.Compare == criteria.AnyNumber)
		MarkModified(parent)
	}
	adjustFieldBlank(field, weightCriteria.Compare == criteria.AnyNumber)
	return popup, field
}

func addQuantityCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey string, numCriteria *criteria.Number) (popup *unison.PopupMenu[string], field *IntegerField) {
	choices := []string{
		i18n.Text("exactly"),
		i18n.Text("at least"),
		i18n.Text("at most"),
	}
	selectedIndex := 0
	switch numCriteria.Compare {
	case criteria.AtLeastNumber:
		selectedIndex = 1
	case criteria.AtMostNumber:
		selectedIndex = 2
	}
	popup = newComparisonPopup(choices, selectedIndex)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		switch p.SelectedIndex() {
		case 0:
			numCriteria.Compare = criteria.EqualsNumber
		case 1:
			numCriteria.Compare = criteria.AtLeastNumber
		case 2:
			numCriteria.Compare = criteria.AtMostNumber
		}
		MarkModified(parent)
	}
	parent.AddChild(popup)
	field = NewIntegerField(targetMgr, targetKey, i18n.Text("Quantity Criteria"),
		func() int { return numCriteria.Qualifier.AsInteger[int]() },
		func(value int) {
			numCriteria.Qualifier = fxp.FromInteger(value)
			MarkModified(parent)
		}, 0, 9999, false, false)
	parent.AddChild(field)
	return popup, field
}

func addLeveledAmountPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey, title string, amount *gurps.LeveledAmount) (field *DecimalField, checkBox *CheckBox) {
	field = addDecimalField(parent, targetMgr, targetKey, i18n.Text("Amount"), "", &amount.Amount, fxp.Min, fxp.Max, true)
	checkBox = addCheckBox(parent, title, &amount.PerLevel)
	return field, checkBox
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

// WrapWithSpan wraps a number of children with a single panel that request to fill in span number of columns.
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

// newApplyCancelButtons adds the Apply Changes and Discard Changes buttons that every editor with a pending set of
// changes has to the toolbar and returns them, disabled until there is something to apply. Clicking apply calls the
// given function and, when it reports success, closes the editor without its usual prompt, which is also all that
// cancel does. showKeys adds the keyboard shortcuts the editors bind to the buttons to their tooltips.
func newApplyCancelButtons(toolbar *unison.Panel, showKeys bool, apply func() bool, closeWithoutPrompt func()) (applyButton, cancelButton *unison.Button) {
	applyText := i18n.Text("Apply Changes")
	cancelText := i18n.Text("Discard Changes")
	applyButton = unison.NewSVGButton(unison.CheckmarkSVG)
	cancelButton = unison.NewSVGButton(svg.Not)
	if showKeys {
		applyButton.Tooltip = newWrappedTooltipWithSecondaryText(applyText, fmt.Sprintf(i18n.Text("%v%v or %v%v"),
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
