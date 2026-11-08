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
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/zeebo/xxh3"
)

var dimmedTextColor = unison.ThemeOnSurface.Derive(func(basedOn unison.ThemeColor) unison.ThemeColor {
	return unison.ThemeColor{
		Light: basedOn.Light.SetAlphaIntensity(0.5),
		Dark:  basedOn.Dark.SetAlphaIntensity(0.5),
	}
})

// adjustedInk is the ink of a value that a pool threshold, or a multiplier, holds away from its base, and of the state
// of a pool whose current threshold has features. Warning is what the encumbrance block uses for the temporary
// condition of being overloaded, and users can retheme it.
var adjustedInk unison.Ink = unison.ThemeWarning

// AttrPanel holds the contents of an attributes block on the sheet.
type AttrPanel struct {
	unison.Panel
	entity      *gurps.Entity
	targetMgr   *TargetMgr
	prefix      string
	hash        uint64
	rowStarts   []int
	kind        int
	nameLabels  map[string]*unison.Label
	valueFields map[string]unison.Paneler
	stateLabels map[string]*unison.Label
	// adjusters holds, for each non-pool attribute row, the function that shows whether its current value is adjusted;
	// see trackAdjustment.
	adjusters map[string]func()
}

// NewPrimaryAttrPanel creates a new primary attributes panel.
func NewPrimaryAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr) *AttrPanel {
	p := newAttrPanel(entity, targetMgr, gurps.PrimaryAttrKind)
	InstallTintFunc(p, colors.TintPrimaryAttributes)
	return p
}

// NewSecondaryAttrPanel creates a new secondary attributes panel.
func NewSecondaryAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr) *AttrPanel {
	p := newAttrPanel(entity, targetMgr, gurps.SecondaryAttrKind)
	InstallTintFunc(p, colors.TintSecondaryAttributes)
	return p
}

// NewPointPoolsPanel creates a new point pools panel.
func NewPointPoolsPanel(entity *gurps.Entity, targetMgr *TargetMgr) *AttrPanel {
	p := newAttrPanel(entity, targetMgr, gurps.PoolAttrKind)
	InstallTintFunc(p, colors.TintPools)
	return p
}

func newAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr, kind int) *AttrPanel {
	a := &AttrPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
		kind:      kind,
	}
	var title string
	switch kind {
	case gurps.PrimaryAttrKind:
		title = i18n.Text("Primary Attributes")
	case gurps.SecondaryAttrKind:
		title = i18n.Text("Secondary Attributes")
	case gurps.PoolAttrKind:
		title = i18n.Text("Point Pools")
	default:
		a.fatalKind()
	}
	initTitledPagePanel(a, title, a.columns(), false, nil)
	a.DrawCallback = a.drawSelf
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	a.hash = a.computeHash(attrs)
	a.rebuild(attrs)
	return a
}

// computeHash returns a hash of the attribute definitions combined with their trait-driven effective placements, so
// that the panel is rebuilt not only when the definitions change, but also when a trait is added, removed, enabled or
// disabled in a way that reveals or hides an attribute.
func (a *AttrPanel) computeHash(attrs *gurps.AttributeDefs) uint64 {
	h := xxh3.New()
	xhash.Num64(h, gurps.Hash64(attrs))
	for _, def := range attrs.List(false) {
		xhash.Num8(h, def.EffectivePlacement(a.entity))
	}
	return h.Sum64()
}

func (a *AttrPanel) drawSelf(gc *unison.Canvas, rect geom.Rect) {
	gc.DrawRect(rect, unison.ThemeBelowSurface.Paint(gc, rect, paintstyle.Fill))
	children := a.Children()
	for i, rowIndex := range a.rowStarts {
		if rowIndex >= len(children) {
			break
		}
		if i&1 == 0 {
			continue
		}
		r := children[rowIndex].FrameRect()
		upTo := len(children)
		if i < len(a.rowStarts)-1 {
			upTo = a.rowStarts[i+1]
		}
		for j := rowIndex + 1; j < upTo; j++ {
			r = r.Union(children[j].FrameRect())
		}
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, unison.ThemeBanding.Paint(gc, r, paintstyle.Fill))
	}
}

