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
  prerequisites when it has any. That paragraph no longer shows above them once the section is expanded.
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
