# Export Templates

Export templates turn a character sheet into a file of your own design: a stat block for a forum post, a wiki page, a
file for a virtual tabletop, or a web page. GCS fills the template in with the sheet's data and writes the result
wherever you choose. The Master Library ships with several templates, and this page explains how to write your own.

## Where templates live

Templates live in the `Output Templates` folder at the top of a library. GCS reads that folder each time you open the
**File > Export** menu, and reads the template file each time you export, so a new or changed template is used right
away, without restarting GCS.

- Keep your own templates in the `Output Templates` folder of your User Library, creating the folder if it isn't
  there. The [Library Explorer](Library%20Explorer) doesn't list template files, so right-click the library and select
  **Show on Disk** to find the folder.
- The Master Library's templates (Fantasy Grounds, HTML, Text, and Obsidian Markdown) are examples to copy. Copy a
  template into your User Library before changing it, since library updates overwrite the Master Library.
- The template's extension is the exported file's extension: a `.txt` template produces a `.txt` file, a `.md`
  template a `.md` file, and so on.

## Running an export

With a character sheet active, select **File > Export**. The menu lists each library's templates under the library's
name, using the file name without its extension. Select one and GCS asks where to save the result, suggesting the
sheet's name with the template's extension. If the template has a mistake in it, an **Export failed** dialog shows the
message. [Troubleshooting](#troubleshooting) explains the common ones.

The same export runs from the command line, without opening the GCS window:

```
gcs -text "My Template.txt" "Dai Blackthorn.gcs" "Sora.gcs"
```

Each sheet's output is written beside the sheet, with the sheet's name and the template's extension, such as
`Dai Blackthorn.txt`. Only character sheets (`.gcs` files) can be exported.

## The first line picks the format

The first line of a template tells GCS how to process the rest of it. It must be one of these, with nothing else on
the line:

| First line             | What GCS does                                                                       |
|------------------------|-------------------------------------------------------------------------------------|
| `GCS Text Template v1` | Copies values into the output as they are: plain text, Markdown, XML, and the like. |
| `GCS HTML Template v1` | Escapes values for display in a web page. See [HTML templates](#html-templates).    |

The line itself isn't copied to the output. As of GCS v5.58.0, a byte order mark at the start of the file and any
spaces or tabs around the line are ignored; earlier versions required an exact match. Any other first line runs the
template through the legacy processor used before GCS v5.15. That processor is still present so that old templates
keep working, but it isn't documented here and receives little new data, so write new templates in one of the two
formats above. As of GCS v5.59.0, its primary and secondary attribute loops offer `@CURRENT`, an attribute's current
value beside `@VALUE`, and its keys for an attribute or pool threshold the sheet doesn't define write nothing rather
than a huge negative number.

Both formats use the template language of the Go programming language. The sections below cover everything a sheet
export needs. The complete reference is at <https://pkg.go.dev/text/template>, with the HTML format's escaping rules
at <https://pkg.go.dev/html/template>.

## Your first template

This walkthrough uses Dai Blackthorn, from the Master Library's `Basic Set/Characters` folder, so you can compare your
output with the output shown here.

1. Create a file named `My Template.txt` in your User Library's `Output Templates` folder, containing:

   ```
   GCS Text Template v1
   Name: {{.Name}} ({{.Points.Total}} points)
   ```

2. Open the sheet and select **File > Export > My Template**. Save the file and open it. It contains:

   ```
   Name: Dai Blackthorn (250 points)
   ```

   Everything outside `{{` and `}}` is copied as is. Inside them, `.Name` is the sheet's name and `.Points.Total` is
   its point total. The dot means "the current value", which starts out as the whole sheet. The [Sheet](#sheet)
   section lists everything it holds.

3. Add a skill list. `range` repeats its body once for each entry in a list, with the dot set to that entry:

   ```
   GCS Text Template v1
   Name: {{.Name}} ({{.Points.Total}} points)
   Skills:
   {{range .Skills}}
     {{.Description}}-{{.Level}} [{{.Points}}]
   {{end}}
   ```

   The result has the right content but is double-spaced, because the line breaks after `{{range .Skills}}` and before
   `{{end}}` are part of the repeated body:

   ```
   Name: Dai Blackthorn (250 points)
   Skills:

     Acrobatics-15 [2]

     Body Sense-16 [1]

     Climbing-18 [1]
   ...
   ```

4. A dash just inside the braces swallows the whitespace, line breaks included, next to an action: `{{-` removes it
   before the action and `-}}` after. Starting the `range` and `end` lines with `{{-` removes the line break before
   each of them, which leaves the indented text lines intact:

   ```
   GCS Text Template v1
   Name: {{.Name}} ({{.Points.Total}} points)
   Skills:
   {{- range .Skills}}
     {{.Description}}-{{.Level}} [{{.Points}}]
   {{- end}}
   ```

   ```
   Name: Dai Blackthorn (250 points)
   Skills:
     Acrobatics-15 [2]
     Body Sense-16 [1]
     Climbing-18 [1]
   ...
   ```

5. The lists include the containers that group their entries as rows of their own, with `.Type` set to `group`. Dai
   has none in his skills, but most sheets do, so limit the line to the skills themselves with `if`:

   ```
   GCS Text Template v1
   Name: {{.Name}} ({{.Points.Total}} points)
   Skills:
   {{- range .Skills}}
   {{- if eq .Type "item"}}
     {{.Description}}-{{.Level}} [{{.Points}}]
   {{- end}}
   {{- end}}
   ```

The rest of this page covers what else the language can do, and the Master Library's `Text - PC` and `HTML - PC`
templates are complete examples to learn from.

## Template language primer

### Actions and comments

Everything between `{{` and `}}` is an action, which GCS evaluates and replaces with its result. Everything else is
copied to the output unchanged. A comment produces nothing:

```
{{/* This is a comment. */}}
```

The `/*` must follow `{{` directly, with no space between them. `{{- /* ... */ -}}` is also allowed.

### The dot and the dollar sign

`.` is the current value. At the top of a template it's the sheet, so `.Name` is the sheet's name. A dot after a field
walks into it: `.Points.Total` is the `Total` field of the sheet's `Points`, and `.Attributes.Primary` is the list of
primary attributes.

`range` and `with` change what the dot refers to inside them. Inside `{{range .Skills}}`, the dot is one skill, so
`.Description` is that skill's description. `$` always refers to the sheet, so use it to reach a sheet field from
inside a loop:

```
{{range .Skills}}{{.Description}} belongs to {{$.Name}}
{{end}}
```

Inside a `define` block, `$` is whatever value the block was given. See [Reusable blocks](#reusable-blocks).

### Fields and methods

A value's fields are listed in the reference at the end of this page, each with its type. Follow the type to see what
you can do with it: `.Points` on the sheet is a [Points](#points) value, so `.Points.Unspent` works, while `.Points` on
a trait is a [Number](#number).

A [Number](#number) also has methods, such as `.StringWithSign`, which are used like fields:
`{{.Points.Total.StringWithSign}}`. A method that takes a parameter is called with the parameter after a space:
`{{.Points.Total.Add .Points.Unspent}}`.

### Conditions

```
{{if .Spells}}Spellcaster{{else if .Traits}}Just traits{{else}}Blank sheet{{end}}
```

`else if` and `else` are optional. A value counts as false when it's `false`, zero, empty text, or an empty list, so
`{{if .Spells}}` tests whether the sheet has any spells and `{{if .Notes}}` whether an item's notes field has anything
in it.

`with` is an `if` that also moves the dot to the value when it isn't empty:

```
{{with .Title}}Title: {{.}}{{else}}Untitled{{end}}
```

### Comparisons and logic

The comparison functions are `eq` (equal), `ne` (not equal), `lt` (less than), `le` (less or equal), `gt` (greater
than), and `ge` (greater or equal). The function comes first, then the two values:

```
{{if gt .SizeModifier 0}}Large{{end}}
{{if eq .Type "item"}}A skill{{else}}A group of skills{{end}}
```

`eq` alone accepts extra values, and is true when the first matches any of the others:
`{{if eq .Type "group" "ancestry" "meta_trait"}}`. Combine tests with `and`, `or`, and `not`, written the same way,
using parentheses to pass one call's result to another:

```
{{if and .Equipped (gt (numberToInt .Quantity) 1)}}Several, equipped{{end}}
{{if not (eq .Type "item")}}Container{{end}}
```

Both sides of a comparison must be the same kind of value. Comparing text with a number is an error, comparing two
pieces of text compares them alphabetically (so `15` sorts before `9`), and comparing a [Number](#number) with a plain
number silently gives the wrong answer. The next section explains how to convert between them.

### Numbers

GCS keeps most quantities (points, levels, costs, attribute values) as a [Number](#number), a fixed-point value with
four decimal places. Printing one works as you'd expect: `{{.Points.Total}}` gives `250`. Comparing or calculating with
one needs a little care:

- Convert it to a plain integer or floating-point number with `numberToInt` or `numberToFloat` before comparing it
  with a literal: `{{if gt (numberToInt .Points) 0}}`.
- Or turn the literal into a Number with `numberFrom`: `{{if eq .Points (numberFrom 1)}}`.
- Calculate with the Number's methods: `.Add`, `.Sub`, `.Mul`, `.Div`, `.Round`, `.Floor`, and the rest listed under
  [Number](#number). `{{.Points.Total.Div (numberFrom 2)}}` is half the point total.
- Format with `.Comma` (thousands separators), `.StringWithSign` (a leading `+` on values that aren't negative), or
  both: `{{.Total.CommaWithSign}}`.
- Some numeric fields are text, because they carry units or symbols: a skill's `.Level` is `15`, a weight is
  `67.266 lb`. `numberFrom` reads the number at the start of such text, so `numberFrom .Equipment.CarriedWeight` is the
  carried weight as a Number, and `numberToInt (numberFrom .Level)` is a skill's level as an integer.

> [!WARNING]
> `{{if eq .Points 1}}` is false even when the points are 1, because `eq` compares the literal with the Number's
> internal representation, which holds four decimal places. Likewise, `printf "%d"` on a Number prints that internal
> value (`2500000` for 250). Convert with `numberToInt` first, or print with `%s` or `%v`.

### Functions and pipelines

A function is called by name, followed by its arguments separated by spaces. Use parentheses to pass one call's result
to another:

```
{{join .Colleges ", "}}
{{repeat "  " .Depth}}{{.Description}}
{{upper (trim .Name)}}
```

The `|` character pipes the result of one call into the last argument of the next:

```
{{.Points.Total | printf "%s points"}}
{{.Name | upper | printf "[%s]"}}
```

`printf` formats values the way C's printf does: `%s` for text, `%d` for an integer, `%.1f` for a floating-point number
with one decimal, `%v` for anything, and `%q` for quoted text. A width pads the value, which lines columns up in a text
export: `%-30s` pads text to 30 characters on the right, `%4s` pads to 4 on the left.

```
{{printf "%-30s %3s [%s]" .Description .Level .Points}}
```

The functions available are the built-in ones below, the [Number functions](#number-functions), and the
[String functions](#string-functions). There are no others: functions from other template systems built on Go's, such
as Hugo's `where` or `add`, don't exist here.

| Built-in function                  | What it does                                                                 |
|------------------------------------|------------------------------------------------------------------------------|
| `eq`, `ne`, `lt`, `le`, `gt`, `ge` | Compare two values. See [Comparisons and logic](#comparisons-and-logic).     |
| `and`, `or`, `not`                 | Combine conditions.                                                          |
| `len`                              | The number of entries in a list: `{{len .Skills}}`.                          |
| `index`                            | One entry of a list, counting from 0, or of a map by key.                    |
| `slice`                            | Part of a list: `{{slice .Skills 0 3}}` is the first three skills.           |
| `printf`, `print`, `println`       | Format values. `print` joins its arguments and `println` adds a line break.  |
| `html`                             | Escapes `<`, `>`, `&` and quotes, for XML or HTML made by a Text template.   |
| `js`, `urlquery`                   | Escape text for use in JavaScript or in a URL query.                         |
| `call`                             | Calls a function value. Not needed for sheet exports.                        |

### Variables

A variable starts with `$`, is created with `:=`, and changed with `=`:

```
{{$sum := numberFrom 0}}
{{range .Skills}}{{$sum = $sum.Add .Points}}{{end}}
Skill points: {{$sum}}
```

A variable created inside `if`, `range`, or `with` lasts until that block's `{{end}}`. Create it before the block when
you need it afterwards, as above.

### Loops

`range` repeats its body for each entry of a list, with the dot set to the entry. An `else` section runs when the list
is empty, and `break` and `continue` work as in most programming languages:

```
{{range .Spells}}{{.Description}}; {{else}}No spells.{{end}}
{{range .Traits}}{{if eq .Type "group"}}{{continue}}{{end}}{{.Description}}; {{end}}
```

To number the entries, give `range` two variables. The index starts at 0:

```
{{range $i, $skill := .Skills}}{{$i}}: {{$skill.Description}}
{{end}}
```

A map, such as `.Attributes.PrimaryByID`, can be ranged over too, in key order. Usually the ordered lists beside the
maps, such as `.Attributes.Primary`, are what you want, since they follow the sheet's attribute order.

### Lists and lookups

`len` counts a list's entries, `index` picks one by position, and `slice` takes a run of them:

```
{{len .Skills}} skills, the first being {{(index .Skills 0).Description}}
{{range slice .Skills 0 3}}{{.Description}}; {{end}}
```

Attributes can be looked up by their ID (see [Attributes](Attributes)) in `.Attributes.PrimaryByID`,
`.Attributes.SecondaryByID`, and `.Attributes.PoolsByID`. `index` with an ID the sheet doesn't have gives an empty
value, and reading a field from that is an error, so wrap the lookup in `with` unless you're sure of the ID:

```
ST {{(index .Attributes.PrimaryByID "st").Value}}
{{with index .Attributes.PoolsByID "hp"}}HP {{.Current}}/{{.Maximum}}{{end}}
```

### Whitespace control

Template text is copied exactly, line breaks and indentation included, so a loop written one action per line
double-spaces its output (see [Your first template](#your-first-template)). A dash just inside the braces removes the
whitespace next to an action: `{{-` removes everything (spaces, tabs, and line breaks) before it, and `-}}` everything
after it. The dash must be separated from the action by a space: `{{- .Name -}}`.

Two reliable patterns:

- Begin each control line (`range`, `if`, `with`, `else`, `end`) with `{{-`, so the line break before it disappears
  while the text lines between them keep their line breaks and indentation.
- Keep a loop's body on the same line as its `range` and `end` actions, as the Master Library's `Text - PC` template
  does, so there are no extra line breaks to remove.

`-}}` removes the line break *and* the indentation at the start of the next line, which is why the first pattern puts
the dash at the start of the control line rather than at its end.

### Reusable blocks

`define` names a block of template, and `template` inserts it, passing the value the block should use as its dot:

```
GCS Text Template v1
{{- define "row"}}{{printf "%-30s" .Description}} [{{.Points}}]
{{end -}}
Traits:
{{range .Traits}}{{template "row" .}}{{end}}
Skills:
{{range .Skills}}{{template "row" .}}{{end}}
```

A `define` block produces no output where it's written, and can go anywhere in the file.

## The sheet's lists

Traits, skills, spells, equipment, and notes come as flat lists in the order they appear on the sheet, with each
container (a folder or group) as a row of its own, followed by its contents. Each row says where it sits:

- `.Type` is `item` for an entry and `group` for a container. Traits and equipment use more specific container types,
  listed under [Trait](#trait) and [Equipment](#equipment).
- `.Depth` is how many containers the row is inside: 0 at the top level, 1 inside one container, and so on.
- `.ID` identifies the row, and `.ParentID` is the ID of the container holding it, or empty at the top level.

A container's `.Points` is the total of its contents, so filter on `.Type` before adding points up. This renders the
trait list as an outline, with the containers in capitals:

```
GCS Text Template v1
Traits:
{{- range .Traits}}
{{repeat "    " .Depth}}{{if eq .Type "item"}}{{.Description}} [{{.Points}}]{{else}}{{upper .Description}}{{end}}
{{- end}}
```

```
Traits:
ADVANTAGES & PERKS
    Absolute Direction [5]
    ESP
        Danger Sense [14]
        Warp [40]
    Flexibility [5]
...
```

Other things to know about the lists:

- Disabled traits, skills, spells, equipment, and notes are left out, along with everything inside a disabled
  container.
- Skills, spells, and other rows that a trait grants to the sheet show 0 points, as they do on the sheet.
- `.Description` is the row's name as the sheet shows it, with substitutions applied and any level or specialization
  included.
- Equipment is split into `.Equipment.Carried` and `.Equipment.Other`. `.Equipped` is only ever true for carried
  equipment.
- `.MeleeWeapons` and `.RangedWeapons` hold the attacks the sheet's weapon tables show, one row per usage (such as a
  sword's swing and thrust), with the sheet's values already worked out: `.Level` is the skill level to attack with,
  and `.Damage`, `.Reach`, and the rest are the text the sheet shows. Attacks from unequipped or uncarried items are
  included only when the sheet's **Show all attacks** setting is on.
- `.Reactions` and `.ConditionalModifiers` are flat as well: their group containers are skipped, and each modifier
  names its group in `.Group`.
- Multi-line text, such as notes, uses plain line breaks.

## HTML templates

A template whose first line is `GCS HTML Template v1` is processed as HTML, and every value is escaped for the place it
lands in: a `<` in a trait's notes becomes `&lt;` in body text, a value inside an `href` attribute is URL-encoded, and
so on. That keeps a sheet from breaking the page, whatever its notes contain. Three things follow from it:

- Line breaks in multi-line text (`.Notes`, `.ModifierNotes`, a spell's `.Rituals`, a trait's `.UserDescription`, and
  a note's `.Description`) are plain line breaks, which a browser runs together. Pass such fields through `htmlLines`,
  which escapes the text and turns each line break into a `<br>` tag: `{{htmlLines .Notes}}`.
- `.EmbeddedPortraitDataURL` and `.GridTemplate` are already marked safe for their places, so
  `<img src="{{.EmbeddedPortraitDataURL}}">` and `grid-template: {{.GridTemplate}};` work as written.
- A value the escaper can't make safe for its position is replaced with `#ZgotmplZ` (in a URL, a scheme other than
  `http`, `https`, or `mailto`) or `ZgotmplZ` (a CSS value with characters it doesn't allow). Put such values in a
  position the escaper treats as text, or build the document in a Text template.

The Master Library's `HTML - PC` template is a complete example, including the portrait and the sheet's page size and
margins (`.Page`).

## Troubleshooting

A mistake in a template stops the export: **File > Export** shows an **Export failed** dialog, and the command line
prints the message and exits. The line number in the message counts from the line after the format line, so add 1 to
find the line in the file. The type names in messages correspond to the sections of the reference below:
`exportedEntity` is the [Sheet](#sheet), and the rest drop the `exported` prefix, so `exportedTrait` is a
[Trait](#trait), `exportedSkill` a [Skill](#skill), and so on. A `fixed64.Int` is a [Number](#number).

- **The template isn't in the File > Export menu.** The folder must be named `Output Templates` and sit directly
  inside the library's folder, and the file name must not start with a dot. The menu's entries are disabled until a
  character sheet is active.
- **The output starts with the format line, and the `{{...}}` actions appear in it unchanged.** The first line wasn't
  recognized, so the legacy processor ran. Check it for a typo or different capitalization. Before GCS v5.58.0, a
  trailing space or a byte order mark also kept the line from being recognized, so with an older version save the
  file as UTF-8 without a byte order mark and remove the space.
- **`function "xyz" not defined`.** Only the functions on this page exist. For a substring's position use `indexStr`,
  since `index` is the built-in list lookup.
- **`can't evaluate field Xyz in type *export.exportedTrait`.** The dot doesn't have that field. The type named is
  what the dot is at that point; check its section of the reference. Common causes are `.Name` inside a `range` over
  traits, skills, or equipment (it's `.Description` there) and reaching for a sheet field inside a loop without `$`.
- **`unexpected EOF`** means an `if`, `range`, `with`, or `define` has no `{{end}}`. **`unexpected {{end}}`** means
  there's one too many.
- **`wrong number of args for hasPrefix: want 2 got 1`.** Check the function's parameter list. `ne`, `lt`, and the
  other comparisons take exactly two values; only `eq` takes more.
- **`incompatible types for comparison`.** The two values are different kinds, such as a Number and text, or a skill's
  text `.Level` and a number. Convert one side, as described under [Numbers](#numbers).
- **A comparison with a Number is always false, or `%d` prints a huge number.** See the warning under
  [Numbers](#numbers).
- **`nil pointer evaluating *export.exportedAttribute.Value`.** An `index` lookup by ID found nothing, because the
  sheet has no attribute with that ID or it's in a different map (pools are in `.Attributes.PoolsByID`). Wrap the
  lookup in `with`.
- **`unexpected "/" in command`.** A comment has a space between `{{` and `/*`. Write `{{/* ... */}}` or
  `{{- /* ... */ -}}`.
- **`undefined variable "$x"`.** The variable was created inside an `if`, `range`, or `with` block and used after its
  `{{end}}`. Create it before the block.
- **Blank lines, or indentation that comes and goes.** See [Whitespace control](#whitespace-control).
- **`&lt;`, `&#39;`, or `&amp;` in the output, or line breaks run together.** The HTML format is escaping values,
  which is right for a web page. Use `htmlLines` for multi-line text, and the Text format for anything that isn't HTML.
- **`#ZgotmplZ` or `ZgotmplZ` in the output.** See [HTML templates](#html-templates).
- **Rows are missing.** Disabled rows and the attacks of unequipped items are left out. See
  [The sheet's lists](#the-sheets-lists).
- **The command line says "No exportable files were provided".** Only `.gcs` character sheets can be exported, not
  templates, loot sheets, or libraries.

The rest of this page is the reference: first the GCS-specific functions, then the data types, starting with the
[Sheet](#sheet) the template is given.

## Number functions

- `numberFrom` — Returns the equivalent `Number` value of `value`. For text, reads the number at the start and ignores
  the rest, so `"15 lb"` gives 15 and text that doesn't start with a number gives 0.
  - Parameter: `value` (int, float64, or string that contains a number)
- `numberToFloat` — Returns the equivalent floating-point number value of `value`
  - Parameter: `value` (`Number`)
- `numberToInt` — Returns the equivalent integer value of `value`, dropping any fraction
  - Parameter: `value` (`Number`)

## String functions

- `caselessEqual` — Returns true if `left` and `right` are the same when case isn't considered
  - Parameter: `left` (string)
  - Parameter: `right` (string)
- `contains` — Returns true if `str` contains `substr`
  - Parameter: `str` (string)
  - Parameter: `substr` (string)
- `hasPrefix` — Returns true if `str` starts with `prefix`
  - Parameter: `str` (string)
  - Parameter: `prefix` (string)
- `hasSuffix` — Returns true if `str` ends with `suffix`
  - Parameter: `str` (string)
  - Parameter: `suffix` (string)
- `htmlLines` — Returns `text` with its HTML special characters escaped and each line break turned into a `<br>` tag,
  marked as safe HTML so the HTML template doesn't escape it again. Use it for multi-line fields such as `.Notes` and
  `.ModifierNotes` in an HTML template. Added in GCS v5.45.
  - Parameter: `text` (string)
- `indexStr` — Returns an index into `str` where `substr` starts, or -1 if `substr` isn't contained in `str`.<br>
  *NOTE: This was named `index` prior to GCS v5.33, but was renamed due to conflicts with built-in functions that Go
  supplies.*
  - Parameter: `str` (string)
  - Parameter: `substr` (string)
- `join` — Returns a new string by concatenating the elements of `elems` to create a single string, placing `sep`
  between each element in the resulting string
  - Parameter: `elems`, ([]string)
  - Parameter: `sep` (string)
- `lastIndexStr` — Returns an index into `str` where the last `substr` starts, or -1 if `substr` isn't contained in
  `str`.<br>
  *NOTE: This was named `lastIndex` prior to GCS v5.33, but was renamed to be consistent with `indexStr`.*
  - Parameter: `str` (string)
  - Parameter: `substr` (string)
- `lower` — Returns the lowercase version of `str`
  - Parameter: `str` (string)
- `repeat` — Returns a new string consisting of `count` copies of `str`
  - Parameter: `str` (string)
  - Parameter: `count` (int)
- `replace` — Returns a copy of `str` with all non-overlapping instances of `old` replaced by `new`
  - Parameter: `str` (string)
  - Parameter: `old` (string)
  - Parameter: `new` (string)
- `split` — Returns a slice of sub-strings by splitting `str` into all sub-strings separated by `sep`
  - Parameter: `str` (string)
  - Parameter: `sep` (string)
- `splitN` — Returns a slice of sub-strings by splitting `str` into at most `count` sub-strings separated by `sep`
  - Parameter: `str` (string)
  - Parameter: `sep` (string)
  - Parameter: `count` (int)
- `trim` — Returns a copy of `str` with all leading and trailing white space removed
  - Parameter: `str` (string)
- `trimPrefix` — Returns a copy of `str` without the `prefix` at the beginning
  - Parameter: `str` (string)
  - Parameter: `prefix` (string)
- `trimSuffix` — Returns a copy of `str` without the `suffix` at the end
  - Parameter: `str` (string)
  - Parameter: `suffix` (string)
- `upper` — Returns the uppercase version of `str`
  - Parameter: `str` (string)

## AllEquipment

The two equipment lists and their totals. See [The sheet's lists](#the-sheets-lists).

### AllEquipment fields

- `.Carried` ([]Equipment)
- `.CarriedValue` (Number)
- `.CarriedWeight` (string)
- `.Other` ([]Equipment)
- `.OtherValue` (Number)

## Attribute

A primary or secondary attribute. See [Attributes](Attributes).

### Attribute fields

- `.CombinedName` (string) — The full name and its abbreviation, such as `Strength (ST)`
- `.FullName` (string) — The full name, such as `Strength`
- `.ID` (string) — The attribute's ID, such as `st`
- `.Name` (string) — The short name, such as `ST`
- `.Points` (Number)
- `.Value` (Number) — The value without the adjustments of the pools' current thresholds, which the points pay for
- Added in GCS v5.59:
  - `.Current` (Number) — The value with the adjustments of the pools' current thresholds, which the sheet shows. The
    same as `.Value` when none apply.

## Attributes

### Attributes fields

- `.Pools` ([]Pool)
- `.Primary` ([]Attribute)
- `.Secondary` ([]Attribute)
- Added in GCS v5.33:
  - `.PoolsByID` (map[string]Pool)
  - `.PrimaryByID` (map[string]Attribute)
  - `.SecondaryByID` (map[string]Attribute)

The lists are in the sheet's attribute order. The maps hold the same values, keyed by attribute ID, for `index`
lookups. Hidden attributes are left out.

## BodyType

### BodyType fields

- `.Locations` ([]HitLocation) — Flat, with each sub-table's locations following the location that holds it, at a
  greater `.Depth`
- `.Name` (string)

## ConditionalModifier

A reaction or conditional modifier. See [The sheet's lists](#the-sheets-lists).

### ConditionalModifier fields

- `.ID` (string)
- `.Situation` (string)
- `.Sources` ([]Source)
- `.Total` (Number)
- Added in GCS v5.52:
  - `.Group` (string) — The group the modifier was given, or an empty string. The `.Reactions` and
    `.ConditionalModifiers` lists are flat: group containers aren't listed, and each modifier names its group here
    instead.

## Encumbrance

One row of the encumbrance table.

### Encumbrance fields

- `.Dodge` (int)
- `.IsCurrent` (boolean) — True for the sheet's current encumbrance level
- `.Level` (int) — 0 for None through 4 for Extra-Heavy
- `.MaxLoad` (string)
- `.Move` (int)
- `.Name` (string)
- `.Penalty` (int) — 0 for None through -4 for Extra-Heavy

## Equipment

One row of an equipment list. See [The sheet's lists](#the-sheets-lists).

### Equipment fields

- `.Cost` (Number) — The cost of one, with modifiers applied
- `.Depth` (int)
- `.Description` (string)
- `.Equipped` (boolean)
- `.ExtendedCost` (Number) — The cost of the quantity, plus the contents of a container
- `.ExtendedWeight` (string) — The weight of the quantity, plus the contents of a container
- `.ID` (string)
- `.LegalityClass` (string)
- `.MaxUses` (int)
- `.ModifierNotes` (string)
- `.Notes` (string)
- `.PageRef` (string)
- `.ParentID` (string)
- `.Quantity` (Number)
- `.Tags` ([]string)
- `.TechLevel` (string)
- `.Type` (string) — `item`, `container` or `group`. Prior to GCS v5.52, every container was `group`.
- `.UnsatisfiedReason` (string)
- `.Uses` (int)
- `.Weight` (string) — The weight of one, with modifiers applied

## HitLocation

### HitLocation fields

- `.Depth` (int)
- `.DR` (string)
- `.Notes` (string) — The DR breakdown the sheet shows in the location's tooltip
- `.Penalty` (int)
- `.RollRange` (string)
- `.Where` (string)

## Lift

The lifting and moving things table, each entry as text with units.

### Lift fields

- `.Basic` (string)
- `.CarryOnBack` (string)
- `.OneHanded` (string)
- `.RunningShove` (string)
- `.ShiftSlightly` (string)
- `.Shove` (string)
- `.TwoHanded` (string)

## Mana

### Mana fields

- `.Cast` (string)
- `.Maintain` (string)

## Margins

### Margins fields

- `.Bottom` (string)
- `.Left` (string)
- `.Right` (string)
- `.Top` (string)

## MeleeWeapon

One attack in the melee weapons table. See [The sheet's lists](#the-sheets-lists).

### MeleeWeapon fields

- `.Block` (string)
- `.BlockParts` (WeaponBlock)
- `.Damage` (string)
- `.Description` (string)
- `.Level` (Number)
- `.Notes` (string)
- `.Parry` (string)
- `.ParryParts` (WeaponParry)
- `.Reach` (string)
- `.ReachParts` (WeaponReach)
- `.Strength` (string)
- `.StrengthParts` (WeaponStrength)
- `.Usage` (string)

## Note

One row of the notes list. See [The sheet's lists](#the-sheets-lists).

### Note fields

- `.Depth` (int)
- `.Description` (string)
- `.ID` (string)
- `.PageRef` (string)
- `.ParentID` (string)
- `.Type` (string) — `item` or `group`
- Added in GCS v5.43:
  - `.Tags` ([]string)

## Number

A special type of number that uses a fixed number of decimal digits (four) for calculations, unlike traditional
floating-point numbers. See [Numbers](#numbers) for how to compare and calculate with one.

### Number methods

- `.Abs` — Returns the absolute value of this value
- `.Add` — Returns the result of adding this number to `value`
  - Parameter: `value` (Number)
- `.Ceil` — Returns the value rounded up to the nearest whole number
- `.Comma` — Returns the string representation of this value, but with commas for values of 1000 and greater
- `.CommaWithSign` — Same as `.Comma`, but prefixes the value with a '+' if it is not negative
- `.Dec` — Returns the value decremented by 1
- `.Div` — Returns the result of dividing this number by `value`
  - Parameter: `value` (Number)
- `.Floor` — Returns the value rounded down to the nearest whole number
- `.Inc` — Returns the value incremented by 1
- `.Max` — Returns the maximum of this value or `value`
  - Parameter: `value` (Number)
- `.Min` — Returns the minimum of this value or `value`
  - Parameter: `value` (Number)
- `.Mod` — Returns the remainder after subtracting all full multiples of the passed-in `value` from this number
  - Parameter: `value` (Number)
- `.Mul` — Returns the result of multiplying this number by `value`
  - Parameter: `value` (Number)
- `.Round` — Returns the nearest integer value, rounding half away from zero
- `.String` — Returns the string representation of this value
- `.StringWithSign` — Same as `.String`, but prefixes the value with a '+' if it is not negative
- `.Sub` — Returns the result of subtracting `value` from this number
  - Parameter: `value` (Number)
- `.Trunc` — Returns the value with its fractional part removed

## Page

The sheet's page settings, as CSS lengths such as `8.5in`.

### Page fields

- `.Height` (string)
- `.Margins` (Margins)
- `.Width` (string)

## Points

### Points fields

- `.Advantages` (Number)
- `.Ancestry` (Number)
- `.Attributes` (Number)
- `.Disadvantages` (Number)
- `.Quirks` (Number)
- `.Skills` (Number)
- `.Spells` (Number)
- `.Total` (Number)
- `.Unspent` (Number)

## Pool

A point pool, such as HP or FP. See [Point Pools](Point%20pools).

### Pool fields

- `.CombinedName` (string)
- `.Current` (Number)
- `.FullName` (string)
- `.ID` (string)
- `.Maximum` (Number)
- `.Name` (string)
- `.Points` (Number)

## RangedWeapon

One attack in the ranged weapons table. See [The sheet's lists](#the-sheets-lists).

### RangedWeapon fields

- `.Accuracy` (string)
- `.AccuracyParts` (WeaponAccuracy)
- `.Bulk` (string)
- `.BulkParts` (WeaponBulk)
- `.Damage` (string)
- `.Description` (string)
- `.Level` (Number)
- `.Notes` (string)
- `.Range` (string)
- `.RangeParts` (WeaponRange)
- `.RateOfFire` (string)
- `.RateOfFireParts` (WeaponRoF)
- `.Recoil` (string)
- `.RecoilParts` (WeaponRecoil)
- `.Shots` (string)
- `.ShotsParts` (WeaponShots)
- `.Strength` (string)
- `.StrengthParts` (WeaponStrength)
- `.Usage` (string)

## Sheet

This is the value the template is given, and what `.` and `$` refer to at the top of the template. All other data is
reached through it.

### Sheet fields

- `.Age` (string)
- `.Attributes` (Attributes)
- `.Birthday` (string)
- `.BodyType` (BodyType)
- `.ConditionalModifiers` ([]ConditionalModifier)
- `.CreatedOn` (string)
- `.EmbeddedPortraitDataURL` (string, marked as a safe URL for HTML templates) — Empty when the sheet has no portrait
- `.Encumbrance` ([]Encumbrance)
- `.Equipment` (AllEquipment)
- `.Eyes` (string)
- `.Gender` (string)
- `.GridTemplate` (string, marked as safe CSS for HTML templates)
- `.Hair` (string)
- `.Handedness` (string)
- `.Height` (string)
- `.Lift` (Lift)
- `.MeleeWeapons` ([]MeleeWeapon)
- `.ModifiedOn` (string)
- `.Name` (string)
- `.Notes` ([]Note)
- `.Organization` (string)
- `.Page` (Page)
- `.Player` (string)
- `.Points` (Points)
- `.RangedWeapons` ([]RangedWeapon)
- `.Reactions` ([]ConditionalModifier)
- `.Religion` (string)
- `.SizeModifier` (int)
- `.Skills` ([]Skill)
- `.Skin` (string)
- `.Spells` ([]Spell)
- `.Swing` (string)
- `.TechLevel` (string)
- `.Thrust` (string)
- `.Title` (string)
- `.Traits` ([]Trait)
- `.Weight` (string)

## Skill

One row of the skills list: a skill, a technique, or a container of them. See
[The sheet's lists](#the-sheets-lists).

### Skill fields

- `.Depth` (int)
- `.Description` (string)
- `.Difficulty` (string) — Such as `DX/A`. Empty for a container.
- `.ID` (string)
- `.Level` (string) — The level as text, such as `15`, or `-` when the skill can't be used. Empty for a container.
- `.ModifierNotes` (string)
- `.Notes` (string)
- `.PageRef` (string)
- `.ParentID` (string)
- `.Points` (Number)
- `.RelativeLevel` (string) — Such as `DX+2`
- `.Tags` ([]string)
- `.Type` (string) — `item` or `group`
- `.UnsatisfiedReason` (string)

## Source

One contributor to a reaction or conditional modifier.

### Source fields

- `.Amount` (Number)
- `.Source` (string)

## Spell

One row of the spells list. See [The sheet's lists](#the-sheets-lists).

### Spell fields

- `.Class` (string)
- `.Colleges` ([]string)
- `.Depth` (int)
- `.Description` (string)
- `.Difficulty` (string) — Such as `IQ/H`. Empty for a container.
- `.Duration` (string)
- `.ID` (string)
- `.Level` (string) — The level as text, such as `15`, or `-` when the spell can't be cast. Empty for a container.
- `.Mana` (Mana)
- `.Notes` (string)
- `.PageRef` (string)
- `.ParentID` (string)
- `.Points` (Number)
- `.RelativeLevel` (string)
- `.Resist` (string)
- `.Rituals` (string)
- `.Tags` ([]string)
- `.TimeToCast` (string)
- `.Type` (string) — `item` or `group`
- `.UnsatisfiedReason` (string)

## Trait

One row of the traits list. See [The sheet's lists](#the-sheets-lists).

### Trait fields

- `.Depth` (int)
- `.Description` (string)
- `.ID` (string)
- `.ModifierNotes` (string)
- `.Notes` (string)
- `.PageRef` (string)
- `.ParentID` (string)
- `.Points` (Number)
- `.Tags` ([]string)
- `.Type` (string) — `item`, or the container type: `group`, `alternative_abilities`, `ancestry`, `attributes` or
  `meta_trait`
- `.UnsatisfiedReason` (string)
- `.UserDescription` (string)
- Added in GCS v5.52:
  - `.PrereqContradiction` (string) — Explains a trait whose own prerequisites are met but which is caught in a
    contradiction with other traits' prerequisites, when the "Disable traits whose prerequisites are unsatisfied"
    sheet setting is on. Empty otherwise. At most one of this and `.UnsatisfiedReason` is set.
- Added in GCS v5.43:
  - `.CR` (int) — The self-control roll target number, or 0 if there is no self-control roll.
  - `.CRFull` (string) — The full self-control roll description (e.g. "CR: 12 (Quite Often)"), or an empty string if
    there is no self-control roll.
  - `.FR` (int) — The frequency-of-appearance roll target number, or 0 if there is no frequency roll.
  - `.FRFull` (string) — The full frequency-of-appearance roll description, or an empty string if there is no
    frequency roll.
  - `.ModifierNotesNoCR` (string) — Same as `.ModifierNotes`, but without the self-control roll portion.
  - `.ModifierNotesNoFR` (string) — Same as `.ModifierNotes`, but without the frequency-of-appearance roll portion.
  - `.ModifierNotesNoRolls` (string) — Same as `.ModifierNotes`, but without either the self-control or
    frequency-of-appearance roll portions.
  - `.Level` (Number) — Will be -1 if the trait is a container or is not leveled.

## WeaponAccuracy

The component parts of a ranged weapon's accuracy.

### WeaponAccuracy fields

- `.Base` (Number)
- `.Jet` (boolean)
- `.Scope` (Number)

## WeaponBlock

The component parts of a melee weapon's block.

### WeaponBlock fields

- `.CanBlock` (boolean)
- `.Modifier` (Number)

## WeaponBulk

The component parts of a ranged weapon's bulk.

### WeaponBulk fields

- `.Giant` (Number)
- `.Normal` (Number)
- `.RetractingStock` (boolean)

## WeaponParry

The component parts of a melee weapon's parry.

### WeaponParry fields

- `.CanParry` (boolean)
- `.Fencing` (boolean)
- `.Modifier` (Number)
- `.Unbalanced` (boolean)

## WeaponRange

The component parts of a ranged weapon's range.

### WeaponRange fields

- `.HalfDamage` (Number)
- `.InMiles` (boolean)
- `.Max` (Number)
- `.Min` (Number)
- `.MusclePowered` (boolean)

## WeaponReach

The component parts of a melee weapon's reach.

### WeaponReach fields

- `.ChangeRequiresReady` (boolean)
- `.CloseCombat` (boolean)
- `.Max` (Number)
- `.Min` (Number)

## WeaponRecoil

The component parts of a ranged weapon's recoil.

### WeaponRecoil fields

- `.Shot` (Number)
- `.Slug` (Number)

## WeaponRoF

The component parts of a ranged weapon's rate of fire.

### WeaponRoF fields

- `.Jet` (boolean)
- `.Mode1` (WeaponRoFMode)
- `.Mode2` (WeaponRoFMode)

## WeaponRoFMode

The component parts of a single rate of fire mode.

### WeaponRoFMode fields

- `.FullAutoOnly` (boolean)
- `.HighCyclicControlledBursts` (boolean)
- `.SecondaryProjectiles` (Number)
- `.ShotsPerAttack` (Number)

## WeaponShots

The component parts of a ranged weapon's shots.

### WeaponShots fields

- `.Count` (Number)
- `.Duration` (Number)
- `.InChamber` (Number)
- `.ReloadTime` (Number)
- `.ReloadTimeIsPerShot` (boolean)
- `.Thrown` (boolean)

## WeaponStrength

The component parts of a weapon's minimum strength.

### WeaponStrength fields

- `.Bipod` (boolean)
- `.Min` (Number)
- `.Mounted` (boolean)
- `.MusketRest` (boolean)
- `.TwoHanded` (boolean)
- `.TwoHandedUnready` (boolean)
