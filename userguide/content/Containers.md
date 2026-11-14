# Containers

Containers are one of the most important organizational tools in GCS. They let you group items for clarity, automate
totals (like weight or points), and power advanced features in [character templates](Character%20Templates).

You can create containers for [Traits](Traits), [Skills](Skills), [Spells](Spells), [Equipment](Equipment),
[Notes](Notes), [Trait Modifiers](Trait%20Modifiers), and [Equipment Modifiers](Equipment%20Modifiers).

Containers let you **group and organize items**. For example:

- Put multiple advantages in an “Ancestry” group.
- Nest a **pouch** inside a **backpack** for equipment.
- Group spells by college or theme.

A container is identified by the **expand/collapse arrow** ![expand folder](images/icons/icn-folderExpand.svg) next to
its name. Depending on the type, you may also see extra details, such as the **container type** (traits) or the
**choice rule** (templates).

Containers can also hold other containers, so you can build complex structures when needed.

- Expand or collapse a container with the arrow ![expand folder](images/icons/icn-folderExpand.svg).
- Show or hide **all containers** with the expand/collapse button ![collapse folder](images/icons/icn-expand.svg).
- Totals (points, weight, value) include all contents automatically, depending on container type.

## Creating a container

To create a container:

1. On your character sheet, go to **Insert**, open the submenu for the kind of item, and choose **New … Container**
   (for example, **Insert > Trait > New Trait Container** or **Insert > Equipment > New Carried Equipment Container**).
   You can also right-click a list and select **New Container**.
2. The container's editor opens. Fill in its fields (see field tables below) and select **Apply Changes**
   ![apply changes](images/icons/icn-applyChanges.svg).
3. The new empty container appears in the appropriate list of your sheet.

You can also turn an existing item into a container, or back, with **Item > Convert > Convert to Container** and
**Convert to Non-Container**. An item must be empty before it can be converted back.

You can rename, tag, or add notes to containers to keep your sheet organized.

## Moving items in and out of containers

To move something into a container:

1. **Drag and drop** the item within the container.
   - The highlight must appear **indented** to place it inside.
   - If the highlight is flat (not indented), the item will go _below_ the container instead of inside it.
2. To move an item back out, drag it until the highlight is no longer indented.

From the keyboard, select the items and use **Item > Move > Move Into Container** (`Ctrl+Shift+Right` /
`Cmd+Shift+Right`), which puts them at the top of the container directly below them, or **Move Out of Container**
(`Ctrl+Shift+Left` / `Cmd+Shift+Left`), which lifts them out to sit just above it.

> [!TIP]
> For equipment, container contents count toward the **container’s weight and value** totals (for example, a backpack
> full of gear).

## Trait container types

A trait container's **Container Type** sets how GCS treats its contents:

| Container Type            | Description                                                                                                                                                                                                                                                      |
|---------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Group**                 | Just groups the traits. The points go to the categories the traits themselves belong to.                                                                                                                                                                         |
| **Alternative Abilities** | Applies the Alternative Abilities rule ([P11](P11)): the most expensive abilities cost full price and the rest cost 20%. **Alternative Slots** sets how many can be in use at once; a container with more than one slot shows the count, such as _Alternate x2_. |
| **Ancestry**              | Marks the container as a racial template. Its points count as **Ancestry** in the Points block, and the **Ancestry** chosen beside it drives the character's random name, gender, age, height, weight, and looks. See [Name Generators](Name%20Generators).      |
| **Attributes**            | Marks traits that adjust attributes, such as _Increased ST_, so their points count as **Attributes** in the Points block.                                                                                                                                        |
| **Meta-Trait**            | Marks the container as a meta-trait ([B262](B262)).                                                                                                                                                                                                              |

## Equipment containers and groups

An equipment container is itself a piece of equipment, such as a backpack: it has a quantity, value, and weight of its
own, and can have modifiers, attacks, and features. A **group** only organizes the equipment in it and is worth and
weighs just what its contents do. Create a group with **Insert > Equipment > New Carried Equipment Group**, or turn a
container into one with **Item > Convert > Convert to Group**. See [Equipment](Equipment).

