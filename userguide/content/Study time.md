# Study Time

You can use GCS to track study hours for [Skills](Skills), [Spells](Spells), or [Traits](Traits). This helps you follow
the **Improvement Through Study** rules ([B292](B292)).

Study only means something on a character sheet, so the **Study** section appears only in editors opened from a sheet.
A row copied into a template or library loses its study entries.

## Changing total hours needed for study

By default, you need **200 hours of study** for 1 character point.

You can change this with the dropdown in the **Study** section's heading, which reads _Studied X of 200 hours (Standard)
for 1 point_, to reflect special cases:

- **Standard:** The default 200 hours.
- **Reduction for Talent level 1-4:** Each Talent level reduces the time needed by **10% per level**.
  - Talent 1: 180 hours
  - Talent 2: 160 hours
  - Talent 3: 140 hours
  - Talent 4: 120 hours

> [!TIP]
> Choose the matching reduction to account for Talents, Magery, or other special rules. This ensures GCS calculates the
> correct required hours.

## Adding study entries

To log study time:

1. Open the editor for the skill, spell, or trait on your character sheet.
2. In the **Study** section, choose the reduction if you need fewer than 200 hours (for example, due to Talents or
   Magery).
3. Select **Add a study entry** ![add](images/icons/icn-add.svg). New entries go at the top of the list.
4. Enter the details for the study entry (see below).

Save your changes when you’re done. See **Saving study entries** below.

## Study entries

Each study entry includes the following fields:

| Field              | Description                                                                                                                                                                    |
|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Study type**     | How you studied. Choose from the **Study types** table below. Hover over the info icon beside it to see the limits that apply to that type, such as the hours per day allowed. |
| **Hours spent**    | Enter the number of hours spent studying.                                                                                                                                      |
| **Hours of study** | Read-only. Shows the hours that count toward the next point, based on the study type.                                                                                          |
| **Note**           | (Optional) Add any notes about this study session.                                                                                                                             |

GCS automatically calculates:

- **Effective study hours**, based on the study type (see below).
- **Total accumulated hours** toward your next point, shown in the section's heading.

You can also enter **negative study hours** to track when study time was _converted into character points_. This lets
you record that you’ve earned a point while keeping a complete study history instead of deleting your old entries.

**Example:** After reaching 200 hours and gaining 1 character point, add a new study entry with **–200 hours** to reset
your progress. Your total hours return to 0, but your log still shows the time you invested.

### Study types

| Study type               | Multiplier | Description                                |
|--------------------------|------------|--------------------------------------------|
| **Self-Taught**          | ×0.5       | Every 2 hours count as 1 hour of learning. |
| **On-the-Job Training**  | ×0.25      | Every 4 hours count as 1 hour of learning. |
| **Professional Teacher** | ×1.0       | Counts at full value.                      |
| **Intensive Training**   | ×2.0       | Every hour counts as 2 hours of learning.  |

## Deleting study entries

To remove a study entry:

- Select **Remove this study entry** ![delete](images/icons/icn-delete.svg) beside it.

## Gaining a point

When you’ve reached the required hours (normally 200), you must:

1. **Add 1 [character point](Character%20Points) to your sheet**. This represents the point earned through study.
2. **Increment the skill, spell, or trait** you were studying once to spend the new point (GCS will subtract from your
   total unspent points when applying the point).
3. **Delete the completed study entries**, or add a negative entry, to reset your log for future study.

> [!IMPORTANT]
> GCS does not automatically award points when you reach the required study hours. You must apply the point to your
> character points and update the skill, spell, or trait yourself.

## Saving study entries

On the editor toolbar:

- **Apply changes:** Select **Apply Changes** ![apply changes](images/icons/icn-applyChanges.svg) to save.
- **Discard changes:** Select **Discard Changes** ![discard changes](images/icons/icn-discardChanges.svg) to exit
  without saving.

---

[Return to Home](Home)
