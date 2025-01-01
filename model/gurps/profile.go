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
	"net/http"
	"os"
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/unison"
)

// ProfileRandom holds the portion of the profile that is affected by the randomizer.
type ProfileRandom struct {
	Name       string     `json:"name,omitempty"`
	Age        string     `json:"age,omitempty"`
	Birthday   string     `json:"birthday,omitempty"`
	Eyes       string     `json:"eyes,omitempty"`
	Hair       string     `json:"hair,omitempty"`
	Skin       string     `json:"skin,omitempty"`
	Handedness string     `json:"handedness,omitempty"`
	Gender     string     `json:"gender,omitempty"`
	Height     fxp.Length `json:"height,omitempty"`
	Weight     fxp.Weight `json:"weight,omitempty"`
}

// Profile holds the profile information for an NPC.
type Profile struct {
	ProfileRandom
	PlayerName        string        `json:"player_name,omitempty"`
	Title             string        `json:"title,omitempty"`
	Organization      string        `json:"organization,omitempty"`
	Religion          string        `json:"religion,omitempty"`
	TechLevel         string        `json:"tech_level,omitempty"`
	PortraitData      []byte        `json:"portrait,omitempty"`
	PortraitImage     *unison.Image `json:"-"`
	SizeModifier      int           `json:"SM,omitempty"`
	SizeModifierBonus fxp.Int       `json:"-"`
}

// Update any derived values.
func (p *Profile) Update(entity *Entity) {
	p.SizeModifierBonus = entity.AttributeBonusFor(SizeModifierID, stlimit.None, nil)
}

// Portrait returns the portrait image, if there is one.
func (p *Profile) Portrait() *unison.Image {
	if p.PortraitImage == nil && len(p.PortraitData) != 0 {
		var err error
		if p.PortraitImage, err = unison.NewImageFromBytes(p.PortraitData, 0.5); err != nil {
			errs.Log(errs.NewWithCause("unable to load portrait data", err))
			p.PortraitImage = nil
			p.PortraitData = nil
			return nil
		}
	}
	return p.PortraitImage
}

// CanExportPortrait returns true if the portrait can be exported.
func (p *Profile) CanExportPortrait() bool {
	return p.PortraitExtension() != ""
}

// PortraitExtension returns the extension for the portrait image.
func (p *Profile) PortraitExtension() string {
	if len(p.PortraitData) == 0 {
		return ""
	}
	switch http.DetectContentType(p.PortraitData) {
	case "image/webp":
		return ".webp"
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	default:
		return ""
	}
}

// ExportPortrait exports the portrait image.
func (p *Profile) ExportPortrait(filePath string) error {
	return errs.Wrap(os.WriteFile(filePath, p.PortraitData, 0o640))
}

// AdjustedSizeModifier returns the adjusted size modifier.
func (p *Profile) AdjustedSizeModifier() int {
	return p.SizeModifier + fxp.As[int](p.SizeModifierBonus)
}

// SetAdjustedSizeModifier sets the adjusted size modifier.
func (p *Profile) SetAdjustedSizeModifier(value int) {
	if value != p.AdjustedSizeModifier() {
		p.SizeModifier = value - fxp.As[int](p.SizeModifierBonus)
	}
}

// AutoFill fills in the default profile entries.
func (p *Profile) AutoFill(entity *Entity) {
	generalSettings := GlobalSettings().GeneralSettings()
	p.TechLevel = generalSettings.DefaultTechLevel
	p.PlayerName = generalSettings.DefaultPlayerName
	p.ApplyRandomizers(entity)
}

// ApplyRandomizers to all randomizable fields, ignoring what may have been there before.
func (p *Profile) ApplyRandomizers(entity *Entity) {
	a := entity.Ancestry()
	p.Gender = a.RandomGender("")
	p.Age = strconv.Itoa(a.RandomAge(entity, p.Gender, 0))
	p.Eyes = a.RandomEyes(p.Gender, "")
	p.Hair = a.RandomHair(p.Gender, "")
	p.Skin = a.RandomSkin(p.Gender, "")
	p.Handedness = a.RandomHandedness(p.Gender, "")
	p.Height = a.RandomHeight(entity, p.Gender, 0)
	p.Weight = a.RandomWeight(entity, p.Gender, 0)
	globalSettings := GlobalSettings()
	generalSettings := globalSettings.GeneralSettings()
	p.Name = a.RandomName(AvailableNameGenerators(globalSettings.Libraries()), p.Gender)
	p.Birthday = generalSettings.CalendarRef(globalSettings.Libraries()).RandomBirthday(p.Birthday)
}
