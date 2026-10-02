# Changes since the last release

## New & Improved

### Accessibility & Keyboard

- GCS now works with screen readers: VoiceOver on macOS, Narrator, NVDA and JAWS on Windows, and Orca on Linux.
- A new general setting, "Include static text and disabled controls in the Tab order for screen readers", makes Tab
  stops of static text, read-only fields and disabled controls, so what a sheet or window shows can be read by moving
  the keyboard focus alone. It is off by default, since a Tab order that runs through all of these is slower to move
  through, and it only takes effect while GCS is serving a screen reader or other assistive technology.
- Lists can now be worked a cell at a time from the keyboard. Press the right arrow on a row that has nothing to open to
  move into its cells, where the arrow keys, Home and End move between them and Space works what the cell holds, such as
  a page reference or a check mark. Escape returns to the row.
- Context menus no longer need a mouse. Press shift+F10 or the Menu key while a list, the Library Explorer, a text field
  or a sheet whose layout is being edited has the keyboard focus. Screen readers also offer the context menu as an
  action.
- Page references shown as links, such as those in the template picker dialog, the calculators and the Sheet Settings,
  can now be followed from the keyboard: Tab to the link and press Space, Return or Enter.
- The portrait on a character sheet can now be changed from the keyboard: Tab to it and press Space.
- Buttons that show only an icon now have tooltips saying what they do.
- Markdown documents, such as a library's .md files and release notes and the License in the Help menu, can now be read
  with a caret and copied from, with the usual mouse and keyboard selection, Copy and Select All in the Edit menu, and
  Space or Return on a link to follow it.

### Appearance

- The title bar of each window now follows the Color Mode chosen in the Colors settings. On macOS the menus and the open
  and save dialogs follow it too, and on Linux the title bar follows it where the window manager supports that.
- A new color, Needs Attention, has been added to the Colors settings. The template picker dialog uses it for what still
  has choices to be made.
- The factory sheet layout now gives the attributes and the hit location table a little more width and the encumbrance
  and lifting column a little less, and makes the Skills list a little narrower than the Traits list beside it. Existing
  sheets keep their layout. To adopt the new one, choose "Reset the Default Layout to Factory Settings" from the block
  layout menu in a sheet's toolbar, then "Reset Layout to Default" on any sheet that should follow it.

### Lists & Filtering

- Library lists can now be filtered with saved, reusable filters, created and chosen from a new popup in the list's
  toolbar. A filter combines conditions on the list's fields, such as a trait's name, tags or points, in "match all of"
  and "match any of" groups that can be nested and negated, and is kept with your settings for use in every list of its
  kind. The "Names Only" checkbox and the tag popup have been removed, since a saved filter does what they did. The
  search field, now labeled Quick Filter, also looks at the tags column, and can be used along with a saved filter to
  narrow the list further.
- While a list is filtered, a hazard-striped banner across the top of the list notes that items cannot be added, removed
  or rearranged until the filter is cleared. Adding items used to be allowed while filtering and no longer is.
- A filtered list now keeps its hierarchy, showing each match beneath the containers that hold it. A container shown
  only because something inside it matched is dimmed, and containers are held open until the filter is cleared.
- A list's context menu no longer repeats what the list holds in its "New" commands, offering "New Container" rather
  than "New Trait Container", for instance, and in a template adds "New Choice". The menu bar keeps the full names.

### Template Choices

- Template choices now live in their own kind of container, a choice, rather than being a setting of an ordinary
  container, whose editor no longer has a Choices row. In a template, "New Trait Choice", "New Skill Choice" and "New
  Spell Choice" in the Item menu add one, and "Convert to Choice" and "Convert to Group" in the Edit menu turn a
  container into a choice and back, warning before removing anything; a trait container must be a plain group to become
  a choice. A choice's editor shows only its name, notes, choices, page reference and page highlight, its type can no
  longer be set to Not Applicable, and one made by count can no longer ask for fewer than 0. The choices in your
  existing templates become choice containers when the template is opened, and a count below 0 is raised to 0.
- Anything a choice has no use for is removed from it when a template is opened or rows are copied or dragged into one:
  modifiers, VTT notes, user description, tags, self-control roll, frequency of appearance, prerequisites, library
  source and the preconfigured, disabled and switched on settings, and a trait choice is made a plain group. A choice is
  dissolved when the template is applied, so none of these ever reached the character. Put a modifier meant for every
  option on each option instead.
- The points column of a template now shows what a choice may come to rather than the total of every option it offers: a
  single figure where the cost is settled, otherwise a range such as "20~45", "10+" or "≤-30". A container holds the
  range of what is inside it, ranges sort by their lower end, and the template picker dialog's running total shows a
  range while a picked option still has choices of its own.
