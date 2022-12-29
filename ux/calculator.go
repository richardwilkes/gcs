/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"math"
	"strings"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var (
	_ unison.Dockable            = &Calculator{}
	_ unison.UndoManagerProvider = &Calculator{}
	_ GroupedCloser              = &Calculator{}
)

// Calculator provides calculations for various physical tasks, such as jumping.
type Calculator struct {
	unison.Panel
	sheet                      *Sheet
	undoMgr                    *unison.UndoManager
	scroll                     *unison.ScrollPanel
	jumpingLabel               *unison.Label
	highJumpResult             *unison.Label
	broadJumpResult            *unison.Label
	throwingDistanceResult     *unison.Label
	throwingDamageResult       *unison.Label
	scale                      int
	jumpRunningStartYards      fxp.Int
	objectWeight               model.Weight
	jumpingExtraEffortPenalty  int
	throwingExtraEffortPenalty int
}

// DisplayCalculator displays the calculator for the given Sheet.
func DisplayCalculator(sheet *Sheet) {
	ws, dc, found := Activate(func(d unison.Dockable) bool {
		if e, ok := d.(*Calculator); ok {
			return e.sheet == sheet
		}
		return false
	})
	if !found && ws != nil {
		e := &Calculator{
			sheet:        sheet,
			scale:        model.GlobalSettings().General.InitialEditorUIScale,
			objectWeight: model.Weight(fxp.One),
		}
		e.Self = e

		e.undoMgr = unison.NewUndoManager(100, func(err error) { jot.Error(err) })
		e.SetLayout(&unison.FlexLayout{Columns: 1})

		content := e.createContent()

		e.scroll = unison.NewScrollPanel()
		e.scroll.SetContent(content, unison.HintedFillBehavior, unison.FillBehavior)
		e.scroll.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			VAlign: unison.FillAlignment,
			HGrab:  true,
			VGrab:  true,
		})

		e.AddChild(e.createToolbar())
		e.AddChild(e.scroll)
		e.ClientData()[AssociatedUUIDKey] = sheet.Entity().ID
		e.scroll.Content().AsPanel().ValidateScrollRoot()
		group := EditorGroup
		p := sheet.AsPanel()
		for p != nil {
			if _, exists := p.ClientData()[AssociatedUUIDKey]; exists {
				group = subEditorGroup
				break
			}
			p = p.Parent()
		}
		PlaceInDock(ws, dc, e, group)
		content.RequestFocus()
	}
}

// UpdateCalculator for the given owner.
func UpdateCalculator(sheet *Sheet) {
	for _, wnd := range unison.Windows() {
		if ws := WorkspaceFromWindow(wnd); ws != nil {
			ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
				for _, other := range dc.Dockables() {
					if c, ok := other.(*Calculator); ok && c.sheet == sheet {
						c.updateAll()
						return true
					}
				}
				return false
			})
		}
	}
}

func (c *Calculator) createToolbar() *unison.Panel {
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))

	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(
		NewScaleField(
			model.InitialUIScaleMin,
			model.InitialUIScaleMax,
			func() int { return model.GlobalSettings().General.InitialEditorUIScale },
			func() int { return c.scale },
			func(scale int) { c.scale = scale },
			nil,
			false,
			c.scroll,
		),
	)

	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
	return toolbar
}

func (c *Calculator) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing * 2)))
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	c.addJumpingSection(content)
	c.addThrowingSection(content)
	return content
}

