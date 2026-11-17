// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculators

import (
	"fmt"
	"image"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// calculatorKey and newCharacterSheetKey are the key binding IDs of the ux actions that open the calculators and a new
// character sheet.
const (
	calculatorKey        = "calculator"
	newCharacterSheetKey = "new.char.sheet"
)

// openNewCharacterSheet opens a new character sheet through its action and returns it.
func openNewCharacterSheet(t *testing.T, screen *unison.HeadlessScreen) *ux.Sheet {
	t.Helper()
	sheet, ok := uxtest.OpenedByAction(t, screen, uxtest.ActionForKey(t, newCharacterSheetKey)).(*ux.Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	return sheet
}

// openCalculator opens the calculators from their menu action, which preselects the active sheet, and returns the
// dockable holding them.
func openCalculator(t *testing.T, screen *unison.HeadlessScreen) *Dockable {
	t.Helper()
	calc, ok := uxtest.OpenedByAction(t, screen, uxtest.ActionForKey(t, calculatorKey)).(*Dockable)
	if !ok {
		t.Fatal("the action must open the calculators")
	}
	return calc
}

// selectCalculatorTab clicks the tab for the given calculator the way a user would and checks that its content is
// what the dockable then shows.
func selectCalculatorTab(t *testing.T, screen *unison.HeadlessScreen, calc *Dockable, tab calculatorTab) {
	t.Helper()
	index := slices.Index(calc.tabs, tab)
	if index < 0 {
		t.Fatalf("%s is not one of the calculators", tab.title())
	}
	screen.Click(screen.PanelCenter(uxtest.PanelsOfType[*unison.Button](calc.tabBar.AsPanel())[index]))
	var shown []*unison.Panel
	var selected int
	screen.Do(func() {
		shown = calc.slot.Children()
		selected = calc.tabBar.SelectedIndex()
	})
	if selected != index || len(shown) != 1 || shown[0] != tab.panel() {
		t.Fatalf("clicking the %s tab must show that calculator alone", tab.title())
	}
}

// calculatorHeader returns the header a calculator's content starts with.
func calculatorHeader(t *testing.T, screen *unison.HeadlessScreen, content *unison.Panel) *ux.TextLabel {
	t.Helper()
	var header *ux.TextLabel
	screen.Do(func() {
		if children := content.Children(); len(children) > 0 {
			if label, isLabel := children[0].Self.(*ux.TextLabel); isLabel {
				header = label
			}
		}
	})
	if header == nil {
		t.Fatal("the calculator must start with its header")
	}
	return header
}

// TestCalculatorTabs verifies the calculators open as one dockable with a tab per calculator, in the documented order,
// that clicking a tab swaps that calculator's content in beneath the tab bar, and that a second request brings the open
// dockable forward rather than opening another. TestCalculatorSheetSources covers the sheet they were opened from.
func TestCalculatorTabs(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openNewCharacterSheet(t, screen)
	calc := openCalculator(t, screen)

	var titles []string
	var shown []*unison.Panel
	screen.Do(func() {
		for _, button := range uxtest.PanelsOfType[*unison.Button](calc.tabBar.AsPanel()) {
			titles = append(titles, button.Text.String())
		}
		shown = calc.slot.Children()
	})
	c.Equal([]string{
		"Explosions & Area Attacks", "Scatter", "Demolition", "Collisions & Falls", "Jumping", "Throwing", "Hiking",
	}, titles, "the tabs must be labeled with the calculators, in order")
	c.Equal([]*unison.Panel{calc.explosion.content}, shown, "the first tab must be showing when the calculators open")

	for _, tab := range calc.tabs {
		selectCalculatorTab(t, screen, calc, tab)
	}
	uxtest.CaptureScreen(t, c, screen, "calculator_hiking_tab")
	selectCalculatorTab(t, screen, calc, calc.jumping)

	// Choosing the menu item again brings the open calculators forward, leaving the chosen tab alone.
	var opened int
	screen.Do(func() {
		before := len(ux.AllDockables())
		uxtest.ActionForKey(t, calculatorKey).Execute(nil)
		opened = len(ux.AllDockables()) - before
	})
	c.Equal(0, opened, "a second request must not open a second set of calculators")
	var selected int
	screen.Do(func() { selected = calc.tabBar.SelectedIndex() })
	c.Equal(slices.Index(calc.tabs, calculatorTab(calc.jumping)), selected,
		"bringing the calculators forward must not change the tab")

	// Squeezing the workspace narrower than the rows must leave the calculators scrolling sideways rather than
	// clipping: the scroll panel's content has to stay as wide as the calculator needs, which includes the rows
	// nested in the groups that come and go, such as the target's posture popup on the explosions tab.
	selectCalculatorTab(t, screen, calc, calc.collision)
	var viewWidth, contentWidth, neededWidth float32
	screen.Do(func() {
		wnd.SetFrameRect(geom.NewRect(0, 0, 900, 900))
		wnd.ValidateLayout()
		viewWidth = calc.scroll.ContentView().FrameRect().Width
		contentWidth = calc.scroll.Content().AsPanel().FrameRect().Width
		_, needed, _ := calc.collision.content.Sizes(geom.Size{Width: viewWidth})
		neededWidth = needed.Width
	})
	c.True(viewWidth < neededWidth, "the narrowed view must be too narrow for the collision rows")
	c.Equal(neededWidth, contentWidth, "the scroll content must stay as wide as the calculator needs")
	selectCalculatorTab(t, screen, calc, calc.explosion)
	var popupRight, viewRight, contentRight float32
	screen.Do(func() {
		wnd.ValidateLayout()
		view := calc.scroll.ContentView()
		viewRight = view.PointToRoot(geom.NewPoint(view.FrameRect().Width, 0)).X
		content := calc.scroll.Content().AsPanel()
		contentRight = content.PointToRoot(geom.NewPoint(content.FrameRect().Width, 0)).X
		if popup, found := uxtest.FirstPanelOfType[*unison.PopupMenu[calculator.TargetPostureChoice]](calc.explosion.content); found {
			popupRight = popup.PointToRoot(geom.NewPoint(popup.FrameRect().Width, 0)).X
		}
	})
	uxtest.CaptureScreen(t, c, screen, "calculator_narrow_explosion")
	c.True(popupRight > 0, "the explosions tab must offer a posture popup")
	c.True(popupRight > viewRight, "the narrowed view must be too narrow for the posture popup")
	c.True(popupRight <= contentRight, "the posture popup must lie within the scroll content, however narrow the view")

	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
}

// TestCalculatorChoicesFollowLanguage verifies that the choices the calculators offer in their popups are translated
// when a calculator is created rather than when the package initializes. The language setting is only applied after the
// package has initialized, so a table translated at that point would show a user who chose a language other than the
// system's the system's language in those popups alone.
func TestCalculatorChoicesFollowLanguage(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	const mark = "»"
	i18n.SetLocalizer(func(text string) string { return mark + text })
	t.Cleanup(func() { i18n.SetLocalizer(nil) })

	type table struct {
		what  string
		names []string
	}
	var tables []table
	var startingExplosive string
	screen.Do(func() {
		hiking := newHikingCalculator()
		collision := newCollisionCalculator()
		demolition := newDemolitionCalculator()
		explosion := newExplosionCalculator()
		scatter := newScatterCalculator()
		throwing := newThrowingCalculator()
		tables = []table{
			{"terrain", choiceNames(hiking.terrain)},
			{"weather", choiceNames(hiking.weather)},
			{"heat", choiceNames(hiking.heat)},
			{"intensity", choiceNames(hiking.intensity)},
			{"fitness", choiceNames(hiking.fitness)},
			{"Recover Energy", choiceNames(hiking.recoverEnergy)},
			{"collision scenario", choiceNames(collision.scenarios)},
			{"collision shape", choiceNames(collision.shapes)},
			{"collision surface", choiceNames(collision.surfaces)},
			{"terminal velocity", choiceNames(collision.terminals)},
			{"collision angle", choiceNames(collision.angles)},
			{"collision restraint", choiceNames(collision.restraints)},
			{"demolition mode", choiceNames(demolition.modes)},
			{"explosive", choiceNames(demolition.explosives)},
			{"attack type", choiceNames(explosion.attackTypes)},
			{"explosion environment", choiceNames(explosion.environments)},
			{"explosion posture", choiceNames(explosion.postures)},
			{"explosion situation", choiceNames(explosion.situations)},
			{"scatter cause", choiceNames(scatter.causes)},
			{"Throwing tier", choiceNames(throwing.throwingTiers)},
			{"Throwing Art tier", choiceNames(throwing.throwingArtTiers)},
		}
		startingExplosive = demolition.explosives[demolition.explosiveIndex].title
	})
	for _, one := range tables {
		c.NotEqual(0, len(one.names), fmt.Sprintf("the %s popup must offer choices", one.what))
		for _, name := range one.names {
			c.True(strings.HasPrefix(name, mark),
				fmt.Sprintf("the %s choice %q must be translated when the calculator is created", one.what, name))
		}
	}
	c.Equal("TNT", startingExplosive, "the demolition calculator must still start on TNT, whose name is not translated")
}

// choiceNames returns what a popup offering the items shows for each of them.
func choiceNames[T fmt.Stringer](items []T) []string {
	names := make([]string, len(items))
	for i, one := range items {
		names[i] = one.String()
	}
	return names
}

// TestEveryCalculatorControlHasAnAccessibleName opens the calculators with a character sheet active and fails for
// every control on any of their tabs that an assistive technology would be handed with no name; see
// uxtest.AXNameAudit. ux's own audit covers the rest of the workspace.
func TestEveryCalculatorControlHasAnAccessibleName(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	audit := uxtest.NewAXNameAudit(t, screen, wnd)
	sheet := openNewCharacterSheet(t, screen)
	calc, isCalc := audit.Open(func() { Display(sheet) }).(*Dockable)
	if !isCalc {
		t.Fatal("the calculators did not open")
	}
	for i := range calc.tabs {
		var title string
		screen.Do(func() {
			calc.tabBar.SelectTab(i)
			title = calc.tabs[i].title()
		})
		audit.Check("calculator: "+title, calc)
	}
	audit.Report()
}

// With a screen reader listening and the FocusForReading setting on, a disabled control, such as the hiking
// calculator's Move while a sheet supplies it, is a tab stop. It stays disabled otherwise: a screen reader is told it
// is unavailable, typing does not change it, and while focused it keeps showing the value it is given.
func TestDisabledControlsJoinTheTabOrderForScreenReaders(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	sheet := openNewCharacterSheet(t, screen)
	calc := openCalculator(t, screen)
	hiking := calc.hiking
	selectCalculatorTab(t, screen, calc, hiking)
	move := hiking.moveField
	rest := hiking.restField
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	enabled := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Enabled() })
		return result
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	text := func() (result string) {
		screen.Do(func() { result = move.Text() })
		return result
	}
	c.False(enabled(move), "the sheet supplies the Move, so its field must be disabled")
	c.True(enabled(rest), "the rest is not the sheet's to supply, so its field must be enabled")

	setFocusForReading(false)
	c.False(focusable(move), "a disabled field must not be a tab stop while the setting is off")
	c.True(focusable(rest), "a field that can be used is always a tab stop")

	setFocusForReading(true)
	c.True(focusable(move), "a disabled field must be a tab stop while the setting is on")
	c.True(focusable(rest))
	c.False(enabled(move), "taking the focus must not enable the field")
	node := screen.AccessibilityNodeFor(move)
	if node == nil {
		t.Fatal("the Move field must be described")
	}
	c.True(node.Disabled, "a screen reader must be told the field is unavailable")
	before := text()
	c.NotEqual("", before, "the sheet must have supplied a Move")
	screen.Do(move.RequestFocus)
	c.Equal(move.AsPanel(), focus(), "the disabled field takes the focus when asked to")
	screen.Type("9")
	c.Equal(before, text(), "nothing typed on a disabled field may change it")
	c.Equal(move.AsPanel(), focus(), "and the focus stays where it was")
	var want string
	screen.Do(func() {
		entity := sheet.Entity()
		attr := entity.Attributes.Set[gurps.BasicMoveID]
		attr.SetMaximum(attr.Maximum() + fxp.Three)
		entity.Recalculate()
		hiking.sheetChanged(sheet)
		want = strconv.Itoa(hiking.move)
	})
	c.NotEqual(before, want, "the sheet's Move must have changed")
	c.Equal(want, text(), "a disabled field holding the focus must show the number it is given")
	c.Equal(move.AsPanel(), focus())

	screen.Do(rest.RequestFocus)
	c.Equal(rest.AsPanel(), focus())
	c.Equal(want, text(), "the focus leaving a disabled field must leave it as it was")
	var got string
	screen.Do(func() { got = strconv.Itoa(hiking.move) })
	c.Equal(want, got, "and the number behind it too")

	var target *unison.Panel
	screen.Do(func() {
		ux.FocusFirstContent(unison.NewPanel(), hiking.content)
		target = wnd.Focus()
	})
	c.NotNil(target, "the calculator holds controls that can be used")
	if target != nil {
		c.True(target.Enabled(), "the focus must start on a control that can be used, not on a %T", target.Self)
	}

	setFocusForReading(false)
	c.False(focusable(move), "a disabled field must not be a tab stop once the setting is off again")
	c.True(focusable(rest))
}

func TestStaticTextShowsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	calc := openCalculator(t, screen)
	explosion := calc.explosion
	selectCalculatorTab(t, screen, calc, explosion)
	setFocusForReading(true)
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	ringInWindow := func() (ring geom.Rect, ok bool) {
		screen.Do(func() {
			var clip geom.Rect
			if ring, clip, ok = ux.StaticTextFocusRing(wnd, wnd.Content()); ok {
				c.Equal(ring, ring.Intersect(clip), "all of the ring can be seen")
				ring = wnd.Content().RectToRoot(ring)
			}
		})
		return ring, ok
	}
	capture := func() *image.NRGBA {
		screen.Sync()
		img := screen.CaptureWindow(wnd)
		if img == nil {
			t.Fatal("the window could not be captured")
		}
		return img
	}
	changedWithin := func(before, after *image.NRGBA, r geom.Rect) (count int) {
		scale := float32(after.Bounds().Dx()) / wnd.ContentRect().Width
		area := image.Rect(int(r.X*scale), int(r.Y*scale), int(r.Right()*scale), int(r.Bottom()*scale)).
			Intersect(after.Bounds()).Intersect(before.Bounds())
		for y := area.Min.Y; y < area.Max.Y; y++ {
			for x := area.Min.X; x < area.Max.X; x++ {
				if before.NRGBAAt(x, y) != after.NRGBAAt(x, y) {
					count++
				}
			}
		}
		return count
	}

	screen.Do(explosion.hotFragmentsBox.RequestFocus)
	c.Equal(explosion.hotFragmentsBox.AsPanel(), focus())
	_, ok := ringInWindow()
	c.False(ok, "a checkbox shows the focus itself")
	before := capture()

	screen.KeyPress(unison.KeyTab, mod.None)
	target, isLabel := focus().Self.(*unison.Label)
	if !isLabel {
		t.Fatalf("Tab from Hot fragments must go to the Target subheader, not to a %T", focus().Self)
	}
	c.Equal("Target", target.String())
	ring, ok := ringInWindow()
	c.True(ok, "static text holding the focus is given a ring")
	var frame geom.Rect
	screen.Do(func() { frame = target.RectToRoot(target.ContentRect(true)) })
	c.Equal(ring, ring.Intersect(frame), "the ring %v lies within the label %v", ring, frame)
	c.True(ring.Width < frame.Width/2, "the ring goes around the text, not the width of the column: %v in %v", ring, frame)
	c.True(ring.Width > 0 && ring.Height > 0)
	after := capture()
	c.True(changedWithin(before, after, ring) > 0, "the ring is drawn")

	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(explosion.target.popup.AsPanel(), focus(), "Tab goes on to the target's Source")
	_, ok = ringInWindow()
	c.False(ok, "a popup shows the focus itself")
	gone := capture()
	c.True(changedWithin(after, gone, ring) > 0, "the ring is gone once the focus has moved on")

	header := calculatorHeader(t, screen, explosion.content)
	screen.Do(header.RequestFocus)
	c.Equal(header.AsPanel(), focus())
	_, ok = ringInWindow()
	c.False(ok, "a label holding links shows the focus itself")

	setFocusForReading(false)
	screen.Do(explosion.hotFragmentsBox.RequestFocus)
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(explosion.target.popup.AsPanel(), focus(), "Tab goes from Hot fragments to the target's Source")
}

