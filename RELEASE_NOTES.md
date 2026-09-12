# Changes since the last release

## New & Improved

- The "Gives a weapon damage modifier of" feature now accepts dice as well as a plain number, so a trait or modifier
  can add "+1d" or "+2d+1x3" to a weapon's damage, or take "-1d" away from it. The dice combine with the weapon's own
  damage the same way its base damage does, so a leading "-" takes away only the dice: "-1d+2" removes one die and
  still adds 2, just as it would in the weapon's damage. The "per level" and "per die" options multiply the dice just
  as they do a number. A percentage cannot be given in dice, so the "as a %" option is unavailable while the modifier
  has dice in it.
- Library lists can now be filtered with saved, reusable filters, chosen from the new Saved Filters popup in the
  list's toolbar. A filter combines conditions on the list's own fields, such as a trait's name, tags, points or
  self-control roll, or a piece of equipment's cost or weight, using "match all of" and "match any of" groups that can
  be nested and negated. Filters are kept with your settings, so each one is available in every list of its kind. The
  "Names Only" checkbox and the tag popup have been removed; the search field still filters as you type and now also
  looks at the tags column. The search field stays available while a saved filter is chosen, and narrows the list
  further: an item is shown only when it passes both the saved filter and what you have typed.
- While a filter is applied, a hazard-striped banner across the top of the list notes that items cannot be added,
  removed or rearranged until the filter is cleared, since only part of the list is in view.
- A filtered list now keeps its hierarchy, showing each matching item beneath the containers that hold it rather than
  in a flat list, so you can see where a match sits. A container that is shown only because something inside it
  matched is dimmed, so the actual matches stand out.
- Deep search now indexes your libraries from the values recorded in each file when it was last saved, instead of
  recalculating every character sheet and running every note's scripts to build its index. Starting GCS, and each
  library update after that, now does far less work.
- A new sheet setting, "Disable traits whose prerequisites are unsatisfied", treats any trait whose prerequisites are
  not met, or whose level exceeds its maximum, as disabled: it contributes no points, features or weapons to the sheet
  until they are met. The trait keeps its own enabled state and the warning that explains what is missing, and comes
  back into play on its own once its prerequisites are satisfied. The setting is off by default.
- When the workspace arrangement is restored on start, the tab that had the keyboard focus when GCS was last quit is
  given the focus again, rather than the Library Explorer always starting with it. The files are also reopened in the
  order their tabs are laid out in, so the recent files list comes out the same from one start to the next.

## Bug Fixes

- (none yet)
