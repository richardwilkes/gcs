# GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to build characters for the [GURPS](http://www.sjgames.com/gurps) 4<sup>th</sup> Edition roleplaying game system.

## Development Setup

GCS is composed of three source projects that I maintain:

- https://github.com/richardwilkes/apple_stubs
- https://github.com/richardwilkes/toolkit
- https://github.com/richardwilkes/gcs

The code is compiled with Java 8. Ant is used to build the product and produce the distribution bundles, however, I use the latest Eclipse release for daily development.

## Building from the command line
1. Clone the source repositories:

  ```
  % git clone https://github.com/richardwilkes/apple_stubs.git
  % git clone https://github.com/richardwilkes/toolkit.git
  % git clone https://github.com/richardwilkes/gcs.git
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

3. Download one (or all) of the JRE packages I've already prepared for use and place them into the launcher directory of the toolkit:

  - [JRE for Mac OS X](http://gcs.trollworks.com/dev_artifacts/jre-mac.zip)  
  - [JRE for 64-bit Windows](http://gcs.trollworks.com/dev_artifacts/jre-windows.zip)  
  - [JRE for 32-bit Windows](http://gcs.trollworks.com/dev_artifacts/jre-windows-32.zip)  
  - [JRE for 64-bit Linux](http://gcs.trollworks.com/dev_artifacts/jre-linux.zip)  
  - [JRE for 32-bit Linux](http://gcs.trollworks.com/dev_artifacts/jre-linux-32.zip)  

4. Build a distribution package for your platform:

  ```
  % cd gcs
  % ant bundle
  ```
