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
	"encoding/json/jsontext"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	uncheck "github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/zeebo/xxh3"
)

// testSkullID is the ID of the skull in the default body.
const testSkullID = "skull"

// showFeaturesPanel shows a featuresPanel for the features of the owner in a window, within a host, so that its
// rebuilds run and its edits can be undone. The panel is expanded, as most tests look at its rows; see
// TestFeaturesPanelStartingState for how it starts out.
func showFeaturesPanel(t *testing.T, screen *unison.HeadlessScreen, entity *gurps.Entity, owner fmt.Stringer, features *gurps.Features, forEquipmentModifier bool) (*featuresPanel, *prereqUndoHost) {
	var p *featuresPanel
	host := &prereqUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
	screen.Do(func() {
		host.Self = host
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		host.KeyDownCallback = func(keyCode unison.KeyCode, _ mod.Modifiers, _ bool) bool {
			if keyCode == unison.KeyEscape {
				host.escapes++
			}
			return true
		}
		p = newFeaturesPanel(entity, owner, features, forEquipmentModifier)
		// Expanded ahead of showing, so that the window is sized for the rows.
		if p.collapse.collapsed {
			p.collapse.toggle()
		}
		host.AddChild(p)
	})
	showInTestWindow(t, screen, 900, host)
	return p, host
}

// newTestFeatures returns a skill bonus, a switchable DR bonus to the skull, and a feature this version of GCS doesn't
// understand, each belonging to the owner.
func newTestFeatures(owner fmt.Stringer) gurps.Features {
	skill := gurps.NewSkillBonus()
	skill.NameCriteria.Qualifier = "Streetwise"
	skill.SetOwner(owner)
	dr := gurps.NewDRBonus()
	dr.Locations = []string{testSkullID}
	dr.SetSwitchable(true)
	dr.SetOwner(owner)
	return gurps.Features{skill, dr, gurps.NewUnknownFeature("future", jsontext.Value(`{"type":"future"}`))}
}

// featureTypes returns the types of the features.
func featureTypes(list gurps.Features) []feature.Type {
	types := make([]feature.Type, 0, len(list))
	for _, one := range list {
		types = append(types, one.FeatureType())
	}
	return types
}

// featuresHash returns a hash of the features.
func featuresHash(list gurps.Features) uint64 {
	h := xxh3.New()
	for _, one := range list {
		one.Hash(h)
	}
	return h.Sum64()
}

// featuresPopup returns the popup with the reference key, failing the test if there is none.
func featuresPopup[T comparable](c check.Checker, p *featuresPanel, key string) *unison.PopupMenu[T] {
	popup, ok := p.FindRefKey(key).Self.(*unison.PopupMenu[T])
	c.True(ok, "expected a popup keyed %s", key)
	return popup
}

// featuresCheckBox returns the checkbox with the reference key, failing the test if there is none.
func featuresCheckBox(c check.Checker, p *featuresPanel, key string) *unison.CheckBox {
	box, ok := p.FindRefKey(key).Self.(*unison.CheckBox)
	c.True(ok, "expected a checkbox keyed %s", key)
	return box
}

// clickFeatureCheckBox puts the checkbox into the given state and runs its click callback, as a user's click does.
func clickFeatureCheckBox(box *unison.CheckBox, on bool) {
	box.State = uncheck.FromBool(on)
	box.ClickCallback()
}

// TestFeaturesPanelSentences checks that each closed row reads as its feature's description, which a screen reader
// hears as a disclosure, and that an unknown feature's row is static text that doesn't open.
func TestFeaturesPanelSentences(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() {
		for i, want := range []struct {
			text string
			role role.Enum
		}{
			{"+1 to skill Streetwise", role.DisclosureTriangle},
			{"+1 DR to the Skull, only while switched on", role.DisclosureTriangle},
			// An unknown feature's sentence is static text.
			{`Unknown feature type "future"; it will be preserved, but ignored`, role.Label},
		} {
			sentence, ok := p.FindRefKey(fmt.Sprintf("%d%s", i, keySentence)).Self.(*sentenceButton)
			c.True(ok, "row %d is a sentence", i)
			if !ok {
				continue
			}
			c.Equal(want.text, sentence.plainText())
			c.Equal(want.text, sentence.Accessibility.Name)
			c.Equal(want.role, sentence.Accessibility.Role)
			c.Equal(want.role == role.Label, sentence.Tooltip != nil, "only an unknown feature's says why")
		}
		c.NotNil(p.FindRefKey("2"+keyMore), "but can still be deleted")
		p.open = "2"
		p.rebuild("")
	})
	c.Equal("", p.open, "an unknown feature doesn't open")
	screen.Do(func() { c.Nil(p.FindRefKey("2" + keyFirst)) })
}

