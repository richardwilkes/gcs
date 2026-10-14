# Changes since the last release

## New & Improved

- (none yet)

## Bug Fixes

- Undo and redo no longer lose edits made around a change that rearranges a list. Editing an item, toggling it, or
  converting it, with a drag, move, deletion, duplication, or applied template before or after it, then undoing and
  redoing past that change, used to leave the edit without effect and could silently drop the changes of an editor
  left open across the undo. This applied to every list on sheets, templates, loot sheets and library files, and to
  the modifier and weapon lists inside an editor.
- Syncing with library sources no longer overwrites data whose kind differs from its source's, such as a trait whose
  source is a trait container or equipment whose source is an equipment group, which used to lose data. The library
  source column marks such data with "≠" instead. Equipment and a physical container still sync with each other, as do
  a note and a note container, and the editor's Sync with Source now allows that for equipment as well.
- The saved filter editor no longer accepts a filter with no conditions, including one holding only empty groups. OK
  stays off until at least one condition has been added.
