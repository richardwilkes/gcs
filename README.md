# GURPS[^1] Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to build
characters for the [GURPS 4<sup>th</sup> Edition](http://www.sjgames.com/gurps) roleplaying game system.

## Branches

### v4-java

This is the current v4.37.1 release branch. This version will no longer receive updates. You can access the
[source code here](https://github.com/richardwilkes/gcs/tree/v4-java).

### master

This is the unreleased v5.0.0 branch for the new version being rewritten using the [Go language](http://go.dev). GCS
relies on another project of mine, [Unison](https://github.com/richardwilkes/unison), for the UI and OS integration. The
[prerequisites](https://github.com/richardwilkes/unison/blob/main/README.md) are therefore the same as for that project.
Once you have the prerequistes, you can build GCS by running the build script: `./build.sh`. Add a `-h` to see available
options.

#### [Remaining issues for v5](https://github.com/richardwilkes/gcs/milestone/1)

[^1]: GURPS is a trademark of Steve Jackson Games, and its rules and art are copyrighted by Steve Jackson Games. All
  rights are reserved by Steve Jackson Games. This game aid is the original creation of Richard A. Wilkes and is
  released for free distribution, and not for resale, under the permissions granted in the
  <a href="http://www.sjgames.com/general/online_policy.html">Steve Jackson Games Online Policy</a>.
