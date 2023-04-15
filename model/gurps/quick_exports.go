/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"sort"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
)

// ExportInfo holds information about a recent export so that it can be redone quickly.
type ExportInfo struct {
	FilePath     string   `json:"file_path"`
	TemplatePath string   `json:"template_path"`
	ExportPath   string   `json:"export_path"`
	LastUsed     jio.Time `json:"last_used"`
}

// QuickExportsData holds the QuickExports data that is written to disk.
type QuickExportsData struct {
	Max     int           `json:"max"`
	Exports []*ExportInfo `json:"exports,omitempty"`
}

// QuickExports holds a list containing information about previous exports.
type QuickExports struct {
	QuickExportsData
}

// NewQuickExports creates a new, empty, QuickExports object.
func NewQuickExports() *QuickExports {
	return &QuickExports{QuickExportsData: QuickExportsData{Max: 20}}
}

// MarshalJSON implements json.Marshaler.
func (q *QuickExports) MarshalJSON() ([]byte, error) {
	sort.Slice(q.Exports, func(i, j int) bool { return q.Exports[i].LastUsed.After(q.Exports[j].LastUsed) })
	if q.Max < 0 {
		q.Max = 0
	}
	if len(q.Exports) > q.Max {
		list := make([]*ExportInfo, q.Max)
		copy(list, q.Exports)
		q.Exports = list
	}
	return json.Marshal(&q.QuickExportsData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (q *QuickExports) UnmarshalJSON(data []byte) error {
	q.QuickExportsData = QuickExportsData{}
	if err := json.Unmarshal(data, &q.QuickExportsData); err != nil {
		return err
	}
	if q.Max < 0 {
		q.Max = 0
	}
	return nil
}

// Empty implements encoding.Empty.
func (q *QuickExports) Empty() bool {
	return len(q.Exports) == 0
}
