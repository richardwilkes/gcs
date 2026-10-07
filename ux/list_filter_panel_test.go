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
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// listFilterPanelTestKey is the list type the panel tests edit filters for: equipment, which has a field of every kind.
const listFilterPanelTestKey = "eqp"

// showListFilterPanel shows a filter editor for the filter in a window, within the content the dialog would hold it in,
// which provides its undo manager. The session's memory of the field last used is swapped out for an empty one, so that
// one test can't influence another.
func showListFilterPanel(t *testing.T, screen *unison.HeadlessScreen, filter *gurps.ListFilter) (*listFilterPanel, *listFilterDialogContent) {
	t.Helper()
	return showListFilterPanelFor(t, screen, filter, filterFieldInfos(gurps.EquipmentFilterFields()))
}

// showListFilterPanelFor shows a filter editor as showListFilterPanel does, offering the fields.
func showListFilterPanelFor(t *testing.T, screen *unison.HeadlessScreen, filter *gurps.ListFilter, fields []filterFieldInfo) (*listFilterPanel, *listFilterDialogContent) {
	t.Helper()
	swapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	var p *listFilterPanel
	var host *listFilterDialogContent
	screen.Do(func() {
		host = newListFilterDialogContent()
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		p = newListFilterPanel(listFilterPanelTestKey, filter, fields)
		host.AddChild(p)
	})
	showInTestWindow(t, screen, 700, host)
	return p, host
}

// newTestListFilter returns a filter whose root requires all of: a name containing "sword", a weight of at most 5 lb,
// none of a "Shield" or "Buckler" tag or a cost of at least 1000, not being a container, a condition on a field this
// version doesn't know, and a node of a kind it doesn't know.
func newTestListFilter() *gurps.ListFilter {
	f := gurps.NewListFilter("Light blades")
	root := f.Root
	name := gurps.NewFilterCondition(root, "name")
	name.Text = criteria.Text{Compare: criteria.ContainsText, Qualifier: "sword"}
	weight := gurps.NewFilterCondition(root, "weight")
	weight.Weight = criteria.Weight{Compare: criteria.AtMostNumber, Qualifier: fxp.WeightFromInteger(5, fxp.Pound)}
	none := gurps.NewFilterGroup(root)
	none.All = false
	none.Not = true
	tags := gurps.NewFilterCondition(none, "tags")
	tags.Text = criteria.Text{Compare: criteria.IsText, Qualifier: "Shield, Buckler"}
	cost := gurps.NewFilterCondition(none, "cost")
	cost.Number = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.FromInteger(1000)}
	none.Children = gurps.FilterNodes{tags, cost}
	container := gurps.NewFilterCondition(root, "container")
	container.Not = true
	future := gurps.NewFilterCondition(root, "future_field")
	unknown := gurps.NewUnknownFilterNode("sparkle", []byte(`{"type":"sparkle"}`))
	unknown.Parent = root
	root.Children = gurps.FilterNodes{name, weight, none, container, future, unknown}
	return f
}

// filterShape returns the field of each of the root's conditions, and for each group its mode, then the shape of its
// children, then ":end".
func filterShape(root *gurps.FilterGroup) []string {
	var shape []string
	var walk func(group *gurps.FilterGroup)
	walk = func(group *gurps.FilterGroup) {
		for _, one := range group.Children {
			switch node := one.(type) {
			case *gurps.FilterCondition:
				shape = append(shape, node.Field)
			case *gurps.FilterGroup:
				shape = append(shape, groupModeOf(node).String()+":")
				walk(node)
				shape = append(shape, ":end")
			default:
				shape = append(shape, "unknown")
			}
		}
	}
	walk(root)
	return shape
}

// filterConditionAt returns the condition at the path. Without one, it fails the test and returns an empty condition,
// so that the caller can go on. Call it on the UI thread.
func filterConditionAt(t *testing.T, p *listFilterPanel, path string) *gurps.FilterCondition {
	t.Helper()
	cond, ok := p.node(path).(*gurps.FilterCondition)
	if !ok {
		t.Errorf("no condition at %s", path)
		return &gurps.FilterCondition{}
	}
	return cond
}

// refAs returns the panel within p that has the ref key as T: the panel itself when T is *unison.Panel, and its Self
// otherwise. Without one, it fails the test and returns false, so that the caller can skip what it would have done with
// it. Call it on the UI thread.
func refAs[T any](t *testing.T, p *unison.Panel, key string) (T, bool) {
	t.Helper()
	var zero T
	one := p.FindRefKey(key)
	if one == nil {
		t.Errorf("nothing has the ref key %s", key)
		return zero, false
	}
	if v, ok := any(one).(T); ok {
		return v, true
	}
	if v, ok := one.Self.(T); ok {
		return v, true
	}
	t.Errorf("%s is a %T, not a %T", key, one.Self, zero)
	return zero, false
}

