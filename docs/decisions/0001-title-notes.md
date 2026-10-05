---
status: proposed
date: 2026-10-04
decision-makers: Lee Saferite
---

# Title notes: parenthetical notes after an item's name

## Context and Problem Statement

GURPS names a lot of things with a parenthetical after them: `Allies (Bob; 15 or less)`, `Phobia (Spiders)`,
`Guns/TL8 (Pistol)`. SJG's formatting guide calls these parenthetical notes. GCS only builds them from fields it knows
about: a skill's specialization and optional specialization, and a trait's self-control roll and frequency of
appearance. Anything else has to be typed into the name, where it can't be turned on by a modifier, filled in by a
nameable, or matched by a prerequisite without matching the whole name.

Modifiers make this worse. Libraries model most of what belongs in the parenthetical as modifiers: an Ally's point
total, frequency of appearance, reliability, intent. Today every enabled modifier is listed in the item's notes, under
its full name, and none of them can reach the title.

## Decision Drivers

* Library authors need to put text after an item's name, from the item itself or from one of its modifiers, so picking
  a modifier can change the name.
* Some modifiers belong in the title and most don't. Putting every modifier in the title gets out of hand quickly.
* Old files must load unchanged, render the same, and nothing about an item's calculated values may change.
* A title note says what the item is. It shouldn't change with play state such as an item's switch or whether it's
  equipped.
* Keep it inside the existing feature, modifier, prerequisite and editor machinery rather than adding a parallel system.

## Considered Options

* A `title_note` feature, plus a short name and a "show in title" flag on modifiers
* A dedicated title notes list field on traits, skills, spells and equipment
* Leave it to the name field and nameables

## Decision Outcome

Chosen option: a `title_note` feature, plus a short name and a "show in title" flag on modifiers. Features already
exist on every item and every modifier that needs one, already honor the modifier's enabled state, and already take
part in nameables, hashing and the editor. The modifier fields cover the common case where the modifier itself is what
belongs in the title, without asking authors to repeat its name in a feature.

### The feature

```json
{"type": "title_note", "text": "@Who@"}
```

* `text` is shown after nameable replacement and trimming. Empty or blank text adds nothing.
* It can be added to traits, skills, spells, equipment, trait modifiers and equipment modifiers, the same hosts the
  editor already offers features on.
* A title note on a modifier belongs to the item that owns the modifier. It applies while the modifier is enabled and
  not inside a disabled container, the same modifiers whose features the entity collects.
* A title note can't be switchable. Every other feature can be marked switchable, so that it only applies while its
  item's switch is on, but that switch is something a player flips during play, and the name shouldn't change with
  it. The editor doesn't offer the check box, the feature never gives its item a switch, and a `switchable` flag in
  the data is ignored.
* The entity ignores it during feature collection. It has no effect on points, levels or any other calculation.
* It describes itself in plain language ("Adds the title note Bob"), in the same form as the features editor's
  sentences for the other feature types.
* Nameable keys in `text` are extracted with the item's, so the nameables prompt asks for them. Replacements come from
  the owning item, for modifier title notes too.

### Modifiers

Trait and equipment modifiers gain three fields:

```json
{"name": "Appears quite often (12-)", "short_name": "12 or less", "show_in_title": true, "hide_notes": true}
```

* `short_name` is optional. When set, it names the modifier in its owner's title and notes in place of its name. It
  takes nameables like the name does. A leveled trait modifier still gets its level after it: `AD 2`.
* `show_in_title` puts the modifier in the title. Off by default, so every existing modifier stays where it is. A
  modifier in the title leaves its owner's notes, unless it has notes of its own to show; then its usual notes entry
  stays, so the notes aren't lost.
* `hide_notes` keeps the modifier's notes out of its owner's notes. A modifier not in the title is still listed there
  by name; one in the title disappears from the notes entirely. It doesn't affect weapon usage notes, which have their
  own flag, or the modifier's own row and prompts.
* All three fields are source data, so they are hashed and synced with the library. Containers don't have them; a modifier
  inside a container can still set them for itself.
* The full name is still used everywhere a modifier is listed as a modifier: its own row, editors and the modifier
  prompt.
* Anything that looks a modifier up by name accepts either name: the prerequisite modifier criteria below, a script's
  `findActiveModifier`, and the legacy text export's modifier lookups.

### Display

All parenthetical parts are joined with `"; "` inside one set of parentheses after the name, in this order:

1. Built-in parts, as today. For a skill: specialization, then optional specialization. For a trait (name column
   only): self-control roll, then frequency of appearance.
2. The item's own title notes, in feature order.
3. For each enabled modifier, in modifier order: its short name (or name) if it shows in the title, then its own title
   notes.

An Ally built as a trait with a `@Who@` title note first, and its point total and frequency modifiers set to show in the
title, reads `Allies (Bob; Built on 25%; 12 or less)`. Its other modifiers stay in the notes.

Title notes are part of each item's `String()`, so they show anywhere GCS names the item that way: list name columns,
prompts, bonus tooltips that name their source, legacy text exports, and the library navigator's search cache
(`Trait.StringWithSavedCalc` includes them too). Prerequisite and default matching against skill and trait names is
unaffected, since it reads the raw fields.

One existing output changes. A skill's specialization and optional specialization were joined with `", "`, as were a
trait's self-control roll and frequency in the name column. Both now use `"; "`, to match everything else in the group
and the formatting guide.

### Prerequisites

* Trait, skill and equipped equipment prerequisites gain `title_note`, a text criteria matched like tags: the qualifier
  may list several values separated by commas, a positive comparison passes if any title note matches any value, and a
  negative one passes only if none do.
