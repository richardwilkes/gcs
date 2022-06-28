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

#### [Remaining issues for v5](https://github.com/richardwilkes/gcs/issues?q=is%3Aopen+is%3Aissue+label%3Av5)