func (a *AttrPanel) fatalKind() {
	xos.ExitWithErr(errs.Newf("unknown attribute panel kind: %d", a.kind))
}

func (a *AttrPanel) columns() int {
	if a.kind == gurps.PoolAttrKind {
		return 6
	}
	return 3
}

func (a *AttrPanel) rebuild(attrs *gurps.AttributeDefs) {
	focusRefKey := a.targetMgr.CurrentFocusRef()
	removeBlockRows(a.AsPanel())
	a.rowStarts = nil
	a.nameLabels = make(map[string]*unison.Label)
	a.valueFields = make(map[string]unison.Paneler)
	a.adjusters = make(map[string]func())
	if a.kind == gurps.PoolAttrKind {
		a.stateLabels = make(map[string]*unison.Label)
	}
	sepCount := 0
	open := true
	for _, def := range attrs.List(false) {
		if def.Relevant(a.entity, a.kind) {
			if def.IsSeparator() {
				a.rowStarts = append(a.rowStarts, len(a.Children()))
				panel := unison.NewPanel()
				panel.SetLayout(&unison.FlexLayout{Columns: 2})
				panel.SetLayoutData(&unison.FlexLayoutData{
					HSpan:  a.columns(),
					HAlign: align.Fill,
					HGrab:  true,
				})
				var rotation float32
				currentSepCount := sepCount
				sepCount++
				if open = def.IsOpen(a.entity, currentSepCount); open {
					rotation = 90
				}
				title := def.CombinedName()
				button := unison.NewButton()
				tabStopForReading(button)
				button.SetLayoutData(&unison.FlexLayoutData{
					HAlign: align.Middle,
					VAlign: align.Middle,
				})
				// The button is nothing but its chevron, so its tooltip is also what a screen reader calls it.
				if title == "" {
					button.Tooltip = newWrappedTooltip(i18n.Text("Show or hide this section"))
				} else {
					button.Tooltip = newWrappedTooltip(i18n.Text("Show or hide %s", title))
				}
				button.HideBase = true
				button.HMargin = 0
				button.VMargin = 0
				size := max(fonts.PageLabelPrimary.Baseline()-2, 6)
				button.Drawable = &unison.DrawableSVG{
					SVG:             unison.CircledChevronRightSVG,
					Size:            geom.NewSize(size, size),
					RotationDegrees: rotation,
				}
				button.ClickCallback = func() {
					def.SetOpen(a.entity, currentSepCount, !def.IsOpen(a.entity, currentSepCount))
					a.forceSync()
					unison.InvokeTaskAfter(a.Window().UpdateCursorNow, time.Millisecond)
				}
				panel.AddChild(button)
				if title == "" {
					sep := unison.NewSeparator()
					sep.SetLayoutData(&unison.FlexLayoutData{
						HAlign: align.Fill,
						VAlign: align.Middle,
						HGrab:  true,
					})
					panel.AddChild(sep)
				} else {
					label := unison.NewLabel()
					label.Font = fonts.PageLabelSecondary
					label.HAlign = align.Middle
					label.OnBackgroundInk = unison.ThemeOnSurface
					label.SetTitle(title)
					label.SetLayoutData(&unison.FlexLayoutData{
						HAlign: align.Fill,
						VAlign: align.Middle,
						HGrab:  true,
					})
					label.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
						_, pref, _ := label.Sizes(geom.Size{})
						paint := unison.ThemeSurfaceEdge.Paint(gc, rect, paintstyle.Stroke)
						paint.SetStrokeWidth(1)
						half := (rect.Width - pref.Width) / 2
						gc.DrawLine(geom.NewPoint(rect.X, rect.CenterY()),
							geom.NewPoint(rect.X+half-2, rect.CenterY()), paint)
						gc.DrawLine(geom.NewPoint(2+rect.Right()-half, rect.CenterY()),
							geom.NewPoint(rect.Right(), rect.CenterY()), paint)
						label.DefaultDraw(gc, rect)
					}
					panel.AddChild(label)
				}
				a.AddChild(panel)
			} else if open {
				attr, ok := a.entity.Attributes.Set[def.ID()]
				if !ok {
					errs.Log(errs.New("unable to locate attribute data"), "id", def.ID())
					continue
				}
				a.rowStarts = append(a.rowStarts, len(a.Children()))
				if a.kind == gurps.PoolAttrKind {
					if def.Type != attribute.PoolRef {
						a.AddChild(a.createPointsField(attr))
					} else {
						a.AddChild(unison.NewPanel())
					}

					var currentField *DecimalField
					currentField = NewDecimalPageField(a.targetMgr, a.prefix+attr.AttrID+":cur",
						i18n.Text("Point Pool Current"), func() fxp.Int {
							if currentField != nil {
								currentField.SetMinMax(currentField.Min(), attr.Maximum())
							}
							return attr.Current()
						},
						func(v fxp.Int) { attr.Damage = (attr.Maximum() - v).Max(0) }, fxp.Min, attr.Maximum(), true)
					// The fields of a pool row sit on either side of the word "of", with the pool's name only after
					// both of them, so each is named outright.
					currentField.Accessibility.Name = i18n.Text("Current %s", def.ResolveFullName())
					a.AddChild(currentField)

					a.AddChild(NewPageLabel(i18n.Text("of")))

					var maximumField unison.Paneler
					if def.Type == attribute.Pool {
						field := NewDecimalPageField(a.targetMgr, a.prefix+attr.AttrID+":max",
							i18n.Text("Point Pool Maximum"), func() fxp.Int { return attr.Maximum() },
							func(v fxp.Int) {
								attr.SetMaximum(v)
								currentField.SetMinMax(currentField.Min(), v)
								currentField.Sync()
							}, fxp.Min, fxp.Max, true)
						field.Accessibility.Name = i18n.Text("Maximum %s", def.ResolveFullName())
						maximumField = field
					} else {
						maximumField = NewNonEditablePageFieldEndFor(func() string { return attr.Maximum().String() })
					}
					a.AddChild(maximumField)
					a.valueFields[def.ID()] = maximumField

					name := NewPageLabel(def.Name)
					a.AddChild(name)
					a.nameLabels[def.ID()] = name
					a.updateBonusTooltips(def.ID(), attr)

					state := NewPageLabel("")
					a.updateThreshold(state, attr)
					a.AddChild(state)
					a.stateLabels[def.ID()] = state
				} else {
					var valueField unison.Paneler
					var adjust func()
					editable := false
					if def.Type == attribute.IntegerRef || def.Type == attribute.DecimalRef {
						field := NewNonEditablePageFieldEndFor(func() string { return attr.Current().String() })
						field.SetLayoutData(&unison.FlexLayoutData{
							HSpan:  2,
							HAlign: align.Fill,
							VAlign: align.Middle,
						})
						valueField = field
						adjust = trackAdjustment(field, attr, adjustableLabelInk(field.Label), nil)
					} else {
						a.AddChild(a.createPointsField(attr))
						editable = true
						if def.AllowsDecimal() {
							field := NewDecimalPageField(a.targetMgr, a.prefix+attr.AttrID, def.CombinedName(),
								func() fxp.Int { return attr.Maximum() },
								func(v fxp.Int) { attr.SetMaximum(v) }, fxp.Min, fxp.Max, true)
							showCurrentWhileIdle(field, func() string { return attr.Current().String() })
							valueField = field
							adjust = trackAdjustment(field, attr, adjustableFieldInk(field.Field), hiddenCurrent(field))
						} else {
							field := NewIntegerPageField(a.targetMgr, a.prefix+attr.AttrID, def.CombinedName(),
								func() int { return attr.Maximum().Floor().AsInteger[int]() },
								func(v int) { attr.SetMaximum(fxp.FromInteger(v)) },
								fxp.Min.Floor().AsInteger[int](), fxp.Max.Floor().AsInteger[int](), false, true)
							showCurrentWhileIdle(field, func() string { return attr.Current().Floor().String() })
							valueField = field
							adjust = trackAdjustment(field, attr, adjustableFieldInk(field.Field), hiddenCurrent(field))
						}
					}
					adjust()
					a.adjusters[def.ID()] = adjust
					a.AddChild(valueField)
					a.valueFields[def.ID()] = valueField
					name := NewPageLabel(def.CombinedName())
					a.AddChild(name)
					a.nameLabels[def.ID()] = name
					a.updateBonusTooltips(def.ID(), attr)
					if editable {
						// The attribute's name is laid out after its value rather than before it. A non-editable value
						// is left alone, since it is static text whose name is the value it shows.
						valueField.AsPanel().Accessibility.LabeledBy = name
					}
				}
			}
		}
	}
	if a.targetMgr != nil {
		if sheet := a.Ancestor[*Sheet](); sheet != nil {
			a.targetMgr.ReacquireFocus(focusRefKey, sheet.toolbar, sheet.scroll.Content())
		}
	}
}

