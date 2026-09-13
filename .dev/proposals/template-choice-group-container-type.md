# Proposal: "Fixed Cost" Trait Container Type

## Status

Draft.

## Problem

Trait containers currently derive their point cost in one of a few fixed ways
(see `Trait.AdjustedPoints` and `container.Type` in
[trait.go](../../model/gurps/trait.go)):

- `Group` / default: sum of children's `AdjustedPoints()`.
- `AlternativeAbilities`: most expensive `AlternativeSlots` children billed in
  full, remainder at 20%.
- `Ancestry`, `Attributes`, `MetaTrait`: same summation as `Group`, but tagged
  differently for `PointsBreakdown` bucketing in
  [entity.go](../../model/gurps/entity.go) (`calculateSingleTraitPoints`).

There's no way to author a container whose point cost is a single fixed
number regardless of its children's individual costs (e.g. "this package of
abilities costs a flat 15 points, however it's organized"), nor one whose
cost is driven by a `TemplatePicker` "Points" choice value ("pick N points
worth of gear, and *that* choice value is what this container costs" as
opposed to being the sum of the picked children).

## Goal

Add a new `container.Type` named `FixedCost` whose
`AdjustedPoints()` is:

1. A manually-set fixed value, if the user has configured one, **or**
2. The resolved value of the container's own `TemplatePicker` when its
   `Type` is `picker.Points`, **or**
3. Falls back to the normal sum-of-children behavior if neither applies.

Manual override always wins over the template-picker-derived value ("manual
override > calculated value" per the request). Children remain normal,
browsable/editable rows regardless of which of the three sources is active —
only the parent's own `AdjustedPoints()` computation changes, exactly as
`AlternativeAbilities` children remain normal rows today even though only a
subset of them contribute to the parent's total at full cost.

`FixedCost` is a template-authoring feature only: it must not be a usable
container type on a character sheet (see "Templates-Only Availability"
below), and its point-derivation must stay consistent with the existing
template-picker validation UI at the moment a character is created from a
template (see "Consistency with Picker Validation" below).

## Data Model Changes

### `container` enum

Add a new value in `cmd/enumgen/main.go`'s `allEnums` entry for
`model/gurps/enums/container` (this is the generator input; `type_gen.go` is
generated — do not hand-edit it):

```go
{Key: "fixed_cost"},
```

then regenerate with `go generate ./cmd/enumgen/main.go`.

### `TraitContainerSyncData` (in [trait.go](../../model/gurps/trait.go))

```go
type TraitContainerSyncData struct {
	Ancestry         string         `json:"ancestry,omitzero"`
	TemplatePicker   TemplatePicker `json:"template_picker,omitzero"`
	ContainerType    container.Type `json:"container_type,omitzero"`
	AlternativeSlots int            `json:"alternative_slots,omitzero"`
	FixedPoints      *fxp.Int       `json:"fixed_points,omitempty"` // new
}
```

`FixedPoints` is a pointer so "unset" (fall through to picker/children) is
distinguishable from an explicit `0`. Follows the existing precedent of
`Trait.savedCurrentLevel *fxp.Int` and `calc.CurrentLevel *fxp.Int` for
optional-int-that-can-legitimately-be-zero fields.

Only meaningful when `ContainerType == container.FixedCost`; it must be
cleared by `ClearUnusedFieldsForType` for every other container type and for
non-containers, matching how `AlternativeSlots` and `Ancestry` are already
reset there.

### `Trait.AdjustedPoints()`

Add a branch before the existing `AlternativeAbilities` special case:

```go
if t.ContainerType == container.FixedCost {
	if t.FixedPoints != nil {
		return *t.FixedPoints
	}
	if t.TemplatePicker.Type == picker.Points && t.TemplatePicker.Qualifier.Compare == criteria.EqualsNumber {
		return t.TemplatePicker.Qualifier.Qualifier
	}
	// fall through to sum-of-children below
}
```

The picker-derived value is only used when `Qualifier.Compare == criteria.EqualsNumber`
("Pick exactly N points"). `Qualifier` is a `criteria.Number`: a value paired with a
comparison (`is`/`is not`/`at least`/`at most`), used to validate a choice, not to state
it. For `AtLeastNumber`/`AtMostNumber`/`NotEqualsNumber`, the qualifier's number is only a
bound checked when the pick was made, not the actual total of what was picked (which may
be higher or lower), so those cases fall through to the normal sum-of-children
computation. This mirrors the existing rule in
[ux/template.go](../../ux/template.go)'s `rawPoints` helper, which trusts a container's
`picker.Points` qualifier only under the same `EqualsNumber` condition.

### Consistency with Picker Validation

Restricting the picker-derived branch to `EqualsNumber` is also what keeps this feature
safe to combine with template-choice validation, and what makes it behave correctly at
character creation:

- While a template is being applied, [ux/template_picker.go](../../ux/template_picker.go)'s
  `processPickerRow` already refuses to let the user close the picker dialog unless
  `tp.Qualifier.Matches(total)` is true (the OK button is disabled otherwise), where `total`
  is the actual summed cost of the checked children. For an `EqualsNumber` picker, this
  guarantees the checked children's total *equals* `tp.Qualifier.Qualifier` by the time the
  dialog can be dismissed — the same value `AdjustedPoints()` would derive from the picker.
  No extra validation is required here; it already exists and cannot be bypassed.
- Once the pick is resolved, the container is copied onto the character's sheet and
  (per "Templates-Only Availability" below) its `ContainerType` reverts from `FixedCost`
  to `Group`. From that point on the container's cost is simply the sum of the (now
  pruned-down-to-the-picked-set) children, which is the same total the dialog validated.
  So the fixed/picker-derived cost that applied while the container lived in the template
  and the summed cost it has after landing on the sheet always agree; there is no window
  where the two disagree.