// focusedRefKey returns the ref key of what has the window's focus, or "" when nothing does. Call it on the UI thread.
func focusedRefKey(wnd *unison.Window) string {
	if focus := wnd.Focus(); focus != nil {
		return focus.RefKey
	}
	return ""
}

// sentenceText returns the text of the sentence of the row at the path, without its emphasis. Call it on the UI thread.
func sentenceText(p *listFilterPanel, path string) string {
	if one := p.FindRefKey(path + keySentence); one != nil {
		if b, ok := one.Self.(*sentenceButton); ok {
			return stripEm.Replace(b.text)
		}
	}
	return ""
}

func TestListFilterPanelRows(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := newTestListFilter()
	before := gurps.Hash64(f.Root)
	p, _ := showListFilterPanel(t, screen, f)
	screen.Do(func() {
		for path, want := range map[string]string{
			"r.0":   `Must have a name that contains "sword"`,
			"r.1":   `Must have a weight that is at most 5 lb`,
			"r.2.0": `Must have tags where at least one is "Shield" or "Buckler"`,
			"r.2.1": `Must have a cost that is at least 1,000`,
			"r.3":   `Must not be a container`,
			"r.4":   `Condition on unknown field "future_field"; it can't be checked`,
			"r.5":   `Unknown filter node type "sparkle"; it will be preserved, but can't be checked`,
		} {
			c.Equal(want, sentenceText(p, path), path)
		}
		for path, want := range map[string]string{"r": "All of", "r.2": "None of"} {
			if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), path+keyPill); ok {
				c.Equal(want, pill.Text(), path)
				c.Contains(strings.ReplaceAll(tooltipText(pill.Tooltip), "\n", " "), "An empty group is left out",
					"%s's pill explains the choices, an empty group's among them", path)
			}
		}
		if group, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.2:group"); ok {
			c.Equal("None of", group.Accessibility.Name)
		}
		for path, tip := range map[string]string{
			"r.4": unknownFilterFieldTooltip(),
			"r.5": preservedFilterNodeTooltip(),
		} {
			if b, ok := refAs[*sentenceButton](t, p.AsPanel(), path+keySentence); ok {
				c.Nil(b.onClick, "%s can't be opened", path)
				c.Equal(wrapTextForTooltip(tip), tooltipText(b.Tooltip), path)
			}
		}
		p.toggle("r.4")
	})
	screen.Do(func() {
		c.Equal("", p.open, "a condition on a field this version doesn't know doesn't open")
		c.Equal(before, gurps.Hash64(f.Root), "showing the filter changes nothing")
	})
}

// TestListFilterPanelListSentences checks the sentences of list conditions that accept anything or compare with an
// empty value, and of a value with space around it, which is quoted so that the space shows.
func TestListFilterPanelListSentences(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := gurps.NewListFilter("")
	anything := gurps.NewFilterCondition(f.Root, "tags")
	none := gurps.NewFilterCondition(f.Root, "tags")
	none.Text = criteria.Text{Compare: criteria.IsText}
	padded := gurps.NewFilterCondition(f.Root, "name")
	padded.Text = criteria.Text{Compare: criteria.IsText, Qualifier: " Axe "}
	f.Root.Children = gurps.FilterNodes{anything, none, padded}
	p, _ := showListFilterPanel(t, screen, f)
	screen.Do(func() {
		c.Equal(`Must have tags that are anything`, sentenceText(p, "r.0"))
		c.Equal(`Must have tags where at least one is ""`, sentenceText(p, "r.1"))
		c.Equal(`Must have a name that is " Axe "`, sentenceText(p, "r.2"))
	})
}

