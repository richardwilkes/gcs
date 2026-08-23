# Changes since the last release

## New & Improved

- The PDF viewer now uses the page labels embedded in a PDF, so pages are identified by the numbering printed in the
  book itself rather than by counting pages in the file. The page field in the toolbar shows and accepts these labels,
  including non-numeric ones such as `iv` for the fourth page of frontmatter. If you had set a page offset for a PDF to
  compensate for covers or frontmatter, you may need to reset it to 0, since the page labels already account for those
  pages.
- Page references may now use non-numeric page labels. When one is used, a colon **must** separate the reference key
  from the label: to reference page `iv` of the Basic Set, use `B:iv`. References that already include a colon, such as
  the Pyramid references (e.g. `PY104:12`), work as they always have — do not add a second colon.
- Page references of `B338` and higher refer to material in Basic Set: Campaigns, the second of the two volumes the
  Basic Set was published as, and are now automatically redirected to the `BX` reference key. Previously, these
  references only opened for those who had combined the two volumes into a single PDF mapped to the `B` key; now they
  work with the volumes kept as separate PDFs. If you use a single PDF instead — whether a combined PDF of the two
  volumes or the new single-volume Basic Set Revised, which keeps the same page numbering for the same content — map
  the `BX` key to it as well as the `B` key. (#1098)
- Features can now be marked as "switchable". A switchable feature only takes effect while the switch of the trait,
  skill, spell, or equipment it belongs to is on. The switch is per item (a modifier's switchable features follow the
  switch of the item the modifier belongs to) and can be toggled from a new column that appears on the character sheet
  lists when any item in the list has switchable features, or from the item's editor. A trait or equipment container can
  have a switch of its own; hold down the Option/Alt key when clicking that switch to also apply the change to
  everything contained within it. Skill and spell containers never have a switch, since a container's own features
  don't apply.
- Spells can now have features, just like traits, skills and equipment. As a side effect, weapons attached to a spell
  now receive any "this weapon" bonuses defined by the spell's own features.
- Adding an item to a template now prompts for its modifier selections and name substitutions, just as adding it to a
  character sheet does. (#1101)
- Items in a template can now be marked "Preconfigured" with a new checkbox in their editors. When a preconfigured item
  is added to a character, the prompts for modifier selections and name substitutions are skipped, since you already
  answered them while building the template. Trait and equipment containers can be preconfigured as well; the flag is
  cleared automatically when an item is copied anywhere other than a template. (#1102)
- Attribute pools can now be hidden, just like primary and secondary attributes. (#1103)
- The PDF viewer displays pages faster. The graphics libraries GCS is built on now use the vector (SIMD) instructions
  of modern processors for work done on the CPU, and PDF pages are always rendered there, so page display — including
  decoding the images inside a PDF — speeds up for everyone. Normal interface drawing happens on the GPU and is
  unaffected, but when GCS has to fall back to CPU rendering because working OpenGL support isn't available, that path
  is now much faster as well: fills and gradients render 40–60% faster, blurs up to 27x faster, and the screen update
  step on Windows and Linux has been sped up too. On older Intel/AMD processors that lack AVX2 support, GCS
  automatically falls back to the previous code; the rendered output is identical either way.

## Bug Fixes

- The `equipped` property of equipment in scripts now reports `false` for items in the Other Equipment list, since
  those items don't affect the character no matter what their equipped flag says. The Extended Value and Extended
  Weight previews in the equipment editor answer that property the same way the sheet does, including for a top-level
  row of the Other Equipment list.
- A "this armor" DR bonus (one that names no hit locations) now covers the locations that the armor's enabled modifiers
  grant DR to, along with those the armor itself grants DR to, whether the "this armor" bonus comes from the armor or
  from one of its modifiers. This can change the DR totals of existing characters whose armor picks up locations from
  its modifiers.
- The per-hit-location equipment lists produced by the legacy text export (the `EQUIPMENT` and `EQUIPMENT_FORMATTED`
  keys of the hit location loop) now include armor whose DR for that location comes from one of its modifiers, which
  the DR reported for the location already accounted for. Each piece of armor is also listed just once, rather than once
  per DR bonus of its own that reaches the location.
- Right-clicking one of the checkmark columns of the sheet's lists (Equipped, a modifier's enabled state, a weapon's
  hidden state) now selects the row and brings up its context menu, just as right-clicking anywhere else in the row
  does, rather than toggling the checkmark.
- Equipping or unequipping an item, whether by clicking its Equipped checkmark or with the Toggle Equipped command, now
  brings the Reactions, Conditional Modifiers and weapon lists onto or off of the sheet as the item's contributions
  come and go, rather than leaving them as they were until something else caused the sheet to be rebuilt.
- A template's equipment list now shows or hides its TL and LC columns as soon as the corresponding sheet settings are
  changed, rather than only after the template is reopened.
- A character sheet's modification timestamp is now updated by every edit made through its lists, not just some of
  them. Applying changes from an item's editor, deleting or duplicating rows, moving equipment between the carried and
  other equipment lists, converting rows to or from containers, answering a modifier or nameables prompt, swapping
  skill defaults, syncing with library sources, applying a template, and editing the point records all left the
  timestamp unchanged before. Several edits also redrew the sheet two or three times over, which on a sheet with many
  rows made them noticeably slower than they needed to be.
- Closing an item's editor no longer scrolls the sheet when the row that was edited is still in view. Before, the
  focus was returned to the list in a way that scrolled the entire list into view, which on a sheet with a long list
  could move the row far from where it had been. Now only the edited row is brought into view, and only when it isn't
  already visible.
- Modifiers now keep their connection to the library item they came from when the trait or equipment item holding them
  is copied to a character sheet or template, so syncing with library sources updates them as well. (#1100)
- Opening a trait, skill, spell, or equipment editor and clicking Apply no longer silently replaces the hidden
  identifiers of every modifier and weapon within the item. Those identifiers are how GCS matches items to their
  library sources, so an edit that changed nothing could still quietly break later syncs. Duplicating an item of your
  own creation also no longer marks the copy as if it had come from a library. (#1106)
- Fixed a memory error in the user interface library that could destabilize GCS on Windows when it created cursors or
  the images shown while dragging.
