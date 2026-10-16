// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/check"
)

// wideHost returns a host whose self has n kids.
func wideHost(n int) (h *fakeHost, self *fakeItem) {
	kids := make([]*fakeItem, n)
	for i := range kids {
		kids[i] = newItem("kid "+strconv.Itoa(i), nil)
	}
	return newHost(), newItem("self", nil).with("kids", kids...)
}

// expensive is a count over every kid that needs them all, costing a few units of work per kid.
func expensive() *Count {
	return &Count{
		Of: "kids", Where: Ref{Node: checkText("name", criteria.StartsWithText, "kid")},
		Quantity: atLeast(1000),
	}
}

// TestBudget checks that a pass that runs out of budget ends Failed where its state depends on what it couldn't do,
// keeps a state that what it did decide, and that a Collect pass that runs out gives way to the YesNo state, marked
// Truncated and with no details.
func TestBudget(t *testing.T) {
	c := check.New(t)
	h, self := wideHost(100)
	small := Options{Budget: 50}

	r := eval(t, expensive(), h, self, small)
	c.Equal(Failed, r.State)
	c.Equal("too complex to decide", r.Reason)
	c.Equal(Unmet, eval(t, expensive(), h, self, Options{}).State)

	decided := &Group{Mode: GroupAny, Children: Nodes{&Const{Value: true}, expensive()}}
	c.Equal(Met, eval(t, decided, h, self, small).State)
	small.Mode = Collect
	r = eval(t, decided, h, self, small)
	c.Equal(Met, r.State)
	c.True(r.Truncated)
	c.Equal(0, len(r.Children))
	r = eval(t, decided, h, self, collect)
	c.False(r.Truncated)
	c.Equal(2, len(r.Children))

	unmet := &Group{Children: Nodes{&Const{}, expensive()}}
	small.Mode = Explain
	r = eval(t, unmet, h, self, small)
	c.Equal(Unmet, r.State)
	c.True(r.Truncated)
	r = eval(t, unmet, h, self, explain)
	c.Equal(Unmet, r.State)
	c.False(r.Truncated)
	c.Equal(2, len(r.Children))
	named := checkText("name", criteria.IsText, "self")
	c.Equal(1, eval(t, named, h, self, collect).Specificity)
	r = eval(t, named, h, self, explain)
	c.Equal(Met, r.State)
	c.Equal(0, r.Specificity, "a state that passes isn't explained")
}

// cancellingHost cancels its context once it has been asked for a field a given number of times.
type cancellingHost struct {
	*fakeHost
	cancel func()
	after  int
}

func (h *cancellingHost) Field(item Item, key string) (Value, bool) {
	h.after--
	if h.after == 0 {
		h.cancel()
	}
	return h.fakeHost.Field(item, key)
}

// TestCancel checks that an evaluation whose context is done, before it starts, while it works through a large
// collection, or while a script waits, returns the context's error instead of a result.
func TestCancel(t *testing.T) {
	c := check.New(t)
	h, self := wideHost(10_000)
	done, cancel := context.WithCancel(t.Context())
	cancel()
	r, err := Evaluate(done, &Const{Value: true}, h, self, Options{})
	c.True(errors.Is(err, context.Canceled))
	c.Equal(Result{}, r)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	host := &cancellingHost{fakeHost: h, cancel: cancel, after: 100}
	_, err = Evaluate(ctx, expensive(), host, self, Options{Budget: 1 << 30})
	c.True(errors.Is(err, context.Canceled))
	c.True(host.after > -cancelCheckInterval, "noticed within a check interval")

	ctx, cancel = context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	_, err = Evaluate(ctx, &Group{Children: Nodes{&Script{Script: "block"}}}, h, self, collect)
	c.True(errors.Is(err, context.DeadlineExceeded))

	r, err = Evaluate(t.Context(), &Const{Value: true}, h, self, Options{})
	c.NoError(err)
	c.Equal(Met, r.State)
	r, err = Evaluate(t.Context(), nil, h, self, Options{})
	c.NoError(err)
	c.Equal(Skipped, r.State)
}

// TestStrings checks the names states and value kinds give themselves in reasons.
func TestStrings(t *testing.T) {
	c := check.New(t)
	c.Equal([]string{"failed", "met", "unmet", "skipped", "?"}, []string{
		Failed.String(), Met.String(), Unmet.String(), Skipped.String(), State(9).String(),
	})
	c.Equal([]string{"absent", "text", "list", "number", "weight", "bool", "unknown"}, []string{
		ValueAbsent.String(), ValueText.String(), ValueList.String(), ValueNumber.String(), ValueWeight.String(),
		ValueBool.String(), ValueKind(9).String(),
	})
}
