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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// showSkillLevelCapPanel shows a cap editor for the cap in a window, within the content the dialog would hold it in,
// which provides its undo manager. The editor names attributes as a new sheet does.
func showSkillLevelCapPanel(t *testing.T, screen *unison.HeadlessScreen, lc *gurps.SkillLevelCap) (*skillLevelCapPanel, *sentenceDialogContent) {
	t.Helper()
	var p *skillLevelCapPanel
	var host *sentenceDialogContent
	screen.Do(func() {
		host = newSentenceDialogContent()
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		p = newSkillLevelCapPanel(gurps.NewEntity(), lc)
		host.AddChild(p)
	})
	showInTestWindow(t, screen, 700, host)
	return p, host
}

// newPanelTestLevelCap returns a cap on skills whose name contains "Magic" at the lowest of 12, skill Thaumatology -2,
// the highest of spell Enchantment and trait Magery +10, IQ +2, and a value of a kind this version doesn't know.
func newPanelTestLevelCap() *gurps.SkillLevelCap {
	lc := gurps.NewSkillLevelCap()
	lc.NameCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Magic"}
	root := lc.Cap
	skill := gurps.NewReferenceLevelCap(nil, root, levelcap.Skill)
	skill.NameCriteria.Qualifier = "Thaumatology"
	skill.Modifier = -fxp.Two
	highest := gurps.NewLevelCapGroup(root)
	highest.Highest = true
	spell := gurps.NewReferenceLevelCap(nil, highest, levelcap.Spell)
	spell.NameCriteria.Qualifier = "Enchantment"
	trait := gurps.NewReferenceLevelCap(nil, highest, levelcap.Trait)
	trait.NameCriteria.Qualifier = "Magery"
	trait.Modifier = fxp.Ten
	highest.Children = gurps.LevelCaps{spell, trait}
	iq := gurps.NewReferenceLevelCap(nil, root, levelcap.Attribute)
	iq.Attribute = gurps.IntelligenceID
	iq.Modifier = fxp.Two
	unknown := gurps.NewUnknownLevelCap("sparkle", []byte(`{"type":"sparkle"}`))
	unknown.Parent = root
	root.Children = gurps.LevelCaps{gurps.NewFixedLevelCap(root, fxp.Twelve), skill, highest, iq, unknown}
	return lc
}

// levelCapSentence returns the text of the sentence of the row at the path, without its emphasis. Call it on the UI
// thread.
func levelCapSentence(t *testing.T, p *skillLevelCapPanel, path string) string {
	t.Helper()
	if b, ok := refAs[*sentenceButton](t, p.AsPanel(), path+keySentence); ok {
		return stripEm.Replace(b.text)
	}
	return ""
}

// levelCapKind returns what the node at the path is: "lowest" or "highest" for a group, and "value" otherwise. Call it
// on the UI thread.
func levelCapKind(p *skillLevelCapPanel, path string) string {
	group := asLevelCapGroup(p.node(path))
	switch {
	case group == nil:
		return "value"
	case group.Highest:
		return "highest"
	default:
		return "lowest"
	}
}

// TestSkillLevelCapPanelRows checks that each value reads as its sentence, that each group's pill and accessible name
// say whether it takes the lowest or the highest of its children, that a value of a kind this version doesn't know is
// static text that says why and doesn't open, and that opening each of the other values changes nothing.
func TestSkillLevelCapPanelRows(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	lc := newPanelTestLevelCap()
	before := gurps.Hash64(lc)
	p, _ := showSkillLevelCapPanel(t, screen, lc)
	screen.Do(func() {
		for path, want := range map[string]string{
			"r.0":   "12",
			"r.1":   "skill Thaumatology -2",
			"r.2.0": "spell Enchantment",
			"r.2.1": "trait Magery +10",
			"r.3":   "IQ +2",
			"r.4":   `an unknown kind of level cap ("sparkle") that needs a newer version of GCS`,
		} {
			c.Equal(want, levelCapSentence(t, p, path), path)
		}
		for path, want := range map[string]string{treeRootPath: "Lowest of", "r.2": "Highest of"} {
			if pill, ok := refAs[*unison.PopupMenu[bool]](t, p.AsPanel(), path+keyPill); ok {
				c.Equal(want, pill.Text(), path)
			}
		}
		if group, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.2:group"); ok {
			c.Equal("Highest of", group.Accessibility.Name)
		}
		c.Equal("spell Enchantment; trait Magery +10", p.treeGroupContents(p.node("r.2")))
		if b, ok := refAs[*sentenceButton](t, p.AsPanel(), "r.4"+keySentence); ok {
			c.Nil(b.onClick, "a value of an unknown kind can't be opened")
			c.Equal(wrapTextForTooltip(unknownLevelCapTooltip()), tooltipText(b.Tooltip), "and says why")
		}
		p.toggle("r.4")
	})
	screen.Do(func() { c.Equal("", p.open, "a value of an unknown kind doesn't open") })
	for _, path := range []string{"r.0", "r.1", "r.2.0", "r.2.1", "r.3"} {
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() { c.NotNil(p.FindRefKey(path+keyFirst), "%s opens", path) })
		c.Equal(before, gurps.Hash64(lc), "opening %s changes nothing", path)
	}
}

