## GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to
build characters for the [GURPS](http://www.sjgames.com/gurps) 4<sup>th</sup> Edition roleplaying
game system.

### Building from the command line

1. Make sure you have JDK 13 installed and set to be used as your default Java compiler. You can
   download it for your platform here: http://jdk.java.net/13/

2. Until the packager is part of the released JDK again, we also need to download a pre-release that
   contains it. You can download it for your platform here: https://jdk.java.net/15/
   
   Do **NOT** place this in your path. The build expects to find it in your home directory. If you
   placed it somewhere else, you'll need to adjust the variable for `jpackage` in the build.xml
   file.

3. Make sure you have Apache ANT installed. You can download it for your platform here:
   https://ant.apache.org/bindownload.cgi

4. If you are building on Windows, you'll need to install WiX Toolset from here:
   https://github.com/wixtoolset/wix3/releases/tag/wix3112rtm

5. Clone the source repositories:
   ```
   % git clone https://github.com/richardwilkes/gcs
   % cd gcs
   % ant clone-deps
   ```

6. Build and bundle the code for your platform:
   ```
   % cd gcs
   % ant bundle
   ```

7. Later, to pull changes that have been committed since your initial build, you can do the
   following to pull those changes and re-build and bundle the code for your platform without going
   through all the steps again:
   ```
   % cd gcs
   % ant pull
   % ant bundle
   ```

As part of doing the `clone-deps` step above, a project for IntelliJ IDEA will have been checked
out into a directory named `gcs_intellij`. This can be used to develop and debug GCS.