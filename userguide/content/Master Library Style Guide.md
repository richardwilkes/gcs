# Master Library Style Guide

## Overview

This page sets out some stylistic guidelines on the master library files.  Additions to the master library should follow these, however they are only guidelines and usability is the primary concern.

Be aware of copyright restrictions in entering data to the library.  Game aid notes and data are generally permitted, such as the name of a trait and a note of its effect on attributes.  Descriptive text or rules on how the game works would generally not be permitted.

## Traits

## Skills

## Spells

## General template notes
* Fill in the Template block at the top: the Name, the Target Points the template is meant to cost, and a Page Reference to where it is published.
* Have all elements of the template contained within a suitably named container with a suitable page reference set.  Where containers are required for multiple areas (traits, skills, spells) ensure they all have page references and the same name.
* Where there are a selection of traits, skills or spells, place them into a choice container (Insert > Trait > New Trait Choice, and so on) whose Choices row says what the user should select.  The choice shows its rule, such as "Pick 1", on the template, and the picker enforces it when the template is applied.
* Where a selection of advantages are offered with an instruction to choose a set number of points, set this as "Points / is at most".  Some traits have variable costs (for example "Ally") which show up in the template options as a range until their modifier choice is made, so the picker can only enforce the final point value once that choice is answered.
* Similarly where a selection of disadvantages are offered with an instruction to choose a set number of (negative) points, set this as "Points / is at least".
* For skills, templates tend to have the instruction to choose a set number with each skill in the selection at the same point cost.  As such can set these to be "Count / is".
* Where a skill option is to spend a certain number of points on another skill, this can be implemented with a dummy skill and pre-requisites.  Add a skill suitably named to tell the user how many points to spend, and on what set of skills.  Give this a point cost equal to the amount to be spent.  Add a pre-requisite that will always fail, such as the character doesn't have a skill with the name equal to the name of the dummy skill.  The result is that there will be a suitably costed point holder on the character sheet with a pre-requisite failed flag alerting the user to add the points to the relevant other skill and delete the dummy skill.
* Can do similar where an option is to increase the level of an existing trait where you may not want to have multiple traits (for example Magery).  Add a dummy trait costed as the difference in cost for adding a level, named to instruct the user to increase the level on the other trait.  And finally add a pre-requisite that will always fail such as the character does not have a trait with the same name as the dummy trait.
* Doing this for spells is a little more awkward, as pre-requisites work differently.  A similar structure can be followed with setting the pre-requisite to something which will always be failed, but without referencing the dummy spell name.  For example, requiring a spell with fictional name, or an absurd level in some trait or spell.
* Items whose modifiers and substitutions are settled by the template should be marked Preconfigured in their editors, so the player isn't asked about them again.  Leave the player's decisions unmarked.
* Use plain @code@ placeholders for anything the player must fill in, and define each code in the item's Nameables section with the choices to offer.

## Racial templates
* Have a trait container with type set to "Ancestry", with the page reference set.
* If a suitable Ancestry option exists, select that in the Ancestry container drop down and this will ensure suitable random names, heights, weights, etc are used.
* It is not necessary to have an "Attributes" container within the Ancestry one, as GCS will just add everything to the Ancestry points total
* For traits with trait modifiers, consider removing any which definitely don't apply to the racial template.  As an example humanoid races with Increased ST would likely have the "No Fine Manipulators", "Size" and "Super-Effort" modifiers deleted from the Increased ST used in their template.  This means that users aren't presented with unneeded options.
* Set the body type in the Template block when the race doesn't use the humanoid hit locations.
* It is not possible to enforce an Ancestry

## Profession templates
* Under the main skills container for the profession, have sub-containers for "Primary Skills", "Secondary Skills" and "Background Skills"

## Characters
* If built from a reference source, such as a GURPS book, add a note with a link to the source page
* Any traits which amend attributes (such as "Increased ST") should be in a Trait Container tagged with type "Attributes"
* Ensure the correct body type is selected.  Select Body Type on the sheet's toolbar, then either edit the hit locations to create a custom type or select a standard one from the menu at the top right.
