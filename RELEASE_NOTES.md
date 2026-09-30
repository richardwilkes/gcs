# Changes since the last release

## New & Improved

- GCS now works with screen readers: VoiceOver on macOS, Narrator, NVDA and JAWS on Windows, and Orca on Linux.
  Character sheets, lists, editors, settings and dialogs can be read and worked with one. What is shown only as a
  picture, an abbreviation or a color is put into words: a column headed by an icon is called by its name, such as
  Page Reference, Library Source or Equipped; Pts, SL and TL are read as Points, Skill Level and Tech Level; and the
  current encumbrance level is announced as such, along with whether more than the maximum load is being carried. A
  list says what its columns are as the keyboard focus arrives on it, the titled blocks of a sheet, such as Identity
  or Miscellaneous, and the header of each calculator are offered as headings, and page references are offered as
  links.
- A new general setting, "Include static text and disabled controls in the Tab order for screen readers", makes Tab
  stops of static text, read-only fields such as a hit location's penalty and DR, and disabled controls, so what a
  sheet or window shows can be read by moving the keyboard focus alone rather than with the screen reader's own review
  cursor. A label that names a field is left out, since it is read along with its field. A disabled control reached
  this way is announced as unavailable and still cannot be changed. The buttons a character sheet otherwise leaves out
  of the Tab order become Tab stops as well: the randomize buttons of the name and description, and the buttons that
  show or hide a section of the attributes or the sub-table of a hit location. Outside of dialogs, static text is
  outlined while it holds the keyboard focus. The setting is off by default, since a Tab order that runs through all
  of these is slower to move through, and it only takes effect while a screen reader is in use.
- Lists can now be worked a cell at a time from the keyboard. Press the right arrow on a row that has nothing to open
  to move into its cells, where the current cell is outlined. The left and right arrows move between the cells, Home
  and End go to the first and last, and the up and down arrows change rows while staying in the same column. Space
  works what the cell holds: on a page reference it opens the reference, and on a check mark, such as Equipped, it
  checks or unchecks it. On a cell holding nothing to press it opens the row's editor, as it does on a row. Escape, or
  the left arrow from the first cell, returns to the row as a whole.
- Context menus no longer need a mouse. Press shift+F10 or the Menu key while a list, the Library Explorer, a text
  field or a sheet whose layout is being edited has the keyboard focus to open its context menu. Screen readers also
  offer the context menu as an action on lists and their rows, the Library Explorer, document tabs and text fields.
- Page references shown as links can now be followed from the keyboard as well as with a click: those beside each
  option of the template picker dialog, those in the headers and text of the calculators, such as "(B104)", and those
  beside the checkboxes of the Sheet Settings. Tab goes to each link in turn, where it is outlined, and Space, Return
  or Enter follows it.
- The portrait on a character sheet can now be changed from the keyboard. Tab to it or click it, then press Space to
  choose a new portrait, as a double-click does. It is outlined while it holds the keyboard focus.
- Buttons that show only an icon now have tooltips saying what they do, such as those that add or remove a feature,
  prerequisite, default or study entry, show or hide the notes of a row, or edit a character's points.
- The title bar of each window now matches the light or dark mode GCS is set to.
- Library lists can now be filtered with saved, reusable filters, created and chosen from a new popup in the list's
  toolbar. A filter combines conditions on the list's own fields, such as a trait's name, tags, points or self-control
  roll, or a piece of equipment's cost or weight, using "match all of" and "match any of" groups that can be nested
  and negated. Filters are kept with your settings, so each one is available in every list of its kind. The "Names
  Only" checkbox and the tag popup have been removed, since a saved filter does what they did. The search field, now
  labeled Quick Filter, still filters as you type and now also looks at the tags column. It can be used along with a
  saved filter to narrow the list further: an item is shown only when it passes both.
- While a list is filtered, whether by a saved filter or by the Quick Filter, a hazard-striped banner across the top
  of the list notes that items cannot be added, removed or rearranged until the filter is cleared. Adding items used
  to be allowed while filtering and no longer is, and the buttons that open or close every row are turned off as well.
