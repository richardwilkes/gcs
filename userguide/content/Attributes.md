# Attributes

The **Primary and Secondary Attribute blocks** track your character’s Basic Attributes ([B14](B14)) and Secondary
Characteristics ([B15](B15)). These values form the foundation of your [character sheet](Character%20Sheet%20Overview)
and affect many other parts of GCS, including [Traits](Traits), [Skills](Skills), and [Spells](Spells).

## Primary attributes

Your **Primary Attributes** are the four basic attributes ([B14](B14)): ST, DX, IQ, and HT.

To set your primary attributes:

- In the **Primary Attributes block**, enter a score for each attribute.

As you change these values, GCS will:

- Show the point cost next to each attribute (number in brackets).
- Update your unspent [character points](Character%20Points) in the **Points block**.
- Adjust related secondary attributes.
- Apply defaults to [Traits](Traits), [Skills](Skills), and [Spells](Spells).

> [!TIP]
> Hover over an attribute's name or value to see where its bonuses come from. The tooltip lists each source and its
> amount. Strength bonuses that apply only to striking, lifting, or throwing are listed separately.

## Secondary attributes

Your **Secondary Attributes** are your secondary characteristics ([B15](B15)) that depend on your primary attributes.

To set your secondary attributes:

- In the **Secondary Attributes block**, enter a score for each value.

## Adjusted attributes

A [pool threshold](Point%20pools) can add to or multiply an attribute while the pool is at that threshold. This
changes only the attribute's current value, leaving its score as it was:

- The sheet shows the current value in the **Warning** color (see [Colors and Fonts](Colors%20and%20Fonts)). Hover over
  it to see the value it's adjusted from and the thresholds adjusting it.
- Select the field to edit the attribute. While you edit, the field shows the value without the adjustments.
- Skills, spells, skill defaults, attribute prerequisites, Move, Dodge, and ST for striking, lifting, and throwing use
  the current value. Point costs, and attributes calculated from this one, such as HP from ST, use the value without the
  adjustments.

## Using attribute options from other games

Some GURPS supplements and house rules use different attributes or point pools. The Master Library includes several of
these options in its `Settings/Attributes` folder.

> [!TIP]
> If you’re not using a supplement or house rule listed here, stick with the default GURPS attributes.

To switch your attributes:

1. On the character sheet toolbar, select **Attributes** ![attributes](images/icons/icn-attributes.svg), or go to
   **Sheet > Attributes**.
2. On the Attributes toolbar, select **Menu** ![menu](images/icons/icn-menu.svg).
3. Choose an option from the table below. They are listed under the **Master Library** heading.
4. Choose whether to:
   - **Remove existing attribute definitions, replacing them with all of these**, or
   - **Merge the checked attribute definitions into the existing ones**. Attributes that already exist keep their
     place, and the rest are added after them.
5. Select **Import**.
6. On the Attributes toolbar, select **Apply Changes** ![apply changes](images/icons/icn-applyChanges.svg) to save.

| Attribute option                            | Notes                                                                                                                                                                    |
|---------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **After the End**                           | Adds Radiation Threshold Points (RP) ([ATE1:24](ATE1:24)).                                                                                                               |
| **Fantastic Dungeon Grappling**             | Adds Control Maximum (CM) levels to Secondary Attributes.                                                                                                                |
| **Mad as Bones**                            | Adds the Sanity attributes used by _Mad as Bones_.                                                                                                                       |
| **PY83 HT based PER**                       | Perception based on HT instead of IQ ([PY83:21](PY83:21)).                                                                                                               |
| **QN and HT based PER**                     | Adds Quintessence (QN) and Quintessence Points (QP) ([PY120:21](PY120:21)); Perception based on HT ([PY83:21](PY83:21)).                                                 |
| **Quintessence**                            | Adds Quintessence (QN) and Quintessence Points (QP) ([PY120:21](PY120:21)).                                                                                              |
| **Separate Will and Per from IQ**           | Will and Perception become Primary Attributes, independent of IQ.                                                                                                        |
| **Star Wars**                               | Adds Force Sight to Secondary Attributes.                                                                                                                                |
| **Star Wars - with Energy Reserve (Force)** | Adds Force Sight to Secondary Attributes and Energy Reserve (Force) Points (ER).                                                                                         |
| **Stress and Derangement**                  | Adds the Stress and Derangement pools ([H141](H141)).                                                                                                                    |
| **Threshold Magic**                         | Adds Threshold and Recharge Rate (Primary Attributes), plus Excess Power Tally and Calamity Roll Bonus (Secondary Attributes), and Power Tally Points (PT) ([T76](T76)). |

