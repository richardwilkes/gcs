// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// PageRefs holds a set of page references.
type PageRefs struct {
	data map[string]*PageRef
}

// PageRef holds a path to a file and an offset for all page references within that file.
type PageRef struct {
	ID     string `json:"-"`
	Path   string `json:"path,omitzero"`
	Offset int    `json:"offset,omitzero"`
}

// NewPageRefsFromFS creates a new set of page references from a file.
func NewPageRefsFromFS(fileSystem fs.FS, filePath string) (*PageRefs, error) {
	var p PageRefs
	if err := jio.Load(fileSystem, filePath, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Save writes the PageRefs to the file as JSON.
func (p *PageRefs) Save(filePath string) error {
	return jio.SaveToFile(filePath, p)
}

// MarshalJSONTo implements json.MarshalerTo.
func (p *PageRefs) MarshalJSONTo(enc *jsontext.Encoder) error {
	if p.data == nil {
		p.data = make(map[string]*PageRef)
	}
	return json.MarshalEncode(enc, &p.data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (p *PageRefs) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if err := json.UnmarshalDecode(dec, &p.data); err != nil {
		return err
	}
	for k, v := range p.data {
		v.ID = k
	}
	return nil
}

// IsZero implements json.isZero.
func (p *PageRefs) IsZero() bool {
	return p == nil || p.data == nil || len(p.data) == 0
}

// Lookup the PageRef for the given ID. If not found or if the path it points to isn't a readable file, returns nil.
func (p *PageRefs) Lookup(id string) *PageRef {
	if ref, ok := p.data[id]; ok && xos.FileIsReadable(ref.Path) {
		r := *ref // Make a copy so that clients can't muck with our data
		return &r
	}
	return nil
}

// Set the PageRef.
func (p *PageRefs) Set(pageRef *PageRef) {
	if pageRef.ID == "" {
		errs.Log(errs.New("invalid page reference ID"), "id", pageRef.ID)
		return
	}
	if p.data == nil {
		p.data = make(map[string]*PageRef)
	}
	r := *pageRef
	p.data[pageRef.ID] = &r
}

// Remove the PageRef for the ID.
func (p *PageRefs) Remove(id string) {
	if p.data != nil {
		delete(p.data, id)
	}
}

// List returns a sorted list of page references.
func (p *PageRefs) List() []*PageRef {
	list := make([]*PageRef, 0, len(p.data))
	for _, v := range p.data {
		r := *v
		list = append(list, &r)
	}
	slices.SortFunc(list, func(a, b *PageRef) int { return xstrings.NaturalCmp(a.ID, b.ID, true) })
	return list
}