func (a *AttrPanel) createPointsField(attr *gurps.Attribute) unison.Paneler {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		points := attr.PointCost()
		f.SetTitleIfChanged("[" + points.String() + "]")
		// Spoken as "20 points", since the brackets are nothing to hear and a bare number does not say what it counts.
		speakAs(f, spokenPoints(points))
		if def := attr.AttributeDef(); def != nil {
			f.Tooltip = newWrappedTooltip(i18n.Text("Points spent on %s", def.CombinedName()))
		}
	})
	field.Font = fonts.PageFieldSecondary
	field.OnBackgroundInk = dimmedTextColor
	field.SetTitle(field.Text.String())
	return field
}

// showCurrentWhileIdle has a field that edits the maximum of a non-pool attribute show the attribute's current value,
// as current renders it, while the field is not being edited.
func showCurrentWhileIdle[T xmath.Integer | xmath.Float](field *NumericField[T], current func() string) {
	field.DisplayFormat = func(T) string { return current() }
	// The field was synced with the exact text when it was created, before the display format was installed.
	field.Sync()
}

// hiddenCurrent returns a function that returns the current value a field set up by showCurrentWhileIdle shows while
// idle, if the field shows the maximum instead, as it does while being edited, or an empty string if not.
func hiddenCurrent[T xmath.Integer | xmath.Float](field *NumericField[T]) func() string {
	return func() string {
		if field.showingDisplayText {
			return ""
		}
		return field.DisplayFormat(field.get())
	}
}

