# Equipment

The **Equipment lists** are where you add your character's gear ([B264](B264)), including weapons and armor, to your
[character sheet](Character%20Sheet%20Overview).

Equipment is divided into two lists:

- **Carried Equipment:** Items on your person (equipped or not). These count toward encumbrance.
- **Other Equipment:** Items you own but aren’t carrying. These do not count toward encumbrance and can’t be equipped.
  Items here don’t provide any bonuses or stats in play.

> [!TIP]
> Keep your equipment lists tidy by collapsing containers, sorting columns (e.g., Name, Weight, Value), or searching for
> items in the character sheet toolbar.

## Tracking encumbrance

Carried items count toward encumbrance ([B17](B17)). The Carried Equipment list automatically totals the weight of all
carried gear and highlights your current encumbrance level ![weight](images/icons/icn-weight.svg) in the **Encumbrance,
Move & Dodge block**.

### Master Library

GCS includes hundreds of official equipment items in the **Master Library**. See [Library Explorer](Library%20Explorer)
for more details.

You can also create your own custom equipment.

## Equipment lists

The equipment lists on your character sheet include the following columns. Select a column heading to sort ascending
![sort ascending](images/icons/icn-sortAscending.svg) or descending ![sort
descending](images/icons/icn-sortDescending.svg).

Each list's heading includes the **total weight and value** for all items in that list.

| Column                                                                          | Description                                                                                                                                                                       |
|---------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **![equipped](images/icons/icn-applyChanges.svg)**                              | If checked, the equipment is **equipped** and any bonuses apply to your character. Available only for Carried Equipment.                                                          |
| **![switched on](images/icons/icn-toggleSwitch.svg)**                           | Appears only when an item has switchable [features](Features). Select the cell to turn the item's switch on or off.                                                               |
| **#**                                                                           | Quantity of the item. Equipment with a quantity of zero is dimmed and adds nothing to the sheet, including its attacks.                                                           |
| **Carried Equipment** / **Other Equipment**                                     | The item's name and any notes, including usage information. Select the chevron ![notes expand](images/icons/icn-notesExpand.svg) to show or hide the notes.                       |
| **TL**                                                                          | Tech level ([B22](B22)). Can be hidden in [Sheet Settings](Sheet%20Settings).                                                                                                     |
| **LC**                                                                          | Legality class ([B267](B267)). Can be hidden in [Sheet Settings](Sheet%20Settings).                                                                                               |
| **![value](images/icons/icn-coins.svg)**                                        | Value of one item.                                                                                                                                                                |
| **![weight](images/icons/icn-weight.svg)**                                      | Weight of one item.                                                                                                                                                               |
| **![stack](images/icons/icn-stack.svg) ![value](images/icons/icn-coins.svg)**   | Total value, including the quantity and any contained items.                                                                                                                      |
| **![stack](images/icons/icn-stack.svg) ![weight](images/icons/icn-weight.svg)** | Total weight, including the quantity and any contained items.                                                                                                                     |
| **![page reference](images/icons/icn-pageReference.svg)**                       | Page reference for the equipment. See [Page References](Page%20References) for more details.                                                                                      |
| **![source](images/icons/icn-source.svg)**                                      | Tracks whether the equipment matches the library it came from. See _Master Library Syncing_ in [Character Sheet Overview](Character%20Sheet%20Overview).                          |

The number of decimal places shown for weights and values can be changed in [Sheet Settings](Sheet%20Settings). Hover
over a rounded value to see the exact figure.

## Adding equipment

To add equipment to your sheet:

1. Open an equipment list from the **Master Library**.
2. Choose the equipment you want.
3. Add an item in one of two ways:
   - Drag it from the list to your sheet, or
   - Right-click it and select **Copy to Character Sheet**.

You can also [create new custom equipment](Custom%20Content):

- On your character sheet, go to **Insert > Equipment > New Carried Equipment** (or **New Other Equipment**), or
  right-click an equipment list and select **New Equipment**.

