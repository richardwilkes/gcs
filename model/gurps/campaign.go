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
	"bytes"
	"context"
	"hash"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/tid"
)

// Campaign holds the data set to be used for a campaign.
type Campaign struct {
	CampaignData
}

// CampaignData holds the campaign file data.
type CampaignData struct {
	Version       int            `json:"version"`
	ID            tid.TID        `json:"id"`
	SheetSettings *SheetSettings `json:"settings,omitempty"`
	Traits        []*Trait       `json:"traits,omitempty"`
	Skills        []*Skill       `json:"skills,omitempty"`
	Spells        []*Spell       `json:"spells,omitempty"`
	Equipment     []*Equipment   `json:"equipment,omitempty"`
	Notes         []*Note        `json:"notes,omitempty"`
	Templates     []*Template    `json:"templates,omitempty"`
	Characters    []*Entity      `json:"characters,omitempty"`
	Documents     []*Document    `json:"documents,omitempty"`
}

// NewCampaignFromFile loads a Campaign from a file.
func NewCampaignFromFile(fileSystem fs.FS, filePath string) (*Campaign, error) {
	var campaign Campaign
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &campaign); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(campaign.Version); err != nil {
		return nil, err
	}
	return &campaign, nil
}

// NewCampaign creates a new Campaign.
func NewCampaign() *Campaign {
	return &Campaign{
		CampaignData: CampaignData{
			ID:            tid.MustNewTID(kinds.Campaign),
			SheetSettings: GlobalSettings().SheetSettings().Clone(nil),
		},
	}
}

// Save the Campaign to a file as JSON.
func (c *Campaign) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, c)
}

// MarshalJSON implements json.Marshaler.
func (c *Campaign) MarshalJSON() ([]byte, error) {
	c.Version = jio.CurrentDataVersion
	return json.Marshal(&c.CampaignData)
}

// Hash writes this object's contents into the hasher.
func (c *Campaign) Hash(h hash.Hash) {
	var buffer bytes.Buffer
	if err := jio.Save(context.Background(), &buffer, c); err != nil {
		errs.Log(err)
		return
	}
	_, _ = h.Write(buffer.Bytes())
}
