/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	primaryAttrKind = iota
	secondaryAttrKind
	poolAttrKind
)

var dimmedPointsColor = unison.ThemeOnSurface.Derive(func(basedOn unison.ThemeColor) unison.ThemeColor {
	return unison.ThemeColor{
		Light: basedOn.Light.SetAlphaIntensity(0.5),
		Dark:  basedOn.Dark.SetAlphaIntensity(0.5),
	}
})

// AttrPanel holds the contents of an attributes block on the sheet.
type AttrPanel struct {
	unison.Panel
	entity      *gurps.Entity
	targetMgr   *TargetMgr
	prefix      string
	crc         uint64
	rowStarts   []int
	kind        int
	stateLabels map[string]*unison.RichLabel
}

// NewPrimaryAttrPanel creates a new primary attributes panel.
func NewPrimaryAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr) *AttrPanel {
	return newAttrPanel(entity, targetMgr, primaryAttrKind)
}

// NewSecondaryAttrPanel creates a new secondary attributes panel.
func NewSecondaryAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr) *AttrPanel {
	return newAttrPanel(entity, targetMgr, secondaryAttrKind)
}

// NewPointPoolsPanel creates a new point pools panel.
func NewPointPoolsPanel(entity *gurps.Entity, targetMgr *TargetMgr) *AttrPanel {
	return newAttrPanel(entity, targetMgr, poolAttrKind)
}

func newAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr, kind int) *AttrPanel {
	a := &AttrPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
		kind:      kind,
	}
	a.Self = a
	a.SetLayout(&unison.FlexLayout{
		Columns:  a.columns(),
		HSpacing: 4,
	})
	a.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  a.hspan(),
		VSpan:  a.vspan(),
		HAlign: align.Fill,
		VAlign: align.Fill,
	})
	var title string
	switch kind {
	case primaryAttrKind:
		title = i18n.Text("Primary Attributes")
	case secondaryAttrKind:
		title = i18n.Text("Secondary Attributes")
	case poolAttrKind:
		title = i18n.Text("Point Pools")
	default:
		a.fatalKind()
	}
	a.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: title},
		unison.NewEmptyBorder(unison.Insets{
			Top:    1,
			Left:   2,
			Bottom: 1,
			Right:  2,
		})))
	a.DrawCallback = a.drawSelf
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	a.crc = attrs.CRC64()
	a.rebuild(attrs)
	return a
}

func (a *AttrPanel) drawSelf(gc *unison.Canvas, rect unison.Rect) {
	gc.DrawRect(rect, unison.ThemeSurface.Paint(gc, rect, paintstyle.Fill))
	children := a.Children()
	for i, rowIndex := range a.rowStarts {
		if rowIndex > len(children) {
			break
		}
		if i&1 == 1 {
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
		gc.DrawRect(r, unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill))
	}
}

func (a *AttrPanel) fatalKind() {
	fatal.WithErr(errs.Newf("unknown attribute panel kind: %d", a.kind))
}

func (a *AttrPanel) columns() int {
	if a.kind == poolAttrKind {
		return 6
	}
	return 3
}

func (a *AttrPanel) hspan() int {
	if a.kind == poolAttrKind {
		return 2
	}
	return 1
}

func (a *AttrPanel) vspan() int {
	if a.kind == secondaryAttrKind {
		return 2
	}
	return 1
}

func (a *AttrPanel) isRelevant(def *gurps.AttributeDef) bool {
	switch a.kind {
	case primaryAttrKind:
		return def.Primary()
	case secondaryAttrKind:
		return def.Secondary()
	case poolAttrKind:
		return def.Pool()
	default:
		a.fatalKind()
		return false
	}
}

