## GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet
editor that allows you to build characters for the
[GURPS 4<sup>th</sup> Edition](http://www.sjgames.com/gurps) roleplaying game
system.

### Building from the command line

**NOTE**: *To build a specific version of GCS, you will need to check out the appropriate release
tag. These directions are for the latest source, which may have experimental code or changes that
are incompatible with the current data files. These build instructions may have also changed since
a given release, so be sure to review them again with the version you plan to build.*

1. Make sure you have JDK 15 installed and set to be used as your default Java compiler. You can
   download it for your platform here: http://jdk.java.net/15/

2. If you are building on Windows, you'll need to install the WiX Toolset from here:
   https://github.com/wixtoolset/wix3/releases/tag/wix3112rtm

3. Clone the source repositories:
   ```
   % git clone https://github.com/richardwilkes/gcs
   ```

4. Build and bundle the code for your platform:

   macOS and Linux:
   ```
   % cd gcs
   % ./bundle.sh
   ```
   Windows:
   ```
   > cd gcs
   > .\bundle.bat
   ```