// TestFeaturesPanelOpenAndClose checks that a row opens to its editor, focusing its first field, that only one row is
// open at a time, and that Done and Escape close it, the latter without reaching the editor.
func TestFeaturesPanelOpenAndClose(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		c.NotNil(p.FindRefKey("0" + keyFirst))
		c.Equal("0:amount", p.Window().Focus().RefKey, "opening a row focuses its first field")
		p.toggle("1")
	})
	screen.Do(func() {
		c.Nil(p.FindRefKey("0"+keyFirst), "opening another row closes the first")
		c.NotNil(p.FindRefKey("1" + keyFirst))
		done, ok := p.FindRefKey("1:done").Self.(*unison.Button)
		c.True(ok)
		done.ClickCallback()
	})
	screen.Do(func() {
		c.Equal("", p.open)
		c.Equal(p.FindRefKey("1"+keySentence), p.Window().Focus(), "closing gives the focus to the sentence")
		p.toggle("1")
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", p.open, "Escape closes the open row")
	c.Equal(0, host.escapes, "and doesn't reach the editor")
}

// TestFeaturesPanelCollapse checks that the title bar collapses the panel to a paragraph of every row's sentence and
// expands it again, as does the paragraph, from a click or the keyboard, that a screen reader hears whether it is
// expanded, that collapsing hands the focus from the rows to the title bar, and that it is no edit and keeps the open
// row open.
func TestFeaturesPanelCollapse(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	c.Equal(3, len(features), "precondition: three features")
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	hash := featuresHash(features)
	expanded := func() bool {
		node := &accessibility.Node{}
		p.collapse.Accessibility.Callback(node)
		c.True(node.Expandable)
		return node.Expanded
	}
	summary := func() *sentenceButton {
		b, ok := p.FindRefKey(featureSummaryKey).Self.(*sentenceButton)
		c.True(ok, "a collapsed panel shows a paragraph")
		return b
	}
	screen.Do(func() {
		c.Equal(role.DisclosureTriangle, p.collapse.Accessibility.Role)
		c.Equal("Features", p.collapse.Accessibility.Name, "the title bar is named for the title")
		c.True(p.collapse.Focusable())
		c.True(expanded(), "precondition: the panel is expanded")
		c.Nil(p.FindRefKey(featureSummaryKey))
		p.toggle("0")
	})
	screen.Do(func() {
		c.Equal("0:amount", p.Window().Focus().RefKey, "precondition: the focus is in the open row")
		p.collapse.MouseUpCallback(geom.Point{X: 1, Y: 1}, 0, mod.None)
	})
	screen.Do(func() {
		c.False(expanded(), "clicking the title bar collapses the panel")
		for _, key := range []string{"0" + keyFirst, "1" + keySentence, featureAddKey} {
			c.Nil(p.FindRefKey(key), "collapsing hides %s", key)
		}
		c.Equal("+1 to skill Streetwise. +1 DR to the Skull, only while switched on. "+
			`Unknown feature type "future"; it will be preserved, but ignored.`, summary().Accessibility.Name)
		c.Equal(p.collapse.AsPanel(), p.Window().Focus(), "the focus moves from the rows to the title bar")
		// Nothing but the title strip and the paragraph takes room: the border's insets, as under a plain title, and
		// the 2 point inset all round.
		paragraph := summary()
		_, pref, _ := p.Sizes(geom.Size{Width: p.FrameRect().Width})
		_, text, _ := paragraph.Sizes(geom.Size{Width: paragraph.FrameRect().Width})
		plain := &TitledBorder{Title: p.collapse.border.Title, Font: p.collapse.border.Font}
		c.Equal(plain.Insets().Height()+4+text.Height, pref.Height, "the collapsed panel is only as tall as it shows")
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("0", p.open, "the open row stays open, even through Escape")
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() {
		c.True(expanded(), "Space expands the panel")
		c.Nil(p.FindRefKey(featureSummaryKey), "and hides the paragraph")
		c.NotNil(p.FindRefKey("0"+keyFirst), "with the open row still open")
		c.Equal(p.collapse.AsPanel(), p.Window().Focus(), "and the focus left on the title bar")
	})
	screen.KeyPress(unison.KeyReturn, mod.None)
	screen.Do(func() {
		c.False(expanded(), "Return collapses it")
		summary().RequestFocus()
	})
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() {
		c.True(expanded(), "the paragraph expands the panel")
		c.Nil(p.FindRefKey(featureSummaryKey))
		c.Equal("0:amount", p.Window().Focus().RefKey, "and the open row takes the focus from it")
	})
	c.Equal(hash, featuresHash(features), "collapsing changes nothing")
	c.False(host.mgr.CanUndo(), "and is not an edit")
	c.Equal(0, host.modified, "nor marks the editor modified")
}

// TestFeaturesPanelStartingState checks that a panel with features starts out collapsed, leaving the focus where the
// window put it, and one without starts out open with its placeholder, and that adding the first feature or deleting the last leaves it
// open.
func TestFeaturesPanelStartingState(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	var empty gurps.Features
	c.Equal(3, len(features), "precondition: three features")
	c.Equal(0, len(empty), "precondition: the list has no features")
	var full, open *featuresPanel
	var field *unison.Field
	host := &prereqUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
	screen.Do(func() {
		host.Self = host
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		field = unison.NewField()
		host.AddChild(field)
		full = newFeaturesPanel(entity, owner, &features, false)
		open = newFeaturesPanel(entity, owner, &empty, false)
		host.AddChild(full)
		host.AddChild(open)
	})
	wnd := showInTestWindow(t, screen, 900, host)
	screen.Do(func() {
		c.True(full.collapse.collapsed, "a panel with features starts out collapsed")
		c.NotNil(full.FindRefKey(featureSummaryKey), "showing the paragraph")
		c.Nil(full.FindRefKey("0"+keySentence), "in place of the rows")
		c.False(open.collapse.collapsed, "a panel without features starts out open")
		c.NotNil(open.FindRefKey(featureEmptyKey), "showing the placeholder")
		c.Equal(field.AsPanel(), wnd.Focus(), "neither takes the focus from the field ahead of them")
		add, ok := open.FindRefKey(featureAddKey).Self.(*unison.Button)
		c.True(ok, "the open panel has an add button")
		if ok {
			add.ClickCallback()
		}
	})
	c.Equal(1, len(empty), "precondition: a feature was added")
	screen.Do(func() {
		c.False(open.collapse.collapsed, "adding the first feature leaves the panel open")
		open.moreEntries("0")[len(open.moreEntries("0"))-1].Act()
	})
	c.Equal(0, len(empty), "precondition: the feature was deleted")
	screen.Do(func() {
		c.False(open.collapse.collapsed, "deleting the last leaves it open")
		c.NotNil(open.FindRefKey(featureEmptyKey))
	})
}

// TestFeaturesPanelCollapseEmpty checks that a collapsed panel with no features says so.
func TestFeaturesPanelCollapseEmpty(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	var features gurps.Features
	c.Equal(0, len(features), "precondition: the list has no features")
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(p.collapse.toggle)
	screen.Do(func() {
		b, ok := p.FindRefKey(featureSummaryKey).Self.(*sentenceButton)
		c.True(ok, "a collapsed panel shows a paragraph")
		if ok {
			c.Equal("No features.", b.plainText())
		}
		c.Nil(p.FindRefKey(featureEmptyKey), "in place of the placeholder")
	})
}

// TestFeaturesPanelTitleBar checks that the title bar covers the strip the title is drawn in, and that the rows start
// where they would under a title that can't be clicked.
func TestFeaturesPanelTitleBar(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() {
		border := p.collapse.border
		c.Equal(border.TitleStrip(p.FrameRect().Size), p.collapse.FrameRect())
		children := p.Children()
		c.True(len(children) > 1 && children[0] == p.collapse.AsPanel(), "the title bar comes first, then the rows")
		if len(children) > 1 {
			plain := &TitledBorder{Title: border.Title, Font: border.Font}
			c.Equal(plain.Insets().Top+2, children[1].FrameRect().Y, "the first row starts below the title")
		}
	})
}

// TestFeaturesPanelBuildingChangesNothing opens every row in turn, holding values a file can hold but the editor
// doesn't offer, and checks that none of it changes the data, since opening an editor must not mark it modified.
func TestFeaturesPanelBuildingChangesNothing(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	weaponSwitch := gurps.NewWeaponBonus(feature.WeaponSwitch)
	weaponSwitch.SetOwner(owner)
	override := gurps.NewSelectorOverride(selector.WeaponDamageType)
	override.Value = "not a damage type"
	override.SetOwner(owner)
	dr := gurps.NewDRBonus()
	dr.Locations = []string{"tail", gurps.AllID, testSkullID}
	dr.Specialization = " "
	dr.SetOwner(owner)
	reduction := gurps.NewCostReduction(gurps.StrengthID)
	reduction.Percentage = fxp.FromInteger(33)
	features := gurps.Features{weaponSwitch, override, dr, reduction}
	c.Equal(wswitch.NotSwitched, weaponSwitch.SwitchType, "precondition: a switch that isn't one of the choices")
	hash := featuresHash(features)
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	for i := range features {
		path := fmt.Sprint(i)
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() { c.NotNil(p.FindRefKey(path + keyFirst)) })
		c.Equal(hash, featuresHash(features), "opening %s changes nothing", path)
	}
	screen.Do(func() {
		c.Equal("by 33%", featuresPopup[fxp.Int](c, p, "3:reduction").Text(), "a reduction that isn't offered is shown")
	})
	c.Equal(wswitch.NotSwitched, weaponSwitch.SwitchType)
	c.Equal("not a damage type", override.Value)
	c.Equal([]string{"tail", gurps.AllID, testSkullID}, dr.Locations)
}

