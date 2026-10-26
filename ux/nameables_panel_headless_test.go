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
	"maps"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// showNameablesPanel shows a nameablesPanel in a window, within a host, so that its rebuilds run and its edits can be
// undone. The row at the path open starts open, so that the window is sized to fit it. texts holds the item's text,
// which a test may change.
//
//nolint:gocritic // The panel replaces the maps themselves
func showNameablesPanel(t *testing.T, screen *unison.HeadlessScreen, defs *map[string]nameable.Detail, texts *[]string,
	links *map[tid.TID]map[string]string, modifiers []nameableModifier, open string,
) (*nameablesPanel, *sentenceUndoHost) {
	var p *nameablesPanel
	host := &sentenceUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
	screen.Do(func() {
		host.Self = host
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		var mods func() []nameableModifier
		if links != nil {
			mods = func() []nameableModifier { return modifiers }
		}
		p = newNameablesPanel(defs, func() map[string]int { return nameableUses(*texts, nil, nil, nil) }, links, mods)
		// Expanded ahead of showing, so that the window is sized for the rows.
		if p.collapse.collapsed {
			p.collapse.toggle()
		}
		if open != "" {
			p.open = open
			p.RemoveAllChildren()
			p.build()
		}
		host.AddChild(p)
	})
	wnd := showInTestWindow(t, screen, 700, host)
	screen.Do(func() { RegisterWindowDragTypes(wnd) })
	return p, host
}

// clickNameablesButton calls the click callback of the panel's button with the reference key.
func clickNameablesButton(c check.Checker, p *nameablesPanel, key string) {
	b, ok := p.FindRefKey(key).Self.(*unison.Button)
	c.True(ok, "expected a button at %s", key)
	if ok {
		b.ClickCallback()
	}
}

func TestNameablesPanelReordersOptionsByDragging(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{"style": {Label: "Style", Options: []string{"Judo", "Karate", "Boxing"}}}
	texts := []string{"@style@"}
	p, host := showNameablesPanel(t, screen, &defs, &texts, nil, nil, pathOf("style"))
	var from, to geom.Point
	screen.Do(func() {
		var chips []*unison.Panel
		p.HasInSelfOrDescendants(func(one *unison.Panel) bool {
			if _, ok := one.ClientData()[nameableOptionKey]; ok {
				chips = append(chips, one)
			}
			return false
		})
		c.Equal(3, len(chips), "each option is a chip")
		if len(chips) != 3 {
			return
		}
		from = screen.PanelPoint(chips[2], geom.NewPoint(12, chips[2].FrameRect().Height/2))
		to = screen.PanelPoint(chips[0], geom.NewPoint(4, chips[0].FrameRect().Height/2))
	})
	screen.Drag(from, to, 10)
	c.Equal([]string{"Boxing", "Judo", "Karate"}, defs["style"].Options)
	screen.Do(host.mgr.Undo)
	c.Equal([]string{"Judo", "Karate", "Boxing"}, defs["style"].Options, "a drop is one step to undo")
}

// TestNameablesStateHashTellsReorderedOptionsApart checks that the hash by which the panel tells an edit from no change
// sees options reordered even when their texts run together the same way in either order, as "x" and "x x" do.
func TestNameablesStateHashTellsReorderedOptionsApart(t *testing.T) {
	c := check.New(t)
	before := &nameablesState{defs: map[string]nameable.Detail{"style": {Options: []string{"x", "x x"}}}}
	after := &nameablesState{defs: map[string]nameable.Detail{"style": {Options: []string{"x x", "x"}}}}
	c.NotEqual(before.hash(), after.hash(), "reordered options are a change")
	c.Equal(before.hash(), before.clone().hash(), "a copy hashes the same")
	linked := before.clone()
	linked.links = map[tid.TID]map[string]string{"m": {"art": "style"}}
	c.NotEqual(before.hash(), linked.hash(), "a link is a change")
}

