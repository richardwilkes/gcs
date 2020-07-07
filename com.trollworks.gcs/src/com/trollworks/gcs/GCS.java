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

import com.trollworks.gcs.character.CmdLineExport;
import com.trollworks.gcs.menu.edit.PreferencesCommand;
import com.trollworks.gcs.menu.file.OpenCommand;
import com.trollworks.gcs.menu.file.OpenDataFileCommand;
import com.trollworks.gcs.menu.file.PrintCommand;
import com.trollworks.gcs.menu.file.QuitCommand;
import com.trollworks.gcs.menu.help.AboutCommand;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.WiderToolTipUI;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.UpdateChecker;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.utility.launchproxy.LaunchProxy;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Desktop;
import java.awt.Desktop.Action;
import java.awt.EventQueue;
import java.awt.Font;
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
import javax.swing.UIManager;

/** The main entry point for the character sheet. */
public class GCS {
    public static final String  WEB_SITE        = "https://gurpscharactersheet.com";
    public static final Version VERSION         = new Version();
    public static final Version LIBRARY_VERSION = new Version();
    public static final String  COPYRIGHT;
    public static final String  COPYRIGHT_FOOTER;
    public static final String  APP_BANNER;
    private static      boolean NOTIFICATION_ALLOWED;

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
        COPYRIGHT_FOOTER = String.format("GCS " + I18n.Text(" is copyrighted ©%s by %s"), years, "Richard A. Wilkes");
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
            buffer.append(COPYRIGHT.replaceAll("©", "(c)"));
        } else {
            buffer.append(COPYRIGHT);
        }
        buffer.append(". All rights reserved.");
        APP_BANNER = buffer.toString();
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
                case "-h":
                case "--help":
                    showHelp();
                    break;
                case "--margins":
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
                    break;
                case "--paper":
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
                    break;
                case "--pdf":
                    generatePDF = true;
                    break;
                case "--png":
                    generatePNG = true;
                    break;
                case "--text":
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
                    break;
                case "-v":
                case "--version":
                    showVersion = true;
                    break;
                default:
                    msgs.add(I18n.Text("unknown option: ") + parts[0]);
                    break;
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

        if (generatePDF || generatePNG || generateText) {
            if (files.isEmpty()) {
                System.err.println(I18n.Text("must specify one or more sheet files to process"));
                System.exit(1);
            }
            System.setProperty("java.awt.headless", Boolean.TRUE.toString());
            initialize();
            try {
                // This is run on the event queue since much of the sheet logic assumes a UI
                // environment and would otherwise cause concurrent modification exceptions, as the
                // detection of whether it was safe to modify data would be inaccurate.
                EventQueue.invokeAndWait(new CmdLineExport(files, generatePDF, generatePNG, generateText, template, margins, paper));
            } catch (Exception exception) {
                exception.printStackTrace(System.err);
                System.exit(1);
            }
            System.exit(0);
        }

        if (GraphicsEnvironment.isHeadless()) {
            System.err.println("there is no valid graphics display");
            System.exit(1);
        }

        LaunchProxy launchProxy = new LaunchProxy(files);
        initialize();

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

    public static void initialize() {
        System.setProperty("apple.laf.useScreenMenuBar", Boolean.TRUE.toString());
        try {
            UIManager.setLookAndFeel(UIManager.getSystemLookAndFeelClassName());
            Font current = UIManager.getFont(Fonts.KEY_STD_TEXT_FIELD);
            UIManager.getDefaults().put(Fonts.KEY_STD_TEXT_FIELD, new Font("SansSerif", current.getStyle(), current.getSize()));
            WiderToolTipUI.installIfNeeded();
        } catch (Exception ex) {
            Log.error(ex);
        }
        Fonts.loadFromPreferences();
    }

    public static void showHelp() {
        System.out.println(APP_BANNER);
        System.out.println();
        System.out.println(I18n.Text("Available options:"));
        System.out.println();
        List<String> options = new ArrayList<>();
        options.add(I18n.Text("-h, --help"));
        options.add(I18n.Text("Displays a description of each option."));
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
