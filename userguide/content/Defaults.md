# Defaults

**Defaults** appear in the **Detail Editor** of [Skills](Skills) and [Spells](Spells), and in the editors of the melee
and ranged attacks that belong to equipment, traits, skills, and spells (see
[Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons)).

**Defaults** represent **Skill Defaults** ([B173](B173)), not GCS [Default Settings](Default%20Settings). They describe
what a skill, spell, or attack falls back on if you don’t spend points on it directly.

For example:

- A **skill** might default from an **attribute** (for example, _DX-5_).
- An **attack** might default from a **skill** (for example, _Broadsword_).
- A spell's **attack** might default from the **spell** itself.
- A **block or parry** could default from your calculated defense score.

> [!NOTE]
> You can add **multiple defaults**. GCS uses the one that yields the best result given Tech Level and other
> conditions. A skill or spell with points in it, other than a Ritual Magic spell, keeps using the default it has been
> using until its list of defaults changes or you [swap](#swapping-defaults) it.

## The Defaults section

The **Defaults** section lists each default as a line of plain text, such as _DX at -5_ or _Skill Broadsword at -2_.
Select the section's title to collapse it to a one-line summary, or to expand it again.

- **Editing:** Select a default to open it. Its controls read as a sentence. Select **Done** (or press `Return` or
  `Escape`) to close it.
- **Reordering:** Drag a default by its grip ![grip](images/icons/grip.svg) to move it. Dragging it into the
  Defaults section of another open editor, whether a skill's, a spell's, or an attack's, copies it there, expanding
  the section if it's collapsed. In a read-only editor, defaults can be dragged out but not rearranged.
- **More actions:** Select **More actions** ![more](images/icons/moreActions.svg) beside a default for
  **Duplicate**, **Move Up**, **Move Down**, and **Delete**.

Every change can be undone with **Edit > Undo**.

## Adding defaults

To add a default:

1. Open the skill's, spell's, or attack's **Detail Editor**.
2. In the **Defaults** section, select **Add** ![add](images/icons/add.svg). When the section is empty, you can
   also select **No defaults. Click here to add one.**
3. Choose a **Source** from the first dropdown (**10**, an attribute, **Parry**, **Block**, **Skill**, or **Spell**).
4. Enter the **Modifier** (e.g., `-5`, `+0`) and any other fields that appear.
5. (Optional) Add a **+ tech level** condition.

To delete a default:

- Select **More actions** ![more](images/icons/moreActions.svg) beside it, then **Delete**.

## How a default is built

Each default entry in GCS reads like a sentence composed of:

- a **Source** (what you’re defaulting **from**),
- a **Modifier** (numeric adjustment), and
- an optional **Tech Level rule** (when it applies).

| Default Field               | Description                                                                                                                      |
|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| **Source**                  | What this default is based on: **10**, an attribute such as **ST** or **DX**, **Parry**, **Block**, **Skill**, or **Spell**.     |
| **Name**                    | Appears when the source is **Skill**, **Spell**, **Parry**, or **Block**. Choose how to match the name, then enter it.           |
| **+ specialization**        | Optional, for the **Skill**, **Parry**, and **Block** sources. Use when a specific specialization is required.                   |
| **+ power source**          | Optional, for the **Spell** source. Default only from a spell of a matching power source ([B77](B77)).                           |
| **+ tag**                   | Optional, wherever **Name** appears. Default from whichever skill or spell the character has that carries the tag.               |
| **Modifier**                | Numeric adjustment added to the source (for example, `-5`, `+0`).                                                                |
| **+ tech level**            | A condition like “**when the tech level is at least 7**”, using [Operators](Operators). The default applies only when it is met. |

## Sources

Sources define what the default is based on. The fields you see depend on the source you select.

### 10

Defaults from a fixed base of 10 with a modifier, for skills that don't depend on an attribute.

### Attribute

Defaults from a character **attribute** (e.g., **ST**, **DX**, **IQ**, **Will**, **Vision**, etc.) with a modifier. For
example, `DX at -5`.

### Parry / Block

Defaults from the **Parry** or **Block** score of a skill with a modifier ([B374](B374)). Choose the skill the same way
as for a **Skill** source, or leave the name as _is anything_ to use any skill.

### Skill

Defaults from another **skill** (optionally with a **specialization** or **tag**) plus a modifier. For example, `Skill
Bow at +0` or `Skill Driving whose specialization is Motorcycle at -4`.

Matching by tag lets one default cover a whole family of skills: Stage Combat can default to whichever combat skill the
character has at -3 instead of listing each one.

A default from a skill uses that skill's level after the sheet's skill level caps (see _Skill Level Caps_ in
[Sheet Settings](Sheet%20Settings)). The defaulting skill is only capped itself if a cap selects it.

### Spell

Defaults from a **spell** (optionally with a **power source** or **tag**) plus a modifier. For example,
`Spell Fireball at -2` or `Spell Fireball whose power source is Arcane at -2`. A spell has no specialization, and
**Parry** and **Block** always draw on a skill, never a spell.

**+ power source** compares the spell's power source ([B77](B77)) with the usual text comparisons (see
[Operators](Operators)). In the defaults of a spell, or of an attack that belongs to one, it also offers
_is the same as this spell's_, which a new **Spell** default there starts with, as in
`Spell Fireball whose power source is the same as this spell's at -2`. This lets a spell's attack default from the
spell itself. Dragged into the defaults of a skill, or of an attack that doesn't belong to a spell, a default with that
choice matches no spell.

