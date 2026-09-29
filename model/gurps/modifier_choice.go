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
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// A modifier container is a group, which only organizes its modifiers, or a choice, which also asks for one of them to
// be picked: exactly one when mandatory, at most one when optional. A choice keeps its data in a TemplatePicker, but
// unlike a template choice it never dissolves; its pick is whichever option is enabled.

// ModifierChoiceProvider is implemented by the modifier types and their editor data.
type ModifierChoiceProvider interface {
	// ModifierChoiceData returns a non-nil pointer to the data only a modifier container holds.
	ModifierChoiceData() *ModifierContainerSyncData
}

var _ ModifierChoiceProvider = &ModifierContainerSyncData{}

// ModifierContainerSyncData holds the modifier sync data that is only applicable to modifier containers.
type ModifierContainerSyncData struct {
	// Choice makes the container a choice when it isn't zero (see normalizeModifierChoice).
	Choice TemplatePicker `json:"choice,omitzero"`
}

// ModifierChoiceData implements ModifierChoiceProvider.
func (m *ModifierContainerSyncData) ModifierChoiceData() *ModifierContainerSyncData {
	return m
}

// IsChoice returns true if the container is a choice rather than a group.
func (m *ModifierContainerSyncData) IsChoice() bool {
	return !m.Choice.IsZero()
}

// IsMandatoryChoice returns true if the container is a choice that asks for exactly one of its options.
func (m *ModifierContainerSyncData) IsMandatoryChoice() bool {
	return m.Choice == newModifierChoicePicker(true)
}

// SetMandatoryChoice makes the container a choice that asks for exactly one of its options when mandatory is true, and
// for at most one otherwise.
func (m *ModifierContainerSyncData) SetMandatoryChoice(mandatory bool) {
	m.Choice = newModifierChoicePicker(mandatory)
}

// normalizeModifierChoice clears whatever a group holds in the picker of a choice, and rewrites a choice in the nearer
// of the two supported forms: an exact count of one or more is mandatory, and anything else optional.
func (m *ModifierContainerSyncData) normalizeModifierChoice() {
	if m.IsChoice() {
		m.SetMandatoryChoice(m.Choice.Type.EnsureValid() == picker.Count &&
			m.Choice.Qualifier.Compare.EnsureValid() == criteria.EqualsNumber && m.Choice.Qualifier.Qualifier >= fxp.One)
	} else {
		m.Choice = TemplatePicker{}
	}
}

func (m *ModifierContainerSyncData) hash(h hash.Hash) {
	if m.IsChoice() {
		m.Choice.Hash(h)
		return
	}
	// A group carries no further sync data, so mark it the same way a nil value is marked elsewhere
	xhash.Num8(h, uint8(255))
}

// newModifierChoicePicker returns the picker data of a modifier choice that asks for exactly one of its options when
// mandatory is true, and for at most one otherwise.
func newModifierChoicePicker(mandatory bool) TemplatePicker {
	tp := newTemplateChoicePicker()
	if !mandatory {
		tp.Qualifier.Compare = criteria.AtMostNumber
	}
	return tp
}

// modifierChoiceData returns the container data of the node when it is a modifier choice, or nil otherwise.
func modifierChoiceData[T Node[T]](node T) *ModifierContainerSyncData {
	if xreflect.IsNil(node) || !node.Container() {
		return nil
	}
	if provider, ok := any(node).(ModifierChoiceProvider); ok {
		if data := provider.ModifierChoiceData(); data.IsChoice() {
			return data
		}
	}
	return nil
}

// IsModifierChoice returns true if the node is a modifier choice.
func IsModifierChoice[T Node[T]](node T) bool {
	return modifierChoiceData(node) != nil
}

// IsMandatoryModifierChoice returns true if the node is a modifier choice that asks for exactly one of its options.
func IsMandatoryModifierChoice[T Node[T]](node T) bool {
	data := modifierChoiceData(node)
	return data != nil && data.IsMandatoryChoice()
}

// ModifierChoiceDescription returns a short description of what the modifier choice asks for, "Pick 1" or "Pick at most
// 1", or an empty string if the node isn't one.
func ModifierChoiceDescription[T Node[T]](node T) string {
	if data := modifierChoiceData(node); data != nil {
		return data.Choice.String()
	}
	return ""
}

// ModifierChoiceFor returns the nearest choice above the modifier, which it is an option of, and false if there is none
// or the modifier is a container. Groups in between don't count, so a nested choice owns its own options.
func ModifierChoiceFor[T Node[T]](mod T) (T, bool) {
	if xreflect.IsNil(mod) || mod.Container() {
		var none T
		return none, false
	}
	return modifierChoiceAbove(mod)
}