// TestFeaturesPanelAdd checks that the add button adds a feature of the type last chosen at the end of the list, opened
// with the focus in its first field, and that the new type becomes the one added next.
func TestFeaturesPanelAdd(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	defer func(last feature.Type) { lastFeatureTypeUsed = last }(lastFeatureTypeUsed)
	lastFeatureTypeUsed = feature.ReactionBonus
	add := func() {
		screen.Do(func() {
			button, ok := p.FindRefKey(featureAddKey).Self.(*unison.Button)
			c.True(ok, "the section has an add button")
			button.ClickCallback()
		})
	}
	add()
	c.Equal(4, len(features))
	c.Equal(feature.ReactionBonus, features[3].FeatureType(), "a feature is added at the end")
	c.Equal("3", p.open, "and opens")
	screen.Do(func() { c.Equal("3:amount", p.Window().Focus().RefKey) })
	bonus, ok := features[3].(gurps.Bonus)
	c.True(ok)
	c.Equal(fmt.Stringer(owner), bonus.Owner(), "belonging to the item")
	c.Equal("Undo Add Feature", host.mgr.UndoTitle())

	screen.Do(func() {
		featuresPopup[featureTypeEntry](c, p, "3:type").Select(featureTypeEntry{featureType: feature.SkillPointBonus})
	})
	c.Equal(feature.SkillPointBonus, lastFeatureTypeUsed)
	add()
	c.Equal(feature.SkillPointBonus, features[4].FeatureType(), "the type chosen last is the one added")
	screen.Do(host.mgr.Undo)
	c.Equal(4, len(features), "adding is undone in one step")
	c.Equal("3", p.open, "undo opens the row that was open")
}

// TestFeaturesPanelLayout checks that the add button sits under the last row, in line with the more buttons, even with
// no rows, that an open row's controls follow its type as a sentence would, with every line after the first indented
// by the same small amount, that a situation fills its line, and that a DR bonus's locations follow the popup that
// picks them, with its chips after them.
func TestFeaturesPanelLayout(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	skill := gurps.NewSkillBonus()
	skill.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Combat"}
	skill.SetOwner(owner)
	reaction := gurps.NewReactionBonus()
	reaction.SetOwner(owner)
	dr := gurps.NewDRBonus()
	dr.Locations = []string{testSkullID, "tail"}
	dr.SetOwner(owner)
	features := gurps.Features{skill, reaction, dr}
	var empty gurps.Features
	c.NotEqual(criteria.AnyText, skill.TagsCriteria.Compare, "precondition: the skill bonus has a chip")
	c.Equal([]string{testSkullID, "tail"}, dr.Locations,
		"precondition: the DR bonus has a list of locations, one of which the body doesn't have")
	c.Equal(0, len(empty), "precondition: the empty list has no features")
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	emptyPanel, _ := showFeaturesPanel(t, screen, entity, owner, &empty, false)
	rect := func(panel *featuresPanel, key string) geom.Rect {
		target := panel.FindRefKey(key)
		c.NotNil(target, "expected %s", key)
		if target == nil {
			return geom.Rect{}
		}
		return target.RectToRoot(target.ContentRect(true))
	}
	// lines checks that the open row at the path starts with its type, and that each of its lines after the first,
	// whether the sentence wrapped or the line holds a situation, chips or a note, starts the indent in from it.
	lines := func(path string) {
		editor := p.FindRefKey(path + keyFirst)
		c.NotNil(editor, "expected row %s to be open", path)
		if editor == nil {
			return
		}
		typ := rect(p, path+":type")
		c.Equal(editor.RectToRoot(editor.ContentRect(false)).X, typ.X, "the type starts row %s", path)
		indent := typ.X + featureIndent
		children := editor.Children()
		prev := typ.X
		for _, child := range children[0].Children()[1:] {
			r := child.RectToRoot(child.ContentRect(true))
			if r.X < prev {
				c.Equal(indent, r.X, "a wrapped line of row %s starts at the indent", path)
			}
			prev = r.X
		}
		for _, line := range children[1:] {
			for _, child := range line.Children() {
				c.Equal(indent, child.RectToRoot(child.ContentRect(true)).X,
					"a line after the sentence of row %s starts at the indent", path)
			}
		}
	}
	screen.Do(func() {
		add, more := rect(p, featureAddKey), rect(p, "2"+keyMore)
		c.Equal(more.Right(), add.Right(), "the add button lines up with the more buttons")
		c.True(add.Y >= more.Bottom(), "under the last row")
		emptyAdd, emptySection := rect(emptyPanel, featureAddKey), emptyPanel.RectToRoot(emptyPanel.ContentRect(false))
		c.Equal(add.Right()-p.RectToRoot(p.ContentRect(false)).Right(), emptyAdd.Right()-emptySection.Right(),
			"and stays at the right with no rows")
		p.toggle("0")
	})
	screen.Do(func() {
		amount, typ := rect(p, "0:amount"), rect(p, "0:type")
		c.True(amount.X > typ.Right(), "the controls follow the type")
		c.Equal(typ.CenterY(), amount.CenterY(), "on its line")
		c.Equal(typ.X+featureIndent, rect(p, "0:tag"+keyChip).X, "and the chips start at the indent")
		lines("0")
		p.toggle("1")
	})
	screen.Do(func() {
		situation, done := rect(p, "1:situation"), rect(p, "1:done")
		c.Equal(rect(p, "1:type").X+featureIndent, situation.X, "the situation starts at the indent")
		c.True(done.X-situation.Right() <= 2*unison.StdHSpacing, "and fills its line")
		lines("1")
		p.toggle("2")
	})
	screen.Do(func() {
		popup, against := rect(p, "2:locations"), rect(p, "2:add against")
		var locations []geom.Rect
		for _, box := range panelsOfType[*unison.CheckBox](p.FindRefKey("2" + keyFirst)) {
			if strings.HasPrefix(box.RefKey, "2:loc ") {
				locations = append(locations, box.RectToRoot(box.ContentRect(true)))
			}
		}
		c.NotEqual(0, len(locations), "the DR bonus shows its locations")
		if len(locations) == 0 {
			return
		}
		first, last := locations[0], locations[len(locations)-1]
		c.True(first.Y > popup.Bottom() || (first.X > popup.Right() && first.CenterY() == popup.CenterY()),
			"the locations follow the popup that picks them")
		c.True(against.Y > last.Bottom() || (against.X > last.Right() && against.CenterY() == last.CenterY()),
			"with the chips after them")
		lines("2")
	})
}

// TestFeaturesPanelTypeSwitchKeepsSwitchable checks that changing a feature's type replaces it in place, carrying over
// whether it is switchable, and leaves the others alone.
func TestFeaturesPanelTypeSwitchKeepsSwitchable(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	first, last := features[0], features[2]
	c.True(features[1].IsSwitchable(), "precondition: the feature being switched is switchable")
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	defer func(last feature.Type) { lastFeatureTypeUsed = last }(lastFeatureTypeUsed)
	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		featuresPopup[featureTypeEntry](c, p, "1:type").Select(featureTypeEntry{featureType: feature.SelectorOverride})
	})
	c.Equal([]feature.Type{feature.SkillBonus, feature.SelectorOverride, feature.Unknown}, featureTypes(features))
	c.True(features[1].IsSwitchable(), "the replacement is still switchable")
	c.True(features[0] == first && features[2] == last, "the other features are untouched")
	c.Equal("1", p.open, "the row stays open")
	screen.Do(func() { c.NotNil(p.FindRefKey("1:switchable"+keyChip), "and shows the switchable chip it kept") })
}