- A manually-set `FixedPoints` intentionally is not required to agree with the picker's
  qualifier or the children's sum — the override always wins per the Goal above, and can
  legitimately be used with no `TemplatePicker` at all, or with one that constrains only
  the *children* selection process, independent of the flat price the author wants to
  charge for the package.

### Hashing

`TraitContainerSyncData.hash` in [trait.go](../../model/gurps/trait.go) must
hash `FixedPoints` (nil vs pointer-to-zero must hash differently) when
`ContainerType == container.FixedCost`, mirroring the existing
`Ancestry` / `AlternativeSlots` cases:

```go
case container.FixedCost:
	if t.FixedPoints != nil {
		xhash.Bool(h, true)
		xhash.Num64(h, *t.FixedPoints)
	} else {
		xhash.Bool(h, false)
	}
```

### `PointsBreakdown` bucketing

`calculateSingleTraitPoints` in [entity.go](../../model/gurps/entity.go)
currently only special-cases `Group`, `Ancestry`, and `Attributes` for
recursion/bucketing; `AlternativeAbilities`, `MetaTrait`, and the new
`FixedCost` type fall through to the default per-trait bucketing (i.e.
they are billed like a single trait, not decomposed into children). No entity
changes should be required here — the new type is expected to behave the same
way `MetaTrait`/`AlternativeAbilities` do today: `AdjustedPoints()` is called
once on the container and the result is bucketed as one item. Verify this
assumption while implementing rather than assuming it holds for every
bucketing consumer (also check
[list_filter_fields.go](../../model/gurps/list_filter_fields.go),
[export.go](../../model/gurps/export.go),
[export_legacy.go](../../model/gurps/export_legacy.go), and
[scripting_trait.go](../../model/gurps/scripting_trait.go), which key off
`ContainerType.Key()`/`.String()` generically and shouldn't need changes
beyond the generated enum gaining the new case).

## Templates-Only Availability

