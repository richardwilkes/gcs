/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs;

import com.trollworks.gcs.cmdline.Export;
import com.trollworks.gcs.cmdline.LoadSave;
import com.trollworks.gcs.menu.edit.PreferencesCommand;
import com.trollworks.gcs.menu.file.OpenCommand;
import com.trollworks.gcs.menu.file.OpenDataFileCommand;
import com.trollworks.gcs.menu.file.PrintCommand;
import com.trollworks.gcs.menu.file.QuitCommand;
import com.trollworks.gcs.menu.help.AboutCommand;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.UpdateChecker;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.utility.launchproxy.LaunchProxy;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Desktop;
import java.awt.Desktop.Action;
import java.awt.EventQueue;
import java.awt.GraphicsEnvironment;
import java.awt.desktop.QuitStrategy;
import java.io.InputStream;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Instant;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;
import java.util.jar.Attributes;
import java.util.jar.Manifest;
import java.util.regex.Pattern;

/** The main entry point for the character sheet. */
public final class GCS {
    public static final  String  WEB_SITE          = "https://gurpscharactersheet.com";
    public static final  Version VERSION           = new Version();
    public static final  String  COPYRIGHT;
    public static final  String  COPYRIGHT_FOOTER;
    public static final  String  APP_BANNER;
    private static final Pattern COPYRIGHT_PATTERN = Pattern.compile("©");
    private static       boolean NOTIFICATION_ALLOWED;

    static {
        // Fix the current working directory, as bundled apps break the normal logic.
        // Sadly, this still doesn't fix stuff referenced from the "default" filesystem
        // class, as it is already initialized to the wrong value and won't pick this
        // change up.
        String pwd = System.getenv("PWD");
        if (pwd != null && !pwd.isEmpty()) {
            System.setProperty("user.dir", pwd);
        }

        // Determine the version
        String years = null;
        try (InputStream in = GCS.class.getModule().getResourceAsStream("/META-INF/MANIFEST.MF")) {
            Attributes attributes = new Manifest(in).getMainAttributes();
            VERSION.extract(attributes.getValue("bundle-version"));
            years = attributes.getValue("bundle-copyright-years");
        } catch (Exception exception) {
            // Ignore... we'll fill in default values below
        }
        if (years == null || years.isBlank()) {
            years = "1998-" + DateTimeFormatter.ofPattern("yyyy").format(Instant.now().atZone(ZoneOffset.UTC));
        }

        // Setup localizations -- must be called AFTER the version is determined
        I18n.initialize();

        // Setup the copyright notices and such that rely on the version and year info
        COPYRIGHT = String.format(I18n.Text("Copyright ©%s by %s"), years, "Richard A. Wilkes");
        COPYRIGHT_FOOTER = String.format("GCS " + I18n.Text("is copyrighted ©%s by %s"), years, "Richard A. Wilkes");
        StringBuilder buffer = new StringBuilder();
        buffer.append("GCS ");
        if (VERSION.isZero()) {
            buffer.append(I18n.Text("(development)"));
        } else {
            buffer.append(VERSION);
        }
        buffer.append('\n');
        if (Platform.isWindows()) {
            // The windows command prompt doesn't understand the copyright symbol, so translate it
            // to something it can deal with.
            buffer.append(COPYRIGHT_PATTERN.matcher(COPYRIGHT).replaceAll("(c)"));
        } else {
            buffer.append(COPYRIGHT);
        }
        buffer.append(". All rights reserved.");
        APP_BANNER = buffer.toString();
    }

    private GCS() {
    }

