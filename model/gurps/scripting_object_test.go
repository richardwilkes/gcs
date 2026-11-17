// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"errors"
	"fmt"
	"testing"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// The members Object.prototype supplies must be reachable from the objects handed to scripts. goja consults a dynamic
// object's prototype only when its Get returns nil, so a Get that answered goja.Undefined() for keys the object does
// not provide shadowed Object.prototype entirely: hasOwnProperty, valueOf and constructor all read as undefined, and
// calling any of them threw a TypeError.
func TestScriptObjectPrototypeIsReachable(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Sample"
	e.Traits = []*Trait{trait}
	e.Recalculate()
	selfArg := ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptTrait(r, trait) }}
	for _, tc := range []struct {
		name   string
		script string
		want   string
	}{
		{name: "hasOwnProperty exists", script: "typeof self.hasOwnProperty", want: "function"},
		{name: "hasOwnProperty for a provided key", script: "String(self.hasOwnProperty('name'))", want: "true"},
		{name: "hasOwnProperty for a missing key", script: "String(self.hasOwnProperty('notAKey'))", want: "false"},
		{name: "constructor", script: "String(self.constructor === Object)", want: "true"},
		{name: "isPrototypeOf", script: "String(Object.prototype.isPrototypeOf(self))", want: "true"},
		{name: "prototype toString", script: "Object.prototype.toString.call(self)", want: "[object Object]"},
		{name: "in for a provided key", script: "String('name' in self)", want: "true"},
		{name: "in for a missing key", script: "String('notAKey' in self)", want: "false"},
		// Reading a key the object does not provide must still come out as undefined rather than picking up something
		// unexpected from the prototype.
		{name: "missing key", script: "typeof self.notAKey", want: "undefined"},
		{name: "provided key", script: "self.name", want: "Sample"},
	} {
		v, err := runScript(0, tc.script, selfArg)
		c.NoError(err, "case %q", tc.name)
		c.Equal(tc.want, v, "case %q", tc.name)
	}
}

// An object that describes itself as a value through valueOf — the attributes bound as $st, $dx and the rest — must
// both answer toString and still turn into that value when converted to a string, which is how a script's result
// reaches the rest of GCS. The two pull against each other: JavaScript consults toString before valueOf, so an object
// with no toString of its own would inherit Object.prototype.toString and convert to "[object Object]", while an object
// with no toString at all leaves `$st.toString()` undefined and a TypeError when called.
func TestScriptObjectWithValueOfConvertsToItsValue(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Recalculate()
	attr := e.Attributes.Find("st")
	c.NotNil(attr)
	attrArg := ScriptArg{Name: "$st", Value: func(r *goja.Runtime) any { return newScriptAttribute(r, attr) }}

	// A new entity's ST is 10, so this is what every conversion below has to produce.
	value, err := runScript(0, "String($st.valueOf())", attrArg)
	c.NoError(err)
	c.Equal("10", value)

	for _, tc := range []struct {
		name   string
		script string
		want   string
	}{
		{name: "toString exists", script: "typeof $st.toString", want: "function"},
		{name: "toString reports the value", script: "$st.toString()", want: value},
		{name: "explicit string conversion", script: "String($st)", want: value},
		{name: "concatenation", script: "'' + $st", want: value},
		// The bare attribute is the shape the sheet actually resolves; runScript delivers its result by converting it
		// to a string, so this is what ResolveToNumber has to be able to parse.
		{name: "result conversion", script: "$st", want: value},
	} {
		var v string
		v, err = runScript(0, tc.script, attrArg)
		c.NoError(err, "case %q", tc.name)
		c.Equal(tc.want, v, "case %q", tc.name)
	}
}

