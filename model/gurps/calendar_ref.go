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
	"context"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/toolbox/errs"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

// CalendarRef holds a named reference to a calendar.
type CalendarRef struct {
	Name     string
	Calendar *calendar.Calendar
}

// AvailableCalendarRefs scans the libraries and returns the available calendars.
func AvailableCalendarRefs(libraries Libraries) []*NamedFileSet {
	return ScanForNamedFileSets(embeddedFS, "embedded_data", true, libraries, CalendarExt)
}

// LookupCalendarRef a CalendarRef by name.
func LookupCalendarRef(name string, libraries Libraries) *CalendarRef {
	for _, lib := range AvailableCalendarRefs(libraries) {
		for _, one := range lib.List {
			if one.Name == name {
				if a, err := NewCalendarRefFromFS(one.FileSystem, one.FilePath); err != nil {
					errs.Log(err, "path", one.FilePath)
				} else {
					return a
				}
			}
		}
	}
	return nil
}

// NewCalendarRefFromFS creates a new CalendarRef from a file.
func NewCalendarRefFromFS(fileSystem fs.FS, filePath string) (*CalendarRef, error) {
	var c calendar.Calendar
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &c); err != nil {
		return nil, err
	}
	return &CalendarRef{
		Name:     xfs.BaseName(filePath),
		Calendar: &c,
	}, nil
}

// RandomBirthday generates a random birthday month and day.
func (c *CalendarRef) RandomBirthday(not string) string {
	year := 1
	base := 0
	if c.Calendar.LeapYear != nil {
		for !c.Calendar.IsLeapYear(year) {
			year++
		}
		base = c.Calendar.MustNewDate(1, 1, year).Days
	}
	daysInYear := c.Calendar.Days(year)
	result := ""
	for range 5 {
		if result = c.Calendar.NewDateByDays(base + rand.NewCryptoRand().Intn(daysInYear)).Format("%M %D"); result != not {
			break
		}
	}
	return result
}
