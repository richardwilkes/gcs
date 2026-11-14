# Scripting: Attribute

> [Return to the Scripting Guide.](../Scripting%20Guide)

This object represents a single Attribute.

Any attributes you’ve defined for the sheet are available as globals using their ID prefixed with a dollar sign, e.g. `$st` refers to the Strength attribute. You can also access these via the [Entity](Entity).

For convenience, if you pass the attribute object as a value, it will resolve as if you had passed its maximum value. e.g. `$hp + 1` when the current HP is 8 with a maximum of 10, it would resolve to a value of 11 (10 + 1). For an attribute other than a pool, the maximum leaves out the adjustments the pools' current thresholds make; use `$st.current` (or `entity.attribute("st").current`) for the value with them. The exception is the base value of an attribute with the ID `striking_st`, `lifting_st`, `throwing_st`, `dodge`, or `move`: when GCS works out that attribute's current value, each attribute other than a pool resolves to its current value there.

| Property  | Type    | Value                                                                                                                                                                                                                                                                                                           | Availability |
|-----------|---------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| id        | string  | The ID for this attribute.                                                                                                                                                                                                                                                                                      | GCS v5.36+   |
| kind      | string  | The kind of attribute this refers to. One of **primary, secondary,** or **pool.**                                                                                                                                                                                                                               | GCS v5.36+   |
| name      | string  | The short name of the attribute, or its full name when it has no short name.                                                                                                                                                                                                                                    | GCS v5.36+   |
| fullName  | string  | The full name of the attribute.                                                                                                                                                                                                                                                                                 | GCS v5.36+   |
| maximum   | number  | For a pool, its maximum. For any other attribute, its value without the adjustments the pools' current thresholds make, which its point cost and the attributes calculated from it use. Undefined when the script isn't running for a character sheet.                                                          | GCS v5.36+   |
| current   | number  | For a pool, the points it has left. For any other attribute, its value with the adjustments the pools' current thresholds make, which skills, spells, Move, Dodge, and ST for striking, lifting, and throwing use. Prior to GCS v5.59, this was always the same as the maximum for attributes other than pools. | GCS v5.36+   |
| isDecimal | boolean | True if values other than whole numbers might be present.                                                                                                                                                                                                                                                       | GCS v5.36+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
