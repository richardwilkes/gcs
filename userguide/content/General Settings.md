# General Settings

You can adjust how GCS behaves using the **General Settings** view.

For other settings options, see [Default Settings](Default%20Settings), [Sheet Settings](Sheet%20Settings),
[Colors and Fonts](Colors%20and%20Fonts), [Menu Keys](Menu%20Keys), and [Saved Filters](Saved%20Filters).

To open General Settings:

- Go to **Settings > General Settings** (`Ctrl+,` / `Cmd+,`).

Changes are applied immediately. You don’t need to save.

## Managing Settings

- **Reset settings to defaults:** Select **Reset** ![reset](images/icons/icn-reset.svg).
- **Import or export settings:** On the Settings toolbar, select **Menu** ![menu](images/icons/icn-menu.svg), then
  choose:
  - **Import…:** Load saved settings into GCS.
  - **Export…:** Save your current settings as a file.
  - Any `.general` settings files found in the `Settings` folder of your libraries are listed at the bottom of the
    menu, ready to import.

## Setting Options

The settings are grouped into sections.

### New Sheets

| Setting                         | Description                                                                                                                          |
|---------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| **Player Name**                 | The player name that appears on new character sheets.                                                                                |
| **Points**                      | Starting points for a new character. See [Character points](Character%20Points).                                                     |
| **Tech Level**                  | The Tech Level ([B22](B22)) given to new character sheets.                                                                           |
| **Calendar**                    | The calendar used to generate random birthdays. Gregorian is built in, and the Master Library adds others, such as Golarion.         |
| **Fill in initial description** | Randomly fills the identity and description blocks on new [character sheets](Character%20Sheet%20Overview).                          |
| **Add natural attacks**         | Automatically adds the Bite, Kick, and Punch natural attacks to new character sheets.                                                |

### Lists

| Setting                                               | Description                                                                                                                                                                                                                                                                    |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Group containers when sorting**                     | Keeps containers together ahead of other rows when a list is sorted, instead of mixing them in by value.                                                                                                                                                                       |
| **Show additional page references when space allows** | Lets the page reference column on character sheets, loot sheets, and templates show more than one reference when there is room, instead of always collapsing to the first one.                                                                                                 |
| **Lock list files when opened**                       | Opens list files locked, so that nothing changes them until you unlock them with **Lock against changes** ![lock](images/icons/icn-lock.svg) on their toolbar. A new list opens unlocked, and lists already open keep their state. See [Library Explorer](Library%20Explorer). |
| **Max Auto Column Width**                             | The widest a column can grow when GCS sizes columns to fit their contents.                                                                                                                                                                                                     |

### Workspace

| Setting                          | Description                                                                                                                                                                                                                 |
|----------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Restore arrangement on start** | Reopens the sheets, libraries, and documents from your last session, in the same arrangement, when you restart GCS.                                                                                                         |
| **Show sheets as pages**         | Opens character sheets as pages, the way they print, instead of as one continuous page. On by default. Each sheet's toolbar can still switch between the two. See [Character Sheet Overview](Character%20Sheet%20Overview). |
| **Open in separate windows**     | Opens the checked document types (Character Sheets, Character Templates, Loot Sheets, Editors, Images, Libraries, Markdown, PDFs, Settings, Sub-Editors) in windows of their own.                                           |
| **Initial scale**                | The starting zoom for lists, editors, sheets, Markdown documents, images, and PDFs. For PDFs you can also choose **Fit Width** or **Fit Page** instead of a fixed scale.                                                    |

### Updates & Library

| Setting                       | Description                                                                                                                                                                                                                                                |
|-------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Check for GCS Updates**     | How often to look for a newer version of GCS: **At launch**, **At launch and every hour**, **At launch and every day**, or **Never**. **Help > Check for GCS updates** always works. See [Installing and Updating GCS](Installing%20and%20Updating%20GCS). |
| **Check for Library Updates** | How often to look for newer versions of your libraries, with the same choices. The [Library Explorer](Library%20Explorer) update buttons always work.                                                                                                      |
| **Deep search**               | Lets the Library Explorer search field match the contents of the checked file types (for example, individual traits inside a traits library), not just file names. See [Library Explorer](Library%20Explorer).                                             |

### Interaction

