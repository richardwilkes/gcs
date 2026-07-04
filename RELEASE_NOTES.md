# Changes since v5.43.0

## New & Improved

- Scripts can now read a trait's total point value via `self.points` (and the `points` field on any trait obtained
  through a script). The value accounts for base points, cost-per-level, trait modifiers, and the reduced cost of
  children within an Alternative Abilities container. (#1053)
- Text fields now show a context menu on right-click containing the standard Cut, Copy, Paste and Select All actions.
  Only the actions that can currently be performed are included, and if none of them apply, no menu is shown.
- Linux only: GCS now installs a desktop icon when it starts, so it shows up properly in the file viewer. This is done
  when launched, so the icon will not show up until you've launched GCS once. Also note that it only applies to
  Gnome-based desktops.

## Bug Fixes

- Linux only: Fix use of NumLock causing mouse clicks to be misinterpreted.
- Key handling now properly masks out sticky modifier keys (i.e. CapsLock & NumLock) before handling them. This affected
  menu accelerator matching (menu command key sequences were ignored), key navigation within open menus, the fallback
  cut/copy/paste/select-all handling in text fields, and Return/Enter in the file dialog's file name field.
- Fixed library "update to latest" option failing on reserved device names. (#1057)
- Fixed substitutions in a skill's optional specialty not being detected. (#1054)
- Fixed crashes on some configurations (primarily Linux) when the primary display cannot be determined.
- Fixed newly opened character sheets, lists, PDFs and other views sometimes not receiving the keyboard focus, or
  losing it immediately, so you can now interact with them via the keyboard right after opening.
- Windows only: Fixed the packaged executable showing a generic icon and missing its version details in the Properties
  dialog, caused by the icon and version resources not being embedded during the release build.
- Fixed a leveled trait's level-adjusting bonus not being applied when the trait's name uses a substitution (such as a
  Talent with a fill-in-the-blank name).
- Fixed a spell's college not having its substitutions applied when accessed from a script.
- Fixed the safeguard that keeps a skill's level from being resolved through itself not working when the skill's name
  uses a substitution, which could cause an error when a script reads that skill's level.
- Fixed a crash in the Hiking distance calculator when a character's Move was 0. In that case, the travel time is now
  shown as a dash, since no distance can be covered. Also fixed the travel time reading "1 days" instead of "1 day"
  for a one-day hike.
- Fixed changes to an equipment modifier's per-level and per-pound cost and weight options not being noticed when
  comparing against a library source, so those options now sync correctly.
- Fixed a crash that could occur when a script requested a random weight for a character with a very low Strength.
- Fixed a trait loaded from a file with a negative level (which can really only occur if someone edited the file by
  hand) being left at that invalid value; such traits are now corrected to a level of 0.
- Fixed a crash when adding or configuring a library with invalid settings; GCS now reports the problem instead of
  quitting.
- Command-line only: Combining the `--text` export option with `--convert` or `--sync` is now reported as an error
  rather than being silently ignored, and requesting a `--text` export with no exportable files is now reported as an
  error instead of doing nothing.
