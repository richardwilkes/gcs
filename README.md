## GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet
editor that allows you to build characters for the
[GURPS 4<sup>th</sup> Edition](http://www.sjgames.com/gurps) roleplaying game
system.

### Building from the command line

1. Make sure you have JDK 14 installed and set to be used as your default Java compiler. You can
   download it for your platform here: http://jdk.java.net/14/

2. If you are building on Windows, you'll need to install the WiX Toolset from here:
   https://github.com/wixtoolset/wix3/releases/tag/wix3112rtm

3. Clone the source repositories:
   ```
   % git clone https://github.com/richardwilkes/gcs
   ```
   and optionally:
   ```
   % git clone https://github.com/richardwilkes/gcs_library
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
