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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// A modifier container is either a group, which only organizes the modifiers it holds, or a choice, which is a group
// that also asks for one of the modifiers it holds to be picked. A choice is either mandatory, asking for exactly one,
// or optional, asking for at most one. A mandatory choice is how a trait or piece of equipment whose price varies is
// given its price, and an optional one is how a set of modifiers is made mutually exclusive.
//
// A choice keeps its data in the same TemplatePicker type a template choice does, but unlike a template choice it is
// never dissolved: its options stay attached, and the choice is made by which of them are enabled, wherever the
// modifiers are. Only a count of exactly one or at most one is supported, which is all the data is ever allowed to
// hold.

// ModifierChoiceProvider is implemented by the modifier types and their editor data, since any of their containers may
// be a choice.
type ModifierChoiceProvider interface {
	// ModifierChoiceData returns a non-nil pointer to the data only a modifier container holds.
	ModifierChoiceData() *ModifierContainerSyncData
}

var _ ModifierChoiceProvider = &ModifierContainerSyncData{}

// ModifierContainerSyncData holds the modifier sync data that is only applicable to modifier containers.
type ModifierContainerSyncData struct {
	// Choice makes the container a choice when it isn't zero. It only ever holds one of the two forms a modifier
	// choice supports (see normalizeModifierChoice).
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

// IsMandatoryChoice returns true if the container is a choice that asks for exactly one of its options. A choice held
// in a form this version doesn't support is read as the nearer of the two that are: an exact count as exactly one, and
// any other as at most one, the choice that asks the least of whoever makes it.
func (m *ModifierContainerSyncData) IsMandatoryChoice() bool {
	return m.IsChoice() && m.Choice.Qualifier.Compare.EnsureValid() == criteria.EqualsNumber
}

// SetMandatoryChoice makes the container a choice that asks for exactly one of its options when mandatory is true, and
// for at most one otherwise.
func (m *ModifierContainerSyncData) SetMandatoryChoice(mandatory bool) {
	m.Choice = newModifierChoicePicker(mandatory)
}

// isSupportedChoice returns true if the container is a choice held in one of the two forms this version supports.
func (m *ModifierContainerSyncData) isSupportedChoice() bool {
	return m.IsChoice() && m.Choice == newModifierChoicePicker(m.IsMandatoryChoice())
}

// normalizeModifierChoice clears whatever a group holds in the picker of a choice. A choice's picker is kept as it is,
// even in a form this version doesn't support, so that a newer one's choice survives being loaded and saved here: it is
// read as the nearer of the two forms that are supported (see IsMandatoryChoice), and only rewritten to one of them
// when the choice is edited.
func (m *ModifierContainerSyncData) normalizeModifierChoice() {
	if !m.IsChoice() {
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
		// Described as the form it is read as, which a choice held in another form may not be.
		return newModifierChoicePicker(data.IsMandatoryChoice()).String()
	}
	return ""
}

// ModifierChoiceFor returns the modifier choice the modifier is an option of: the nearest container above it that is a
// choice. A group only organizes what it holds, so a modifier in a group within a choice is still one of the choice's
// options, while a choice within a choice has options of its own. The second return is false when the modifier isn't
// an option of any choice, which a container never is.
func ModifierChoiceFor[T Node[T]](mod T) (T, bool) {
	if xreflect.IsNil(mod) || mod.Container() {
		var none T
		return none, false
	}
	return modifierChoiceAbove(mod)
}

// ModifierEnabledChanges works out every modifier whose enabled state changes when each of the modifiers is set to the
// state want gives for it, and returns them, in order, along with the state each ends up in. Turning on an option of a
// choice also turns off the choice's other options, since a choice never has more than one enabled, and when two
// options of one choice are turned on together the later one wins. The pick of a mandatory choice on a sheet can't be
// turned off (see IsLockedModifierChoiceSelection), so a request to do that is left out. Containers are always enabled,
// so they are left out too.
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
	return targets, enabled
}

