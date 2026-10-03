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
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &ScriptPrereq{}

// ScriptPrereq represents a script-based prerequisite.
type ScriptPrereq struct {
	Parent *PrereqList `json:"-"`
	Type   prereq.Type `json:"type"`
	Name   string      `json:"name,omitzero"`
	Script string      `json:"script"`
}

// NewScriptPrereq creates a new ScriptPrereq.
func NewScriptPrereq() *ScriptPrereq {
	var s ScriptPrereq
	s.Type = prereq.Script
	return &s
}

// PrereqType implements Prereq.
func (s *ScriptPrereq) PrereqType() prereq.Type {
	return s.Type
}

// ParentList implements Prereq.
func (s *ScriptPrereq) ParentList() *PrereqList {
	return s.Parent
}

// SetParentList implements Prereq.
func (s *ScriptPrereq) SetParentList(list *PrereqList) {
	s.Parent = list
}

// Clone implements Prereq.
func (s *ScriptPrereq) Clone(parent *PrereqList) Prereq {
	clone := *s
	clone.Parent = parent
	return &clone
}

// Hash implements Prereq.
func (s *ScriptPrereq) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, s.Type)
	xhash.StringWithLen(h, s.Name)
	xhash.StringWithLen(h, s.Script)
}

// FillWithNameableKeys implements Prereq.
func (s *ScriptPrereq) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, s.Name, s.Script)
}

// ResolvedName returns the name with the replacements applied and the surrounding space trimmed.
func (s *ScriptPrereq) ResolvedName(replacements map[string]string) string {
	return strings.TrimSpace(nameable.Apply(s.Name, replacements))
}

// Describe implements Prereq. It returns the name, or a generic description when there is none.
func (s *ScriptPrereq) Describe(_ *Entity, replacements map[string]string, _ func(string) string) string {
	if name := s.ResolvedName(replacements); name != "" {
		return name
	}
	return i18n.Text("Passes a custom check")
}

// Satisfied implements Prereq.
func (s *ScriptPrereq) Satisfied(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	met, reason, _ := s.Evaluate(entity, exclude)
	if !met && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(reason)
	}
	return met
}

// Evaluate runs the script against the entity for the item given as exclude. A result of "" or "true" is met. A result
// of "false" is unmet, with the reason saying this prerequisite's description failed, and any other result is unmet,
// with that text as the reason. failed is true when the script could not produce a result, because it threw, timed out
// or nested too deeply; the reason then says so.
func (s *ScriptPrereq) Evaluate(entity *Entity, exclude any) (met bool, reason string, failed bool) {
	script := s.Script
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
		script = nameable.Apply(script, replacements)
	}
	script = strings.TrimSpace(script)
	if script != "" && !strings.HasPrefix(script, scriptStart) {
		script = scriptStart + script + scriptEnd
	}
	var self ScriptSelfProvider
	switch what := exclude.(type) {
	case *Equipment:
		self = deferredNewScriptEquipment(what)
	case *Skill:
		self = deferredNewScriptSkill(what)
	case *Spell:
		self = deferredNewScriptSpell(what)
	case *Trait:
		self = deferredNewScriptTrait(what)
	}
	result, failed := resolveText(entity, self, script)
	switch {
	case failed:
		if name := s.ResolvedName(replacements); name != "" {
			return false, fmt.Sprintf(i18n.Text(`Couldn't check "%s": %s`), name, result), true
		}
		return false, fmt.Sprintf(i18n.Text("Couldn't run a custom check: %s"), result), true
	case result == "" || result == "true":
		return true, "", false
	case result == "false":
		return false, fmt.Sprintf(i18n.Text("Failed: %s"), s.Describe(entity, replacements, plainText)), false
	default:
		return false, result, false
	}
}
