# Changes since the last release

## New & Improved

- The calculators have been reworked and gathered into a single Calculators view, opened from View → Calculators or
  from the calculator button on a character sheet, with a tab for each. New calculators cover Explosions & Area
  Attacks (B413-B415), Scatter (B414), Demolition (B415), and Collisions & Falls (B430-B432). The existing Jumping,
  Throwing and Hiking calculators can now draw their numbers from any open character sheet or have them typed in,
  follow the extra effort rules more closely (B356-B357), and explain the rolls and their consequences. Hiking also
  works out the day's FP cost hour by hour (B426), including the slowdown and collapse of a tired hiker, a rest
  halfway through the day (B427), and the effects of Fit, Very Fit (B55) and Recover Energy (B248).
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
- Added an ancestry editor, opened from File → New Ancestry, for building the ancestry files that drive the random
  name, gender, age, height, weight, hair, eye, skin and handedness choices on a character sheet. Give the ancestry a
  name, fill in the common options, and add gender-specific overrides; everything is undoable. Each ancestry opens in
  an editor of its own, so several can be open at once, and an ancestry that is already open is brought forward rather
  than opened twice. The toolbar menu opens any ancestry from your libraries, or one from a file elsewhere. Save
  writes back to the file you opened, while for a new ancestry it opens a save dialog that starts in your User
  Library's Settings/Ancestries folder, and Save As writes a copy elsewhere. Ancestry and name generator files can also
  be opened with File → Open, from the recent files list, or by dropping them onto the window. As before, an ancestry
  in your User Library takes precedence over one of the same name from the Master Library or built into GCS. (#607)
- Added a name generator editor, opened from File → New Name Generator, for the .names files that ancestries draw
  their random names from. Choose how names are generated (a training name picked at random, new names built letter
  by letter or from runs of vowels and consonants, or several generators combined with a separator), whether training
  names are lowercased and results capitalized, and the training names themselves with optional weights. A button
  imports names from a text file, one per line, and rows can be selected with click, shift-click and command-click and
  removed together. Sample names from the current definition are shown as you edit. Saving, opening from your
  libraries, having several open at once, and undo work as in the ancestry editor, and each name generator row in the
  ancestry editor has a button that opens that generator for editing. (#607)
- Scripts can now read three more fields of the `entity` object: `otherEquipment`, the equipment not currently
  carried, and `wealthCarried` and `wealthNotCarried`, the total value of the carried and the other equipment.

## Bug Fixes

- (None)