- Equipment can now be offered as a template choice. In a template, "New Equipment Choice" adds one, and "Convert to
  Choice" turns an equipment group into one. An equipment choice is picked by count, by extended value or by extended
  weight, and is always equipped. The picker dialog shows each option's quantity, value and weight, and a choice picked
  by value or weight lets the quantity of each option be changed while picking. Until the choice is made, the value and
  weight it may come to are shown as a range.
- A group inside a template choice can now offer what it holds one by one rather than being picked as a unit. Uncheck
  "Picked as a Unit" in its editor, shown only for a group in a template choice. The choice then counts and costs the
  options inside the group, which the picker dialog shows indented beneath the group's name. On the character, each such
  group keeps only what was picked from it, and one with nothing picked is left out.

### Equipment Groups

- Equipment containers now have a type. A container is still a piece of equipment that holds other equipment, such as a
  backpack, and every existing container stays one. A group only organizes the equipment in it, with no quantity, value,
  weight, modifiers, weapons, features or other properties of its own, so it is worth and weighs just what its contents
  do. "New Carried Equipment Group" and "New Other Equipment Group" in the Item menu add one, and "Convert to Group" and
  "Convert to Container" in the Edit menu turn one into the other, warning first about what a group can't keep. A group
  can't be given modifiers, and generating treasure from a loot sheet picks from what a group holds rather than the
  group itself.
- Saved list filters can test whether an equipment container is a container or a group, and scripts can read it as the
  container's `kind`. In Go template exports, the Type of an equipment container is now "container" or "group" rather
  than always "group", so a template that tests for "group" to find equipment containers needs to allow for both. The
  legacy text export gives CONTAINER or GROUP to match, and leaves a group's quantity, value and weight empty.

### Modifiers

- Modifier containers are now called groups, so "New Trait Modifier Container" and "New Equipment Modifier Container"
  are now "New Trait Modifier Group" and "New Equipment Modifier Group", keeping their shortcuts. A modifier group can
  be made a choice: a mandatory choice asks for exactly one of its modifiers, such as the one that sets the price of a
  trait whose cost varies, and an optional choice asks for at most one, which makes its modifiers mutually exclusive. A
  choice shows a "Pick 1" or "Pick at most 1" tag on its row. "New Trait Modifier Choice" and "New Equipment Modifier
  Choice" add one, and "Convert to Choice" and "Convert to Group" turn a group into a choice and back.
- Unlike a template choice, a modifier choice keeps all of its modifiers wherever it goes. The one picked is the one
  enabled, and turning one on turns off the one that was on. A modifier added, duplicated, moved or dropped into a
  choice that already has its pick arrives turned off.
- When modifiers are asked about, a choice's modifiers are offered as radio buttons, with "None" added wherever the pick
  may be left out. When a trait or piece of equipment is added to a character or loot sheet, a mandatory choice must be
  made before the prompt can be accepted; elsewhere, such as in a template, it may be left for later. A preconfigured
  item keeps the picks its choices already have and is only asked about a mandatory choice that has none.
- On a character or loot sheet a mandatory choice always needs its pick, which can be changed but not turned off, so its
  Enabled box is disabled in the modifier's editor. Making a choice mandatory there picks its first option if none is
  on. A mandatory choice left without a pick for any other reason, such as a library sync or deleting or moving its pick
  away, is flagged with a "Modifier choice required" tag on the item and a "Required" tag on the choice.
- In a template or library, a mandatory choice on an item that isn't marked preconfigured is asked about again when the
  item reaches a sheet, so any pick it has is only a default. Until then the points a trait may cost, and the value and
  weight of a piece of equipment, are shown as a range in the lists and editors. A preconfigured item whose mandatory
  choices all have their picks shows a single figure, and wherever a single figure is needed, as in the totals, the
  least is counted.
- The editor for a modifier group or choice now shows only the name, notes, choice, tags, page reference, page highlight
  and library source. The other fields only ever applied to a modifier, and a group's VTT notes are no longer kept.
- The Preconfigured setting, which marks an item whose modifiers are already settled so that they are not asked about
  again, can now be set in a library as well as in a template, and is kept when items are dragged into a library. It is
  never shown on a character or loot sheet.
- Modifiers can now be applied to traits and equipment without drag & drop. Select the modifiers in a modifier library
  and choose Apply Modifier from the Edit menu or the list's context menu, then pick the open sheet, template or library
  list and the traits or equipment that should receive them. A filtered or empty list is not offered, nor are choices or
  equipment groups. Each recipient gets its own copy of the modifiers, enabled. A group or choice of modifiers applied
  to a character or loot sheet asks which of its contents should be enabled, and a character sheet asks for any @Name@
  substitutions. Canceling any question changes nothing, and the whole application is undone in a single step.
