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
	"math"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/weight"
)

var (
	_ unison.Dockable            = &Calculator{}
	_ unison.UndoManagerProvider = &Calculator{}
	_ GroupedCloser              = &Calculator{}

	terrain = []terrainModifier{
		{Name: i18n.Text("Broken Ground"), Modifier: fxp.Half},
		{Name: i18n.Text("Deep Snow"), Modifier: fxp.Fifth, IsSnow: true},
		{Name: i18n.Text("Desert"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Desert, Hard-packed"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Forest"), Modifier: fxp.Half},
		{Name: i18n.Text("Forest, Dense"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Forest, Light"), Modifier: fxp.One},
		{Name: i18n.Text("Frozen Lake"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Frozen River"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Hills, Rolling"), Modifier: fxp.One},
		{Name: i18n.Text("Hills, Steep"), Modifier: fxp.Half},
		{Name: i18n.Text("Jungle"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Mountains"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Mud"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Plains, Level"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Road, Cobblestone"), Modifier: fxp.One, IsRoad: true},
		{Name: i18n.Text("Road, Dirt"), Modifier: fxp.One, ModifierInRain: fxp.Fifth, IsRoad: true, Default: true},
		{Name: i18n.Text("Road, Gravel"), Modifier: fxp.One, ModifierInRain: fxp.Fifth, IsRoad: true},
		{Name: i18n.Text("Road, Paved"), Modifier: fxp.OneAndAQuarter, ModifierInRain: fxp.One, IsRoad: true},
		{Name: i18n.Text("Sand"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Sand, Hard-packed"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Swamp"), Modifier: fxp.Fifth},
	}

	weather = []terrainModifier{
		{Name: i18n.Text("Normal"), Modifier: fxp.One, Default: true},
		{Name: i18n.Text("Rain"), Modifier: fxp.Half, IsRain: true},
		{Name: i18n.Text("Sleet"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Snow"), Modifier: fxp.Half, IsSnow: true},
		{Name: i18n.Text("Snow, Heavy"), Modifier: fxp.Quarter, IsSnow: true},
	}
)

type terrainModifier struct {
	Name           string
	Modifier       fxp.Int
	ModifierInRain fxp.Int
	IsRoad         bool
	IsRain         bool
	IsSnow         bool
	IsIce          bool
	Default        bool
}

func (t terrainModifier) String() string {
	return t.Name
}

// Calculator provides calculations for various physical tasks, such as jumping.
type Calculator struct {
	unison.Panel
	sheet                      *Sheet
	undoMgr                    *unison.UndoManager
	content                    *unison.Panel
	scroll                     *unison.ScrollPanel
	jumpingLabel               *unison.Label
	highJumpResult             *unison.Label
	broadJumpResult            *unison.Label
	throwingDistanceResult     *unison.Label
	throwingDamageResult       *unison.Label
	hikingResult               *unison.Label
	scale                      int
	jumpingRunningStartYards   fxp.Int
	throwingObjectWeight       fxp.Weight
	jumpingExtraEffortPenalty  int
	throwingExtraEffortPenalty int
	hikingExtraEffortPenalty   int
	terrainIndex               int
	weatherIndex               int
	usingSkis                  bool
	usingSkates                bool
	roadsAreCleared            bool
	successfulHikingRoll       bool
}

// DisplayCalculator displays the calculator for the given Sheet.
func DisplayCalculator(sheet *Sheet) {
	if Activate(func(d unison.Dockable) bool {
		if c, ok := d.AsPanel().Self.(*Calculator); ok {
			return c.sheet == sheet
		}
		return false
	}) {
		return
	}
	c := &Calculator{
		sheet:                sheet,
		scale:                gurps.GlobalSettings().General.InitialEditorUIScale,
		throwingObjectWeight: fxp.Weight(fxp.One),
		terrainIndex:         slices.IndexFunc(terrain, func(t terrainModifier) bool { return t.Default }),
		weatherIndex:         slices.IndexFunc(weather, func(t terrainModifier) bool { return t.Default }),
	}
	c.Self = c

	c.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	c.SetLayout(&unison.FlexLayout{Columns: 1})

	c.createContent()

	c.scroll = unison.NewScrollPanel()
	c.scroll.SetContent(c.content, behavior.HintedFill, behavior.Fill)
	c.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	c.AddChild(c.createToolbar())
	c.AddChild(c.scroll)
	c.ClientData()[AssociatedIDKey] = sheet.Entity().ID
	c.content.ValidateScrollRoot()
	group := dgroup.Editors
	p := sheet.AsPanel()
	for p != nil {
		if _, exists := p.ClientData()[AssociatedIDKey]; exists {
			group = dgroup.SubEditors
			break
		}
		p = p.Parent()
	}
	PlaceInDock(c, group, false)
	c.content.RequestFocus()
}

// UpdateCalculator for the given owner.
func UpdateCalculator(sheet *Sheet) {
	for _, other := range AllDockables() {
		c, ok := other.(*Calculator)
		if !ok || c.sheet != sheet {
			continue
		}
		c.updateJumpingResult()
		c.updateThrowingResult()
		c.updateHikingResult()
		c.content.MarkForLayoutRecursively()
		c.content.MarkForRedraw()
		break
	}
}

func (c *Calculator) createToolbar() *unison.Panel {
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))

	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
			func() int { return c.scale },
			func(scale int) { c.scale = scale },
			nil,
			false,
			c.scroll,
		),
	)

	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	return toolbar
}

