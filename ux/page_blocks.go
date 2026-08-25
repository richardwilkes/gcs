// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

// newSheetBlockPanel creates the panel for the block with the given key. The ten list blocks are not built here -- the
// sheet, the template and the page exporter each have their own reasons for building those themselves -- so nil comes
// back for one of those, as well as for a key that names no block at all.
func newSheetBlockPanel(key string, entity *gurps.Entity, targetMgr *TargetMgr) unison.Paneler {
	switch key {
	case gurps.BlockPortraitKey:
		return NewPortraitPanel(entity)
	case gurps.BlockIdentityKey:
		return NewIdentityPanel(entity, targetMgr)
	case gurps.BlockMiscellaneousKey:
		return NewMiscPanel(entity, targetMgr)
	case gurps.BlockPointsKey:
		return NewPointsPanel(entity, targetMgr)
	case gurps.BlockDescriptionKey:
		return NewDescriptionPanel(entity, targetMgr)
	case gurps.BlockPrimaryAttributesKey:
		return NewPrimaryAttrPanel(entity, targetMgr)
	case gurps.BlockSecondaryAttributesKey:
		return NewSecondaryAttrPanel(entity, targetMgr)
	case gurps.BlockPointPoolsKey:
		return NewPointPoolsPanel(entity, targetMgr)
	case gurps.BlockBodyKey:
		return NewBodyPanel(entity, targetMgr)
	case gurps.BlockDamageKey:
		return NewDamagePanel(entity, targetMgr)
	case gurps.BlockEncumbranceKey:
		return NewEncumbrancePanel(entity)
	case gurps.BlockLiftingKey:
		return NewLiftingPanel(entity)
	default:
		return nil
	}
}