`FixedCost` must only be usable while a trait lives in a `Template`
([template.go](../../model/gurps/template.go)) or a library; it must never persist as the
`ContainerType` of a trait attached to a character `Entity`
([entity.go](../../model/gurps/entity.go)). Both types already implement `DataOwner`
([node.go](../../model/gurps/node.go)), and are told apart by `OwningEntity()`:
`Template.OwningEntity()` always returns `nil`, while `Entity.OwningEntity()` returns
itself (non-nil). That is the existing, reliable signal for "this data now belongs to a
character," and needs no new plumbing.

Enforcement needs two parts:

1. **Automatic reversion when a trait lands on a character.** Add a small helper, e.g.
   `(t *Trait) demoteFixedCostForOwner(owner DataOwner)`, that sets
   `t.ContainerType = container.Group` and `t.FixedPoints = nil` when
   `t.ContainerType == container.FixedCost && owner != nil && owner.OwningEntity() != nil`.
   Call it from both:
   - `Trait.Clone` ([trait.go](../../model/gurps/trait.go)), right after `other` is built and
     before its children are cloned (each child that is itself a `FixedCost` container will
     be corrected by its own recursive `Clone` call using the same `owner`). This is the path
     used by dragging/copying a trait from a library or template onto a sheet
     (`Node.CloneForTarget` in [ux/table_node.go](../../ux/table_node.go)) and by
     `applyTemplateToSheetWithPickers`'s `cloneRows` in
     [ux/template.go](../../ux/template.go), which is exactly the "apply template to a
     character" / character-creation path this needs to cover.
   - `Trait.SetDataOwner` ([trait.go](../../model/gurps/trait.go)), so that a trait
     reassigned to a different owner after the fact (e.g. hand-edited save data, or any
     future code path that reparents rows without going through `Clone`) is normalized the
     next time it is attached, rather than only at load or apply time.
2. **Hide the option in the editor.** [ux/trait_editor.go](../../ux/trait_editor.go)
   currently passes the full `container.Types` list to the "Container Type" popup
   unconditionally. When editing a trait whose `DataOwner().OwningEntity() != nil` (i.e. a
   trait that already belongs to a character sheet), the popup's option list must omit
   `container.FixedCost`, so a character's traits can never have it selected in the first
   place. Templates and library files (`OwningEntity() == nil`) continue to see the full
   list.

Combined, these two changes mean: `FixedCost` can be authored and combined with a
`TemplatePicker` only in a template or library, is never an option once a trait is being
edited directly on a sheet, and is transparently converted to a plain `Group` (summing
whatever children remain) the moment it is copied onto one, including during template
application at character creation.

## UI Changes ([ux/trait_editor.go](../../ux/trait_editor.go))

The container-type editor currently shows an "Alternative Slots" integer
field only for `AlternativeAbilities` and an "Ancestry" popup only for
`container.Ancestry`, toggled via `adjustFieldBlank`/`adjustPopupBlank`.
Add a parallel "Fixed Points" numeric field (nil-able, e.g. via a checkbox +
`IntegerField` or a text field that treats blank as "unset") shown only when
`ContainerType == container.FixedCost`:

- When the field is blank/disabled, `FixedPoints` stays `nil` and the
  container falls back to the `TemplatePicker` if it's configured as
  `picker.Points`, else sums children.