// trackAdjustment has the value field of a non-pool attribute's row drawn in adjustedInk, and spoken as adjusted from
// the attribute's maximum, while the attribute's current value is adjusted (see adjustedFrom), and returns the function
// that brings the field up to date. setInk is called with the new state each time it changes. hidden, if not nil,
// returns the current value while the field shows the maximum instead (see hiddenCurrent), and the field is then spoken
// as adjusted to that value.
func trackAdjustment(field unison.Paneler, attr *gurps.Attribute, setInk func(adjusted bool), hidden func() string) func() {
	var from string // The maximum the current value is adjusted from, or empty when it isn't adjusted.
	addAccessibilityCallback(field, func(node *accessibility.Node) {
		if from == "" {
			return
		}
		if hidden != nil {
			if current := hidden(); current != "" {
				node.Name = i18n.Text("%s, adjusted to %s", node.Name, current)
				return
			}
		}
		node.Name = i18n.Text("%s, adjusted from %s", node.Name, from)
		if node.Role == role.Label {
			// The name of static text is its text, so, as speakAs does, the text is withheld now that the name says
			// more, since macOS would speak both.
			node.Text = nil
		}
	})
	return func() {
		was := from != ""
		from = adjustedFrom(attr)
		if adjusted := from != ""; adjusted != was {
			setInk(adjusted)
		}
	}
}

