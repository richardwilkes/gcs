# Changes since v5.44.0

## New & Improved

- Improved the merging of points for identical skills and spells added to a character sheet, so a matching entry adds
  its points to the existing one rather than creating a duplicate row. This now covers additional cases: entries with a
  tech level, entries whose nameable substitutions are resolved as they are added, duplicate entries within a single
  template, and skills or spells dragged or copied from another character sheet.
- Added the ability for a hidden attribute to be revealed with a chosen placement (Automatic, Primary, or Secondary)
  whenever the character has a named trait. The attribute settings now show "Placement [Hidden] unless trait [name] is
  present, then [placement]". (#845)
- Added support for internal anchor links in Markdown. Headings are now automatically assigned anchors, and links to
  them (e.g. `[Scripting](#scripting)`, or page references such as `md:User Guide/Scripting Guide#code`) scroll the
  target heading to the top of the view, revealing the section it introduces. (#651)
- Added support for a comma-separated list of tags in a feature's tag qualifier, so a single feature can grant its
  bonus to anything matching any one of the listed tags without stacking (e.g. `Sword, Axe, Polearm`). This applies to
  every tag-based criteria, including skill, spell, and their point bonuses, weapon bonuses, and equipment
  prerequisites. (#1008)
- Added a new "Sets the value of" feature for replacing a field with a chosen value, rather than adjusting it by a
  number or toggling a flag. Because such a field can hold only one value, these are resolved absolutely
  instead of stacking: when more than one applies to the same field, the one with the highest priority wins, ties are
  broken in favor of the more specific match, and the winning value along with the ones it overrode is shown in the
  tooltip.
- Added a new "Gives equipment a maximum uses modifier of" feature for raising or lowering a piece of equipment's
  maximum uses. As with equipment modifier costs, the operation is taken from what you enter: a plain number adds (e.g.
  `1` or `-1`), a value ending in `%` adjusts by a percentage (e.g. `-10%`), and a value with an `x` multiplies (e.g.
  `x2`). It can optionally scale per level, and can apply to the equipment it is attached to ("to this equipment") or to
  other equipment matched by name and tags ("to equipment whose name"). The resolved maximum is always kept within the
  range 0 to 9,999,999. When such a feature lowers the maximum below a piece of equipment's current remaining uses, the
  remaining uses shown (and adjusted by the uses commands) are now capped at the new maximum; the stored value is left
  untouched until you change it or save the file, at which point it is brought into range.
- Added a "Reset Uses to Maximum" command alongside the existing "Increase Uses" and "Decrease Uses" commands, available
  from the menus, the equipment context menu, and as an assignable key binding.
- Added a maximum level to leveled traits, shown alongside the Level and Cost Per Level fields in the trait editor. It
  may be a plain number or a script expression (e.g. one that varies with SM or ST), and the editor displays the
  resolved value. A trait whose level exceeds its maximum is flagged on the character sheet with the same warning used
  for unmet prerequisites, and points are still computed from the actual level. When picking a trait's level while
  applying a template, the maximum is shown and the pickable level is capped to it. A companion "Gives a trait maximum
  level modifier of" feature can raise or lower the maximum without editing the trait's definition (preserving library
  source sync): as with the equipment uses adjustment, a plain number adds, a value ending in `%` adjusts by a
  percentage, and a value with an `x` multiplies; it can optionally scale per level, and can apply to the trait it is
  attached to ("to this trait") or to other traits matched by name and tags ("to traits whose name"). (#1060)
- Added a "Cursor Size" setting to the General Settings, allowing the size of the mouse cursor to be adjusted. Also
  added two new theme colors to control the foreground and background colors of the cursors. In addition, cursors are
  now drawn from their vector artwork at each display's exact resolution rather than being scaled from a single
  fixed-size image, making them noticeably sharper on Windows systems using a fractional display scale and on
  non-Retina Mac displays.
- GCS is now built entirely from Go. The window management, drawing, text layout, and PDF generation previously done by
  a bundled native graphics library, and the PDF display previously done by a bundled native PDF library, are now done
  by pure Go replacements, so GCS no longer contains any platform-specific native code. On Linux, only the standard X11
  and OpenGL client libraries (libX11.so.6 and libGL.so.1) are needed at runtime; fontconfig and freetype are no longer
  used. The list of available fonts is now built by scanning the fonts installed on the system directly, so the first
  launch after upgrading may take a moment longer while the font index is built; it is cached for later launches.
- The PDF viewer used for page references was rebuilt on the new pure-Go engine. Page display, text search, links, the
  table of contents, and password-protected documents all work as before, and the new engine was developed and tested
  against the old one to keep page layout, search hit positions, and link locations identical. Text is drawn with a
  slightly heavier weight, tuned to match the way macOS Preview renders text, and documents that don't embed their
  fonts now substitute from a font set bundled with GCS rather than using system fonts, so some PDFs may look a little
  different than they did before. Exporting a sheet to PDF goes through the same new machinery, and a failure while
  writing the exported file is now reported instead of silently leaving a truncated file behind.
- PDFs are now displayed with continuous scrolling: pages are stacked one after another and you scroll straight through
  the document instead of viewing a single page at a time. The page number field, the navigation buttons, and the
  Back/Forward history follow the page currently at the top of the view, pages just outside the view are rendered ahead
  of time so scrolling stays smooth, and search highlights and links work on every visible page rather than only the
  current one.
- Opening a PDF no longer freezes GCS while the file is read. A large PDF — especially one the operating system has to
  retrieve from cloud storage first — could previously lock up the application for a long time before the view
  appeared. The view now opens immediately and fills in once the document is ready, showing a "Loading…" message if
  that takes more than a moment, and a PDF that can't be opened now reports the failure in the view instead of failing
  silently. A PDF opened in a window of its own also now gets a window sized to show a full page.
- When a working hardware-accelerated (OpenGL) display isn't available — in a virtual machine, over a remote desktop
  session, or with missing or broken graphics drivers — GCS now falls back to drawing with the processor instead of
  failing to open a window. Drawing is slower this way, but the application runs.

## Bug Fixes

- Fixed the handling of data files containing a feature or prerequisite type this version of GCS doesn't recognize, such
  as one written by a newer release. Previously, an unrecognized feature was silently loaded as an empty attribute bonus
  and an unrecognized prerequisite as an empty, always-satisfied prerequisite list, and saving the file then wrote those
  replacements back out, permanently destroying the original data. Such entries are now kept exactly as they were read
  and written back out unchanged. They are shown in the editors so you can see they're present (and delete them if you
  want to), an unrecognized feature has no effect, and an unrecognized prerequisite is reported as unsatisfied rather
  than being quietly treated as met.
- Fixed double-counting of +3 and defense bonus for parry/block weapon defaults. This now also covers the case where
  the default is for the other defense: a parry-type default feeding a weapon's block (or a block-type default feeding
  its parry) is resolved from the skill it names, so it is halved once and given the bonus belonging to the defense
  being computed, rather than being halved a second time and carrying the wrong defense's bonus. Adjustments to the
  weapon's skill — the penalty for a minimum ST higher than your own, and bonuses aimed at this weapon's skill — are
  now folded into the skill level before it is halved, rather than being applied to the already-halved defense, where
  they counted for twice as much. A weapon that names the same skill both ways, as the Tonfa in the High Tech library
  does, no longer shows an inflated parry or block.
