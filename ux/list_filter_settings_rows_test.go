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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSummarizeListFilter verifies the summary of a filter's conditions in the Saved Filters view, which reads a root
// that requires all of what it holds as a plain list and names any other group for how it combines what it holds.
func TestSummarizeListFilter(t *testing.T) {
	c := check.New(t)
	fields := filterFieldInfos(gurps.TraitFilterFields(), nil)
	f := gurps.NewListFilter("x")
	c.Equal("No conditions", summarizeListFilter(f, fields))
	cond := gurps.NewFilterCondition(f.Root, "tags")
	cond.Text.Compare = criteria.IsText
	cond.Text.Qualifier = "Mental"
	f.Root.Children = append(f.Root.Children, cond, gurps.NewFilterGroup(f.Root))
	c.Equal("Must have tags where at least one is Mental; All of (No conditions)", summarizeListFilter(f, fields))
	f.Root.All = false
	c.Equal("Any of (Must have tags where at least one is Mental; All of (No conditions))",
		summarizeListFilter(f, fields))
	f.Root.Not = true
	c.Equal("None of (Must have tags where at least one is Mental; All of (No conditions))",
		summarizeListFilter(f, fields), "a root that requires none of what it holds")
	f.Root.All = true
	c.Equal("Not all of (Must have tags where at least one is Mental; All of (No conditions))",
		summarizeListFilter(f, fields), "a root that requires all of what it holds, negated, is no plain list")
}