// TestSkillLevelCapPanelAdd checks what the root's add button and a group's more menu offer, and that adding a value of
// each kind, to the root or to a nested group, puts it at the end of that group, owned by it, open with the focus in
// its first field, as one step to undo.
func TestSkillLevelCapPanelAdd(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	lc := newPanelTestLevelCap()
	p, host := showSkillLevelCapPanel(t, screen, lc)
	kinds := []string{"Fixed Level", "Skill Level", "Spell Level", "Trait Level", "Attribute Value"}
	screen.Do(func() {
		c.Equal(slices.Concat([]string{"Cap Value"}, kinds, []string{"Structure", "Lowest of Group", "Highest of Group"}),
			menuLabels(p.treeAddEntries(lc.Cap, treeRootPath)))
		c.Equal(slices.Concat([]string{"Add Cap Value"}, kinds, []string{
			"Add Structure", "Lowest of Group", "Highest of Group", "-", "Duplicate", "Move Up", "Move Down",
			"Wrap in Group", "-", "Delete",
		}), menuLabels(p.moreEntries(p.node("r.2"), "r.2")))
	})
	for _, one := range []struct {
		label string
		kind  levelcap.Type
		first string
	}{
		{label: "Fixed Level", kind: levelcap.Fixed, first: "level"},
		{label: "Skill Level", kind: levelcap.Skill, first: "name"},
		{label: "Spell Level", kind: levelcap.Spell, first: "name"},
		{label: "Trait Level", kind: levelcap.Trait, first: "name"},
		{label: "Attribute Value", kind: levelcap.Attribute, first: "modifier"},
	} {
		for _, path := range []string{treeRootPath, "r.2"} {
			var at string
			screen.Do(func() {
				group := asLevelCapGroup(p.node(path))
				at = childPath(path, len(group.Children))
				if path == treeRootPath {
					menuAction(p.treeAddEntries(group, path), one.label)()
				} else {
					menuAction(p.moreEntries(group, path), one.label)()
				}
			})
			screen.Do(func() {
				added := p.node(at)
				c.NotNil(added, "%s goes at the end of %s", one.label, path)
				if added != nil {
					c.Equal(one.kind, added.CapType(), one.label)
					c.True(added.ParentGroup() == asLevelCapGroup(p.node(path)), "%s is owned by %s", one.label, path)
				}
				c.Equal(at, p.open, "%s opens", one.label)
				c.Equal(at+":"+one.first, focusedRefKey(p.Window()), "%s focuses its first field", one.label)
				host.undoMgr.Undo()
			})
			screen.Do(func() { c.Nil(p.node(at), "adding %s to %s is one step to undo", one.label, path) })
		}
	}
}