func (c *Calculator) createContent() {
	c.content = unison.NewPanel()
	c.content.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing * 2)))
	c.content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	c.addJumpingSection()
	c.addThrowingSection()
	c.addHikingSection()
}

func (c *Calculator) addJumpingSection() {
	c.content.AddChild(c.createHeader(i18n.Text("Jumping"), "BX352", "Jumping", 0))

	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	field := NewDecimalField(nil, "", i18n.Text("Jump Running Start"),
		func() fxp.Int { return c.jumpingRunningStartYards },
		func(v fxp.Int) {
			c.jumpingRunningStartYards = v
			c.updateJumpingResult()
		},
		0, fxp.Max, false, false)
	wrapper.AddChild(field)
	c.jumpingLabel = unison.NewLabel()
	wrapper.AddChild(c.jumpingLabel)
	c.content.AddChild(wrapper)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	wrapper.AddChild(NewIntegerField(nil, "", i18n.Text("Jumping Extra Effort Penalty"),
		func() int { return c.jumpingExtraEffortPenalty },
		func(v int) {
			c.jumpingExtraEffortPenalty = v
			c.updateJumpingResult()
		},
		-100, 0, false, false))
	label := unison.NewLabel()
	label.SetTitle(i18n.Text("penalty for extra effort"))
	wrapper.AddChild(label)
	c.content.AddChild(wrapper)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	divider := unison.NewSeparator()
	divider.SetBorder(unison.NewEmptyBorder(unison.NewVerticalInsets(unison.StdVSpacing * 2)))
	divider.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	wrapper.AddChild(divider)
	label = unison.NewLabel()
	label.SetTitle(i18n.Text("High Jump:"))
	wrapper.AddChild(label)
	c.highJumpResult = c.createResultLabel()
	wrapper.AddChild(c.highJumpResult)
	label = unison.NewLabel()
	label.SetTitle(i18n.Text("Broad Jump:"))
	wrapper.AddChild(label)
	c.broadJumpResult = c.createResultLabel()
	c.updateJumpingResult()
	wrapper.AddChild(c.broadJumpResult)
	c.content.AddChild(wrapper)
}