// adjustedFrom returns the maximum of a non-pool attribute as text if the pools' current thresholds or the multipliers
// hold its current value away from the value it has without them, or an empty string if not. That value is the
// maximum or, for an attribute standing in for Striking ST, Lifting ST or Throwing ST, the maximum with the ST bonuses
// under the same limitation added and the sum floored, as gurps.Attribute.Current takes them, since those bonuses are
// part of the ST in use rather than an adjustment of it.
func adjustedFrom(attr *gurps.Attribute) string {
	maximum := attr.Maximum()
	unadjusted := maximum
	if e := attr.Entity; e != nil {
		switch attr.AttrID {
		case gurps.StrikingStrengthID:
			unadjusted = (maximum + e.StrikingStrengthBonus).Floor()
		case gurps.LiftingStrengthID:
			unadjusted = (maximum + e.LiftingStrengthBonus).Floor()
		case gurps.ThrowingStrengthID:
			unadjusted = (maximum + e.ThrowingStrengthBonus).Floor()
		}
	}
	if attr.Current() == unadjusted {
		return ""
	}
	return maximum.String()
}

// adjustableFieldInk returns a function that draws the field's text in adjustedInk, or in the inks it has now.
func adjustableFieldInk(field *unison.Field) func(adjusted bool) {
	onEditable, onBackground := field.OnEditableInk, field.OnBackgroundInk
	return func(adjusted bool) {
		if adjusted {
			field.OnEditableInk, field.OnBackgroundInk = adjustedInk, adjustedInk
		} else {
			field.OnEditableInk, field.OnBackgroundInk = onEditable, onBackground
		}
		field.MarkForRedraw()
	}
}

// adjustableLabelInk returns a function that draws the label's text in adjustedInk, or in the ink it has now.
func adjustableLabelInk(label *unison.Label) func(adjusted bool) {
	normal := label.OnBackgroundInk
	return func(adjusted bool) {
		ink := normal
		if adjusted {
			ink = adjustedInk
		}
		label.OnBackgroundInk = ink
		label.Text.AdjustDecorations(func(d *unison.TextDecoration) { d.OnBackgroundInk = ink })
		label.MarkForRedraw()
	}
}

// adjustedTooltip returns the tooltip text followed by the lines naming the bonuses that adjust the current value it
// describes, as gurps.Entity.CurrentAdjustmentTooltip lists them, if there are any.
func adjustedTooltip(text, adjustments string) string {
	if adjustments == "" {
		return text
	}
	return text + "\n" + i18n.Text("Adjusted by:") + "\n" + adjustments
}

func spokenPoints(points fxp.Int) string {
	if points == fxp.One || points == fxp.NegOne {
		return i18n.Text("%s point", points.String())
	}
	return i18n.Text("%s points", points.String())
}

// Sync the panel to the current data.
func (a *AttrPanel) Sync() {
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	if hash := a.computeHash(attrs); hash != a.hash {
		a.hash = hash
		a.rebuild(attrs)
	} else {
		// The labels only exist for the rows that were actually built, so no further filtering is needed here.
		for id := range a.nameLabels {
			if attr, ok := a.entity.Attributes.Set[id]; ok {
				a.updateBonusTooltips(id, attr)
			}
		}
		for _, adjust := range a.adjusters {
			adjust()
		}
		for id, label := range a.stateLabels {
			if attr, ok := a.entity.Attributes.Set[id]; ok {
				a.updateThreshold(label, attr)
			}
		}
	}
	MarkForLayoutWithinDockable(a)
}