// TestSkillLevelCapPanelEmptyRoot checks that a new cap's empty root shows its placeholder alone, with no pill, until a
// group type is chosen for it, which sets the root's own kind rather than adding a group, and which undo takes back,
// and that an empty root that takes the highest shows its pill.
func TestSkillLevelCapPanelEmptyRoot(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	lc := gurps.NewSkillLevelCap()
	p, host := showSkillLevelCapPanel(t, screen, lc)
	screen.Do(func() {
		c.Nil(p.FindRefKey(treeRootPath+keyPill), "an untouched empty root shows no pill")
		if empty, ok := refAs[*unison.Button](t, p.AsPanel(), treeRootPath+":empty"); ok {
			c.Equal("No cap values. Click here to add one.", empty.Text.String())
		}
		menuAction(p.treeAddEntries(lc.Cap, treeRootPath), "Highest of Group")()
	})
	screen.Do(func() {
		c.True(lc.Cap.Highest, "Highest of Group makes the bare root take the highest")
		c.Equal(0, len(lc.Cap.Children), "rather than adding a group")
		c.NotNil(p.FindRefKey(treeRootPath+keyPill), "and shows its pill")
		host.undoMgr.Undo()
	})
	screen.Do(func() {
		c.False(lc.Cap.Highest, "undo takes it back")
		c.Nil(p.FindRefKey(treeRootPath+keyPill), "pill and all")
		menuAction(p.treeAddEntries(lc.Cap, treeRootPath), "Lowest of Group")()
	})
	screen.Do(func() {
		c.False(lc.Cap.Highest, "Lowest of Group leaves the bare root taking the lowest")
		c.NotNil(p.FindRefKey(treeRootPath+keyPill), "and shows its pill")
	})

	highest := gurps.NewSkillLevelCap()
	highest.Cap.Highest = true
	p, _ = showSkillLevelCapPanel(t, screen, highest)
	screen.Do(func() {
		c.NotNil(p.FindRefKey(treeRootPath+keyPill), "an empty root that takes the highest shows its pill")
		if empty, ok := refAs[*unison.Button](t, p.AsPanel(), treeRootPath+":empty"); ok {
			c.Equal("Empty group. Add a cap value or drag one here.", empty.Text.String())
		}
	})
}

// TestSkillLevelCapPanelTypeChange checks that switching a value to another kind replaces it with one of that kind,
// handing the name and adjustment on between values resting on a skill, spell or trait but not to or from a fixed
// level or an attribute, and that undo puts the value back as it was.
func TestSkillLevelCapPanelTypeChange(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	plain := func(s string) string { return s }
	for _, one := range []struct {
		from, to levelcap.Type
		want     string
	}{
		{from: levelcap.Skill, to: levelcap.Spell, want: "spell Thaumatology -2"},
		{from: levelcap.Spell, to: levelcap.Trait, want: "trait Thaumatology -2"},
		{from: levelcap.Trait, to: levelcap.Skill, want: "skill Thaumatology -2"},
		{from: levelcap.Skill, to: levelcap.Attribute, want: "DX"},
		{from: levelcap.Attribute, to: levelcap.Skill, want: `skill ""`},
		{from: levelcap.Skill, to: levelcap.Fixed, want: "12"},
		{from: levelcap.Fixed, to: levelcap.Spell, want: `spell ""`},
	} {
		lc := gurps.NewSkillLevelCap()
		var source gurps.LevelCap
		if one.from == levelcap.Fixed {
			source = gurps.NewFixedLevelCap(lc.Cap, fxp.FromInteger(14))
		} else {
			ref := gurps.NewReferenceLevelCap(nil, lc.Cap, one.from)
			ref.Modifier = -fxp.Two
			if one.from == levelcap.Attribute {
				ref.Attribute = gurps.IntelligenceID
			} else {
				ref.NameCriteria.Qualifier = "Thaumatology"
			}
			if one.from == levelcap.Skill {
				ref.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Ritual"}
			}
			source = ref
		}
		lc.Cap.Children = gurps.LevelCaps{source}
		before := gurps.Hash64(lc)
		p, host := showSkillLevelCapPanel(t, screen, lc)
		screen.Do(func() { p.toggle("r.0") })
		screen.Do(func() {
			if popup, ok := refAs[*unison.PopupMenu[levelcap.Type]](t, p.AsPanel(), "r.0:type"); ok {
				popup.Select(one.to)
			}
		})
		screen.Do(func() {
			changed := p.node("r.0")
			c.Equal(one.to, changed.CapType(), "%s to %s", one.from, one.to)
			c.Equal(one.want, changed.Describe(p.entity, plain), "%s to %s", one.from, one.to)
			c.True(changed.ParentGroup() == lc.Cap, "%s to %s keeps the value in its group", one.from, one.to)
			c.Equal("r.0", p.open, "%s to %s leaves the row open", one.from, one.to)
			host.undoMgr.Undo()
		})
		c.Equal(before, gurps.Hash64(lc), "undoing %s to %s puts the value back", one.from, one.to)
	}
}

