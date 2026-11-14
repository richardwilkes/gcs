# Saved Filters

A **saved filter** is a named set of conditions that narrows a library list to the items that match it. For example,
_Cheap Advantages_ could show only the traits tagged `Advantage` that cost 10 points or less.

Saved filters belong to a list type (traits, trait modifiers, skills, spells, equipment, equipment modifiers or
notes), and every library list of that type offers them. They are stored in your GCS settings, not in your libraries,
so they are available whichever library file you open.

## Using a saved filter

Each library list has a saved filter dropdown at the right end of its toolbar.

<img src="images/screenshots/scrn-gcs-savedFiltersPopup.png" alt="Saved filter dropdown" width="630">

- Select a filter to show only the items that match it.
- Select **No Filter** to show every item again.
- Select **New Filter…**, **Edit Filter…** or **Delete Filter…** to manage filters without leaving the list. Edit and
  delete act on the filter currently in use.

While a filter is in use, a striped banner runs across the top of the list. Containers that hold a matching item stay
visible, dimmed, so you can see where the item sits. The **Quick Filter** field and the **Any Tag** dropdown (see
[Tags](Tags)) still work, and narrow the results further.

<img src="images/screenshots/scrn-gcs-savedFiltersFiltered.png" alt="A list with a saved filter in use" width="1392">

> [!NOTE]
> Adding, removing and rearranging items is disabled while a list is filtered. Select **No Filter** and clear the Quick
> Filter to make those changes.

## Building a filter

The filter editor describes each filter it holds as a plain sentence, such as _Must have tags where at least one is
Advantage_.

<img src="images/screenshots/scrn-gcs-savedFilterEditor.png" alt="The filter editor" width="732">

- **Name:** Each filter needs a name that no other filter of the same list type uses, ignoring case.
- **Filters:** Select **Add** ![add](images/icons/icn-add.svg) to add a filter or a group. When the editor is empty,
  select its placeholder to add a filter straight away. Select a filter's sentence to change it: choose the field to
  test, how to compare it (see [Operators](Operators)), and the value. An **is** or **is not** condition on a field
  with a known set of values, such as the tags, colleges, tech level, or legality class, lists the values found in the
  list below the field as you type; press `Down` to move through them and `Return` to fill one in.
- **Fields:** A filter can test an item's name, notes, tags, page reference, points, level, and the fields particular
  to its type, such as a skill's difficulty and specializations, a spell's college and class, equipment's quantity,
  value, weight, and contents, whether a container is a container or a group, and the names, notes, and tags of a
  trait's or equipment's modifiers. Modifier conditions test enabled modifiers only unless you add the **+ disabled
  modifiers** chip.
- **Groups:** The pill at the top of a group sets how it combines what it holds:
  - **All of:** everything in the group must match.
  - **Any of:** at least one thing in the group must match.
  - **None of:** nothing in the group may match.
  - **Not all of:** at least one thing in the group must not match.
- **Order:** Drag a row by its grip to move it.

**OK** stays unavailable until the filter has a name and at least one filter in it. A message beside the buttons says
when the name is already in use or there is nothing in the filter yet.

## Managing saved filters

To see and manage every saved filter in one place:

- Go to **Settings > Saved Filters…**.

The filters are listed under their list type, with a summary of each filter's conditions.

<img src="images/screenshots/scrn-gcs-savedFilters.png" alt="The Saved Filters settings" width="952">

Select one or more filters, then use the toolbar. Selecting a list type selects all of its filters. Right-click a
filter for the same commands.

- **New Filter** ![add](images/icons/icn-add.svg): Choose a list type, then build the filter in the editor.
- **Edit Filter** ![edit](images/icons/icn-edit.svg): Open the selected filter in the editor. Double-clicking a filter
  does the same.
- **Duplicate** ![copy](images/icons/icn-copy.svg): Copy each selected filter. A copy is named after its source, such
  as _Social (2)_.
- **Merge** ![merge](images/icons/icn-merge.svg): Combine two or more filters of the same list type into a new one.
- **Delete** ![delete](images/icons/icn-delete.svg): Delete the selected filters. The Delete key does the same.

Every change can be undone with **Edit > Undo**, including a change made from a list's saved filter dropdown while the
Saved Filters settings are open.

### Merging filters

Merging creates a new filter that combines the selected filters. The originals are kept.

<img src="images/screenshots/scrn-gcs-savedFiltersMerge.png" alt="Merging two filters" width="335">

- **Name:** The name for the new filter.
- **Match:** Choose **All of** to require every original filter to match, or **Any of** to require at least one.

You can open the merged filter in the editor afterwards to adjust it.

## Sharing saved filters

Saved filters can be exported to a `.filters` file and imported from one, to share them with your group or move them
to another computer. Select **Menu** ![menu](images/icons/icn-menu.svg) on the Saved Filters toolbar:

- **Import…:** Choose a `.filters` file to import from.
- **Export All…:** Save every saved filter to a file.
- **Export Selected…:** Save only the selected filters to a file.

Before anything is imported, GCS lists the filters in the file by list type so you can choose which ones to keep.
Filters with nothing in them are left out.

<img src="images/screenshots/scrn-gcs-savedFiltersImport.png" alt="Choosing filters to import" width="620">

- A filter you already have, with the same name and the same conditions, is left unchecked.
- A filter whose name is already in use, with different conditions, is imported with a number added to its name, such
  as _Shields (2)_.

> [!TIP]
> Put a `.filters` file in the `Settings` folder of any library, such as your User Library, and it appears at the
> bottom of the Saved Filters menu, ready to import without browsing for it.

---

[Return to Home](Home)