## Importing and exporting attributes

You can share custom attribute settings between characters or reuse them in different sheets by importing or exporting.

To import or export attributes:

1. On the Attributes toolbar, select **Menu** ![menu](images/icons/icn-menu.svg).
2. Select one of the following:
   - **Import…:** Load a saved set of attributes into your sheet, replacing or merging as described above.
   - **Export…:** Save the current attributes as a file.

> [!TIP]
> Save an exported `.attr` file in the `Settings` folder of your User Library and it appears in the menu alongside the
> Master Library's options.

## Editing attributes

You can edit the underlying calculations for attributes or create custom attributes.

> [!WARNING]
> Editing attributes can affect calculations across your character sheet. Proceed with caution.

To edit or add attributes:

1. On the character sheet toolbar, select **Attributes** ![attributes](images/icons/icn-attributes.svg).
2. On the editor, do one of the following:
   - **Edit** an existing attribute. Each attribute is collapsed to a header showing its ID, name, and type; select the
     header to show its fields, and again to hide them.
   - Select **Add Attribute** ![add](images/icons/icn-add.svg) on the toolbar to add a new attribute, already expanded.
3. Enter values in the form using the fields below.
4. On the Attributes toolbar, select **Apply Changes** ![apply changes](images/icons/icn-applyChanges.svg) to save.

To expand every attribute at once, select **Expands/collapses all attribute definitions**
![toggle expansion](images/icons/icn-expand.svg) on the Attributes toolbar. When any attribute is already expanded, it
collapses them all instead.

Every change can be undone with **Edit > Undo** while the view is open. Undoing a change to a collapsed attribute
expands it.

### Attribute fields

| Field              | Description                                                                                                                                                                                                                                                                                                              |
|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ID**             | Unique variable name used for references, in lowercase.                                                                                                                                                                                                                                                                  |
| **Attribute Type** | Choose from **Integer**, **Decimal**, **Pool**, or a **Separator**. The “Display Only” variants create a read-only field, and **Pool (Display Only for Maximum)** lets you track the current value of a pool whose maximum is calculated. Anything below a separator can be collapsed or expanded on the sheet.          |
| **Placement**      | Where the attribute appears on the character sheet: **Automatic**, **Primary**, **Secondary**, or **Hidden**. With **Automatic**, the attribute appears in the Primary or Secondary block depending on whether its base value is a primary attribute. Pools are always placed automatically, in the Point Pools block.   |
| **Trait Name**     | Shown when the placement is **Hidden**. Enter the name of a trait that, when present on the character, reveals the attribute with the **Placement When Present** you choose. Leave it empty to always keep the attribute hidden.                                                                                         |
| **Short Name**     | The label shown on the character sheet.                                                                                                                                                                                                                                                                                  |
| **Full Name**      | Optional. If entered, displays after the short name.                                                                                                                                                                                                                                                                     |
| **Base Value**     | Starting value for the attribute. Supports scripting and attribute IDs (for example, `Math.floor($basic_speed)`). See [Scripting Guide](Scripting%20Guide).                                                                                                                                                              |
| **Cost per Point** | Point cost for each point above or below the base value.                                                                                                                                                                                                                                                                 |
| **SM Reduction**   | The reduction in cost for each SM greater than 0, as a percentage.                                                                                                                                                                                                                                                       |
| **Name**           | Separators only. The title shown on the separator.                                                                                                                                                                                                                                                                       |

### Point pools

See [Point pools](Point%20pools) for details on customizing your pools.

## Managing attributes

- **Discard changes:** Select **Discard Changes** ![discard changes](images/icons/icn-discardChanges.svg).
- **Delete an attribute:** Select **Remove attribute** ![delete](images/icons/icn-delete.svg) beside it. The last
  attribute can't be deleted.
- **Reorder attributes:** Drag an attribute using its grip ![grip](images/icons/icn-grip.svg).
- **Reset attributes to defaults:** Select **Reset** ![reset](images/icons/icn-reset.svg) to restore the sheet's
  attributes to the current [Default Attributes](Default%20Settings).

## Default attributes

You can change the default set of attributes for new characters in [Default Attributes](Default%20Settings).

---

[Return to Home](Home)