- A filtered list now keeps its hierarchy, showing each matching item beneath the containers that hold it rather than
  in a flat list, so you can see where a match sits. A container that is shown only because something inside it
  matched is dimmed, so the actual matches stand out. Containers are shown open while the list is filtered, and cannot
  be closed until the filter is cleared.
- The points column of a template now shows what a choice may come to, rather than the total of every option it
  offers. A choice of exactly 20 points shows 20, and where the cost is not settled until the choice is made a range
  such as "20~45" is shown, or "10+" or "≤-30" where only one end of it is known. A container holds the range of what
  is inside it. Ranges sort by their lower end, and the template picker dialog's running total shows a range while a
  picked option still presents choices of its own.
- Modifiers can now be applied to traits and equipment without drag & drop. Select the modifiers in a modifier library
  and choose Apply Modifier from the Edit menu or the list's context menu. You are asked which open character sheet,
  template or library list to apply them to, or loot sheet for equipment modifiers (skipped when only one qualifies),
  then which of its traits or equipment should receive them, each shown with the containers it sits in. Each receives
  its own copy of the modifiers, enabled, and is selected and revealed afterwards. A group or choice of modifiers
  applied to a character or loot sheet asks which of its contents should be enabled; elsewhere its contents keep the
  state they had in the library. A character sheet also asks for any @Name@ substitutions the modifiers need. Canceling
  any of the questions changes nothing, and the whole application is undone in a single step.
- Dropping modifiers onto traits or equipment no longer asks which of the recipient's modifiers should be enabled: the
  dropped modifiers simply arrive enabled. Dropping a group or choice of modifiers onto a character or loot sheet asks
  which of its contents should be enabled, and traits or equipment that arrive with modifiers of their own, as when a
  template is applied, are asked about as before.
- Copying or dragging items from a template or library onto a character sheet, or dragging them onto a loot sheet, now
  goes through the same steps as applying a template, in the same order: the template choices, the modifiers, the
  @Name@ substitutions and, on a character sheet, the ancestry question and the offer to randomize the profile again.
  Every question is asked before anything is added, so canceling any of them leaves the sheet exactly as it was, and
  the whole addition is undone in a single step. Applying a template follows suit: its ancestry question now comes
  after the others rather than first, and canceling any of its questions abandons the whole apply.
- Copying or dragging items from one character sheet to another now asks whether to disable the existing ancestry
  when an ancestry is among them, and offers to randomize the profile again, as applying a template does. Copying or
  dragging items from one template to another no longer asks about their modifiers or @Name@ substitutions.
- The questions asked while copying, dragging or applying now say what they are part of. Each has a title, such as
  "Apply Template: Modifiers", and a line saying what is under way, such as "Applying template Knight to Sir Bob". A
  question about a particular row names the containers the row sits in, and one asked of several rows in turn counts
  them, as in "Select Modifiers (1 of 2) for:". The offer to randomize the profile lists the fields it would replace.
- Template choices now live in their own kind of container, a choice, rather than being a setting of an ordinary
  container, whose editor no longer has a Choices row. In a template, the new "New Trait Choice", "New Skill Choice"
  and "New Spell Choice" commands in the Item menu add one, and "Convert to Choice" and "Convert to Group" in the Edit
  menu turn an existing container into a choice and back. A trait container must be a plain group to become a choice.
  Both conversions warn before removing anything. Since a choice is replaced by the options chosen from it when the
  template is applied, its editor shows only its name, notes, choices, page reference and page highlight. The choices
  in your existing templates become choice containers when the template is opened.
- Modifiers on a template choice are not supported, nor is anything else a choice has no use for. A choice is dissolved
  when the template is applied, and anything on it is left behind, so a modifier placed on one never reached the
  character. Opening a template, or copying or dragging rows into one, now removes from each choice its modifiers, VTT
  notes, user description, tags, self-control roll, frequency of appearance, prerequisites, library source and its
  preconfigured, disabled and switched on settings, and makes a trait choice a plain group. Put a modifier meant for
  every option on each option instead.
