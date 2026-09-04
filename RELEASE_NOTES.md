# Changes since the last release

## New & Improved

- Hovering over an attribute's name or value on the character sheet now shows where its bonuses come from, listing each
  source and its amount. Strength bonuses that apply only to striking, lifting or throwing are listed separately.
- Spell prerequisites can now require a particular power source, so spells from another power source no longer count
  toward them (per B77). A new spell prerequisite defaults to the same power source as the spell being edited, but you
  can choose "is anything" to accept any power source, or type in a specific one. Prerequisites in existing files are
  unchanged until you explicitly change them.
- Skill and weapon defaults can now match by tag. Alongside the name and specialization, a default can name a tag, so
  Stage Combat can default to whichever combat skill the character actually has at -3 instead of listing each one.
  Existing files are unchanged until you fill in a tag.
- A skill with points in it keeps using the default it has been using rather than silently switching to another one.
  Changing its list of defaults, whether by editing them or by syncing from the library, now picks the best default
  afresh.
- Dropping a modifier onto one of several selected traits or equipment items now adds it to all of them. Every item
  that will receive it is highlighted while you drag, and a single undo removes it from all of them again. If the
  modifier has substitutions to fill in, the prompt says which item each copy belongs to. Dropping onto an item that
  isn't part of the selection still affects only that item. (#1096)
- Toggle State and its shortcut now work inside the detail editors. Select any number of trait or equipment modifiers
  to turn them on or off, or any number of melee or ranged weapons to hide or show them, and the whole selection
  changes at once and can be undone in one step. The command is also in the right-click menu of those lists. (#1074)
- Items in any list that can be rearranged by dragging can now be rearranged from the keyboard. Move Up and Move Down
  shift the selected items one place among their neighbors, Move Out of Container lifts them out of their container
  to sit just above it, and Move Into Container places them at the top of the container directly below them, opening
  it if it was closed. Items selected together keep their order, and a selection spread across several containers
  moves within each of them. The commands are in the Edit menu and the right-click menu, bound to Shift-Command-Arrow
  on macOS and Shift-Ctrl-Arrow elsewhere, and can be rebound in the Menu Keys settings. (#1076)
- Alternative Abilities containers can now have more than one slot. A new Alternative Slots field in the trait editor
  sets how many of the container's abilities can be in use at the same time. That many of the most expensive
  abilities are charged at full cost and the rest at 20%. A container with more than one slot shows the count in the
  traits list, such as "Alternate x2". Existing containers have one slot, so their costs are unchanged. (#1125)

## Bug Fixes

- Bonuses granted by a trait modifier are now attributed to both the trait and the modifier in tooltips, as bonuses
  from equipment modifiers already were.
- The tooltip for a per-level weapon bonus attached to a trait modifier now reports the same amount the weapon
  actually receives, rather than an amount based on the trait's level.
- Copying items from a library or template into a character sheet, loot sheet or template now prompts for modifier
  choices and substitutions again. Applying a template was not affected. (#1126)
- Dialogs now open focused and ready for the keyboard. On Windows, and sometimes on macOS, a dialog opened from a menu
  or a toolbar button could come up without keyboard focus, so Enter and Escape did nothing until it had been clicked
  on first. (#1124)
- Fixed a crash on macOS that could happen as soon as the first window opened, seen on an Intel Mac running macOS
  26.2.
