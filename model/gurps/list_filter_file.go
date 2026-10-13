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
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/jio"
)

type listFiltersData struct {
	Version     int                      `json:"version"`
	ListFilters map[string][]*ListFilter `json:"list_filters,omitempty"`
}

// NewListFiltersFromFile loads saved filters, by list type key, from a file. They are made valid the same way the ones
// in the settings are.
func NewListFiltersFromFile(fileSystem fs.FS, filePath string) (map[string][]*ListFilter, error) {
	var data listFiltersData
	if err := jio.LoadVersionedFile(fileSystem, filePath, &data, &data.Version); err != nil {
		return nil, err
	}
	s := Settings{ListFilters: data.ListFilters}
	s.ensureListFiltersValidity()
	return s.ListFilters, nil
}

// SaveListFilters writes saved filters, by list type key, to a file.
func SaveListFilters(filters map[string][]*ListFilter, filePath string) error {
	return jio.SaveToFile(filePath, &listFiltersData{
		Version:     jio.CurrentDataVersion,
		ListFilters: filters,
	})
}
