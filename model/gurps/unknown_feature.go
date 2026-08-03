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
	"hash"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Feature = &UnknownFeature{}

// UnknownFeature holds a feature whose type this version of GCS doesn't understand, most likely because it was written
// by a newer version. The original JSON is retained verbatim and written back out unchanged when the file is saved, so
// that merely loading and saving a file cannot destroy data. It participates in no calculations.
type UnknownFeature struct {
	// Kind is the unrecognized value that was found in the "type" field.
	Kind string
	// Data is the original JSON for the feature, exactly as it was read.
	Data jsontext.Value
}

// NewUnknownFeature creates a new UnknownFeature holding a copy of the passed-in data.
func NewUnknownFeature(kind string, data jsontext.Value) *UnknownFeature {
	return &UnknownFeature{
		Kind: kind,
		Data: slices.Clone(data),
	}
}

// FeatureType implements Feature.
func (u *UnknownFeature) FeatureType() feature.Type {
	return feature.Unknown
}

// Clone implements Feature.
func (u *UnknownFeature) Clone() Feature {
	return NewUnknownFeature(u.Kind, u.Data)
}

// FillWithNameableKeys implements Feature.
func (u *UnknownFeature) FillWithNameableKeys(_, _ map[string]string) {
}

// MarshalJSONTo implements json.MarshalerTo. The original data is written back out as-is.
func (u *UnknownFeature) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(u.Data)
}

// Hash writes this object's contents into the hasher.
func (u *UnknownFeature) Hash(h hash.Hash) {
	if u == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, feature.Unknown)
	xhash.StringWithLen(h, u.Kind)
	xhash.BytesWithLen(h, u.Data)
}
