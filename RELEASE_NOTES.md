# Changes since the last release

## New & Improved

- The tag filter popup is back in the toolbar of library lists, between the search field and the saved filters. It
  offers the tags of the items the saved filter in force lets through, and shows only the items carrying every tag
  chosen (Shift-click, or Command-click on macOS and Ctrl-click elsewhere, to choose more than one). The chosen tags
  are cleared when another saved filter, or None, is chosen, and are not saved. Its Save Filter command opens the New
  Filter editor with the saved filter and the chosen tags combined.

## Bug Fixes

- Syncing with library sources no longer overwrites data whose kind differs from its source's, such as a trait whose
  source is a trait container or equipment whose source is an equipment group, which used to lose data. The library
  source column marks such data with "≠" instead. Equipment and a physical container still sync with each other, as do
  a note and a note container, and the editor's Sync with Source now allows that for equipment as well.
- The saved filter editor no longer accepts a filter with no conditions, including one holding only empty groups. OK
  stays off until at least one condition has been added.
- The saved filter editor no longer offers values for "is" and "is not" that a condition could never match, such as a
  tag holding a nameable marker with choices.
