# GURPS Character Sheet

GURPS Character Sheet (GCS) is a stand-alone, interactive, character sheet editor that allows you to build characters for the [GURPS](http://www.sjgames.com/gurps) 4<sup>th</sup> Edition roleplaying game system.

## Building from the command line

1. Make sure you have JDK 10 installed. You can download it from Oracle for your platform here: http://www.oracle.com/technetwork/java/javase/downloads/index.html

2. Clone the source repositories:

  ```
  % git clone https://github.com/richardwilkes/com.lowagie.text
  % git clone https://github.com/richardwilkes/gcs
  % git clone https://github.com/richardwilkes/gcs_library
  % git clone https://github.com/richardwilkes/gnu.trove
  % git clone https://github.com/richardwilkes/org.apache.commons.logging
  % git clone https://github.com/richardwilkes/org.apache.fontbox
  % git clone https://github.com/richardwilkes/org.apache.pdfbox
  % git clone https://github.com/richardwilkes/toolkit
  ```

2. Build a distribution package:

  ```
  % cd gcs
  % ./bundle
  ```
