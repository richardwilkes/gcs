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
	"fmt"
	"hash"

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
	json.MarshalToFunc(func(_ *jsontext.Encoder, _ omitCalcMarkerType) error { return json.SkipFunc }),
)

// omitCalc reports whether the marshal in progress is one whose "calc" members will be discarded, meaning the work of
// computing those derived values — which runs the scripts in the data — can be skipped entirely. A marshal carrying
// anyone else's marshalers compares unequal and gets the full encoding.
func omitCalc(enc *jsontext.Encoder) bool {
	marshalers, ok := json.GetOption(enc.Options(), json.WithMarshalers)
	return ok && marshalers == omitCalcMarker
}

// Hashable is an object that can be hashed.
type Hashable interface {
	Hash(hash.Hash)
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
// Such a comparison must be a function of the data alone. The "calc" members are derived values, and computing them
// runs the scripts the data owner wrote; a script is abandoned once it exceeds the permitted per-script execution time
// and then resolves to 0 rather than its real value. Whether that happens depends on how busy the machine is, so two
// encodings of identical data could otherwise disagree — and the caller would report changes that were never made.
// Leaving the derived values out also spares the caller from running those scripts at all, which matters when the
// comparison is made as often as one asking whether an editor is modified.
//
// The encoding is deterministic, since the bytes are compared directly and json/v2 is otherwise free to write a map's
// entries in a different order each time.
func MarshalWithoutCalc(in any) ([]byte, error) {
	data, err := json.Marshal(in, json.Deterministic(true), json.WithMarshalers(omitCalcMarker))
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return data, nil
}

// NodesToHashesByID traverses the provided nodes and generates hashes.
func NodesToHashesByID[T NodeTypes](result map[tid.TID]HashAndData, data ...T) {
	Traverse(func(one T) bool {
		node := AsNode(one)
		id := node.ID()
		if _, exists := result[id]; !exists {
			result[id] = HashAndData{
				Hash: Hash64(node),
				Data: one,
			}
		}
		return false
	}, false, false, data...)
}

// TIDFromHashedString creates a TID from a string.
func TIDFromHashedString(kind byte, s string) tid.TID {
	h := xxh3.New()
	xhash.StringWithLen(h, s)
	buffer := h.Sum(make([]byte, 0, 12))
	for len(buffer) < 12 {
		buffer = append(buffer, 0)
	}
	if len(buffer) > 12 {
		buffer = buffer[:12]
	}
	return tid.TID(fmt.Sprintf("%c%s", kind, base64.RawURLEncoding.EncodeToString(buffer)))
}