When several spells match, the one with the highest level is used. As with a skill, that's the spell's level after the
sheet's skill level caps. A skill or spell never defaults from a spell with no points in it, other than a Ritual Magic
spell.

> [!WARNING]
> Skill and spell names must match exactly as they appear in the library or on your sheet. Typos or mismatches mean the
> default won’t work. Comparisons ignore case.

## Tech Level conditions

Each default can include a **+ tech level** condition. This is used if the default only applies for a specific Tech
Level. Use **[Operators](Operators)** (_is_, _is not_, _is at least_, _is at most_) to control when the default applies.
The condition is checked against the Tech Level of the skill or spell being defaulted from.

## Defaults of skills and spells

The best default that applies gives a skill or spell a level even with no points in it, and the points you spend on it
add to what that default is worth ([B173](B173)). A Wildcard skill or spell ignores its defaults.

A Ritual Magic spell (see [Spells](Spells)) has the higher of the level its **Base Skill** gives it and the level of its
best default. Its points raise only the level from the Base Skill, so they don't add to the default, and the default
isn't held to the Base Skill's level. It always uses its best default.

On the sheet, a skill or spell that defaults from another skill or spell names it in a note, such as
_Default: Thaumatology-2_, wherever **Modifiers** are shown (see **Where to display** in
[Sheet Settings](Sheet%20Settings)). A Ritual Magic spell shows the note only while its default gives it a higher level
than its Base Skill does.

## Swapping defaults

On a character sheet, select a skill or spell and choose **Item > Swap Defaults** (`Ctrl+Shift+X` / `Cmd+Shift+X`), or
right-click it and select **Swap Defaults**, to change which default it uses:

- A skill or spell with several usable defaults cycles to the next one.
- Otherwise, it trades places with the skill or spell it defaults from, or with one that defaults from it, if each has a
  default to the other. Afterward, the one that defaulted from the other no longer does, and the other defaults from it
  instead. A skill and a spell can trade places this way. Since a Ritual Magic spell always uses its best default, it
  trades places only when the result would last: once it gives up its default, it must still have a level, and once it
  defaults from the other, that must be its best default and give it a higher level than its **Base Skill** does.

The command acts on whichever of the Skills and Spells lists has the keyboard focus. When neither does, it acts on the
Skills list if anything is selected there, and otherwise on the Spells list. Techniques, Ritual Magic spells, items
with no points, and granted items can't be swapped.

## Saving defaults

On the **Detail Editor** toolbar:

- **Apply changes:** Select **Apply Changes** ![apply changes](images/icons/applyChanges.svg) to save.
- **Discard changes:** Select **Discard Changes** ![discard changes](images/icons/discardChanges.svg) to exit
  without saving.

---

[Return to Home](Home)