// TestListFilterPanelPluralFields checks that a condition on a field whose title is plural agrees with it, in its
// sentence and in the choices of its comparison, for text and numbers alike.
func TestListFilterPanelPluralFields(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := gurps.NewListFilter("")
	notes := gurps.NewFilterCondition(f.Root, "notes")
	notes.Not = true
	notes.Text = criteria.Text{Compare: criteria.StartsWithText, Qualifier: "Cheap"}
	points := gurps.NewFilterCondition(f.Root, "points")
	points.Number = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.FromInteger(5)}
	name := gurps.NewFilterCondition(f.Root, "name")
	name.Text = criteria.Text{Compare: criteria.StartsWithText, Qualifier: "A"}
	f.Root.Children = gurps.FilterNodes{notes, points, name}
	p, _ := showListFilterPanelFor(t, screen, f, filterFieldInfos(gurps.TraitFilterFields()))
	screen.Do(func() {
		c.Equal(`Must not have notes that start with "Cheap"`, sentenceText(p, "r.0"))
		c.Equal(`Must have points that are at least 5`, sentenceText(p, "r.1"))
		c.Equal(`Must have a name that starts with "A"`, sentenceText(p, "r.2"), "a singular title is left alone")
		p.toggle("r.0")
	})
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[criteria.StringComparison]](t, p.AsPanel(), "r.0:textcmp"); ok {
			c.Equal("that start with", popup.Text())
			items := make([]string, 0, popup.ItemCount())
			for i := range popup.ItemCount() {
				item, _ := popup.ItemAt(i)
				items = append(items, popup.ItemRendererCallback(item))
			}
			c.Equal([]string{
				"that are anything", "that are", "that are not", "that contain", "that do not contain",
				"that start with", "that do not start with", "that end with", "that do not end with",
			}, items, "every choice agrees with notes")
		}
		p.toggle("r.1")
	})
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[criteria.NumericComparison]](t, p.AsPanel(), "r.1:numbercmp"); ok {
			c.Equal("that are at least", popup.Text())
			items := make([]string, 0, popup.ItemCount())
			for i := range popup.ItemCount() {
				item, _ := popup.ItemAt(i)
				items = append(items, popup.ItemRendererCallback(item))
			}
			c.Equal([]string{"that are anything", "that are", "that are not", "that are at least", "that are at most"},
				items, "every choice agrees with points")
		}
		p.toggle("r.2")
	})
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[criteria.StringComparison]](t, p.AsPanel(), "r.2:textcmp"); ok {
			c.Equal("that starts with", popup.Text(), "a singular title keeps the singular choices")
		}
	})
}

func TestListFilterPanelEmptyRoot(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := gurps.NewListFilter("")
	p, host := showListFilterPanel(t, screen, f)
	screen.Do(func() {
		c.Nil(p.FindRefKey("r"+keyPill), "an untouched empty root shows no pill")
		if empty, ok := refAs[*unison.Button](t, p.AsPanel(), "r:empty"); ok {
			c.Equal("No conditions. Click here to add one.", empty.Text.String())
		}
		c.Equal([]string{"Condition", "New Condition", "Structure", "All of Group", "Any of Group"},
			menuLabels(p.treeAddEntries(f.Root, "r")))
		menuAction(p.treeAddEntries(f.Root, "r"), "All of Group")()
	})
	screen.Do(func() {
		c.NotNil(p.FindRefKey("r"+keyPill), "choosing the group type shows the pill")
		c.True(f.Root.All)
		c.Equal(0, len(f.Root.Children), "and adds no group")
		host.undoMgr.Undo()
	})
	screen.Do(func() {
		c.Nil(p.FindRefKey("r"+keyPill), "which undo takes back")
		c.Equal("r"+keyAdd, focusedRefKey(p.Window()), "giving the focus to what adds to the root")
	})

	for _, mode := range []filterGroupMode{filterAnyOf, filterNoneOf, filterNotAllOf} {
		f = gurps.NewListFilter("")
		mode.apply(f.Root)
		p, _ = showListFilterPanel(t, screen, f)
		screen.Do(func() {
			c.NotNil(p.FindRefKey("r"+keyPill), "an empty root that is %s shows its pill", mode)
			if empty, ok := refAs[*unison.Button](t, p.AsPanel(), "r:empty"); ok {
				c.Equal("Empty group. Add a condition or drag one here.", empty.Text.String())
			}
		})
	}

	// An empty root that shows its pill has one add button, beside its placeholder, and adding a group to it adds a
	// group rather than changing its own type.
	f = gurps.NewListFilter("")
	filterNoneOf.apply(f.Root)
	p, _ = showListFilterPanel(t, screen, f)
	screen.Do(func() {
		c.Nil(p.FindRefKey("r"+keyMore), "the head has no add button of its own")
		c.NotNil(p.FindRefKey("r"+keyAdd), "the placeholder's add button is the one")
		menuAction(p.treeAddEntries(f.Root, "r"), "All of Group")()
	})
	screen.Do(func() {
		c.Equal(filterNoneOf, groupModeOf(f.Root), "the root keeps its type")
		group, ok := p.node("r.0").(*gurps.FilterGroup)
		c.True(ok && group.All && !group.Not, "and holds a new all of group")
		c.NotNil(p.FindRefKey("r"+keyMore), "which brings back the head's add button")
	})

	// A root given a head by a group type and then a pill choice keeps it while it holds something and after that is
	// undone, and shows its placeholder alone again once undo takes back both choices. Choosing a group type for it
	// then sets its type.
	f = gurps.NewListFilter("")
	p, host = showListFilterPanel(t, screen, f)
	screen.Do(func() { menuAction(p.treeAddEntries(f.Root, "r"), "Any of Group")() })
	screen.Do(func() {
		c.Equal(filterAnyOf, groupModeOf(f.Root), "Any of Group makes the bare root any of")
		if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), "r"+keyPill); ok {
			pill.Select(filterNoneOf)
		}
	})
	screen.Do(func() {
		c.Equal(filterNoneOf, groupModeOf(f.Root), "the pill makes it none of")
		menuAction(p.treeAddEntries(f.Root, "r"), "New Condition")()
	})
	screen.Do(func() {
		c.Equal(1, len(f.Root.Children))
		host.undoMgr.Undo()
	})
	screen.Do(func() {
		c.Equal(0, len(f.Root.Children), "undo empties the root")
		c.NotNil(p.FindRefKey("r"+keyPill), "which keeps its head")
		host.undoMgr.Undo()
		host.undoMgr.Undo()
	})
	screen.Do(func() {
		c.Nil(p.FindRefKey("r"+keyPill), "undoing the choices too leaves the placeholder alone")
		c.Equal(filterAllOf, groupModeOf(f.Root))
		menuAction(p.treeAddEntries(f.Root, "r"), "Any of Group")()
	})
	screen.Do(func() { c.Equal(filterAnyOf, groupModeOf(f.Root), "choosing a group type then sets it again") })

	// A root that only a pill choice gave a head keeps it once it is empty again, even back at All of.
	f = gurps.NewListFilter("")
	p, _ = showListFilterPanel(t, screen, f)
	screen.Do(func() { menuAction(p.treeAddEntries(f.Root, "r"), "New Condition")() })
	for _, mode := range []filterGroupMode{filterAnyOf, filterAllOf} {
		screen.Do(func() {
			if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), "r"+keyPill); ok {
				pill.Select(mode)
			}
		})
		screen.Do(func() { c.Equal(mode, groupModeOf(f.Root), "the pill makes the root %s", mode) })
	}
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Delete")() })
	screen.Do(func() {
		c.Equal(0, len(f.Root.Children), "deleting its condition empties the root")
		c.NotNil(p.FindRefKey("r"+keyPill), "which keeps the head the pill choice gave it")
	})

	// Choosing All of in the pill of an empty root keeps the pill.
	f = gurps.NewListFilter("")
	filterAnyOf.apply(f.Root)
	p, _ = showListFilterPanel(t, screen, f)
	screen.Do(func() {
		if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), "r"+keyPill); ok {
			pill.Select(filterAllOf)
		}
	})
	screen.Do(func() {
		c.True(f.Root.All)
		c.NotNil(p.FindRefKey("r"+keyPill), "the pill stays once a mode has been chosen")
	})
}