// TestScriptObjectToString verifies that a script object keeps a toString of its own. The one it is given otherwise
// reports "[object Object]" for an object with no valueOf, and lets an exception valueOf throws reach the script.
func TestScriptObjectToString(t *testing.T) {
	c := check.New(t)
	fn := func(r *goja.Runtime, f func() goja.Value) func() goja.Value {
		return func() goja.Value { return r.ToValue(func(goja.FunctionCall) goja.Value { return f() }) }
	}
	for _, tc := range []struct {
		name  string
		kvMap func(r *goja.Runtime) map[string]func() goja.Value
		want  string
		err   string
	}{
		{
			name: "own toString",
			kvMap: func(r *goja.Runtime) map[string]func() goja.Value {
				return map[string]func() goja.Value{
					"toString": fn(r, func() goja.Value { return r.ToValue("own") }),
					"valueOf":  fn(r, func() goja.Value { return r.ToValue(7) }),
				}
			},
			want: "own",
		},
		{
			name: "no valueOf",
			kvMap: func(r *goja.Runtime) map[string]func() goja.Value {
				return map[string]func() goja.Value{"name": func() goja.Value { return r.ToValue("x") }}
			},
			want: "[object Object]",
		},
		{
			name: "valueOf throws",
			kvMap: func(r *goja.Runtime) map[string]func() goja.Value {
				return map[string]func() goja.Value{
					"valueOf": fn(r, func() goja.Value { panic(r.NewTypeError("boom")) }),
				}
			},
			err: "boom",
		},
	} {
		var v string
		var err error
		c.NotPanics(func() {
			v, err = runScript(0, "String(obj)", ScriptArg{
				Name:  "obj",
				Value: func(r *goja.Runtime) any { return r.NewDynamicObject(NewScriptObject(r, tc.kvMap(r))) },
			})
		}, "case %q", tc.name)
		if tc.err != "" {
			ex, ok := errors.AsType[*goja.Exception](err)
			c.True(ok, "case %q: expected the script exception, got %v", tc.name, err)
			if ok {
				c.Contains(ex.Error(), tc.err, "case %q", tc.name)
			}
			continue
		}
		c.NoError(err, "case %q", tc.name)
		c.Equal(tc.want, v, "case %q", tc.name)
	}
}

// TestScriptObjectIsReadOnly verifies that a script can list the properties of a script object but not assign or
// delete any.
func TestScriptObjectIsReadOnly(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.Name = "Sample"
	selfArg := ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptTrait(r, trait) }}
	for _, tc := range []struct {
		name   string
		script string
		want   string
	}{
		{name: "keys", script: "String(Object.keys(self).includes('name'))", want: "true"},
		{
			name:   "assign",
			script: "try { self.name = 'x'; 'assigned' } catch (e) { e.name + ' ' + self.name }",
			want:   "TypeError Sample",
		},
		{
			name:   "delete",
			script: "try { delete self.name; 'deleted' } catch (e) { e.name + ' ' + self.name }",
			want:   "TypeError Sample",
		},
	} {
		v, err := runScript(0, tc.script, selfArg)
		c.NoError(err, "case %q", tc.name)
		c.Equal(tc.want, v, "case %q", tc.name)
	}
}

