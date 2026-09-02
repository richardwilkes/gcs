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

## Bug Fixes

- Bonuses granted by a trait modifier are now attributed to both the trait and the modifier in tooltips, as bonuses from
  equipment modifiers already were.
- The tooltip for a per-level weapon bonus attached to a trait modifier now reports the same amount the weapon actually
  receives, rather than an amount based on the trait's level.
