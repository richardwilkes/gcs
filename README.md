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

1. Make sure you have JDK 10 installed. You can download it for your platform
   here: http://www.oracle.com/technetwork/java/javase/downloads/index.html

2. Make sure you have Apache ANT installed. You can download it for your
   platform here: https://ant.apache.org/bindownload.cgi

3. Clone the source repositories:

  ```
  % git clone https://github.com/richardwilkes/gcs
  % cd gcs
  % ant clone-deps
  ```

4. Build and bundle the code for your platform:

  ```
  % cd gcs
  % ant bundle
  ```
