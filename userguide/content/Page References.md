# Page References

GCS supports **page references for GURPS books** to help you connect features in the tool to the corresponding GURPS
rules and materials.

You can use page references to:

- Open a PDF to the specific page
- Look up a page in a physical book

## Book Abbreviations

For a list of official book abbreviations used in GCS, visit [GCS Page
References](https://gurpscharactersheet.com/page_references.html).

Page references use a combination of the book abbreviation (the **key**) and the page and follow this format:
`key:page`. The colon can be left out when the key ends in a letter and the page is a number.

For example, `B16` references _GURPS Basic Set: Characters_, page 16.

> [!IMPORTANT]
> You must include the colon:
>
> - If the key ends in a number. For example, `DF2:25` references _Dungeon Fantasy 2: Dungeons_, page 25.
> - If the page isn't a number, such as the roman numerals of a book's front matter. For example, `B:iv` references
>   page _iv_ of the Basic Set.

A page reference field can hold several references separated by commas, such as `B35, P45`. Page reference columns on
character sheets show as many of them as fit, each clickable; when they don't all fit, the last one shown ends with
`+` and a tooltip lists them all. You can turn this off with **Show additional page references when space allows** in
[General Settings](General%20Settings).

References of `B338` and higher refer to _Basic Set: Campaigns_, the second volume of the Basic Set, and automatically
use the `BX` key. If you keep the two volumes as separate PDFs, map `B` and `BX` to each; if you have a single combined
PDF (or the single-volume _Basic Set Revised_), map both keys to it.

A page reference field can also hold a web link (`https://…`) or a link to a Markdown file in a library, or a page of
this guide (`md:User%20Guide/Home`). See [Markdown Guide](Markdown%20Guide).

## Opening a page reference

Once a PDF is mapped (see below), you can open it directly from a page reference:

- Select the page reference in a list on your sheet or in a library.
- Or select the row and go to **Item > Open Page Reference** (`Ctrl+G` / `Cmd+G`), which opens the first reference of
  each selected row, or **Open Each Page Reference** (`Ctrl+Shift+G` / `Cmd+Shift+G`), which opens all of them.
- In an item's editor, select the bookmark ![page reference](images/icons/icn-pageReference.svg) beside the **Page
  Reference** label.

GCS opens the PDF to the page and searches it for the item's name, or for the text in the item's **Page Highlight**
field if there is one, so the entry is highlighted.

> [!NOTE]
> If your PDF opens to the wrong page, you may need to adjust the **page offset** under **Settings > Page Reference
> Mappings**. GCS uses the page labels printed in the PDF, so most books need no offset, but some PDFs number their
> pages differently from the book.

## Mapping a PDF

If you own a PDF version of a book, you can map it to GCS and quickly open to the specific page from a page reference.

The first time you select a page reference whose key has no mapping, GCS asks whether to create one, and names the PDF
the key is normally mapped to.

To map a PDF to a page reference key:

1. Select an unmapped page reference in GCS.
2. Select **Yes** to create a mapping. Note the name of the suggested book.
3. In your file explorer, browse to the location of the PDF, select it, then select **Open**.

GCS will remember this mapping and use it automatically in the future for any page reference with that key. Select
**No** to skip the key for now, or **Cancel** to stop opening the remaining references.

> [!TIP]
> You can map **any PDF** - such as house rules or custom supplements - to a page reference key. This is useful for
> custom content you’ve added to your library. Just make sure your custom key doesn’t conflict with one of the book
> abbreviations listed on the website.

## Page Reference Mappings

To manage existing PDF mappings:

- Go to **Settings > Page Reference Mappings**.

The view lists each mapping as a row with its **Key**, **Page Offset**, and **PDF**. From there, you can:

- Adjust the **Page Offset** to fix page numbering issues. Use a positive offset if the PDF has extra pages at the
  beginning, or a negative offset if it starts later.
- Select the PDF's path to choose a **different PDF** for the key. A mapping whose PDF can no longer be read is flagged
  with a warning, and GCS asks for a new PDF the next time the key is used.
- Select **Remove** ![delete](images/icons/icn-delete.svg) beside a mapping to delete it.
- Select **Reset** ![reset](images/icons/icn-reset.svg) to remove all mappings, or use **Menu**
  ![menu](images/icons/icn-menu.svg) to import or export them as a `.refs` file.

> [!NOTE]
> By default, PDFs open inside GCS. If you prefer to use your system's PDF viewer instead, you can change this under
> [General Settings](General%20Settings).

## The PDF viewer

PDFs opened in GCS scroll continuously through the document. The toolbar has:

- **Scale** and auto-scaling (**No Auto-Scaling**, **Fit Width**, or **Fit Page**). The defaults are set in [General
  Settings](General%20Settings).
- A **Page** field that shows and accepts the page labels printed in the book, including ones such as `iv`, plus
  **Back** and **Forward** through the pages you've visited and **First Page**, **Previous Page**, **Next Page**, and
  **Last Page** buttons.
- **Toggle the Sidebar**, which shows the PDF's table of contents when it has one.
- A **Page Search** field that highlights matches on the pages as you type.

Drag to select text, which you can copy with **Edit > Copy**. Hold `Alt` (`Option` on macOS) while dragging, or drag
with the right mouse button, to pan the page instead. Links in the PDF can be clicked.

---

[Return to Home](Home)
