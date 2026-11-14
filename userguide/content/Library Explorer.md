# Library Explorer

In GCS, **libraries** hold all the traits, skills, spells, equipment, and other content you'll add to your [character
sheet](Character%20Sheet%20Overview) to build characters.

- The **Master Library** comes with GCS and includes official GURPS content.
- The **User Library** is where you save and organize your own custom content.
- You can also mark **Favorites** to quickly open the files and folders you use most often.

All of these are available in the **Library Explorer** on the left sidebar of the GCS window.

GCS reads your libraries in the background, so a library on a slow disk or a network share doesn't hold up the window.
The libraries show as empty until that first read finishes.

To explore libraries:

- Select the arrows ![expand folder](images/icons/icn-folderExpand.svg) to expand or collapse folders.
- Double-click a folder to open or close it.
- Select **Opens/closes all hierarchical rows** ![toggle expansion](images/icons/icn-expand.svg) on the toolbar to open
  or close all folders at once.

To open a file:

- Double-click it. The file opens in the **Document Workspace** as its own tab.

## Favorites

If you use a file or folder often, mark it as a **Favorite** so it's always at the top of the Library Explorer.

To create a favorite:

- Select the file or folder and select **Toggle Favorite** ![favorite](images/icons/icn-favorite.svg) on the toolbar,
  or
- Right-click it and select **Toggle Favorite**.

It now appears under **Favorites**, where a favorite folder shows its contents. To remove it, repeat the same steps,
either on the original or on its copy under Favorites.

> [!IMPORTANT]
> You can favorite files and folders inside a library, but not a whole library or individual items inside a file.

## Master Library

The **Master Library** is the built-in collection that comes with GCS. It includes hundreds of official traits, skills,
spells, and equipment taken from GURPS books and other supplements. Many items include notes and explanations, and when
added to your sheet they automatically adjust your stats and [points](Character%20Points) for you.

> [!TIP]
> Use the **Search** field on the toolbar to find files by name. As you type, the matching folders and files are listed
> below the field, each with its icon and location. Choose one with a click, or with the arrow keys and `Return`, to
> select it in the tree. Press `Escape` to hide the list.
>
> <img src="images/screenshots/scrn-libraryExplorer-search.png" alt="Searching the Library Explorer" width="428">
>
> You can also search for individual items within files by checking the file types to search under **Deep search** in
> [General Settings](General%20Settings). A file that matched on its contents opens with the search text already in its
> own search field when you double-click it. While GCS is still indexing the contents of your files, the list ends with
> the note _File contents are still being indexed. Matches on contents will appear when that finishes._, and the
> matches on contents are added once it does.

The Master Library is updated regularly. You can see the current version number next to its name (for example, _v5.16_).
By default, GCS checks for library updates each time it launches. You can change how often it checks with the **Check
for Library Updates** option in [General Settings](General%20Settings).

To update the Master Library:

- Select the Master Library, then select **Update** ![update](images/icons/icn-download.svg) on the toolbar, or select
  the **Update** button that appears beside its name when a new version is available. A progress bar shows the download
  and installation, and **Cancel** stops the update and leaves the library as it was.
- To see what's new, select **Release Notes** ![release notes](images/icons/icn-releaseNotes.svg).

