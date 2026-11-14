# Scripting Guide

> Available in GCS v5.36.0 and later.

Some fields, such as notes, support using JavaScript to resolve their content. Versions of GCS prior to v5.36.0 used a
simple expression evaluator instead. Expressions from those old data files are automatically transformed into the
equivalent JavaScript.

## Where scripts can be used

Fields that are scriptable have a Scripting Guide ![scripting guide](images/icons/icn-scriptingGuide.svg) button next
to them to bring up this documentation. They come in two kinds:

- Fields that resolve to a **number**, such as an attribute's base value, a pool threshold, a trait's maximum level,
  an equipment value or weight, or an attack's damage, may contain either a number or a script. The whole field is run
  as a script when it isn't a plain number, so no tags are needed: `Math.floor($st / 2)`.
- Fields that resolve to **text**, such as notes, must have the scripting portion(s) wrapped in tags, like this:

```
My ST + DX is <script>$st + $dx</script> and my name is <script>entity.name</script>.
```

The script's value is the value of its last expression, or whatever a top-level `return` statement returns. A script
that uses `return` outside any function is treated as a function body, so its result is `undefined` if it reaches the
end without returning.

A [prerequisite](Prerequisites) can also be a script. There, the script's last value decides: `true` or empty text
means the requirement is met, `false` means it isn't, and any other text means it isn't, with that text shown as the
reason.

## Globals

When your JavaScript code is called, the following globals are available for you to use:

- [console](Scripting/Console)
- [dice](Scripting/Dice)
- [entity](Scripting/Entity): The character sheet. Always defined, but check its `exists` property before using the
  rest, since a script in a library or template has no character to read.
- [Math](Scripting/Math%20Extensions) extensions
- [measure](Scripting/Measure)
- `self`: The object that the script is attached to. It isn't defined at all for scripts that have no owner, such as
  the height, weight, and age scripts of an ancestry, so test for it with `typeof self !== "undefined"`. Possible
  types:
  - [Attribute](Scripting/Attribute), for an attribute's base value and its pool thresholds
  - [Equipment](Scripting/Equipment)
  - [Equipment Modifier](Scripting/Equipment%20Modifier)
  - [Note](Scripting/Note)
  - [Skill](Scripting/Skill)
  - [Spell](Scripting/Spell)
  - [Trait](Scripting/Trait)
  - [Trait Modifier](Scripting/Trait%20Modifier)
  - [Weapon](Scripting/Weapon), for an attack's usage notes and damage
- [Top-Level Functions](Scripting/Top-Level%20Functions)
  - formatNum
  - iff
  - signedValue

In addition, when the script runs for a character sheet, any attributes you've defined for the sheet are available as
globals using their ID prefixed with a dollar sign, e.g. `$st` refers to the Strength attribute. Used as a number,
`$st` gives the attribute's value without the adjustments of the pools' current thresholds (a pool's maximum), which
is what point costs and attributes such as HP are calculated from, while `$st.current` gives the value with them (the
points a pool has left). See [Attribute](Scripting/Attribute).

## Rules and limits

- Scripts run in strict mode. Variables you declare stay local to the script, and the globals above can't be replaced.
- Disabled items, and equipment with a quantity of 0, are invisible to scripts: they are left out of the lists and
  searches the [entity](Scripting/Entity) provides.
- A script that runs longer than the **Script Execution Limit** in [General Settings](General%20Settings) (0.05
  seconds by default) is stopped, and the field reports that the script timed out.
- A script whose value depends on itself, directly or through other fields, is stopped after 20 levels and reports a
  possible circular reference.
- An error thrown by a script becomes the field's text, so mistakes are easy to spot.

---

***Last updated for v5.59.0***
