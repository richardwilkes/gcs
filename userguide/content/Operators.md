# Operators

Operators define how GCS compares values in fields for [Prerequisites](Prerequisites), [Features](Features),
[Defaults](Defaults), [Saved Filters](Saved%20Filters), and [Character Templates](Character%20Templates).

Some operators are used with **text fields** (names, notes, tags), while others are used with **numbers** (levels,
modifiers).

## Available operators

| Operator                              | Works with   | Description                                                      |
|---------------------------------------|--------------|------------------------------------------------------------------|
| `is anything`                         | Text, Number | Always passes                                                    |
| `is` / `is not`                       | Text, Number | Matches exactly / does not match                                 |
| `is at least` / `is at most`          | Number       | Minimum or maximum value                                         |
| `contains` / `does not contain`       | Text         | Must include / must not include the value                        |
| `starts with` / `does not start with` | Text         | Must (not) begin with the value                                  |
| `ends with` / `does not end with`     | Text         | Must (not) end with the value                                    |

Text comparisons ignore case. `is` and `is not` also ignore spaces at either end of the value.

Where a field holds a list, such as an item's tags, a spell's colleges, or a weapon's damage type, the operator is
checked against each entry: `at least one tag is Sword` passes when any tag matches, and `all tags are not Sword`
passes only when none do. In these fields you can also enter several values separated by commas, such as `Sword, Axe`
or `cut, imp`, to match any one of them.

An empty value with `is` has a meaning of its own: a specialization or title note that `is` nothing means the item has
none, and a tag criterion that `is` nothing means the item has no tags at all.

## Using operators

- In **Prerequisites**, operators control how rules check for required traits, skills, or attributes.
- In **Features**, operators define which skills, spells, traits, or weapons a bonus applies to.
- In **Defaults**, operators pick the skill a default draws from and set Tech Level rules.
- In **Saved Filters**, operators decide which items a library list shows.
- In **Character Templates**, operators set how many items, or how many points, a choice requires.

---

[Return to Home](Home)
