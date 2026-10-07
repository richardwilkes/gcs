// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package themeset

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// entry is a themed int whose live value sits behind a pointer, like the ThemeColors and IndirectFonts the real
// entries wrap.
type entry struct {
	id    string
	value *int
}

func (e *entry) Key() string           { return e.id }
func (e *entry) Value() int            { return *e.value }
func (e *entry) SetValue(v int)        { *e.value = v }
func newEntry(id string, v int) *entry { return &entry{id: id, value: &v} }

// The provider has to be a zero-size type, so the entries it hands out are package state that each test resets.
var (
	live    []*entry
	factory []*entry
	applied int
)

type testProvider struct{}

func (testProvider) Current() []*entry { return live }
func (testProvider) Factory() []*entry { return factory }
func (testProvider) Applied()          { applied++ }

type testSet = Set[int, *entry, testProvider]

// setup installs three entries, "a", "b" and "c", whose factory values are 1, 2 and 3 and whose live values have been
// edited to 10, 20 and 30.
func setup(t *testing.T) {
	t.Helper()
	factory = []*entry{newEntry("a", 1), newEntry("b", 2), newEntry("c", 3)}
	live = []*entry{newEntry("a", 10), newEntry("b", 20), newEntry("c", 30)}
	applied = 0
	t.Cleanup(func() {
		factory = nil
		live = nil
		applied = 0
	})
}

func liveValues() map[string]int {
	m := make(map[string]int, len(live))
	for _, one := range live {
		m[one.id] = *one.value
	}
	return m
}

func TestMarshalWritesFactoryOrderAndFillsInFromFactory(t *testing.T) {
	c := check.New(t)
	setup(t)

	// Only "c" is defined, and the live theme is deliberately different from both it and the factory.
	s := testSet{data: map[string]int{"c": 33, "unknown": 99}}
	data, err := jio.Marshal(&s)
	c.NoError(err)
	c.Equal(`{"a":1,"b":2,"c":33}`, string(data),
		"keys are written in factory order, undefined ones use the factory value, unknown ones are dropped")

	var empty testSet
	data, err = jio.Marshal(&empty)
	c.NoError(err)
	c.Equal(`{"a":1,"b":2,"c":3}`, string(data), "a zero value writes the factory values, not the live ones")
}

func TestUnmarshalFillsInMissingKeysFromFactory(t *testing.T) {
	c := check.New(t)
	setup(t)

	s := testSet{data: map[string]int{"a": 77}} // Stale data that must not survive the load.
	c.NoError(jio.Unmarshal([]byte(`{"b":22}`), &s))
	c.Equal(map[string]int{"a": 1, "b": 22, "c": 3}, s.data, "the loaded key is kept and the rest come from the factory")

	c.NoError(jio.Unmarshal([]byte(`{"b":22,"extra":5}`), &s))
	c.Equal(5, s.data["extra"], "unknown keys are retained")

	err := jio.Unmarshal([]byte(`{"b":"not a number"}`), &s)
	c.HasError(err)
	c.Equal(map[string]int{"a": 1, "b": 2, "c": 3}, s.data, "a failed load leaves a complete factory set behind")
}

func TestUnmarshalRejectsNonObject(t *testing.T) {
	c := check.New(t)
	setup(t)

	var s testSet
	c.HasError(jio.Unmarshal([]byte(`[1,2,3]`), &s))
	c.Equal(map[string]int{"a": 1, "b": 2, "c": 3}, s.data)
}

func TestCaptureCurrentCopiesLiveValues(t *testing.T) {
	c := check.New(t)
	setup(t)

	var s testSet // A zero value, as first-run settings hold.
	s.CaptureCurrent()
	c.Equal(map[string]int{"a": 10, "b": 20, "c": 30}, s.data, "capturing records every live value")

	*live[0].value = 11
	c.Equal(10, s.data["a"], "the captured value is a copy, not an alias of the live value")

	s.CaptureCurrent()
	c.Equal(11, s.data["a"], "capturing again replaces the stale value")
}

func TestMakeCurrentAppliesOnlyDefinedValuesThenNotifies(t *testing.T) {
	c := check.New(t)
	setup(t)

	s := testSet{data: map[string]int{"a": 5, "unknown": 99}}
	s.MakeCurrent()
	c.Equal(map[string]int{"a": 5, "b": 20, "c": 30}, liveValues(), "only defined keys are applied")
	c.Equal(1, applied, "the provider is told once after the values are applied")

	var empty testSet
	empty.MakeCurrent()
	c.Equal(map[string]int{"a": 5, "b": 20, "c": 30}, liveValues(), "a zero value applies nothing")
	c.Equal(2, applied, "but still notifies, so the theme refreshes")
}

func TestResetRestoresFactoryValues(t *testing.T) {
	c := check.New(t)
	setup(t)

	var s testSet
	s.Reset()
	c.Equal(map[string]int{"a": 1, "b": 2, "c": 3}, s.data, "resetting a zero value fills in every factory value")

	s.data["a"] = 50
	s.data["stale"] = 60
	s.Reset()
	c.Equal(1, s.data["a"], "a modified value is restored")
	c.Equal(60, s.data["stale"], "unknown keys are left alone")
	c.Equal(map[string]int{"a": 10, "b": 20, "c": 30}, liveValues(), "the live theme is untouched until MakeCurrent")
}

func TestResetOneRestoresOnlyTheNamedValue(t *testing.T) {
	c := check.New(t)
	setup(t)

	var s testSet
	s.ResetOne("b")
	c.Equal(map[string]int{"b": 2}, s.data, "resetting one value on a zero value defines just that value")

	s.data["a"] = 50
	s.data["b"] = 51
	s.ResetOne("a")
	c.Equal(map[string]int{"a": 1, "b": 51}, s.data, "only the named value is restored")

	var other testSet
	other.ResetOne("unknown")
	c.Equal(0, len(other.data), "resetting an unknown key adds nothing and doesn't panic on the nil map")
}

func TestRoundTripThroughJSON(t *testing.T) {
	c := check.New(t)
	setup(t)

	var s testSet
	s.CaptureCurrent()
	data, err := jio.Marshal(&s)
	c.NoError(err)
	var loaded testSet
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(s.data, loaded.data)
}