// The id and parentID properties every script wrapper exposes must arrive in JavaScript as primitive strings. Handing
// goja a tid.TID (a named string type) rather than a plain string makes it build a String object instead: typeof
// reports "object", === against a string literal is always false, and Array.prototype.includes — which compares with
// SameValueZero — never matches, so the ids are unusable for the comparisons scripts need to make.
func TestScriptIDsArePrimitiveStrings(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, true)
	childTrait := NewTrait(e, trait, false)
	trait.Children = []*Trait{childTrait}
	traitMod := NewTraitModifier(e, nil, false)

	skill := NewSkill(e, nil, true)
	childSkill := NewSkill(e, skill, false)
	skill.Children = []*Skill{childSkill}

	spell := NewSpell(e, nil, true)
	childSpell := NewSpell(e, spell, false)
	spell.Children = []*Spell{childSpell}

	equipment := NewEquipment(e, nil, true)
	childEquipment := NewEquipment(e, equipment, false)
	equipment.Children = []*Equipment{childEquipment}
	equipmentMod := NewEquipmentModifier(e, nil, false)

	note := NewNote(e, nil, true)
	childNote := NewNote(e, note, false)
	note.Children = []*Note{childNote}

	weapon := NewWeapon(equipment, true)

	for _, tc := range []struct {
		name     string
		provider ScriptSelfProvider
		expr     string // yields the id under test
		want     tid.TID
	}{
		{name: "trait id", provider: deferredNewScriptTrait(trait), expr: "self.id", want: trait.TID},
		{name: "trait parentID", provider: deferredNewScriptTrait(childTrait), expr: "self.parentID", want: trait.TID},
		{name: "trait modifier id", provider: deferredNewScriptTraitModifier(traitMod), expr: "self.id", want: traitMod.TID},
		{name: "skill id", provider: deferredNewScriptSkill(skill), expr: "self.id", want: skill.TID},
		{name: "skill parentID", provider: deferredNewScriptSkill(childSkill), expr: "self.parentID", want: skill.TID},
		{name: "spell id", provider: deferredNewScriptSpell(spell), expr: "self.id", want: spell.TID},
		{name: "spell parentID", provider: deferredNewScriptSpell(childSpell), expr: "self.parentID", want: spell.TID},
		{name: "equipment id", provider: deferredNewScriptEquipment(equipment), expr: "self.id", want: equipment.TID},
		{
			name:     "equipment parentID",
			provider: deferredNewScriptEquipment(childEquipment),
			expr:     "self.parentID",
			want:     equipment.TID,
		},
		{
			name:     "equipment modifier id",
			provider: deferredNewScriptEquipmentModifier(equipmentMod),
			expr:     "self.id",
			want:     equipmentMod.TID,
		},
		{name: "note id", provider: deferredNewScriptNote(note), expr: "self.id", want: note.TID},
		{name: "note parentID", provider: deferredNewScriptNote(childNote), expr: "self.parentID", want: note.TID},
		{name: "weapon id", provider: deferredNewScriptWeapon(weapon), expr: "self.id", want: weapon.TID},
	} {
		// The value must be a primitive string, not a String object.
		c.Equal("string", ResolveScript(e, tc.provider, "typeof "+tc.expr), "case %q", tc.name)

		c.Equal(string(tc.want), ResolveScript(e, tc.provider, tc.expr), "case %q", tc.name)

		// It must also compare equal to that id with the operators scripts use. A String object fails all of these.
		for _, cmp := range []struct {
			label  string
			script string
		}{
			{label: "===", script: fmt.Sprintf("%s === %q", tc.expr, tc.want)},
			{label: "includes", script: fmt.Sprintf("[%q].includes(%s)", tc.want, tc.expr)},
			{label: "indexOf", script: fmt.Sprintf("[%q].indexOf(%s) === 0", tc.want, tc.expr)},
			{label: "object key", script: fmt.Sprintf("({%q: 1})[%s] === 1", tc.want, tc.expr)},
		} {
			c.Equal("true", ResolveScript(e, tc.provider, cmp.script), "case %q via %s", tc.name, cmp.label)
		}
	}

	// A missing parent still reports undefined rather than an id.
	c.Equal("undefined", ResolveScript(e, deferredNewScriptTrait(trait), "typeof self.parentID"))
	c.Equal("undefined", ResolveScript(e, deferredNewScriptEquipment(equipment), "typeof self.parentID"))
	c.Equal("undefined", ResolveScript(e, deferredNewScriptNote(note), "typeof self.parentID"))
}

