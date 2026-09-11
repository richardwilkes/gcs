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
	"errors"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// savedCalcMarkerType exists only to be named by savedCalcMarker. Like omitCalcMarkerType (see hashable.go), it must
// stay a private, non-interface type, so that the standard library's own fast paths are left in place.
type savedCalcMarkerType struct{}

// savedCalcMarker is the unmarshalers option that marks an unmarshal as one which keeps the derived values recorded in
// the file's "calc" objects instead of deriving them again. It is the counterpart of omitCalcMarker, and like it the
// registered function is what makes it usable and is not to be removed: json.JoinUnmarshalers returns a nil
// *json.Unmarshalers when handed nothing.
var savedCalcMarker = json.JoinUnmarshalers(
	json.UnmarshalFromFunc(func(_ *jsontext.Decoder, _ *savedCalcMarkerType) error { return errors.ErrUnsupported }),
)

// keepSavedCalc reports whether the unmarshal in progress is one that keeps the saved "calc" values (see
// NewEntityFromFileWithSavedCalc). An unmarshal carrying anyone else's unmarshalers compares unequal and gets the
// ordinary treatment, in which the "calc" objects are ignored and the entity is recalculated once loaded.
func keepSavedCalc(dec *jsontext.Decoder) bool {
	unmarshalers, ok := json.GetOption(dec.Options(), json.WithUnmarshalers)
	return ok && unmarshalers == savedCalcMarker
}

// savedTraitCalc is the part of a trait's "calc" object that Trait.StringWithSavedCalc needs.
type savedTraitCalc struct {
	CurrentLevel *fxp.Int `json:"current_level"`
}

// savedNoteCalc is the part of a note's "calc" object that Note.StringWithSavedCalc needs.
type savedNoteCalc struct {
	ResolvedText string `json:"resolved_text"`
}

// savedCalc decodes a node's "calc" object into a T. A "calc" object is written for third parties and never trusted on
// load, so one that is absent, or that does not decode as a T -- a file written by hand or by a release that recorded
// something else there -- yields the zero T, whose fields the callers treat as "nothing recorded".
func savedCalc[T any](raw jsontext.Value) (calc T) {
	if len(raw) != 0 {
		var decoded T
		if err := json.Unmarshal(raw, &decoded); err == nil {
			calc = decoded
		}
	}
	return calc
}