| Setting                                         | Description                                                                                                                                                                                                                   |
|-------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Initial click on text field selects all**     | Highlights all text when you first click inside a field.                                                                                                                                                                      |
| **Tab stops on static text for screen readers** | While a screen reader is in use, makes labels, read-only fields, and disabled controls Tab stops so they can be read by moving the keyboard focus alone. Off by default, since it makes the Tab order slower to move through. |
| **Tooltip Delay**                               | How long to hover before a tooltip appears (0-30 seconds).                                                                                                                                                                    |
| **Tooltip Dismissal**                           | How long a tooltip remains visible (1-3600 seconds).                                                                                                                                                                          |
| **Scroll Wheel Multiplier**                     | Adjusts scroll speed.                                                                                                                                                                                                         |
| **Cursor Size**                                 | The size of the mouse cursor, in points. The cursor colors can be changed in [Colors and Fonts](Colors%20and%20Fonts).                                                                                                        |

### Advanced

| Setting                     | Description                                                                                                                                     |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| **Monitor Resolution**      | Sets resolution. If `0 ppi`, GCS uses the resolution your monitor reports. See **Monitor Resolution** below.                                    |
| **Image Export Resolution** | Resolution for exported images.                                                                                                                 |
| **Script Execution Limit**  | How long a single script may run before it is stopped. See the [Scripting Guide](Scripting%20Guide).                                            |
| **External PDF Viewer**     | Lets you use a different program to open PDFs. Keep blank to open PDFs with GCS. See **External PDF Settings** below.                           |
| **Interface Locale**        | Changes the GCS language. Leave blank to use the system default. Takes effect the next time GCS is started.                                     |
| **Settings Path**           | Where GCS saves its settings. Select the button beside it to copy the path.                                                                     |
| **Translations Path**       | Where GCS looks for translations. Select the button beside it to copy the path.                                                                 |
| **Log Path**                | Where GCS writes its log. Crash details are written to a `gcs-crash.log` file in the same folder. Select the button beside it to copy the path. |

## External PDF Settings

By default, PDFs for [Page References](Page%20References) open within GCS. You can set GCS to use your system’s PDF
viewer (for example, Adobe Acrobat).

> [!IMPORTANT]
> You must use the executable (`.exe`) file path, not a shortcut path.

To use an external PDF viewer:

1. In **External PDF Viewer**, enter the executable path.
2. Append the arguments your viewer needs, using these placeholders where the values should go:
   - `$FILE` is replaced with the path to the PDF, based on the **Page Reference Mappings** in [Page
     References](Page%20References).
   - `$PAGE` is replaced with the page number, with the mapping's page offset applied.
   - `$TEXT` is replaced with the text to highlight on the page.

> **_Example:_** For Adobe Acrobat on Windows, enter `C:\Program Files\Adobe\Acrobat DC\Acrobat\Acrobat.exe /A
> "page=$PAGE" "$FILE"`.

In most cases, you'll want to surround `$FILE` and `$TEXT` with quotes. Note that this might still fail, e.g. if the
value itself contains quotes.

To use GCS as the PDF viewer:

- Clear the **External PDF Viewer** field so it's blank.

## Monitor Resolution

This setting controls how GCS scales its interface based on your monitor’s pixel density (PPI). You usually won’t need
to adjust this unless GCS appears scaled incorrectly.

Normally, GCS detects your monitor resolution automatically. However, on some systems, especially Windows setups with
multiple monitors or custom scaling, the reported PPI can be inaccurate. When that happens, the interface or character
sheets may appear too small or too large.

If you notice incorrect scaling:

1. Use [this online calculator](https://www.sven.de/dpi/) to find your monitor’s true PPI.
2. Enter that value in **Monitor Resolution**.
3. Restart GCS to apply the change.

> [!TIP]
> **Windows:** Right-click your desktop, select **Display settings**, and check the **Display resolution** value under
> **Scale and layout**. If your **Scale** setting is higher than 100%, try reducing the **Monitor Resolution** value
> slightly in GCS.
>
> **macOS:** Go to **System Settings > Displays**, then check your **Resolution** and **Scale** options. Choose
> **Default for display** or **Scaled** and note the resolution used.
>
> **Linux:** Open your display settings (for example, **Settings > Displays** in GNOME) and confirm your **Resolution**
> and **Scale** values.

---

[Return to Home](Home)
