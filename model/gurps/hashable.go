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
	"encoding/base64"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"hash"
	"maps"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/zeebo/xxh3"
)

// omitCalcMarkerType exists only to give omitCalcMarker a type to be registered against. Nothing is ever of this type,
// so the function registered for it is never called.
type omitCalcMarkerType struct{}

// omitCalcMarker marks a marshal whose "calc" members are going to be thrown away, so that the marshalers can skip
// producing them. Only its identity carries meaning: the one function in it is for a type that never appears in any
// data, so it alters nothing about how anything is encoded.
//
// A trait, skill or spell a grant put on the sheet also leaves out, under the marker, the details it takes from that
// grant (see withoutGrantDetails). A marshal that is saved, as SellMoneyList's is, must therefore never use the marker
// for such rows, or the file would lose those details.
//
// json/v2 has no context to thread through a marshal and no way to define an option of one's own — the option lookup
// switches over the standard library's own option types, and the interface an option must satisfy cannot be
// implemented outside it. What it does have is an options bag that reaches every nested MarshalJSONTo through
// jsontext.Encoder.Options, and introspecting it from a MarshalJSONTo is exactly what it is documented for. A
// *json.Marshalers is the one option that can be given a value of our own choosing, so it is what carries the flag.
//
// The registered function is what makes this usable, and is not to be removed: json.JoinMarshalers returns a nil
// *json.Marshalers when handed nothing, and a nil one stored in the options panics the moment a value of an interface
// type is marshaled. The type it names must also stay a private, non-interface one, so that the standard library's own
// fast path for "any" values is left in place.
var omitCalcMarker = json.JoinMarshalers(
	json.MarshalToFunc(func(_ *jsontext.Encoder, _ omitCalcMarkerType) error { return errors.ErrUnsupported }),
)

// omitCalc reports whether the marshal in progress is one whose "calc" members will be discarded, meaning the work of
// computing those derived values — which runs the scripts in the data — can be skipped entirely, and whose granted rows
// leave out the details their grants give them (see omitCalcMarker). A marshal carrying anyone else's marshalers
// compares unequal and gets the full encoding.
func omitCalc(enc *jsontext.Encoder) bool {
	marshalers, ok := json.GetOption(enc.Options(), json.WithMarshalers)
	return ok && marshalers == omitCalcMarker
}

// Hashable is an object that can be hashed.
type Hashable interface {
	Hash(hash.Hash)
}

// syncKind returns the kind that starts the hash of the data. It is the kind of the data's TID, except for data that
// shares a kind with the other form it converts into (see sharedSyncKind).
func syncKind(data interface{ ID() tid.TID }) byte {
	if kind := sharedSyncKind(data); kind != 0 {
		return kind
	}
	return tidKind(data.ID())
}

// sharedSyncKind returns the kind the data shares with the other form it can be converted into, since the two have the
// same data to sync, or 0 if it shares none: plain equipment and a physical container share the kind of plain
// equipment, and a note and a note container the kind of a note.
func sharedSyncKind(data any) byte {
	switch one := data.(type) {
	case *Equipment:
		if !one.Container() || one.IsPhysicalContainer() {
			return kinds.Equipment
		}
	case *Note:
		return kinds.Note
	}
	return 0
}

// tidKind returns the kind of the TID, or 0 if it is empty.
func tidKind(id tid.TID) byte {
	if id != "" {
		return id[0]
	}
	return 0
}

// hashList writes a length-prefixed list of Hashable objects into the hasher.
func hashList[T Hashable](h hash.Hash, list []T) {
	xhash.Num64(h, len(list))
	for _, one := range list {
		one.Hash(h)
	}
}

// hashStrings writes a length-prefixed list of strings into the hasher.
func hashStrings[T ~string](h hash.Hash, list []T) {
	xhash.Num64(h, len(list))
	for _, one := range list {
		xhash.StringWithLen(h, one)
	}
}

// HashAndData is a combination of a hash and some data.
type HashAndData struct {
	Hash uint64
	Data any
}