// Covers the properties every node script wrapper shares -- parent, container and tags -- both as a top-level node and
// as a child. parent must be the wrapper for the actual parent, reachable recursively up the chain, and undefined at
// the top, and a parent lists the child among its children; container must follow the node's kind; and tags must be a
// copy, so a script cannot alter the node.
func TestScriptNodeIdentityProperties(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, true)
	childTrait := NewTrait(e, trait, false)
	trait.Children = []*Trait{childTrait}
	trait.Tags = []string{"Mental", "Physical"}

	skill := NewSkill(e, nil, true)
	childSkill := NewSkill(e, skill, false)
	skill.Children = []*Skill{childSkill}
	skill.Tags = []string{"Combat"}

	spell := NewSpell(e, nil, true)
	childSpell := NewSpell(e, spell, false)
	spell.Children = []*Spell{childSpell}
	spell.Tags = []string{"Fire", "Air"}

	equipment := NewEquipment(e, nil, true)
	childEquipment := NewEquipment(e, equipment, false)
	equipment.Children = []*Equipment{childEquipment}
	equipment.Tags = []string{"Gear"}

	note := NewNote(e, nil, true)
	childNote := NewNote(e, note, false)
	note.Children = []*Note{childNote}
	note.Tags = []string{"Reminder"}

	for _, tc := range []struct {
		name     string
		top      ScriptSelfProvider
		child    ScriptSelfProvider
		wantTags string
	}{
		{
			name:     "trait",
			top:      deferredNewScriptTrait(trait),
			child:    deferredNewScriptTrait(childTrait),
			wantTags: "Mental,Physical",
		},
		{
			name:     "skill",
			top:      deferredNewScriptSkill(skill),
			child:    deferredNewScriptSkill(childSkill),
			wantTags: "Combat",
		},
		{
			name:     "spell",
			top:      deferredNewScriptSpell(spell),
			child:    deferredNewScriptSpell(childSpell),
			wantTags: "Fire,Air",
		},
		{
			name:     "equipment",
			top:      deferredNewScriptEquipment(equipment),
			child:    deferredNewScriptEquipment(childEquipment),
			wantTags: "Gear",
		},
		{
			name:     "note",
			top:      deferredNewScriptNote(note),
			child:    deferredNewScriptNote(childNote),
			wantTags: "Reminder",
		},
	} {
		c.Equal("undefined", ResolveScript(e, tc.top, "typeof self.parent"), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, tc.top, "self.container"), "case %q", tc.name)
		c.Equal(tc.wantTags, ResolveScript(e, tc.top, "self.tags.join(',')"), "case %q", tc.name)

		// A child reaches its parent's wrapper through parent, and the chain ends there.
		c.Equal("false", ResolveScript(e, tc.child, "self.container"), "case %q", tc.name)
		c.Equal("", ResolveScript(e, tc.child, "self.tags.join(',')"), "case %q", tc.name)
		c.Equal(tc.top.ID, ResolveScript(e, tc.child, "self.parent.id"), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, tc.child, "self.parent.id === self.parentID"), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, tc.child, "self.parent.container"), "case %q", tc.name)
		c.Equal(tc.wantTags, ResolveScript(e, tc.child, "self.parent.tags.join(',')"), "case %q", tc.name)
		c.Equal("undefined", ResolveScript(e, tc.child, "typeof self.parent.parent"), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, tc.child, "self.parent.children.some(n => n.id === self.id)"), "case %q",
			tc.name)

		// tags is a copy: a script that pushes onto it does not alter the node.
		c.Equal("1", ResolveScript(e, tc.child, "self.tags.push('x')"), "case %q", tc.name)
		c.Equal("", ResolveScript(e, tc.child, "self.tags.join(',')"), "case %q", tc.name)
	}
}

// The notes property every wrapper exposes yields the node's secondary text even when the sheet is configured not to
// display notes at all, since a script wants the content, not the layout decision.
func TestScriptNotesReportAllSecondaryText(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.NotesDisplay = display.NotShown

	trait := NewTrait(e, nil, false)
	trait.LocalNotes = "trait notes"
	traitMod := NewTraitModifier(e, nil, false)
	traitMod.LocalNotes = "trait modifier notes"
	skill := NewSkill(e, nil, false)
	skill.LocalNotes = "skill notes"
	spell := NewSpell(e, nil, false)
	spell.LocalNotes = "spell notes"
	equipment := NewEquipment(e, nil, false)
	equipment.LocalNotes = "equipment notes"
	equipmentMod := NewEquipmentModifier(e, nil, false)
	equipmentMod.LocalNotes = "equipment modifier notes"

	for _, tc := range []struct {
		name     string
		provider ScriptSelfProvider
		want     string
	}{
		{name: "trait", provider: deferredNewScriptTrait(trait), want: "trait notes"},
		{name: "trait modifier", provider: deferredNewScriptTraitModifier(traitMod), want: "trait modifier notes"},
		{name: "skill", provider: deferredNewScriptSkill(skill), want: "skill notes"},
		{name: "spell", provider: deferredNewScriptSpell(spell), want: "spell notes"},
		{name: "equipment", provider: deferredNewScriptEquipment(equipment), want: "equipment notes"},
		{
			name:     "equipment modifier",
			provider: deferredNewScriptEquipmentModifier(equipmentMod),
			want:     "equipment modifier notes",
		},
	} {
		c.Equal("true", ResolveScript(e, tc.provider, fmt.Sprintf("self.notes.includes(%q)", tc.want)), "case %q",
			tc.name)
	}
}