// KeepModifierChoiceRules applies the rules of the choice the modifier is an option of after something set the
// modifier's enabled state directly, as its editor does: an option that is now enabled becomes the pick, turning off
// the choice's others, and the pick of a mandatory choice on a sheet, which can't be turned off, is turned back on.
// wasEnabled is the modifier's state before it was set.
func KeepModifierChoiceRules[T Node[T]](mod T, wasEnabled bool) {
	choice, ok := ModifierChoiceFor(mod)
	if !ok {
		return
	}
	if mod.Enabled() {
		for _, other := range ModifierChoiceOptions(choice) {
			if other != mod && other.Enabled() {
				SetModifierEnabled(other, false)
			}
		}
		return
	}
	if wasEnabled && IsOnSheet(mod) && IsMandatoryModifierChoice(choice) && !ModifierChoiceIsResolved(choice) {
		SetModifierEnabled(mod, true)
	}
}

// ModifierChoiceOptions returns the options of the modifier choice, in the order they are listed: the modifiers beneath
// it that are its own options (see ModifierChoiceFor). Returns nil when the node isn't a modifier choice.
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

// ConvertToModifierChoice converts the modifier group to a modifier choice, if it can be. The new choice is a
// mandatory one. Nothing a group holds is lost, since a choice holds all of it too, but no more than one of its options
// can stay enabled, so only the first that is keeps that. On a sheet a mandatory choice must be made, so
// when none of the options is enabled there, the first is.
func ConvertToModifierChoice[T Node[T]](node T) {
	if !CanConvertToModifierChoice(node) {
		return
	}
	any(node).(ModifierChoiceProvider).ModifierChoiceData().SetMandatoryChoice(true) //nolint:errcheck // CanConvertToModifierChoice checked this
	EnsureModifierChoiceRules(node)
	// The options the group held were options of the choice around it, which may have had its pick among them.
	if outer, ok := modifierChoiceAbove(node); ok {
		EnsureModifierChoiceRules(outer)
	}
}

// EnsureModifierChoiceRules brings a modifier choice that has just come into being, or just become mandatory, into line
// with the rules of a choice: no more than one of its options stays enabled, and on a sheet, where a mandatory choice
// must have its pick, the first of its options is picked when none is. A node that isn't a modifier choice is left
// alone.
func EnsureModifierChoiceRules[T Node[T]](choice T) {
	if !IsModifierChoice(choice) {
		return
	}
	SettleModifierChoices(nil, choice)
	if IsOnSheet(choice) && !ModifierChoiceIsResolved(choice) {
		SetModifierEnabled(ModifierChoiceOptions(choice)[0], true)
	}
}

// ModifierChoiceIsResolved returns true unless the node is a mandatory modifier choice none of whose options is
// enabled. Such a choice is unresolved: it is allowed anywhere but on a sheet, where the choice has to have
// been made. A choice with no options has nothing to pick from, so it can't be left unresolved.
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

// UnresolvedModifierChoices returns the unresolved mandatory choices among the modifiers and their children (see
// ModifierChoiceIsResolved).
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

// unresolvedModifierChoiceText explains which mandatory choices among the modifiers of an item on a sheet
// have yet to be made, the one place that isn't allowed, or returns an empty string when there are none or the item
// isn't on a sheet.
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

// modifierChoiceRequired returns true if the node is a mandatory modifier choice left unresolved on a sheet,
// the one place that isn't allowed.
func modifierChoiceRequired[T Node[T]](node T) bool {
	return IsOnSheet(node) && !ModifierChoiceIsResolved(node)
}

// fillModifierChoiceCell fills in what the description cell of a modifier shows about a choice: what it asks for, and
// whether it is required and yet to be made, which is explained ahead of the rest of its tooltip.
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