func TestListFilterPanelAddCondition(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := newTestListFilter()
	p, host := showListFilterPanel(t, screen, f)
	screen.Do(func() {
		lastFilterFieldKeyUsed[listFilterPanelTestKey] = "weight"
		c.Equal([]string{
			"Add Condition", "New Condition", "Add Structure", "All of Group", "Any of Group", "-", "Duplicate",
			"Move Up", "Move Down", "Wrap in Group", "-", "Delete",
		}, menuLabels(p.moreEntries(p.node("r.2"), "r.2")))
		menuAction(p.treeAddEntries(f.Root, "r"), "New Condition")()
	})
	screen.Do(func() {
		c.Equal("r.6", p.open, "a new condition goes at the end, open")
		cond := filterConditionAt(t, p, "r.6")
		c.Equal("weight", cond.Field, "on the field last chosen")
		c.Equal(criteria.AnyNumber, cond.Weight.Compare, "accepting anything")
		c.Equal("r.6:field", focusedRefKey(p.Window()), "the field popup takes the focus")
		menuAction(p.moreEntries(p.node("r.2"), "r.2"), "Any of Group")()
	})
	screen.Do(func() {
		group, ok := p.node("r.2.2").(*gurps.FilterGroup)
		c.True(ok && !group.All, "a group added through a group's more menu goes at its end")
		c.Equal("r.2.2"+keyMore, focusedRefKey(p.Window()), "and its more button takes the focus")
		host.undoMgr.Undo()
		host.undoMgr.Undo()
	})
	screen.Do(func() {
		c.Equal([]string{"name", "weight", "None of:", "tags", "cost", ":end", "container", "future_field", "unknown"},
			filterShape(f.Root), "each addition is one step to undo")
	})
}

