# Changes since v5.44.0

## New & Improved

- Added the ability for a hidden attribute to be revealed with a chosen placement (Automatic, Primary, or Secondary)
  whenever the character has a named trait. The attribute settings now show "Placement [Hidden] unless trait [name] is
  present, then [placement]". (#845)
- Added support for internal anchor links in Markdown. Headings are now automatically assigned anchors, and links to
  them (e.g. `[Scripting](#scripting)`, or page references such as `md:User Guide/Scripting Guide#code`) scroll the
  target heading to the top of the view, revealing the section it introduces. (#651)

## Bug Fixes

- Fixed another case of the key handling not properly masking out sticky modifier keys, this time affecting use of the
  Tab key to move focus between fields.
- Fixed and extended the merging of identical entries added to a character sheet, so a matching entry now adds to the
  existing one rather than creating a duplicate row: skills and spells combine their points, and leveled traits combine
  their levels (only when their modifiers, including which are enabled, are identical). This works whether the entry
  comes from applying a template, dragging, or copying, and in the cases that previously failed: entries with a tech
  level, entries whose modifiers or nameable substitutions are resolved as they are added, duplicate entries within a
  single template, and entries dragged or copied from another character sheet.
- Fixed spell prerequisite counting so that a spell which itself requires the spell being checked is no longer counted
  toward that spell's own prerequisites, avoiding a circular prerequisite relationship. (#737)
- Fixed markdown page references whose path is URL-encoded (e.g. `md:User%20Guide/Scripting%20Guide`) so they resolve to
  the correct file, just like their non-encoded equivalents.
- Fixed the display of a skill whose optional specialization resolves to an empty string, so it no longer shows an empty
  set of parentheses `()` after the skill name.
- Fixed the Linux desktop integration so the application window is correctly associated with its launcher icon (added
  the missing `StartupWMClass` and the matching window `WM_CLASS`). (#1059)