* Spell prerequisites gain a `title_note` match type, "with a title note which".
* The criteria sees the title notes from features and from modifiers shown in the title. A skill's specialization is
  never treated as a title note, because it has its own criteria.
* Descriptions follow the existing wording: "Has trait Allies with the title note Bob", "with a title note that
  contains ...", "Knows at least 1 spell whose title note is Blue".
* Trait and equipped equipment prerequisites also gain `modifier`, a text criteria matched the same way against the
  names of the item's enabled modifiers. Each modifier offers both its name and its short name, without its level, so
  either one matches. Before this, the only way to test for a modifier was the trait prerequisite's notes criteria,
  which searched the notes text, and a modifier moved to the title or renamed by its short name drops out of that text.
* A modifier shown in the title passes both a modifier check (by either name) and a title note check (by the name it
  shows).

### Editor

* The feature row has the type popup on its first line and a full-width text field below, so the popup can't squeeze
  the text.
* The trait and equipment modifier editors get a Short Name field under the name, followed by a "Show in Title Notes"
  check box. Under the notes are "Show in Owner's Notes" (checked by default, the inverse of `hide_notes`) and "Show in
  Weapon Usage" (the existing weapon notes option, renamed to match).
* In the prerequisite editor, "+ title note" adds a chip worded like the tag chip ("and at least one title note is").
  Trait and equipped equipment prerequisites also get "+ modifier" ("and at least one modifier is").

### Compatibility

* No existing data changes meaning or rendering, apart from the separator noted above. The new fields are omitted when
  unset.
* Opened in an older GCS, a title note feature is kept as an unknown feature and preserved on save, in versions that
  have that (since August 2026). Earlier ones drop it. The new modifier fields are ignored and dropped on save.
* An older GCS ignores the new prerequisite fields and drops them on save. It also doesn't know the `title_note` spell
  match type and falls back to matching by name, so such a prerequisite means something else there.
* The new fields are part of the feature, modifier and prerequisite hashes. Source matching computes both sides'
  hashes with the running code and never stores them, so this doesn't flag existing library items as changed.

### Consequences

* Good, because any item or modifier can add to its name, and a library can move the modifiers that matter into the
  title one at a time, without changing anything else.
* Good, because prerequisites can target a specific variant (Allies with Bob) without matching on the full name, and
  can test for a modifier directly instead of searching the notes text.
* Bad, because skill names and the trait name column change their separator from a comma to a semicolon, which will
  show up in anything that compares rendered names, such as exported sheets.
* Bad, because "title note" sits next to the existing notes, and the two are easy to mix up in conversation. The UI and
  data always say "title note" in full to keep them apart.

### Confirmation

`model/gurps/title_note_test.go` covers display order and joining, nameables, enabled and disabled modifiers, modifiers
shown in the title with and without short names and levels, the notes area with and without hidden notes, the switch having no effect, the new
modifier fields' JSON and hashing, all four prerequisite types with all eight comparisons, the modifier criteria by
name and short name, and their descriptions. `ux/prereq_panel_test.go` covers the title note and modifier chips. The editors were also checked in a headless session.

## Pros and Cons of the Options

### A `title_note` feature, plus modifier fields

* Good, because modifiers get the feature for free, and the flag covers the common case of a modifier naming itself.
* Good, because enabled state, nameables, hashing, cloning and the editor already handle features and modifiers.
* Neutral, because display order is feature order, then modifier order, which the author controls.
* Bad, because a title note feature sits among bonuses in the features list, which isn't where an author would first
  look for part of a name.

### A dedicated title notes field

* Good, because it would sit next to the name in the editor.
* Bad, because modifiers would need the same field, and the enabled and nameable handling would be written a second
  time.
* Bad, because it adds a field to four item types and two modifier types for something most items never use.

### The name field and nameables

* Good, because it needs no code.
* Bad, because a modifier can't add to the name, and nothing can be matched apart from the name.

## More Information

### Naming

The guide's own term is "parenthetical note". "Title note" says where the text goes, and was chosen knowing it sits
close to the existing notes. Earlier drafts called it a "specifier". Other single words were ruled out because they're
taken: "Qualifier" is the field name on every criteria, and "Detail" is the template picker's row details.

### The formatting guide

The guide puts modifiers after all other notes, in alphabetical order, and gives some traits a fixed order (Ally:
description; point total; frequency; modifiers). This design keeps author order instead: features first, then
modifiers in their list order. An author following the guide orders the trait's title notes and modifiers to match.

### Out of scope

* Sorting modifiers in the title alphabetically, or rendering them in the guide's full `Name, Notes, ±X%` form.
* Moving or re-rendering the self-control roll and frequency of appearance (`CR12`, `FR12`) to match the guide.
* Replacing or removing title notes. If it's added, the plan is an optional "replaces" criteria on a title note,
  limited to the same item and its modifiers and applied after everything is collected.
* Title notes that reach other items. A trait adding a title note to a skill would make an item's name depend on the
  whole sheet's feature state.
* More than one title note check in a prerequisite, such as "has the title note Bob and a title note starting with
  15". A prerequisite holds one fixed criteria per field today, so this belongs with a broader change that lets a
  prerequisite hold a list of checks.
* An item identity built from the name and its title notes, for matching or merging rows. An earlier draft had an
  `identity` flag on each note, and it read as confusing. If identity comes back, a `display_only` flag that defaults
  to off would keep every existing title note part of the identity without touching any data.
* Title notes in the saved `calc` data, for tools that read a sheet without evaluating it.
