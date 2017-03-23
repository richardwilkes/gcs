# GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to build characters for the [GURPS](http://www.sjgames.com/gurps) 4<sup>th</sup> Edition roleplaying game system.

## Development Setup

GCS is composed of four source projects that I maintain:

- https://github.com/richardwilkes/apple_stubs
- https://github.com/richardwilkes/toolkit
- https://github.com/richardwilkes/gcs
- https://github.com/richardwilkes/gcs_library

The code is compiled with Java 8. Ant is used to build the product and produce the distribution bundles, however, I use the latest Eclipse release for daily development.

## Building from the command line
1. Clone the source repositories:

  ```
  % git clone https://github.com/richardwilkes/apple_stubs.git
  % git clone https://github.com/richardwilkes/toolkit.git
  % git clone https://github.com/richardwilkes/gcs.git
  % git clone https://github.com/richardwilkes/gcs_library.git
  ```

2. Compile the code:

  ```
  % cd apple_stubs
  % ant
  % cd ../toolkit
  % ant
  % cd ../gcs
  % ant
  ```

  If you want to build a distribution package, you'll also need to:

3. Download one (or all) of the JRE packages from Oracle. Click the button to download the JRE from the [main download page](http://www.oracle.com/technetwork/java/javase/downloads/index.html) and put it in the toolkit/launcher folder. This will take you to a form where you can download the packages for each platform.

4. Build a distribution package for your platform:

  ```
  % cd gcs
  % ant bundle
  ```
