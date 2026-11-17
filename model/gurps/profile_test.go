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
	"log/slog"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestProfilePortraitDataSurvivesDecodeFailure verifies that portrait data the current build cannot decode is kept,
// since a different build may be able to decode it and the next save would otherwise drop it; that the failed decode
// is not retried; and that replacing the data gets a fresh attempt.
func TestProfilePortraitDataSurvivesDecodeFailure(t *testing.T) {
	c := check.New(t)
	logged := countLogs(t, slog.LevelError)
	data := []byte("this is not a valid image")
	var p Profile
	p.PortraitData = data
	c.Nil(p.Portrait())
	c.Equal(data, p.PortraitData)
	c.Nil(p.Portrait())
	c.Equal(data, p.PortraitData)
	c.Equal(int32(1), logged.Load(), "the failed decode is logged once, not retried")
	saved, err := jio.Marshal(&p)
	c.NoError(err)
	c.Contains(string(saved), `"portrait":`, "the data is still saved")

	other := []byte("other data that is not a valid image")
	p.SetPortraitData(other)
	c.Equal(other, p.PortraitData, "the new data is stored")
	c.Nil(p.Portrait())
	c.Equal(int32(2), logged.Load(), "replacing the data clears the failed state, so the new data is decoded")
}

// TestProfileSetPortraitDataDiscardsImage verifies that replacing the portrait data drops the image decoded from the
// old data.
func TestProfileSetPortraitDataDiscardsImage(t *testing.T) {
	c := check.New(t)
	var p Profile
	p.PortraitImage = &unison.Image{} // Stands in for an image decoded from earlier data.
	p.SetPortraitData(nil)
	c.Nil(p.Portrait())
}

// TestProfilePortraitExtension verifies that a portrait's extension follows the image format its data holds, and that
// a portrait can be exported only when that format is recognized.
func TestProfilePortraitExtension(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name string
		data []byte
		ext  string
	}{
		{"none", nil, ""},
		{"png", []byte("\x89PNG\r\n\x1a\n"), ".png"},
		{"jpeg", []byte("\xFF\xD8\xFF\xE0"), ".jpg"},
		{"gif", []byte("GIF89a"), ".gif"},
		{"bmp", []byte("BM"), ".bmp"},
		{"webp", []byte("RIFF\x00\x00\x00\x00WEBPVP"), ".webp"},
		{"unrecognized", []byte("not an image"), ""},
	} {
		p := Profile{PortraitData: tc.data}
		c.Equal(tc.ext, p.PortraitExtension(), "%s: extension", tc.name)
		c.Equal(tc.ext != "", p.CanExportPortrait(), "%s: exportable", tc.name)
	}
}

// TestProfileSetAdjustedSizeModifier verifies that setting the adjusted size modifier stores the base value that,
// with the current bonus, yields it.
func TestProfileSetAdjustedSizeModifier(t *testing.T) {
	c := check.New(t)
	p := Profile{SizeModifier: 1, SizeModifierBonus: fxp.Two}
	p.SetAdjustedSizeModifier(5)
	c.Equal(3, p.SizeModifier, "the bonus is taken off")
	c.Equal(5, p.AdjustedSizeModifier())
}

// TestProfileAutoFill verifies that AutoFill takes the tech level and player name from the general settings and gives
// every randomizable field a value.
func TestProfileAutoFill(t *testing.T) {
	c := check.New(t)
	var p Profile
	p.AutoFill(NewEntity())
	general := GlobalSettings().GeneralSettings()
	c.Equal(general.DefaultTechLevel, p.TechLevel)
	c.Equal(general.DefaultPlayerName, p.PlayerName)
	for field, value := range map[string]string{
		"gender":     p.Gender,
		"age":        p.Age,
		"eyes":       p.Eyes,
		"hair":       p.Hair,
		"skin":       p.Skin,
		"handedness": p.Handedness,
		"name":       p.Name,
		"birthday":   p.Birthday,
	} {
		c.NotEqual("", value, "%s is filled in", field)
	}
	c.True(p.Height > 0, "height is filled in")
	c.True(p.Weight > 0, "weight is filled in")
}
