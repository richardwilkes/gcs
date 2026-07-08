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
- Added a new "Gives equipment a maximum uses modifier of" feature for raising or lowering a piece of equipment's
  maximum uses. As with equipment modifier costs, the operation is taken from what you enter: a plain number adds (e.g.
  `1` or `-1`), a value ending in `%` adjusts by a percentage (e.g. `-10%`), and a value with an `x` multiplies (e.g.
  `x2`). It can optionally scale per level, and can apply to the equipment it is attached to ("to this equipment") or to
  other equipment matched by name and tags ("to equipment whose name"). The resolved maximum is always kept within the
  range 0 to 9,999,999.
- Added a "Reset Uses to Maximum" command alongside the existing "Increase Uses" and "Decrease Uses" commands, available
  from the menus, the equipment context menu, and as an assignable key binding.
- When a feature lowers a piece of equipment's maximum uses below its current remaining uses, the remaining uses shown
  (and adjusted by the uses commands) are now capped at the new maximum. The stored value is left untouched until you
  change it or save the file, at which point it is brought into range.
- Added a maximum level to leveled traits, shown alongside the Level and Cost Per Level fields in the trait editor. It
  may be a plain number or a script expression (e.g. one that varies with SM or ST), and the editor displays the
  resolved value. A trait whose level exceeds its maximum is flagged on the character sheet with the same warning used
  for unmet prerequisites, and points are still computed from the actual level. When picking a trait's level while
  applying a template, the maximum is shown and the pickable level is capped to it. A companion "Gives a trait maximum
  level modifier of" feature can raise or lower the maximum without editing the trait's definition (preserving library
  source sync): as with the equipment uses adjustment, a plain number adds, a value ending in `%` adjusts by a
  percentage, and a value with an `x` multiplies; it can optionally scale per level, and can apply to the trait it is
  attached to ("to this trait") or to other traits matched by name and tags ("to traits whose name"). (#1060)

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
- Fixed a rare crash that could occur while scrolling when blank space at the edge of the content was collapsed. Scrolled
  views (tables, lists, and other scrollable panels) now settle correctly instead of getting caught in a loop.
- Fixed a startup failure on Linux where the application window could fail to be created (a `BadMatch` error) on some
  graphics drivers, most notably NVIDIA.
- Fixed saving and copying files failing on network drives (such as certain SMB/CIFS mounts) that don't allow changing
  file permissions. Preserving the original file permissions is now best-effort, so an otherwise-valid save or copy no
  longer aborts when the drive rejects the permission change.
