# Fixed-cost trait containers

Choose **Fixed Cost** in a trait container's editor to give a package a fixed point cost while authoring a template
or a trait library. Its row displays a **Fixed** tag. The **Fixed Points** field sits beside the container settings;
the separate **Choices** controls still configure how children are selected when applying a template.

## How points are calculated

For an enabled Fixed Cost container, the first applicable rule determines its cost:

1. A value entered in **Fixed Points** overrides both the picker value and the children's total.
2. If Fixed Points is unset and Choices specifies **exactly** a number of **points**, that number supplies the cost.
3. Otherwise, the cost is the sum of the children's adjusted costs.

A count-based choice or an at-most, at-least, or not-equal points condition does not supply a fixed cost. Those
conditions describe selection constraints, so they fall back to the child total when there is no manual value.
An effectively disabled container costs zero, including when a manual value is present.

For example, with children totaling 30 points and a choice of exactly 20 points, an unset Fixed Points field gives
the container a cost of 20. Entering 15 changes its cost to 15. Entering 0 makes its cost zero.

## Setting and clearing Fixed Points

An empty field means **unset**, represented in the model as `FixedPoints == nil`. It enables the picker/children
fallback described above. An explicitly entered **0** is a real manual value and overrides that fallback.

There are two ways to restore the unset value:

- Delete the textbox contents, leaving it empty.
- Click **Clear** beside the textbox.

Both methods restore `nil`, even if the previous value was zero. Entering 0 again sets an explicit zero. Apply the
editor changes to save them. Changing the container type away from Fixed Cost also clears the fixed value and
disables the field and its Clear button.

## Templates, libraries, and character sheets

Fixed Cost is available for templates and trait libraries, not for character-sheet editing. Cloning or assigning
one to an owner with a character entity converts it to **Group** and clears Fixed Points, including for nested
containers. Cloning it to a template or library preserves its fixed-cost settings.

Once on a character sheet, the resulting group is charged for its remaining children. The template's manual cost
does not become a discount or surcharge on the character.

The Choices picker validates the selected children's total against its qualifier; a parent's manual fixed value
does not substitute for that total. Existing picker behavior replaces a choice container with its selected children.
To keep an enclosing package on the sheet, put a choice group inside the Fixed Cost container. The enclosing
container becomes a Group, and the selected children replace the inner choice group.

## Stored data

The container type is `container.FixedCost` in Go and `fixed_cost` in saved data. The optional `fixed_points` field
is omitted when unset. An explicit zero is stored, survives saving and loading, and has a different source hash
from an unset value.