// TestSkillLevelCapPanelGroupKinds checks that a group's pill switches it between taking the lowest and the highest of
// its children, that Wrap in Group puts a value in a group of the other kind than the one holding it, and that Ungroup
// is offered only where taking the group away keeps what it caps at: for a group of its parent's kind, or of one value.
func TestSkillLevelCapPanelGroupKinds(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	lc := newPanelTestLevelCap()
	p, host := showSkillLevelCapPanel(t, screen, lc)
	kind := func(path string) string {
		var k string
		screen.Do(func() { k = levelCapKind(p, path) })
		return k
	}
	offersUngroup := func(path string) bool {
		var offered bool
		screen.Do(func() { offered = slices.Contains(menuLabels(p.moreEntries(p.node(path), path)), "Ungroup") })
		return offered
	}
	c.False(offersUngroup("r.2"), "a group of two of the other kind than its parent can't be ungrouped")
	screen.Do(func() {
		if pill, ok := refAs[*unison.PopupMenu[bool]](t, p.AsPanel(), "r.2"+keyPill); ok {
			pill.Select(false)
		}
	})
	c.Equal("lowest", kind("r.2"), "the pill makes the group take the lowest")
	c.True(offersUngroup("r.2"), "a group of its parent's kind can be ungrouped")
	screen.Do(host.undoMgr.Undo)
	c.Equal("highest", kind("r.2"), "undo takes the pill's choice back")

	screen.Do(func() { menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Wrap in Group")() })
	c.Equal("highest", kind("r.0"), "a value of a lowest-of group is wrapped in a highest-of group")
	c.Equal("value", kind("r.0.0"), "which holds it")
	c.True(offersUngroup("r.0"), "a group of one can be ungrouped whatever its kind")
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.2.1"), "r.2.1"), "Wrap in Group")() })
	c.Equal("lowest", kind("r.2.1"), "a value of a highest-of group is wrapped in a lowest-of group")
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Ungroup")() })
	screen.Do(func() {
		_, fixed := p.node("r.0").(*gurps.FixedLevelCap)
		c.True(fixed, "Ungroup puts the value back in place of its group")
		if fixed {
			c.True(p.node("r.0").ParentGroup() == lc.Cap, "owned by the group's parent")
		}
	})
}

// TestSkillLevelCapPanelMatcher checks that the matcher offers the criteria its target consults, a specialization for
// skills and a college and power source for spells, and tags for either, with the buttons that add them after the
// chips in use; that a change of target clears, in the same step, the criteria it no longer consults; and that undo
// puts them back.
func TestSkillLevelCapPanelMatcher(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	chipKeys := func(p *skillLevelCapPanel) []string {
		var keys []string
		screen.Do(func() {
			if tags, ok := refAs[*unison.Panel](t, p.AsPanel(), levelCapMatcherKey+":tag"+keyChip); ok {
				for _, child := range tags.Parent().Children() {
					keys = append(keys, child.RefKey)
				}
			}
		})
		return keys
	}

	lc := gurps.NewSkillLevelCap()
	lc.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Occult"}
	p, _ := showSkillLevelCapPanel(t, screen, lc)
	c.Equal([]string{"m:tag:chip", "m:add specialization"}, chipKeys(p),
		"skills offer a specialization, whose button follows the chips in use")

	ritual := criteria.Text{Compare: criteria.IsText, Qualifier: "Ritual"}
	fire := criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
	arcane := criteria.Text{Compare: criteria.IsText, Qualifier: "Arcane"}
	lc = gurps.NewSkillLevelCap()
	lc.Target = levelcap.SkillsAndSpells
	lc.NameCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Magic"}
	lc.SpecializationCriteria = ritual
	lc.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Occult"}
	lc.CollegeCriteria = fire
	lc.PowerSourceCriteria = arcane
	before := gurps.Hash64(lc)
	p, host := showSkillLevelCapPanel(t, screen, lc)
	c.Equal([]string{"m:specialization:chip", "m:college:chip", "m:power:chip", "m:tag:chip"}, chipKeys(p),
		"skills and spells offer every criterion")
	for _, one := range []struct {
		target                         levelcap.Target
		keys                           []string
		specialization, college, power criteria.Text
	}{
		{target: levelcap.Skills, keys: []string{"m:specialization:chip", "m:tag:chip"}, specialization: ritual},
		{target: levelcap.Spells, keys: []string{"m:college:chip", "m:power:chip", "m:tag:chip"}, college: fire, power: arcane},
	} {
		screen.Do(func() {
			if popup, ok := refAs[*unison.PopupMenu[levelcap.Target]](t, p.AsPanel(), levelCapMatcherKey+":target"); ok {
				popup.Select(one.target)
			}
		})
		c.Equal(one.keys, chipKeys(p), one.target.String())
		screen.Do(func() {
			c.Equal(one.target, lc.Target)
			c.Equal(one.specialization, lc.SpecializationCriteria, "%s: specialization", one.target)
			c.Equal(one.college, lc.CollegeCriteria, "%s: college", one.target)
			c.Equal(one.power, lc.PowerSourceCriteria, "%s: power source", one.target)
			c.Equal("Magic", lc.NameCriteria.Qualifier, "%s keeps the name", one.target)
			c.Equal("Occult", lc.TagsCriteria.Qualifier, "%s keeps the tags", one.target)
			host.undoMgr.Undo()
		})
		c.Equal(before, gurps.Hash64(lc), "undoing the change to %s puts back what it cleared", one.target)
	}
}

// TestSkillLevelCapPanelControlNamesDiffer checks that no two controls of an open value of each kind share an
// accessible name, and that none shares one with the matcher's controls, whether the specializations of the matcher and
// of a skill's value are unused, offered by their add buttons, or in use, as chips.
func TestSkillLevelCapPanelControlNamesDiffer(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	for _, used := range []bool{false, true} {
		lc := newPanelTestLevelCap()
		if used {
			ritual := criteria.Text{Compare: criteria.IsText, Qualifier: "Ritual"}
			lc.SpecializationCriteria = ritual
			if skill, ok := lc.Cap.Children[1].(*gurps.ReferenceLevelCap); ok {
				skill.SpecializationCriteria = ritual
			}
		}
		p, _ := showSkillLevelCapPanel(t, screen, lc)
		var wnd *unison.Window
		var matcherControls []*unison.Panel
		screen.Do(func() {
			wnd = p.Window()
			matcher := p.FindRefKey(levelCapMatcherKey + ":target")
			for matcher != nil && matcher.Parent() != p.AsPanel() {
				matcher = matcher.Parent()
			}
			if matcher != nil {
				matcher.HasInSelfOrDescendants(func(one *unison.Panel) bool {
					if one.Focusable() {
						matcherControls = append(matcherControls, one)
					}
					return false
				})
			}
		})
		c.NotNil(screen.AccessibilityTree(wnd))
		matcherNames := make(map[string]bool)
		for _, one := range matcherControls {
			if node := screen.AccessibilityNodeFor(one); node != nil {
				matcherNames[node.Name] = true
			}
		}
		c.True(matcherNames["Name"] && matcherNames["Name Comparison"], "used %v: the matcher's name controls", used)
		for _, path := range []string{"r.0", "r.1", "r.2.0", "r.2.1", "r.3"} {
			screen.Do(func() { p.toggle(path) })
			var controls []*unison.Panel
			screen.Do(func() {
				if editor, ok := refAs[*unison.Panel](t, p.AsPanel(), path+keyFirst); ok {
					editor.HasInSelfOrDescendants(func(one *unison.Panel) bool {
						if one.Focusable() {
							controls = append(controls, one)
						}
						return false
					})
				}
			})
			c.NotNil(screen.AccessibilityTree(wnd))
			c.True(len(controls) >= 2, path)
			names := make(map[string]bool)
			for _, one := range controls {
				node := screen.AccessibilityNodeFor(one)
				c.NotNil(node, path)
				if node == nil {
					continue
				}
				c.NotEqual("", node.Name, "used %v, %s: every control has a name", used, path)
				c.False(names[node.Name], "used %v, %s: %s is shared", used, path, node.Name)
				c.False(matcherNames[node.Name], "used %v, %s: %s is the matcher's as well", used, path, node.Name)
				names[node.Name] = true
			}
		}
	}
}
