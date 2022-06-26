# GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to build characters
for the [GURPS 4<sup>th</sup> Edition](http://www.sjgames.com/gurps) roleplaying game system.

<hr>

## Branches

### v4-java

This is the current v4.37.1 release branch.

### master

This is the unreleased v5.0.0 branch for the new version being rewritten using the [Go language](http://go.dev).

#### GCS-specific work that still needs to be done

- Add undo records for edit operations that don't already have them
- Implement prompting for substitution text when moving items onto a sheet
- Add monitoring of the library directories for file changes
    - Perhaps also add manual refresh option, for those platforms where disk monitoring is less than optimal
- Settings editors
    - Attributes
    - Body Type
- Library configuration dialogs
- Completion of menu item actions
    - Item
        - Copy to Character Sheet
        - Copy to Template
        - Apply Template to Character Sheet
    - Library
        - Update <library> to <version>
        - Change Library Locations
    - Settings
        - Attributes...
        - Default Attributes...
        - Body Type...
        - Default Body Type...
- Printing support for sheets (requires support in unison first)

#### [Unison](https://github.com/richardwilkes/unison)-specific work that still needs to be done

- When closing a tab, focus the next tab in the same dockable area before moving the focus to another dockable area
- Printing support
- Carefully comb over the interface and identify areas where things aren't working well on Windows and Linux, since I
  spend nearly all of my development time on macOS and may not have noticed deficiencies on the other platforms