func (c *Calculator) addThrowingSection() {
	c.content.AddChild(c.createHeader(i18n.Text("Throwing"), "BX355", "Throwing", unison.StdVSpacing*3))

	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	wrapper.AddChild(NewWeightField(nil, "", i18n.Text("Object Weight"),
		c.sheet.Entity(),
		func() fxp.Weight { return c.throwingObjectWeight },
		func(v fxp.Weight) {
			c.throwingObjectWeight = v
			c.updateThrowingResult()
		},
		0, fxp.Weight(fxp.Max), false))
	label := unison.NewLabel()
	label.SetTitle(i18n.Text("object"))
	wrapper.AddChild(label)
	c.content.AddChild(wrapper)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	wrapper.AddChild(NewIntegerField(nil, "", i18n.Text("Throwing Extra Effort Penalty"),
		func() int { return c.throwingExtraEffortPenalty },
		func(v int) {
			c.throwingExtraEffortPenalty = v
			c.updateThrowingResult()
		},
		-100, 0, false, false))
	label = unison.NewLabel()
	label.SetTitle(i18n.Text("penalty for extra effort"))
	wrapper.AddChild(label)
	c.content.AddChild(wrapper)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	divider := unison.NewSeparator()
	divider.SetBorder(unison.NewEmptyBorder(unison.NewVerticalInsets(unison.StdVSpacing * 2)))
	divider.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	wrapper.AddChild(divider)
	label = unison.NewLabel()
	label.SetTitle(i18n.Text("Distance:"))
	wrapper.AddChild(label)
	c.throwingDistanceResult = c.createResultLabel()
	wrapper.AddChild(c.throwingDistanceResult)
	label = unison.NewLabel()
	label.SetTitle(i18n.Text("Damage:"))
	wrapper.AddChild(label)
	c.throwingDamageResult = c.createResultLabel()
	wrapper.AddChild(c.throwingDamageResult)
	c.updateThrowingResult()
	c.content.AddChild(wrapper)
}

