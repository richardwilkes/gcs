# Changes since v5.44.0

## New & Improved

- Added the ability for a hidden attribute to be revealed with a chosen placement (Automatic, Primary, or Secondary)
  whenever the character has a named trait. The attribute settings now show "Placement [Hidden] unless trait [name] is
  present, then [placement]". (#845)
- Added support for internal anchor links in Markdown. Headings are now automatically assigned anchors, and links to
  them (e.g. `[Scripting](#scripting)`, or page references such as `md:User Guide/Scripting Guide#code`) scroll the
  target heading to the top of the view, revealing the section it introduces. (#651)

## Bug Fixes

- Fixed markdown page references whose path is URL-encoded (e.g. `md:User%20Guide/Scripting%20Guide`) so they resolve to
  the correct file, just like their non-encoded equivalents.
- Fixed the Linux desktop integration so the application window is correctly associated with its launcher icon (added
  the missing `StartupWMClass` and the matching window `WM_CLASS`). (#1059)
- Fixed spell prerequisite counting so that a spell which itself requires the spell being checked is no longer counted
  toward that spell's own prerequisites, avoiding a circular prerequisite relationship. (#737)
