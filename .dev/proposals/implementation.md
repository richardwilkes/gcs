# Fixed Cost Implementation Plan

Implement the proposal as small, independently reversible commits. Each step should leave the repository buildable and should include its focused validation before the next step begins.
Try building and running the tests after each step to ensure correctness.

## 1. Add the enum

- Add `fixed_cost` to `cmd/enumgen/main.go`.
- Run `go generate ./cmd/enumgen/main.go`.
- Commit the generator input and generated output together.
- Validate with `go test ./model/gurps/...`.

This commit adds only the new enum surface and can be reverted independently.

## 2. Add persisted model state

- Add `FixedPoints *fxp.Int` to `TraitContainerSyncData`.
- Update `ClearUnusedFieldsForType`.
- Update `TraitContainerSyncData.hash`.
- Add serialization and hash tests, including `nil` versus pointer-to-zero.

Do not change point calculation in this step.

## 3. Implement fixed-cost calculation

- Add the `FixedCost` branch to `Trait.AdjustedPoints`.
- Add focused model tests for:
  - Manual fixed value overriding children and picker values.
  - An exact points picker supplying the value.
  - Fallback to the summed child cost.
  - An explicit fixed value of zero.
  - `AtMost`, `AtLeast`, and `NotEquals` falling back to the child sum.

Validate with the relevant `model/gurps` tests.

## 4. Enforce template-only ownership

- Add a helper that demotes `FixedCost` to `Group` and clears `FixedPoints` when the owner has an entity.
- Call it from `Trait.Clone` and `Trait.SetDataOwner`.
- Add tests for cloning onto:
  - An `Entity`, which must demote the trait.
  - A `Template` or library owner, which must preserve `FixedCost`.

Keep ownership enforcement separate from UI work.

## 5. Hide the option on character sheets

- Filter `container.FixedCost` from the container-type popup when the trait owner has an entity.
- Add editor tests confirming that sheet traits cannot select it while template and library traits can.

This is a UI-only change and should remain independently revertible.

## 6. Add the Fixed Points editor

- Add a nil-aware numeric field beside the existing container settings.
- Keep it separate from the `Choices` template-picker UI.
- Add tests for setting, clearing, and explicitly setting zero.
- Confirm changing away from `FixedCost` clears the field.

These are additional editor requirements to the largely implemented feature:

- Keep the existing behavior where an empty textbox means `nil`.
- Add an explicit clear action to the textbox so users have a second way to
  clear the value and restore `nil` without manually deleting the text.
- Test both the empty-textbox path and the explicit clear action, including
  that neither path is confused with an explicitly entered zero.

## 7. Add the visible row tag

- Update `Trait.CellData` / `TraitDescriptionColumn` to show a visible `Fixed` tag, matching the existing `Meta` presentation.
- Add or extend the trait cell-data test.

The tag should identify fixed-cost containers regardless of whether their value comes from `FixedPoints` or an exact points picker.

## 8. Test picker application behavior

First add focused validation coverage around the existing `tp.Qualifier.Matches(total)` behavior. If the modal callback is difficult to test directly, extract a small pure helper for the comparison rather than testing UI state indirectly.

Cover these cases:

- Exact qualifier equal to the selected child total: accepted.
- Exact qualifier different from the selected child total: rejected.
- `AtMost` below the selected child total: rejected.
- `AtMost` equal to or above the selected child total: accepted.
- `AtLeast` above the selected child total: rejected.
- `AtLeast` equal to or below the selected child total: accepted.

For the fixed-cost edge cases, make the selected children sum to the manual fixed value. This proves validation uses the children’s actual total rather than trusting the parent’s fixed value.

Then add the end-to-end template-to-sheet test confirming that the copied trait becomes `Group` and that its remaining children sum to the validated picker total.

## 9. Full verification

Run:

```sh
gofmt -w <changed-go-files>
go test ./model/gurps/...
go test ./ux/...
go test ./...
./build.sh -a
```

Use the narrower tests after each implementation step; run the full build only after all slices are complete.

## 10. Document the completed feature

- Update the relevant project documentation to describe `FixedCost`, its
  template-only ownership, its manual/picker/children precedence, and the
  distinction between an unset value (`nil`) and an explicit zero.
- Document both ways to clear the Fixed Points field: leaving the textbox
  empty and using the explicit clear action.
- Keep this documentation update additive and separate from the already
  implemented model and UI behavior.

## Boundaries

Keep `AdjustedPoints` tests separate from picker-validation tests. The former determines the modeled cost of a container; the latter determines whether selected children satisfy a picker constraint. This separation keeps failures local and allows either behavior to be reverted without leaving the other half of the feature active.