    /**
     * The main entry point for the character sheet.
     *
     * @param args Arguments to the program.
     */
    public static void main(String[] args) {
        boolean      showVersion  = false;
        boolean      generatePDF  = false;
        boolean      generatePNG  = false;
        boolean      generateText = false;
        boolean      loadSave     = false;
        Path         template     = null;
        String       margins      = null;
        String       paper        = null;
        List<Path>   files        = new ArrayList<>();
        List<String> msgs         = new ArrayList<>();
        int          length       = args.length;
        for (int i = 0; i < length; i++) {
            String arg = args[i];
            if (i == 0 && Platform.isMacintosh() && arg.startsWith("-psn_")) {
                continue;
            }
            if (arg.startsWith("-")) {
                if (arg.startsWith("=")) {
                    files.add(Paths.get(arg));
                    continue;
                }
                String[] parts = arg.split("=", 2);
                switch (parts[0]) {
                case "-h", "--help" -> showHelp();
                case "--margins" -> {
                    boolean missingMarginsArg = false;
                    if (parts.length > 1) {
                        if (parts[1].isBlank()) {
                            missingMarginsArg = true;
                        } else {
                            margins = parts[1];
                        }
                    } else {
                        i++;
                        if (i < length && !args[i].startsWith("-")) {
                            margins = args[i];
                        } else {
                            missingMarginsArg = true;
                        }
                    }
                    if (missingMarginsArg) {
                        msgs.add(I18n.Text("missing argument for --margins"));
                    }
                }
                case "--paper" -> {
                    boolean missingPaperArg = false;
                    if (parts.length > 1) {
                        if (parts[1].isBlank()) {
                            missingPaperArg = true;
                        } else {
                            paper = parts[1];
                        }
                    } else {
                        i++;
                        if (i < length && !args[i].startsWith("-")) {
                            paper = args[i];
                        } else {
                            missingPaperArg = true;
                        }
                    }
                    if (missingPaperArg) {
                        msgs.add(I18n.Text("missing argument for --paper"));
                    }
                }
                case "--pdf" -> generatePDF = true;
                case "--png" -> generatePNG = true;
                case "--text" -> {
                    generateText = true;
                    boolean missingTemplateArg = false;
                    if (parts.length > 1) {
                        if (parts[1].isBlank()) {
                            missingTemplateArg = true;
                        } else {
                            template = Paths.get(parts[1]);
                        }
                    } else {
                        i++;
                        if (i < length && !args[i].startsWith("-")) {
                            template = Paths.get(args[i]);
                        } else {
                            missingTemplateArg = true;
                        }
                    }
                    if (missingTemplateArg) {
                        msgs.add(I18n.Text("missing argument for --text"));
                    }
                }
                case "--loadsave" -> loadSave = true;
                case "-v", "--version" -> showVersion = true;
                default -> msgs.add(I18n.Text("unknown option: ") + parts[0]);
                }
            } else {
                files.add(Paths.get(arg));
            }
        }

        if (showVersion) {
            System.out.println(VERSION);
            System.exit(0);
        }

        if (!msgs.isEmpty()) {
            for (String msg : msgs) {
                System.err.println(msg);
            }
            System.exit(1);
        }

        if (loadSave) {
            LoadSave.process(files);
            System.exit(0);
        }

        if (generatePDF || generatePNG || generateText) {
            Export.process(files, generatePDF, generatePNG, generateText, template, margins, paper);
            System.exit(0);
        }

        if (GraphicsEnvironment.isHeadless()) {
            System.err.println("there is no valid graphics display");
            System.exit(1);
        }

        LaunchProxy launchProxy = new LaunchProxy(files);
        UIUtilities.initialize();

        if (Desktop.isDesktopSupported()) {
            Desktop desktop = Desktop.getDesktop();
            if (desktop.isSupported(Action.APP_ABOUT)) {
                desktop.setAboutHandler(AboutCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_PREFERENCES)) {
                desktop.setPreferencesHandler(PreferencesCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_OPEN_FILE)) {
                desktop.setOpenFileHandler(OpenCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_PRINT_FILE)) {
                desktop.setPrintFileHandler(PrintCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_QUIT_HANDLER)) {
                desktop.setQuitHandler(QuitCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_QUIT_STRATEGY)) {
                desktop.setQuitStrategy(QuitStrategy.NORMAL_EXIT);
            }
            if (desktop.isSupported(Action.APP_SUDDEN_TERMINATION)) {
                desktop.disableSuddenTermination();
            }
        }

        UpdateChecker.check();
        launchProxy.setReady(true);

        EventQueue.invokeLater(() -> {
            Workspace.get();
            OpenDataFileCommand.enablePassThrough();
            for (Path file : files) {
                OpenDataFileCommand.open(file);
            }
            if (Platform.isMacintosh() && System.getProperty("java.home").toLowerCase().contains("/apptranslocation/")) {
                WindowUtils.showError(null, Text.wrapToCharacterCount(I18n.Text("macOS has translocated GCS, restricting access to the file system and preventing access to the data library. To fix this, you must quit GCS, then run the following command in the terminal after cd'ing into the GURPS Character Sheet folder:\n\n"), 60) + "xattr -d com.apple.quarantine \"/Applications/GCS.app\"");
            }
            setNotificationAllowed(true);
        });
    }

