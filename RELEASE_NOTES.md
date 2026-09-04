# Changes since the last release

## New & Improved

- You can now choose how many decimal places the character sheet shows for the height and weight in the Description
  block, for equipment weights, and for equipment values. Each has its own setting in the sheet settings, from "As
  Needed" (every decimal place the value has, as before) down to whole numbers, along with an option to pad with
  trailing zeros so that columns line up. Only the display changes: values are still stored and calculated at full
  precision, hovering over a rounded equipment weight, value or total shows the exact figure, and clicking into the
  height or weight field shows the exact value for editing. Templates and loot sheets follow the equipment settings in
  the default sheet settings. PDF and image exports and printing show the sheet as it is displayed, while the detail
  editors, library lists and text template exports continue to show full precision. (#1070)
- The damage type that a DR feature applies against can now use substitution markers, such as "@Damage Type@", so a
  trait or modifier from a library can ask which damage type it should protect against when you add it. (#1017)
- A new "Multiply cost by level" checkbox in the trait modifier editor, on by default, controls whether a leveled
  modifier's cost adjustment is multiplied by its level. Turn it off to stop multiplying it, while any per-level
  features the modifier carries still scale with the level. That lets a "Limited, -40%" limitation on Damage
  Resistance 3 remove the DR against everything and grant DR against a chosen damage type instead, with the -40%
  charged once rather than once per level, so the trait costs 9 points rather than 3. (#1017)

## Bug Fixes

- (None)