- The existing "Choices" template-picker UI in
  [ux/choices.go](../../ux/choices.go) already lets any
  `TemplatePickerProvider` configure a `picker.Points` picker; no changes
  needed there beyond ensuring `TraitContainerSyncData.TemplatePickerData()`
  continues to expose it for this new container type (it already does, since
  that method isn't gated on `ContainerType`).

Also update the trait row cell (`Trait.CellData`,
[trait.go](../../model/gurps/trait.go) `TraitDescriptionColumn` case) to show a
visible inline tag for `FixedCost`, matching the existing
`AlternativeAbilities`/`Ancestry`/`Attributes`/`MetaTrait` tags (for example,
`i18n.Text("Fixed")`). This tag is required so fixed-value containers are
visibly identifiable in the interface, regardless of whether their value comes
from `FixedPoints` or an exact `TemplatePicker`.

The "Fixed Points" field is a separate labeled field in `trait_editor.go`,
not part of the "Choices" wrapper in [ux/choices.go](../../ux/choices.go).
"Choices" stays scoped to the `TemplatePicker` concept alone; the container's
manual override is a distinct, independent setting, so it gets its own field
next to "Container Type", toggled via `adjustFieldBlank` the same way
"Alternative Slots" is today.

## Serialization / Compatibility

- New field `fixed_points` is `omitempty` and pointer-typed, so old files
  round-trip unchanged (field absent → `nil` → fallback behavior).
- New container-type key (`fixed_cost`) must be appended (not inserted) in the `allEnums`
  list to avoid shifting existing byte values that may be persisted... note:
  `container.Type` serializes via `Key()` (a string), not the raw byte, so
  ordering in the Go `const` block doesn't affect on-disk compatibility, only
  in-memory equality/ordinal comparisons (e.g. sorting `container.Types` in
  the UI popup). Still, appending at the end is the least surprising choice
  and matches how `MetaTrait` was added after the others.

## Testing

- Unit tests in `model/gurps/trait_test.go`, following
  `TestAlternativeAbilitiesCost`/`TestAlternativeAbilitiesMultipleSlots` as a
  template:
  - Fixed points set → `AdjustedPoints()` returns it regardless of children.
  - Fixed points unset, `TemplatePicker` set to `picker.Points` with
    `EqualsNumber` → returns the qualifier value.
  - Neither set → falls back to sum of children (existing `Group` behavior).
  - Fixed points set to `0` (pointer to zero, not nil) → returns `0`, not the
    picker/children value (proves the pointer's nil-ness, not its numeric
    value, is the switch).
- Hash stability test: confirm `Trait.Hash` differs when `FixedPoints`
  transitions between `nil` and any pointer value, and between different
pointee values, for a trait with `ContainerType == container.FixedCost`.
- Extend `organize_traits_test.go`'s container-type table test if organizing
  logic is expected to treat this new type like `MetaTrait`/
  `AlternativeAbilities` (i.e. not further reorganized as a plain `Group`).
- A test cloning a `FixedCost` trait with a non-nil `owner.OwningEntity()` (i.e. onto an
  `Entity`) confirms `ContainerType` becomes `Group` and `FixedPoints` becomes `nil` on the
  clone, while cloning the same trait onto a `Template` (`OwningEntity() == nil`) leaves it
  untouched.
- An end-to-end test (or an `ux` headless test alongside the existing
  `applyTemplateToSheetWithPickers` tests) applying a template containing a `FixedCost`
  container with an `EqualsNumber` `picker.Points` `TemplatePicker` to a sheet, confirming
  the resulting trait on the sheet is a plain `Group` whose summed children equal the
  original qualifier's value.
- Extend the picker application coverage with failure cases: an `EqualsNumber` picker
  whose qualifier does not equal the selected children's summed cost must not be
  accepted, an `AtMost` qualifier below that cost must fail, and an `AtLeast` qualifier
  above that cost must fail. Also cover compatible `AtMost`/`AtLeast` bounds so the test
  distinguishes genuine validation failures from the expected fallback behavior for
  non-`EqualsNumber` comparisons; these comparisons must use the selected children's
  actual summed cost, not the container's manual fixed-cost value.
- A test confirming the trait editor's "Container Type" popup omits `FixedCost` when the
  edited trait's data owner has a non-nil `OwningEntity()`.

## Out of Scope / Open Questions

- None currently open; prior questions were resolved above (children stay normal,
  browsable rows; the picker-derived branch requires `EqualsNumber`; the "Fixed Points"
  field is a separate labeled field in `trait_editor.go`; `FixedCost` is templates/library
  only and reverts to `Group` on characters).