func (c *Calculator) addJumpingSection(content *unison.Panel) {
	content.AddChild(c.createHeader(i18n.Text("Jumping"), "B352", "Jumping", 0))

	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HSpacing: unison.StdHSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	label := unison.NewLabel()
	label.Text = i18n.Text("With a")
	wrapper.AddChild(label)
	field := NewDecimalField(nil, "", i18n.Text("Jump Running Start"),
		func() fxp.Int { return c.jumpRunningStartYards },
		func(v fxp.Int) {
			c.jumpRunningStartYards = v
			c.updateJumpingResult()
		},
		0, fxp.Max, false, false)
	wrapper.AddChild(field)
	c.jumpingLabel = unison.NewLabel()
	wrapper.AddChild(c.jumpingLabel)
	wrapper.AddChild(NewIntegerField(nil, "", i18n.Text("Jumping Extra Effort Penalty"),
		func() int { return c.jumpingExtraEffortPenalty },
		func(v int) {
			c.jumpingExtraEffortPenalty = v
			c.updateJumpingResult()
		},
		-100, 0, false, false))
	label = unison.NewLabel()
	label.Text = i18n.Text("penalty for extra effort:")
	wrapper.AddChild(label)
	content.AddChild(wrapper)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 4}))
	label = unison.NewLabel()
	label.Text = i18n.Text("High Jump:")
	wrapper.AddChild(label)
	c.highJumpResult = unison.NewLabel()
	wrapper.AddChild(c.highJumpResult)
	label = unison.NewLabel()
	label.Text = i18n.Text("Broad Jump:")
	wrapper.AddChild(label)
	c.broadJumpResult = unison.NewLabel()
	c.updateJumpingResult()
	wrapper.AddChild(c.broadJumpResult)
	content.AddChild(wrapper)
}

func (c *Calculator) addThrowingSection(content *unison.Panel) {
	content.AddChild(c.createHeader(i18n.Text("Throwing"), "B355", "Throwing", unison.StdVSpacing*3))

	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HSpacing: unison.StdHSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	label := unison.NewLabel()
	label.Text = i18n.Text("A")
	wrapper.AddChild(label)
	wrapper.AddChild(NewWeightField(nil, "", i18n.Text("Object Weight"),
		c.sheet.Entity(),
		func() model.Weight { return c.objectWeight },
		func(v model.Weight) {
			c.objectWeight = v
			c.updateThrowingResult()
		},
		0, model.Weight(fxp.Max), false))
	label = unison.NewLabel()
	label.Text = i18n.Text("object and a")
	wrapper.AddChild(label)
	wrapper.AddChild(NewIntegerField(nil, "", i18n.Text("Throwing Extra Effort Penalty"),
		func() int { return c.throwingExtraEffortPenalty },
		func(v int) {
			c.throwingExtraEffortPenalty = v
			c.updateThrowingResult()
		},
		-100, 0, false, false))
	label = unison.NewLabel()
	label.Text = i18n.Text("penalty for extra effort:")
	wrapper.AddChild(label)
	content.AddChild(wrapper)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 4}))
	label = unison.NewLabel()
	label.Text = i18n.Text("Distance:")
	wrapper.AddChild(label)
	c.throwingDistanceResult = unison.NewLabel()
	wrapper.AddChild(c.throwingDistanceResult)
	label = unison.NewLabel()
	label.Text = i18n.Text("Damage:")
	wrapper.AddChild(label)
	c.throwingDamageResult = unison.NewLabel()
	wrapper.AddChild(c.throwingDamageResult)
	c.updateThrowingResult()
	content.AddChild(wrapper)
}

func (c *Calculator) createHeader(text, linkRef, linkHighlight string, topMargin float32) *unison.Panel {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{Columns: 3})
	if topMargin > 0 {
		wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: topMargin}))
	}

	first := unison.NewLabel()
	first.Text = text + " ("
	first.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.LabelFont.Descriptor()
			desc.Size += 2
			desc.Weight = unison.BoldFontWeight
			return desc
		},
	}
	wrapper.AddChild(first)

	link := NewLink(linkRef, func() {
		OpenPageReference(nil, linkRef, linkHighlight, nil)
	})
	link.Font = first.Font
	wrapper.AddChild(link)

	last := unison.NewLabel()
	last.Text = ")"
	last.Font = first.Font
	wrapper.AddChild(last)
	return wrapper
}

// TitleIcon implements unison.Dockable
func (c *Calculator) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  svg.Calculator,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable
func (c *Calculator) Title() string {
	return fmt.Sprintf(i18n.Text("Calculator for %s"), c.sheet.String())
}

