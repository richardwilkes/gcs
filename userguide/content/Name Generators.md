# Name Generators

When you select **Randomize** ![randomize](images/icons/icn-randomize.svg) beside the **Name** on a character sheet,
GCS builds a name with the **name generators** of the character's **ancestry**. An ancestry also decides the random
gender, age, height, weight, hair, eyes, skin, and handedness, and is chosen by the **Ancestry** [container](Containers)
among the character's traits.

The Master Library includes ancestries and name generators for humans, dwarves, elves, halflings, and others. You can
edit them, or make your own, with the editors described here, or by writing the files by hand.

## Ancestry and name generator files

- An **ancestry** is a `.ancestry` file. It holds the common options, a set of genders that can override them, and the
  list of name generators that make up a name.
- A **name generator** is a `.names` file. An ancestry refers to it by its file name, without the extension.

Both kinds of file are found in the `Settings` folder (or any folder inside it) of a library. When several libraries
hold a file with the same name, the one in your **User Library** wins, then the **Master Library**, then any other
libraries, and last the generators built into GCS. New files are saved to your User Library's `Settings/Ancestries`
folder by default, so a copy there overrides the Master Library's version without editing the Master Library.

## The ancestry editor

To create an ancestry, go to **File > New > New Ancestry**. To edit one, open it with **File > Open**, or choose it from
the **Menu** ![menu](images/icons/icn-menu.svg) on the editor's toolbar, which lists the ancestries in your libraries.
The editor's toolbar also has **Save**, **Save As…**, and **Reset**, and every change can be undone.

| Field               | Description                                                                                                                                                                                                                                                                                                                       |
|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Name**            | A descriptive name. Traits select an ancestry by its file name, so this is mostly shown when a template asks about replacing a character's ancestry, and used as the default file name when saving.                                                                                                                               |
| **Common Options**  | The options used when the chosen gender doesn't provide its own: the **Height Script**, **Weight Script**, and **Age Script** (see the [Scripting Guide](Scripting%20Guide); for example, `entity.randomHeightInInches($st)`), and the weighted lists of **Hair**, **Eyes**, **Skin**, and **Handedness** options to choose from. |
| **Name Generators** | The name generators used to produce a name. Their output is joined, in order, with a space. Select **Edit this name generator** ![edit](images/icons/icn-edit.svg) beside one to open it.                                                                                                                                         |
| **Genders**         | The genders that may be chosen at random, each with a **Weight** (its relative likelihood) and a **Name**, and any of the common options it overrides. A new ancestry starts with _Male_ and _Female_.                                                                                                                            |

In the option lists, the **Weight** of each entry is its relative likelihood; only the ratio between the weights in a
list matters. Drag entries by their grip ![grip](images/icons/icn-grip.svg) to reorder them.

## The name generator editor

To create a name generator, go to **File > New > New Name Generator**. To edit one, open its `.names` file, choose it
from the **Menu** ![menu](images/icons/icn-menu.svg) on the editor's toolbar, or select **Edit this name generator**
beside it in an ancestry. The **Sample Names** at the top show names made from the current definition, and
**Generate new sample names** ![randomize](images/icons/icn-randomize.svg) makes a fresh set, so you can check your
changes as you make them.

| Field                                                  | Description                                                                                                                                                                                                                                                                                     |
|--------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Type**                                               | How names are generated: **Simple** picks one of the training names at random; **Markov Letter** builds new names letter by letter from patterns in the training names; **Markov Run** builds them from the runs of vowels and consonants; **Compound** joins the output of several generators. |
| **Lowercase training names before use**                | Whether the result is lowercased before the next step.                                                                                                                                                                                                                                          |
| **Capitalize the first letter of each generated name** | Whether each generated name starts with an uppercase letter.                                                                                                                                                                                                                                    |
| **Depth** (Markov Letter only)                         | How many preceding letters are considered when choosing the next one, from 1 to 5. Lower values produce more random names; higher values stay closer to the training names. 0 uses the default of 3.                                                                                            |
| **Separator** (Compound only)                          | Placed between the outputs of the combined generators, such as a space.                                                                                                                                                                                                                         |
| **Generators** (Compound only)                         | The generators whose outputs are joined, in order. Each is edited in place, and compounds can nest.                                                                                                                                                                                             |
| **Built-in Training Data**                             | A built-in set of training names: American male, female, or last names, weighted by how common they are, or unweighted. Choose **None** to provide your own list.                                                                                                                               |
| **Training Names**                                     | The names the generator learns from, each with a **Weight**. Select **Import training names from a text file, one name per line** to load a list; a name may end with `: 3` to give it a weight, and duplicates are merged.                                                                     |

## File format

