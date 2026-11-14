# Melee and Ranged Weapons

The **Melee Attacks list** and **Ranged Attacks list** on your [character sheet](Character%20Sheet%20Overview) show all
of your character’s attacks ([B268–B276](B268)), their usage statistics, and related notes. GCS calls each entry an
**attack**, since one weapon can be used several ways.

These lists include:

- **Natural attacks** (bite, kick, punch).
- **Weapons** from [Equipment](Equipment).
- **Innate attacks** from [Traits](Traits), **techniques** from [Skills](Skills), and **melee and missile spells** from
  [Spells](Spells).
- **Modifier attacks** from [Trait Modifiers](Trait%20Modifiers) and [Equipment Modifiers](Equipment%20Modifiers).

Attacks belong to the item that provides them. An item with a **Melee Attack Usage** or **Ranged Attack Usage**
automatically appears in the appropriate list. To change or delete an attack, edit the item it belongs to.

A modifier's attacks belong to the trait or equipment it modifies, and appear only while the modifier is enabled. They
are listed under that item's name and work like its own attacks: they use the item's level for leveled damage (unless
the trait modifier's **Attacks use this modifier's level** is checked) and, for equipment, its **Rated ST**; the
features of the item and its enabled modifiers that target its attacks apply to them; and modifiers marked **Show notes
in attack usage** add their notes. Their `@code@` placeholders are the modifier's (see _Codes shared with modifiers_ in
[Substitutions](Substitutions)). To change or delete one, edit the modifier.

> [!NOTE]
> Normally, only attacks from **carried** _and_ **equipped** items appear here. If you enable **Show all attacks** in
> [Sheet Settings](Sheet%20Settings), GCS will also display attacks from items that are _not_ carried or equipped.
> These are clearly marked as `(not carried)` or `(not equipped)`, as appropriate.

## Master Library

GCS includes hundreds of official weapons in the **Master Library**. See [Library Explorer](Library%20Explorer) for more
details.

You can also create your own custom weapons.

## Melee Attacks list

The **Melee Attacks list** on your character sheet includes the following columns. Select a column heading to sort
ascending ![sort ascending](images/icons/icn-sortAscending.svg) or descending ![sort
descending](images/icons/icn-sortDescending.svg).

| Column           | Description                                                                                |
|------------------|--------------------------------------------------------------------------------------------|
| **Melee Attack** | The name of the item the attack belongs to.                                                |
| **Usage**        | Name for the entry. A weapon can have multiple usages.                                     |
| **SL**           | The calculated skill level with the attack. Hover over it to see how it was worked out.    |
| **Parry**        | The Parry score with this weapon ([B269](B269)), or **No** if it can't parry.              |
| **Block**        | The Block score with this weapon ([B327](B327)), or **No** if it can't block.              |
| **Damage**       | Damage done by the attack ([B268](B268)), with any bonuses applied.                        |
| **Reach**        | The weapon’s reach in combat ([B269](B269)).                                               |
| **ST**           | Minimum Strength needed to use the weapon effectively ([B270](B270)).                      |

## Ranged Attacks list

The **Ranged Attacks list** on your character sheet includes the following columns. Select a column heading to sort
ascending ![sort ascending](images/icons/icn-sortAscending.svg) or descending ![sort
descending](images/icons/icn-sortDescending.svg).

| Column            | Description                                                            |
|-------------------|------------------------------------------------------------------------|
| **Ranged Attack** | The name of the item the attack belongs to.                            |
| **Usage**         | Name for the entry. A weapon can have multiple usages.                 |
| **SL**            | The calculated skill level with the attack.                            |
| **Acc**           | Accuracy bonus from aiming ([B269](B269)).                             |
| **Damage**        | Damage done by the attack ([B268](B268)).                              |
| **Range**         | Half-damage and maximum range ([B269](B269)).                          |
| **RoF**           | Rate of Fire (how many shots per attack) ([B270](B270)).               |
| **Shots**         | The number of shots a weapon can fire before reloading ([B270](B270)). |
| **Bulk**          | The weapon’s size and handiness ([B270](B270)).                        |
| **Recoil**        | How easy the weapon is to fire ([B271](B271)).                         |
| **ST**            | Minimum Strength needed to use the weapon effectively ([B270](B270)).  |

> [!TIP]
> Turn on **Hide unused columns in the melee & ranged attack tables** in [Sheet Settings](Sheet%20Settings) to leave
> out columns that are empty for every attack, such as Recoil for a character with no firearms.

