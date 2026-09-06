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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// bonusTagsCriteria returns the tags criteria of one of the bonus types that migrated from "category" to "tags".
func bonusTagsCriteria(c check.Checker, f Feature) criteria.Text {
	switch b := f.(type) {
	case *SkillBonus:
		return b.TagsCriteria
	case *SkillPointBonus:
		return b.TagsCriteria
	case *SpellBonus:
		return b.TagsCriteria
	case *SpellPointBonus:
		return b.TagsCriteria
	case *WeaponBonus:
		return b.TagsCriteria
	default:
		c.Errorf("unexpected feature type %T", f)
		return criteria.Text{}
	}
}

// loadSingleFeature loads a feature list holding a single feature from its JSON and returns that feature.
func loadSingleFeature(c check.Checker, data string) Feature {
	var loaded Features
	c.NoError(jio.Unmarshal([]byte(data), &loaded), "%s: should load", data)
	c.Equal(1, len(loaded), "%s: a single feature should load", data)
	if len(loaded) != 1 {
		return nil
	}
	return loaded[0]
}

// TestBonusLegacyCategoryMigratesToTags verifies that each bonus type that once stored its tags criteria under
// "category" adopts a legacy "category" when it has no "tags", and prefers "tags" when both are present.
func TestBonusLegacyCategoryMigratesToTags(t *testing.T) {
	c := check.New(t)
	for _, typ := range []feature.Type{
		feature.SkillBonus,
		feature.SkillPointBonus,
		feature.SpellBonus,
		feature.SpellPointBonus,
		feature.WeaponBonus,
	} {
		key := typ.Key()
		f := loadSingleFeature(c, `[{"type":"`+key+`","category":{"compare":"is","qualifier":"Old"}}]`)
		if f == nil {
			continue
		}
		c.Equal(typ, f.FeatureType(), "%s: type survives the load", key)
		c.Equal(textCriteria(criteria.IsText, "Old"), bonusTagsCriteria(c, f),
			"%s: a legacy category becomes the tags criteria", key)

		f = loadSingleFeature(c, `[{"type":"`+key+`","tags":{"compare":"contains","qualifier":"New"},"category":{"compare":"is","qualifier":"Old"}}]`)
		if f == nil {
			continue
		}
		c.Equal(textCriteria(criteria.ContainsText, "New"), bonusTagsCriteria(c, f),
			"%s: tags win over a legacy category", key)

		f = loadSingleFeature(c, `[{"type":"`+key+`"}]`)
		if f == nil {
			continue
		}
		c.True(bonusTagsCriteria(c, f).IsZero(), "%s: no tags and no category leaves the tags criteria empty", key)
	}
}

// TestWeaponBonusLegacyPerLevel verifies that a weapon bonus still migrates the legacy "per_level" flag to PerDie
// alongside its legacy category.
func TestWeaponBonusLegacyPerLevel(t *testing.T) {
	c := check.New(t)
	f := loadSingleFeature(c, `[{"type":"`+feature.WeaponBonus.Key()+`","per_level":true,"category":{"compare":"is","qualifier":"Old"}}]`)
	if f == nil {
		return
	}
	bonus, ok := f.(*WeaponBonus)
	c.True(ok, "loads as a weapon bonus")
	if !ok {
		return
	}
	c.True(bonus.PerDie, "a legacy per_level flag becomes PerDie")
	c.Equal(textCriteria(criteria.IsText, "Old"), bonus.TagsCriteria, "the legacy category is migrated as well")
}