// A label that follows a field, such as its units, is the field's description rather than an element of its own, so it
// is no tab stop with static text in the Tab order.
func TestTrailingLabelsDescribeTheirFields(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	openNewCharacterSheet(t, screen)
	calc := openCalculator(t, screen)
	hiking := calc.hiking
	selectCalculatorTab(t, screen, calc, hiking)
	setFocusForReading(true)
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}

	var trailing *ux.TextLabel
	screen.Do(func() {
		children := hiking.restField.Parent().Children()
		if len(children) == 2 {
			if label, isLabel := children[1].Self.(*ux.TextLabel); isLabel {
				trailing = label
			}
		}
	})
	if trailing == nil {
		t.Fatal("the rest field must be followed by its label")
	}
	node := screen.AccessibilityNodeFor(hiking.restField)
	if node == nil {
		t.Fatal("the rest field must be described")
	}
	c.Equal("minutes of rest halfway through the day (0 for none)", node.Description)
	c.False(focusable(trailing), "the label is heard with the field, so it is no tab stop")
	if label := screen.AccessibilityNodeFor(trailing); label != nil {
		c.True(label.Ignored, "the label is heard with the field, so it is not described, but is %+v", label)
	}
	// The description follows the label's text as it changes.
	screen.Do(func() { trailing.SetTitle("minutes of rest") })
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	node = screen.AccessibilityNodeFor(hiking.restField)
	c.Equal("minutes of rest", node.Description)
	// A label holding a link is described, and is a tab stop so that the link can be reached.
	screen.Do(func() { trailing.SetTitle("minutes of rest (B426)") })
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	c.True(focusable(trailing), "a label holding a link is a tab stop")
	label := screen.AccessibilityNodeFor(trailing)
	if label == nil {
		t.Fatal("a label holding a link must be described")
	}
	c.False(label.Ignored)
	c.Equal(role.Label, label.Role)
	c.Equal(1, len(label.Children), "the link is an element within the label")
	c.Equal("minutes of rest (B426)", screen.AccessibilityNodeFor(hiking.restField).Description)
}