// TestFeaturesPanelTypeHeadings checks that the type popup files the types it offers under headings, in the order of
// the types within each group, with a separator between groups, and that the headings can't be chosen.
func TestFeaturesPanelTypeHeadings(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		popup := featuresPopup[featureTypeEntry](c, p, "0:type")
		var types []feature.Type
		var headings []string
		group := -1
		for i := range popup.ItemCount() {
			entry, ok := popup.ItemAt(i)
			switch {
			case !ok:
				c.True(i != 0 && len(headings) != 0, "a separator only comes between groups")
			case entry.heading != "":
				c.False(popup.ItemEnabledAt(i), "%s can't be chosen", entry.heading)
				headings = append(headings, entry.heading)
			default:
				c.True(popup.ItemEnabledAt(i))
				c.True(featureTypeGroup(entry.featureType) >= group, "%s is in order", entry.featureType.Key())
				group = featureTypeGroup(entry.featureType)
				c.Equal(featureTypeGroups()[group], headings[len(headings)-1], "%s is under its heading",
					entry.featureType.Key())
				types = append(types, entry.featureType)
			}
		}
		c.Equal(featureTypeGroups(), headings)
		want := slices.Clone(p.featureTypesList())
		slices.SortStableFunc(want, func(a, b feature.Type) int { return featureTypeGroup(a) - featureTypeGroup(b) })
		c.Equal(want, types, "every type offered, in order within its group")
		c.Equal(feature.SkillBonus.String(), popup.Text(), "with the row's type chosen")
	})
}

// TestFeaturesPanelSwitchablePill checks that the switchable pill of an open row adds and removes its feature's flag,
// each as one step to undo, for every type of feature.
func TestFeaturesPanelSwitchablePill(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewEquipment(entity, nil, true)
	features := make(gurps.Features, 0, len(feature.SelectableTypes))
	fp := newFeaturesPanel(entity, owner, &features, false)
	for _, one := range feature.SelectableTypes {
		features = append(features, fp.createFeatureForType(one))
	}
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	click := func(key string) {
		screen.Do(func() {
			buttons := panelsOfType[*unison.Button](p.FindRefKey(key))
			c.NotEqual(0, len(buttons), "expected a button within %s", key)
			if len(buttons) != 0 {
				buttons[len(buttons)-1].ClickCallback()
			}
		})
	}
	for i, one := range features {
		path := fmt.Sprint(i)
		name := one.FeatureType().Key()
		c.False(one.IsSwitchable(), "precondition: %s starts out not switchable", name)
		screen.Do(func() { p.toggle(path) })
		click(path + ":add switchable")
		c.True(features[i].IsSwitchable(), "adding the pill makes %s switchable", name)
		c.Equal("Undo Add Switchable", host.mgr.UndoTitle())
		screen.Do(func() { c.Nil(p.FindRefKey(path+":add switchable"), "%s shows the chip instead", name) })
		click(path + ":switchable" + keyChip)
		c.False(features[i].IsSwitchable(), "removing the chip makes %s not switchable", name)
		c.Equal("Undo Remove Switchable", host.mgr.UndoTitle())
		screen.Do(host.mgr.Undo)
		c.True(features[i].IsSwitchable(), "undoing the removal makes %s switchable again", name)
		screen.Do(host.mgr.Undo)
		c.False(features[i].IsSwitchable(), "undoing the addition makes %s not switchable again", name)
	}
}

// TestFeaturesPanelMoreMenu checks that Duplicate, Move up, Move down and Delete change the list as they say, that the
// open row stays open wherever it goes, and that the moves are offered only where there is room.
func TestFeaturesPanelMoreMenu(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	act := func(path, label string) {
		screen.Do(func() {
			action := prereqMenuAction(p.moreEntries(path), label)
			c.NotNil(action, "%s offers %s", path, label)
			if action != nil {
				action()
			}
		})
	}
	screen.Do(func() {
		c.Nil(prereqMenuAction(p.moreEntries("0"), "Move Up"), "nothing above the top")
		c.Nil(prereqMenuAction(p.moreEntries("2"), "Move Down"), "nothing below the bottom")
		p.toggle("1")
	})
	dr := features[1]
	act("1", "Move Up")
	c.Equal([]feature.Type{feature.DRBonus, feature.SkillBonus, feature.Unknown}, featureTypes(features))
	c.Equal("0", p.open, "the open row moves with its feature")
	screen.Do(func() { c.Equal("0"+keyMore, p.Window().Focus().RefKey, "the moved row's more button takes the focus") })
	act("0", "Move Down")
	c.Equal([]feature.Type{feature.SkillBonus, feature.DRBonus, feature.Unknown}, featureTypes(features))
	c.Equal("1", p.open)
	act("1", "Duplicate")
	c.Equal([]feature.Type{feature.SkillBonus, feature.DRBonus, feature.DRBonus, feature.Unknown}, featureTypes(features))
	c.True(features[1] == dr && features[2] != dr, "the copy follows the original")
	c.Equal(gurps.Hash64(dr), gurps.Hash64(features[2]), "and is the same")
	act("0", "Delete")
	c.Equal([]feature.Type{feature.DRBonus, feature.DRBonus, feature.Unknown}, featureTypes(features))
	c.Equal("0", p.open, "the open row keeps its place as rows above it go")
	act("0", "Delete")
	c.Equal("", p.open, "deleting the open row leaves none open")
	act("1", "Delete")
	act("0", "Delete")
	c.Equal(0, len(features))
	screen.Do(func() {
		c.Equal(featureAddKey, p.Window().Focus().RefKey, "deleting the last gives the add button the focus")
	})
}

// TestFeaturesPanelUndo checks that typing in a field is recorded as one snapshot of the list, that a change to the
// list's shape is another, that undo and redo install copies of their snapshots, and that a field's own undo is never
// recorded.
func TestFeaturesPanelUndo(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		field, ok := p.FindRefKey("0:name").Self.(*StringField)
		c.True(ok, "the open skill bonus has a name field")
		field.RequestFocus()
		field.SetSelection(0, len(field.Text()))
	})
	screen.Type("Pickpocket")
	skill := func() *gurps.SkillBonus {
		one, ok := features[0].(*gurps.SkillBonus)
		c.True(ok)
		return one
	}
	c.Equal("Pickpocket", skill().NameCriteria.Qualifier)
	c.Equal("Undo Name", host.mgr.UndoTitle())
	screen.Do(func() { prereqMenuAction(p.moreEntries("1"), "Delete")() })
	c.Equal(2, len(features))
	c.Equal("Undo Delete Feature", host.mgr.UndoTitle())
	installed := features[0]
	screen.Do(host.mgr.Undo)
	c.Equal([]feature.Type{feature.SkillBonus, feature.DRBonus, feature.Unknown}, featureTypes(features))
	c.True(installed != features[0], "undo installs a copy")
	c.Equal("Pickpocket", skill().NameCriteria.Qualifier)
	screen.Do(host.mgr.Undo)
	c.Equal("Streetwise", skill().NameCriteria.Qualifier, "the typing was one edit")
	c.False(host.mgr.CanUndo())
	screen.Do(func() {
		field, ok := p.FindRefKey("0:name").Self.(*StringField)
		c.True(ok, "the row the typing was done in is open")
		c.Equal("Streetwise", field.Text())
		c.Equal(field.AsPanel(), field.Window().Focus(), "and the field has the focus")
	})
	screen.Do(host.mgr.Redo)
	c.Equal("Pickpocket", skill().NameCriteria.Qualifier)
	screen.Do(host.mgr.Redo)
	c.Equal([]feature.Type{feature.SkillBonus, feature.Unknown}, featureTypes(features))
}