- Fixed another case of the key handling not properly masking out sticky modifier keys (CapsLock & NumLock), this time
  affecting use of the Tab key to move focus between fields.
- Fixed spell prerequisite counting (for things like "6 spells from the Air college") so that a spell which itself
  requires the spell being checked is no longer counted toward that spell's own prerequisites, avoiding a circular
  prerequisite relationship. (#737)
- Fixed the display of a skill whose optional specialization resolves to an empty string, so it no longer shows an empty
  set of parentheses `()` after the skill name.
- Fixed the Linux desktop integration so the application window is correctly associated with its launcher icon (added
  the missing `StartupWMClass` and the matching window `WM_CLASS`). (#1059)
- Fixed a technique's skill default so that switching its specialization comparison to "whose specialization is
  anything" now matches any specialization, rather than staying locked on the specialization text left over from a prior
  "is" selection. (#1061)
- Fixed the scripting skill lookups (`entity.findSkills`, `entity.skillLevel`, and a skill container's `find`) so a
  specialization argument now matches a skill by its optional specialization as well as its required specialization.
  (#1062)
- Fixed a rare crash that could occur while scrolling when blank space at the edge of the content was collapsed.
  Scrolled views (tables, lists, and other scrollable panels) now settle correctly instead of getting caught in a loop.
- Fixed a startup failure on Linux where the application window could fail to be created (a `BadMatch` error) on some
  graphics drivers, most notably NVIDIA.
- Fixed saving and copying files failing on network drives (such as certain SMB/CIFS mounts) that don't allow changing
  file permissions. Preserving the original file permissions is now best-effort, so an otherwise-valid save or copy no
  longer aborts when the drive rejects the permission change.
- Fixed the missing row highlight when dragging a modifier over a trait or piece of equipment, so the row that will
  receive the drop is highlighted again on character sheets, loot sheets, templates, and in the list editors. (#1065)
- Linux only: Fixed dialog windows losing proper focus and trapping the mouse cursor at the corner of the screen when
  running under XWayland. (#1064)
- Linux only: Fixed modifier keys (Shift, Control, etc.) sometimes not being detected as pressed. (#1069)
- Fixed dialogs and newly opened windows sometimes appearing on the primary display instead of the display being worked
  on, which was most noticeable on Linux and Windows when opened from a menu. New windows and dialogs are now placed
  relative to the frontmost window, falling back to the primary display only when no window is available. This also
  covers the system Open and Save dialogs on Windows.
- Fix unary operator handling and panics in legacy expression conversion.
- Fixed various issues in the legacy text export templates, among them the tag include and exclude filters never
  matching a tag that contains a colon. Only the individual colon-separated portions of a tag were compared, so naming a
  full tag such as "Advantage: Mental" matched nothing; the whole tag is now matched as well.
- Fixed two problems with the settings views when they are in windows of their own, which is the case when the Settings
  group is set to open in its own window: the Page Reference Mappings view didn't show the newly chosen PDF for a
  mapping until it was closed and reopened, and the Attributes, Body Type, and Sheet Settings views for a character
  sheet or template were left open when the sheet or template they belong to was closed.
- Fixed the cost shown for a leveled trait modifier so it reflects the modifier's current level. A modifier costing
  +10% per level and set to 3 levels displayed +10% in the modifier list, even though the points charged were correct;
  it now displays the total (+30%). Also fixed a modifier set to take its level from the trait it is attached to
  showing no level or level-scaled cost in the trait editor's modifier list until something else forced a
  recalculation. (#1079) The cost of such a "use level from owner" modifier was further computed from the trait's
  current level rather than the levels actually paid for. Levels granted to the trait by a feature are free, so they no
  longer add enhancement or limitation percentages to its cost. The level the modifier displays, and that its own
  per-level features use, still reflects the granted levels.
- Fixed a trait whose self-control roll is set to "Never resist" not showing its "No CR" notation after the trait's
  name.
- Fixed a crash when a trait has a per-level bonus that adjusts its own level, or two traits each adjust the other's
  level. The adjustment is now resolved from the trait's own unadjusted level, breaking the cycle.
- Fixed the point cost of an Alternative Abilities container whose children all have negative point costs. Every child
  was billed at 20% of its cost because no "most expensive" child was ever found; the most expensive child is now
  charged in full and the rest at 20%, as already happened when the children have positive costs.
- Fixed a trait container's modifiers being temporarily taken over by whichever child was having its points computed.
  This could make a modifier on the container display its nameable substitutions using the wording of the last child
  costed, and let the container's substitutions leak into a child. The container's modifiers now stay with the
  container while still being costed against each child that inherits them.
- Fixed the nameable substitutions of an equipment modifier being taken from the wrong item, or lost outright, when a
  piece of equipment is duplicated or edited. A duplicate kept its modifiers pointed at the original, so they showed the
  original's substitutions and followed along as it was changed — most visible in an equipment list, where nothing later
  puts it right. In older data files, where such substitutions are still stored on the modifier rather than on the
  equipment it belongs to, they were moved onto the original instead of the duplicate, leaving the duplicate showing raw
  placeholders such as "@Material@"; opening one of these items in the equipment editor and applying the changes
  discarded them entirely.
- Fixed a hidden weapon usage disappearing when another usage on the same character was identical to it apart from
  being hidden. Only one of the two was listed, making it impossible to reveal the hidden one again; both now appear in
  the places where the Hide checkbox is shown.
- Fixed a weapon bulk bonus inventing a "giant" bulk value for a weapon that doesn't have one: a weapon with a bulk of
  -3 given a -1 bonus showed -4/-1 instead of -4. A weapon that actually has a separate giant bulk still has both
  values adjusted.
- Fixed a percentage armor divisor bonus replacing a weapon's armor divisor with the percentage of it instead of
  adjusting it by that percentage: a +10% bonus on a (2) divisor produced (0.2) rather than (2.2).
- Fixed the reload time being dropped from a weapon's Shots value whenever something separates it from the shot count,
  such as a thrown weapon's "T(1)" or a duration like "3x3s(2)". The reload time was silently lost each time the file
  was saved and reloaded; it is now kept.
- Fixed the Tbone damage progressions losing the strength-based part of a weapon's damage when the weapon's base damage
  uses dice with a different number of sides — a 1d3 weapon showed 2d3 instead of 2d3+2, for example. The Basic Set
  progression was unaffected.
- Fixed a weapon's per-level base damage being scaled by the level count twice when it includes a dice multiplier. A
  trait whose weapon does "1dx2" per level showed 3dx6 at 3 levels instead of 3dx2, tripling the damage again on top of
  the dice it had already tripled. Per-level damage without a multiplier was unaffected.
- Fixed multipliers entered with a capital "X" (e.g. "X2") or the multiplication sign "×" (e.g. "×2") in equipment
  modifier cost and weight fields and trait modifier cost fields. They were recognized as multipliers, but the number
  was not read, silently turning a cost into x1 and a weight or trait modifier cost into x0. All three spellings now
  work, whether the marker comes before or after the number.
- Fixed a "to this armor" DR bonus being applied once per matching DR entry, rather than once in total, when a single
  piece of equipment provides several DR entries covering the same hit location. Hit locations that differ only in
  capitalization are now also treated as the same location.
- Fixed the skill penalty applied when a skill's required equipment is missing (-5, or -10 for a tech-level skill) also
  being applied to other copies of that skill that differ only in their optional specialization. The penalty now
  applies solely to the skill whose prerequisite is unmet.
- Fixed the "contained weight" prerequisite for a container whose quantity is greater than one. The weight tested
  included the container's own weight multiplied by its quantity rather than just the contents of a single container,
  so the prerequisite could be reported as unmet when it was satisfied.
- Fixed the level shown for a ritual magic spell when the character doesn't have the ritual magic skill it is based on.
  The unresolvable level was treated as 0, so spell bonuses such as Magery could raise it to a level the character
  hadn't earned, and a negative net bonus could produce an absurdly large level; such a spell now shows "-" as its
  level.
- Fixed a skill whose optional specialization differs from its library source being reported as matching that source.
  It is now flagged as modified and brought back into line by "Sync with Library Sources".
- Fixed stray commas appearing in the unmet prerequisites tooltip for a skill prerequisite that names a specialization
  or an optional specialization.
- Fixed a crash when displaying, editing, or copying a technique whose data file is missing its "Defaults To"
  information, as can happen with hand-edited files. Such a technique now loads with an empty default that can be filled
  in.
- Fixed the randomizers on a character sheet when the character's ancestry defines only one choice for a field. The
  value being replaced was always excluded from the choices, leaving nothing to pick, so the Gender field was cleared
  and the hair, eye color, skin, and handedness fields fell back to the built-in default of "Brown" (or "Right" for
  handedness) — a value the ancestry doesn't offer at all. The ancestry's lone choice is now kept in every case, and an
  ancestry with no options at all leaves the gender untouched.
- Fixed the DR tooltip for a hit location nested inside another location repeating its summary once for every level of
  nesting.
- Fixed data files with upper- or mixed-case file extensions (e.g. ".ANCESTRY") being silently skipped when scanning a
  library's Settings folder, which affected ancestries, calendars, attribute and body type settings, and name
  generators.
- Fixed a character's portrait being permanently discarded when the image couldn't be decoded. The data was cleared as
  soon as the sheet was displayed and lost for good at the next save; it is now kept intact, since a different build of
  GCS may be able to read it.
- Fixed two problems that caused a settings file to be abandoned in favor of factory defaults. A damaged or unreadable
  file was silently replaced and then overwritten at the next save, destroying it; the problem is now logged and the
  offending file is renamed with a ".bad" suffix so its contents can be recovered. A sheet settings file written in the
  older format that nests its content under a "sheet_settings" entry appeared to contain nothing; it now loads
  correctly.
- Fixed downloading a library update treating an HTTP error response — a rate limit message, an expired link, a server
  error — as if it were the library archive itself. The download now stops with a clear message naming the address and
  the status returned, instead of failing later with a confusing error.
- Fixed changing a library's location in the Library Settings triggering a flurry of redundant re-scans of the library
  navigator, and a possible crash when the new location couldn't be monitored for changes.
- Fixed several values reported by the scripting API. A huge nonsensical number was returned for the level or relative
  level of a skill or spell whose level cannot be computed; `skill.level`, `skill.relativeLevel`, `spell.level`,
  `spell.relativeLevel`, and `entity.skillLevel` now report 0 in that case. `spell.techLevel` now returns an empty
  string rather than `undefined` for a spell with no tech level, and `spell.difficulty` now returns the difficulty's key
  (e.g. "h") rather than its display name (e.g. "Hard"), matching what `skill.difficulty` has always returned — a script
  that relied on the old `spell.difficulty` value will need updating. Finally, the `id` and `parentID` properties of
  traits, trait modifiers, skills, spells, equipment, equipment modifiers, notes, and weapons were handed to the script
  engine as objects rather than plain text, so comparing one against a known ID (with `===`, `indexOf`, or `includes`)
  always failed, making the IDs unusable; they are now plain text.
- Fixed two ways one script could disrupt the scripts that run after it in the same interpreter. A timeout belonging to
  a script that had just finished could land on the next script, failing it with "script execution timed out" when it
  hadn't actually timed out. A script that replaced a built-in or left variables behind on the global object could
  silently change the behavior of every script run afterwards, including those of other open documents; the built-ins
  are now protected and anything a script leaves behind is cleared away before the interpreter is reused.
- Stopped the log from filling with script resolution errors while a script expression is being typed into an item
  editor. The editors re-evaluate the text on every keystroke to build their live previews, so the half-finished
  expression naturally fails until it is complete; those intermediate failures are no longer logged, while failures
  anywhere else still are.
- Fixed a crash in the command-line file converter (the "-c" option) when one of the supplied paths could not be read;
  it is now skipped and the remaining paths are still converted.
- Fixed keyboard and mouse handling while a modal dialog is open. Typing now always goes to the dialog, even when the
  operating system still considers the window behind it focused, clicks on the blocked window are ignored rather than
  acted upon, scrolling the window behind the dialog still works, and a mouse button that was held down when the dialog
  appeared no longer leaves GCS convinced the button is still down afterward.
- Fixed Tab and Shift-Tab moving focus into hidden fields, which made the focus appear to vanish.
- Fixed pop-up and context menus opened in windows that have no menu bar of their own — dialogs and the separate
  settings windows — not closing when clicking elsewhere, moving the window, or switching to another application, and
  keystrokes leaking through to the window behind them while they were open.
- Fixed three errors in where a row lands when moved by dragging: dragging a container row onto one of its own children
  moved the container out to the top level of the list instead of refusing the drop, dropping a row back at the position
  it already occupies nudged it up one row, and dragging equipment from one equipment list to another inserted it at a
  shifted position rather than where it was dropped.
- Fixed a row's disclosure triangle opening or closing when a click or drag that began somewhere else happened to end
  on top of it.
- Fixed a crash that could occur when the contents of a list changed while a row was being clicked or dragged.
- Fixed the insertion marker shown while dragging a document tab pointing at the wrong place when more tabs were open
  than fit across the window.
- Fixed scroll bar thumbs running past the end of their track and jumping under the pointer while being dragged, which
  was most noticeable in long documents where the thumb is at its smallest.
- Fixed the color swatches in the Colors settings accepting an image file dropped onto them, which replaced the color
  with a picture.
- Fixed the animated progress bar shown while updating libraries: its moving indicator was drawn slightly out of
  position, and each redraw of the window started an extra animation loop, making a long-running update progressively
  heavier on the processor.
- Fixed the Print dialog freezing while it queried the selected printer, which could last many seconds for a printer
  that was slow to answer or no longer on the network. The dialog now stays responsive and disables its options until
  the printer's capabilities arrive.
- Fixed custom key bindings that use two or more modifier keys (for example, Shift and Control together) being
  discarded when GCS was restarted, leaving the command with no shortcut at all.
- Fixed the height and vertical alignment of a line of text that mixes fonts of different sizes, such as inline code
  within a markdown paragraph. Such lines were measured using only the first font on the line and could be clipped;
  they now get the room they need and share a common baseline.
- Fixed two cases of markdown links not resolving. A page reference whose path is URL-encoded (e.g.
  `md:User%20Guide/Scripting%20Guide`) now resolves to the correct file, just like its non-encoded equivalent, and a
  heading that contains bold, italic, or code formatting now uses its full text, so the anchor automatically generated
  for it and the links pointing at it resolve correctly.
- Fixed an image in a markdown document that failed to load on the first attempt never being retried, leaving it
  missing for as long as the document was open.
- Fixed a number of errors in how SVG artwork is interpreted, including the order in which multiple transforms are
  applied, compact elliptical arc coordinates, shapes reused from elsewhere in the file, gradients referenced before
  they are defined, and shapes hidden by masks. Also fixed SVG artwork being drawn out of position when stretched to
  fill an area whose proportions don't match the artwork's own.
- Fixed pasting text copied from some other applications, where the copied data was offered under a more specific
  description than GCS was looking for and was therefore ignored.
- Fixed several slow memory leaks: closed windows, discarded images, and system objects created while drawing were
  being held onto for the life of the session. Long sessions now use noticeably less memory.
- Fixed GCS shutting down uncleanly when the system asks it to quit, such as when logging out or shutting down, or when
  Control-C is pressed in a terminal that launched it. This is now handled the same way as choosing Quit from the menu,
  so open files get their normal chance to be saved.
- Fixed `--convert` destroying the contents of theme color (`.colors`) and font (`.fonts`) settings files. Instead of
  bringing such a file up to the current data format, GCS wrote whatever colors and fonts it happened to be running with
  at the time, silently replacing the saved theme. These files now keep their own settings.
- Fixed a crash when using Reset in the Font Settings on a fresh installation, before any settings had been saved.
- Fixed the penalty applied when a spell's equipment prerequisite isn't met being applied to every spell on the sheet
  rather than just the spell that requires the missing equipment.
- Fixed a default to an attribute the character doesn't have being treated as usable when the "Use Half-Stat Defaults"
  option is on. Such a default now stays unresolvable, instead of producing a nonsense level that was stored as the
  skill's adjusted level and offered as a choice in Swap Defaults.
- Fixed a "Gives a trait maximum level modifier of" feature imposing a maximum on a trait that doesn't have one. A
  trait with an empty Maximum Level is unlimited, but a broadly-scoped bonus of +2 was computed from a base of zero and
  capped the trait at level 2, flagging it on the sheet — turning a feature meant to raise a limit into one that
  created it.
- Fixed the "retracting stock" weapon switch, which could be chosen in the feature editor but had no effect. It now
  adds or removes a weapon's retracting stock, along with the tooltip describing the folded-stock statistics.
- Fixed a script expression entered in one of a weapon's damage fields being silently cut short when it was written
  without spaces, which is how arithmetic is normally written. A base damage of `2*self.level` was taken as the dice
  specification it happens to start with — a flat +2 — and the script was never run. These fields are now treated as
  dice only when the entire entry is a dice specification; anything else is evaluated as a script.
- Fixed a weapon being given a brand new internal ID every time it was edited, which also happened again with each undo
  and redo. The change was written into the file at the next save, and it caused the weapon's row to lose its selection
  after an edit.
- Fixed the tooltip for a "per die" weapon bonus reporting the amount for a single die rather than the amount actually
  applied. A +1 per die accuracy bonus on a 3d weapon correctly raised its accuracy by 3, but the tooltip read "+1 (+1
  per die) to weapon accuracy". This affected every weapon statistic other than damage. Minimum ST and effective ST
  bonuses continue to count as a single die, since the weapon's damage dice depend on them.
- Fixed the minimum reach of a melee weapon that also has close combat being reset to 1 every time the file was saved
  and reloaded. A weapon with a reach of "C,2-3" came back as "C,1-3", and one with "C,5" as "C,1-5", because the
  close combat marker was read as the minimum. Note that "C-5" continues to mean a range running from close combat out
  to 5, and so is still equivalent to "C,1-5".
- Fixed the "New Equipment Modifiers Library" command not appearing in the Menu Keys settings, where it shared an
  internal identifier with "New Equipment Library". It now has its own entry and can be assigned a key binding.
- Fixed the reset button beside an individual command in the Menu Keys settings appearing to do nothing. The binding
  was restored to its default, but the key shown next to the command still displayed the old one until the settings
  were closed and reopened.
- Fixed typing in the notes field of a hit location that shares its ID with another, as the Right and Left Leg (and
  Arm) of the standard humanoid body do. The first keystroke sent the focus to the other location's field, so the rest
  of what you typed was recorded against the wrong location, and undoing the edit restored the text into the wrong
  location as well.
- Fixed edits to the second and later attributes added with the "Add Attribute" button in the Attribute Settings being
  applied to the first one added instead. Undoing a change to such an attribute wrote the restored text into the
  earlier addition's matching field, and the focus could afterward land on the wrong attribute's field.
- Fixed the Extended Value and Extended Weight shown in the equipment editor ignoring your unapplied edits whenever a
  modifier scales with the equipment itself. Changing the Level no longer left a "per level" cost or weight modifier
  stuck on the level the editor was opened with, and changing the weight (or turning a modifier on or off) no longer
  left a "per pound" cost modifier stuck on the original weight. These previews now match what the character sheet
  shows once you apply the changes, instead of jumping to a different number afterward.
- Windows only: Fixed mouse dragging that leaves the window. Dragging rows past the edge of a list to auto-scroll and
  releasing the mouse button outside the window now work correctly, instead of the button appearing to remain held
  down and every later mouse movement being treated as a drag.
- Windows only: Fixed the Shift and Windows keys occasionally remaining "stuck" as if still held down after switching
  away from GCS and back.
- Windows only: Fixed modal dialogs being centered incorrectly on monitors with a display scale other than 100%.
- Linux only: Fixed character entry on international keyboard layouts: characters that require the AltGr key (such as
  @, €, \ and | on many European layouts) now type correctly, and letters typed with Caps Lock engaged are now
  uppercase.
- Linux only: Fixed window and dialog placement on systems using display scaling (an Xft.dpi other than 96), which
  could put a window on the wrong monitor or well outside the intended area.
- Linux only: Fixed GCS spinning at full processor usage when its connection to the display server was lost, such as
  when a remote session is dropped; it now exits cleanly.
- Linux only: Fixed a tiny stray window briefly appearing on Wayland desktops whenever a file open or save dialog was
  used.
- macOS only: Fixed quitting from the Dock or logging out forcing GCS to exit even when the prompt about unsaved
  changes was canceled. Canceling now aborts the quit, just as it does when using the Quit menu item.