// ModifierEnabledChanges returns, in order, the modifiers whose enabled state changes when each is set to what want
// gives for it, and the state each ends up in. Turning an option on turns its choice's other options off, the later of
// two wins, and a locked pick (see IsLockedModifierChoiceSelection) and containers are left out.
func ModifierEnabledChanges[T Node[T]](modifiers []T, want func(T) bool) (targets []T, enabled map[T]bool) {
	enabled = make(map[T]bool)
	set := func(node T, on bool) {
		if _, seen := enabled[node]; !seen {
			targets = append(targets, node)
		}
		enabled[node] = on
	}
	isEnabled := func(node T) bool {
		if on, seen := enabled[node]; seen {
			return on
		}
		return node.Enabled()
	}
	for _, node := range modifiers {
		if xreflect.IsNil(node) || node.Container() {
			continue
		}
		on := want(node)
		if !on && IsLockedModifierChoiceSelection(node) {
			continue
		}
		if choice, ok := ModifierChoiceFor(node); ok && on {
			for _, other := range ModifierChoiceOptions(choice) {
				if other != node && isEnabled(other) {
					set(other, false)
				}
			}
		}
		set(node, on)
	}
	// A modifier asked for the state it is already in, or turned on and then off again by a later pick, doesn't change.
	targets = slices.DeleteFunc(targets, func(node T) bool {
		if enabled[node] == node.Enabled() {
			delete(enabled, node)
			return true
		}
		return false
	})
	return targets, enabled
}

// ModifierChoiceOptions returns the options of the modifier choice in order (see ModifierChoiceFor), or nil.
func ModifierChoiceOptions[T Node[T]](choice T) []T {
	if !IsModifierChoice(choice) {
		return nil
	}
	var options []T
	var collect func(list []T)
	collect = func(list []T) {
		for _, one := range list {
			switch {
			case !one.Container():
				options = append(options, one)
			case !IsModifierChoice(one):
				collect(one.NodeChildren())
			}
		}
	}
	collect(choice.NodeChildren())
	return options
}

// CanConvertToModifierChoice returns true if the node is a modifier group that can be converted to a modifier choice.
func CanConvertToModifierChoice[T Node[T]](node T) bool {
	if xreflect.IsNil(node) || !node.Container() || IsModifierChoice(node) {
		return false
	}
	_, ok := any(node).(ModifierChoiceProvider)
	return ok
}

// ConvertToModifierChoice makes the modifier group a mandatory choice, if it can be, and brings it and any choice
// around it into line with the rules (see EnsureModifierChoiceRules).
func ConvertToModifierChoice[T Node[T]](node T) {
	if !CanConvertToModifierChoice(node) {
		return
	}
	any(node).(ModifierChoiceProvider).ModifierChoiceData().SetMandatoryChoice(true) //nolint:errcheck // CanConvertToModifierChoice checked this
	EnsureModifierChoiceRules(node)
	if outer, ok := modifierChoiceAbove(node); ok {
		EnsureModifierChoiceRules(outer)
	}
}

// EnsureModifierChoiceRules settles a modifier choice that just came into being or became mandatory, and on a sheet
// picks its first option when a mandatory one has none.
func EnsureModifierChoiceRules[T Node[T]](choice T) {
	if !IsModifierChoice(choice) {
		return
	}
	SettleModifierChoices(nil, choice)
	if modifierChoiceRequired(choice) {
		SetModifierEnabled(ModifierChoiceOptions(choice)[0], true)
	}
}

// ModifierChoiceIsResolved returns false only for a mandatory modifier choice that has options but none enabled.
func ModifierChoiceIsResolved[T Node[T]](choice T) bool {
	if !IsMandatoryModifierChoice(choice) {
		return true
	}
	options := ModifierChoiceOptions(choice)
	if len(options) == 0 {
		return true
	}
	for _, one := range options {
		if one.Enabled() {
			return true
		}
	}
	return false
}

// UnresolvedModifierChoices returns the unresolved mandatory choices among the modifiers and their children.
func UnresolvedModifierChoices[T Node[T]](modifiers ...T) []T {
	var list []T
	Traverse(func(node T) bool {
		if !ModifierChoiceIsResolved(node) {
			list = append(list, node)
		}
		return false
	}, false, false, modifiers...)
	return list
}

// unresolvedModifierChoiceText names the unresolved mandatory choices among the modifiers of an item on a sheet, or
// returns "" when there are none or the item isn't on a sheet.
func unresolvedModifierChoiceText[T Node[T], M Node[M]](item T, modifiers []M) string {
	if !IsOnSheet(item) {
		return ""
	}
	unresolved := UnresolvedModifierChoices(modifiers...)
	if len(unresolved) == 0 {
		return ""
	}
	names := make([]string, 0, len(unresolved))
	for _, one := range unresolved {
		names = append(names, one.String())
	}
	return i18n.Text("A modifier must be picked for each of these choices: ") + strings.Join(names, ", ")
}

