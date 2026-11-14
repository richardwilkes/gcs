# Scripting: Note

> [Return to the Scripting Guide.](../Scripting%20Guide)

This object represents a single Note or Note container.

The **For** column indicates whether the property or method is defined for a container (C), a leaf (L), or both.

| For | Property    | Type                | Value                                                   | Availability |
|:---:|-------------|---------------------|---------------------------------------------------------|--------------|
| CL  | id          | string              | The object ID.                                          | GCS v5.43+   |
| CL  | parentID    | string \| undefined | The parent's object ID or undefined if there isn't one. | GCS v5.43+   |
| CL  | parent      | object \| undefined | The parent Note or undefined if there isn't one.        | GCS v5.43+   |
| CL  | description | string              | The description.                                        | GCS v5.43+   |
| CL  | tags        | string[]            | The associated tags.                                    | GCS v5.43+   |
| CL  | container   | boolean             | True if this is a container.                            | GCS v5.43+   |
|  C  | children    | object[]            | The contained child Notes.                              | GCS v5.43+   |

| For | Method | Arguments                        | Result Type | Result                                                                                                                                                                                                               | Availability |
|:---:|--------|----------------------------------|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
|  C  | find   | description: string, tag: string | object[]    | Returns the Notes within the hierarchy of this Note that match the given **description** and have the given **tag**, case-insensitively. When matching, an empty or undefined value passed in will match everything. | GCS v5.43+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