// IsLockedModifierChoiceSelection returns true if the modifier is the option picked for a mandatory choice on a sheet.
// It can't be turned off there, since that would leave the choice unresolved; picking another option is how the choice
// is changed.
func IsLockedModifierChoiceSelection[T Node[T]](mod T) bool {
	if xreflect.IsNil(mod) || !mod.Enabled() || !IsOnSheet(mod) {
		return false
	}
	choice, ok := ModifierChoiceFor(mod)
	return ok && IsMandatoryModifierChoice(choice)
}

// settleLoadedModifierChoices keeps each modifier choice among the nodes and their children that is held in a form this
// version supports to no more than one enabled option, as SettleModifierChoices does. A choice held in another form, as
// a newer version may have saved it, keeps its options as they are, since that version may allow it more than one.
func settleLoadedModifierChoices[T Node[T]](nodes ...T) {
	Traverse(func(node T) bool {
		if data := modifierChoiceData(node); data != nil && data.isSupportedChoice() {
			settleModifierChoice(nil, node)
		}
		return false
	}, false, false, nodes...)
}

// SettleModifierChoices turns off all but one of the enabled options of each modifier choice among the nodes and their
// children, since a choice never has more than one enabled. The one kept is the first enabled option that incoming
// doesn't report as having just arrived, so that the pick a choice already had survives an option being added, moved
// or pasted into it, or failing that the first enabled option. A nil incoming reports nothing as having arrived.
// Returns true if anything was turned off.
func SettleModifierChoices[T Node[T]](incoming func(T) bool, nodes ...T) bool {
	changed := false
	Traverse(func(node T) bool {
		if IsModifierChoice(node) && settleModifierChoice(incoming, node) {
			changed = true
		}
		return false
	}, false, false, nodes...)
	return changed
}

// settleModifierChoice turns off all but one of the enabled options of the modifier choice, as SettleModifierChoices
// does, and returns true if anything was turned off.
func settleModifierChoice[T Node[T]](incoming func(T) bool, choice T) bool {
	var enabled []T
	for _, one := range ModifierChoiceOptions(choice) {
		if one.Enabled() {
			enabled = append(enabled, one)
		}
	}
	if len(enabled) < 2 {
		return false
	}
	keep := enabled[0]
	if incoming != nil {
		for _, one := range enabled {
			if !incoming(one) {
				keep = one
				break
			}
		}
	}
	for _, one := range enabled {
		if one != keep {
			SetModifierEnabled(one, false)
		}
	}
	return true
}

// SetModifierEnabled sets the enabled state of the node, if it is a modifier.
func SetModifierEnabled[T Node[T]](node T, enabled bool) {
	if gm, ok := any(node).(GeneralModifier); ok {
		gm.SetEnabled(enabled)
	}
}

// ConvertFromModifierChoice converts the modifier choice to a modifier group, discarding its choice. Within another
// choice, its options become options of that one, which keeps the pick it already had over any of theirs.
func ConvertFromModifierChoice[T Node[T]](node T) {
	data := modifierChoiceData(node)
	if data == nil {
		return
	}
	data.Choice = TemplatePicker{}
	settleModifierChoicesAround(node)
}

// settleModifierChoicesAround settles the modifier choices within the container and the choice around it, if any, once
// the container may have become a choice or stopped being one. One that stopped being a choice has handed its options
// to the choice around it, which keeps the pick it already had over any of theirs.
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

// maxModifierChoiceVariants is the most ways of making the open mandatory choices of a set of modifiers that
// modifierChoiceRange will work through. No real trait or piece of equipment comes anywhere near it; it is there so
// that a pathological file can't lock up the display. Past it, the choices are treated as made with the picks they
// have, so that every figure agrees with every other.
const maxModifierChoiceVariants = 4096

// IsOnSheet returns true if the node belongs to a character sheet or a loot sheet. That is where every modifier choice
// has been made, and so where a mandatory one may not be left without its pick. A node elsewhere, or with no owner at
// all, isn't on a sheet.
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

