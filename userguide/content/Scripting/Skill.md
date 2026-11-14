# Scripting: Skill

> [Return to the Scripting Guide.](../Scripting%20Guide)

This object represents a single Skill or Skill container. Disabled skills are never visible to scripts.

The **For** column indicates whether the property or method is defined for a container (C), a leaf (L), or both.

| For | Property               | Type                  | Value                                                                                                                           | Availability |
|:---:|------------------------|-----------------------|---------------------------------------------------------------------------------------------------------------------------------|--------------|
|  CL | id                     | string                | The object ID.                                                                                                                  | GCS v5.36+   |
|  CL | parentID               | string \| undefined   | The parent's object ID or undefined if there isn't one.                                                                         | GCS v5.36+   |
|  CL | parent                 | object \| undefined   | The parent Skill or undefined if there isn't one.                                                                               | GCS v5.41+   |
|  CL | name                   | string                | The name.                                                                                                                       | GCS v5.36+   |
|  L  | specialization         | string                | The required specialization.                                                                                                    | GCS v5.36+   |
|  L  | optionalSpecialization | string                | The optional specialization.                                                                                                    | GCS v5.43+   |
|  L  | techLevel              | string                | The Tech Level, or an empty string if one is not required.                                                                      | GCS v5.36+   |
|  L  | kind                   | string                | The kind of Skill, one of **skill** or **technique.**                                                                           | GCS v5.36+   |
|  L  | attribute              | string                | The attribute ID used to determine the final level.                                                                             | GCS v5.36+   |
|  L  | difficulty             | string                | The difficulty key, one of **e, a, h, vh,** or **w**.                                                                           | GCS v5.36+   |
|  L  | points                 | number                | The number of points spent on this Skill. 0 for a skill put on the sheet by a grant.                                            | GCS v5.36+   |
|  L  | level                  | number                | The computed level, with any skill level caps from [Sheet Settings](../Sheet%20Settings) applied, or 0 if it can't be computed. | GCS v5.36+   |
|  L  | relativeLevel          | number                | The computed level relative to the controlling attribute, with any skill level caps applied, or 0 if it can't be computed.      | GCS v5.36+   |
|  L  | weapons                | object[]              | Returns a list of [Weapons](Weapon) attached to this skill.                                                                     | GCS v5.41+   |
|  CL | notes                  | string                | The notes.                                                                                                                      | GCS v5.39+   |
|  CL | tags                   | string[]              | The associated tags. Always an array, which is empty when there are none.                                                       | GCS v5.36+   |
|  CL | container              | boolean               | True if this is a container.                                                                                                    | GCS v5.36+   |
|  L  | switchedOn             | boolean               | True if the switch controlling any switchable features of this Skill is on.                                                     | GCS v5.47+   |
|  C  | children               | object[] \| undefined | The contained, enabled child Skills. Undefined if this isn't a container.                                                       | GCS v5.36+   |

| For | Method      | Arguments                                         | Result Type | Result                                                                                                                                                                                                                                                                                                                     | Availability |
|:---:|-------------|---------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
|  C  | find        | name: string, specialization: string, tag: string | object[]    | Returns the Skills within the hierarchy of this container that match the given **name** and **specialization**, and have the given **tag**, case-insensitively. The specialization matches either the required or the optional specialization. When matching, an empty or undefined value passed in will match everything. | GCS v5.36+   |
|  L  | findWeapons | melee: boolean, name: string, usage: string       | object[]    | Returns the [Weapons](Weapon) that match the given **name** and **usage**, case-insensitively. When matching, an empty or undefined value passed in will match everything.                                                                                                                                                 | GCS v5.41+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