    private static void showHelp() {
        System.out.println(APP_BANNER);
        System.out.println();
        System.out.println(I18n.Text("Available options:"));
        System.out.println();
        List<String> options = new ArrayList<>();
        options.add(I18n.Text("-h, --help"));
        options.add(I18n.Text("Displays a description of each option."));
        options.add(I18n.Text("--loadsave"));
        options.add(I18n.Text("Load and then save all files specified on the command line. If a directory is specified, it will be traversed recursively and all files found will be loaded and saved. This operation is intended to easily bring files up to the current version's data format. After all files have been processed, GCS will exit."));
        options.add(I18n.Text("--margins <margins>"));
        options.add(I18n.Text("When generating PDF or PNG from the command line, allows you to specify the margins to use, rather than the ones embedded in the file. The top, left, bottom, and right margins must all be specified in inches, separated by colons, such as '1:1:1:1'."));
        options.add(I18n.Text("--paper <size>"));
        options.add(I18n.Text("When generating PDF or PNG from the command line, allows you to specify a paper size to use, rather than the one embedded in the file. Valid choices are: LETTER, A4, or the width and height, expressed in inches and separated by an 'x', such as '5x7'."));
        options.add(I18n.Text("--pdf"));
        options.add(I18n.Text("Create PDF versions of sheets specified on the command line."));
        options.add(I18n.Text("--png"));
        options.add(I18n.Text("Create PNG versions of sheets specified on the command line."));
        options.add(I18n.Text("--text <file>"));
        options.add(I18n.Text("Create text versions of sheets specified on the command line using the specified template file."));
        options.add(I18n.Text("-v, --version"));
        options.add(I18n.Text("Displays the program version."));
        int longest = 0;
        int length  = options.size();
        for (int i = 0; i < length; i += 2) {
            String option = options.get(i);
            if (longest < option.length()) {
                longest = option.length();
            }
        }
        String optionFormat = String.format("  %%-%ds  ", Integer.valueOf(longest));
        String indent       = Text.makeFiller(longest + 4, ' ');
        for (int i = 0; i < length; i += 2) {
            System.out.printf(optionFormat, options.get(i));
            String[] lines = Text.wrapToCharacterCount(options.get(i + 1), 76 - longest).split("\n");
            int      count = lines.length;
            for (int j = 0; j < count; j++) {
                if (j != 0) {
                    System.out.print(indent);
                }
                System.out.println(lines[j]);
            }
        }
        System.exit(0);
    }

    /** @return Whether it is OK to put up a notification dialog yet. */
    public static synchronized boolean isNotificationAllowed() {
        return NOTIFICATION_ALLOWED;
    }

    /** @param allowed Whether it is OK to put up a notification dialog yet. */
    public static synchronized void setNotificationAllowed(boolean allowed) {
        NOTIFICATION_ALLOWED = allowed;
    }
}
