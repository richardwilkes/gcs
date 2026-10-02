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
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestScriptPrereqEvaluate verifies how a script's result maps to met, the reason and the failed state.
func TestScriptPrereqEvaluate(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	for _, one := range []struct {
		name, script, reason, failedPrefix string
		met                                bool
	}{
		{name: "Big", script: "", met: true},
		{name: "Big", script: "true", met: true},
		{name: "Big", script: "1 < 2", met: true},
		{name: "Big", script: "false", reason: "Big"},
		{script: "false", reason: "Passes a custom check"},
		{name: "Big", script: `"Needs\n* more"`, reason: "Needs\n* more"},
		{name: "Big", script: `throw new Error("boom")`, failedPrefix: `Couldn't check "Big": `},
		{script: `throw new Error("boom")`, failedPrefix: "Couldn't run a custom check: "},
		{name: "Big", script: "<script>1 < 2</script> or <script>false</script>", reason: "true or false"},
	} {
		p := NewScriptPrereq()
		p.Name = one.name
		p.Script = one.script
		met, reason, failed := p.Evaluate(entity, nil)
		c.Equal(one.met, met, one.script)
		c.Equal(one.failedPrefix != "", failed, one.script)
		if failed {
			c.True(strings.HasPrefix(reason, one.failedPrefix), "%s: %q", one.script, reason)
			c.True(strings.Contains(reason, "boom"), "%s: %q", one.script, reason)
		} else {
			c.Equal(one.reason, reason, one.script)
		}
		var tooltip xbytes.InsertBuffer
		c.Equal(met, p.Satisfied(entity, nil, &tooltip, "\n- ", nil), one.script)
		if !met {
			c.Equal("\n- "+reason, tooltip.String(), one.script)
		}
	}
}

// TestScriptPrereqName verifies that the name describes the prerequisite, takes nameable replacements, offers its
// markers for filling in and is part of the hash.
func TestScriptPrereqName(t *testing.T) {
	c := check.New(t)
	p := NewScriptPrereq()
	c.Equal("Passes a custom check", p.Describe(nil, plainText))
	p.Name = "Knows @Lore@"
	c.Equal("Knows Demons", p.Describe(map[string]string{"Lore": "Demons"}, plainText))
	keys := make(map[string]string)
	p.FillWithNameableKeys(keys, nil)
	_, exists := keys["Lore"]
	c.True(exists, "the name's marker must be offered for filling in")
	hash := func() []byte {
		h := sha256.New()
		p.Hash(h)
		return h.Sum(nil)
	}
	before := hash()
	p.Name = "Other"
	c.NotEqual(before, hash(), "the name must be part of the hash")
}

// TestResolveScriptCachedErrorStaysFailed verifies that a script error is still reported as a failure when the result
// comes from the cache, rather than being taken for the script's own result.
func TestResolveScriptCachedErrorStaysFailed(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	text := `throw new Error("boom")`
	first, failed := resolveScript(entity, ScriptSelfProvider{}, text)
	c.True(failed, "a thrown error is a failure")
	c.True(entity.scriptCache[scriptResolveKey{text: text}].err, "the failure must be cached")
	second, failed := resolveScript(entity, ScriptSelfProvider{}, text)
	c.True(failed, "a cached error is still a failure")
	c.Equal(first, second)
	c.Equal(first, ResolveScript(entity, ScriptSelfProvider{}, text), "ResolveScript returns the same text")
	result, failed := resolveScript(entity, ScriptSelfProvider{}, "1 + 1")
	c.False(failed, "a result is not a failure")
	c.Equal("2", result)
	_, failed = resolveScript(entity, ScriptSelfProvider{}, "1 + 1")
	c.False(failed, "a cached result is not a failure")
}
