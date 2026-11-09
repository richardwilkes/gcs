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
	"cmp"
	"encoding"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/filternode"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

const dataSchemaGoldenPath = "testdata/data_schema.golden"

var (
	updateDataSchema = flag.Bool("update", false,
		"rewrite "+dataSchemaGoldenPath+" from the current code and data version constants")
	releaseDataSchema = flag.Bool("release", false,
		"also fail unless "+dataSchemaGoldenPath+" was regenerated for the current data version")
)

var (
	importPathDirsRE = regexp.MustCompile(`(?:[\w.~-]+/)+`)
	gurpsQualifierRE = regexp.MustCompile(`\bgurps\.`)
)

// TestDataSchemaChangesBumpTheVersion fails when what GCS writes to its data files has changed since
// testdata/data_schema.golden was regenerated (the last step before a release) without jio.CurrentDataVersion being
// raised, or when something was removed without jio.MinimumReaderVersion being raised too. The schema is fingerprinted
// from the json/v2 members of every struct reachable from the data file roots, the values of every enum met along the
// way, the concrete type behind each kind of the polymorphic lists, and the theme color and font keys. For a type with
// a hand-written marshaler only its tagged members are recorded, and what that marshaler writes beyond them, such as
// the computed "calc" objects, the maps behind PageRefs and KeyBindings, or the text forms of the attack types, is not
// seen. The golden keeps the last release's schema while the constants are raised; it is rewritten with -update as the
// last step before tagging, which the release workflow enforces by running this test with -release.
func TestDataSchemaChangesBumpTheVersion(t *testing.T) {
	c := check.New(t)
	current := dataSchemaFingerprint()
	if *updateDataSchema {
		var buffer strings.Builder
		fmt.Fprintf(&buffer, "version %d min_reader %d\n", jio.CurrentDataVersion, jio.MinimumReaderVersion)
		for _, line := range current {
			buffer.WriteString(line)
			buffer.WriteByte('\n')
		}
		c.NoError(os.WriteFile(dataSchemaGoldenPath, []byte(buffer.String()), 0o644))
		return
	}
	const regenerate = "go test ./model/gurps -run TestDataSchema -update"
	data, err := os.ReadFile(dataSchemaGoldenPath)
	if errors.Is(err, fs.ErrNotExist) {
		c.Fatalf("%s is missing; create it with: %s", dataSchemaGoldenPath, regenerate)
	}
	c.NoError(err)
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(data)), "\r", ""), "\n")
	var version, minReader int
	if _, err = fmt.Sscanf(lines[0], "version %d min_reader %d", &version, &minReader); err != nil {
		c.Fatalf("%s does not start with its version line; recreate it with: %s", dataSchemaGoldenPath, regenerate)
	}
	added, removed := stringSetDifference(current, lines[1:]), stringSetDifference(lines[1:], current)
	changed := len(added) != 0 || len(removed) != 0
	c.True(version <= jio.CurrentDataVersion && minReader <= jio.MinimumReaderVersion,
		"the data version went backwards: %s records version %d min_reader %d, but jio.CurrentDataVersion is %d and jio.MinimumReaderVersion is %d",
		dataSchemaGoldenPath, version, minReader, jio.CurrentDataVersion, jio.MinimumReaderVersion)
	var addedText, removedText strings.Builder
	for _, line := range added {
		addedText.WriteString("\n+ " + line)
	}
	for _, line := range removed {
		removedText.WriteString("\n- " + line)
	}
	const untilRelease = "leave the golden as is until the release, whose last step before tagging is: " + regenerate
	c.False(changed && version == jio.CurrentDataVersion,
		"what the data files hold has changed since the last release, so bump jio.CurrentDataVersion, then %s%s%s",
		untilRelease, addedText.String(), removedText.String())
	c.False(len(removed) != 0 && minReader == jio.MinimumReaderVersion,
		"something the data files held at the last release was removed, which is a breaking change, so bump jio.MinimumReaderVersion too, then %s%s",
		untilRelease, removedText.String())
	if *releaseDataSchema {
		c.False(changed || version != jio.CurrentDataVersion || minReader != jio.MinimumReaderVersion,
			"%s was not regenerated for this release; run: %s", dataSchemaGoldenPath, regenerate)
	}
}