// modifierChoiceRequired returns true if the node is a mandatory modifier choice left unresolved on a sheet.
func modifierChoiceRequired[T Node[T]](node T) bool {
	return IsOnSheet(node) && !ModifierChoiceIsResolved(node)
}

// fillModifierChoiceCell fills in what a modifier's description cell shows about a choice.
func fillModifierChoiceCell[T Node[T]](node T, data *CellData) {
	data.ChoiceInfo = ModifierChoiceDescription(node)
	if data.ChoiceRequired = modifierChoiceRequired(node); data.ChoiceRequired {
		explanation := i18n.Text("One of these modifiers must be picked.")
		if data.Tooltip == "" {
			data.Tooltip = explanation
		} else {
			data.Tooltip = explanation + "\n---\n" + data.Tooltip
		}
	}
}

// IsLockedModifierChoiceSelection returns true if the modifier is the pick of a mandatory choice on a sheet, which can
// be switched to another option but not turned off.
func IsLockedModifierChoiceSelection[T Node[T]](mod T) bool {
	if xreflect.IsNil(mod) || !mod.Enabled() || !IsOnSheet(mod) {
		return false
	}
	choice, ok := ModifierChoiceFor(mod)
	return ok && IsMandatoryModifierChoice(choice)
}

// SettleModifierChoices turns off all but one enabled option of each modifier choice among the nodes and their
// children. It keeps the first that incoming (which may be nil) doesn't report as just arrived, so an existing pick
// survives an option being added, or else the first. Returns true if anything was turned off.
func SettleModifierChoices[T Node[T]](incoming func(T) bool, nodes ...T) bool {
	changed := false
	Traverse(func(node T) bool {
		if !IsModifierChoice(node) {
			return false
		}
		var enabled []T
		for _, one := range ModifierChoiceOptions(node) {
			if one.Enabled() {
				enabled = append(enabled, one)
			}
		}
		if len(enabled) < 2 {
			return false
		}
		keep := enabled[0]
		if incoming != nil {
			if i := slices.IndexFunc(enabled, func(one T) bool { return !incoming(one) }); i != -1 {
				keep = enabled[i]
			}
		}
		for _, one := range enabled {
			if one != keep {
				SetModifierEnabled(one, false)
			}
		}
		changed = true
		return false
	}, false, false, nodes...)
	return changed
}

// SetModifierEnabled sets the enabled state of the node, if it is a modifier.
func SetModifierEnabled[T Node[T]](node T, enabled bool) {
	if gm, ok := any(node).(GeneralModifier); ok {
		gm.SetEnabled(enabled)
	}
}

// ConvertFromModifierChoice converts the modifier choice to a group. Within another choice, that one keeps its pick
// over the options it gains.
func ConvertFromModifierChoice[T Node[T]](node T) {
	data := modifierChoiceData(node)
	if data == nil {
		return
	}
	data.Choice = TemplatePicker{}
	settleModifierChoicesAround(node)
}

// settleModifierChoicesAround settles the choices within the container, which may have just become or stopped being a
// choice, and the choice around it, which keeps its own pick over any the container handed it.
func settleModifierChoicesAround[T Node[T]](container T) {
	SettleModifierChoices(nil, container)
	if outer, ok := modifierChoiceAbove(container); ok {
		SettleModifierChoices(func(one T) bool { return isWithin(one, container) }, outer)
	}
}

// modifierChoiceAbove returns the nearest modifier choice above the node, if any.
func modifierChoiceAbove[T Node[T]](node T) (T, bool) {
	for parent := node.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
		if IsModifierChoice(parent) {
			return parent, true
		}
	}
	var none T
	return none, false
}

// isWithin returns true if the node is held, at any depth, by the container.
func isWithin[T Node[T]](node, container T) bool {
	for parent := node.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
		if parent == container {
			return true
		}
	}
	return false
}

// maxModifierChoiceVariants caps the ways of making open choices that modifierChoiceRange works through, so that a
// pathological file can't lock up the display; past it, choices count as made with the picks they have. A trait
// container's choices multiply with those of each trait inside, and the cap covers them together.
const maxModifierChoiceVariants = 4096

// IsOnSheet returns true if the node belongs to a character or loot sheet, where every modifier choice is made.
func IsOnSheet[T Node[T]](node T) bool {
	return !xreflect.IsNil(node) && IsSheetOwner(node.DataOwner())
}

// IsSheetOwner returns true if the data owner is a character sheet or a loot sheet (see IsOnSheet).
func IsSheetOwner(owner DataOwner) bool {
	if xreflect.IsNil(owner) {
		return false
	}
	if owner.OwningEntity() != nil {
		return true
	}
	_, isLoot := owner.(*Loot)
	return isLoot
}