// TestFeaturesPanelDragAndDrop checks that a dragged feature goes before or after a row by which half of the row it is
// over, that a drop is one step to undo and redo, that the open row follows its feature, and that a drop onto or beside
// itself, or of a feature from another panel, changes nothing.
func TestFeaturesPanelDragAndDrop(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	otherFeatures := newTestFeatures(owner)
	var other *featuresPanel
	screen.Do(func() { other = newFeaturesPanel(entity, owner, &otherFeatures, false) })
	point := func(onto string, fraction float32) geom.Point {
		target := p.FindRefKey(onto + keyMore).Parent()
		r := p.RectFromRoot(target.RectToRoot(target.ContentRect(true)))
		return geom.NewPoint(r.X+r.Width/3, r.Y+r.Height*fraction)
	}
	spot := func(data any, onto string, fraction float32) (target *unison.Panel, at int) {
		screen.Do(func() { target, at = p.dropAt(point(onto, fraction), data) })
		return target, at
	}
	drag := func(from, onto string, fraction float32) (accepted bool) {
		screen.Do(func() {
			where := point(onto, fraction)
			data := &rowDrag{panel: p.AsPanel(), path: from}
			p.dragOver(where, data)
			accepted = p.dropTarget != nil
			p.drop(where, data)
		})
		return accepted
	}
	var row2 *unison.Panel
	screen.Do(func() { row2 = p.FindRefKey("2" + keyMore).Parent() })
	target, at := spot(&rowDrag{panel: p.AsPanel(), path: "0"}, "2", 0.2)
	c.Equal(row2, target, "a feature can be dropped on another's row")
	c.Equal(dropBefore, at, "the top half of a row is before it")
	target, at = spot(&rowDrag{panel: p.AsPanel(), path: "0"}, "2", 0.8)
	c.Equal(row2, target)
	c.Equal(dropAfter, at, "the bottom half is after it")
	target, _ = spot(&rowDrag{panel: other.AsPanel(), path: "0"}, "2", 0.2)
	c.Nil(target, "a feature from another panel can't be dropped")
	target, _ = spot("0", "2", 0.2)
	c.Nil(target, "nor can anything else")

	c.False(drag("0", "0", 0.2), "a row can't be dropped onto itself")
	c.True(drag("1", "0", 0.8), "after the row above it")
	c.True(drag("1", "2", 0.2), "and before the row below it")
	c.Equal([]feature.Type{feature.SkillBonus, feature.DRBonus, feature.Unknown}, featureTypes(features),
		"but neither moves it")
	c.False(host.mgr.CanUndo(), "nor records a step")

	screen.Do(func() { p.toggle("1") })
	c.True(drag("0", "2", 0.8))
	c.Equal([]feature.Type{feature.DRBonus, feature.Unknown, feature.SkillBonus}, featureTypes(features))
	c.Equal("0", p.open, "the open row moves with its feature")
	c.Equal("Undo Move Feature", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.Equal([]feature.Type{feature.SkillBonus, feature.DRBonus, feature.Unknown}, featureTypes(features),
		"a drop is one step to undo")
	c.Equal("1", p.open)
	c.False(host.mgr.CanUndo())
	screen.Do(host.mgr.Redo)
	c.Equal([]feature.Type{feature.DRBonus, feature.Unknown, feature.SkillBonus}, featureTypes(features), "and to redo")
	c.Equal("0", p.open)
	c.True(drag("2", "0", 0.2), "before the first row")
	c.Equal([]feature.Type{feature.SkillBonus, feature.DRBonus, feature.Unknown}, featureTypes(features))
	c.Equal("1", p.open)
}

// TestFeaturesPanelDragFromRow checks that a row can be dragged by its sentence without also opening it, and that a
// click on a sentence still opens its row.
func TestFeaturesPanelDragFromRow(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	var sentence, row *unison.Panel
	var below geom.Point
	screen.Do(func() {
		registerWindowDragTypes(p.Window())
		sentence = p.FindRefKey("0" + keySentence)
		row = p.FindRefKey("2" + keyMore).Parent()
		below = geom.NewPoint(40, row.FrameRect().Height*0.8)
	})
	screen.Drag(screen.PanelCenter(sentence), screen.PanelPoint(row, below), 10)
	c.Equal([]feature.Type{feature.DRBonus, feature.Unknown, feature.SkillBonus}, featureTypes(features))
	c.Equal("", p.open, "the drag didn't also click")
	screen.Do(func() { sentence = p.FindRefKey("0" + keySentence) })
	screen.Click(screen.PanelCenter(sentence))
	c.Equal("0", p.open)
}

// TestFeaturesPanelChips checks that an optional criterion is added as a chip and removed again, and that one with no
// comparison of its own, such as a group, shows once added even before it holds anything.
func TestFeaturesPanelChips(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	skill := gurps.NewSkillBonus()
	skill.SetOwner(owner)
	reaction := gurps.NewReactionBonus()
	reaction.SetOwner(owner)
	features := gurps.Features{skill, reaction}
	c.Equal(criteria.AnyText, skill.TagsCriteria.Compare, "precondition: the skill bonus has no tags criterion")
	c.Equal("", reaction.Group, "precondition: the reaction bonus has no group")
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	click := func(key string) {
		screen.Do(func() {
			buttons := panelsOfType[*unison.Button](p.FindRefKey(key))
			c.NotEqual(0, len(buttons), "expected a button within %s", key)
			if len(buttons) != 0 {
				buttons[len(buttons)-1].ClickCallback()
			}
		})
	}
	screen.Do(func() { p.toggle("0") })
	click("0:add tag")
	c.Equal(criteria.IsText, skill.TagsCriteria.Compare, "adding the chip adds the criterion")
	screen.Do(func() { c.Equal("0:tagcmp", p.Window().Focus().RefKey, "and focuses its comparison") })
	click("0:tag" + keyChip)
	c.Equal(criteria.AnyText, skill.TagsCriteria.Compare, "removing the chip removes it")

	screen.Do(func() { p.toggle("1") })
	click("1:add group")
	screen.Do(func() { c.NotNil(p.FindRefKey("1:group"+keyChip), "an added group shows while it is empty") })
	screen.Type("Social")
	c.Equal("Social", reaction.Group)
	click("1:group" + keyChip)
	c.Equal("", reaction.Group)
	screen.Do(func() { c.Nil(p.FindRefKey("1:group" + keyChip)) })
}

// TestFeaturesPanelEmpty checks that a list with no features shows a placeholder before the add button, that clicking
// it adds a feature of the type last chosen and opens it, as the add button does, and that undo brings it back.
func TestFeaturesPanelEmpty(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	var features gurps.Features
	c.Equal(0, len(features), "precondition: the list has no features")
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	defer func(last feature.Type) { lastFeatureTypeUsed = last }(lastFeatureTypeUsed)
	lastFeatureTypeUsed = feature.ReactionBonus
	var empty *unison.Button
	screen.Do(func() {
		var ok bool
		empty, ok = p.FindRefKey(featureEmptyKey).Self.(*unison.Button)
		c.True(ok, "an empty list shows a placeholder")
		if !ok {
			return
		}
		c.Equal("No features. Add one to get started.", empty.Text.String())
		add := p.FindRefKey(featureAddKey)
		c.NotNil(add)
		r, a := empty.RectToRoot(empty.ContentRect(true)), add.RectToRoot(add.ContentRect(true))
		c.True(a.X >= r.Right(), "the add button follows it")
		c.Equal(r.CenterY(), a.CenterY(), "on its line")
	})
	if empty == nil {
		return
	}
	screen.Do(empty.ClickCallback)
	c.Equal([]feature.Type{feature.ReactionBonus}, featureTypes(features), "clicking it adds a feature of the last type")
	c.Equal("0", p.open, "and opens it")
	screen.Do(func() {
		c.Equal("0:amount", p.Window().Focus().RefKey, "with the focus in its first field")
		c.Nil(p.FindRefKey(featureEmptyKey), "and the placeholder goes")
	})
	c.Equal("Undo Add Feature", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.Equal(0, len(features))
	screen.Do(func() { c.NotNil(p.FindRefKey(featureEmptyKey), "undo brings the placeholder back") })
}

// TestFeaturesPanelSituation checks that a new reaction bonus or conditional modifier starts with no situation, so that
// its field shows the hint, and that the field grows to fit its text rather than scrolling within itself.
func TestFeaturesPanelSituation(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	var features gurps.Features
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	reaction, ok := p.createFeatureForType(feature.ReactionBonus).(*gurps.ReactionBonus)
	c.True(ok)
	conditional, ok := p.createFeatureForType(feature.ConditionalModifier).(*gurps.ConditionalModifierBonus)
	c.True(ok)
	if reaction == nil || conditional == nil {
		return
	}
	c.Equal("", reaction.Situation, "a new reaction bonus has no situation")
	c.Equal("", conditional.Situation, "nor does a new conditional modifier")
	features = gurps.Features{reaction, conditional}
	for i, hint := range []string{"from/to target", "Triggering Condition"} {
		path := strconv.Itoa(i)
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() {
			field, isField := p.FindRefKey(path + ":situation").Self.(*StringField)
			c.True(isField, "row %s has a situation field", path)
			if !isField {
				return
			}
			c.Equal("", field.Text())
			c.Equal(hint, field.Watermark, "row %s shows its hint", path)
			c.False(field.AutoScroll, "row %s's field doesn't scroll within itself", path)
		})
	}
}

// TestFeaturesPanelOptionalSpecialization checks that a skill bonus selected by name and a skill point bonus offer an
// optional specialization criterion, which a skill bonus selected another way does not.
func TestFeaturesPanelOptionalSpecialization(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	byName := gurps.NewSkillBonus()
	byName.SetOwner(owner)
	points := gurps.NewSkillPointBonus()
	points.SetOwner(owner)
	thisWeapon := gurps.NewSkillBonus()
	thisWeapon.SelectionType = skillsel.ThisWeapon
	thisWeapon.SetOwner(owner)
	features := gurps.Features{byName, points, thisWeapon}
	c.Equal(skillsel.Name, byName.SelectionType, "precondition: the first skill bonus is selected by name")
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	for i, optional := range []*criteria.Text{
		&byName.OptionalSpecializationCriteria, &points.OptionalSpecializationCriteria,
	} {
		path := strconv.Itoa(i)
		c.Equal(criteria.AnyText, optional.Compare, "precondition: row %s has no optional specialization", path)
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() {
			add, ok := p.FindRefKey(path + ":add optspecialization").Self.(*unison.Button)
			c.True(ok, "row %s offers an optional specialization", path)
			if ok {
				add.ClickCallback()
			}
		})
		c.Equal(criteria.IsText, optional.Compare, "adding it to row %s adds the criterion", path)
		screen.Do(func() { c.NotNil(p.FindRefKey(path + ":optspecialization" + keyChip)) })
	}
	screen.Do(func() { p.toggle("2") })
	screen.Do(func() {
		c.Nil(p.FindRefKey("2:add optspecialization"), "a skill bonus selected another way doesn't offer it")
	})
}