func (c *Calculator) String() string {
	return c.Title()
}

// Tooltip implements unison.Dockable
func (c *Calculator) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable
func (c *Calculator) Modified() bool {
	return false
}

// CloseWithGroup implements GroupedCloser
func (c *Calculator) CloseWithGroup(other unison.Paneler) bool {
	return c.sheet != nil && c.sheet == other
}

// MayAttemptClose implements GroupedCloser
func (c *Calculator) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(c)
}

// AttemptClose implements GroupedCloser
func (c *Calculator) AttemptClose() bool {
	if !CloseGroup(c) {
		return false
	}
	if dc := unison.Ancestor[*unison.DockContainer](c); dc != nil {
		dc.Close(c)
	}
	return true
}

// UndoManager implements unison.UndoManagerProvider
func (c *Calculator) UndoManager() *unison.UndoManager {
	return c.undoMgr
}

func (c *Calculator) computeJump(broad bool) fxp.Int {
	entity := c.sheet.Entity()
	basicMove := entity.Attributes.Current("basic_move")
	basicMoveWithoutRun := basicMove

	// Adjust Basic Move for running
	if c.jumpRunningStartYards > 0 {
		enhMove := -fxp.One
		model.Traverse(func(t *model.Trait) bool {
			if strings.EqualFold(t.Name, "enhanced move (ground)") && t.IsLeveled() {
				if enhMove == -fxp.One {
					enhMove = t.Levels
				} else {
					enhMove += t.Levels
				}
			}
			return false
		}, true, false, entity.Traits...)
		basicMove += c.jumpRunningStartYards
		if enhMove > 0 {
			if adjusted := basicMoveWithoutRun.Mul(enhMove + fxp.One); adjusted > basicMove {
				basicMove = adjusted
			}
		}
	}

	// Adjust Basic Move for Jumping skill
	model.Traverse(func(s *model.Skill) bool {
		if strings.EqualFold(s.Name, "jumping") {
			s.UpdateLevel()
			level := s.LevelData.Level.Div(fxp.Two).Trunc()
			if level > basicMove {
				basicMove = level
			}
			if level > basicMoveWithoutRun {
				basicMoveWithoutRun = level
			}
			return true
		}
		return false
	}, true, true, entity.Skills...)

	// Adjust Basic Move for high strength
	st := entity.Attributes.Current("st")
	if c.jumpingExtraEffortPenalty < 0 {
		st = st.Mul(fxp.From(-5*c.jumpingExtraEffortPenalty).Div(fxp.Hundred) + fxp.One).Trunc()
	}
	if basicLift := entity.BasicLiftForST(st); basicLift > entity.Profile.Weight {
		adjusted := st.Div(fxp.Four).Trunc()
		if adjusted > basicMove {
			basicMove = adjusted
		}
		if adjusted > basicMoveWithoutRun {
			basicMoveWithoutRun = adjusted
		}
	}

	// Determine base distance
	var multiplier, reduction fxp.Int
	if broad {
		multiplier = fxp.Two
		reduction = fxp.Three
	} else {
		multiplier = fxp.Six
		reduction = fxp.Ten
	}
	distance := (basicMove.Mul(multiplier) - reduction).Min((basicMoveWithoutRun.Mul(multiplier) - reduction).Mul(fxp.Two))

	// Adjust for encumbrance
	distance = distance.Mul(fxp.One - fxp.From(int(entity.EncumbranceLevel(false))).Mul(fxp.Two).Div(fxp.Ten))

	// Adjust for Super Jump
	levels := -fxp.One
	model.Traverse(func(t *model.Trait) bool {
		if strings.EqualFold(t.Name, "super jump") && t.IsLeveled() {
			if levels == -fxp.One {
				levels = t.Levels
			} else {
				levels += t.Levels
			}
		}
		return false
	}, true, false, entity.Traits...)
	if levels > 0 {
		distance = distance.Mul(fxp.From(math.Pow(2, fxp.As[float64](levels))))
	}
	if broad {
		distance = distance.Mul(fxp.Twelve)
	}
	return distance.Trunc()
}

