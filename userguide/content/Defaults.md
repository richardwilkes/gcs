# Defaults

**Defaults** appear in the **Detail Editor** of [Skills](Skills) and in the editors of the melee and ranged attacks
that belong to equipment, traits, skills, and spells (see [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons)).

**Defaults** represent **Skill Defaults** ([B173](B173)), not GCS [Default Settings](Default%20Settings). They describe
what a skill or an attack falls back on if you don’t spend points on it directly.

For example:

- A **skill** might default from an **attribute** (for example, _DX-5_).
- An **attack** might default from a **skill** (for example, _Broadsword_).
- A **block or parry** could default from your calculated defense score.

> [!NOTE]
> You can add **multiple defaults**. GCS uses the one that yields the best result given Tech Level and other
> conditions. A skill with points in it keeps using the default it has been using until its list of defaults changes.

## The Defaults section

The **Defaults** section lists each default as a line of plain text, such as _DX at -5_ or _Skill Broadsword at -2_.
Select the section's title to collapse it to a one-line summary, or to expand it again.

- **Editing:** Select a default to open it. Its controls read as a sentence. Select **Done** (or press `Return` or
  `Escape`) to close it.
- **Reordering:** Drag a default by its grip ![grip](images/icons/icn-grip.svg) to move it. Dragging it into the
  Defaults section of another open editor, whether a skill's or an attack's, copies it there, expanding the section if
  it's collapsed. In a read-only editor, defaults can be dragged out but not rearranged.
- **More actions:** Select **More actions** ![more](images/icons/icn-moreActions.svg) beside a default for
  **Duplicate**, **Move Up**, **Move Down**, and **Delete**.

Every change can be undone with **Edit > Undo**.

## Adding defaults

To add a default:

1. Open the skill's or attack's **Detail Editor**.
2. In the **Defaults** section, select **Add** ![add](images/icons/icn-add.svg). When the section is empty, you can
   also select **No defaults. Click here to add one.**
3. Choose a **Source** from the first dropdown (**10**, an attribute, **Parry**, **Block**, or **Skill**).
4. Enter the **Modifier** (e.g., `-5`, `+0`) and any other fields that appear.
5. (Optional) Add a **+ tech level** condition.

To delete a default:

- Select **More actions** ![more](images/icons/icn-moreActions.svg) beside it, then **Delete**.

## How a default is built

Each default entry in GCS reads like a sentence composed of:

- a **Source** (what you’re defaulting **from**),
- a **Modifier** (numeric adjustment), and
- an optional **Tech Level rule** (when it applies).

| Default Field               | Description                                                                                                                      |
|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| **Source**                  | What this default is based on: **10**, an attribute such as **ST** or **DX**, **Parry**, **Block**, or **Skill**.                |
| **Name**                    | Appears when the source is **Skill**, **Parry**, or **Block**. Choose how to match the skill name, then enter it.                |
| **+ specialization**        | Optional, for skill-based sources. Use when a specific specialization is required.                                               |
| **+ tag**                   | Optional, for skill-based sources. Default from whichever skill the character has that carries the tag.                          |
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

> [!WARNING]
> Skill names must match exactly as they appear in the library or on your sheet. Typos or mismatches mean the default
> won’t work. Comparisons ignore case.

## Tech Level conditions

Each default can include a **+ tech level** condition. This is used if the default only applies for a specific Tech
Level. Use **[Operators](Operators)** (_is_, _is not_, _is at least_, _is at most_) to control when the default applies.
The condition is checked against the Tech Level of the skill being defaulted from.

## Swapping defaults

On a character sheet, select a skill and choose **Item > Swap Defaults** (`Ctrl+Shift+X` / `Cmd+Shift+X`) to change
which default it uses:

- A skill with several usable defaults cycles to the next one.
- Otherwise, the skill swaps with the skill it defaults from, so the points move to the other skill and this one is
  left at its default level.

Techniques, skills with no points, and granted skills can't be swapped.

## Saving defaults

On the **Detail Editor** toolbar:

- **Apply changes:** Select **Apply Changes** ![apply changes](images/icons/icn-applyChanges.svg) to save.
- **Discard changes:** Select **Discard Changes** ![discard changes](images/icons/icn-discardChanges.svg) to exit
  without saving.

---

[Return to Home](Home)
