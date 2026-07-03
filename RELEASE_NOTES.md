# Changes since v5.43.0

## New & Improved

- Text fields now show a context menu on right-click containing the standard Cut, Copy, Paste and Select All actions.
  Only the actions that can currently be performed are included, and if none of them apply, no menu is shown.

## Bug Fixes

- Linux only: Fix use of NumLock causing mouse clicks to be misinterpreted.
- Key handling now properly masks out sticky modifier keys (i.e. CapsLock & NumLock) before handling them. This affected
  menu accelerator matching (menu command key sequences were ignored), key navigation within open menus, the fallback
  cut/copy/paste/select-all handling in text fields, and Return/Enter in the file dialog's file name field.
- Fixed library "update to latest" option failing on reserved device names. (#1057)
- Fixed substitutions in a skill's optional specialty not being detected. (#1054)