// modifierChoicePicks maps each mandatory modifier choice settled on while costing to its pick, or to nil when it
// counts as made with the picks it has. A trait container fixes these for every trait inside it.
type modifierChoicePicks[M comparable] map[M]M

// openMandatoryModifierChoices returns the mandatory choices among the modifiers still to be made that fixed holds no
// pick for, and the ways of making them, capped at one past maxModifierChoiceVariants. Off a sheet a choice is open
// unless it has a pick and what it is asked about on is preconfigured; a choice with no options never is.
func openMandatoryModifierChoices[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, modifiers []M, fixed modifierChoicePicks[M]) (open []M, variants int) {
	variants = 1
	if xreflect.IsNil(item) || IsOnSheet(item) {
		return nil, variants
	}
	Traverse(func(mod M) bool {
		if _, isFixed := fixed[mod]; isFixed || !IsMandatoryModifierChoice(mod) ||
			(IsNodePreconfigured(modifierAskedAboutOn(item, mod)) && ModifierChoiceIsResolved(mod)) {
			return false
		}
		if options := ModifierChoiceOptions(mod); len(options) != 0 {
			open = append(open, mod)
			variants = min(variants*len(options), maxModifierChoiceVariants+1)
		}
		return false
	}, false, false, modifiers...)
	return open, variants
}

// modifierAskedAboutOn returns the container the item inherits the modifier from, which is asked about it, or else
// the item. The modifier's target can't simply be used, since an editor works on a copy of the item whose own
// modifiers still point at the original.
func modifierAskedAboutOn[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, mod M) T {
	if target := mod.Target(); !xreflect.IsNil(target) && isWithin(item, target) {
		return target
	}
	return item
}

// hasOpenMandatoryModifierChoice returns true if the item has an open mandatory modifier choice within the cap.
func hasOpenMandatoryModifierChoice[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, modifiers []M) bool {
	open, variants := openMandatoryModifierChoices(item, modifiers, nil)
	return len(open) != 0 && variants <= maxModifierChoiceVariants
}

// modifierChoiceRange returns the span of eval over each way of making the open mandatory choices among the modifiers,
// with those in fixed made as it says. eval gets the non-container modifiers in order, with only each such choice's
// pick among its options, enabled. Returns false, without calling eval, when there is nothing to enumerate.
func modifierChoiceRange[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, modifiers []M, fixed modifierChoicePicks[M], eval func([]M) NumericRange) (NumericRange, bool) {
	open, variants := openMandatoryModifierChoices(item, modifiers, fixed)
	if variants > maxModifierChoiceVariants {
		open = nil
	}
	// The choice each option of an open choice, or of one fixed holds a pick for, is an option of.
	choiceOf := make(map[M]M)
	var leaves []M
	Traverse(func(mod M) bool {
		leaves = append(leaves, mod)
		if choice, ok := ModifierChoiceFor(mod); ok {
			if _, isFixed := fixed[choice]; isFixed || slices.Contains(open, choice) {
				choiceOf[mod] = choice
			}
		}
		return false
	}, false, true, modifiers...)
	if len(choiceOf) == 0 {
		return NumericRange{}, false
	}
	variant := make([]M, 0, len(leaves))
	ranges := make([]NumericRange, 0, variants)
	eachModifierChoicePick(open, fixed, func(picks modifierChoicePicks[M]) {
		variant = variant[:0]
		for _, one := range leaves {
			choice, isOption := choiceOf[one]
			switch {
			case !isOption || xreflect.IsNil(picks[choice]):
				variant = append(variant, one)
			case picks[choice] == one:
				variant = append(variant, one.enabledVariant())
			}
		}
		ranges = append(ranges, eval(variant))
	})
	return rangeForPickerByCount(newTemplateChoicePicker().Qualifier, ranges), true
}

// eachModifierChoicePick calls fn with fixed plus each combination of picks for the choices, or just once with fixed
// when there are none. The map is reused between calls.
func eachModifierChoicePick[M Node[M]](choices []M, fixed modifierChoicePicks[M], fn func(picks modifierChoicePicks[M])) {
	options := make([][]M, len(choices))
	for i, one := range choices {
		options[i] = ModifierChoiceOptions(one)
	}
	picks := make(modifierChoicePicks[M], len(fixed)+len(choices))
	maps.Copy(picks, fixed)
	which := make([]int, len(choices))
	for {
		for i, one := range choices {
			picks[one] = options[i][which[i]]
		}
		fn(picks)
		// Advance to the next way of making the choices, stopping once every one has been tried.
		i := 0
		for ; i < len(which); i++ {
			which[i]++
			if which[i] < len(options[i]) {
				break
			}
			which[i] = 0
		}
		if i == len(which) {
			return
		}
	}
}