## Template choices

In a [character template](Character%20Templates), a **choice** is a special kind of container that asks the player to
pick from its contents when the template is applied. A trait, skill, spell, or equipment choice shows its rule after its
name, such as _Pick 1_ or _Pick 20 points worth_.

To create one in a template:

- Go to **Insert > Trait > New Trait Choice** (or the **Skill**, **Spell**, or **Equipment** equivalent), or right-click
  a list and select **New Choice**.
- Or select an existing container and choose **Item > Convert > Convert to Choice**. **Convert to Group** turns a
  choice back into an ordinary container, dropping its rule.

A choice's editor has only **Name**, **Notes**, **Choices**, **Page Reference**, and **Page Highlight**. The **Choices**
row sets the rule with [Operators](Operators):

- **Count:** The player must pick a number of items (_is_, _is at least_, _is at most_, …).
- **Points:** The player's picks must total a point value (traits, skills, and spells).
- **Value** or **Weight:** The player's picks must total a money value or a weight (equipment).

### Example choices

| Statement                                           | Choice                 |
|-----------------------------------------------------|------------------------|
| **Advantages**: _X points chosen from among..._     | `Points` `is at most`  |
| **Disadvantages**: _-X points chosen from among..._ | `Points` `is at least` |
| **Skills**: _Select X skills from:_                 | `Count` `is`           |

Containers inside a choice are normally picked as a unit. Clear **Picked as a Unit** in a container's editor to offer
what it holds one by one instead. You can also **nest choices** (for example, a “Pick 1 of 3 Skills” choice inside a
“Pick 10 Points of Skills” choice); the picker asks about the inner choice when its container is picked.

> [!TIP]
> If you’re not building templates, you can ignore choices. Only a template can hold them: copying a choice to a
> character sheet makes the choice on the way, so only the picked options arrive.

For recommendations on how to structure choices effectively, see **Best practices for templates** on [Character
templates](Character%20Templates).

## Trait container fields

Containers for [Traits](Traits) include the following fields:

| Trait Container Field       | Description                                                                                                                                                              |
|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Name**                    | The container’s name (e.g., “Advantages,” “Psionic Powers,” “Alternate Form”).                                                                                           |
| **Notes**                   | Notes that appear under the container. Select the Scripting Guide ![scripting guide](images/icons/icn-scriptingGuide.svg) for instructions on how to include JavaScript. |
| **VTT Notes**               | Notes for use in your Virtual Tabletop system. See your VTT’s documentation.                                                                                             |
| **User Description**        | Additional notes for your reference. These appear as a tooltip when you hover, and only exist on character sheets.                                                       |
| **Tags**                    | Comma-separated [Tags](Tags) for organization.                                                                                                                           |
| **Preconfigured**           | Only in templates and libraries. See [Substitutions](Substitutions).                                                                                                     |
| **Picked as a Unit**        | Only for a group inside a template choice. Clear it to offer the group's contents one by one.                                                                            |
| **Enabled**                 | If checked, the container and its contents are active on your character sheet.                                                                                           |
| **Switched On**             | Whether the switchable [features](Features) of the container's modifiers are on.                                                                                         |
| **Self-Control**            | The self-control roll to resist ([B120](B120)). The CR rating will show in parentheses on your character sheet.                                                          |
| **Self-Control Adjustment** | Adjustments tied to self-control rolls, such as reaction or fright check penalties.                                                                                      |
| **Frequency of Appearance** | How often the container's subject appears in play ([B36](B36)).                                                                                                          |
| **Container Type**          | Defines special grouping rules: Group, Alternative Abilities, Ancestry, Attributes, or Meta-Trait (see above).                                                           |
| **Ancestry**                | Available when Container Type is Ancestry. Choose the ancestry (for example, Human or Elf).                                                                              |
| **Alternative Slots**       | Available when Container Type is Alternative Abilities. How many of the container's abilities may be in use at once.                                                     |
| **Page Reference**          | Page number in the source book. See [Page References](Page%20References) for details.                                                                                    |
| **Page Highlight**          | Text to highlight in the PDF. Leave blank unless the default highlights fail.                                                                                            |
| **Nameables**               | Defines the `@code@` placeholders the container uses. See [Substitutions](Substitutions).                                                                                |
| **Prerequisites**           | Any requirements for this container. See [Prerequisites](Prerequisites).                                                                                                 |
| **Trait Modifier**          | Modifiers that apply to every trait in the container. See [Trait Modifiers](Trait%20Modifiers).                                                                          |