func (c *Calculator) addHikingSection() {
	c.content.AddChild(c.createHeader(i18n.Text("Hiking"), "BX351", "Hiking", unison.StdVSpacing*3))

	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))

	terrainPopup := unison.NewPopupMenu[terrainModifier]()
	roadsAreClearedCheckbox := unison.NewCheckBox()
	usingSkisCheckbox := unison.NewCheckBox()
	usingSkatesCheckbox := unison.NewCheckBox()
	successfulHikingRollCheckbox := unison.NewCheckBox()
	extraEffortPenaltyField := NewIntegerField(nil, "", i18n.Text("Hiking Extra Effort Penalty"),
		func() int { return c.hikingExtraEffortPenalty },
		func(v int) {
			c.hikingExtraEffortPenalty = v
			c.updateHikingResult()
		},
		-100, 0, false, false)
	hikingAdjuster := func() {
		switch {
		case c.usingSkis:
			successfulHikingRollCheckbox.SetTitle(i18n.Text("Made a successful Skiing (B221) roll"))
			usingSkatesCheckbox.SetEnabled(false)
		case c.usingSkates:
			successfulHikingRollCheckbox.SetTitle(i18n.Text("Made a successful Skating (B220) roll"))
			usingSkisCheckbox.SetEnabled(false)
		default:
			successfulHikingRollCheckbox.SetTitle(i18n.Text("Made a successful Hiking (B200) roll"))
			usingSkatesCheckbox.SetEnabled(true)
			usingSkisCheckbox.SetEnabled(true)
		}
		w := weather[c.weatherIndex]
		roadsAreClearedCheckbox.SetEnabled(terrain[c.terrainIndex].IsRoad && (w.IsIce || w.IsSnow))
		extraEffortPenaltyField.SetEnabled(c.successfulHikingRoll)
		c.content.MarkForLayoutRecursively()
		c.content.MarkForLayoutRecursivelyUpward()
		c.content.MarkForRedraw()
	}

	label := unison.NewLabel()
	label.SetTitle(i18n.Text("Terrain:"))
	wrapper.AddChild(label)

	terrainPopup.AddItem(terrain...)
	terrainPopup.SelectIndex(c.terrainIndex)
	terrainPopup.SelectionChangedCallback = func(popup *unison.PopupMenu[terrainModifier]) {
		c.terrainIndex = popup.SelectedIndex()
		hikingAdjuster()
		c.updateHikingResult()
	}
	wrapper.AddChild(terrainPopup)

	label = unison.NewLabel()
	label.SetTitle(i18n.Text("Weather:"))
	wrapper.AddChild(label)

	weatherPopup := unison.NewPopupMenu[terrainModifier]()
	weatherPopup.AddItem(weather...)
	weatherPopup.SelectIndex(c.weatherIndex)
	weatherPopup.SelectionChangedCallback = func(popup *unison.PopupMenu[terrainModifier]) {
		c.weatherIndex = popup.SelectedIndex()
		hikingAdjuster()
		c.updateHikingResult()
	}
	wrapper.AddChild(weatherPopup)

	c.content.AddChild(wrapper)

	roadsAreClearedCheckbox.SetTitle(i18n.Text("Roads are cleared"))
	roadsAreClearedCheckbox.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	roadsAreClearedCheckbox.ClickCallback = func() {
		c.roadsAreCleared = roadsAreClearedCheckbox.State == check.On
		c.updateHikingResult()
	}
	c.content.AddChild(roadsAreClearedCheckbox)

	usingSkisCheckbox.SetTitle(i18n.Text("Using skis"))
	usingSkisCheckbox.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	usingSkisCheckbox.ClickCallback = func() {
		c.usingSkis = usingSkisCheckbox.State == check.On
		hikingAdjuster()
		c.updateHikingResult()
	}
	c.content.AddChild(usingSkisCheckbox)

	usingSkatesCheckbox.SetTitle(i18n.Text("Using skates"))
	usingSkatesCheckbox.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	usingSkatesCheckbox.ClickCallback = func() {
		c.usingSkates = usingSkatesCheckbox.State == check.On
		hikingAdjuster()
		c.updateHikingResult()
	}
	c.content.AddChild(usingSkatesCheckbox)

	hikingAdjuster()
	successfulHikingRollCheckbox.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	successfulHikingRollCheckbox.ClickCallback = func() {
		c.successfulHikingRoll = successfulHikingRollCheckbox.State == check.On
		hikingAdjuster()
		c.updateHikingResult()
	}
	c.content.AddChild(successfulHikingRollCheckbox)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	wrapper.AddChild(extraEffortPenaltyField)
	label = unison.NewLabel()
	label.SetTitle(i18n.Text("penalty for extra effort."))
	wrapper.AddChild(label)
	c.content.AddChild(wrapper)

	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing * 2}))
	divider := unison.NewSeparator()
	divider.SetBorder(unison.NewEmptyBorder(unison.NewVerticalInsets(unison.StdVSpacing * 2)))
	divider.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	wrapper.AddChild(divider)
	c.hikingResult = c.createResultLabel()
	c.updateHikingResult()
	wrapper.AddChild(c.hikingResult)
	label = unison.NewLabel()
	label.SetTitle(i18n.Text(" per full day"))
	wrapper.AddChild(label)
	c.content.AddChild(wrapper)
}

func (c *Calculator) createResultLabel() *unison.Label {
	label := unison.NewLabel()
	label.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.DefaultLabelTheme.Font.Descriptor()
			desc.Weight = weight.Bold
			return desc
		},
	}
	return label
}

