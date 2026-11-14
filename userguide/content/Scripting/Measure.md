# Scripting: Measure

> [Return to the Scripting Guide.](../Scripting%20Guide)

This object provides measurement utilities.

Valid units for lengths are: **in, ft, ft_in, yd, mi, cm, m,** and **km.**

Valid units for weights are: **oz, lb, \#, tn, t, g,** and **kg.**

| Method               | Arguments                                       | Result Type | Result                                                                                                                                             | Availability |
|----------------------|-------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| formatLength         | inches: number, toUnits: string                 | string      | Formats a length (in inches) into a string with the given units.                                                                                   | GCS v5.36+   |
| lengthToInches       | value: number, fromUnits: string                | number      | Converts a length value of the given units into inches.                                                                                            | GCS v5.36+   |
| stringLengthToInches | str: string, defaultUnits: string               | number      | Converts a string containing a length into inches. If no units are found in the string, then **defaultUnits** will be used to interpret the value. | GCS v5.36+   |
| formatWeight         | pounds: number, toUnits: string                 | string      | Formats a weight (in pounds) into a string with the given units.                                                                                   | GCS v5.36+   |
| weightToPounds       | value: number, fromUnits: string                | number      | Converts a weight value of the given units into pounds.                                                                                            | GCS v5.36+   |
| stringWeightToPounds | str: string, defaultUnits: string               | number      | Converts a string containing a weight into pounds. If no units are found in the string, then 'defaultUnits' will be used to interpret the value.   | GCS v5.36+   |
| rangeModifier        | yards: number                                   | number      | Converts the given **yards** into a range modifier.                                                                                                | GCS v5.36+   |
| sizeModifier         | yards: number                                   | number      | Converts the given **yards** into a size modifier.                                                                                                 | GCS v5.36+   |
| modifier             | length: number, units: string, forSize: boolean | number      | Converts the given length into either a range modifier (if **forSize** is false) or a size modifier (if **forSize** is true).                      | GCS v5.36+   |
| modifierToYards      | modifier: number                                | number      | Converts a range or size modifier into yards.                                                                                                      | GCS v5.36+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