func (c *Calculator) updateJumpingResult() {
	c.highJumpResult.Text = c.distanceToText(c.computeJump(false))
	c.broadJumpResult.Text = c.distanceToText(c.computeJump(true))
	var units string
	if c.useMeters() {
		units = i18n.Text("meter")
	} else {
		units = i18n.Text("yard")
	}
	c.jumpingLabel.Text = fmt.Sprintf(i18n.Text("%s running start and a"), units)
}

func (c *Calculator) updateThrowingResult() {
	if c.objectWeight <= 0 {
		c.throwingDistanceResult.Text = i18n.Text("None")
		c.throwingDamageResult.Text = i18n.Text("None")
		return
	}
	entity := c.sheet.Entity()

	// Determine bonuses for skills
	var distanceBonus, damageBonus int
	model.Traverse(func(s *model.Skill) bool {
		if strings.EqualFold(s.Name, "throwing art") {
			s.UpdateLevel()
			if s.LevelData.RelativeLevel >= fxp.One {
				if distanceBonus < 2 {
					distanceBonus = 2
				}
				if damageBonus < 2 {
					damageBonus = 2
				}
			} else if s.LevelData.RelativeLevel >= 0 {
				if distanceBonus < 1 {
					distanceBonus = 1
				}
				if damageBonus < 1 {
					damageBonus = 1
				}
			}
		}
		if strings.EqualFold(s.Name, "throwing") {
			s.UpdateLevel()
			if s.LevelData.RelativeLevel >= fxp.Two {
				if distanceBonus < 2 {
					distanceBonus = 2
				}
			} else if s.LevelData.RelativeLevel >= fxp.One {
				if distanceBonus < 1 {
					distanceBonus = 1
				}
			}
		}
		return false
	}, true, true, entity.Skills...)

	// Determine distance modifier based on weight ratio
	st := entity.Attributes.Current("st")
	if c.throwingExtraEffortPenalty < 0 {
		st = st.Mul(fxp.From(-5*c.throwingExtraEffortPenalty).Div(fxp.Hundred) + fxp.One).Trunc()
	}
	st += fxp.From(distanceBonus)
	basicLift := entity.BasicLiftForST(st)
	var weightRatio fxp.Int
	if basicLift > 0 {
		weightRatio = fxp.Int(c.objectWeight).Div(fxp.Int(basicLift))
	}
	var modifier fxp.Int
	switch {
	case weightRatio <= fxp.Twentieth:
		modifier = fxp.ThreeAndAHalf
	case weightRatio <= fxp.Tenth:
		modifier = fxp.TwoAndAHalf
	case weightRatio <= fxp.PointOneFive:
		modifier = fxp.Two
	case weightRatio <= fxp.Fifth:
		modifier = fxp.OneAndAHalf
	case weightRatio <= fxp.Quarter:
		modifier = fxp.OnePointTwo
	case weightRatio <= fxp.ThreeTenths:
		modifier = fxp.OnePointOne
	case weightRatio <= fxp.TwoFifths:
		modifier = fxp.One
	case weightRatio <= fxp.Half:
		modifier = fxp.FourFifths
	case weightRatio <= fxp.ThreeQuarters:
		modifier = fxp.SevenTenths
	case weightRatio <= fxp.One:
		modifier = fxp.ThreeFifths
	case weightRatio <= fxp.OneAndAHalf:
		modifier = fxp.TwoFifths
	case weightRatio <= fxp.Two:
		modifier = fxp.ThreeTenths
	case weightRatio <= fxp.TwoAndAHalf:
		modifier = fxp.Quarter
	case weightRatio <= fxp.Three:
		modifier = fxp.Fifth
	case weightRatio <= fxp.Four:
		modifier = fxp.PointOneFive
	case weightRatio <= fxp.Five:
		modifier = fxp.PointOneTwo
	case weightRatio <= fxp.Six:
		modifier = fxp.Tenth
	case weightRatio <= fxp.Seven:
		modifier = fxp.PointZeroNine
	case weightRatio <= fxp.Eight:
		modifier = fxp.PointZeroEight
	case weightRatio <= fxp.Nine:
		modifier = fxp.PointZeroSeven
	case weightRatio <= fxp.Ten:
		modifier = fxp.PointZeroSix
	case weightRatio <= fxp.Twelve:
		modifier = fxp.Twentieth
	}
	inches := st.Mul(modifier).Mul(fxp.ThirtySix).Trunc()
	if inches <= fxp.One {
		c.throwingDistanceResult.Text = i18n.Text("The object is too heavy for you to throw")
		c.throwingDamageResult.Text = i18n.Text("None")
		return
	}

	// Determine damage based on weight ratio
	thrust := entity.Thrust()
	thrust.Modifier += thrust.Count * damageBonus
	basicLift = entity.BasicLiftForST(st - fxp.From(distanceBonus))
	if basicLift > 0 {
		weightRatio = fxp.Int(c.objectWeight).Div(fxp.Int(basicLift))
	} else {
		weightRatio = 0
	}
	switch {
	case weightRatio <= fxp.Eighth:
		thrust.Modifier -= thrust.Count * 2
	case weightRatio <= fxp.Quarter:
		thrust.Modifier -= thrust.Count
	case weightRatio <= fxp.Half:
	case weightRatio <= fxp.One:
		thrust.Modifier += thrust.Count
	case weightRatio <= fxp.Two:
	case weightRatio <= fxp.Four:
		thrust.Modifier -= thrust.Count / 2
	default:
		thrust.Modifier -= thrust.Count
	}

	c.throwingDistanceResult.Text = c.distanceToText(inches)
	c.throwingDamageResult.Text = thrust.StringExtra(entity.SheetSettings.UseModifyingDicePlusAdds)
}