func TestListFilterPanelFieldChange(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := newTestListFilter()
	p, host := showListFilterPanel(t, screen, f)
	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() {
		popup, ok := refAs[*unison.PopupMenu[string]](t, p.AsPanel(), "r.0:field")
		if !ok {
			return
		}
		c.Equal("have a name", popup.Text())
		popup.Select("cost")
	})
	screen.Do(func() {
		cond := filterConditionAt(t, p, "r.0")
		c.Equal("cost", cond.Field)
		c.Equal(criteria.Text{}, cond.Text, "the old field's criterion is cleared")
		c.Equal(criteria.Number{}, cond.Number, "and the new one's accepts anything")
		c.Equal("cost", lastFilterFieldKeyUsed[listFilterPanelTestKey],
			"the field is remembered for the next condition")
		if popup, ok := refAs[*unison.PopupMenu[criteria.NumericComparison]](t, p.AsPanel(), "r.0:numbercmp"); ok {
			c.Equal(len(criteria.NumericComparisons), popup.ItemCount(), "which offers anything")
		}
		c.Nil(p.FindRefKey("r.0:number"), "and no number while it is anything")
		if not, ok := refAs[*unison.PopupMenu[bool]](t, p.AsPanel(), "r.0:not"); ok {
			not.Select(true)
		}
	})
	screen.Do(func() {
		c.True(filterConditionAt(t, p, "r.0").Not, "Must not negates the condition")
		host.undoMgr.Undo()
		host.undoMgr.Undo()
	})
	screen.Do(func() {
		cond := filterConditionAt(t, p, "r.0")
		c.Equal("name", cond.Field, "a field change is one step to undo")
		c.Equal(criteria.Text{Compare: criteria.ContainsText, Qualifier: "sword"}, cond.Text, "criteria and all")
		c.False(cond.Not)
	})
}

func TestListFilterPanelPill(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := newTestListFilter()
	p, _ := showListFilterPanel(t, screen, f)
	for _, one := range []struct {
		mode     filterGroupMode
		all, not bool
	}{
		{filterAllOf, true, false},
		{filterAnyOf, false, false},
		{filterNotAllOf, true, true},
		{filterNoneOf, false, true},
	} {
		screen.Do(func() {
			if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), "r.2"+keyPill); ok {
				pill.Select(one.mode)
			}
		})
		screen.Do(func() {
			group, ok := p.node("r.2").(*gurps.FilterGroup)
			c.True(ok)
			if ok {
				c.Equal(one.all, group.All, one.mode.String())
				c.Equal(one.not, group.Not, one.mode.String())
				c.Equal(one.mode, groupModeOf(group))
			}
		})
	}
}

func TestListFilterPanelMoreMenu(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := newTestListFilter()
	p, host := showListFilterPanel(t, screen, f)
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Wrap in Group")() })
	screen.Do(func() {
		c.Equal([]string{
			"Any of:", "name", ":end", "weight", "None of:", "tags", "cost", ":end", "container",
			"future_field", "unknown",
		}, filterShape(f.Root), "a wrapped node goes in a group of the other kind")
		c.True(slices.Contains(menuLabels(p.moreEntries(p.node("r.0"), "r.0")), "Ungroup"),
			"a group of one can be ungrouped")
		if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), "r.0"+keyPill); ok {
			pill.Select(filterNoneOf)
		}
	})
	screen.Do(func() {
		c.False(slices.Contains(menuLabels(p.moreEntries(p.node("r.0"), "r.0")), "Ungroup"),
			"but not once it is negated")
		host.undoMgr.Undo()
	})
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Ungroup")() })
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.3"), "r.3"), "Duplicate")() })
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.1"), "r.1"), "Move Down")() })
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.1.0"), "r.1.0"), "Move Up")() })
	screen.Do(func() {
		c.Equal("r.1"+keyMore, focusedRefKey(p.Window()), "Move Up steps out of the group, ahead of it")
		c.Equal("weight", filterConditionAt(t, p, "r.1").Field)
		menuAction(p.moreEntries(p.node("r.1"), "r.1"), "Move Down")()
	})
	screen.Do(func() {
		c.Equal("r.1.0"+keyMore, focusedRefKey(p.Window()), "the moved node's more button takes the focus")
		menuAction(p.moreEntries(p.node("r.4"), "r.4"), "Delete")()
	})
	screen.Do(func() {
		c.Equal("r.4"+keyMore, focusedRefKey(p.Window()), "after a delete, what came next takes the focus")
		c.Equal([]string{"name", "None of:", "weight", "tags", "cost", ":end", "container", "container", "unknown"},
			filterShape(f.Root), "moving down steps into a group, and the unknown field is gone")
		group, ok := p.node("r.1").(*gurps.FilterGroup)
		c.True(ok)
		if ok {
			for _, child := range group.Children {
				c.True(child.ParentGroup() == group, "the moved node belongs to its new group")
			}
		}
		host.undoMgr.Undo()
	})
	screen.Do(func() {
		c.Equal("r.4"+keyMore, focusedRefKey(p.Window()), "undoing the delete gives its more button the focus")
		c.Equal("future_field", filterConditionAt(t, p, "r.4").Field, "and puts it back")
	})
}

