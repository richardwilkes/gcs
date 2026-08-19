# Changes since the last release

## New & Improved

- Page references for `B:338+` (Basic Set: Campaigns) are redirected to use `BX` automatically. This allows split PDF users to open `B338+` page references. (#1098)
- Features can now be marked as "switchable". A switchable feature only takes effect while the switch of the trait,
  skill, spell, or equipment it belongs to is on. The switch is per item (a modifier's switchable features follow the
  switch of the item the modifier belongs to) and can be toggled from a new column that appears on the character sheet
  lists when any item in the list has switchable features, or from the item's editor. Hold down the Option/Alt key when
  clicking the column to also apply the change to everything contained within the item.
- Spells can now have features, just like traits, skills and equipment. As a side effect, weapons attached to a spell now
  receive any "this weapon" bonuses defined by the spell's own features.

## Bug Fixes

- The `equipped` property of equipment in scripts now reports `false` for items in the Other Equipment list, since
  those items don't affect the character no matter what their equipped flag says.