// Hash64 returns a 64-bit hash of the Hashable object.
func Hash64(in Hashable) uint64 {
	h := xxh3.New()
	in.Hash(h)
	return h.Sum64()
}

// SameItem reports whether two rows with the same Hash64 describe the same item: they also need the same nameable
// replacements and, for rows that carry a tech level, the same tech level, neither of which is part of the hash.
func SameItem[T interface {
	Hashable
	NameableReplacements() map[string]string
}](a, b T) bool {
	if !maps.Equal(a.NameableReplacements(), b.NameableReplacements()) {
		return false
	}
	if tla, ok := any(a).(TechLevelProvider); ok {
		if tlb, ok2 := any(b).(TechLevelProvider); ok2 {
			return sameTechLevel(tla, tlb)
		}
	}
	return true
}

// HashJSON writes the JSON encoding of the given object into the hasher, without any of the "calc" members. This is how
// the top-level documents — the ones whose whole contents are hashed to decide whether they have unsaved changes — are
// hashed.
//
// The "calc" members hold derived values (attribute maximums, skill levels, resolved notes, and so on) that exist only
// so third-party tools can read a saved file without reimplementing GCS. They add nothing to a hash that already covers
// every field they are derived from, and including them makes the hash something other than a function of the
// document's data: computing them runs the scripts the data owner wrote, and a script is abandoned once it exceeds the
// permitted per-script execution time, resolving to 0 instead of its real value. On a machine that is briefly too busy
// for a script to finish in time, an untouched document would otherwise hash differently than it did a moment earlier
// and start reporting itself as modified.
//
// For the same reason, a trait, skill or spell a grant put on the sheet leaves out the details it takes from the grant
// giving it (see withoutGrantDetails), since which grant that is turns on prerequisites, which may be scripts.
//
// The marshal is marked, which is what the data objects consult to leave those values out, so the work of producing
// them is never done here at all.
func HashJSON(h hash.Hash, in any) {
	if err := json.MarshalWrite(h, in, json.Deterministic(true), json.WithMarshalers(omitCalcMarker)); err != nil {
		errs.Log(err)
	}
}

// MarshalWithoutCalc returns the JSON encoding of the given object without any of the "calc" members, in the same form
// HashJSON hands to a hasher. It is for the callers that compare the encodings of two pieces of document data to decide
// whether one differs from the other, such as an open editor asking whether it holds unapplied changes.
//
// Such a comparison must be a function of the data alone, which the derived "calc" values are not (see HashJSON).
// Leaving them out also spares the caller from running the data's scripts, which matters for a comparison made as often
// as one asking whether an editor is modified.
//
// The encoding is deterministic, since the bytes are compared directly and json/v2 is otherwise free to write a map's
// entries in a different order each time.
func MarshalWithoutCalc(in any) ([]byte, error) {
	data, err := jio.Marshal(in, json.Deterministic(true), json.WithMarshalers(omitCalcMarker))
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return data, nil
}

// NodesToHashesByID traverses the provided nodes and generates hashes.
func NodesToHashesByID[T Node[T]](result map[tid.TID]HashAndData, data ...T) {
	Traverse(func(one T) bool {
		id := one.ID()
		if _, exists := result[id]; !exists {
			result[id] = HashAndData{
				Hash: Hash64(one),
				Data: one,
			}
		}
		return false
	}, false, false, data...)
}

// TIDFromHashedString creates a TID of the given kind from the given strings. Each string is hashed along with its
// length, so two different splits of the same text produce different TIDs, and a single string produces the same TID it
// always has.
func TIDFromHashedString(kind byte, parts ...string) tid.TID {
	h := xxh3.New()
	for _, part := range parts {
		xhash.StringWithLen(h, part)
	}
	buffer := h.Sum(make([]byte, 0, 12))
	for len(buffer) < 12 {
		buffer = append(buffer, 0)
	}
	if len(buffer) > 12 {
		buffer = buffer[:12]
	}
	return tid.TID(fmt.Sprintf("%c%s", kind, base64.RawURLEncoding.EncodeToString(buffer)))
}
