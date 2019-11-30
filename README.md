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
   You should install this such that it is in your path **AFTER** JDK 13.

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
