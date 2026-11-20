# Library Filter

A **Library Filter** view shows the items of many library lists of one type together, so you can filter them all at
once. For example, it can gather every skill in every library into one list, then show only those tagged `Combat`. It
works with traits, trait modifiers, skills, spells, equipment, equipment modifiers, and notes.

The view only shows what its lists hold, so nothing in it can be changed. You can open as many Library Filter views as
you like, each showing its own lists.

## Opening a library filter

- Go to **View > New Library Filter** (`Ctrl+L` / `Cmd+L`). The view starts on the type and source set you chose last,
  or on **All Traits** the first time.
- In the [Library Explorer](Library%20Explorer), select files, folders, or libraries, right-click, and select **Open in
  Library Filter**. Selected files must all be lists of the same type. When the selected folders and libraries hold
  lists of several types, the menu lists the types beneath **Open in Library Filter:** for you to choose one.

The view opens where library lists open, and its tab is named after what it shows, such as _All Skills_.

## Choosing what to show

The lists a view shows come from its **source set**, which you choose from the **Source Set** dropdown on its toolbar.
Under a heading for each type, the dropdown lists:

- **All Traits**, **All Skills**, and so on: every list of that type in every library.
- **Selected Traits**, **Selected Skills**, and so on: what you opened the view on from the Library Explorer. This set
  isn't saved, and choosing another set discards it. Use **Save Set As…** to keep it.
- The source sets you've saved for that type.

Choosing a set of another type switches the view to that type. The **Quick Filter** text carries over, but the chosen
tags and saved filter are cleared.

A saved set names libraries, folders, and files. A library or folder includes every list of the set's type inside it,
even lists added to it later. Like [saved filters](Saved%20Filters), source sets are stored in your GCS settings, not in
your libraries.

The commands at the top of the dropdown manage the saved sets:

- **New Set…:** Build a new set in the set editor.
- **Edit Set…:** Change the set the view shows. Other views showing it change with it.
- **Save Set As…:** Save what the view shows as a new set, starting from a copy of it in the editor. Saving an **All**
  set names each of your libraries, so a library you add later isn't included.
- **Delete Set…:** Delete the set the view shows, once you confirm it. The view switches to its type's **All** set,
  while other views showing the set keep showing the same lists, as an unsaved set.

**Edit Set…** and **Delete Set…** are only available while the view shows a saved set.

### The set editor

The editor is titled after the command that opened it: **New Set**, **Edit Set**, or **Save Set As**.

- **Name:** Each set needs a name that no other set of the same type uses, ignoring case. It can't be the name of the
  type's **All** or **Selected** set either, such as _All Skills_, so **Save Set As…** on **All Skills** suggests _All
  Skills (2)_.
- **Type:** The type of list the set shows. Changing it drops the files you checked, but keeps the folders and
  libraries.
- **Checklist:** Each library is listed in bold, with its folders and lists of the chosen type beneath it. Check a
  library, folder, or file to include it.
  - The boxes inside a checked library or folder are checked too, and can't be changed, since it includes them.
  - A mixed box means something inside it is checked, but not the library or folder itself, so lists added to it later
    aren't included.
  - Libraries, folders, and files the set names that no longer exist are listed first, under **Can't be found**.
    Uncheck one to drop it from the set.

**OK** stays unavailable until the set has a name and at least one checked box. A message beside the buttons says
when the name is already in use or nothing is checked.

## Filtering and sorting

The **Quick Filter**, the **Any Tag** dropdown (see [Tags](Tags)), and the **No Filter** dropdown of [saved
filters](Saved%20Filters) work just as they do on a [library list's toolbar](Library%20Explorer#library-list-toolbar),
and offer the same saved filters. Matching items keep their place under their containers, and a container shown only
because it holds a match is dimmed.

The last column, **File**, names the list each item comes from. Hover over it to see the list's library, its path
within the library, and its location on your computer. When two lists share a name, the column adds the folder, and if
need be the library, to tell them apart. The **Quick Filter** doesn't search this column, so typing _magic_ doesn't
match every item of a list named _Magic_.

Select a column heading to sort by it. Sorting only changes the order the view shows.

## Working with the items

Nothing can be added, changed, removed, or rearranged here, as the line beneath the toolbar says. Double-click an item
to see it in a read-only editor. You can still:

- Drag items onto a character sheet, template, or another list, or use **Item > Copy to Character Sheet** and **Item >
  Copy to Template**. Copies remember the list they came from, just as they do when copied from the list itself.
- Use **Item > Add as Grant** on traits, skills, and spells (see [Features](Features)).
- Open an item's [page references](Page%20References).
- Use **Item > Apply Modifier** on trait and equipment modifiers.
- Use **Sheet > Buy Selected Equipment…**, or **Buy…** from the right-click menu, on equipment (see
  [Equipment](Equipment)).

### Show in Source File

To change an item, select it and choose **Item > Show in Source File** (`Ctrl+Shift+I` / `Cmd+Shift+I`), or right-click
it and select **Show in Source File**. GCS opens the list the item comes from and selects it there, opening the
containers around it and scrolling it into view.

- Items from several lists open each of those lists. The list of the first selected item ends up in front.
- If more than 10 lists would open, GCS asks first.
- If a filter in the list hides any of the items, the list's filters are cleared.
- If some of the items can't be found in their lists, as happens when a list has changed since the view read it, GCS
  tells you how many.
- If the list opens [locked](Library%20Explorer#locking-a-list), unlock it to make your changes.

## Status and reloading

The line at the bottom of the view counts its files and rows, such as _12 files, 840 rows_. While a filter is in use,
it says how many of the rows match, such as _12 files, 35 of 840 rows match_. It also says how many of the set's
libraries, folders, and files can't be found, and how many files couldn't be read. Hover over it to see which ones. A
file that can't be read is left out, and the others are still shown.

The view keeps itself up to date: it reads its files again when they change, whether in GCS or another program, and
when lists are added to or removed from what it shows. While a library is being updated, it waits for the update to
finish. Select **Reads the library files again** on the toolbar to read them all again now. Until the Library Explorer
has finished reading your libraries, the view waits for it.

## Good to know

- Library Filter views aren't reopened when GCS restarts.
- Column widths are remembered for each type.
- Opening or closing a container or a note here opens or closes it in its list too, and the other way around.

---

[Return to Home](Home)
