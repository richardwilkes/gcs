# Changes since the last release

## New & Improved

- Skills, spells and equipment on a sheet can now be disabled, as traits can, from the Enabled check box in their
  editor, and skills and spells also with Toggle State. A disabled item, along with everything in a disabled
  container, is treated as if it weren't on the sheet: it costs no points, has no level, adds no features, weapons,
  weight or value, and satisfies no prerequisites or defaults. Copying one into a library or template enables it again.

## Bug Fixes

- Syncing with library sources no longer overwrites data whose kind differs from its source's, such as a trait whose
  source is a trait container or equipment whose source is an equipment group, which used to lose data. The library
  source column marks such data with "≠" instead. Equipment and a physical container still sync with each other, as do
  a note and a note container, and the editor's Sync with Source now allows that for equipment as well.
- The saved filter editor no longer accepts a filter with no conditions, including one holding only empty groups. OK
  stays off until at least one condition has been added.