- Equipment containers now have a type. A container is still a piece of equipment that holds other equipment, such as
  a backpack, and every existing container stays one. A group only organizes the equipment in it: it has no quantity,
  value, weight, tech level, legality class, uses, rated ST, level, modifiers, weapons, features or prerequisites of
  its own, so it is worth and weighs just what its contents do. The new "New Carried Equipment Group" and "New Other
  Equipment Group" commands in the Item menu add one. In the Edit menu, "Convert to Group" turns a container into a
  group, warning first about what the group can't keep, and "Convert to Container" turns a group back into a
  container. A group can't be given modifiers, whether by dropping them onto it or with "Apply Modifier". Generating
  treasure from a loot sheet picks from what a group holds rather than from the group itself.
- Saved list filters can test whether an equipment container is a container or a group, and scripts can read it as the
  container's `kind`. In Go template exports, the Type of an equipment container is now "container" or "group", rather
  than always "group", so a template that tests for "group" to find equipment containers needs to allow for both. The
  legacy text export now gives CONTAINER or GROUP to match, and leaves a group's quantity, value and weight empty.
- Equipment can now be offered as a template choice. In a template, "New Equipment Choice" adds one, and "Convert to
  Choice" turns an equipment group into one. An equipment choice is picked by count, by value or by weight, where the
  value and weight counted are the extended ones, and its editor shows only its name, notes, choices, page reference
  and page highlight. The picker dialog shows each option's quantity, value and weight, and a choice picked by value
  or weight lets the quantity of each option that isn't itself a group be changed while picking. Until the choice is
  made, a template shows the value and weight the choice may come to as a range, and so do the containers holding it
  and the list's totals.
- A list's context menu no longer repeats what the list holds in each of its "New" commands. A trait list's now offers
  "New Trait", "New Container" and "New Choice", and an equipment list's "New Equipment", "New Container", "New Group"
  and "New Choice". The menu bar keeps the full names, since it can add to any list.
- Modifier containers are now called groups, and a group can be made a choice. A mandatory choice asks for exactly one
  of its modifiers, such as the one that sets the price of a trait whose cost varies, and an optional choice asks for at
  most one, which makes its modifiers mutually exclusive. "New Trait Modifier Choice" and "New Equipment Modifier
  Choice" add one, and "Convert to Choice" and "Convert to Group" turn a group into a choice and back, in a modifier
  library or in a trait or equipment editor.
- Unlike a template choice, a modifier choice keeps all of its modifiers wherever it goes. The one picked is the one
  enabled, and no more than one is ever enabled: turning one on, in the list or in its editor, turns off the one that
  was on. A new modifier added to a choice arrives turned off, and so does one duplicated, moved or dropped into a
  choice that already has its pick.
- When modifiers are asked about, a choice's modifiers are offered as radio buttons, and an optional choice adds
  "None". When a trait or piece of equipment is added to a character or loot sheet, a mandatory choice must be made
  before the prompt can be accepted; elsewhere, such as in a template, it may be left without a pick. A trait or piece
  of equipment marked preconfigured takes the picks its choices already have and is only asked about a mandatory choice
  that has none.
- On a character or loot sheet a mandatory choice always needs its pick. The pick can be changed but not turned off,
  so its Enabled box is disabled in the modifier's editor, and making a choice mandatory there, by converting a group
  or in the choice's editor, picks its first option. A mandatory choice left without a pick for any other reason, such
  as a library sync or deleting, moving or dragging its pick away, is flagged on both the item and the choice for you
  to resolve.
- In a template or library, a mandatory choice on an item that isn't marked preconfigured is asked about again when
  the item reaches a sheet, so any pick it has is only a default. Until then the points a trait may cost, and the value
  and weight of a piece of equipment, are shown as a range, in the lists and in the trait and equipment editors, and a
  trait container counts the same pick for everything inside it. A preconfigured item whose mandatory choices all have
  their picks shows a single figure. Wherever a single figure is needed, as in the totals, the least is counted.