func TestListFilterPanelKeepsUnknownData(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := newTestListFilter()
	p, _ := showListFilterPanel(t, screen, f)
	var original *gurps.UnknownFilterNode
	screen.Do(func() {
		var ok bool
		original, ok = p.node("r.5").(*gurps.UnknownFilterNode)
		c.True(ok, "the fixture's unknown node is at r.5")
	})
	if original == nil {
		return
	}
	data := bytes.Clone(original.Data)
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.5"), "r.5"), "Duplicate")() })
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.6"), "r.6"), "Move Up")() })
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.6"), "r.6"), "Delete")() })
	screen.Do(func() {
		c.Equal(6, len(f.Root.Children))
		kept, isUnknown := p.node("r.5").(*gurps.UnknownFilterNode)
		c.True(isUnknown, "the copy is left where the original was")
		if isUnknown {
			c.True(kept != original, "and is a copy")
			c.True(bytes.Equal(data, kept.Data), "holding the original data byte for byte")
		}
	})
}

func TestListFilterPanelDragAndDrop(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := newTestListFilter()
	p, host := showListFilterPanel(t, screen, f)
	drag := func(from, onto string, fraction float32, capture string) (accepted bool) {
		var where geom.Point
		data := &rowDrag{panel: p.AsPanel(), path: from}
		screen.Do(func() {
			more, ok := refAs[*unison.Panel](t, p.AsPanel(), onto+keyMore)
			if !ok {
				return
			}
			target := more.Parent()
			r := p.RectFromRoot(target.RectToRoot(target.ContentRect(true)))
			where = geom.NewPoint(r.X+r.Width/3, r.Y+r.Height*fraction)
			p.dragOver(where, data)
			accepted = p.dropTarget != nil
		})
		if capture != "" {
			captureScreen(t, c, screen, capture)
		}
		screen.Do(func() { p.drop(where, data) })
		return accepted
	}
	shape := func() []string {
		var s []string
		screen.Do(func() { s = filterShape(f.Root) })
		return s
	}
	c.False(drag("r.2", "r.2.0", 0.5, ""), "a group can't go into itself")
	c.True(drag("r.3", "r.2", 0.8, "list_filter_drop_into_group"), "below the top of a group's head is into it")
	c.Equal([]string{"name", "weight", "None of:", "tags", "cost", "container", ":end", "future_field", "unknown"},
		shape())
	c.True(drag("r.0", "r.2.0", 0.2, "list_filter_drop_before_row"), "the top of a row is before it")
	c.Equal([]string{"weight", "None of:", "name", "tags", "cost", "container", ":end", "future_field", "unknown"},
		shape())
	screen.Do(host.undoMgr.Undo)
	c.Equal([]string{"name", "weight", "None of:", "tags", "cost", "container", ":end", "future_field", "unknown"},
		shape(), "a drop is one step to undo")

	// Beside the last child of a group is after the group.
	screen.Do(func() {
		group, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.2:group")
		if !ok {
			return
		}
		more, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.2.2"+keyMore)
		if !ok {
			return
		}
		last := more.Parent()
		where := geom.NewPoint(p.RectFromRoot(group.RectToRoot(group.ContentRect(true))).X+4,
			p.RectFromRoot(last.RectToRoot(last.ContentRect(true))).Bottom()-2)
		data := &rowDrag{panel: p.AsPanel(), path: "r.0"}
		p.dragOver(where, data)
		c.Equal(group, p.dropTarget, "the group is the target")
		c.Equal(dropAfter, p.dropWhere, "after it")
		p.drop(where, data)
	})
	c.Equal([]string{"weight", "None of:", "tags", "cost", "container", ":end", "name", "future_field", "unknown"},
		shape())

	// Onto the placeholder of an empty group is into it.
	screen.Do(func() { menuAction(p.treeAddEntries(f.Root, treeRootPath), "Any of Group")() })
	screen.Do(func() {
		empty, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.5:empty")
		if !ok {
			return
		}
		r := p.RectFromRoot(empty.RectToRoot(empty.ContentRect(true)))
		where := geom.NewPoint(r.CenterX(), r.CenterY())
		data := &rowDrag{panel: p.AsPanel(), path: "r.0"}
		p.dragOver(where, data)
		c.Equal(dropInto, p.dropWhere, "the placeholder takes the row into its group")
		p.drop(where, data)
	})
	c.Equal([]string{
		"None of:", "tags", "cost", "container", ":end", "name", "future_field", "unknown", "Any of:",
		"weight", ":end",
	}, shape())
}