Some equipment has **modifiers** or [substitution **placeholders**](Substitutions). GCS prompts you for them when the
item is added, and you can change them later in the **Equipment Editor**.

### Weapons and armor

When adding weapons or armor:

- **Weapons ([B267](B267)):** Equipped weapons appear on the **Melee Attacks** and/or **Ranged Attacks** lists with
  their stats. See [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons) for more details.
- **Armor ([B282](B282)):** Equipped armor updates the **Damage Resistance (DR)** values on the body type block based
  on the protection the armor provides.

## Buying and selling equipment

Your character can pay for gear with the money they carry, and sell gear for the money it brings in. Buying spends
carried, non-container equipment tagged `Money`, such as the coins in the Master Library (a tag like `Money: GURPS
Standard Medieval Coins` counts too). Sales pay out in the coins set as **Sell Money** in the sheet's [Sheet
Settings](Sheet%20Settings), which also hold the default **Buy Percent** and **Sell Percent**.

### Buying

1. Go to **Sheet > Buy Equipment…**. The **Buy Equipment** pane opens for the sheet.
2. Add the items to buy:
   - Drop equipment into the pane from any equipment list, including a library, a loot sheet, or the sheet itself.
   - Or select items in an equipment list and choose **Sheet > Buy Selected Equipment…** or **Buy…** from the
     right-click menu.
   - While the pane is open, equipment dragged onto the sheet from elsewhere joins the purchase instead.
   - Drop [equipment modifiers](Equipment%20Modifiers) onto an item in the pane to adjust it before buying.
3. Set the quantity of each item, and change the **Transaction** name and **Percent of Full Price** if you like. The
   pane shows the total and warns when the character doesn't carry enough money.
4. Select **Finalize Purchase…**. The dialog picks the fewest coins that cover the cost. Change the amounts to spend, or
   uncheck a kind of coin to leave it out, then select **Finalize**. Any change due is returned in the sheet's Sell
   Money coins or, failing that, the kinds of coin paid.

Bought items arrive unequipped in Carried Equipment, merged into matching items unless **Merge with matching items** is
turned off. The purchase is a single undoable edit.

### Selling

1. Select the items to sell in the **Carried Equipment** list.
2. Go to **Sheet > Sell Equipment…** or choose **Sell…** from the right-click menu.
3. In the dialog, set the quantity of each item to sell and the **Percent of full price**. A container is sold with its
   contents, which the dialog lists. The dialog shows what you get, and warns when no Sell Money is configured or the
   total can't be paid exactly.
4. Select **Sell**.

The payment arrives in the sheet's Sell Money coins, merged into matching money already carried unless **Merge with
matching money already carried** is turned off. The sale is a single undoable edit.

### Shopping Record

Every purchase and sale is logged in the sheet's **Shopping Record**, with its time, transaction name, kind, items and
quantities, and the money that changed hands. Open it with **Sheet > Shopping Record** or the **Shopping Record**
![shopping record](images/icons/icn-releaseNotes.svg) button on the character sheet toolbar. Entries can't be edited,
but you can select entries and press `Delete` to remove them.

## Managing equipment

### Equipping and carrying items

Items are equipped by default when added to **Carried Equipment**. Equipped items provide bonuses and update stats on
your sheet.

- Clear the **check mark** to carry an item without equipping it.
- You can also right-click and select **Toggle State** (or go to **Item > Toggle State**, `Ctrl+'` / `Cmd+'`) to
  equip/unequip quickly.

When you unequip a **[container](Containers)**, all items inside it are also automatically unequipped. Those contained
items don’t contribute any of their features until the container is re-equipped, or you move them out and equip them
individually.

### Moving items around

Move items to the **Other Equipment list** if you aren’t carrying them:

- Drag it to the **Other Equipment** list,
- Right-click it and select **Move to Other Equipment**, or
- Select **Item > Move > Move to Other Equipment**.

