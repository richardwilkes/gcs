# Changes since the last release

## New & Improved

- The saved filter editor is easier to get started with. The name field shows an example of a name, and a duplicate
  name gets a red edge instead of filling the field red. Why OK is off, a duplicate name or a filter with nothing in it,
  is spelled out beside the buttons. The editor's rows are now called filters rather than conditions, and clicking the
  "No filters" placeholder adds one directly.

## Bug Fixes

- Syncing with library sources no longer overwrites data whose kind differs from its source's, such as a trait whose
  source is a trait container or equipment whose source is an equipment group, which used to lose data. The library
  source column marks such data with "≠" instead. Equipment and a physical container still sync with each other, as do
  a note and a note container, and the editor's Sync with Source now allows that for equipment as well.
- The saved filter editor no longer accepts a filter with no conditions, including one holding only empty groups. OK
  stays off until at least one condition has been added.