// openMandatoryModifierChoices returns the options of each mandatory choice among the modifiers of the item being
// costed that has yet to be made, or nothing when there are too many ways of making them to work through (see
// maxModifierChoiceVariants). On a sheet every choice has been made. Elsewhere a mandatory choice is asked about when
// the item reaches a sheet, and so is still to be made, unless the item is marked preconfigured and the choice already
// has its pick, which is then taken without asking. A choice with no options has nothing to make.
func openMandatoryModifierChoices[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, modifiers []M) [][]M {
	if xreflect.IsNil(item) || IsOnSheet(item) {
		return nil
	}
	var open [][]M
	Traverse(func(mod M) bool {
		if !IsMandatoryModifierChoice(mod) || (IsNodePreconfigured(modifierAskedAboutOn(item, mod)) &&
			ModifierChoiceIsResolved(mod)) {
			return false
		}
		if options := ModifierChoiceOptions(mod); len(options) != 0 {
			open = append(open, options)
		}
		return false
	}, false, false, modifiers...)
	count := 1
	for _, options := range open {
		if count *= len(options); count > maxModifierChoiceVariants {
			return nil
		}
	}
	return open
}

// modifierAskedAboutOn returns what the modifier, one of those the item being costed is subject to, is asked about on:
// the container above the item it belongs to when the item inherits it from one, since a container is asked about its
// own modifiers, and the item itself otherwise. Whether that is preconfigured decides whether the modifier's choice is
// taken as made. The modifier's target can't simply be used, since an editor works on a copy of the item whose own
// modifiers still point at the original.
func modifierAskedAboutOn[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, mod M) T {
	if target := mod.Target(); !xreflect.IsNil(target) && target != item {
		for parent := item.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
			if parent == target {
				return target
			}
		}
	}
	return item
}

// hasOpenMandatoryModifierChoice returns true if a mandatory choice among the modifiers of the item being costed has
// yet to be made.
func hasOpenMandatoryModifierChoice[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, modifiers []M) bool {
	return len(openMandatoryModifierChoices(item, modifiers)) != 0
}

// modifierChoiceRange returns the span of what eval reports for each way the open mandatory choices among the
// modifiers of the item being costed can be made. eval is handed the modifiers as they would stand with the choices
// made: a flat list of the modifiers that aren't containers, in their usual order, with just the option picked from
// each open choice among them, enabled. The second return is false, and eval is never called, when there is no open
// choice to make, or too many ways of making them (see maxModifierChoiceVariants). Wherever a single number is needed
// instead, an open choice counts as the least it may come to.
func modifierChoiceRange[M ModifierNode[M, T], T ModifiableNode[T, M]](item T, modifiers []M, eval func([]M) NumericRange) (NumericRange, bool) {
	choices := openMandatoryModifierChoices(item, modifiers)
	if len(choices) == 0 {
		return NumericRange{}, false
	}
	count := 1
	for _, options := range choices {
		count *= len(options)
	}
	optionOf := make(map[M]int)
	for i, options := range choices {
		for _, one := range options {
			optionOf[one] = i
		}
	}
	var leaves []M
	Traverse(func(mod M) bool {
		leaves = append(leaves, mod)
		return false
	}, false, true, modifiers...)
	picks := make([]int, len(choices))
	variant := make([]M, 0, len(leaves))
	ranges := make([]NumericRange, 0, count)
	for {
		variant = variant[:0]
		for _, one := range leaves {
			if i, isOption := optionOf[one]; !isOption {
				variant = append(variant, one)
			} else if choices[i][picks[i]] == one {
				variant = append(variant, one.enabledVariant())
			}
		}
		ranges = append(ranges, eval(variant))
		// Advance to the next way of making the choices, stopping once every one has been tried.
		i := 0
		for ; i < len(picks); i++ {
			picks[i]++
			if picks[i] < len(choices[i]) {
				break
			}
			picks[i] = 0
		}
		if i == len(picks) {
			break
		}
	}
	return spanOfNumericRanges(ranges), true
}
