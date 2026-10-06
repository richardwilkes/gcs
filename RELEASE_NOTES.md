# Changes since the last release

## New & Improved

- The ID, Source ID, Source Library and Source Path fields have been removed from the item editors. They are now in a
  menu opened by a new button on the editor's toolbar. Choosing one copies its value. The menu also shows whether the
  item still matches its library source and offers Sync with Source and Clear Source. Like any other change made in
  the editor, these take effect when the editor's changes are applied.
- The features editor has been rebuilt to work like the prerequisites editor. Each feature is shown as a short
  sentence, such as "+1 per level to skill Streetwise", that opens to edit it. Optional criteria and switchable are
  added as pills, and the type menu groups the types under headings. Features can be duplicated, moved or dragged to
  reorder them, new ones are added at the end of the list, and every change can be undone.
  The section opens collapsed to a paragraph of its features' sentences when it has any; clicking its title or the
  paragraph expands it.
- The prerequisites section of the editors works the same way, opening collapsed to a paragraph describing its
  prerequisites when it has any. That paragraph no longer shows above them once the section is expanded. A group's
  add button has moved into its more menu, except for the top group's, which takes the place of a more menu, and an
  empty group's, which now sits beside its placeholder. Choosing All of Group or Any of Group for an empty top group
  that already shows its pill now adds a group, rather than changing the top group's type.
- Wherever prerequisite and feature sentences compare text, as in their "whose ..." and specialization clauses, a value
  with a comma, a double quote, or space at either end is now quoted, so that where it starts and stops shows, and one
  holding a double quote is put in single quotes.
- The editor for saved list filters has been rebuilt to work like the prerequisites editor. Each condition is shown as a
  short sentence, such as "Must not be a container", that opens to edit it. Each group's pill says how it combines
  what it holds: all of it must match, any of it, none of it, or not all of it. Conditions and groups can be
  duplicated, moved, wrapped in a group or dragged into place, a group can be ungrouped when that keeps what the
  filter asks for, new ones are added at the end, and every change can be undone. Escape closes the open condition,
  or with none open, cancels the editor. A value with a comma, a double quote, or space at either end is quoted, as is
  each of several values given as a list.
- The saved filters popup now lists New Filter, Edit Filter and Delete Filter first, so they stay within reach of a
  long list of filters.
- Weapon bonus tooltips on the sheet use shorter names, such as "to accuracy", and the rate of fire bonuses say which
  mode they apply to.
- A new weapon bonus that picks weapons by their skill no longer starts out requiring a relative skill level of at least
  0, which left out weapons used at a lower level. A relative skill level can still be added to it.

## Bug Fixes

- Sync with Source now works on items in library lists.
- A technique whose default is based on an attribute rather than a skill, such as Neck Snap, no longer shows as
  differing from its library source after being added to a sheet, edited or synced.
- Unticking the last location of a DR bonus quietly turned it into one that applies to the armor it's attached to.
  The last location can no longer be unticked.
- A saved filter's condition on text or a list, such as the notes or the tags, no longer passes an item that has none
  of it. A condition that the tags be anything now needs a tag that isn't blank, one that no tag be "Shield" no longer
  passes an item with no tags, and negating a condition that the notes be anything finds the items without notes. Saved
  filters with such conditions may show different results.
- An empty group in a saved filter set to None of or Not all of hid every item. An empty group now asks for nothing,
  whichever way it is set.
