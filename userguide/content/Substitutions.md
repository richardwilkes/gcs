# Substitutions

Many items in the Master Library need details that only make sense when you add them to a character. For example,
_Enemy_ needs to know who the enemy is, and _Hobby Skill_ needs a specialization. **Substitution placeholders** let an
item ask for these details. A placeholder is a code between two `@` characters, such as `@Who@`. When you add the item
to a character sheet or template, GCS prompts you for a value and uses your answer everywhere the placeholder appears.

You can use placeholders in most text fields of traits, skills, spells, equipment, and notes, including the text in
their modifiers, attacks, features, prerequisites, and defaults.

## When GCS asks for substitutions

GCS opens the substitutions prompt when:

- You add an item that has placeholders to a character sheet or template, whether by dragging it from a library,
  copying it, buying it, or applying a template.
- You select **Set Substitutions** ![substitutions](images/icons/icn-substitutions.svg) in the toolbar of an item's
  editor. This button is only enabled when the item has placeholders. The editors for attacks, trait modifiers, and
  equipment modifiers don't have it; use the editor of the trait or equipment that owns them instead.

## Providing values

The prompt shows one row per placeholder, with the placeholder's name on the left and a field for its value on the
right. If several items need values, for example when you apply a template, they all appear in the same prompt. Hover
over a name to see its description, if the placeholder has one.

The field depends on how the placeholder is defined (see **Nameables** below):

- If the placeholder allows any text, the field is editable. Type a value, or pick one of the suggested options, which
  are listed below the field as you type.
- Otherwise, the field is a dropdown and you must select one of the listed options.

Every list starts with **«not set»**, and includes **«empty»** when the placeholder allows a blank answer:

- **«not set»** leaves the placeholder unresolved. GCS displays it as `@Code@` and asks for it again the next time
  the item is copied to a sheet or template, or a template containing it is applied. Selecting **«not set»** is also
  the only way to clear a value you set earlier.
- **«empty»** replaces the placeholder with nothing.

## Nameables

Each `@code@` an item uses can be defined in the **Nameables** section of the item's editor. The definition is what
turns a bare code into a question: the name the player sees, a description, the options to choose from, and whether
any text or a blank answer is allowed. The prompt labels and limits each answer from these definitions, so the text
only needs the plain `@code@`.

The section lists each code as a sentence, such as _`@Who@` asks for **Who**, suggesting A rival, A deity, or any
text; used once_. Select the section's title to collapse it to a summary, or to expand it again.

To define a placeholder:

1. Type `@code@` wherever the answer should appear. A code may use letters, digits, and underscores.
2. In the **Nameables** section, a row reads _`@code@` isn't defined_. Select **Define** beside it. Until it is
   defined, the player is asked for it as any text.
3. Select the row to open it, then fill in the sentence:
   - **asks for:** The name shown to the player in place of the code.
   - **+ option:** Add a choice to offer. Enter the text and press `Return`. Drag the option chips to reorder them.
   - **+ any text:** Let the player type something other than the options.
   - **+ not blank:** Refuse a blank answer.
   - **+ description:** Help text shown when the player hovers over the name.
4. Select **Done**.

Select **Add a nameable** ![add](images/icons/icn-add.svg) to define a code before using it, and **More actions**
![more](images/icons/icn-moreActions.svg) beside a row to delete a definition.

### Codes shared with modifiers

A trait's or equipment's modifiers can use placeholders too. In the owner's editor, the Nameables section lists each
code a modifier uses and says where its answer comes from:

- A modifier code that the owner also uses **shares the item's answer**, so the player is asked only once.
- Otherwise it **is asked for under its own code**. Open the row to make it **take the answer to** one of the owner's
  other codes instead.

In a modifier's own editor, a code the owner defines reads _is defined by the item the modifier belongs to_.

## Preconfigured items in templates and libraries

When you edit an item in a template or library, its editor includes a **Preconfigured** checkbox. Turn it on once
you've made the modifier selections and substitutions the item should always use. When the item is added to a sheet,
GCS skips the prompts for it and only asks for placeholders that still have no value, and for mandatory modifier
choices that still have no pick. This keeps template application short for players. See [Character
Templates](Character%20Templates) for more on building templates.

## Writing placeholders in the text

Older library data packs the whole definition into the placeholder itself, and GCS still reads these forms:

| Form                        | Meaning                                                                                                                                                           |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `@Label@`                   | Any text, including an empty value. With no definition in Nameables, this is how a bare code behaves.                                                             |
| `@Label|One|Two@`           | Must select _One_ or _Two_. Add `*` as a part to also allow free-form text, `?` to allow an empty value, and `tt(text)` to add a description.                     |
| `@Label: One, Two, etc.@`   | A short label, a colon, and a comma-separated list ending in "etc." becomes a label with those items as choices. Free-form text and empty values are allowed.     |
| `@Label: text@`             | A short label and a colon becomes a label, with the rest of the text shown as the description.                                                                   |

A placeholder written this way reads _`@code@` defines itself_ in the Nameables section. For new content, use a plain
`@code@` and define it in Nameables, so it reads the same everywhere it is used.

To use a literal `|` or `\` in a label, choice, or description, put a backslash in front of it: `\|` or `\\`. To show a
literal `@` in a text field without starting a placeholder, write `\@`.

> [!NOTE]
> Nameable definitions are saved by GCS v5.56.0 and later. An older GCS drops them when it saves the item and asks for
> each code as plain text.

---

[Return to Home](Home)
