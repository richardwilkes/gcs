# Equipment Modifiers

**Equipment Modifiers** ([B345](B345)) can be applied to [Equipment](Equipment) as either **enhancements** or
**limitations**. These adjust the base cost and weight of items and allow you to add [Features](Features) and attacks
to equipment.

## Master Library

GCS includes libraries of official equipment modifiers in the **Master Library** that are taken from GURPS books and
other supplements. See [Library Explorer](Library%20Explorer) for more details.

You can also create your own custom equipment modifiers.

## Equipment Modifier list

The **Equipment Modifier** table in the Equipment Editor includes the following columns. Select a column heading to
sort ascending ![sort ascending](images/icons/icn-sortAscending.svg) or descending ![sort
descending](images/icons/icn-sortDescending.svg).

| Column                                                    | Description                                                                                                                                                       |
|-----------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **![enabled](images/icons/icn-applyChanges.svg)**         | If checked, the modifier is **enabled** and its effects apply to the item.                                                                                        |
| **Equipment Modifier**                                    | The modifier’s name and any notes. Select the chevron ![notes expand](images/icons/icn-notesExpand.svg) to show or hide the notes.                                |
| **TL**                                                    | The Tech Level.                                                                                                                                                   |
| **Cost Adjustment**                                       | The calculated adjustment to the equipment’s monetary cost.                                                                                                       |
| **Weight Adjustment**                                     | The calculated adjustment to the equipment's weight.                                                                                                              |
| **Tags**                                                  | Tags used to help organize modifiers.                                                                                                                             |
| **![page reference](images/icons/icn-pageReference.svg)** | Page reference for the modifier. See [Page References](Page%20References).                                                                                        |
| **![source](images/icons/icn-source.svg)**                | Tracks whether the modifier matches the library it came from. See _Master Library Syncing_ in [Character Sheet Overview](Character%20Sheet%20Overview).           |

## Adding equipment modifiers

You must add modifiers to existing [Equipment](Equipment).

To add a modifier from a library:

1. Open the **Equipment Editor** for the equipment you want to modify.
2. Open an equipment modifiers list from the **Master Library**.
3. Drag the modifier to the **Equipment Modifier** table in the editor.

To add modifiers without dragging:

1. Select the modifiers in an equipment modifiers list.
2. Go to **Item > Apply Modifier**, or right-click and select **Apply Modifier**.
3. Choose the open sheet, template, loot sheet, or list, then the equipment that should receive them. Each item
   receives its own copy of the modifiers, enabled.

You can also drop modifiers straight onto equipment in a list, or onto an item in the **Buy Equipment** pane before
buying it.

To create a new custom modifier:

- Open the Equipment Editor for the equipment you want to modify.
- Go to **Insert > Equipment > New Equipment Modifier** (`Ctrl+Alt+F` / `Cmd+Option+F`), or right-click the Equipment
  Modifier table and select **New Equipment Modifier**.

Modifier **groups** and **choices** work the same way as for [Trait Modifiers](Trait%20Modifiers): a choice tagged
**Pick 1** or **Pick at most 1** makes its modifiers exclusive. Add them with **Insert > Equipment > New Equipment
Modifier Group** and **New Equipment Modifier Choice**. The prompt that asks about modifiers when you add equipment to a
sheet works the same way too, including **Auto Pick** and **Auto Pick All** for choices that must be made.

## Managing modifiers

You can manage modifiers directly in the Equipment Editor:

- **Enable/disable:** Check or uncheck the box to apply or remove a modifier’s effects. You can also select modifiers
  and use **Item > Toggle State** to flip several at once.
- **Reorder:** Drag modifiers, or use the **Item > Move** commands.
- **Delete:** Right-click the modifier and select **Delete** (or select it and go to **Edit > Delete**).

## Editing equipment modifiers

You can edit an equipment modifier in a few ways:

- Double-click it to open the **Equipment Modifier Editor**.
- Right-click it and select **Open Detail Editor**.
- Select an equipment modifier and go to **Item > Open Detail Editor**.

Save your changes when you're done from the **Equipment Modifier Editor toolbar** (see below).

## Equipment Modifier Editor fields

The Equipment Modifier Editor includes the following fields:

| Equipment Modifier field          | Description                                                                                                                                            |
|-----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Name**                          | The equipment modifier's name.                                                                                                                         |
| **Show in title notes**           | If checked, the modifier is listed in parentheses after the item's name, instead of in the item's notes.                                               |
| **Tech Level**                    | The Tech Level.                                                                                                                                        |
| **Notes**                         | Notes about what the equipment modifier does. These appear with the equipment on your character sheet. Supports [Markdown](Markdown%20Guide).          |
| **Show notes in owner's notes**   | If checked (the default), the modifier's notes appear in the item's notes. Clear it to list the modifier by name only.                                 |
| **Show notes in attack usage**    | If checked, the notes also appear in the usage notes of the item's attacks.                                                                            |
| **VTT Notes**                     | Notes for use in your Virtual Tabletop system. See your VTT’s documentation.                                                                           |
| **Enabled**                       | If checked, the modifier is active and applies its effects.                                                                                            |
| **Cost Modifier**                 | Adjusts the item’s monetary cost. See **Cost and Weight Modifiers** section below.                                                                     |
| **Weight Modifier**               | Adjusts the item’s weight. See **Cost and Weight Modifiers** section below.                                                                            |
| **Per Level**                     | If checked, applies the cost/weight change for each level of the equipment.                                                                            |
| **Per Pound**                     | If checked, applies the cost change for each pound the equipment weighs.                                                                               |
| **Tags**                          | Comma-separated [Tags](Tags) to help organize modifiers.                                                                                               |
| **Page Reference**                | Page number for the modifier. See [Page References](Page%20References) for details.                                                                    |
| **Page Highlight**                | Text to highlight in the PDF. Leave blank unless the default highlights fail.                                                                          |
| **Nameables**                     | Defines the `@code@` placeholders the modifier uses, and shows which are defined by the item it belongs to. See [Substitutions](Substitutions).        |
| **Features**                      | Lists attribute, condition, skill, spell, reaction, or attack changes applied by the modifier. See [Features](Features) for details.                   |
| **Melee Attack Usage**            | Melee attacks the modifier gives the item while enabled, shown under the item's name. See [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons).  |
| **Ranged Attack Usage**           | Ranged attacks the modifier gives the item while enabled, shown under the item's name. See [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons). |

## Cost and Weight Modifiers

Unlike [Trait Modifiers](Trait%20Modifiers), equipment modifiers adjust an item’s **money cost** and **weight** instead
of character points.

Each modifier has two parts:

1. A **value** (for example, `+5`, `-10%`, `x2`, `+1 CF`).
2. A **stage** (selected from the dropdown).

### Accepted formats

Both **Cost** and **Weight** modifiers accept multiple formats. You can enter **whole numbers, decimals, or fractions**:

- **Flat changes**: `+5`, `-5 lb`, `-0.2`
- **Percentages**: `+10%`, `-20%`
- **Multipliers**: `x2`, `x3.2`, `x2/3`
- **Cost Factor (CF)**: `+2 CF`, `-0.2 CF` (cost only, at the base cost stage)

Cost Factor (CF) represents **multiplicative adjustments** to an item’s base price.

- A **positive CF** increases cost by adding multiples of the base price.
- A **negative CF** decreases cost by subtracting multiples of the base price.

For example, a $500 Broadsword with **+2 CF** (Fine quality) costs: `$500 base + ($500 x 2 CF) = $1,500`.

### Cost Modifier stages

- **To original cost:** Adjusts the raw book cost.
- **To base cost:** Adjusts after TL/CF changes. Only multipliers and CF values apply here.
- **To final base cost:** Adjusts after other modifiers, but before the final calculation.
- **To final cost:** Adjusts the very end result.

### Weight Modifier stages

- **To original weight:** Adjusts the listed book weight.
- **To base weight:** Adjusts after TL/CF changes.
- **To final base weight:** Adjusts after other modifiers, but before the final calculation.
- **To final weight:** Adjusts the very end result.

> [!TIP]
> Multiple cost or weight modifiers can stack. Be careful which stage you apply them at — this changes the math
> significantly. Turn on **Show equipment modifier cost & weight adjustments** in [Sheet Settings](Sheet%20Settings) to
> see each modifier's adjustment in the item's notes.

## Equipment Modifier Editor toolbar

The toolbar above the Equipment Modifier Editor gives you quick access to common actions:

| Button                                                  | Name                    | Description                                                                                             |
|---------------------------------------------------------|-------------------------|---------------------------------------------------------------------------------------------------------|
| ![help](images/icons/icn-help.svg)                      | **Help**                | Opens this page from the User Guide.                                                                    |
| 100%                                                    | **Scale**               | Adjusts the display size of the Equipment Modifier Editor.                                              |
| ![apply changes](images/icons/icn-applyChanges.svg)     | **Apply Changes**       | Saves and closes the Equipment Modifier Editor (`Ctrl+Return` / `Cmd+Return`).                          |
| ![discard changes](images/icons/icn-discardChanges.svg) | **Discard Changes**     | Exits without saving changes (`Escape`).                                                                |
| ![source](images/icons/icn-source.svg)                  | **ID & Library Source** | Shows the modifier's ID and library source, and offers **Sync with Source** and **Clear Source**.       |

---

[Return to Home](Home)