// TestListFilterPanelPassesEscape checks the rows' own handling of Escape, which passEscape sets for the filter editor:
// it closes the open row and is taken, and with no row open, it is left for what holds the panel. The dialog closes
// the open row before the rows see the key, which TestListFilterEditorKeysHeadless covers.
func TestListFilterPanelPassesEscape(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	p, _ := showListFilterPanel(t, screen, newTestListFilter())
	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() { c.True(p.keyDown(unison.KeyEscape, mod.None, false), "Escape closes the open row") })
	screen.Do(func() {
		c.Equal("", p.open)
		c.False(p.keyDown(unison.KeyEscape, mod.None, false), "and with none open, is left for the dialog")
	})
}

// TestListFilterPanelFieldKinds switches one condition through a field of every kind and back, giving each a criterion
// that accepts less than anything first, checking that each field brings the controls its kind compares with and
// clears what the last one had. A switch between two fields of the same kind, points and levels, clears it too.
func TestListFilterPanelFieldKinds(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	setCriteria := func(cond *gurps.FilterCondition) {
		cond.Text = criteria.Text{Compare: criteria.ContainsText, Qualifier: "x"}
		cond.Number = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.FromInteger(5)}
		cond.Weight = criteria.Weight{Compare: criteria.AtMostNumber, Qualifier: fxp.WeightFromInteger(5, fxp.Pound)}
	}
	switchTo := func(p *listFilterPanel, field, control string) {
		screen.Do(func() {
			setCriteria(filterConditionAt(t, p, "r.0"))
			if popup, ok := refAs[*unison.PopupMenu[string]](t, p.AsPanel(), "r.0:field"); ok {
				popup.Select(field)
			}
		})
		screen.Do(func() {
			cond := filterConditionAt(t, p, "r.0")
			c.Equal(field, cond.Field, field)
			c.True(cond.Text.IsZero(), "%s: the text criterion is cleared", field)
			c.True(cond.Number.IsZero(), "%s: the number criterion is cleared", field)
			c.True(cond.Weight.IsZero(), "%s: the weight criterion is cleared", field)
			for _, key := range []string{"r.0:textcmp", "r.0:numbercmp", "r.0:weightcmp"} {
				c.Equal(key == control, p.FindRefKey(key) != nil, "%s: %s", field, key)
			}
		})
	}
	p, _ := showListFilterPanel(t, screen, newTestListFilter())
	screen.Do(func() { p.toggle("r.0") })
	for _, one := range []struct {
		field, control string
	}{
		{"tags", "r.0:textcmp"},
		{"cost", "r.0:numbercmp"},
		{"weight", "r.0:weightcmp"},
		{"container", ""},
		{"name", "r.0:textcmp"},
	} {
		switchTo(p, one.field, one.control)
	}

	f := gurps.NewListFilter("")
	f.Root.Children = gurps.FilterNodes{gurps.NewFilterCondition(f.Root, "points")}
	p, _ = showListFilterPanelFor(t, screen, f, filterFieldInfos(gurps.TraitFilterFields()))
	screen.Do(func() { p.toggle("r.0") })
	switchTo(p, "levels", "r.0:numbercmp")
}

// TestListFilterPanelListEditor checks what an open condition on a list offers: comparisons worded for a list, and a
// field whose tooltip says how to give several values.
func TestListFilterPanelListEditor(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := gurps.NewListFilter("")
	tags := gurps.NewFilterCondition(f.Root, "tags")
	tags.Text = criteria.Text{Compare: criteria.IsText, Qualifier: "Shield"}
	f.Root.Children = gurps.FilterNodes{tags}
	p, _ := showListFilterPanel(t, screen, f)
	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[criteria.StringComparison]](t, p.AsPanel(), "r.0:textcmp"); ok {
			items := make([]string, 0, popup.ItemCount())
			for i := range popup.ItemCount() {
				item, _ := popup.ItemAt(i)
				items = append(items, popup.ItemRendererCallback(item))
			}
			c.Equal([]string{
				"that are anything", "where at least one is", "where none is", "where at least one contains",
				"where none contains", "where at least one starts with", "where none starts with",
				"where at least one ends with", "where none ends with",
			}, items)
		}
		if field, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.0:text"); ok {
			c.Contains(tooltipText(field.Tooltip), "commas", "the field says how to give several values")
		}
	})
}

