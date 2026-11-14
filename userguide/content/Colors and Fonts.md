# Colors and Fonts

You can customize the appearance of GCS and your [character sheet](Character%20Sheet%20Overview) with **colors** and
**fonts**. This allows you to highlight important sections, match your group’s style, or just make your sheet easier
to read.

For other settings options, see [General Settings](General%20Settings).

To open the **Colors** and **Fonts** settings:

- **Colors:** Go to **Settings > Colors.**
- **Fonts:** Go to **Settings > Fonts**.

## Managing Settings

You can manage color and font settings the same way as other settings:

- **Reset settings to defaults:** Select **Reset** ![reset](images/icons/icn-reset.svg).
- **Import or export settings:** On the Colors or Fonts toolbars, select **Menu** ![menu](images/icons/icn-menu.svg),
  then choose:
  - **Import…:** Load saved settings into GCS.
  - **Export…:** Save your current settings as a file.
  - Any `.colors` or `.fonts` files found in the `Settings` folder of your libraries are listed at the bottom of the
    menu, ready to import.

## Colors

You can change the colors GCS uses for its interface and for your character sheet.

### Color mode

At the top of the view, choose how GCS uses **light** and **dark** colors:

- **Light:** Always use light backgrounds and text.
- **Dark:** Always use dark backgrounds and text.
- **Automatic:** Follow your computer’s system preference.

The title bar of each window follows the color mode. On macOS the menus and the open and save dialogs follow it too.

### Theme Colors

Theme colors define the look of the GCS interface. For each, you can set separate values for **light mode** and **dark
mode** in the **Light** and **Dark** columns.

| Theme Color           | Description                                                                                                         |
|-----------------------|---------------------------------------------------------------------------------------------------------------------|
| **Surface**           | The background color of most windows.                                                                               |
| **Header**            | Background of headers and section titles.                                                                           |
| **Banding**           | Alternating row backgrounds for readability.                                                                        |
| **Focus**             | Highlight color for selected items and the control with the keyboard focus.                                         |
| **Tooltip**           | Background of tooltips.                                                                                             |
| **Error**             | Highlight for errors, such as an unmet prerequisite.                                                                |
| **Warning**           | Highlight for warnings.                                                                                             |
| **Needs Attention**   | Marks what still needs a decision, such as choices left to make in the template picker.                             |
| **Success**           | Marks a prerequisite the character meets in the [Prerequisites](Prerequisites) editor and a script that passed.     |
| **Failure**           | Marks a prerequisite the character doesn't meet and a script that failed.                                           |
| **Cursor Foreground** | The color of the mouse cursor. Its size is set in [General Settings](General%20Settings).                           |
| **Cursor Background** | The outline color of the mouse cursor.                                                                              |
| **Grouping 1**        | The color of the **All of** pill and rail in the prerequisites and saved filter editors.                            |
| **Grouping 2**        | The color of the **Any of** pill and rail in the prerequisites and saved filter editors.                            |

### Sheet Block Tints

You can also apply **tints** to specific blocks on the character sheet (for example, [Traits](Traits), [Skills](Skills),
or [Spells](Spells)).

- Each block can have its own color, again with separate light and dark values.
- **Alpha** (transparency) must be greater than `0` to enable the tint. An alpha of `0` turns it off.
- Higher Alpha numbers **do not** result in lighter or darker colors.

Examples:

- Light blue tint for **Skills** to make them stand out.
- Grey tint for **Equipment** to separate gear from abilities.

### Choosing a color

When you select a color box, the **color picker** appears.

You can set colors in several ways:

- By adjusting **Red, Green, Blue, Alpha** values (0-255 or %).
- Using **Hue, Saturation, Brightness** sliders.
- Typing a **CSS color** (a name such as `red`, or `#ff0000`, `rgba(255,0,0,0.5)`, or `hsl(0,100%,50%)`).

The preview panel shows both the **new color** and the **original color**.

Next to every color option is a **reset button** ![reset](images/icons/icn-reset.svg). Select this to return that color
to the GCS default.

## Fonts

The **Fonts** settings let you control the font family, size, weight, and style used throughout GCS.

Each kind of text (labels, fields, notes, etc.) can have its own font settings.

### Font uses

| Field                                        | Description                                                                              |
|----------------------------------------------|------------------------------------------------------------------------------------------|
| **System / System (Emphasized)**             | Fonts used by GCS interface text.                                                        |
| **Label**                                    | Text labels in editors and settings.                                                     |
| **Field / Secondary Fields**                 | Values entered by the user in the GCS interface.                                         |
| **Keyboard**                                 | Font for keyboard shortcut text, such as in the Menu Keys settings.                      |
| **Page Primary / Secondary Fields**          | Values on your character sheet (like attributes, skill levels, and notes).               |
| **Page Primary / Secondary Labels**          | Labels for the blocks and columns on your sheet.                                         |
| **Page Primary / Secondary Footer**          | Fonts for the page footer on your sheet.                                                 |
| **Base Markdown**                            | Font for rendered [Markdown](Markdown%20Guide) text in notes and documents.              |
| **Monospaced**                               | Used for scripts and other fixed-width text.                                             |

### Font options

Each font entry lets you choose:

- **Size** (in points, e.g., `10`).
- **Font Family** (e.g., _Roboto_, _DejaVu Sans Mono_). Depends on what fonts are installed on your system.
- **Weight**: from _Thin_ to _Extra-Black_, including _Light_, _Regular_, _Medium_, _Bold_, and _Black_.
- **Slant**: _Upright_, _Italic_, or _Oblique_.
- **Spacing**: from _Ultra-Condensed_ to _Ultra-Expanded_. _Standard_ is the default.

Next to every font option is a **reset button** ![reset](images/icons/icn-reset.svg). Select this to return that font to
the GCS default.

> [!IMPORTANT]
> Changing fonts usually requires restarting GCS to display content correctly.

---

[Return to Home](Home)
