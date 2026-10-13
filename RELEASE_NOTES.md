# Changes since the last release

## New & Improved

- Settings > Saved Filters… lists the saved filters of every library list type in one place, where they can be
  created, edited, duplicated, merged and deleted, with each change undoable. Filters can also be exported to and
  imported from a `.filters` file, all of them or only those selected.

## Bug Fixes

- Syncing with library sources no longer overwrites data whose kind differs from its source's, such as a trait whose
  source is a trait container or equipment whose source is an equipment group, which used to lose data. The library
  source column marks such data with "≠" instead. Equipment and a physical container still sync with each other, as do
  a note and a note container, and the editor's Sync with Source now allows that for equipment as well.
- The saved filter editor no longer accepts a filter with no conditions, including one holding only empty groups. OK
  stays off until at least one condition has been added.