You can move items back to **Carried Equipment** the same way.

To rearrange items within a list, drag them, or select them and use **Item > Move > Move Up**, **Move Down**, **Move
Into Container**, and **Move Out of Container** (`Ctrl+Shift+Arrow` / `Cmd+Shift+Arrow`).

### Adjusting quantities

You can increase (or decrease) the quantity of items in a few ways:

- Right-click it and select **Increment** (or **Decrement**).
- Select an item and go to **Item > Adjust > Increment** (or **Decrement**).
- Double-click it to open the **Equipment Editor**, then update the **Quantity** field.

### Adjusting levels

You can change the Tech Level or Equipment Level of an item:

- **Increase Tech Level:** Right-click the item and select **Increase Tech Level** (`Ctrl+]` / `Cmd+]`).
- **Increase Equipment Level:** Right-click the item and select **Increase Equipment Level**.

You can also change levels from the **Item > Adjust** menu.

### Tracking uses

For items with a **Maximum Uses**, such as potions or ammunition, right-click the item and select **Increase Uses**
(`Ctrl+Up` / `Cmd+Up`), **Decrease Uses** (`Ctrl+Down` / `Cmd+Down`), or **Reset Uses to Maximum**. The remaining uses
appear in the item's notes.

### Organizing items with containers and groups

[Containers](Containers) are items that can hold other items, such as backpacks or pouches (and other containers). These
help organize gear and automatically total contained weight and value. A **group** only organizes the equipment in it:
it has no quantity, value, weight, modifiers, attacks, or features of its own, so it is worth and weighs just what its
contents do.

Containers and groups show an arrow ![expand folder](images/icons/icn-folderExpand.svg) next to their name. Select the
arrow to expand/collapse the contents.

You can convert an item to a container (or a non-container), and a container to a group (or back):

- Right-click an item and select **Convert to Container**, **Convert to Non-Container**, **Convert to Group**, or
  **Convert to Container**, or
- Select an item and use the **Item > Convert** menu.

You cannot convert an item back to a non-container if it still contains equipment. You must move items out of the
container first. Converting a container to a group drops what a group can't hold, and GCS warns you first. See
[Containers](Containers) for more details.

### Deleting equipment

To delete equipment:

- Right-click and select **Delete**, or
- Select the item and go to **Edit > Delete**.

## Editing equipment

You can edit equipment in a few ways:

- Double-click it to open the **Equipment Editor**.
- Right-click it and select **Open Detail Editor**.
- Select an item and go to **Item > Open Detail Editor** (`Ctrl+I` / `Cmd+I`).

Save your changes when you're done from the **Equipment Editor toolbar** (see below).

## Equipment Editor fields

The Equipment Editor includes the following fields:

