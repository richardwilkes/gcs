# Changes since the last release

## New & Improved

- Page references for `B:338+` (Basic Set: Campaigns) are redirected to use `BX` automatically. This allows split PDF users to open `B338+` page references. (#1098)
- Features can now be marked as "switchable". A switchable feature only takes effect while the switch of the trait,
  skill, spell, or equipment it belongs to is on. The switch is per item (a modifier's switchable features follow the
  switch of the item the modifier belongs to) and can be toggled from a new column that appears on the character sheet
  lists when any item in the list has switchable features, or from the item's editor. A trait or equipment container can
  have a switch of its own; hold down the Option/Alt key when clicking that switch to also apply the change to
  everything contained within it. Skill and spell containers never have a switch, since a container's own features
  don't apply.
- Spells can now have features, just like traits, skills and equipment. As a side effect, weapons attached to a spell
  now receive any "this weapon" bonuses defined by the spell's own features.

## Bug Fixes

- The `equipped` property of equipment in scripts now reports `false` for items in the Other Equipment list, since
  those items don't affect the character no matter what their equipped flag says. The Extended Value and Extended
  Weight previews in the equipment editor answer that property the same way the sheet does, including for a top-level
  row of the Other Equipment list.
- A "this armor" DR bonus (one that names no hit locations) now covers the locations that the armor's enabled modifiers
  grant DR to, along with those the armor itself grants DR to, whether the "this armor" bonus comes from the armor or
  from one of its modifiers. This can change the DR totals of existing characters whose armor picks up locations from
  its modifiers.
- The per-hit-location equipment lists produced by the legacy text export (the `EQUIPMENT` and `EQUIPMENT_FORMATTED`
  keys of the hit location loop) now include armor whose DR for that location comes from one of its modifiers, which
  the DR reported for the location already accounted for. Each piece of armor is also listed just once, rather than once
  per DR bonus of its own that reaches the location.
- Right-clicking one of the checkmark columns of the sheet's lists (Equipped, a modifier's enabled state, a weapon's
  hidden state) now selects the row and brings up its context menu, just as right-clicking anywhere else in the row
  does, rather than toggling the checkmark.
- Equipping or unequipping an item, whether by clicking its Equipped checkmark or with the Toggle Equipped command, now
  brings the Reactions, Conditional Modifiers and weapon lists onto or off of the sheet as the item's contributions
  come and go, rather than leaving them as they were until something else caused the sheet to be rebuilt.
- The per-hit-location equipment lists produced by the legacy text export now include armor whose DR reaches the
  location through the location that contains it (for body types that use sub-tables), matching the DR reported for
  the location.
- A template's equipment list now shows or hides its TL and LC columns as soon as the corresponding sheet settings are
  changed, rather than only after the template is reopened.
- A character sheet's modification timestamp is now updated by every edit made through its lists, not just some of
  them. Applying changes from an item's editor, deleting or duplicating rows, moving equipment between the carried and
  other equipment lists, converting rows to or from containers, answering a modifier or nameables prompt, swapping
  skill defaults, syncing with library sources, applying a template, and editing the point records all left the
  timestamp unchanged before. Several edits (inserting new items, dragging rows in, moving equipment between the two
  lists, and changing the sheet settings among them) also updated the sheet two or three times over for a single edit,
  which on a sheet with many rows made them noticeably slower than they needed to be.
