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
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// situationBonus is the shared portion of a bonus that is keyed by a free-form situation description rather than by a
// selector. ConditionalModifierBonus and ReactionBonus embed it; they remain distinct types because the entity and
// the feature editor dispatch on them, and each keeps its own Clone and nil-guarded Hash.
type situationBonus struct {
	Situation string `json:"situation,omitzero"`
	// Group, when not empty, files this bonus under a container of that name in the table that displays it.
	Group string `json:"group,omitzero"`
	LeveledAmount
	BonusOwner `json:"-"`
}

// situationKeyed is implemented by the bonuses that embed situationBonus. It lets the entity collect them generically.
type situationKeyed interface {
	Bonus
	situation() string
	group() string
}

func (s *situationBonus) situation() string {
	return s.Situation
}

func (s *situationBonus) group() string {
	return s.Group
}

// FillWithNameableKeys implements Feature.
func (s *situationBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, s.Situation, s.Group)
}

// SetLeveledOwner implements Bonus.
func (s *situationBonus) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (s *situationBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// hashSituation writes the situation, group and amount into the hasher. The concrete types write their own type and
// switch first, so that the nil guard and the type discriminator stay with them.
func (s *situationBonus) hashSituation(h hash.Hash) {
	xhash.StringWithLen(h, s.Situation)
	xhash.StringWithLen(h, s.Group)
	s.Hash(h) // the embedded LeveledAmount
}

// condModKey identifies one aggregated row: the group it is filed under (empty for none) and its situation. Both have
// already had nameable replacements applied, so two owners that fill "@Target@" in differently land on separate keys.
type condModKey struct {
	group     string
	situation string
}

// condModCollector accumulates the conditional modifiers (or reactions) gathered from an entity's features, merging
// the contributions that share a group and a situation. entityID is the ID of that entity and namespace is the sheet
// block key the rows are destined for; both are mixed into each group container's TID so that a group of the same name
// on another sheet, or in the other table of the same sheet, keeps its own disclosure state.
type condModCollector struct {
	entityID  tid.TID
	namespace string
	entries   map[condModKey]*ConditionalModifier
}

func newCondModCollector(entityID tid.TID, namespace string) *condModCollector {
	return &condModCollector{
		entityID:  entityID,
		namespace: namespace,
		entries:   make(map[condModKey]*ConditionalModifier),
	}
}

// add records amt from source against the given group and situation, creating the row on the first contribution to it.
// Surrounding whitespace is trimmed from the group, so that " Combat" and "Combat" are the same group and a group of
// nothing but spaces is no group at all.
func (c *condModCollector) add(source, group, situation string, amt fxp.Int) {
	key := condModKey{group: strings.TrimSpace(group), situation: situation}
	if existing, ok := c.entries[key]; ok {
		existing.Add(source, amt)
		return
	}
	c.entries[key] = newConditionalModifierInGroup(source, key.group, situation, amt)
}

// rows turns what was collected into the root rows of the table. keep, when not nil, selects the modifiers to retain;
// it is applied before the groups are formed, so a group left without any members is never created rather than shown
// empty. The order produced here is final: the tables that show these rows have their header sorting disabled.
func (c *condModCollector) rows(keep func(*ConditionalModifier) bool) []*ConditionalModifier {
	roots := make([]*ConditionalModifier, 0, len(c.entries))
	groups := make(map[string]*ConditionalModifier)
	for key, one := range c.entries {
		if keep != nil && !keep(one) {
			continue
		}
		if key.group == "" {
			roots = append(roots, one)
			continue
		}
		group, exists := groups[key.group]
		if !exists {
			group = NewConditionalModifierGroup(c.entityID, c.namespace, key.group)
			groups[key.group] = group
			roots = append(roots, group)
		}
		one.SetParent(group)
		group.Children = append(group.Children, one)
	}
	// Both levels have to be sorted, since the map iteration order above is random.
	for _, group := range groups {
		slices.SortFunc(group.Children, compareCondModRows)
	}
	slices.SortFunc(roots, compareCondModRows)
	return roots
}

// situationModifiersFromFeatureList adds the amount of each feature in the list that is a T to the conditional
// modifier for its group and situation in c, creating the modifier if this is the first contribution to that pair. The
// group and situation have the owner's nameable replacements applied first, so that two features reading "@Target@"
// resolve to separate entries when their owners fill the key in differently.
func situationModifiersFromFeatureList[T situationKeyed](source string, features Features, c *condModCollector) {
	for _, f := range features {
		if bonus, ok := f.(T); ok {
			replacements := bonusReplacements(bonus)
			c.add(source, nameable.Apply(bonus.group(), replacements), nameable.Apply(bonus.situation(), replacements),
				bonus.AdjustedAmount())
		}
	}
}
