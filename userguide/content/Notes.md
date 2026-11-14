# Notes

You can add **Notes** to your [character sheet](Character%20Sheet%20Overview) for extra details, reminders, or
references.

Notes are useful for:

- Campaign background and story notes.
- Notes for your gameplay.
- Quick rules references or house rules.

Notes always show their **full content** in the Notes block of your character sheet (you don’t have to open them
separately to read). Notes support **[Markdown](Markdown%20Guide)** and **Scripting** for formatting and automation, so
a note can show calculated values, such as `<script>entity.name</script>`.

## Adding notes

To add a note:

- Go to **Insert > Note > New Note** (`Ctrl+Shift+N` / `Cmd+Shift+N`), or
- Right-click the Notes list and select **New Note**.

You can also keep notes in a notes library (**File > New > New Notes Library**) and drag them onto sheets, or download
the community-maintained **GURPS Rules Lookup** notes file with **Settings > Download GURPS Rules Lookup File** and open
it from your User Library.

### Organizing notes with containers

Notes can be placed in [Containers](Containers) to group related notes or collapse long lists.

For example:

- A container for **House Rules** with several note entries inside.

Notes can also be tagged, so you can find them with [saved filters](Saved%20Filters) in a notes library. See
[Tags](Tags).

## Deleting a note

To delete a note:

- Right-click it and select **Delete**, or
- Select it and go to **Edit > Delete**.

## Editing notes

You can edit a note in a few ways:

- Double-click it to open the **Note Editor**.
- Right-click it and select **Open Detail Editor**.
- Select a note and go to **Item > Open Detail Editor** (`Ctrl+I` / `Cmd+I`).

Type your notes directly in the **Notes** field. The **Markdown Preview** below it shows how the note will look as you
type.

Save your changes when you're done from the **Note Editor toolbar** (see below).

## Note Editor fields

The Note Editor includes the following fields:

| Note Field           | Description                                                                                                                                                                                            |
|----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Notes**            | A multi-line text field for notes. Supports [Markdown](Markdown%20Guide) and scripts; select the Scripting Guide ![scripting guide](images/icons/icn-scriptingGuide.svg) for details.                  |
| **Tags**             | Comma-separated [Tags](Tags) to help organize notes.                                                                                                                                                   |
| **Preconfigured**    | Only in templates and libraries. If checked, the note's substitutions are already settled, so GCS doesn't ask about them when the note is added to a sheet. See [Substitutions](Substitutions).        |
| **Page Reference**   | The page number in the source. Select the bookmark ![page reference](images/icons/icn-pageReference.svg) beside the label to open it. See [Page References](Page%20References).                        |
| **Page Highlight**   | Text to highlight in the PDF. Leave blank unless the default highlights fail.                                                                                                                          |
| **Nameables**        | Defines the `@code@` placeholders the note uses, with the choices offered for each. See [Substitutions](Substitutions).                                                                                |
| **Markdown Preview** | Preview of how the note will render on your character sheet, with any scripts resolved.                                                                                                                |

## Note Editor toolbar

The toolbar above the Note Editor gives you quick access to common actions:

| Button                                                  | Name                    | Description                                                                                                                                                                                  |
|:--------------------------------------------------------|-------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ![help](images/icons/icn-help.svg)                      | **Help**                | Opens this page from the User Guide.                                                                                                                                                         |
| 100%                                                    | **Scale**               | Adjusts the display size of the Note Editor.                                                                                                                                                 |
| ![apply changes](images/icons/icn-applyChanges.svg)     | **Apply Changes**       | Saves and closes the Note Editor (`Ctrl+Return` / `Cmd+Return`).                                                                                                                             |
| ![discard changes](images/icons/icn-discardChanges.svg) | **Discard Changes**     | Exits without saving changes (`Escape`).                                                                                                                                                     |
| ![substitutions](images/icons/icn-substitutions.svg)    | **Set Substitutions**   | Lets you enter values for notes with placeholders in their text.                                                                                                                             |
| ![source](images/icons/icn-source.svg)                  | **ID & Library Source** | Shows the note's ID and, if it came from a library, its source. Select an entry to copy it. Offers **Sync with Source** and **Clear Source**, which take effect when you apply your changes. |

---

[Return to Home](Home)
