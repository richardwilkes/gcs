# Changes since the last release

## New & Improved

- Hovering over an attribute's name or value on the character sheet now shows where its bonuses come from, listing each
  source and its amount. Strength bonuses limited to striking, lifting or throwing are listed in sections of their own.
- Spell prerequisites can now require a particular power source, so spells from another power source no longer count
  toward them (per B77). A new spell prerequisite asks for the same power source as the spell being edited, but you can
  choose "is anything" to accept any of them, or type in a specific power source to look for. Prerequisites in existing
  files are unchanged until you explicitly change them.
- Skill defaults can now be selected by tag. Alongside the name and specialization, a default to a skill can ask for a
  tag, so Stage Combat can default to whichever combat skill the character actually has at -3 instead of having to name
  each one. Weapon defaults can do the same. Existing files are unchanged until you fill in a tag.
- A skill with points in it keeps using the default it has been using rather than silently switching to another one,
  but that no longer applies across a change to its list of defaults: editing them, or syncing them from the library,
  now picks the best default afresh.
- Dropping a modifier onto one of several selected traits or equipment items now adds it to all of them. Every item
  that will receive it is highlighted while you drag, and a single undo removes it from all of them again. If the
  modifier has substitutions to fill in, the prompt says which item each copy belongs to. Dropping onto an item that
  isn't part of the selection still affects only that item. (#1096)
- Toggle State and its shortcut now work inside the detail editors. Select any number of trait or equipment modifiers
  to turn them on or off, or any number of melee or ranged weapons to hide or show them, and the whole selection is
  changed at once and taken back with a single undo. The command is also available from the right-click menu of those
  lists. (#1074)

## Bug Fixes

- Bonuses granted by a trait modifier are now attributed to both the trait and the modifier in tooltips, as bonuses from
  equipment modifiers already were.
- The tooltip for a per-level weapon bonus attached to a trait modifier now reports the same amount the weapon actually
  receives, rather than an amount based on the trait's level.
- Copying items from a template to a character is now correctly triggering modifier and nameable configuration. Applying
  templates was not affected by this bug. (#1126)
- Dialogs now open focused and ready for the keyboard. On Windows, a dialog opened from a menu or a toolbar button,
  such as when applying a template to a character sheet, could come up without the keyboard focus, so Enter and Escape
  did nothing until it had been clicked on first. macOS could show the same problem in some cases. (#1124)