Both files hold JSON. An ancestry's fields are `name`, `common_options`, and `gender_options`; each options block has
`name`, `height_script`, `weight_script`, `age_script`, the `hair_options`, `eye_options`, `skin_options`, and
`handedness_options` lists of `{weight, value}` entries, and `name_generators`, a list of generator file names. The
rest of this page describes the name generator files, which you may want to write by hand for large training lists.

Files that GCS saves also start with a `file_type` field, `"ancestry"` or `"names"`, and a `version` field. You can
leave both out of a file you write by hand, but if you include `file_type`, it must match the file's extension, or GCS
refuses to load the file with a message such as _The file holds "ancestry" data, but "names" data was expected._

There are four types of name generator:

- `simple`
- `markov_letter`
- `markov_run`
- `compound`

### Common fields

Each of the name generators has some fields in common with each other, which are described here before the actual file
definitions, below.

- `type`: this must be present and its value must be one of the name generator types listed above. The older name
  `markov_chain` is still accepted for `markov_letter`.
- `no_lowered`: this field is optional, and when not present is assumed to be false. When false, the resulting name will
  be forced to lowercase before being returned. When true, it is left as-is.
- `no_first_to_upper`: this field is optional, and when not present is assumed to be false. When false, the resulting
  name will have its first letter forced to uppercase before being returned. When true, it is left as-is. This
  adjustment runs after any adjustment made for the no_lowered option.

There are three different possibilities for the training data, which are looked for in order and the first field found
is used:

- `built_in_training_data`: this indicates one of the built-in data sets should be used. This field may be any one of
  these values:
  - `american_male`: first names of American males from [baby names](http://www.ssa.gov/oact/babynames), weighted for
    how common they are.
  - `american_female`: first names of American females from [baby names](http://www.ssa.gov/oact/babynames), weighted
    for how common they are.
  - `american_last`: last names of Americans from
    [2010 surnames](https://www2.census.gov/topics/genealogy/2010surnames), weighted for how common they are.
  - `unweighted_american_male`: first names of American males from [baby names](http://www.ssa.gov/oact/babynames).
  - `unweighted_american_female`: first names of American females from [baby names](http://www.ssa.gov/oact/babynames).
  - `unweighted_american_last`: last names of Americans from
    [2010 surnames](https://www2.census.gov/topics/genealogy/2010surnames).
  - `none`: the default, which means the training data comes from one of the fields below.
- `weighted_training_data`: this is json object containing pairs of names and weights, like this:

    ```json
    {
      "john": 45,
      "richard": 22
    }
    ```

- `training_data`: this is a json array containing a list of names, like this:

    ```json
    [
      "john",
      "richard"
    ]
    ```

The examples below all use `training_data`, but could have easily been written to use `built_in_training_data` or
`weighted_training_data` instead. When GCS saves a file, it writes `training_data` if every weight is 1 and
`weighted_training_data` otherwise.

### Simple

As the name suggests, this is the simplest of the name generators.

```json
{
  "type": "simple",
  "no_lowered": false,
  "no_first_to_upper": false,
  "training_data": [
    "first",
    "second"
  ]
}
```

### Markov Letter

This option uses [Markov chains](https://en.wikipedia.org/wiki/Markov_chain) from the letters in the training data to
generate names.

```json
{
  "type": "markov_letter",
  "no_lowered": false,
  "no_first_to_upper": false,
  "depth": 3,
  "training_data": [
    "first",
    "second"
  ]
}
```

The `depth` field is optional and will default to 3 if not specified or 0. It can range from 1 to 5 and indicates how
many consecutive letters should be considered for the chain; values above 5 are treated as 5.

### Markov Run

This option uses [Markov chains](https://en.wikipedia.org/wiki/Markov_chain) from decomposing the training data into
runs of vowels and consonants to generate names.

```json
{
  "type": "markov_run",
  "no_lowered": false,
  "no_first_to_upper": false,
  "training_data": [
    "first",
    "second"
  ]
}
```

### Compound

This is the most complex of the available options. It allows you to build a name up using multiple other name
generators.

```json
{
  "type": "compound",
  "no_lowered": false,
  "no_first_to_upper": false,
  "separator": " ",
  "compound": [
    {
      "type": "simple",
      "no_lowered": false,
      "no_first_to_upper": false,
      "training_data": [
        "first",
        "second"
      ]
    },
    {
      "type": "markov_letter",
      "no_lowered": false,
      "no_first_to_upper": false,
      "depth": 3,
      "training_data": [
        "first",
        "second"
      ]
    }
  ]
}
```

The `separator` field is optional and will default to an empty string if not specified. It can be any fixed text you'd
like placed between each piece of generated text from the name generators this generator is stringing together. For
example, if you specify a separator value of "-" and you had two generators listed in the compound field which generated
"Lloyd" and "Atkinson", the result would be "Lloyd-Atkinson".

The `compound` field contains a json array of other name generators to use for the components of this generator. The
results of each will be strung together, optionally with some separator text between each.

---

[Return to Home](Home)
