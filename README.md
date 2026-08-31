# GURPS Character Sheet

[![Go Reference](https://pkg.go.dev/badge/github.com/richardwilkes/gcs/v5.svg)](https://pkg.go.dev/github.com/richardwilkes/gcs/v5)
[![Build](https://github.com/richardwilkes/gcs/actions/workflows/build.yml/badge.svg)](https://github.com/richardwilkes/gcs/actions/workflows/build.yml)
[![Release](https://github.com/richardwilkes/gcs/actions/workflows/release.yml/badge.svg)](https://github.com/richardwilkes/gcs/actions/workflows/release.yml)

GURPS[^1] Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to build
characters for the [GURPS Fourth Edition](http://www.sjgames.com/gurps) roleplaying game system.

GCS relies on another project of mine, [Unison](https://github.com/richardwilkes/unison),
for the UI and OS integration. The [prerequisites](https://github.com/richardwilkes/unison/blob/main/README.md) are
therefore the same as for that project. Once you have the prerequistes, you can build GCS by running the build script:
`./build.sh`. Add a `-h` to see available options.

## Generative AI Use Policy

Generative AI isn't forbidden in the GURPS Character Sheet project, and you're welcome to use it while preparing a
contribution. If you do, the expectations below apply.

Keep in mind that Generative AI can produce a large amount of plausible-looking code that takes a long time to review
carefully, and that it can't learn to avoid repeating the same mistake. Code that compiles and passes the tests can
still be wrong, can duplicate functionality that already exists elsewhere in the code base, and can quietly change
behavior beyond what was asked for. Those problems are easy to overlook in a large diff and hard to catch in review.

When you submit:

* **Say which parts were AI-generated.** A short note on the Pull Request or Issue is enough.
* **Take as much responsibility for your submission as if you wrote it yourself.** You're expected to have read every
  change, understood all of it, and fixed every error you noticed before asking for review. Be prepared to explain why
  the code does what it does, not just what it does.
* **Build and run it first.** Run `./build.sh -a` and confirm it completes cleanly, then exercise your change in the
  running application. Passing tests are necessary, but they aren't a substitute for trying it out.
* **Keep the change focused.** Don't let the tool reformat or rewrite code unrelated to what you're changing, and don't
  add new dependencies without discussing them first. Follow the conventions of the surrounding code.
* **Keep the discussion respectful and on-topic.** The relative merits of Generative AI won't be settled in a discussion
  thread, and we all have to collaborate afterwards.

[^1]: GURPS is a trademark of Steve Jackson Games, and its rules and art are copyrighted by Steve Jackson Games. All
rights are reserved by Steve Jackson Games. This game aid is the original creation of Richard A. Wilkes and is
released for free distribution, and not for resale, under the permissions granted in the
[Steve Jackson Games Online Policy](http://www.sjgames.com/general/online_policy.html).