## Skill or Spell container fields

[Skills](Skills) and [Spells](Spells) containers support the following fields:

| Skill / Spell Container Field | Description                                                                                   |
|-------------------------------|-----------------------------------------------------------------------------------------------|
| **Name**                      | The container’s name (for example, “Combat Skills” or “Fire Spells”).                         |
| **Notes**                     | Notes that appear under the container.                                                        |
| **VTT Notes**                 | Notes for use in your Virtual Tabletop system. See your VTT’s documentation.                  |
| **Tags**                      | Comma-separated tags to help organize containers.                                             |
| **Picked as a Unit**          | Only for a container inside a template choice. Clear it to offer the contents one by one.     |
| **Page Reference**            | The page number in the source book. See [Page References](Page%20References).                 |
| **Page Highlight**            | Text to highlight in the PDF. Leave blank unless the default highlights fail.                 |

## Equipment container fields

Containers for [Equipment](Equipment) share the same fields as normal equipment, and groups have only the name, notes,
tags, and page reference fields.

See _Organizing items with containers and groups_ in [Equipment](Equipment) for more on converting items to containers
and back. A container must be empty before it can be converted back to a non-container.

## Notes container fields

Containers for [Notes](Notes) have the same fields as notes:

| Note Container Field | Description                                                                             |
|----------------------|-----------------------------------------------------------------------------------------|
| **Notes**            | A multi-line text field for notes. Supports scripting and [Markdown](Markdown%20Guide). |
| **Tags**             | Comma-separated [Tags](Tags) to help organize notes.                                    |
| **Page Reference**   | The page number in the source. See [Page References](Page%20References).                |
| **Page Highlight**   | Text to highlight in the PDF. Leave blank unless the default highlights fail.           |
| **Nameables**        | Defines the `@code@` placeholders the notes use. See [Substitutions](Substitutions).    |
| **Markdown Preview** | Preview of how the notes will render.                                                   |

## Modifier groups and choices

[Trait Modifiers](Trait%20Modifiers) and [Equipment Modifiers](Equipment%20Modifiers) can be grouped too. A modifier
**group** just organizes modifiers. A modifier **choice** makes them exclusive: a mandatory choice asks for exactly one
of its modifiers (_Pick 1_), such as the one that sets the price of a trait whose cost varies, and an optional choice
asks for at most one (_Pick at most 1_). See [Trait Modifiers](Trait%20Modifiers).

## Container Editor toolbar

The toolbar above a container's editor gives you quick access to common actions:

| Button                                                  | Name                    | Description                                                                                            |
|:--------------------------------------------------------|-------------------------|--------------------------------------------------------------------------------------------------------|
| ![help](images/icons/icn-help.svg)                      | **Help**                | Opens the help page from the User Guide.                                                               |
| 100%                                                    | **Scale**               | Adjusts the display size of the editor.                                                                |
| ![apply changes](images/icons/icn-applyChanges.svg)     | **Apply Changes**       | Saves and closes the editor.                                                                           |
| ![discard changes](images/icons/icn-discardChanges.svg) | **Discard Changes**     | Exits without saving changes.                                                                          |
| ![substitutions](images/icons/icn-substitutions.svg)    | **Set Substitutions**   | Lets you enter values for containers with placeholders in their text.                                  |
| ![source](images/icons/icn-source.svg)                  | **ID & Library Source** | Shows the container's ID and library source, and offers **Sync with Source** and **Clear Source**.     |

---

[Return to Home](Home)
