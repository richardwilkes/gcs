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
	return i18n.Text("A custom check")
}

// Satisfied implements Prereq. An unmet prerequisite writes its description, followed by the reason it gives, if any.
// A reason of more than one line goes on the lines below it, one level deeper.
func (s *ScriptPrereq) Satisfied(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	result, _ := s.evaluate(entity, exclude, tooltip, prefix)
	return result == CheckMet
}

// evaluate is Satisfied, returning the script's result and the reason it gives.
func (s *ScriptPrereq) evaluate(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string) (result CheckResult, reason string) {
	result, reason = s.Evaluate(entity, exclude)
	reason = strings.TrimSpace(reason)
	if result == CheckMet || tooltip == nil {
		return result, reason
	}
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	name := s.Describe(entity, replacements, plainText)
	multiLine := strings.Contains(reason, "\n")
	tooltip.WriteString(prefix)
	switch {
	case reason == "":
		tooltip.WriteString(name)
	case result == CheckFailed && multiLine:
		fmt.Fprintf(tooltip, i18n.Text("%s (couldn't run):"), name)
	case result == CheckFailed:
		fmt.Fprintf(tooltip, i18n.Text("%s (couldn't run: %s)"), name, reason)
	case multiLine:
		fmt.Fprintf(tooltip, i18n.Text("%s:"), name)
	default:
		fmt.Fprintf(tooltip, i18n.Text("%s (%s)"), name, reason)
	}
	if multiLine {
		nested := strings.ReplaceAll(prefix, "\n", "\n\t")
		for line := range strings.SplitSeq(reason, "\n") {
			tooltip.WriteString(nested)
			tooltip.WriteString(strings.TrimRight(line, "\r"))
		}
	}
	return result, reason
}

// Evaluate runs the script against the entity for the item given as exclude. A result of "" or "true" is met, "false"
// is unmet with no reason, and any other result is unmet, with that text as the reason. It has failed when the script
// could not produce a result, because it threw, timed out or nested too deeply; the reason is then the error.
func (s *ScriptPrereq) Evaluate(entity *Entity, exclude any) (result CheckResult, reason string) {
	script := s.Script
	if na, ok := exclude.(nameable.Accesser); ok {
		script = nameable.Apply(script, na.NameableReplacements())
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
	text, failed := resolveText(entity, self, script)
	switch {
	case failed:
		return CheckFailed, text
	case text == "" || text == "true":
		return CheckMet, ""
	case text == "false":
		return CheckUnmet, ""
	default:
		return CheckUnmet, text
	}
}
