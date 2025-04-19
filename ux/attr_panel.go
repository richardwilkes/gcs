// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"time"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
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
	hash        uint64
	rowStarts   []int
	kind        int
	stateLabels map[string]*unison.Label
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
	case gurps.PrimaryAttrKind:
		title = i18n.Text("Primary Attributes")
	case gurps.SecondaryAttrKind:
		title = i18n.Text("Secondary Attributes")
	case gurps.PoolAttrKind:
		title = i18n.Text("Point Pools")
	default:
		a.fatalKind()
	}
	a.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: title},
		unison.NewEmptyBorder(unison.NewSymmetricInsets(2, 1))))
	a.DrawCallback = a.drawSelf
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	a.hash = gurps.Hash64(attrs)
	a.rebuild(attrs)
	return a
}

func (a *AttrPanel) drawSelf(gc *unison.Canvas, rect unison.Rect) {
	gc.DrawRect(rect, unison.ThemeBelowSurface.Paint(gc, rect, paintstyle.Fill))
	children := a.Children()
	for i, rowIndex := range a.rowStarts {
		if rowIndex > len(children) {
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
	fatal.WithErr(errs.Newf("unknown attribute panel kind: %d", a.kind))
}

func (a *AttrPanel) columns() int {
	if a.kind == gurps.PoolAttrKind {
		return 6
	}
	return 3
}

func (a *AttrPanel) hspan() int {
	if a.kind == gurps.PoolAttrKind {
		return 2
	}
	return 1
}

func (a *AttrPanel) vspan() int {
	if a.kind == gurps.SecondaryAttrKind {
		return 2
	}
	return 1
}

func (a *AttrPanel) rebuild(attrs *gurps.AttributeDefs) {
	focusRefKey := a.targetMgr.CurrentFocusRef()
	a.RemoveAllChildren()
	a.rowStarts = nil
	if a.kind == gurps.PoolAttrKind {
		a.stateLabels = make(map[string]*unison.Label)
	}
	sepCount := 0
	open := true
	for _, def := range attrs.List(false) {
		if def.Relevant(a.kind) {
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
				button := unison.NewButton()
				button.SetFocusable(false)
				button.SetLayoutData(&unison.FlexLayoutData{
					HAlign: align.Middle,
					VAlign: align.Middle,
				})
				button.HideBase = true
				button.HMargin = 0
				button.VMargin = 0
				size := max(fonts.PageLabelPrimary.Baseline()-2, 6)
				button.Drawable = &unison.DrawableSVG{
					SVG:             unison.CircledChevronRightSVG,
					Size:            unison.NewSize(size, size),
					RotationDegrees: rotation,
				}
				button.ClickCallback = func() {
					def.SetOpen(a.entity, currentSepCount, !def.IsOpen(a.entity, currentSepCount))
					a.forceSync()
					unison.InvokeTaskAfter(a.Window().UpdateCursorNow, time.Millisecond)
				}
				panel.AddChild(button)
				title := def.CombinedName()
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
					label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
						_, pref, _ := label.Sizes(unison.Size{})
						paint := unison.ThemeSurfaceEdge.Paint(gc, rect, paintstyle.Stroke)
						paint.SetStrokeWidth(1)
						half := (rect.Width - pref.Width) / 2
						gc.DrawLine(rect.X, rect.CenterY(), rect.X+half-2, rect.CenterY(), paint)
						gc.DrawLine(2+rect.Right()-half, rect.CenterY(), rect.Right(), rect.CenterY(), paint)
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
					a.AddChild(currentField)

					a.AddChild(NewPageLabel(i18n.Text("of")))

					if def.Type == attribute.Pool {
						a.AddChild(NewDecimalPageField(a.targetMgr, a.prefix+attr.AttrID+":max",
							i18n.Text("Point Pool Maximum"), func() fxp.Int { return attr.Maximum() },
							func(v fxp.Int) {
								attr.SetMaximum(v)
								currentField.SetMinMax(currentField.Min(), v)
								currentField.Sync()
							}, fxp.Min, fxp.Max, true))
					} else {
						a.AddChild(NewNonEditablePageFieldEnd(func(field *NonEditablePageField) {
							field.SetTitle(attr.Maximum().String())
						}))
					}

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
							field.SetTitle(attr.Maximum().String())
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
	}
	if a.targetMgr != nil {
		if sheet := unison.Ancestor[*Sheet](a); sheet != nil {
			a.targetMgr.ReacquireFocus(focusRefKey, sheet.toolbar, sheet.scroll.Content())
		}
	}
}

func (a *AttrPanel) createPointsField(attr *gurps.Attribute) unison.Paneler {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := "[" + attr.PointCost().String() + "]"; text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f.AsPanel())
		}
		if def := attr.AttributeDef(); def != nil {
			f.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("Points spent on %s"), def.CombinedName()))
		}
	})
	field.Font = fonts.PageFieldSecondary
	field.OnBackgroundInk = dimmedPointsColor
	field.SetTitle(field.Text.String())
	return field
}

// Sync the panel to the current data.
func (a *AttrPanel) Sync() {
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	if hash := gurps.Hash64(attrs); hash != a.hash {
		a.hash = hash
		a.rebuild(attrs)
	} else if a.kind == gurps.PoolAttrKind {
		for _, def := range attrs.List(false) {
			if def.Pool() && def.Type != attribute.PoolSeparator {
				id := def.ID()
				if label, exists := a.stateLabels[id]; exists {
					if attr, ok := a.entity.Attributes.Set[id]; ok {
						label.DrawCallback = label.DefaultDraw
						if threshold := attr.CurrentThreshold(); threshold != nil {
							label.Text = unison.NewSmallCapsText("["+threshold.State+"]", &unison.TextDecoration{
								Font:            fonts.PageLabelPrimary,
								OnBackgroundInk: unison.ThemeOnSurface,
							})
							if threshold.Explanation != "" {
								label.Tooltip = newWrappedTooltip(threshold.Explanation)
							}
						} else {
							label.Text = unison.NewSmallCapsText("", &unison.TextDecoration{
								Font:            fonts.PageLabelPrimary,
								OnBackgroundInk: unison.ThemeOnSurface,
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

func (a *AttrPanel) forceSync() {
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	a.hash = gurps.Hash64(attrs)
	a.rebuild(attrs)
	MarkForLayoutWithinDockable(a)
}