## Adding attacks

Attacks are not created directly in the Melee or Ranged Attacks lists. Instead:

- Add a weapon the same way you add [Equipment](Equipment), [Traits](Traits), [Skills](Skills), or [Spells](Spells).
- If the item, or one of its enabled modifiers, has **Melee Attack Usage** or **Ranged Attack Usage** entries in its
  editor, they automatically appear in the appropriate list.

Some weapons can have **multiple usages**. For example, a spear may have a **thrust impaling** attack and a **swing
cutting** attack, which appear as separate rows.

To add natural attacks (bite, kick, and punch):

- Go to **Sheet > Add Natural Attacks**. This adds a _Natural Attacks_ trait with all three. New sheets get it
  automatically unless you turn off **Add natural attacks** in [General Settings](General%20Settings).

## Managing and editing attacks

You manage and edit attacks through the item they belong to.

To edit an attack’s details:

1. Open the editor of the item it belongs to (double-click the item in its list on the sheet). For a modifier's
   attack, open the modifier's editor from there as well.
2. In the **Melee Attack Usage** or **Ranged Attack Usage** table, double-click the attack to open the **Melee Attack
   Editor** or **Ranged Attack Editor**.
3. Make your changes and select **Apply Changes** ![apply changes](images/icons/icn-applyChanges.svg) in the attack
   editor, then again in the modifier's editor, if any, and the item's editor.

Double-clicking an attack in the Melee Attacks or Ranged Attacks list on the sheet does steps 1 and 2 for you.

To add a new attack to an item or modifier, open its editor and go to **Insert > New Melee Attack** (`Ctrl+Shift+M` /
`Cmd+Shift+M`) or **Insert > New Ranged Attack** (`Ctrl+Shift+R` / `Cmd+Shift+R`), or right-click the attack table and
choose the same command. To hide an attack from the sheet without deleting it, check **Hide** in the table or select
the attack and use **Item > Toggle State**.

> [!NOTE]
> An attack must have at least one **default skill** (e.g., Broadsword, Spear) to calculate its skill level. See
> [Defaults](Defaults). Hover over the attack's **SL** on the sheet to see which skill it used and what adjusted it.

## Melee Attack Editor fields

The Melee Attack Editor includes the following fields:

| Field                                                    | Description                                                                                                                                                                                                     |
|----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Hide**                                                 | If checked, hides the attack from the Melee Attacks list on your character sheet.                                                                                                                               |
| **Usage**                                                | Name for the entry. A weapon can have multiple usages.                                                                                                                                                          |
| **Notes**                                                | Notes that appear with the attack entry. Supports [Markdown](Markdown%20Guide) and scripts.                                                                                                                     |
| **Parry**                                                | If checked, the weapon can parry. Enter the parry modifier ([B269](B269)) beside it.                                                                                                                            |
| **Fencing**                                              | Uses fencing parry rules ([B404](B404)).                                                                                                                                                                        |
| **Unbalanced**                                           | Weapon can’t attack and parry on the same turn ([B269](B269)).                                                                                                                                                  |
| **Block**                                                | If checked, the weapon can block. Enter the block modifier ([B327](B327)) beside it.                                                                                                                            |
| **Damage**                                               | The strength basis of the damage (**thr** or **sw**, including lifting, telekinetic, and IQ-based variants, or **None** for fixed damage) ([B268](B268)). Check **Leveled** to multiply it by the item's level. |
| **Multiply ST used for sw or thr damage calculation by** | Changes how Strength modifies damage.                                                                                                                                                                           |
| **Non-leveled damage modifier**                          | Dice or a number added to the damage, unaffected by levels. May be a script.                                                                                                                                    |
| **Leveled damage modifier**                              | Dice or a number added to the damage per level of the item.                                                                                                                                                     |
| **Damage modifier per die**                              | A modifier applied to each die of damage.                                                                                                                                                                       |
| **Armor Divisor**                                        | Adjusts how good the attack is at penetrating DR ([B378](B378)).                                                                                                                                                |
| **Type**                                                 | The type of injury or effect of the attack (cut, imp, cr, etc.) ([B268](B268)).                                                                                                                                 |
| **Fragmentation**                                        | Fragmentation damage ([B414](B414)): the dice, their armor divisor, and their type. The divisor and type are disabled until dice are entered.                                                                   |
| **Reach**                                                | The maximum reach of the weapon ([B269](B269)), with the **Minimum** reach beside it.                                                                                                                           |
| **Close Combat**                                         | Allows use in close combat ([B391](B391)).                                                                                                                                                                      |
| **Reach Change Requires Ready**                          | Changing reach requires a Ready maneuver ([B366](B366)).                                                                                                                                                        |
| **Minimum ST**                                           | Minimum Strength needed to use the weapon ([B270](B270)).                                                                                                                                                       |
| **Two-handed**                                           | Requires two hands ([B270](B270)).                                                                                                                                                                              |
| **Two-handed & unready**                                 | Requires two hands and becomes unready after each attack ([B270](B270)).                                                                                                                                        |
| **Defaults**                                             | Lists default skills for this attack ([B271](B271)). See [Defaults](Defaults).                                                                                                                                  |

