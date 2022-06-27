# GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to build characters
for the [GURPS 4<sup>th</sup> Edition](http://www.sjgames.com/gurps) roleplaying game system.

<hr>

## Branches

### v4-java

This is the current v4.37.1 release branch. You can access it [here](https://github.com/richardwilkes/gcs/tree/v4-java).

### master

This is the unreleased v5.0.0 branch for the new version being rewritten using the [Go language](http://go.dev). GCS
relies on another project of mine, [Unison](https://github.com/richardwilkes/unison), for the UI and OS integration. The
prerequisites are therefore the same as for that project and are listed
[here](https://github.com/richardwilkes/unison/blob/main/README.md). Once you have the prerequistes, you can build GCS
by running the build script: `./build.sh`. Add a `-h` to see available options.

#### GCS-specific work that still needs to be done

- Some undo operations don't work correctly because the underlying user interface object is discarded during a rebuild
  of the display. Need to provide a mechanism to discover the new object the data should be attached to. An example is
  the fields in the top section of the sheet... some edits cause a rebuild, rather than a simple update, and when that
  happens, the associated undos that came before the rebuild can no longer function.
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