- Dropping modifiers onto traits or equipment no longer asks which should be enabled: they arrive enabled. Dropping a
  group or choice of modifiers onto a character or loot sheet still asks which of its contents should be enabled, as do
  traits or equipment that arrive with modifiers of their own. On a character sheet the @Name@ question still follows,
  and canceling it removes the dropped modifiers.

### Template Picker Dialog

- A container picked as a unit in the template picker dialog can now be opened to show what it holds. A choice inside it
  is shown by its name and rule, such as "(pick 1)", since its options are picked in its own dialog.
- The template picker dialog can now make a pick's modifier choices, and the picks of a choice offered within the
  choice, as the picks are made: a button beside the row, shown in the Needs Attention color while something is still
  open, puts up the modifier prompt or that choice's own dialog. The prompt has Clear Selections and Override buttons,
  and Override keeps what is answered so far and leaves the rest for when the template is applied. Once answered, the
  row shows a fixed cost and isn't asked about again when the template is applied; until then it shows the range it may
  cost.
- The template picker dialog's running total now reads what the picks come to against the target, as in "40~70 / 60",
  and shows whether the rule is met, could still be met, is met but something picked needs attention, or isn't met. OK
  is only enabled when the rule is met; anything else needs Override. A choice within the choice whose picks miss its
  rule is marked in red with a tooltip saying why, and a notice under the list says what is left to do.

### Copying, Dragging & Applying Templates

- Copying or dragging items from a template or library onto a character or loot sheet now goes through the same steps as
  applying a template, in the same order: the template choices, the modifiers, the @Name@ substitutions and, on a
  character sheet, the ancestry question and the offer to randomize the profile. Every question is asked before anything
  is added, so canceling leaves the sheet as it was, and the whole addition is undone in a single step. Applying a
  template now asks its ancestry question after the others, and canceling any of its questions abandons the whole apply.
- Copying or dragging items from one character sheet to another now asks whether to disable the existing ancestry when
  an ancestry is among them, and offers to randomize the profile. Copying or dragging items from one template to another
  no longer asks about their modifiers or @Name@ substitutions.
- The questions asked while copying, dragging or applying now say what they are part of, with a title such as "Apply
  Template: Modifiers", a line saying what is under way, and the containers the row in question sits in.

### Features & Prerequisites

- The "Gives a weapon damage modifier of" feature now accepts dice as well as a plain number, so a feature can add "+1d"
  or "+2d+1" to a weapon's damage or take "-1d" away. The dice combine with the weapon's own damage the same way its
  base damage does, and the "per level" and "per die" options multiply the dice just as they do a number. The "as a %"
  option is unavailable while the modifier has dice in it.
