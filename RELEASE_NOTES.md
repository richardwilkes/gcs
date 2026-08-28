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
- Edit ▸ Organize Traits, available on character sheets, templates and trait library lists, files each top-level trait
  that isn't a container into an Advantages, Perks, Disadvantages, Quirks, Features or Languages container, creating
  any that don't exist yet, placing them at the top of the list in that order, and sorting their contents by name. A
  trait carrying a Language tag goes into Languages; otherwise a tag naming exactly one of the other categories
  decides, and anything else is placed by its point cost (a disabled trait by the cost it would have if enabled). Only
  containers of the group type are reused, and any of these containers left empty is removed. The whole rearrangement
  is a single undoable step, and the command is also available from the traits list's context menu. On a trait library
  list it is disabled while the content filter is in use, since a filtered list can't be rearranged.
- General Settings has two new choices, "Check for GCS Updates" and "Check for Library Updates". Each can be set to
  check at launch (the default, and what GCS has always done), at launch and every hour, at launch and every day, or
  never. Help ▸ Check for GCS updates and the Library Explorer's update buttons still work no matter what these are
  set to: when no check has been made yet, the Library Explorer's buttons look for the library's releases when
  clicked. Turning either check on from never runs it right away.
- When a check finds a newer version of GCS, a pulsing "Software Update Available" button appears in the Library
  Explorer's toolbar. Click it to read the release notes and install the update. The button settles down after a
  couple of minutes. It pulses again when a newer release turns up, and whenever it reappears after having been
  hidden, as it is by Help ▸ Check for GCS updates and by "Check for GCS Updates" being switched back on from never.
  The update window now opens on its own only the first time a given release is seen -- later launches just show the
  button -- and Help ▸ Check for GCS updates, which is usable whenever a check isn't already running, always opens it.
  With "Check for GCS Updates" set to never, the button isn't shown.
- Added more capabilities to the substitution placeholders - See the User Guide or PR (#1113) for more details.

## Bug Fixes

- The version numbers shown for libraries in the Library Explorer no longer depend on an update check having been
  made.
- A skill's optional specialization is no longer treated as part of its library data. Setting one on a skill that came
  from a library used to make the skill show as not matching its source, and Sync with Source (or the sheet's sync-all
  button) then wiped the optional specialization out. Such a skill now shows as matching its source and keeps its
  optional specialization through a sync.