func (c *Calculator) createHeader(text, linkRef, linkHighlight string, topMargin float32) *unison.Panel {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{Columns: 3})
	if topMargin > 0 {
		wrapper.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: topMargin}))
	}

	first := unison.NewLabel()
	first.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.LabelFont.Descriptor()
			desc.Size += 2
			desc.Weight = weight.Bold
			return desc
		},
	}
	first.SetTitle(text + " (")
	wrapper.AddChild(first)

	linkTheme := unison.DefaultLinkTheme
	linkTheme.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.LabelFont.Descriptor()
			desc.Weight = weight.Bold
			return desc
		},
	}
	link := unison.NewLink(linkRef, "", linkRef, linkTheme, func(_ unison.Paneler, _ string) {
		OpenPageReference(linkRef, linkHighlight, nil)
	})
	wrapper.AddChild(link)

	last := unison.NewLabel()
	last.Font = first.Font
	last.SetTitle(")")
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
	return AttemptCloseForDockable(c)
}

// UndoManager implements unison.UndoManagerProvider
func (c *Calculator) UndoManager() *unison.UndoManager {
	return c.undoMgr
}

func (c *Calculator) computeJump(broad bool) fxp.Int {
	entity := c.sheet.Entity()
	basicMove := entity.Attributes.Current(gurps.BasicMoveID)
	basicMoveWithoutRun := basicMove

	// Adjust Basic Move for running
	if c.jumpingRunningStartYards > 0 {
		enhMove := -fxp.One
		gurps.Traverse(func(t *gurps.Trait) bool {
			if strings.EqualFold(t.NameWithReplacements(), "enhanced move (ground)") && t.IsLeveled() {
				if enhMove == -fxp.One {
					enhMove = t.Levels
				} else {
					enhMove += t.Levels
				}
			}
			return false
		}, true, false, entity.Traits...)
		basicMove += c.jumpingRunningStartYards
		if enhMove > 0 {
			if adjusted := basicMoveWithoutRun.Mul(enhMove + fxp.One); adjusted > basicMove {
				basicMove = adjusted
			}
		}
	}

	// Adjust Basic Move for Jumping skill
	gurps.Traverse(func(s *gurps.Skill) bool {
		if strings.EqualFold(s.NameWithReplacements(), "jumping") {
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
	st := entity.LiftingStrength()
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
	gurps.Traverse(func(t *gurps.Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), "super jump") && t.IsLeveled() {
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
	c.highJumpResult.SetTitle(c.distanceToText(c.computeJump(false)))
	c.highJumpResult.MarkForLayoutRecursivelyUpward()
	c.broadJumpResult.SetTitle(c.distanceToText(c.computeJump(true)))
	c.broadJumpResult.MarkForLayoutRecursivelyUpward()
	var units string
	if c.useMeters() {
		units = i18n.Text("meter")
	} else {
		units = i18n.Text("yard")
	}
	c.jumpingLabel.SetTitle(fmt.Sprintf(i18n.Text("%s running start"), units))
	c.jumpingLabel.MarkForLayoutRecursivelyUpward()
}

func (c *Calculator) updateThrowingResult() {
	if c.throwingObjectWeight <= 0 {
		c.throwingDistanceResult.SetTitle(i18n.Text("None"))
		c.throwingDamageResult.SetTitle(i18n.Text("None"))
		return
	}
	entity := c.sheet.Entity()

	// Determine bonuses for skills
	var distanceBonus, damageBonus int
	gurps.Traverse(func(s *gurps.Skill) bool {
		switch strings.ToLower(s.NameWithReplacements()) {
		case "throwing art":
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
		case "throwing":
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
	st := entity.LiftingStrength() - entity.LiftingStrengthBonus
	if c.throwingExtraEffortPenalty < 0 {
		st = st.Mul(fxp.From(-5*c.throwingExtraEffortPenalty).Div(fxp.Hundred) + fxp.One).Trunc()
	}
	st += fxp.From(distanceBonus)
	basicLift := entity.BasicLiftForST(st)
	var weightRatio fxp.Int
	if basicLift > 0 {
		weightRatio = fxp.Int(c.throwingObjectWeight).Div(fxp.Int(basicLift))
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
		c.throwingDistanceResult.SetTitle(i18n.Text("The object is too heavy for you to throw"))
		c.throwingDamageResult.SetTitle(i18n.Text("None"))
		return
	}

	// Determine damage based on weight ratio
	thrust := entity.Thrust()
	thrust.Modifier += thrust.Count * damageBonus
	basicLift = entity.BasicLiftForST(st - fxp.From(distanceBonus))
	if basicLift > 0 {
		weightRatio = fxp.Int(c.throwingObjectWeight).Div(fxp.Int(basicLift))
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

	c.throwingDistanceResult.SetTitle(c.distanceToText(inches))
	c.throwingDamageResult.SetTitle(thrust.StringExtra(entity.SheetSettings.UseModifyingDicePlusAdds))
	c.throwingDistanceResult.MarkForLayoutRecursivelyUpward()
	c.throwingDamageResult.MarkForLayoutRecursivelyUpward()
}

func (c *Calculator) updateHikingResult() {
	entity := c.sheet.Entity()
	distance := fxp.From(entity.Move(entity.EncumbranceLevel(false)) * 10)

	// Adjust for enhanced move (ground), if any
	enhMove := -fxp.One
	gurps.Traverse(func(t *gurps.Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), "enhanced move (ground)") && t.IsLeveled() {
			if enhMove == -fxp.One {
				enhMove = t.Levels
			} else {
				enhMove += t.Levels
			}
		}
		return false
	}, true, false, entity.Traits...)
	if enhMove > 0 {
		distance = distance.Mul(fxp.One + enhMove)
	}

	// Adjust for terrain
	t := terrain[c.terrainIndex]
	mod := t.Modifier
	if t.IsIce && c.usingSkates {
		mod = fxp.OneAndAQuarter
	}
	if t.IsSnow && c.usingSkis {
		mod = fxp.One
	}

	// Adjust for weather
	w := weather[c.weatherIndex]
	switch {
	case w.IsRain:
		if t.IsRoad {
			if t.ModifierInRain != 0 {
				mod = t.ModifierInRain
			}
		} else {
			mod = mod.Mul(w.Modifier)
		}
	case w.IsSnow:
		if t.IsRoad {
			mod = fxp.One
		}
		if (!t.IsRoad || !c.roadsAreCleared) && !c.usingSkis {
			mod = mod.Mul(w.Modifier)
		}
	case w.IsIce:
		if t.IsRoad {
			mod = fxp.One
		}
		if (!t.IsRoad || !c.roadsAreCleared) && !c.usingSkates {
			mod = mod.Mul(w.Modifier)
		}
	}
	distance = distance.Mul(mod)

	// Adjust for making the hiking/skiing/skating check
	mod = fxp.One
	if c.successfulHikingRoll {
		mod = fxp.OnePointTwo
		if c.hikingExtraEffortPenalty < 0 {
			mod += fxp.From(-5 * c.hikingExtraEffortPenalty).Div(fxp.Hundred)
		}
	}
	distance = distance.Mul(mod)

	var units string
	if c.useMeters() {
		// miles -> inches -> GURPS kilometers
		distance = distance.Mul(fxp.MileInInches).Div(fxp.ThirtySixThousand)
		if distance == fxp.One {
			units = i18n.Text("kilometer")
		} else {
			units = i18n.Text("kilometers")
		}
	} else {
		if distance == fxp.One {
			units = i18n.Text("mile")
		} else {
			units = i18n.Text("miles")
		}
	}
	c.hikingResult.SetTitle(fmt.Sprintf("%s %s", distance.Comma(), units))
	c.hikingResult.MarkForLayoutRecursivelyUpward()
}

func (c *Calculator) useMeters() bool {
	units := c.sheet.Entity().SheetSettings.DefaultLengthUnits
	return units == fxp.Centimeter || units == fxp.Meter || units == fxp.Kilometer
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
