# Changes since the last release

## New & Improved

- The arrangement of the blocks on a character sheet is now edited directly on the sheet. Click the new Edit Layout
  button in the sheet's toolbar (or choose View ▸ Edit Sheet Layout) to switch into layout editing, where you can:
  - Drag any block -- the portrait, identity, attributes, body type, the lists, and so on -- to place it beside,
    above, or below another block, or drop it onto the seam between two blocks to sit alongside both of them.
  - Drag the divider between side-by-side blocks to change how the width is shared between them.
  - Drag a block's bottom edge to give it a minimum height.
  - Click the × in a block's top right corner to hide it. The portrait has a second button beside the × that makes
    its picture area square.
  - Right-click a block for a menu with these same commands.

  Layout changes can be undone and redone like any other edit to the sheet, and pressing Escape abandons a drag in
  progress.
- The menu button beside the Edit Layout button holds the remaining layout commands: bringing hidden blocks back,
  resetting the sheet to the default layout, making the current sheet's layout the default for new sheets, and
  restoring the default layout to its factory arrangement.
- The "Block Layout" text setting has been removed from Sheet Settings, since the layout is now edited on the sheet
  itself. Existing sheets keep the list arrangement they had.
- The portrait block can now be any shape the layout gives it. The picture is drawn in the largest square that fits,
  centered within the block.
- When printing or exporting a sheet, a block that doesn't fit in the space left on a page is now moved whole to the
  next page rather than being cut in two. Lists continue to be split across pages as before, and blocks that sit side
  by side are kept together.

## Bug Fixes

- A skill's optional specialization is no longer treated as part of its library data. Setting one on a skill that came
  from a library used to make the skill show as not matching its source, and Sync with Source (or the sheet's sync-all
  button) then wiped the optional specialization out. Such a skill now shows as matching its source and keeps its
  optional specialization through a sync.