// updateBonusTooltips sets the tooltip on both the name label and the value field of an attribute's row, listing the
// sources of the bonuses currently affecting it. Point pool labels show only the abbreviated name, so their full name
// leads the tooltip.
func (a *AttrPanel) updateBonusTooltips(id string, attr *gurps.Attribute) {
	text := a.bonusTooltipText(attr)
	if label, ok := a.nameLabels[id]; ok {
		label.Tooltip = wrappedTooltipOrNil(text)
	}
	if field, ok := a.valueFields[id]; ok {
		setFieldTooltip(field, wrappedTooltipOrNil(text))
	}
}

// bonusTooltipText returns the tooltip text for an attribute, or an empty string if it doesn't need one. A non-pool
// attribute whose current value is adjusted (see adjustedFrom) leads with its maximum; the bonuses that adjust it are
// listed after those that make up the maximum, among the modifiers BonusTooltip lists.
func (a *AttrPanel) bonusTooltipText(attr *gurps.Attribute) string {
	var parts []string
	if a.kind == gurps.PoolAttrKind {
		if def := attr.AttributeDef(); def != nil && def.FullName != "" {
			parts = append(parts, def.FullName)
		}
	} else if from := adjustedFrom(attr); from != "" {
		parts = append(parts, i18n.Text("Adjusted from %s", from))
	}
	if tooltip := attr.BonusTooltip(); tooltip != "" {
		parts = append(parts, tooltip)
	}
	return strings.Join(parts, "\n")
}

func wrappedTooltipOrNil(text string) *unison.Panel {
	if text == "" {
		return nil
	}
	return newWrappedTooltip(text)
}

var poolThresholdDescFont = &unison.DynamicFont{
	Resolver: func() unison.FontDescriptor {
		fd := fonts.PageLabelPrimary.Descriptor()
		fd.Size *= 0.8
		return fd
	},
}

// updateThreshold shows the state of the pool's current threshold, if it has one, in adjustedInk when the threshold has
// features it doesn't ignore (see gurps.ThresholdIgnores), which are then listed in the tooltip after its explanation,
// which is Markdown, as the plain text they are described with.
func (a *AttrPanel) updateThreshold(label *unison.Label, attr *gurps.Attribute) {
	label.DrawCallback = label.DefaultDraw
	label.Font = poolThresholdDescFont
	var txt string
	var tip *unison.Panel
	var ink unison.Ink = unison.ThemeOnSurface
	if threshold := attr.CurrentThreshold(); threshold != nil {
		txt = "[" + threshold.String() + "]"
		var explanation strings.Builder
		if threshold.Explanation != "" {
			explanation.WriteString(threshold.ResolveExplanation(attr))
		}
		listed := false
		defs := gurps.AttributeDefsFor(a.entity)
		for _, f := range threshold.Features {
			if gurps.ThresholdIgnores(defs, f) {
				continue
			}
			switch {
			case listed:
				explanation.WriteByte('\n')
			case explanation.Len() != 0:
				explanation.WriteString("\n\n")
			}
			listed = true
			explanation.WriteString("- ")
			explanation.WriteString(escapeMarkdown(gurps.DescribeThresholdFeature(f, a.entity, nil,
				func(s string) string { return s })))
		}
		if listed {
			ink = adjustedInk
		}
		if explanation.Len() != 0 {
			tip = newMarkdownTooltip(explanation.String(), "")
		}
	}
	label.Text = unison.NewSmallCapsText(txt, &unison.TextDecoration{
		Font:            poolThresholdDescFont,
		OnBackgroundInk: ink,
	})
	label.Tooltip = tip
}

func (a *AttrPanel) forceSync() {
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	a.hash = a.computeHash(attrs)
	a.rebuild(attrs)
	MarkForLayoutWithinDockable(a)
}