| Equipment Editor Field       | Description                                                                                                                                                                                                                        |
|------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Name**                     | The item’s name.                                                                                                                                                                                                                   |
| **Notes**                    | Notes about the item. These appear below the item in the list. Select the Scripting Guide ![scripting guide](images/icons/icn-scriptingGuide.svg) for instructions on including JavaScript. Supports [Markdown](Markdown%20Guide). |
| **VTT Notes**                | Notes for use in your Virtual Tabletop system. See your VTT’s documentation.                                                                                                                                                       |
| **Tech Level**               | The Tech Level ([B22](B22)) associated with the item, if applicable.                                                                                                                                                               |
| **Legality Class**           | The Legality Class ([B267](B267)).                                                                                                                                                                                                 |
| **Quantity**                 | How many of the item you have.                                                                                                                                                                                                     |
| **Equipped**                 | Whether the item is equipped. Equipped items apply any bonuses to your character. Only shown for carried equipment.                                                                                                                |
| **Value**                    | Value of one item. May be a number or a script. See the [Scripting Guide](Scripting%20Guide).                                                                                                                                      |
| **Extended Value**           | Read-only. Total value based on quantity, modifiers, and contents. Shows a range while a modifier choice is still open.                                                                                                            |
| **Weight**                   | Weight of one item, with optional units. May be a number or a script.                                                                                                                                                              |
| **Extended Weight**          | Read-only. Total weight based on quantity, modifiers, and contents.                                                                                                                                                                |
| **Ignore weight for skills** | If checked, it excludes the item’s weight from encumbrance penalties for skills. It does not reduce overall encumbrance.                                                                                                           |
| **Switched On**              | Whether the item's switchable [features](Features) are on.                                                                                                                                                                         |
| **Preconfigured**            | Only in templates and libraries. If checked, the item's modifier choices and substitutions are already settled, so GCS doesn't ask about them when the item is added to a sheet. See [Substitutions](Substitutions).               |
| **Uses Left**                | The number of uses available. Greyed out unless Maximum Uses > 0.                                                                                                                                                                  |
| **Maximum Uses**             | Maximum number of uses. **Adjusted Maximum Uses** shows the maximum after any [features](Features) that change it.                                                                                                                 |
| **Rated ST**                 | Strength rating for using the item. Equipment with a rated ST uses this value instead of the user's ST.                                                                                                                            |
| **Level**                    | The item’s level, used by features and modifiers that have per-level effects.                                                                                                                                                      |
| **Tags**                     | Comma-separated [Tags](Tags) to help organize items.                                                                                                                                                                               |
| **Page Reference**           | The page number where the item appears. Select the bookmark ![page reference](images/icons/icn-pageReference.svg) beside the label to open it. See [Page References](Page%20References).                                           |
| **Page Highlight**           | Text to highlight in the PDF. Leave blank unless the default highlights fail.                                                                                                                                                      |
| **Nameables**                | Defines the `@code@` placeholders the item uses, with the choices offered for each. See [Substitutions](Substitutions).                                                                                                            |
| **Prerequisites**            | Requirements for this item. See [Prerequisites](Prerequisites) for details.                                                                                                                                                        |
| **Features**                 | Lists any attribute, condition, or stat changes the item provides. See [Features](Features) for details.                                                                                                                           |
| **Equipment Modifier**       | Any modifications to the item's base cost and weight. See [Equipment Modifiers](Equipment%20Modifiers) for details.                                                                                                                |
| **Melee Attack Usage**       | Melee attack stats, if applicable. See [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons).                                                                                                                                 |
| **Ranged Attack Usage**      | Ranged attack stats, if applicable. See [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons).                                                                                                                                |

Groups have only the **Name**, **Notes**, **VTT Notes**, **Equipped**, **Tags**, **Page Reference**, and **Page
Highlight** fields.

## Equipment Editor toolbar

The toolbar above the Equipment Editor gives you quick access to common actions:

| Button                                                  | Name                    | Description                                                                                                                                                                                  |
|:--------------------------------------------------------|-------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ![help](images/icons/icn-help.svg)                      | **Help**                | Opens this page from the User Guide.                                                                                                                                                         |
| 100%                                                    | **Scale**               | Adjusts the display size of the Equipment Editor.                                                                                                                                            |
| ![apply changes](images/icons/icn-applyChanges.svg)     | **Apply Changes**       | Saves and closes the Equipment Editor (`Ctrl+Return` / `Cmd+Return`).                                                                                                                        |
| ![discard changes](images/icons/icn-discardChanges.svg) | **Discard Changes**     | Exits without saving changes (`Escape`).                                                                                                                                                     |
| ![substitutions](images/icons/icn-substitutions.svg)    | **Set Substitutions**   | Lets you enter values for equipment with placeholders in their text.                                                                                                                         |
| ![source](images/icons/icn-source.svg)                  | **ID & Library Source** | Shows the item's ID and, if it came from a library, its source. Select an entry to copy it. Offers **Sync with Source** and **Clear Source**, which take effect when you apply your changes. |

---

[Return to Home](Home)
