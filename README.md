## GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet
editor that allows you to build characters for the
[GURPS](http://www.sjgames.com/gurps) 4<sup>th</sup> Edition roleplaying game
system.

### Development

I generally use the latest `Eclipse IDE for Eclipse Commiters` package from
http://www.eclipse.org/downloads/eclipse-packages/ for developing GCS. If you
are using Windows, make sure you set the
`General > Workspace > Default Text Encoding` preference to UTF8, as many
source files contain UTF8 characters and the Windows default encoding will
choke on them.

### Building from the command line

1. Make sure you have JDK 13 installed and set to be used as your default
   Java compiler. You can download it for your platform here:
   http://jdk.java.net/13/

2. Until the packager is part of the released JDK again, we also need to
   download a pre-release that contains it. You can download it for your
   platform here: https://jdk.java.net/jpackage/
   Do **NOT** place this in your path. The build expects to find it in your
   home directory. If you placed it somewhere else, you'll need to adjust
   the variable for `jpackage` in the build.xml file.

3. Make sure you have Apache ANT installed. You can download it for your
   platform here: https://ant.apache.org/bindownload.cgi

4. Clone the source repositories:

   ```
   % git clone https://github.com/richardwilkes/gcs
   % cd gcs
   % ant clone-deps
   ```

5. Build and bundle the code for your platform:

   ```
   % cd gcs
   % ant bundle
   ```

   Linux is not currently bundling correctly. If you're on that platform,
   you can either roll your repos back to a known working version or forego
   the bundling and do this instead:

   ```
   % cd gcs
   % ant deps build
   ```

   ... and then this to run it:
   
   ```
   % java --module-path ../java_modules --module com.trollworks.gcs/com.trollworks.gcs.app.GCS
   ```
