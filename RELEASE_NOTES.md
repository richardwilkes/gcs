# Changes since the last release

## New & Improved

- The ID, Source ID, Source Library and Source Path fields have been removed from the item editors. They are now in a
  menu opened by a new button on the editor's toolbar. Choosing one copies its value. The menu also shows whether the
  item still matches its library source and offers Sync with Source and Clear Source. Like any other change made in
  the editor, these take effect when the editor's changes are applied.
- A new title note feature adds text in parentheses after the name of the trait, skill, spell or piece of equipment
  it is on, the way a skill's specialization is shown, such as the "Bob" in "Allies (Bob)". It can hold substitution
  markers, like a name. On a modifier, it is shown after the name of the item the modifier belongs to while the
  modifier is enabled.
- Trait and equipment modifiers have a new Short Name field and a "Show in Title Notes" check box. A modifier set to
  show in the title notes is listed in parentheses after its owner's name, by its short name if it has one, rather
  than in its owner's notes, where it stays only if it has notes of its own. The short name is also used in the
  owner's notes, and when a script or an export template looks a modifier up by name. A new "Show in Owner's Notes"
  check box, on by default, can keep a modifier's notes out of its owner's notes. "Also show notes in weapon usage" is
  now "Show in Weapon Usage".
- Trait, skill and equipped equipment prerequisites can now check an item's title notes, and spell prerequisites can
  count spells by their title notes. Trait and equipped equipment prerequisites can also check for a modifier by its
  name or short name. Both checks work as tag checks do: a list of values separated by commas matches if any one of
  them does.
- The parts in parentheses after a name are now separated by semicolons, as the GURPS books do, so a skill reads
  "Guns (Pistol; Revolver)" rather than "Guns (Pistol, Revolver)".

## Bug Fixes

- Sync with Source now works on items in library lists.
- A technique whose default is based on an attribute rather than a skill, such as Neck Snap, no longer shows as
  differing from its library source after being added to a sheet, edited or synced.