- The "Gives a reaction modifier of" and "Gives a conditional modifier of" features can now be given an optional group.
  On the character sheet, entries with a group are gathered inside a collapsible container named for the group in the
  Reaction Modifiers or Conditional Modifiers table, and the "Group containers when sorting" general setting decides
  whether the groups are listed ahead of the ungrouped entries or mixed in with them by name. The group may use the same
  @Name@ substitutions as the situation text, and exports carry it as the Group field in Go templates and @GROUP in the
  legacy text templates. (#282)
- A new sheet setting, "Disable traits whose prerequisites are unsatisfied", treats any trait whose prerequisites are
  not met, or whose level exceeds its maximum, as disabled until they are: it contributes no points, features or weapons
  to the sheet and does not count toward the prerequisites of anything else. The trait keeps its own enabled state and
  the warning that explains what is missing. Where prerequisites contradict one another, such as two traits that each
  require the other's absence, the traits caught in the contradiction are left enabled and flagged with a "Contradictory
  prerequisite(s)" tag whose tooltip explains it. The setting is off by default. A saved sheet records the explanation
  at the end of the trait's unsatisfied reason or in the new prereq_contradiction field, which Go template exports offer
  as PrereqContradiction, and both the Go template and legacy text exports leave out a trait the sheet disabled as they
  do any other disabled trait. (#342)
- A sheet whose values can never settle, because its prerequisites, defaults, features or scripts depend on one another
  in a circle, now says so with a "Data never settles" notice in its toolbar.

### Other

- When the workspace is restored on start, the tab that had the keyboard focus when GCS was last quit gets it again, and
  files are reopened in the order of their tabs, so the recent files list comes out the same from one start to the next.
- The Library Explorer's deep search now builds its index from the values saved in each file rather than recalculating
  every character sheet and running every note's scripts, so it is ready sooner after GCS starts or libraries are
  updated.
- The weight given to an option in the ancestry editor, or to a training name in the name generator editor, can now be
  as high as 999,999 rather than 9,999.
- Added the page reference key DFM6 for *Dungeon Fantasy Monsters 6: Tiny Terrors*.

### Older Versions of GCS

- Files saved by this version still open in older versions of GCS, but those versions know nothing of equipment groups
  and choices, modifier choices, groups picked from separately, dice in a weapon damage modifier, groups of reaction and
  conditional modifiers, or the new prerequisite sheet setting. They treat a group as an ordinary container and a choice
  as a plain group, keep only the plain number of a damage modifier given in dice, and ignore the rest. Saving a file
  there loses all of that for good.

## Bug Fixes

### Templates & Choices

- Canceling any of the questions asked while creating a new character sheet from a template no longer leaves an empty
  sheet behind.
- The tag describing a template choice, such as "Pick 1", no longer keeps a selected row's colors after the row is
  deselected.
- Copying or dragging a template's choice onto a character sheet now makes the choice on the way, as applying the
  template does, so only the options chosen arrive rather than every option with nothing left to choose. Dragging a
  choice into a library now asks first, and removes the choices if you continue, since only a template can hold them.
- Copying or dragging a skill or spell into a template no longer folds its points into a matching option of one of the
  template's choices, which turned a required skill into an optional one, and copying a choice into a template no longer
  loses an option that matches a skill or spell the template already requires.
- Character sheets and libraries that hold template choices left there by older versions now have the choices removed
  when they are opened, since only a template can hold them. The options inside are kept.
- A template choice no longer keeps the library source of the container it was made from, since syncing with that source
  would have quietly removed the choices.
- The question of whether to disable the character's existing ancestry, and the offer to randomize the profile again,
  are no longer put when the only new ancestry is an option of a template choice that was not chosen.
- The question of whether to disable the character's existing ancestry now names the ancestry containers involved, as
  they appear in the Traits list, and lists all of them rather than only the first on each side, since several
  containers can use the same ancestry.

### Modifiers & Substitutions

- Opening a character or loot sheet now clears the Preconfigured setting an older version left on items nested inside
  others, which quietly skipped the modifier and @Name@ questions for modifiers dropped onto them.
- Providing @Name@ substitutions for a modifier dropped onto a trait or piece of equipment on a character sheet no
  longer discards the substitutions it already had, including those of its disabled modifiers, and a substitution shared
  by several of the dropped modifiers is asked for once per recipient. The "Set Substitutions" button in the trait and
  equipment editors, copying or dragging items onto a sheet, and applying a template likewise keep the substitutions of
  a disabled modifier.
- Items dropped into a closed container on a character sheet, loot sheet or template are now asked about their modifiers
  and @Name@ substitutions like any others, and the container is opened to show them.

### Sheets & Editors

- A skill whose level comes from one of its defaults, and the weapons that use it, now show the right level as soon as
  what it defaults from changes, as when a leveled trait raises an attribute, rather than at the next edit.
- Opening and closing the editor of a trait container without changing anything no longer asks whether to save changes,
  and the blank Alternative Slots field of a container that isn't a set of alternative abilities no longer has a red
  error outline.
- Dragging a divider in the sheet layout editor and letting go of it where it started no longer adds an edit to the undo
  history.
- The choices offered by the popups of the calculators, such as the Hiking calculator's terrain and weather, now appear
  in the language chosen in the General Settings rather than the system's language.
- The script functions dice.add and dice.subtract now accept a bare modifier, such as "+3" or "-2", on either side, so
  that dice.add("1d-2", "+3") gives "1d+1" instead of failing with "dice sides must match". (#1138)

### Lists

- The button that shows or hides a row's notes no longer pokes out past the edge of its column, nor pushes the row's
  tag, such as "Pick 1", out of the cell, when the name beside it wraps onto more than one line.
- With "Group containers when sorting" turned on, sorting traits by points, or equipment by quantity, value or weight,
  now keeps the containers together ahead of the other rows in order of their values, or after them when the sort is
  reversed, rather than sorting each as though it held zero.
- Right-clicking one of several selected rows on Windows and Linux no longer reduces the selection to that row.
- Selecting a row near the bottom of a library list no longer scrolls the list when the row was already in view.

### Windows & Workspace

- Right-clicking the close button of a tab no longer closes the tab.
- Closing the last tab in the workspace now moves the keyboard focus to the Library Explorer, and pressing Tab when
  nothing has the focus now brings the focus into the window.
- Scrolling with the mouse wheel while dragging rows or a tab on macOS or Windows now shows what is under the pointer
  scrolling as it goes, rather than leaving it as it was until the drop.
- A document or editor shown in a window of its own now opens at least as wide as the widest of its tooltips, which were
  otherwise squeezed to fit it.
- Resetting the General Settings, or loading them from a file, now puts a change to "Group containers when sorting" or
  "Show additional page references when space allows" into effect right away.
