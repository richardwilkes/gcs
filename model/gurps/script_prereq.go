// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &ScriptPrereq{}

// ScriptPrereq represents a script-based prerequisite.
type ScriptPrereq struct {
	Parent *PrereqList `json:"-"`
	Type   prereq.Type `json:"type"`
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
	xhash.StringWithLen(h, s.Script)
}

// FillWithNameableKeys implements Prereq.
func (s *ScriptPrereq) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(s.Script, m, existing)
}

// Satisfied implements Prereq.
func (s *ScriptPrereq) Satisfied(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	script := s.Script
	if na, ok := exclude.(nameable.Accesser); ok {
		script = nameable.Apply(script, na.NameableReplacements())
	}
	script = strings.TrimSpace(script)
	if script != "" && !strings.HasPrefix(script, "<script>") {
		script = "<script>" + script + "</script>"
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
	if result := ResolveText(entity, self, script); result != "" {
		if tooltip != nil {
			fmt.Fprintf(tooltip, "%s%s", prefix, result)
		}
		return false
	}
	return true
}
