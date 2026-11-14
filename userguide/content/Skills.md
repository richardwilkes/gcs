# Skills

The **Skills list** on your [character sheet](Character%20Sheet%20Overview) shows all of your character’s skills
([B167](B167)) and techniques ([B229](B229)), their levels, and points spent.

## Master Library

GCS includes the full list of official skills in the **Master Library**. See [Library Explorer](Library%20Explorer) for
more details.

You can also create your own custom skills.

## Skills list

The skills list on your character sheet includes the following columns. Select a column heading to sort ascending
![sort ascending](images/icons/icn-sortAscending.svg) or descending ![sort
descending](images/icons/icn-sortDescending.svg).

| Column                                                | Description                                                                                                                                                                                                              |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ![switched on](images/icons/icn-toggleSwitch.svg)     | Appears only when a skill has switchable [features](Features). Select the cell to turn the skill's switch on or off.                                                                                                     |
| **Skill / Technique**                                 | The skill or technique name, specializations (in parentheses), and any notes. Select the chevron ![notes expand](images/icons/icn-notesExpand.svg) to show or hide the notes for this skill.                             |
| **SL**                                                | The calculated skill level, based on its controlling attribute and points spent. Hover over it to see how it was calculated.                                                                                             |
| **RSL**                                               | The relative skill level compared to its controlling attribute. For example, `DX+2`.                                                                                                                                     |
| **Pts**                                               | The total points you’ve invested in the skill.                                                                                                                                                                           |
| ![page reference](images/icons/icn-pageReference.svg) | Page reference for the skill. See [Page References](Page%20References) for more details.                                                                                                                                 |
| ![source](images/icons/icn-source.svg)                | Tracks whether the skill matches the library it came from. See _Master Library Syncing_ in [Character Sheet Overview](Character%20Sheet%20Overview).                                                                     |

A skill whose [prerequisites](Prerequisites) aren't met is tagged **Unsatisfied prerequisite(s)**, and a skill put on
the sheet by another item's grant is marked with a chain icon ![chain](images/icons/icn-chainLink.svg).

The sheet's skill level caps can hold a skill below the level its points buy. A capped skill's **SL** tooltip names the
cap, and the Skill Editor's **Level** shows the capped level. See _Skill Level Caps_ in [Sheet
Settings](Sheet%20Settings).

## Adding skills

To add a skill:

1. Open a skills list from the **Master Library**.
2. Choose the skill you want.
3. Add the skill in one of two ways:
   - Drag it from the list to your sheet, or
   - Right-click it and select **Copy to Character Sheet**.

You can also [create a new custom skill](Custom%20Content):

- On your character sheet, go to **Insert > Skill > New Skill** (or **New Technique**), or right-click the Skills list
  and select **New Skill**.

> [!IMPORTANT]
> Some skills have [prerequisites](Prerequisites). If you add one without meeting its requirements, GCS tags it with
> **Unsatisfied prerequisite(s)**. Hover over the tag to see what’s missing.

When you add a skill:

- GCS highlights the block where it can be dropped. No matter where you drag it, it goes to the right location.
- The **Points block** updates automatically. Points spent on skills reduce your unspent points.
- The skill’s starting level is calculated automatically, based on its controlling attribute. Changing an attribute can
  change the skill level.
- Adding a skill the character already has, with the same specializations and tech level, adds its points to the
  existing one instead of creating a second row.
- Some skills include [substitution **placeholders**](Substitutions), such as a specialization to fill in. GCS prompts
  you for them when the skill is added.

### Points and costs

- Most skills require at least **1 point** to purchase.
- Higher skill levels cost more points, based on the buying skills rules and the **Skill Cost Table** ([B170](B170)).
- A skill with no points in it shows the level it has at default, but doesn't count as a skill the character has for
  [prerequisites](Prerequisites).

## Managing skills

You can change skills on your sheet in a few ways.

### Incrementing points vs. increasing levels

