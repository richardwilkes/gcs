# Changes since v5.44.0

## New & Improved

- Added the ability for a hidden attribute to be revealed with a chosen placement (Automatic, Primary, or Secondary)
  whenever the character has a named trait. The attribute settings now show "Placement [Hidden] unless trait [name] is
  present, then [placement]". (#845)
- Added support for internal anchor links in Markdown. Headings are now automatically assigned anchors, and links to
  them (e.g. `[Scripting](#scripting)`, or page references such as `md:User Guide/Scripting Guide#code`) scroll the
  target heading to the top of the view, revealing the section it introduces. (#651)
- Added support for a comma-separated list of tags in a feature's tag qualifier, so a single feature can grant its
  bonus to anything matching any one of the listed tags without stacking (e.g. `Sword, Axe, Polearm`). This applies to
  every tag-based criteria, including skill, spell, and their point bonuses, weapon bonuses, and equipment
  prerequisites. (#1008)
- Added a new "Sets the value of" feature for replacing a field with a chosen value, rather than adjusting it by a
  number or toggling a flag. Because such a field can hold only one value, these are resolved absolutely
  instead of stacking: when more than one applies to the same field, the one with the highest priority wins, ties are
  broken in favor of the more specific match, and the winning value along with the ones it overrode is shown in the
  tooltip.
- Added a new "Adjusts the maximum uses" feature for raising or lowering a piece of equipment's maximum uses. As with
  equipment modifier costs, the operation is taken from what you enter: a plain number adds (e.g. `1` or `-1`), a value
  ending in `%` adjusts by a percentage (e.g. `-10%`), and a value with an `x` multiplies (e.g. `x2`). It can optionally
  scale per level, and can apply to the equipment it is attached to ("to this equipment") or to other equipment matched
  by name and tags ("to equipment whose name"). The resolved maximum is always kept within the range 0 to 9,999,999.
- Added a "Reset Uses to Maximum" command alongside the existing "Increase Uses" and "Decrease Uses" commands, available
  from the menus, the equipment context menu, and as an assignable key binding.
- When a feature lowers a piece of equipment's maximum uses below its current remaining uses, the remaining uses shown
  (and adjusted by the uses commands) are now capped at the new maximum. The stored value is left untouched until you
  change it or save the file, at which point it is brought into range.

## Bug Fixes

- Fixed another case of the key handling not properly masking out sticky modifier keys (CapsLock & NumLock), this time
  affecting use of the Tab key to move focus between fields.
- Fixed and extended the merging of identical entries added to a character sheet, so a matching entry now adds to the
  existing one rather than creating a duplicate row: skills and spells combine their points, and leveled traits combine
  their levels (only when their modifiers, including which are enabled, are identical). This works whether the entry
  comes from applying a template, dragging, or copying, and in the cases that previously failed: entries with a tech
  level, entries whose modifiers or nameable substitutions are resolved as they are added, duplicate entries within a
  single template, and entries dragged or copied from another character sheet.
- Fixed spell prerequisite counting (for things like "6 spells from the Air college") so that a spell which itself
  requires the spell being checked is no longer counted toward that spell's own prerequisites, avoiding a circular
  prerequisite relationship. (#737)
- Fixed markdown page references whose path is URL-encoded (e.g. `md:User%20Guide/Scripting%20Guide`) so they resolve to
  the correct file, just like their non-encoded equivalents.
- Fixed the display of a skill whose optional specialization resolves to an empty string, so it no longer shows an empty
  set of parentheses `()` after the skill name.
- Fixed the Linux desktop integration so the application window is correctly associated with its launcher icon (added
  the missing `StartupWMClass` and the matching window `WM_CLASS`). (#1059)
- Fixed a technique's skill default so that switching its specialization comparison to "whose specialization is anything"
  now matches any specialization, rather than staying locked on the specialization text left over from a prior "is"
  selection. (#1061)
- Fixed the scripting skill lookups (`entity.findSkills`, `entity.skillLevel`, and a skill container's `find`) so a
  specialization argument now matches a skill by its optional specialization as well as its required specialization.
  (#1062)