// dataSchemaFingerprint returns the sorted lines that describe what the data files can hold.
func dataSchemaFingerprint() []string {
	s := &dataSchema{lines: make(map[string]bool), seen: make(map[reflect.Type]bool)}
	for _, root := range []reflect.Type{
		reflect.TypeFor[EntityData](),
		reflect.TypeFor[TemplateData](),
		reflect.TypeFor[LootData](),
		reflect.TypeFor[CampaignData](),
		reflect.TypeFor[listData[*Trait]](),
		reflect.TypeFor[listData[*TraitModifier]](),
		reflect.TypeFor[listData[*Equipment]](),
		reflect.TypeFor[listData[*EquipmentModifier]](),
		reflect.TypeFor[listData[*Skill]](),
		reflect.TypeFor[listData[*Spell]](),
		reflect.TypeFor[listData[*Note]](),
		reflect.TypeFor[standaloneBodyData](),
		reflect.TypeFor[attributeDefsData](),
		reflect.TypeFor[ancestryData](),
		reflect.TypeFor[listFiltersData](),
		reflect.TypeFor[skillLevelCapsData](),
		reflect.TypeFor[Settings](),
		reflect.TypeFor[GeneralSettings](),
		reflect.TypeFor[SheetSettings](),
		reflect.TypeFor[KeyBindings](),
		reflect.TypeFor[keyBindingsData](),
		reflect.TypeFor[PageRefs](),
		reflect.TypeFor[pageRefsData](),
		reflect.TypeFor[PageRef](), // PageRefs writes its unexported map of these by hand.
		reflect.TypeFor[nameGeneratorData](),
	} {
		s.walk(root)
	}
	addDataSchemaKinds(s, "Feature", feature.Types, allocFeature)
	addDataSchemaKinds(s, "Prereq", prereq.Types, allocPrereq)
	addDataSchemaKinds(s, "LevelCap", levelcap.Types, allocLevelCap)
	addDataSchemaKinds(s, "FilterNode", filternode.Types, allocFilterNode)
	for _, one := range colors.Factory() {
		s.add("keys", "colors", one.Key())
	}
	for _, one := range fonts.FactoryFonts() {
		s.add("keys", "fonts", one.Key())
	}
	return slices.Sorted(maps.Keys(s.lines))
}

type dataSchema struct {
	lines map[string]bool
	seen  map[reflect.Type]bool
}

func (s *dataSchema) add(fields ...string) {
	s.lines[strings.Join(fields, " ")] = true
}

// walk adds the lines for t and everything reachable from it, once per type.
func (s *dataSchema) walk(t reflect.Type) {
	if s.seen[t] {
		return
	}
	s.seen[t] = true
	if s.addEnum(t) {
		return
	}
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		s.walk(t.Elem())
	case reflect.Map:
		s.walk(t.Key())
		s.walk(t.Elem())
	case reflect.Struct:
		// A hand-written marshaler by convention writes the members with an explicit name, while untagged exported
		// fields hold runtime state, so those are left out but still walked, as such a marshaler may write a map or
		// text form built from them.
		name := dataSchemaTypeName(t)
		handWritten := hasHandWrittenMarshaler(t)
		for _, m := range schemaMembers(t) {
			if m.hasName || !handWritten {
				desc := "raw"
				if m.typ != reflect.TypeFor[jsontext.Value]() {
					desc = dataSchemaTypeName(m.typ)
				}
				s.add("struct", name, m.name, desc)
			}
			s.walk(m.typ)
		}
	}
}

// addEnum adds a line for each value of t and returns true if t is an enum: a named integer type with an EnsureValid()
// method. Each value from 0 to 255 for which EnsureValid returns the value it was given is recorded by its Key(), or by
// its number for an enum without a Key() string method, which is written as that number.
func (s *dataSchema) addEnum(t reflect.Type) bool {
	if t.Name() == "" {
		return false
	}
	var set func(v reflect.Value, i int)
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		set = func(v reflect.Value, i int) { v.SetInt(int64(i)) }
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		set = func(v reflect.Value, i int) { v.SetUint(uint64(i)) }
	default:
		return false
	}
	ensure, ok := t.MethodByName("EnsureValid")
	if !ok || ensure.Type.NumIn() != 1 || ensure.Type.NumOut() != 1 || ensure.Type.Out(0) != t {
		return false
	}
	key, hasKey := t.MethodByName("Key")
	hasKey = hasKey && key.Type.NumIn() == 1 && key.Type.NumOut() == 1 && key.Type.Out(0).Kind() == reflect.String
	name := dataSchemaTypeName(t)
	for i := range 256 {
		v := reflect.New(t).Elem()
		set(v, i)
		if !ensure.Func.Call([]reflect.Value{v})[0].Equal(v) {
			continue
		}
		if hasKey {
			s.add("enum", name, cmp.Or(key.Func.Call([]reflect.Value{v})[0].String(), `""`))
		} else {
			s.add("enum", name, strconv.Itoa(i))
		}
	}
	return true
}

