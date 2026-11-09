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
	"encoding/json/jsontext"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestBodyFileKeepsUnknownMembers verifies that saving a body type loaded from a file, or a clone of it, writes back the
// members of the file this release does not know.
func TestBodyFileKeepsUnknownMembers(t *testing.T) {
	c := check.New(t)
	const data = `{"type":"body","version":6,"name":"Test","roll":"3d","locations":[{"id":"eye","choice_name":"Eyes","table_name":"Eyes"}],"zzz_future":{"a":1}}`
	body, err := NewBodyFromFile(fstest.MapFS{"Test.body": {Data: []byte(data)}}, "Test.body")
	c.NoError(err)
	path := filepath.Join(t.TempDir(), "Test.body")
	for _, one := range []struct {
		name string
		body *Body
	}{
		{name: "the body", body: body},
		{name: "its clone", body: body.Clone(nil, nil)},
	} {
		c.NoError(one.body.Save(path), one.name)
		var saved []byte
		saved, err = os.ReadFile(path)
		c.NoError(err, one.name)
		c.Equal(`{"a":1}`, jsonMember(t, saved, "zzz_future"), one.name)
	}
}

// TestFactoryFilesHoldNoUnknownMembers verifies that the factory attribute, body type and ancestry files leave nothing
// among the members this release does not know, both as loaded and once saved and loaded again, so that every member
// they hold, or that saving them writes, is one their decoding reads.
func TestFactoryFilesHoldNoUnknownMembers(t *testing.T) {
	checkHoldsNoUnknownMembers(t, "Standard.attr", NewAttributeDefsFromFile, (*AttributeDefs).Save)
	checkHoldsNoUnknownMembers(t, "Humanoid.body", NewBodyFromFile, (*Body).Save)
	checkHoldsNoUnknownMembers(t, "Human.ancestry", NewAncestryFromFile, (*Ancestry).Save)
}

// checkHoldsNoUnknownMembers runs a subtest that loads the named factory file, saves it and loads it again, checking
// that each load holds no unknown members.
func checkHoldsNoUnknownMembers[T any](t *testing.T, name string, load func(fs.FS, string) (T, error), save func(T, string) error) {
	t.Run(name, func(t *testing.T) {
		c := check.New(t)
		data, err := load(embeddedFS, "embedded_data/"+name)
		c.NoError(err)
		c.Equal([]string(nil), filledFallbacks(data), "as loaded")
		dir := t.TempDir()
		c.NoError(save(data, filepath.Join(dir, name)))
		data, err = load(os.DirFS(dir), name)
		c.NoError(err)
		c.Equal([]string(nil), filledFallbacks(data), "once saved and loaded again")
	})
}

// filledFallbacks returns, for each fallback reachable from v that holds members this release does not know, its path
// and contents. Data this release wrote leaves each one empty, unless a member its decoding doesn't name is written.
func filledFallbacks(v any) []string {
	valueType := reflect.TypeFor[jsontext.Value]()
	type visit struct {
		ptr uintptr
		typ reflect.Type
	}
	seen := make(map[visit]bool)
	var paths []string
	var walk func(rv reflect.Value, path string)
	walk = func(rv reflect.Value, path string) {
		switch rv.Kind() {
		case reflect.Pointer:
			if key := (visit{ptr: rv.Pointer(), typ: rv.Type()}); !rv.IsNil() && !seen[key] {
				seen[key] = true
				walk(rv.Elem(), path)
			}
		case reflect.Interface:
			if !rv.IsNil() {
				walk(rv.Elem(), path)
			}
		case reflect.Struct:
			for i := range rv.NumField() {
				field := rv.Type().Field(i)
				fv := rv.Field(i)
				switch {
				case field.Type != valueType:
					walk(fv, path+"."+field.Name)
				case (field.Tag.Get("json") == ",embed" || field.Name == "fileUnknown") && fv.Len() != 0:
					paths = append(paths, path+"."+field.Name+": "+string(fv.Bytes()))
				}
			}
		case reflect.Slice, reflect.Array:
			for i := range rv.Len() {
				walk(rv.Index(i), fmt.Sprintf("%s[%d]", path, i))
			}
		case reflect.Map:
			for iter := rv.MapRange(); iter.Next(); {
				walk(iter.Value(), fmt.Sprintf("%s[%v]", path, iter.Key()))
			}
		default:
		}
	}
	walk(reflect.ValueOf(v), "")
	return paths
}