**Increment/Decrement:**

- Adds or removes **one point** at a time.
- Depending on the skill cost table ([B170](B170)), several points may be required before the skill increases to the
  next level.

To add or remove points:

- Right-click the skill and select **Increment** or **Decrement**, or
- Go to **Item > Adjust > Increment** (`Ctrl+=` / `Cmd+=`) or **Decrement** (`Ctrl+-` / `Cmd+-`).

**Increase/Decrease Skill Level:**

- Raises or lowers the skill’s **actual level** directly.
- GCS automatically adds or removes the correct number of points based on the skill cost table.

To increase or decrease the skill level:

- Right-click the skill and select **Increase Skill Level** or **Decrease Skill Level**, or
- Go to **Item > Adjust > Increase Skill Level** (`Ctrl+/` / `Cmd+/`) or **Decrease Skill Level** (`Ctrl+.` /
  `Cmd+.`).

> [!TIP]
> Use **Increment** when you’re tracking gradual training. Use **Increase Skill Level** if you want to jump directly to
> the next full level.

### Swapping defaults

A skill can be bought up from another skill it defaults to. Select a skill and go to **Item > Swap Defaults**
(`Ctrl+Shift+X` / `Cmd+Shift+X`) to cycle through its usable [defaults](Defaults), or to swap the points with the skill
it defaults from.

### Changing the tech level

For skills with a tech level, right-click the skill and select **Increase Tech Level** (`Ctrl+]` / `Cmd+]`) or
**Decrease Tech Level** (`Ctrl+[` / `Cmd+[`).

### Tracking study time

You can also track study time for skills. See [Study time](Study%20time) for more details.

### Organizing skills with containers

Skills can be grouped into [Containers](Containers) for easier navigation.

For Example:

- Put **Broadsword** and **Axe/Mace** into a “Melee Combat Skills” container.
- In a [character template](Character%20Templates), use a **skill choice** to enforce “pick X skills” rules.

Drag skills to rearrange them, or select them and use **Item > Move > Move Up**, **Move Down**, **Move Into
Container**, and **Move Out of Container**.

### Deleting a skill

To delete a skill:

- Right-click it and select **Delete**, or
- Select it and go to **Edit > Delete**.

## Editing skills

You can edit a skill in a few ways:

- Double-click it to open the **Skill Editor**.
- Right-click it and select **Open Detail Editor**.
- Select a skill and go to **Item > Open Detail Editor** (`Ctrl+I` / `Cmd+I`).

Save your changes when you're done from the **Skill Editor toolbar** (see below).

## Skill Editor fields

The Skill Editor includes the following fields:

| Skill Editor Field          | Description                                                                                                                                                                                                                          |
|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Name**                    | The skill’s name.                                                                                                                                                                                                                    |
| **Required Specialization** | The specialty the skill must be taken in, if any ([B169](B169)).                                                                                                                                                                     |
| **Optional Specialization** | An optional specialty ([B169](B169)). Setting one makes the skill one step easier, shown in **Adjusted Difficulty**.                                                                                                                 |
| **Tech Level**              | The Tech Level associated with the skill, if applicable. This field is greyed out unless **Required** is checked. Leave it blank in a library or template to fill in the character's TL when the skill is added.                     |
| **Required**                | If checked, the skill has a Tech Level ([B168](B168)).                                                                                                                                                                               |
| **Notes**                   | Notes about the skill. These appear below the skill in the list. Select the Scripting Guide ![scripting guide](images/icons/icn-scriptingGuide.svg) for instructions on including JavaScript. Supports [Markdown](Markdown%20Guide). |
| **VTT Notes**               | Notes for use in your Virtual Tabletop system. See your VTT’s documentation.                                                                                                                                                         |
| **Tags**                    | Comma-separated [Tags](Tags) to help organize skills.                                                                                                                                                                                |
| **Switched On**             | Whether the skill's switchable [features](Features) are on.                                                                                                                                                                          |
| **Preconfigured**           | Only in templates and libraries. If checked, the skill's substitutions are already settled, so GCS doesn't ask about them when the skill is added to a sheet. See [Substitutions](Substitutions).                                    |
| **Difficulty**              | Select the controlling attribute ([B167](B167)) and the difficulty level ([B168](B168)). Options include Easy (E), Average (A), Hard (H), Very Hard (VH), and Wildcard (W) ([B175](B175)).                                           |
| **Encumbrance Penalty**     | Multiplies the current encumbrance level. Applies to skills like Stealth, Climbing, and Swimming ([B17](B17)).                                                                                                                       |
| **Points**                  | Total character points invested in this skill. Only shown on character sheets and templates.                                                                                                                                         |
| **Level**                   | Read-only. Shows the calculated skill level and relative skill level (`12/DX+2`). Levels are calculated based on the GURPS skill cost table ([B170](B170)).                                                                          |
| **Page Reference**          | The page number where the skill appears. Select the bookmark ![page reference](images/icons/icn-pageReference.svg) beside the label to open it. See [Page References](Page%20References).                                            |
| **Page Highlight**          | Text to highlight in the PDF. Leave blank unless the default highlights fail.                                                                                                                                                        |
| **Nameables**               | Defines the `@code@` placeholders the skill uses, with the choices offered for each. See [Substitutions](Substitutions).                                                                                                             |
| **Prerequisites**           | Requirements for this skill. See [Prerequisites](Prerequisites).                                                                                                                                                                     |
| **Defaults**                | Lists default skills and values ([B173](B173)). See [Defaults](Defaults).                                                                                                                                                            |
| **Features**                | Lists attribute, condition, skill, spell, reaction, or attack changes applied by the skill. See [Features](Features).                                                                                                                |
| **Melee Attack Usage**      | Melee attack statistics ([B268](B268)), including defaults. See [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons).                                                                                                          |
| **Ranged Attack Usage**     | Ranged attack statistics ([B268](B268)), including defaults. See [Melee and Ranged Weapons](Melee%20and%20Ranged%20Weapons).                                                                                                         |
| **Study**                   | Only on character sheets. Tracks study time for this skill. See [Study time](Study%20time).                                                                                                                                          |

### Techniques

A technique ([B229](B229)) uses the same editor, with these differences:

- There are no specializations or tech level.
- **Defaults To** names the skill (or attribute) the technique is based on and the adjustment from it, such as
  _Skill Karate whose specialization is anything at -2_.
- **Cannot exceed default skill level by more than** sets the technique's limit, if it has one.
- **Difficulty** is **A** (Average) or **H** (Hard) only.

## Skill Editor toolbar

The toolbar above the Skill Editor gives you quick access to common actions:

| Button                                                  | Name                    | Description                                                                                                                                                                                   |
|:--------------------------------------------------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ![help](images/icons/icn-help.svg)                      | **Help**                | Opens this page from the User Guide.                                                                                                                                                          |
| 100%                                                    | **Scale**               | Adjusts the display size of the Skill Editor.                                                                                                                                                 |
| ![apply changes](images/icons/icn-applyChanges.svg)     | **Apply Changes**       | Saves and closes the Skill Editor (`Ctrl+Return` / `Cmd+Return`).                                                                                                                             |
| ![discard changes](images/icons/icn-discardChanges.svg) | **Discard Changes**     | Exits without saving changes (`Escape`).                                                                                                                                                      |
| ![substitutions](images/icons/icn-substitutions.svg)    | **Set Substitutions**   | Lets you enter values for skills with placeholders in their specialization or notes.                                                                                                          |
| ![source](images/icons/icn-source.svg)                  | **ID & Library Source** | Shows the skill's ID and, if it came from a library, its source. Select an entry to copy it. Offers **Sync with Source** and **Clear Source**, which take effect when you apply your changes. |

---

[Return to Home](Home)