- The template picker dialog can now make a pick's modifier choices, and the picks of a choice offered within the
  choice, as the picks are made: a button beside the row puts up the modifier prompt or that choice's own dialog. Once
  answered, the row shows a fixed cost and isn't asked about again when the template is applied. Until then a row headed
  for a character or loot sheet shows the range it may cost, rather than whatever its picks happen to be, and a
  preconfigured item is only asked about a mandatory choice that has no pick.
- The template picker dialog's running total now reads what the picks come to against the target, as in "40~70 / 60",
  and shows whether the rule is met, could still be met once the choices below are made, is met but something picked
  below needs attention, or isn't met. OK is only enabled when the rule is met; anything else needs Override. A choice
  within the choice whose picks miss its rule, or come to more or less than its rule expects, is marked in red on its
  row with a tooltip saying why, and a line under the list says how far off the picks are and what is left to do.
- The editor for a modifier group or choice now shows only the name, notes, choice, tags, page reference and library
  source. The other fields only ever applied to a modifier, and a group's VTT notes are no longer kept.
- The "Gives a weapon damage modifier of" feature now accepts dice as well as a plain number, so a trait, modifier or
  piece of equipment can add "+1d" or "+2d+1" to a weapon's damage, or take "-1d" away from it. The dice combine with
  the weapon's own damage the same way its base damage does, so a leading "-" takes away only the dice: "-1d+2"
  removes one die and still adds 2, just as it would in the weapon's damage. The "per level" and "per die" options
  multiply the dice just as they do a number. A percentage cannot be given in dice, so the "as a %" option is
  unavailable while the modifier has dice in it.