// TestListFilterPanelSavesUnknownNodes shows a filter loaded with a node of a kind this version doesn't know, changes
// the rest of it, and checks that saving writes the node back out byte for byte.
func TestListFilterPanelSavesUnknownNodes(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	const unknownNode = `{"type":"future_node","weird":[1,2]}`
	var f gurps.ListFilter
	c.NoError(jio.Unmarshal([]byte(`{"name":"Saved","root":{"type":"group","all":true,"children":[`+unknownNode+
		`,{"type":"condition","field":"tags"},{"type":"condition","field":"name"}]}}`), &f))
	p, _ := showListFilterPanel(t, screen, &f)
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.2"), "r.2"), "Move Up")() })
	screen.Do(func() {
		if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), "r"+keyPill); ok {
			pill.Select(filterAnyOf)
		}
	})
	var out []byte
	var err error
	screen.Do(func() { out, err = jio.Marshal(&f) })
	c.NoError(err)
	c.Contains(string(out), unknownNode, "the unknown node is saved as it was")
	c.Contains(string(out), `"all":false`, "along with the change")
}

// TestListFilterPanelKeepsUnknownField shows a filter with a condition on a field this version doesn't know, changes
// the rest of it, and checks that saving keeps the condition's field, its negation and its criteria.
func TestListFilterPanelKeepsUnknownField(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	f := gurps.NewListFilter("Saved")
	text := criteria.Text{Compare: criteria.IsText, Qualifier: "Sword"}
	number := criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.FromInteger(3)}
	future := gurps.NewFilterCondition(f.Root, "future_field")
	future.Not = true
	future.Text = text
	future.Number = number
	tags := gurps.NewFilterCondition(f.Root, "tags")
	name := gurps.NewFilterCondition(f.Root, "name")
	f.Root.Children = gurps.FilterNodes{future, tags, name}
	p, _ := showListFilterPanel(t, screen, f)
	screen.Do(func() {
		c.True(future == p.node("r.0"), "the future field's condition is at r.0")
		c.True(tags == p.node("r.1"), "the tags condition is at r.1")
		c.True(name == p.node("r.2"), "the name condition is at r.2")
		menuAction(p.moreEntries(p.node("r.2"), "r.2"), "Move Up")()
	})
	screen.Do(func() {
		if pill, ok := refAs[*unison.PopupMenu[filterGroupMode]](t, p.AsPanel(), "r"+keyPill); ok {
			pill.Select(filterAnyOf)
		}
	})
	screen.Do(func() { p.toggle("r.2") })
	screen.Do(func() {
		if field, ok := refAs[*unison.PopupMenu[string]](t, p.AsPanel(), "r.2:field"); ok {
			field.Select("cost")
		}
	})
	var out []byte
	var err error
	screen.Do(func() { out, err = jio.Marshal(f) })
	c.NoError(err)
	var saved gurps.ListFilter
	c.NoError(jio.Unmarshal(out, &saved))
	c.Equal([]string{"future_field", "name", "cost"}, filterShape(saved.Root), "the other changes are saved")
	c.False(saved.Root.All, "the pill's choice among them")
	if len(saved.Root.Children) == 0 {
		return
	}
	kept, ok := saved.Root.Children[0].(*gurps.FilterCondition)
	c.True(ok, "the condition on the future field is saved as a condition")
	if ok {
		c.Equal("future_field", kept.Field, "keeping its field")
		c.True(kept.Not, "its negation")
		c.Equal(text, kept.Text, "its text criterion")
		c.Equal(number, kept.Number, "and its number criterion")
	}
}

// TestListFilterPanelControlNamesDiffer checks that the controls of an open condition of each kind, and the group
// pills, have names of their own for a screen reader.
func TestListFilterPanelControlNamesDiffer(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	p, _ := showListFilterPanel(t, screen, newTestListFilter())
	var pill *unison.Panel
	var wnd *unison.Window
	screen.Do(func() {
		pill = p.FindRefKey("r.2" + keyPill)
		wnd = p.Window()
	})
	c.NotNil(screen.AccessibilityTree(wnd))
	pillNode := screen.AccessibilityNodeFor(pill)
	c.NotNil(pillNode)
	if pillNode != nil {
		c.Equal("Match", pillNode.Name, "the group pill says what it chooses")
	}
	for _, path := range []string{"r.0", "r.1", "r.2.0", "r.2.1", "r.3"} {
		screen.Do(func() { p.toggle(path) })
		names := make(map[string]bool)
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
		for _, one := range controls {
			node := screen.AccessibilityNodeFor(one)
			c.NotNil(node, path)
			if node == nil {
				continue
			}
			c.NotEqual("", node.Name, "%s: every control has a name", path)
			c.False(names[node.Name], "%s: %s is shared", path, node.Name)
			names[node.Name] = true
		}
		c.True(names["Must"], "%s: the Must popup is named for itself", path)
	}
}
