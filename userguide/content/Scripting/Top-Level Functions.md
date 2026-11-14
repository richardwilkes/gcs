# Scripting: Top-Level Functions

> [Return to the Scripting Guide.](../Scripting%20Guide)

These are the top-level functions.

| Method      | Arguments                                             | Result Type | Result                                                                                                                                                                                                                                                         | Availability |
|-------------|-------------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| formatNum   | value: number, withCommas: boolean, withSign: boolean | string      | Returns the **value** as a string with commas every third digit from the right if **withCommas** is true and with a leading `+` if it is zero or positive and **withSign** is true.                                                                            | GCS v5.39+   |
| iff         | condition: boolean, trueValue: any, falseValue: any   | any         | Returns **trueValue** if **condition** is true, otherwise returns **falseValue.** This exists primarily to support the automatic conversions of the old expression engine into Javascript and is essentially the same as `condition ? trueValue : falseValue`. | GCS v5.36+   |
| signedValue | value: number                                         | string      | Returns the **value** as a string with a leading `+` if it is zero or positive.                                                                                                                                                                                | GCS v5.36+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
