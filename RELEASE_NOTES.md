# Changes since v5.44.0

## New & Improved


## Bug Fixes

- Fixed the Linux desktop integration so the application window is correctly associated with its launcher icon (added
  the missing `StartupWMClass` and the matching window `WM_CLASS`). (#1059)
- Fixed spell prerequisite counting so that a spell which itself requires the spell being checked is no longer counted
  toward that spell's own prerequisites, avoiding a circular prerequisite relationship. (#737)
