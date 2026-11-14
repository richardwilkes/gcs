# Installing and Updating GCS

GURPS Character Sheet (GCS) is available for free for macOS, Windows, and Linux.

Follow the instructions below to download and install GCS for your operating system.

## System requirements

| Operating system | Requirements                                                                                                                                               |
|------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **macOS**        | macOS 11 (Big Sur) or newer, on Apple silicon or Intel.                                                                                                    |
| **Windows**      | Windows 10 or newer, on x86-64 or ARM64.                                                                                                                   |
| **Linux**        | glibc 2.35 or newer (Ubuntu 22.04, Debian 12, Fedora 36, or equivalent), on x86-64 or ARM64. Only the standard X11 and OpenGL client libraries are needed. |

## Download GCS

- Open your browser and go to the [GURPS Character Sheet website](https://gurpscharactersheet.com/) to download the
  latest version of GCS.

## Install on macOS

1. From the [GCS website](https://gurpscharactersheet.com/), select the correct version of **macOS** for your system
   ([Apple silicon vs Intel](https://support.apple.com/en-gb/116943)).
2. Open the downloaded disk image.
3. In the window that opens, drag the **GCS** icon into the **Applications** folder.
4. Launch GCS from your Applications folder as you would any other application on your Mac.

> [!NOTE]
> Run GCS from your Applications folder, not from the disk image. GCS can't update itself while it runs from the disk
> image.

## Install GCS on Windows

1. From the [GCS website](https://gurpscharactersheet.com/), select the **Windows** download for your system (x86-64 or
   ARM64).
2. Open the downloaded installation file and follow the installation wizard.
3. Launch GCS as you would any other application on Windows.

> [!NOTE]
> After installing, you may see a "Windows protected your PC” warning when launching GCS. See the
> [Troubleshooting](Troubleshooting) page for more details on how to fix this.

## Install GCS on Linux

### Quick install

1. From the [GCS website](https://gurpscharactersheet.com), select the **Linux** download for your system (x86-64 or
   ARM64) to download the `.tgz` file.
2. Extract it to a folder.
3. Run `./gcs` from the extracted folder (or double-click the file).

### Full install instructions

1. From the [GCS website](https://gurpscharactersheet.com), select the **Linux** download for your system to download
   the `.tgz` file (tarball).
2. Extract the `.tgz` file to a folder of your choice.
3. Launch GCS using one of the following methods:
   - **From the terminal:** Open a terminal in the extracted folder and run `./gcs`.
   - **By clicking:** Double-click the `gcs` file in your file manager.
   - **Via a desktop launcher:** Create a shortcut/launcher pointing to the `gcs` file for easy access. On GNOME-based
     desktops, GCS installs a desktop icon for itself the first time it is launched.

## Updating GCS

By default, GCS checks for a new version each time it launches. You can change how often it checks (at launch, hourly,
daily, or never) with the **Check for GCS Updates** option in [General Settings](General%20Settings), and you can check
at any time with **Help > Check for GCS updates**.

When a new version is available:

1. A pulsing **Software Update Available** button appears in the [Library Explorer](Library%20Explorer) toolbar. The
   update window also opens on its own when the check GCS makes at launch finds a release it hasn't shown you before.
2. Select the button to open the update window, which shows the **release notes** for the new version.
3. Select **Install & Restart**. GCS downloads the release for your system and verifies it, with a progress window
   whose **Cancel** button stops the update. It then quits, installs the new version over the copy you're running, and
   restarts.

If GCS can't update itself (for example, because it was installed by Homebrew or a Linux package manager, is running
from the disk image, or is in a folder it can't write to), the window explains why and offers **Download Page**
instead. Select it to open the [GURPS Character Sheet website](https://gurpscharactersheet.com/), then follow the
installation steps above to install the new version by hand.

Updating GCS does **not** affect your existing libraries or character sheets. They remain on your computer and will be
available after the update.

To update the **Master Library**, see the [Library Explorer](Library%20Explorer) page.

## Releases

For details on the latest versions, including new features and bug fixes, see the [GCS Releases page on
GitHub](https://github.com/richardwilkes/gcs/releases) or select **Help > Release Notes**.

## Next Steps

With GCS installed, you’re ready to [explore the GCS interface](Explore%20the%20GCS%20interface) to understand how
everything is organized.

---

[Return to Home](Home)