func (a *AttrPanel) rebuild(attrs *gurps.AttributeDefs) {
	focusRefKey := a.targetMgr.CurrentFocusRef()
	a.RemoveAllChildren()
	a.rowStarts = nil
	if a.kind == poolAttrKind {
		a.stateLabels = make(map[string]*unison.RichLabel)
	}
	for _, def := range attrs.List(false) {
		if a.isRelevant(def) {
			a.rowStarts = append(a.rowStarts, len(a.Children()))
			if def.IsSeparator() {
				a.AddChild(NewPageInternalHeader(def.CombinedName(), a.columns()))
			} else {
				attr, ok := a.entity.Attributes.Set[def.ID()]
				if !ok {
					errs.Log(errs.New("unable to locate attribute data"), "id", def.ID())
					continue
				}
				if a.kind == poolAttrKind {
					a.AddChild(a.createPointsField(attr))

					var currentField *DecimalField
					currentField = NewDecimalPageField(a.targetMgr, a.prefix+attr.AttrID+":cur",
						i18n.Text("Point Pool Current"), func() fxp.Int {
							if currentField != nil {
								currentField.SetMinMax(currentField.Min(), attr.Maximum())
							}
							return attr.Current()
						},
						func(v fxp.Int) { attr.Damage = (attr.Maximum() - v).Max(0) }, fxp.Min, attr.Maximum(), true)
					a.AddChild(currentField)

					a.AddChild(NewPageLabel(i18n.Text("of")))

					maximumField := NewDecimalPageField(a.targetMgr, a.prefix+attr.AttrID+":max",
						i18n.Text("Point Pool Maximum"), func() fxp.Int { return attr.Maximum() },
						func(v fxp.Int) {
							attr.SetMaximum(v)
							currentField.SetMinMax(currentField.Min(), v)
							currentField.Sync()
						}, fxp.Min, fxp.Max, true)
					a.AddChild(maximumField)

					name := NewPageLabel(def.Name)
					if def.FullName != "" {
						name.Tooltip = newWrappedTooltip(def.FullName)
					}
					a.AddChild(name)

					if threshold := attr.CurrentThreshold(); threshold != nil {
						state := NewPageLabel("[" + threshold.State + "]")
						if threshold.Explanation != "" {
							state.Tooltip = newWrappedTooltip(threshold.Explanation)
						}
						a.AddChild(state)
						a.stateLabels[def.ID()] = state
					} else {
						a.AddChild(unison.NewPanel())
					}
				} else {
					if def.Type == attribute.IntegerRef || def.Type == attribute.DecimalRef {
						field := NewNonEditablePageFieldEnd(func(field *NonEditablePageField) {
							field.Text = attr.Maximum().String()
						})
						field.SetLayoutData(&unison.FlexLayoutData{
							HSpan:  2,
							HAlign: align.Fill,
							VAlign: align.Middle,
						})
						a.AddChild(field)
					} else {
						a.AddChild(a.createPointsField(attr))
						if def.AllowsDecimal() {
							a.AddChild(NewDecimalPageField(a.targetMgr, a.prefix+attr.AttrID, def.CombinedName(),
								func() fxp.Int { return attr.Maximum() },
								func(v fxp.Int) { attr.SetMaximum(v) }, fxp.Min, fxp.Max, true))
						} else {
							a.AddChild(NewIntegerPageField(a.targetMgr, a.prefix+attr.AttrID, def.CombinedName(),
								func() int { return fxp.As[int](attr.Maximum().Trunc()) },
								func(v int) { attr.SetMaximum(fxp.From(v)) }, fxp.As[int](fxp.Min.Trunc()), fxp.As[int](fxp.Max.Trunc()), false, true))
						}
					}
					a.AddChild(NewPageLabel(def.CombinedName()))
				}
			}
		}
		switch a.kind {
		case primaryAttrKind:
		case secondaryAttrKind:
		case poolAttrKind:
		default:
			a.fatalKind()
		}
	}
	if a.targetMgr != nil {
		if sheet := unison.Ancestor[*Sheet](a); sheet != nil {
			a.targetMgr.ReacquireFocus(focusRefKey, sheet.toolbar, sheet.scroll.Content())
		}
	}
}

func (a *AttrPanel) createPointsField(attr *gurps.Attribute) unison.Paneler {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := "[" + attr.PointCost().String() + "]"; text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
		if def := attr.AttributeDef(); def != nil {
			f.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("Points spent on %s"), def.CombinedName()))
		}
	})
	field.Font = gurps.PageFieldSecondaryFont
	field.OnBackgroundInk = dimmedPointsColor
	return field
}

// Sync the panel to the current data.
func (a *AttrPanel) Sync() {
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	if crc := attrs.CRC64(); crc != a.crc {
		a.crc = crc
		a.rebuild(attrs)
	} else if a.kind == poolAttrKind {
		for _, def := range attrs.List(false) {
			if def.Pool() && def.Type != attribute.PoolSeparator {
				id := def.ID()
				if label, exists := a.stateLabels[id]; exists {
					if attr, ok := a.entity.Attributes.Set[id]; ok {
						label.DrawCallback = label.DefaultDraw
						if threshold := attr.CurrentThreshold(); threshold != nil {
							label.Text = unison.NewSmallCapsText("["+threshold.State+"]", &unison.TextDecoration{
								Font:       gurps.PageLabelPrimaryFont,
								Foreground: unison.ThemeOnSurface,
							})
							if threshold.Explanation != "" {
								label.Tooltip = newWrappedTooltip(threshold.Explanation)
							}
						} else {
							label.Text = unison.NewSmallCapsText("", &unison.TextDecoration{
								Font:       gurps.PageLabelPrimaryFont,
								Foreground: unison.ThemeOnSurface,
							})
							label.Tooltip = nil
						}
					}
				}
			}
		}
	}
	MarkForLayoutWithinDockable(a)
}
