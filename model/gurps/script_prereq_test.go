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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestScriptPrereqEvaluate verifies how a script's result maps to met, the reason and the failed state, and what an
// unmet one writes.
func TestScriptPrereqEvaluate(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	for _, one := range []struct {
		name, script, reason, text string
		met, failed                bool
	}{
		{name: "Big", script: "", met: true},
		{name: "Big", script: "true", met: true},
		{name: "Big", script: "1 < 2", met: true},
		{name: "Big", script: "false", text: "Big"},
		{script: "false", text: "A custom check"},
		{name: "Big", script: `"Too weak"`, reason: "Too weak", text: "Big (Too weak)"},
		{name: "Big", script: `"Needs\n* more"`, reason: "Needs\n* more", text: "Big:\n\t- Needs\n\t- * more"},
		{name: "Big", script: `throw new Error("boom")`, failed: true, text: "Big (couldn't run: "},
		{script: `throw new Error("boom")`, failed: true, text: "A custom check (couldn't run: "},
		{
			name: "Big", script: "<script>1 < 2</script> or <script>false</script>", reason: "true or false",
			text: "Big (true or false)",
		},
	} {
		p := NewScriptPrereq()
		p.Name = one.name
		p.Script = one.script
		result, reason := p.Evaluate(entity, nil)
		met, failed := result == CheckMet, result == CheckFailed
		c.Equal(one.met, met, one.script)
		c.Equal(one.failed, failed, one.script)
		if failed {
			c.True(strings.Contains(reason, "boom"), "%s: %q", one.script, reason)
		} else {
			c.Equal(one.reason, reason, one.script)
		}
		var tooltip xbytes.InsertBuffer
		c.Equal(met, p.Satisfied(entity, nil, &tooltip, "\n- ", nil), one.script)
		if !met {
			c.True(strings.HasPrefix(tooltip.String(), "\n- "+one.text), "%s: %q", one.script, tooltip.String())
		}
	}
}

// TestScriptPrereqStoppedRunsOncePerRecalculation verifies that a script prerequisite stopped for running too long is
// not run again before the next recalculation, whether or not its item carries a grant: judged again in the meantime,
// it is still reported as stopped, though it would now answer at once.
func TestScriptPrereqStoppedRunsOncePerRecalculation(t *testing.T) {
	// The time limit is lowered only for the recalculation in which the script is meant to run out of it, since a
	// legitimate script has run out of so short a limit on a slow runner, and nothing checked depends on whether one
	// does there.
	recalculateWithShortLimit := func(e *Entity) {
		prev := scriptExecTimeLimitOverride.Load()
		SetScriptExecTimeLimitForTesting(fxp.FromStringForced("0.01"))
		defer scriptExecTimeLimitOverride.Store(prev)
		SuppressScriptResolveErrorLogging(e.Recalculate)
	}
	c := check.New(t)
	for _, carriesGrant := range []bool{false, true} {
		e := NewEntity()
		training := addTraitWithFeatures(e, "Training")
		if carriesGrant {
			training.Features = Features{
				newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore),
			}
		}
		training.Prereq = NewPrereqList()
		addScriptPrereq(training.Prereq, `entity.hasTrait("Ready") || (() => { for (;;); })()`)
		recalculateWithShortLimit(e)
		c.NotEqual("", training.UnsatisfiedReason, "carries a grant %v: the script is stopped", carriesGrant)
		addTraitWithFeatures(e, "Ready")
		stopped := abandonedScripts(e)
		var result CheckResult
		SuppressScriptResolveErrorLogging(func() { result = training.Prereq.Evaluate(e, training, nil) })
		c.Equal(CheckFailed, result, "carries a grant %v: judged again, it is not run again", carriesGrant)
		c.Equal(int64(1), abandonedScripts(e)-stopped, "carries a grant %v: and is reported as stopped", carriesGrant)
		SuppressScriptResolveErrorLogging(e.Recalculate)
		c.Equal("", training.UnsatisfiedReason, "carries a grant %v: the next recalculation runs it again", carriesGrant)
	}
}

// TestScriptPrereqName verifies that the name describes the prerequisite, takes nameable replacements, offers its
// markers for filling in and is part of the hash.
func TestScriptPrereqName(t *testing.T) {
	c := check.New(t)
	p := NewScriptPrereq()
	c.Equal("A custom check", p.Describe(nil, nil, plainText))
	p.Name = "Knows @Lore@"
	c.Equal("Knows Demons", p.Describe(nil, map[string]string{"Lore": "Demons"}, plainText))
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

// TestResolveScriptSyntaxErrorNamedOnce verifies that a script that won't compile is reported as a SyntaxError once.
func TestResolveScriptSyntaxErrorNamedOnce(t *testing.T) {
	c := check.New(t)
	result, failed := resolveScript(NewEntity(), ScriptSelfProvider{}, "nope(")
	c.True(failed)
	c.True(strings.HasPrefix(result, "SyntaxError: ") && !strings.Contains(result, "SyntaxError: SyntaxError:"), result)
}