// TestFeaturesPanelRelativeSkillLevel checks that a weapon bonus selected by its required skill leaves its relative
// skill level as an optional chip, that a new one has none, that one at least 0, which a file can hold, shows as the
// filter it is, and that removing the chip removes the criterion.
func TestFeaturesPanelRelativeSkillLevel(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	fresh := gurps.NewWeaponBonus(feature.WeaponBonus)
	fresh.SetOwner(owner)
	zero := gurps.NewWeaponBonus(feature.WeaponBonus)
	zero.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber}
	zero.SetOwner(owner)
	features := gurps.Features{fresh, zero}
	c.Equal(wsel.WithRequiredSkill, fresh.SelectionType, "precondition: a new weapon bonus is selected by skill")
	c.Equal(criteria.AnyNumber, fresh.RelativeLevelCriteria.Compare, "a new weapon bonus has no level criterion")
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		c.Nil(p.FindRefKey("0:level"+keyChip), "a new weapon bonus shows no level chip")
		add, ok := p.FindRefKey("0:add level").Self.(*unison.Button)
		c.True(ok, "but offers one")
		if ok {
			add.ClickCallback()
		}
	})
	c.Equal(criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}, fresh.RelativeLevelCriteria)
	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		c.NotNil(p.FindRefKey("1:level"+keyChip), "a level of at least 0 is a filter, so it shows its chip")
	})
	screen.Do(func() {
		buttons := panelsOfType[*unison.Button](p.FindRefKey("1:level" + keyChip))
		c.NotEqual(0, len(buttons))
		if len(buttons) != 0 {
			buttons[len(buttons)-1].ClickCallback()
		}
	})
	c.Equal(criteria.AnyNumber, zero.RelativeLevelCriteria.Compare, "removing the chip removes the criterion")
}

// TestFeaturesPanelFocusAfterRebuild checks that a choice from a popup leaves the focus on the popup, and that when the
// control that had the focus goes, the open row takes it rather than the first row of the panel.
func TestFeaturesPanelFocusAfterRebuild(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := newTestFeatures(owner)
	skill, ok := features[0].(*gurps.SkillBonus)
	c.True(ok, "precondition: the first feature is a skill bonus")
	if !ok {
		return
	}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		popup := featuresPopup[skillsel.Type](c, p, "0:selection")
		popup.RequestFocus()
		popup.Select(skillsel.WeaponsWithName)
	})
	c.Equal(skillsel.WeaponsWithName, skill.SelectionType, "precondition: the popup made its choice")
	screen.Do(func() { c.Equal("0:selection", p.Window().Focus().RefKey, "the popup keeps the focus") })

	features = append(gurps.Features{gurps.NewSkillBonus()}, features...)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		field, isField := p.FindRefKey("1:name").Self.(*StringField)
		c.True(isField, "precondition: the open row has a name field")
		if isField {
			field.RequestFocus()
		}
		skill.SelectionType = skillsel.ThisWeapon
		p.rebuild("")
	})
	screen.Do(func() {
		c.Nil(p.FindRefKey("1:name"), "precondition: the field that had the focus is gone")
		c.Equal("1:amount", p.Window().Focus().RefKey, "the open row takes the focus")
	})
}