> [!CAUTION]
> **Do not** make changes directly in the Master Library. Updates replace the entire library, and your edits will be
> lost. If you want to customize something, copy it into your User Library first. To guard against accidental edits,
> [lock](#locking-a-list) the lists you open.

By default, content from the Master Library is saved in: `<home directory> > GCS > Master Library`.

### Library list toolbar

When you open a list file from a library, such as _Basic Set Traits_, its toolbar gives you tools to find things in it
and to lock it against changes:

<img src="images/screenshots/scrn-toolbar-library.png" alt="Library list toolbar" width="1158">

| Button                                               | Name                                                  | Description                                                                                                                   |
|:-----------------------------------------------------|-------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| 100%                                                 | **Scale**                                             | Change the display size of the list.                                                                                          |
| ![lock](images/icons/icn-lock.svg)                   | **Lock against changes**                              | Lock the list so that nothing changes it until you select the button again. See **Locking a list** below.                     |
| ![toggle expansion](images/icons/icn-expand.svg)     | **Opens/closes all hierarchical rows**                | Show or hide all items within containers at once.                                                                             |
| ![toggle notes](images/icons/icn-openCloseNotes.svg) | **Opens/closes all embedded notes**                   | Show or hide all embedded notes.                                                                                              |
| ![size to fit](images/icons/icn-sizeToFit.svg)       | **Sets the width of each column to fit its contents** | Resize every column to fit what it holds.                                                                                     |
| N/A                                                  | **Quick Filter**                                      | Type to show only the rows whose text, in any column including the tags, contains what you type.                              |
| **Any Tag** ![tag](images/icons/icn-tag.svg)         | **Tag Filter**                                        | Show only the items carrying the chosen [tags](Tags). Shift-click, or `Ctrl`-click (`Cmd`-click on macOS), to choose several. |
| **No Filter** ![filter](images/icons/icn-filter.svg) | **Saved Filters**                                     | Choose a [saved filter](Saved%20Filters) to narrow the list by any of its fields, or create, edit, and delete saved filters.  |

While a list is filtered, a striped banner notes that items can't be added, removed, or rearranged until the filter is
cleared.

#### Locking a list

Every list file (traits, trait modifiers, skills, spells, equipment, equipment modifiers, and notes) has a **Lock
against changes** ![lock](images/icons/icn-lock.svg) button on its toolbar. Select it to protect the file from
accidental changes, such as while you browse the Master Library. The button stays highlighted while the list is locked;
select it again to unlock the list. To have list files always open locked, turn on **Lock list files when opened** in
[General Settings](General%20Settings). A new list always starts unlocked.

While a list is locked:

- You can't add, delete, duplicate, or rearrange items, drop anything onto the list, or use the **Item > Move**,
  **Item > Adjust**, **Item > Convert**, **Item > Sync with Source**, **Item > Clear Source**, or **Sheet > Organize
  Traits** commands on it.
- The check boxes in its rows can't be changed, and the detail editors you open from it are read-only.
- Selecting a column heading doesn't sort the list. The sort you had comes back when you unlock it.
- **Edit > Undo** and **Edit > Redo** only undo and redo changes to the **Scale**. Earlier changes to the list can be
  undone again once you unlock it.

You can still select, search, and filter its items, follow page references, change the scale, and save a copy with
**File > Save As…**. You can also copy items out of it: drag them onto a sheet, template, or another list, or use **Copy
to Character Sheet** or **Copy to Template**. From a locked modifier list, **Apply Modifier** still works.

Locking or unlocking a list first closes the editors you opened from it. If one of them has unsaved changes, GCS asks
whether to save them; canceling leaves the lock as it was.

## User Library

The **User Library** is your personal space for [custom content](Custom%20Content). It’s where you keep anything that
belongs to _you_ rather than the official Master Library, such as:

- Saved [character sheets](Character%20Sheet%20Overview), [templates](Character%20Templates), and [loot
  sheets](Loot%20Sheets)
- Homebrew traits, skills, spells, or equipment
- [Notes](Notes) and [Markdown](Markdown%20Guide) documents related to your character or game
- Your own ancestries and [name generators](Name%20Generators)
- Anything you don’t want to lose when GCS updates

By default, your User Library is saved in: `<home directory> > GCS > User Library`.

Anything saved here will not change when the Master Library updates.

> [!TIP]
> Right-click a library, folder, or file and select **Show on Disk** to open its location on your computer.

### Creating or Importing Libraries

You can create your own list files or add whole libraries, including ones shared on GitHub.

- **Create a new list file:** Go to **File > New** and choose the type (for example, **New Traits Library**, **New
  Skills Library**, or **New Equipment Library**). GCS opens a new, empty list for you to add [custom
  content](Custom%20Content) to. Save it into your User Library and it appears in the Library Explorer.
- **Import an existing file:** If you’ve downloaded a file (or received one from another player), select **File >
  Open**, choose the file, and then save it into your User Library.
- **Add another library:** Select **Add Library** ![add](images/icons/icn-add.svg) on the toolbar. Give it a **Name**
  and a **Path** on your computer, and, for a library published on GitHub, the **GitHub Account** and **Repository** it
  comes from. GCS then offers updates for it just as it does for the Master Library. Select **Configure**
  ![configure](images/icons/icn-configure.svg) to change these settings later.

### Organizing files

- **New folder:** Select a folder or library, then select **New Folder** ![new folder](images/icons/icn-newFolder.svg)
  on the toolbar and enter a name.
- **Rename:** Select a file or folder, then select **Rename** ![rename](images/icons/icn-sign.svg) and enter the new
  name. Open documents follow the rename.
- **Move:** Drag files and folders to another folder, or use your computer's file manager. The Library Explorer notices
  changes made outside GCS.

### Deleting Libraries and Files

> [!CAUTION]
> When you delete a file or folder, GCS **permanently deletes** it from your computer, not just from the Library
> Explorer. Removing a library you added only removes it from the list and leaves its files on disk.

To delete a file or folder:

- Select it and select **Delete** ![delete](images/icons/icn-delete.svg) on the toolbar, press the `Delete` key, or
- Right-click it and select **Delete**.

## Library Explorer toolbar

The toolbar above the Library Explorer gives you quick access to common actions depending on what's selected:

<img src="images/screenshots/scrn-toolbar-libraryExplorer.png" alt="Library Explorer toolbar" width="428">

| Button                                              | Name                                   | Description                                                                                                                |
|-----------------------------------------------------|----------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| ![Help](images/icons/icn-help.svg)                  | **Help**                               | Open this page from the User Guide.                                                                                        |
| 100%                                                | **Scale**                              | Change the text size in the Library Explorer.                                                                              |
| ![toggle expansion](images/icons/icn-expand.svg)    | **Opens/closes all hierarchical rows** | Show or hide all folders at once.                                                                                          |
| ![Add library](images/icons/icn-add.svg)            | **Add Library**                        | Add a library from a folder on your computer or a GitHub repository.                                                       |
| ![Update](images/icons/icn-download.svg)            | **Update**                             | Replace the selected library with its latest release.                                                                      |
| ![Release Notes](images/icons/icn-releaseNotes.svg) | **Release Notes**                      | See what’s new in the selected library's releases.                                                                         |
| ![Configure](images/icons/icn-configure.svg)        | **Configure**                          | Change the name, location, or GitHub settings of the selected library.                                                     |
| ![New folder](images/icons/icn-newFolder.svg)       | **New Folder**                         | Create a new folder to organize files.                                                                                     |
| ![Rename](images/icons/icn-sign.svg)                | **Rename**                             | Rename the selected file or folder.                                                                                        |
| ![Delete](images/icons/icn-delete.svg)              | **Delete**                             | Permanently delete the selected files or folders from your computer, or remove an added library from the list.             |
| ![Favorite](images/icons/icn-favorite.svg)          | **Toggle Favorite**                    | Add or remove the selected files and folders from Favorites.                                                               |
| ![Software update](images/icons/icn-download.svg)   | **Software Update Available**          | Appears only when a new version of GCS is available. See [Installing and Updating GCS](Installing%20and%20Updating%20GCS). |
| N/A                                                 | **Search**                             | Find files by name, or by their contents with **Deep search** turned on.                                                   |

## File types

Each file type in GCS has its own extension and icon. Here’s what you’ll see in the Library Explorer:

| File type               | Icon                                                              | Extension(s)                                    | Description                                                                                                                                          |
|-------------------------|-------------------------------------------------------------------|-------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Character Sheet**     | ![Character Sheet](images/icons/icn-gcsSheet.svg)                 | `.gcs`                                          | Complete characters you build and play. See [Character Sheet Overview](Character%20Sheet%20Overview).                                                |
| **Character Template**  | ![Template](images/icons/icn-gcsTemplate.svg)                     | `.gct`                                          | Templates you can reuse for new characters. See [Character templates](Character%20Templates).                                                        |
| **Loot Sheet**          | ![loot](images/icons/icn-loot.svg)                                | `.loot`                                         | Lists to track [Loot](Loot%20Sheets) obtained during play.                                                                                           |
| **Traits**              | ![Trait](images/icons/icn-gcsTraits.svg)                          | `.adq`                                          | [Traits](Traits), including advantages, disadvantages, and quirks.                                                                                   |
| **Trait Modifiers**     | ![Trait Modifiers](images/icons/icn-gcsTraitModifiers.svg)        | `.adm`                                          | [Trait Modifiers](Trait%20Modifiers).                                                                                                                |
| **Skills**              | ![Skills](images/icons/icn-gcsSkills.svg)                         | `.skl`                                          | [Skills](Skills) and techniques.                                                                                                                     |
| **Spells**              | ![Spells](images/icons/icn-spells.svg)                            | `.spl`                                          | [Spells](Spells).                                                                                                                                    |
| **Equipment**           | ![Equipment](images/icons/icn-gcsEquipment.svg)                   | `.eqp`                                          | [Equipment](Equipment).                                                                                                                              |
| **Equipment Modifiers** | ![EquipmentModifiers](images/icons/icn-gcsEquipmentModifiers.svg) | `.eqm`                                          | [Equipment Modifiers](Equipment%20Modifiers).                                                                                                        |
| **Notes**               | ![Notes](images/icons/icn-gcsNotes.svg)                           | `.not`                                          | [Notes](Notes).                                                                                                                                      |
| **Ancestry**            | ![Ancestry](images/icons/icn-ancestry.svg)                        | `.ancestry`                                     | The random names, genders, ages, heights, weights, and looks of an ancestry. Opens in the ancestry editor. See [Name Generators](Name%20Generators). |
| **Name Generator**      | ![Name Generator](images/icons/icn-naming.svg)                    | `.names`                                        | A generator of random names for ancestries. Opens in the name generator editor. See [Name Generators](Name%20Generators).                            |
| **Markdown**            | ![Markdown file](images/icons/icn-fileMarkdown.svg)               | `.md`                                           | Documents written in [Markdown](Markdown%20Guide) format.                                                                                            |
| **PDF**                 | ![PDF file](images/icons/icn-filePDF.svg)                         | `.pdf`                                          | PDFs, opened in the built-in viewer or your [external viewer](General%20Settings).                                                                   |
| **Image**               | ![Image file](images/icons/icn-fileImage.svg)                     | `.png`, `.jpg`, `.gif`, `.webp`, `.bmp`, `.svg` | Images or graphics.                                                                                                                                  |

Settings files, such as exported attributes, body types, saved filters, and skill level caps, live in each library's
`Settings` folder. They aren't shown in the Library Explorer, but the settings views that use them list them in their
**Menu** ![menu](images/icons/icn-menu.svg). Export templates live in the `Output Templates` folder and appear under
**File > Export**.

## Library Settings

Most players won’t need this, but advanced users can connect a library to a GitHub repository to share content with
others. Select a library and choose **Configure** ![configure](images/icons/icn-configure.svg) to set its **GitHub
Account**, **Repository**, and, for private repositories, a read-only **GitHub Access Token**. GCS then scans the
repository for release tags and offers them as updates. Check **Use the most recent commit (possibly unreleased) of
this repository** to follow the repository's latest commit instead of its releases.

---

[Return to Home](Home)