func TestNameablesPanelPills(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{"style": {Label: "Style", FreeForm: true}}
	texts := []string{"@style@"}
	p, host := showNameablesPanel(t, screen, &defs, &texts, nil, nil, pathOf("style"))
	screen.Do(func() { clickNameablesButton(c, p, "n:style:add notblank") })
	c.True(defs["style"].NotBlank, "the pill disallows a blank answer")
	c.Equal("Undo Disallow a Blank Answer", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.False(defs["style"].NotBlank)

	// A new option is typed into a chip of its own, and each Return adds one.
	screen.Do(func() { clickNameablesButton(c, p, "n:style:add option") })
	screen.Do(func() {
		field, ok := p.FindRefKey("n:style:newoption").Self.(*StringField)
		c.True(ok, "the new option's field shows")
		if ok {
			c.True(field.FrameRect().Width >= unison.NewText(field.Watermark, &unison.TextDecoration{Font: field.Font}).Width(),
				"and is wide enough to show its watermark")
			field.SetText("Judo")
			field.KeyDownCallback(unison.KeyReturn, 0, false)
		}
	})
	c.Equal([]string{"Judo"}, defs["style"].Options)
	screen.Do(func() { c.NotNil(p.FindRefKey("n:style:newoption"), "the field stays for the next one") })
}

func TestNameablesPanelDefinesWhatTheTextRefersTo(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var defs map[string]nameable.Detail
	texts := []string{"Speaks @lang@ with @Accent: Posh, Rural, etc.@"}
	p, host := showNameablesPanel(t, screen, &defs, &texts, nil, nil, "")
	screen.Do(func() {
		c.Nil(p.FindRefKey("u:Accent: Posh, Rural, etc.:define"), "a marker that defines itself has nothing to define")
		clickNameablesButton(c, p, "u:lang:define")
	})
	c.Equal(map[string]nameable.Detail{"lang": {Label: "lang", FreeForm: true}}, defs)
	c.Equal(pathOf("lang"), p.open, "the new definition opens for editing")
	screen.Do(host.mgr.Undo)
	c.Equal(0, len(defs), "defining is one step to undo")

	// A code typed into the item's text shows up without anything else happening in the panel.
	texts = []string{texts[0], "and @tongue@"}
	screen.Do(func() { DeepSync(host) })
	screen.Do(func() { c.NotNil(p.FindRefKey("u:tongue:define"), "the new code is listed as not defined") })
}

func TestNameablesPanelLinksModifierCodes(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{"style": {Label: "Style", FreeForm: true}}
	texts := []string{"@style@"}
	var links map[tid.TID]map[string]string
	id := tid.MustNewTID('m')
	modifiers := []nameableModifier{{
		id:   id,
		name: "Practiced",
		defs: map[string]nameable.Detail{"art": {Label: "Art"}, "style": {Label: "Style"}},
		uses: map[string]int{"art": 1, "style": 1},
	}}
	p, host := showNameablesPanel(t, screen, &defs, &texts, &links, modifiers, "")
	itemSentence := func() string {
		var text string
		screen.Do(func() {
			if sentence, ok := p.FindRefKey(pathOf("style") + keySentence).Self.(*sentenceButton); ok {
				text = sentence.plainText()
			}
		})
		return text
	}
	c.Equal("@style@ asks for Style as any text; used once, and once in a modifier", itemSentence(),
		"a modifier's use of a code it shares with the item counts for the item's")
	sharedSentence := func() string {
		var text string
		screen.Do(func() {
			if sentence, ok := p.FindRefKey("m:" + string(id) + ":style" + keySentence).Self.(*sentenceButton); ok {
				c.Nil(sentence.onClick, "a code the item also has has nothing to link")
				text = sentence.plainText()
			}
		})
		return text
	}
	c.Equal("@style@ on Practiced shares the item's answer; used once", sharedSentence())
	modifiers[0].disabled = true
	screen.Do(func() { DeepSync(host) })
	c.Equal("@style@ on Practiced (disabled) shares the item's answer; used once", sharedSentence(),
		"a change to the modifier in the editor's modifier list shows")
	art := "m:" + string(id) + ":art"
	screen.Do(func() { p.toggle(art) })
	screen.Do(func() {
		popup, ok := p.FindRefKey(art + ":link").Self.(*unison.PopupMenu[string])
		c.True(ok, "a code the item doesn't have can be linked")
		if ok {
			popup.Select("style")
		}
	})
	c.Equal(map[tid.TID]map[string]string{id: {"art": "style"}}, links)
	c.Equal("@style@ asks for Style as any text; used once, and 2 times in modifiers", itemSentence(),
		"so does a use of a code linked to it")

	// Renaming the item's code carries the link along with it, and its row stays open under the new code.
	screen.Do(func() { p.toggle(pathOf("style")) })
	screen.Do(func() {
		field, ok := p.FindRefKey(nameablesCodeKey).Self.(*StringField)
		c.True(ok)
		if ok {
			field.SetText("schoo")
			field.SetText("school")
		}
	})
	c.Equal(map[tid.TID]map[string]string{id: {"art": "school"}}, links)
	c.Equal(pathOf("school"), p.open)
	screen.Do(func() { clickNameablesButton(c, p, pathOf("school")+keyDone) })
	c.Equal("", p.open, "Done closes the renamed row")
	screen.Do(host.mgr.Undo)
	c.Equal(map[tid.TID]map[string]string{id: {"art": "style"}}, links, "a rename is one step to undo")
}

// TestNameablesPanelShowsLinkOverSameNamedCode covers a modifier's code that is linked to one of the item's codes while
// the item has a code of the same name: the model applies the link first, so the row must say so and stay editable,
// and the use must count for the linked code.
func TestNameablesPanelShowsLinkOverSameNamedCode(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{
		"style": {Label: "Style", FreeForm: true},
		"art":   {Label: "Art", FreeForm: true},
	}
	texts := []string{"@style@ @art@"}
	id := tid.MustNewTID('m')
	links := map[tid.TID]map[string]string{id: {"art": "style"}}
	modifiers := []nameableModifier{{
		id:   id,
		name: "Practiced",
		defs: map[string]nameable.Detail{"art": {Label: "Art"}},
		uses: map[string]int{"art": 1},
	}}
	p, _ := showNameablesPanel(t, screen, &defs, &texts, &links, modifiers, "")
	sentenceAt := func(path string) (text string, editable bool) {
		screen.Do(func() {
			if panel := p.FindRefKey(path + keySentence); panel != nil {
				if sentence, ok := panel.Self.(*sentenceButton); ok {
					text = sentence.plainText()
					editable = sentence.onClick != nil
				}
			}
		})
		return text, editable
	}
	art := "m:" + string(id) + ":art"
	text, editable := sentenceAt(art)
	c.Equal("@art@ on Practiced takes the answer to @style@; used once", text,
		"the link is what the modifier's code resolves through, not the item's same-named code")
	c.True(editable, "the row stays editable so the link can be changed or removed")
	text, _ = sentenceAt(pathOf("style"))
	c.Equal("@style@ asks for Style as any text; used once, and once in a modifier", text,
		"the use counts for the linked code")
	text, _ = sentenceAt(pathOf("art"))
	c.Equal("@art@ asks for Art as any text; used once", text, "and not for the item's same-named code")

	// Removing the link lets the modifier's code share the item's answer.
	screen.Do(func() { p.toggle(art) })
	screen.Do(func() {
		popup, ok := p.FindRefKey(art + ":link").Self.(*unison.PopupMenu[string])
		c.True(ok, "the linked code can be edited")
		if ok {
			c.Equal(-1, popup.IndexOfItem("art"), "a code is not offered a link to itself")
			popup.Select("")
		}
	})
	c.Equal(0, len(links))
	screen.Do(func() { clickNameablesButton(c, p, art+keyDone) })
	text, editable = sentenceAt(art)
	c.Equal("@art@ on Practiced shares the item's answer; used once", text)
	c.False(editable, "once it shares the item's answer, there is nothing to link")
	text, _ = sentenceAt(pathOf("art"))
	c.Equal("@art@ asks for Art as any text; used once, and once in a modifier", text)
}

// TestNameablesPanelRenameThroughModifierCode covers renaming the item's code by typing through the name of a
// modifier's code linked to it: the link follows each keystroke, so it must survive the moment the two names match,
// when it points at itself. Applying the editor's data is what drops such a link (see TestApplyDropsSelfLinks).
func TestNameablesPanelRenameThroughModifierCode(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{"style": {Label: "Style", FreeForm: true}}
	texts := []string{"@style@"}
	id := tid.MustNewTID('m')
	links := map[tid.TID]map[string]string{id: {"art": "style"}}
	modifiers := []nameableModifier{{id: id, name: "Practiced", uses: map[string]int{"art": 1}}}
	p, _ := showNameablesPanel(t, screen, &defs, &texts, &links, modifiers, pathOf("style"))
	screen.Do(func() {
		field, ok := p.FindRefKey(nameablesCodeKey).Self.(*StringField)
		c.True(ok)
		if ok {
			field.SetText("art")
			c.Equal(map[tid.TID]map[string]string{id: {"art": "art"}}, links,
				"the link is kept while the names match, in case the rename goes on")
			field.SetText("arts")
		}
	})
	c.Equal(map[tid.TID]map[string]string{id: {"art": "arts"}}, links, "the link follows the rest of the rename")
}

func TestNameablesPanelCollapses(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{"style": {Label: "Style", FreeForm: true}}
	texts := []string{"@style@ speaks @lang@ and @lang@"}
	var p *nameablesPanel
	screen.Do(func() {
		p = newNameablesPanel(&defs, func() map[string]int { return nameableUses(texts, nil, nil, nil) }, nil, nil)
		c.True(p.collapse.collapsed, "a panel with nameables starts collapsed")
		paragraph, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
		c.True(ok, "collapsed, it shows a paragraph")
		if ok {
			c.Equal("@style@ asks for Style as any text; used once. @lang@ is used 2 times but isn't defined.",
				paragraph.plainText(), "a code used but not defined shows while collapsed too")
		}
	})
	showInTestWindow(t, screen, 700, p)
	screen.Do(p.collapse.toggle)
	screen.Do(func() { c.NotNil(p.FindRefKey(pathOf("style")+keySentence), "expanded, it shows the rows") })

	var none map[string]nameable.Detail
	screen.Do(func() {
		empty := newNameablesPanel(&none, func() map[string]int { return nil }, nil, nil)
		c.False(empty.collapse.collapsed, "a panel with none starts open to add one")
		c.NotNil(empty.FindRefKey("nameables:empty"), "with its placeholder")
	})
}

// The trait editor's section edits the editor's copy, so the trait only gets its nameables and the links for its
// modifiers' codes when the edit is applied.
func TestTraitEditorAppliesNameablesAndLinks(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Trained in @style@"
	mod := gurps.NewTraitModifier(sheet.Entity(), nil, false)
	mod.Name = "@art@ focus"
	trait.SetModifiers([]*gurps.TraitModifier{mod})
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	p, ok := uxtest.FirstPanelOfType[*nameablesPanel](content)
	c.True(ok, "expected a nameables panel in the trait editor")
	if !ok {
		return
	}
	style := map[string]nameable.Detail{"style": {Label: "Style", FreeForm: true}}
	*p.defs = maps.Clone(style)
	p.setLink(mod.TID, "art", "style")
	c.Nil(trait.Nameables, "the trait is untouched until the edit is applied")
	c.Nil(trait.NameableLinks)

	e.editorData.ApplyTo(trait)
	c.Equal(style, trait.Nameables)
	c.Equal(map[tid.TID]map[string]string{mod.TID: {"art": "style"}}, trait.NameableLinks)
	trait.Replacements = map[string]string{"style": "Judo"}
	c.Equal("Judo focus", trait.Modifiers[0].NameWithReplacements(), "the applied link resolves")
}

func TestNameablesPanelNameFieldFillsItsRow(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{"style": {Label: "Style"}}
	texts := []string{"@style@"}
	p, _ := showNameablesPanel(t, screen, &defs, &texts, nil, nil, pathOf("style"))
	screen.Do(func() {
		field := p.FindRefKey("n:style:name")
		c.NotNil(field)
		if field == nil {
			return
		}
		row := field.Parent().ContentRect(false)
		c.Equal(row.Right(), field.FrameRect().Right(), "the name takes the rest of its row")
		c.True(field.FrameRect().Width > 300, "which is wider than its minimum")
	})
}

// What is typed into a field of an open row survives the sync that typing into another field of it brings about, which
// must not hand the stale text of the row's copy back to the first field's setter.
func TestNameablesPanelFieldsKeepWhatWasTyped(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defs := map[string]nameable.Detail{"style": {Label: "Style", Tooltip: "Which?"}}
	texts := []string{"@style@"}
	p, host := showNameablesPanel(t, screen, &defs, &texts, nil, nil, pathOf("style"))
	screen.Do(func() {
		field, ok := p.FindRefKey("n:style:name").Self.(*StringField)
		c.True(ok, "the open row has a name field")
		if ok {
			field.RequestFocus()
			field.SetSelectionToEnd()
		}
	})
	screen.Type("s")
	c.Equal("Styles", defs["style"].Label)
	screen.KeyPress(unison.KeyTab, 0)
	screen.Do(func() {
		focus := p.Window().Focus()
		c.NotNil(focus, "Tab moves the focus")
		if focus != nil {
			c.Equal("n:style:descfield", focus.RefKey, "to the description field")
			if field, ok := focus.Self.(*StringField); ok {
				field.SetSelectionToEnd()
			}
		}
	})
	screen.Type("!")
	c.Equal("Which?!", defs["style"].Tooltip)
	c.Equal("Styles", defs["style"].Label, "the name keeps what was typed into it")
	c.Equal("Undo Description", host.mgr.UndoTitle(), "and nothing else is recorded")
	screen.Do(host.mgr.Undo)
	c.Equal("Which?", defs["style"].Tooltip)
	c.Equal("Styles", defs["style"].Label, "undoing the description leaves the name alone")
}