// TestFeaturesPanelDRLocations checks that a DR bonus's checkbox grid adds and removes locations, refusing to remove
// the last, that the location popup offers "to this armor" only to an equipment modifier, and that choosing a list of
// locations from "all" starts it with the torso.
func TestFeaturesPanelDRLocations(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	traitDR := gurps.NewDRBonus()
	traitDR.SetOwner(trait)
	traitFeatures := gurps.Features{traitDR}
	modifier := gurps.NewEquipmentModifier(entity, nil, false)
	modifierDR := gurps.NewDRBonus()
	modifierDR.SetOwner(modifier)
	modifierFeatures := gurps.Features{modifierDR}
	c.Equal([]string{gurps.TorsoID}, traitDR.Locations, "precondition: the trait's bonus covers the torso")
	c.Equal([]string{gurps.TorsoID}, modifierDR.Locations, "precondition: the modifier's bonus covers the torso")
	traitPanel, _ := showFeaturesPanel(t, screen, entity, trait, &traitFeatures, false)
	modifierPanel, _ := showFeaturesPanel(t, screen, entity, modifier, &modifierFeatures, true)
	screen.Do(func() {
		traitPanel.toggle("0")
		modifierPanel.toggle("0")
	})
	screen.Do(func() {
		c.Equal(-1, featuresPopup[string](c, traitPanel, "0:locations").IndexOfItem("to this armor"),
			"a trait's bonus doesn't offer to this armor")
		c.Equal(0, featuresPopup[string](c, modifierPanel, "0:locations").IndexOfItem("to this armor"),
			"an equipment modifier's bonus does")
		clickFeatureCheckBox(featuresCheckBox(c, traitPanel, "0:loc "+testSkullID), true)
	})
	c.Equal([]string{testSkullID, gurps.TorsoID}, traitDR.Locations, "ticking a box adds its location")
	screen.Do(func() { clickFeatureCheckBox(featuresCheckBox(c, traitPanel, "0:loc "+gurps.TorsoID), false) })
	c.Equal([]string{testSkullID}, traitDR.Locations, "unticking one removes it")
	screen.Do(func() { clickFeatureCheckBox(featuresCheckBox(c, traitPanel, "0:loc "+testSkullID), false) })
	c.Equal([]string{testSkullID}, traitDR.Locations, "the last one stays")
	screen.Do(func() {
		c.Equal(uncheck.On, featuresCheckBox(c, traitPanel, "0:loc "+testSkullID).State, "and shows as ticked")
		featuresPopup[string](c, modifierPanel, "0:locations").Select("to this armor")
	})
	c.Equal(0, len(modifierDR.Locations), "to this armor has no locations")
	screen.Do(func() {
		c.Nil(modifierPanel.FindRefKey("0:loc "+gurps.TorsoID), "and no grid")
		featuresPopup[string](c, modifierPanel, "0:locations").Select("to all locations")
	})
	c.Equal([]string{gurps.AllID}, modifierDR.Locations)
	screen.Do(func() { featuresPopup[string](c, modifierPanel, "0:locations").Select("to these locations:") })
	c.Equal([]string{gurps.TorsoID}, modifierDR.Locations)
}

// TestFeaturesPanelContainedWeightReductionOnlyForContainers checks that only a container offers the contained weight
// reduction, while a feature of that type that something else holds is still shown as such.
func TestFeaturesPanelContainedWeightReductionOnlyForContainers(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	container := gurps.NewEquipment(entity, nil, true)
	containerFeatures := gurps.Features{gurps.NewContainedWeightReduction()}
	item := gurps.NewEquipment(entity, nil, false)
	itemFeatures := gurps.Features{gurps.NewContainedWeightReduction()}
	c.True(container.Container(), "precondition: the container is one")
	c.False(item.Container(), "precondition: the item isn't one")
	containerPanel, _ := showFeaturesPanel(t, screen, entity, container, &containerFeatures, false)
	itemPanel, _ := showFeaturesPanel(t, screen, entity, item, &itemFeatures, false)
	c.True(slices.Contains(containerPanel.featureTypesList(), feature.ContainedWeightReduction))
	c.False(slices.Contains(itemPanel.featureTypesList(), feature.ContainedWeightReduction))
	screen.Do(func() {
		containerPanel.toggle("0")
		itemPanel.toggle("0")
	})
	screen.Do(func() {
		c.Equal(feature.ContainedWeightReduction.String(), featuresPopup[featureTypeEntry](c, containerPanel, "0:type").Text())
		c.Equal(feature.ContainedWeightReduction.String(), featuresPopup[featureTypeEntry](c, itemPanel, "0:type").Text())
	})
}

// TestFeaturesPanelSelectorFieldChange checks that choosing another field for a selector override resets its value to
// the field's first, and clears the usage criterion when the field belongs to traits, which have no usage.
func TestFeaturesPanelSelectorFieldChange(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	override := gurps.NewSelectorOverride(selector.WeaponDamageType)
	override.UsageCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Thrown"}
	override.SetOwner(owner)
	features := gurps.Features{override}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	defer func(last selector.Field) { lastSelectorFieldUsed = last }(lastSelectorFieldUsed)
	var traitField selector.Field
	var found bool
	for _, one := range selector.Fields {
		if gurps.SelectorFieldDescriptorFor(one).Scope == gurps.SelectorScopeTrait {
			traitField, found = one, true
			break
		}
	}
	c.True(found, "precondition: some field belongs to traits")
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { featuresPopup[selector.Field](c, p, "0:field").Select(traitField) })
	d := gurps.SelectorFieldDescriptorFor(traitField)
	want := ""
	if len(d.SuggestedStates) != 0 {
		want = d.SuggestedStates[0]
	}
	c.Equal(want, override.Value, "the value starts over")
	c.Equal(criteria.Text{Compare: criteria.AnyText}, override.UsageCriteria, "the usage criterion is cleared")
	c.Equal(traitField, lastSelectorFieldUsed)
	screen.Do(func() { c.Nil(p.FindRefKey("0:add usage"), "and isn't offered") })
}

// TestFeaturesPanelWeaponDamageBonus checks that the amount of a weapon damage bonus takes a dice specification as well
// as a number, that "as a %" is suspended while it holds dice and comes back once they go, unless it was turned off,
// that text in any other form leaves the bonus alone, that only the damage bonus takes dice, and that the ST bonuses
// offer no per-die choice.
func TestFeaturesPanelWeaponDamageBonus(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	damage := gurps.NewWeaponBonus(feature.WeaponBonus)
	damage.SelectionType = wsel.ThisWeapon
	damage.Amount = fxp.Ten
	damage.Percent = true
	damage.SetOwner(owner)
	accuracy := gurps.NewWeaponBonus(feature.WeaponAccBonus)
	accuracy.SetOwner(owner)
	minST := gurps.NewWeaponBonus(feature.WeaponMinSTBonus)
	minST.SetOwner(owner)
	features := gurps.Features{damage, accuracy, minST}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	var field *StringField
	screen.Do(func() {
		var ok bool
		field, ok = p.FindRefKey("0:amount").Self.(*StringField)
		c.True(ok, "the damage bonus amount is a text field")
		c.Equal("+10", field.Text())
		c.True(featuresCheckBox(c, p, "0:percent").Enabled())
		c.NotNil(p.FindRefKey("0:perdie"))
	})
	if field == nil {
		return
	}
	screen.Do(func() { field.SetText("2d+1x3") })
	expected, ok := gurps.ParseBonusDice("2d+1x3")
	c.True(ok)
	c.Equal(expected, damage.Dice, "a dice specification is stored as the bonus's dice")
	c.Equal(fxp.Int(0), damage.Amount, "and clears the flat amount")
	c.False(damage.Percent, "entering dice turns the percentage off")
	screen.Do(func() {
		percent := featuresCheckBox(c, p, "0:percent")
		c.False(percent.Enabled(), "and disables it")
		c.Equal(uncheck.Off, percent.State)
		field.SetText("2d+1x3 cr")
		c.False(field.ValidateCallback(), "text that is not a bonus is flagged")
	})
	c.Equal(expected, damage.Dice, "and leaves the dice alone")
	screen.Do(func() { field.SetText("+10") })
	c.True(damage.Percent, "the percentage comes back once the dice are gone")
	screen.Do(func() {
		c.True(featuresCheckBox(c, p, "0:percent").Enabled())
		clickFeatureCheckBox(featuresCheckBox(c, p, "0:percent"), false)
		field.SetText("+1d")
		field.SetText("+3")
	})
	c.False(damage.Percent, "a percentage the user turned off is not turned back on")

	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		_, isDecimal := p.FindRefKey("1:amount").Self.(*DecimalField)
		c.True(isDecimal, "another weapon bonus keeps its numeric amount")
		c.NotNil(p.FindRefKey("1:perdie"))
		p.toggle("2")
	})
	screen.Do(func() { c.Nil(p.FindRefKey("2:perdie"), "a minimum ST bonus has no per-die choice") })
}