func (c *Calculator) useMeters() bool {
	units := c.sheet.Entity().SheetSettings.DefaultLengthUnits
	return units == model.Centimeter || units == model.Meter || units == model.Kilometer
}

func (c *Calculator) distanceToText(inches fxp.Int) string {
	var buffer strings.Builder
	if c.useMeters() {
		meters := inches.Div(fxp.ThirtySix).Mul(fxp.Hundred).Round().Div(fxp.Hundred)
		if meters == fxp.One {
			buffer.WriteString(i18n.Text("1 meter"))
		} else {
			fmt.Fprintf(&buffer, i18n.Text("%s meters"), meters.Comma())
		}
	} else {
		if inches >= fxp.ThirtySix {
			yards := inches.Div(fxp.ThirtySix).Trunc()
			if yards == fxp.One {
				buffer.WriteString(i18n.Text("1 yard"))
			} else {
				fmt.Fprintf(&buffer, i18n.Text("%s yards"), yards.Comma())
			}
			inches -= yards.Mul(fxp.ThirtySix)
		}
		if inches >= fxp.Twelve {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			feet := inches.Div(fxp.Twelve).Trunc()
			if feet == fxp.One {
				buffer.WriteString(i18n.Text("1 foot"))
			} else {
				fmt.Fprintf(&buffer, i18n.Text("%s feet"), feet.Comma())
			}
			inches -= feet.Mul(fxp.Twelve)
		}
		if inches > 0 || buffer.Len() == 0 {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			if inches == fxp.One {
				buffer.WriteString(i18n.Text("1 inch"))
			} else {
				fmt.Fprintf(&buffer, i18n.Text("%s inches"), inches.Comma())
			}
		}
	}
	return buffer.String()
}

func (c *Calculator) updateAll() {
	c.updateJumpingResult()
	c.updateThrowingResult()
	content := c.scroll.Content().AsPanel()
	content.MarkForLayoutRecursively()
	content.MarkForRedraw()
}
