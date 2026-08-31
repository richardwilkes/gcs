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
  progress. The menu button beside the Edit Layout button holds the remaining layout commands: bringing hidden blocks
  back, resetting the sheet to the default layout, making the current sheet's layout the default for new sheets, and
  restoring the default layout to its factory arrangement.
- Because the layout is now edited on the sheet itself, the "Block Layout" text setting has been removed from Sheet
  Settings. Existing sheets keep the arrangement they had.
- The portrait block can now be any shape the layout gives it. The picture is drawn in the largest square that fits,
  centered within the block.
- When printing or exporting a sheet, a block that doesn't fit in the space left on a page is now moved whole to the
  next page rather than being cut in two. Lists are still split across pages as before, and blocks that sit side by
  side stay together.
- The new Edit ▸ Organize Traits command, available on character sheets, templates and trait library lists, files
  each top-level trait that isn't already a container into an Advantages, Perks, Disadvantages, Quirks, Features or
  Languages container, creating any that don't exist yet, placing them at the top of the list in that order, and
  sorting their contents by name. A trait tagged with one of these categories goes into that container; anything else
  is placed by its point cost. Any of these containers left empty afterward is removed. The whole rearrangement is a
  single undoable step, and the command is also available from the traits list's context menu. On a trait library
  list it is disabled while the content filter is in use.
- General Settings has two new options, "Check for GCS Updates" and "Check for Library Updates". Each can be set to
  check at launch (the default, and what GCS has always done), at launch and every hour, at launch and every day, or
  never. Turning a check on from never runs it right away, and Help ▸ Check for GCS updates and the Library
  Explorer's update buttons still work no matter how these are set.
- When a check finds a newer version of GCS, a pulsing "Software Update Available" button now appears in the Library
  Explorer's toolbar. Click it to read the release notes and install the update. The update window opens on its own
  only the first time a given release is seen -- after that, just the button appears -- and Help ▸ Check for GCS
  updates always opens it.
- Substitution placeholders can now offer a list of choices to pick from and say whether an empty or free-form entry
  is allowed. In the "Provide substitutions" dialog, each placeholder starts out as «not set» until you choose or
  enter a value. See the User Guide for more details.
- Searching the Library Explorer with deep search enabled is now much faster, and an active search's results refresh
  automatically as files change on disk.

## Bug Fixes

- The version numbers shown for libraries in the Library Explorer no longer depend on an update check having been
  made.
- Setting an optional specialization on a skill that came from a library no longer makes the skill show as modified
  from its source, and Sync with Source (or the sheet's sync-all button) no longer wipes the optional specialization
  out.
