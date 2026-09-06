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
	"encoding/json/v2"

	"github.com/richardwilkes/toolbox/v2/errs"
)

// unmarshalTypedList decodes a JSON array of polymorphic objects, each of which carries a "type" key naming its
// concrete kind. extract resolves that string to the caller's enum, reporting whether it was recognized; alloc returns
// a fresh, empty value of the concrete type for a recognized kind, or the zero T for a kind it has no representation
// for; and unknown wraps the raw JSON of anything that cannot be decoded into a concrete type so that it survives a
// load/save cycle unchanged rather than being discarded.
//
// The type is extracted as a string and resolved with extract rather than being unmarshaled directly into the enum:
// the generated enums' UnmarshalText maps anything they don't recognize onto their first value, which would silently
// turn data written by a newer version of GCS into a bogus value of that first type.
func unmarshalTypedList[T, K comparable](dec *jsontext.Decoder, extract func(string) (K, bool), alloc func(K) T, unknown func(kind string, raw jsontext.Value) T) ([]T, error) {
	var v []jsontext.Value
	if err := json.UnmarshalDecode(dec, &v); err != nil {
		return nil, errs.Wrap(err)
	}
	var zero T
	result := make([]T, len(v))
	for i, one := range v {
		var justTypeData struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(one, &justTypeData, dec.Options()); err != nil {
			return nil, errs.Wrap(err)
		}
		var target T
		if kind, known := extract(justTypeData.Type); known {
			target = alloc(kind)
		}
		if target == zero {
			// Either the kind wasn't recognized at all, or it is known but has no concrete type (the Unknown value
			// itself, or a kind alloc has no case for). Preserve rather than discard.
			result[i] = unknown(justTypeData.Type, one)
			continue
		}
		if err := json.Unmarshal(one, &target, dec.Options()); err != nil {
			return nil, errs.Wrap(err)
		}
		result[i] = target
	}
	return result, nil
}
