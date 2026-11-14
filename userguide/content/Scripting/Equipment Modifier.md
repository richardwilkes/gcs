# Scripting: Equipment Modifier

> [Return to the Scripting Guide.](../Scripting%20Guide)

This object represents a single Equipment Modifier. Only active Equipment Modifiers (which are never containers) can be accessed via scripting.

| Property   | Type                | Value                                                                                    | Availability |
|------------|---------------------|------------------------------------------------------------------------------------------|--------------|
| id         | string              | The object ID.                                                                           | GCS v5.41+   |
| attachedTo | object \| undefined | The [Equipment](Equipment) this modifier is attached to or undefined if there isn't one. | GCS v5.41+   |
| name       | string              | The name.                                                                                | GCS v5.41+   |
| techLevel  | string              | The Tech Level.                                                                          | GCS v5.41+   |
| weapons    | object[]            | Returns a list of [Weapons](Weapon) this modifier gives its Equipment.                   | GCS v5.58+   |
| notes      | string              | The notes.                                                                               | GCS v5.41+   |
| tags       | string[]            | The associated tags. Always an array, which is empty when there are none.                | GCS v5.41+   |

| Method      | Arguments                                   | Result Type | Result                                                                                                                                                                                  | Availability |
|-------------|---------------------------------------------|-------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| findWeapons | melee: boolean, name: string, usage: string | object[]    | Returns the [Weapons](Weapon) in `weapons` that match the given **name** and **usage**, case-insensitively. When matching, an empty or undefined value passed in will match everything. | GCS v5.58+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