## Ranged Attack Editor fields

The Ranged Attack Editor includes the following fields:

| Field                             | Description                                                                                                                                                                                                             |
|-----------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Hide**                          | If checked, hides the attack from the Ranged Attacks list on your character sheet.                                                                                                                                      |
| **Usage**                         | Name for the entry. A weapon can have multiple usages.                                                                                                                                                                  |
| **Notes**                         | Notes that appear with the attack entry. Supports [Markdown](Markdown%20Guide) and scripts.                                                                                                                             |
| **Accuracy**                      | Base accuracy bonus ([B269](B269)), with the **Scope** bonus beside it ([HT155](HT155)).                                                                                                                                |
| **Jet**                           | Attack is treated as a jet (continuous stream) ([B106](B106)). Accuracy and rate of fire don't apply.                                                                                                                   |
| **Damage**                        | The same damage fields as a melee attack: strength basis, **Leveled**, ST multiplier, damage modifiers, armor divisor, type, and fragmentation.                                                                         |
| **Range**                         | The maximum range ([B269](B269)), with the **½ Damage** range and **Minimum** range beside it.                                                                                                                          |
| **Muscle-powered**                | The ranges are multiplied by the user's ST ([B269](B269)).                                                                                                                                                              |
| **Ranges are in miles**           | Range is measured in miles instead of yards.                                                                                                                                                                            |
| **Rate of Fire**                  | For each of **Mode 1** and **Mode 2**: shots per attack and secondary projectiles ([B373](B373)), with **Fully Automatic Only** ([B408](B408)) and **High-cyclic Controlled Bursts** ([HT83](HT83)) options.            |
| **Shots**                         | Number of shots before reloading ([B270](B270)), with the shots **In Chamber**, the **Duration** of a continuous attack, the **Reload Time**, whether reloading is **Per Shot**, and whether it is a **Thrown Weapon**. |
| **Bulk**                          | Size and handling penalty ([B270](B270)), with the bulk **For Giants** and whether the weapon has a **Retracting Stock**.                                                                                               |
| **Recoil**                        | Minimum margin of success needed for extra hits ([B271](B271)), with the recoil **For Slugs**.                                                                                                                          |
| **Minimum ST**                    | Minimum Strength needed to use the weapon ([B270](B270)), with **Two-handed**, **Two-handed & unready**, **Has bipod**, **Mounted**, and **Uses a musket rest** ([B412](B412)).                                         |
| **Defaults**                      | Lists default skills for this attack ([B275](B275)). See [Defaults](Defaults).                                                                                                                                          |

> [!TIP]
> Damage fields accept dice such as `1d+2` or `2d-1x3`, and the other numeric fields accept a number. A field can also
> hold a script that works out the value. See the [Scripting Guide](Scripting%20Guide). [Features](Features) on
> traits, skills, and equipment can adjust any of these stats, and **Sets the value of** can replace one outright.

## Attack Editor toolbar

The toolbar above the attack editor gives you quick access to common actions:

| Button                                                  | Name                | Description                                                        |
|:--------------------------------------------------------|---------------------|--------------------------------------------------------------------|
| ![help](images/icons/icn-help.svg)                      | **Help**            | Opens this page from the User Guide.                               |
| 100%                                                    | **Scale**           | Adjusts the display size of the attack editor.                     |
| ![apply changes](images/icons/icn-applyChanges.svg)     | **Apply Changes**   | Saves and closes the attack editor (`Ctrl+Return` / `Cmd+Return`). |
| ![discard changes](images/icons/icn-discardChanges.svg) | **Discard Changes** | Exits without saving changes (`Escape`).                           |

---

[Return to Home](Home)
