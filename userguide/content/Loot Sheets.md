# Loot Sheets

Loot Sheets are a special type of sheet in GCS used to **track treasure, rewards, or collected items** during gameplay.
They’re especially useful for GMs to generate treasure hoards or for players to keep a shared record of loot.

## Creating a loot sheet

To create a new loot sheet:

1. Go to **File > New > New Loot Sheet.**
2. Fill in the fields of the sheet.
3. Add any [Equipment](Equipment) or [Notes](Notes) by dragging them from libraries or other sheets, or with **Insert >
   Equipment > New Equipment**.
4. Organize items with [Containers](Containers) or groups if needed.

Equipment with **modifiers** or [substitution **placeholders**](Substitutions) prompts for them when added, just as on
a character sheet.

## Opening a loot sheet

- **From the User Library:** Double-click the sheet in the [Library Explorer](Library%20Explorer).
- **From another location:** Select **File > Open**, browse to the file, and select **Open**.

## Saving a loot sheet

To save your work:

- **Save:** Select **File > Save**.
- **Save As:** Select **File > Save As**.

## Treasure generation

Loot Sheets can also generate **random treasure hoards** based on weighted probabilities.

1. Create a Loot Sheet with the items you want as possible treasure.
2. For each item, set the **Quantity** to weight its likelihood of being chosen (higher quantities = more common). A
   group's contents are picked from individually.
3. Select **Generate a treasure horde from this loot sheet** ![treasure](images/icons/icn-wand.svg).
4. Enter a **Target (Minimum) Value** and **Maximum Value** for the treasure hoard, then select **OK**.
5. GCS creates a **new Loot Sheet** with a randomized selection of items that fits within the chosen value range. The
   old loot sheet remains unchanged.

For example, if your Loot Sheet contains:

- **Copper coins (100)**
- **Silver coins (10)**
- **Gold coin (1)**

Then copper is **100× more likely** to be selected than gold, and silver is **10× more likely** than gold.

## Giving loot to a character

Drag items from a loot sheet onto a character sheet to give them to the character, or select them and use **Copy to
Character Sheet**. To make the character pay for them instead, select the items and choose **Buy…** from the
right-click menu, which adds them to the character's **Buy Equipment** pane. See [Equipment](Equipment).

## Loot Sheet fields

The **Loot Sheet** includes the following fields:

| Loot Sheet Field | Description                                                                                                                                                                                                            |
|------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Name**         | The loot sheet's name.                                                                                                                                                                                                 |
| **Location**     | Notes about where the loot was obtained.                                                                                                                                                                               |
| **Session**      | Notes about the game session when the loot was found.                                                                                                                                                                  |
| **Equipment**    | Items that make up the loot. Standard [Equipment](Equipment) details are included, like weight, value, and quantity. The columns shown and the decimal places follow the [Default Sheet Settings](Default%20Settings). |
| **Notes**        | Any relevant [Notes](Notes) regarding the loot.                                                                                                                                                                        |

## Loot Sheet toolbar

The toolbar above the Loot Sheet gives you quick access to common actions:

<img src="images/screenshots/scrn-toolbar-lootSheet.png" alt="Loot Sheet Toolbar" width="1156">

| Button                                               | Name                                               | Description                                                                                                   |
|:-----------------------------------------------------|----------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| ![info](images/icons/icn-info.svg)                   | **Info**                                           | Hover to see a reminder that holding `Alt` (`Option` on macOS) while using the mouse wheel changes the scale. |
| 100%                                                 | **Scale**                                          | Adjust the display size of your loot sheet.                                                                   |
| ![toggle expansion](images/icons/icn-expand.svg)     | **Opens/closes all hierarchical rows**             | Show or hide all items within containers at once.                                                             |
| ![toggle notes](images/icons/icn-openCloseNotes.svg) | **Opens/closes all embedded notes**                | Show or hide all embedded notes on your sheet.                                                                |
| ![sync sources](images/icons/icn-syncSources.svg)    | **Sync with all sources in this sheet**            | Sync all items on the loot sheet with the libraries they came from.                                           |
| ![treasure](images/icons/icn-wand.svg)               | **Generate a treasure horde from this loot sheet** | Generate a random treasure hoard based on the sheet’s items and settings.                                     |
| N/A                                                  | **Search**                                         | Type to list the matching rows on the loot sheet. Choose one to select it and scroll to it.                   |

---

[Return to Home](Home)