- The "Gives a reaction modifier of" and "Gives a conditional modifier of" features can now be given an optional
  group. On the character sheet, entries with a group are gathered inside a collapsible container named for the group
  in the Reaction Modifiers or Conditional Modifiers table, while entries without one stay where they were. The
  "Group containers when sorting" general setting decides whether the groups are listed ahead of the ungrouped
  entries or mixed in with them by name. The group may use the same @Name@ substitutions as the situation text. In
  exports, each entry now also carries its group: as the Group field in Go templates, and as @GROUP in the legacy
  text templates. (#282)
- A new sheet setting, "Disable traits whose prerequisites are unsatisfied", treats any trait whose prerequisites are
  not met, or whose level exceeds its maximum, as disabled: until they are met, it contributes no points, features or
  weapons to the sheet, and does not count toward the prerequisites of anything else. The trait keeps its own enabled
  state and the warning that explains what is missing, and comes back into play on its own once its prerequisites are
  satisfied. Where prerequisites contradict one another, such as two traits that each require the other's absence, no
  choice satisfies them all, so the traits caught in the contradiction are left enabled and flagged, with the
  contradiction explained in the flag's tooltip. The setting is off by default. In a saved sheet, the unsatisfied
  reason recorded for a trait the sheet disabled, or for one whose own prerequisites are unmet within a contradiction,
  ends with an explanation of that, and a trait whose own prerequisites are met but which is caught in a contradiction
  has that explained in the new prereq_contradiction field. Go template exports offer the latter as the
  PrereqContradiction trait field, and leave out a trait the sheet disabled as they do any other disabled trait. (#342)
- A sheet whose values can never settle, because its prerequisites, defaults, features or scripts depend on one
  another in a circle, now says so with a "Data never settles" notice in its toolbar.
- When the workspace arrangement is restored on start, the tab that had the keyboard focus when GCS was last quit is
  given the focus again, rather than the Library Explorer always starting with it. The files are also reopened in the
  order their tabs are laid out in, so the recent files list comes out the same from one start to the next.
- The Library Explorer's deep search now builds its index from the values saved in each file rather than recalculating
  every character sheet and running every note's scripts, so it takes far less work and is ready sooner after GCS
  starts or libraries are updated.
- The weight given to an option in the ancestry editor, or to a training name in the name generator editor, can now be
  as high as 999,999 rather than 9,999.
- Added the page reference key DFM6 for *Dungeon Fantasy Monsters 6: Tiny Terrors*.

## Bug Fixes

- The button that shows or hides a row's notes no longer pokes out past the edge of its column when the name beside it
  wraps onto more than one line.
- Dragging a divider in the sheet layout editor and letting go of it where it started no longer adds an edit to the undo
  history.
- Canceling a template's choices while creating a new character sheet from the template no longer leaves an empty
  sheet behind.
- A skill whose level comes from one of its defaults now shows the right level as soon as what it defaults from
  changes, as when a leveled trait raises an attribute, and so do the weapons that use the skill. They used to show
  the level from before the change until the next edit.
- Opening and closing the editor of a trait container without changing anything no longer asks whether to save
  changes.
- The tag describing a template choice, such as "Pick 1", no longer keeps a selected row's colors after its row is
  deselected. This happened when the template was changed while the row was selected.
- The Preconfigured setting, which marks an item whose modifiers are already settled so that they are not asked about
  again, can now be set in a library as well as in a template, and is kept when items are dragged into a library. It
  is never shown on a character or loot sheet, where it means nothing, and opening a sheet now clears any an older
  version left on nested items.
- Providing @Name@ substitutions for a modifier dropped onto a trait or piece of equipment on a character sheet no
  longer discards the substitutions that trait or equipment already had, including those for the other modifiers
  dropped with it and for its disabled modifiers. A substitution shared by several of the dropped modifiers is now
  asked for once for each trait or piece of equipment receiving them. The "Set Substitutions" button in the trait and
  equipment editors, copying or dragging items onto a sheet, and applying a template likewise no longer discard the
  substitutions of a disabled modifier, which showed its raw @Name@ markers when it was re-enabled.
- Items dropped into a closed container on a character sheet, loot sheet or template are now asked about their
  modifiers and @Name@ substitutions like any others; they used to be skipped. The container is opened to show them.
- Copying or dragging a template's choice, such as "Pick 60 points worth", onto a character sheet no longer brings it
  across with every option still in it and nothing left to make the choice. The choice is now made on the way, just as
  when applying the template, and only the options chosen arrive. Dragging a choice into a library now asks first, and
  removes the choices if you continue, since only a template can hold them.
- Character sheets and libraries that already hold template choices, left there by older versions, now have the
  choices removed when they are opened, since only a template can hold them. The options inside are kept.
- A template choice no longer keeps the library source of the container it was made from. Syncing with that source
  would have quietly removed the choices, since the library's copy has none.
- The question of whether to disable the character's existing ancestry, and the offer to randomize the profile again,
  are no longer put when the only new ancestry is one of the options of a template choice that was not chosen.
- The question of whether to disable the character's existing ancestry now names the ancestry containers involved, as
  they appear in the Traits list, rather than the ancestry each one uses, and lists all of them rather than only the
  first on each side. Several containers can use the same ancestry with details of their own, and it was impossible
  to tell which was which when both were shown as "Human".
- The script functions dice.add and dice.subtract now accept a bare modifier, such as "+3" or "-2", on either side, so
  that dice.add("1d-2", "+3") gives "1d+1" instead of failing with "dice sides must match". Only two specifications
  that both have dice of different sizes are still refused. (#1138)
- With the "Group containers when sorting" general setting turned on, sorting traits by points, equipment by
  quantity, value or weight, or reaction and conditional modifiers by their modifier now lists the containers first,
  in order of their values. They used to be sorted as though each held zero, which left them among the other rows.
- Right-clicking one of several selected rows on Windows and Linux no longer reduces the selection to that row, so the
  context menu's commands act on everything that was selected.
- Right-clicking the close button of a tab no longer closes the tab.
- Selecting a row near the bottom of a library list no longer scrolls the list when the row was already in view.
- Scrolling with the mouse wheel while dragging rows or a tab now shows the list scrolling as it goes, rather than
  leaving it as it was until the drop.
- A window now opens at least as wide as the widest of its tooltips, which were otherwise squeezed to fit it.
- Resetting the General Settings, or loading them from a file, now puts a change to "Group containers when sorting"
  or "Show additional page references when space allows" into effect right away, as changing it by hand does.