// TestFeaturesPanelPercentSuspensionSurvivesRebuild checks that "as a %", suspended while a weapon damage bonus holds
// dice, comes back once they go even after the row has been closed and reopened, or another control has rebuilt the
// panel.
func TestFeaturesPanelPercentSuspensionSurvivesRebuild(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	damage := gurps.NewWeaponBonus(feature.WeaponBonus)
	damage.SelectionType = wsel.ThisWeapon
	damage.Amount = fxp.Ten
	damage.Percent = true
	damage.SetOwner(owner)
	features := gurps.Features{damage}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	setAmount := func(text string) {
		screen.Do(func() {
			field, ok := p.FindRefKey("0:amount").Self.(*StringField)
			c.True(ok, "the damage bonus amount is a text field")
			if ok {
				field.SetText(text)
			}
		})
	}
	screen.Do(func() { p.toggle("0") })
	setAmount("+1d")
	c.False(damage.Percent, "precondition: dice suspend the percentage")
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { p.toggle("0") })
	setAmount("+10")
	c.True(damage.Percent, "the percentage comes back after the row is closed and reopened")

	setAmount("+1d")
	c.False(damage.Percent, "precondition: dice suspend the percentage again")
	screen.Do(func() {
		featuresPopup[wsel.Type](c, p, "0:selection").Select(wsel.WithRequiredSkill)
	})
	c.Equal(wsel.WithRequiredSkill, damage.SelectionType, "precondition: the popup rebuilt the panel")
	setAmount("+10")
	c.True(damage.Percent, "the percentage comes back after another control rebuilt the panel")
}

// TestFeaturesPanelPercentSuspensionSurvivesUndo checks that undo and redo suspend "as a %" for the copies they install
// as it was suspended in the snapshots they return to, and only for those.
func TestFeaturesPanelPercentSuspensionSurvivesUndo(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	damage := gurps.NewWeaponBonus(feature.WeaponBonus)
	damage.SelectionType = wsel.ThisWeapon
	damage.Amount = fxp.Ten
	damage.Percent = true
	damage.SetOwner(owner)
	features := gurps.Features{damage}
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	setAmount := func(text string) {
		screen.Do(func() {
			field, ok := p.FindRefKey("0:amount").Self.(*StringField)
			c.True(ok, "the damage bonus amount is a text field")
			if ok {
				field.SetText(text)
			}
		})
	}
	installed := func() *gurps.WeaponBonus {
		one, ok := features[0].(*gurps.WeaponBonus)
		c.True(ok)
		return one
	}
	screen.Do(func() { p.toggle("0") })
	setAmount("1d")
	c.False(damage.Percent, "precondition: dice suspend the percentage")
	screen.Do(func() { clickFeatureCheckBox(featuresCheckBox(c, p, "0:perlevel"), true) })
	c.True(damage.PerLevel, "precondition: the bonus is per level")
	screen.Do(host.mgr.Undo)
	c.True(installed() != damage, "precondition: undo installs a copy")
	c.False(installed().PerLevel, "precondition: undo takes back per level")
	setAmount("+10")
	c.True(installed().Percent, "the percentage comes back for the copy undo installed")

	screen.Do(host.mgr.Undo)
	c.False(installed().Dice.IsZero(), "precondition: undo returns to the dice")
	c.Equal(1, len(p.percentSuspended), "with the percentage suspended")
	screen.Do(host.mgr.Undo)
	c.True(installed().Percent, "precondition: undo returns to before the dice")
	c.Equal(0, len(p.percentSuspended), "where it was not suspended")
	screen.Do(host.mgr.Redo)
	c.False(installed().Dice.IsZero(), "precondition: redo returns to the dice")
	c.Equal(1, len(p.percentSuspended), "with the percentage suspended")
	setAmount("+10")
	c.True(installed().Percent, "the percentage comes back for the copy redo installed")
}

// TestFeaturesPanelCheckBoxUndo checks that each click on a checkbox is a step of its own to undo, rather than being
// run together with the clicks before it as typing is.
func TestFeaturesPanelCheckBoxUndo(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	skill := gurps.NewSkillBonus()
	skill.SetOwner(owner)
	features := gurps.Features{skill}
	c.False(skill.PerLevel, "precondition: the skill bonus is not per level")
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { clickFeatureCheckBox(featuresCheckBox(c, p, "0:perlevel"), true) })
	screen.Do(func() { clickFeatureCheckBox(featuresCheckBox(c, p, "0:perlevel"), false) })
	perLevel := func() bool {
		one, ok := features[0].(*gurps.SkillBonus)
		c.True(ok)
		return one.PerLevel
	}
	c.False(perLevel())
	screen.Do(host.mgr.Undo)
	c.True(perLevel(), "undo takes back the untick alone")
	c.True(host.mgr.CanUndo(), "leaving the tick")
	screen.Do(host.mgr.Undo)
	c.False(perLevel())
	c.False(host.mgr.CanUndo())
}

// TestFeaturesPanelCreatesEverySelectableType checks that the editor creates a feature for every type the user can
// pick, carrying that type and, for bonuses, its owner. The weapon bonuses share one constructor keyed by type, so a
// weapon type missing from feature.Type.IsWeaponBonus would fall through to the "unknown feature type" arm and yield
// nil here.
func TestFeaturesPanelCreatesEverySelectableType(t *testing.T) {
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	var features gurps.Features
	panel := newFeaturesPanel(entity, trait, &features, false)
	for _, one := range feature.SelectableTypes {
		t.Run(one.Key(), func(t *testing.T) {
			c := check.New(t)
			f := panel.createFeatureForType(one)
			c.NotNil(f)
			if f == nil {
				return
			}
			c.Equal(one, f.FeatureType())
			if bonus, ok := f.(gurps.Bonus); ok {
				c.Equal(trait, bonus.Owner(), "owner")
			}
			_, isWeaponBonus := f.(*gurps.WeaponBonus)
			c.Equal(one.IsWeaponBonus(), isWeaponBonus, "weapon bonus")
		})
	}
}
