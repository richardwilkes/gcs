# Markdown Guide

Use **Markdown syntax** to add structure and style to your [Notes](Notes) in GCS. This page shows the most common
formatting options with examples.

Markdown is also supported in the notes field in the **Detail Editors** for [Traits](Traits), [Skills](Skills),
[Spells](Spells), [Equipment](Equipment), and other advanced features, and in the `.md` documents kept in your
libraries. This guide is written in Markdown too.

## Reading and editing Markdown documents

Double-click a `.md` file in the [Library Explorer](Library%20Explorer) to open it. The toolbar has a **Markdown
Guide** button that opens this page, a **Scale** field, and, for files on disk, **Toggle Edit Mode**
![edit](images/icons/icn-edit.svg), which switches between the rendered document and a plain text editor. **File > New
> New Markdown File** creates a new document, starting in edit mode.

In the rendered view, dragging with the mouse scrolls the document. Use the keyboard to read and copy text:

- The arrow keys, `Home`, `End`, `Page Up`, and `Page Down` move a reading caret. Hold `Shift` to extend the
  selection.
- **Edit > Copy** and **Edit > Select All** work on the selection.
- With the caret on a link, press `Return` or `Space` to follow it. Clicking a link follows it too.

## Quick Reference

| **Element**             | Syntax                                                                          | Output                                                                               |
|-------------------------|---------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| **Headings**            | `# Heading 1`<br>`## Heading 2`<br>`### Heading 3`                              | (headings shown below)                                                               |
| **Bold**                | `**bold**` or `__bold__`                                                        | **bold text**                                                                        |
| **Italic**              | `*italic*` or `_italic_`                                                        | _italic text_                                                                        |
| **Bold+Italic**         | `***text***` or `___text___`                                                    | _**bold & italic text**_                                                             |
| **Strikethrough**       | `~~text~~`                                                                      | ~~strikethrough text~~                                                               |
| **Quote**               | `> quoted text`                                                                 | > quoted text                                                                        |
| **Alerts**              | `> [!NOTE]`<br>`> The alert text.`                                              | (alerts shown below)                                                                 |
| **Inline code**         | `` `code` ``                                                                    | `inline code`                                                                        |
| **Code block**          | ` ```code``` `                                                                  | (block shown below)                                                                  |
| **Horizontal rule**     | `---` or `***`                                                                  | ---                                                                                  |
| **Link**                | External: `[text](https://example.com)`<br>Internal: `[text](Markdown%20Guide)` | [text](https://example.com)<br>[text](Markdown%20Guide)                              |
| **Markdown Link**       | `[Markdown Guide](md:User%20Guide/Markdown%20Guide)`                            | [Markdown Guide](md:User%20Guide/Markdown%20Guide)                                   |
| **Anchor link**         | `[Tables](#tables)`                                                             | [Tables](#tables)                                                                    |
| **Page reference link** | `[Basic Set, page 45](B45)`                                                     | [Basic Set, page 45](B45)                                                            |
| **Image**               | `![alt](url)`                                                                   | ![alt](https://gurpscharactersheet.com/images/app_icon.svg)                          |
| **Sized image**         | `<img src="url" alt="alt" width="32">`                                          | <img src="https://gurpscharactersheet.com/images/app_icon.svg" alt="alt" width="32"> |
| **List (unordered)**    | `- item`<br>`- item`                                                            | - test<br>- test                                                                     |
| **List (ordered)**      | `1. item`<br>`2. item`                                                          | 1. item<br>2. item                                                                   |
| **Task list**           | `- [ ] to do`<br>`- [x] done`                                                   | (task list shown below)                                                              |
| **Table**               | See details below                                                               | (table shown below)                                                                  |
| **Escape Markdown**     | `\**escape\**`                                                                  | \*\*escape\*\*                                                                       |

## Headings

Headings create titles and organize your content. Use `#` symbols before the text. The more `#` you use, the smaller
the heading.

```
# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6 (smallest)
```

# Heading 1

## Heading 2

### Heading 3

#### Heading 4

##### Heading 5

###### Heading 6

You can also create **Heading 1** or **Heading 2** with underlines:

```
Heading 1
=========

Heading 2
---------
```

Heading 1
=========

Heading 2
---------

---

## Paragraphs

Start a new paragraph by leaving a blank line between blocks of text.

```
This is one paragraph.

This is a second paragraph.
```

This is one paragraph.

This is a second paragraph.

---

## Line Breaks

End a line with two spaces or use `<br>` for a manual break.

```
First line (with two spaces)
Second line

First line<br>
Second line
```

First line (with two spaces)
Second line

First line<br>
Second line

---

## Styling Text

- _Italics_: wrap text in one `*` or `_`
- **Bold**: wrap text in two `*` or `_`
- _**Bold & Italic**_: wrap text in three `*` or `_`
- ~~Strikethrough~~: wrap text in two `~`

```
*Italic* or _Italic_
**Bold** or __Bold__
***Bold & Italic*** or ___Bold & Italic___
~~Strikethrough~~
```

---

## Block Quotes

Use `>` to mark quoted text. Add more `>` for nested quotes.

```
> a quoted line
> > a nested quote
```

> a quoted line
>
> > a nested quote

---

## Alerts

Use alerts to emphasize critical information. Alerts are displayed with distinctive colors and icons to indicate the
significance of the content.

```

> [!NOTE]
> Useful information the reader should know, even when skimming.

> [!TIP]
> Helpful advice for doing things better or more easily.

> [!IMPORTANT]
> Key information needed to achieve a goal.

> [!WARNING]
> Urgent info that needs immediate attention to avoid problems.

> [!CAUTION]
> Advises about risks or negative outcomes of certain actions.

```

> [!NOTE]
> Useful information the reader should know, even when skimming.

> [!TIP]
> Helpful advice for doing things better or more easily.

> [!IMPORTANT]
> Key information needed to achieve a goal.

> [!WARNING]
> Urgent info that needs immediate attention to avoid problems.

> [!CAUTION]
> Advises about risks or negative outcomes of certain actions.

---

## Code

- **Inline code**: wrap text in single backticks.
- **Code block**: wrap text in triple backticks.

````
This is `inline code`.

    ```
    This is a code block with
        its formatting preserved as typed.
    ```
````

This is `inline code`.

```
    This is a code block with
        its formatting preserved as typed.
```

Inline code is shown exactly as you type it: a backslash or an entity such as `&amp;` inside the backticks appears as
written rather than being interpreted.

---

## Horizontal Rules

Use three or more `*`, `-`, or `_` on a line by themselves:

```
***
---
___
```

---

---

---

---

## Links

Plain web addresses that start with `https://`, `http://`, or `www.` are auto-detected and shown as you typed them:
https://gurpscharactersheet.com

To add custom text, use `[text](URL)`:

```
[Visit GCS](https://gurpscharactersheet.com)
```

[Visit GCS](https://gurpscharactersheet.com)

You can also add tooltips if you hover over the link:

```
[Visit GCS](https://gurpscharactersheet.com "GURPS Character Sheet")
```

[Visit GCS](https://gurpscharactersheet.com "GURPS Character Sheet")

The text of a link can be styled, and can hold inline code or an [image](#images):

```
[Visit **GCS** _now_](https://gurpscharactersheet.com)
```

[Visit **GCS** _now_](https://gurpscharactersheet.com)

A link too long to fit on one line wraps onto the next, as the rest of the text does.

### Relative paths

You can link to files in the same library using **relative paths**. This makes your notes easier to share because
they don’t depend on absolute locations.

```
[Markdown Guide](./Markdown%20Guide)
```

[Markdown Guide](./Markdown%20Guide)

**Rules for relative paths:**

- Use **forward slashes `/`** (not Windows `\`).
- Replace spaces with `%20`.
- The markdown extension (`.md`) can be omitted in your links.
- `./` means “this folder.”
- `../` means “go up one folder.”
- `../../` means "go up two folders."

> [!IMPORTANT]
> If a file name or folder has spaces, always replace them with `%20` in your link.

### GCS-specific links

In addition to web links, GCS supports special link types:

**Markdown file links (`md:`):**
Use this to reference a Markdown file in any library, wherever the document with the link lives. GCS looks for the
path inside each library's `Markdown` folder first, and then inside the library itself, checking your User Library
before the Master Library.

A link that starts with `md:User%20Guide/` always opens a page of the User Guide built into GCS, whatever your
libraries hold. Its folder and page names are matched ignoring case.

```
[Markdown Guide](md:User%20Guide/Markdown%20Guide)
```

[Markdown Guide](md:User%20Guide/Markdown%20Guide)

**Page references (like `B45`):**
Use this shorthand to link directly to a page in a GURPS book. See [Page References](Page%20References) for more
details.

```
[Basic Set, page 45](B45)
```

[Basic Set, page 45](B45)

### Anchors

Every heading gets an anchor made from its text: lowercased, with spaces turned into `-` and punctuation dropped, so
_## Horizontal Rules_ becomes `#horizontal-rules`. Link to one with `#` to scroll it to the top of the view:

```
[Jump to the tables section](#tables)
```

[Jump to the tables section](#tables)

To link to a heading in another document, put the anchor on the end of a relative link, such as
`[x](Other%20Page#section)`, or of an `md:` link:

```
[Scripting globals](md:User%20Guide/Scripting%20Guide#globals)
```

You can give a heading an anchor of your own by adding `{#my-id}` to the end of it.

---

## Images

Start with `!`, then use the same format as links:

```
![Alt text](https://gurpscharactersheet.com/images/app_icon.svg)
```

![Alt text](https://gurpscharactersheet.com/images/app_icon.svg)

You can also use relative paths for images (see links above):

```
![Logo](../images/logo.gif)
```

> [!TIP]
> Use relative paths for images stored in your library so they don’t break if you move the file.

To make an image a link, put the image in place of the link's text:

```
[![GCS](https://gurpscharactersheet.com/images/app_icon.svg)](https://gurpscharactersheet.com)
```

[![GCS](https://gurpscharactersheet.com/images/app_icon.svg)](https://gurpscharactersheet.com)

An image wider than the space it's shown in, including in a list or a quote, is scaled down to fit.

### Sizing images

An image is shown at its natural size. To show it at another size, write it as an HTML `<img>` tag instead, with the
size in pixels in `width`, `height`, or both:

```
<img src="https://gurpscharactersheet.com/images/app_icon.svg" alt="GCS" width="32">
```

<img src="https://gurpscharactersheet.com/images/app_icon.svg" alt="GCS" width="32">

- `src`, `alt`, and `title` work as the URL, alt text, and title of `![alt](url "title")` do.
- With only one of `width` and `height`, the other keeps the image's proportions. Sizes are plain numbers, such as
  `32`; percentages and other units are ignored. An image given a size wider than the space it's shown in is still
  scaled down to fit.
- A tag can go anywhere an image can: within text, in a table cell, a list item, or a heading, or as the text of a
  link. Tags alone on their own lines, with a blank line before and after, are shown as a paragraph.

---

## Lists

### Unordered lists

Use `-`, `*`, or `+`:

```
- Barbarian
- Knight
- Mage
```

- Barbarian
- Knight
- Mage

### Ordered lists

Use numbers:

```
1. First
2. Second
3. Third
```

1. First
2. Second
3. Third

You can start numbering anywhere, and GCS will auto-increment.

### Nested lists

Indent items to nest them (and mix and match ordered and unordered lists):

```
1. First list item
   - First nested list item
     - Second nested list item
   - More...
2. Another...
```

1. First list item
   - First nested list item
     - Second nested list item
   - More...
2. Another...

Each level of nesting in an unordered list has its own bullet: a disc, then a circle, then a square.

### Spacing between items

A list whose items are separated by blank lines is spaced out like paragraphs. Without the blank lines, the items sit
close together:

```
- Spaced out

- Like paragraphs
```

- Spaced out

- Like paragraphs

### Task lists

Start a list item with `[ ]` for a task still to do, or `[x]` for one that's done, followed by a space:

```
- [x] Pick an ancestry
- [ ] Buy equipment
```

- [x] Pick an ancestry
- [ ] Buy equipment

GCS shows a check box in place of the item's bullet. The check boxes only show the state written in the text; selecting
one doesn't change it.

---

## Tables

Use pipes `|` and hyphens `-` to make tables:

```
| First  | Second   |
|--------|----------|
| Line 1 | Column 2 |
| Line 2 | Column 2 |
```

| First  | Second   |
|--------|----------|
| Line 1 | Column 2 |
| Line 2 | Column 2 |

### Alignment

Use `:` for alignment in header rows:

```
| Left | Right | Center |
|:-----|------:|:------:|
| aa   | bb    | cc     |
```

| Left | Right | Center |
|:-----|------:|:------:|
| aa   |    bb |   cc   |

Note that tables don't have to have consistent column sizing:

```
| Left | Right | Centered |
|:---|---:|:---:|
| aa | bb | cc |
| aaaa | bbbb | cccc |
```

| Left | Right | Centered |
|:-----|------:|:--------:|
| aa   |    bb |    cc    |
| aaaa |  bbbb |   cccc   |

A table too wide for the space it's shown in has its widest columns narrowed, with the text in their cells wrapped to
fit.

---

## Escaping Markdown

Use a backslash `\` before a character to prevent formatting.

```
These asterisks \*will not\* make the text bold.
```

These asterisks \*will not\* make the text bold.

---

## What isn't supported

GCS renders the common Markdown described on this page, plus the GitHub-style tables, strikethrough, task lists, and
alerts. It doesn't render:

- HTML, apart from `<br>` for a line break, `<hr>` for a horizontal rule, and `<img>` for an image. Other tags are
  dropped.
- Footnotes and definition lists.

---

[Return to Home](Home)