// addDataSchemaKinds adds a line naming the concrete type alloc returns for each kind of a polymorphic list, or "-"
// for a kind that has none, and walks those types.
func addDataSchemaKinds[K interface{ Key() string }, T any](s *dataSchema, name string, kinds []K, alloc func(K) T) {
	s.walk(reflect.TypeFor[K]())
	for _, kind := range kinds {
		concrete := "-"
		if v := any(alloc(kind)); v != nil {
			t := reflect.TypeOf(v)
			concrete = dataSchemaTypeName(t)
			s.walk(t)
		}
		s.add("kind", name, kind.Key(), concrete)
	}
}

// dataSchemaTypeName returns the name of t without pointers or import paths, and without the qualifier for this
// package. Other packages keep their qualifier, since several enum types share a name.
func dataSchemaTypeName(t reflect.Type) string {
	if t.Name() == "" {
		switch t.Kind() {
		case reflect.Pointer:
			return dataSchemaTypeName(t.Elem())
		case reflect.Slice:
			return "[]" + dataSchemaTypeName(t.Elem())
		case reflect.Array:
			return fmt.Sprintf("[%d]%s", t.Len(), dataSchemaTypeName(t.Elem()))
		case reflect.Map:
			return "map[" + dataSchemaTypeName(t.Key()) + "]" + dataSchemaTypeName(t.Elem())
		}
	}
	name := strings.ReplaceAll(strings.ReplaceAll(t.String(), "*", ""), "interface {}", "any")
	return gurpsQualifierRE.ReplaceAllString(importPathDirsRE.ReplaceAllString(name, ""), "")
}

type schemaMember struct {
	name    string
	typ     reflect.Type
	depth   int
	hasName bool
}

// schemaMembers returns the members json/v2 reads and writes for the struct type t, following its rules for promoting
// the members of embedded structs and for resolving conflicts between their names. Fallback fields that collect
// unknown members are left out.
func schemaMembers(t reflect.Type) []schemaMember {
	type queued struct {
		typ           reflect.Type
		depth         int
		visitChildren bool
	}
	queue := []queued{{typ: t, visitChildren: true}}
	seen := map[reflect.Type]bool{t: true}
	all := make([]schemaMember, 0, t.NumField())
	for len(queue) != 0 {
		q := queue[0]
		queue = queue[1:]
		for sf := range q.typ.Fields() {
			tag := sf.Tag.Get("json")
			if tag == "-" || (!sf.IsExported() && !sf.Anonymous) {
				continue
			}
			name, options, _ := strings.Cut(tag, ",")
			hasName := name != ""
			if !hasName {
				name = sf.Name
			}
			ft := sf.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if slices.Contains(strings.Split(options, ","), "embed") ||
				(sf.Anonymous && !hasName && ft.Kind() == reflect.Struct) {
				if ft.Kind() == reflect.Struct {
					if q.visitChildren {
						queue = append(queue, queued{typ: ft, depth: q.depth + 1, visitChildren: !seen[ft]})
					}
					seen[ft] = true
				}
				continue
			}
			if sf.IsExported() || (sf.Anonymous && ft.Kind() == reflect.Struct) {
				all = append(all, schemaMember{name: name, typ: sf.Type, depth: q.depth, hasName: hasName})
			}
		}
	}
	unnamed := func(m schemaMember) int {
		if m.hasName {
			return 0
		}
		return 1
	}
	slices.SortStableFunc(all, func(a, b schemaMember) int {
		return cmp.Or(strings.Compare(a.name, b.name), cmp.Compare(a.depth, b.depth), cmp.Compare(unnamed(a), unnamed(b)))
	})
	var members []schemaMember
	for len(all) != 0 {
		n := 1
		for n < len(all) && all[n-1].name == all[n].name {
			n++
		}
		if n == 1 || all[0].depth != all[1].depth || all[0].hasName != all[1].hasName {
			members = append(members, all[0])
		}
		all = all[n:]
	}
	return members
}

// hasHandWrittenMarshaler returns true if json/v2 calls a method of t, or of a pointer to it, to write it.
func hasHandWrittenMarshaler(t reflect.Type) bool {
	pt := reflect.PointerTo(t)
	return pt.Implements(reflect.TypeFor[json.MarshalerTo]()) || pt.Implements(reflect.TypeFor[json.Marshaler]()) ||
		pt.Implements(reflect.TypeFor[encoding.TextAppender]()) || pt.Implements(reflect.TypeFor[encoding.TextMarshaler]())
}

// stringSetDifference returns the sorted strings in a that are not in b.
func stringSetDifference(a, b []string) []string {
	inB := make(map[string]bool, len(b))
	for _, one := range b {
		inB[one] = true
	}
	result := make([]string, 0, len(a))
	for _, one := range a {
		if !inB[one] {
			result = append(result, one)
		}
	}
	slices.Sort(result)
	return result
}
