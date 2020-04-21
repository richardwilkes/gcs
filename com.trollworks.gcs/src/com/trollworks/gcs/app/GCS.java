/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.WiderToolTipUI;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Font;
import java.awt.GraphicsEnvironment;
import java.io.InputStream;
import java.net.URI;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Instant;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;
import java.util.jar.Attributes;
import java.util.jar.Manifest;
import javax.swing.ToolTipManager;
import javax.swing.UIManager;

/** The main entry point for the character sheet. */
public class GCS {
    public static final String WEB_SITE = "https://gurpscharactersheet.com";
    public static final Path   APP_HOME_PATH;
    public static final long   VERSION;
    public static final String COPYRIGHT;
    public static final String COPYRIGHT_BANNER;
    public static final String APP_BANNER;

    static {
        // Fix the current working directory, as bundled apps break the normal logic.
        // Sadly, this still doesn't fix stuff referenced from the "default" filesystem
        // class, as it is already initialized to the wrong value and won't pick this
        // change up.
        String pwd = System.getenv("PWD");
        if (pwd != null && !pwd.isEmpty()) {
            System.setProperty("user.dir", pwd);
        }

        Path path;
        try {
            path = Paths.get(System.getProperty("java.home"));
            if (path.endsWith("Contents/runtime/Contents/Home")) {
                // Running inside a macOS package
                path = path.getParent().getParent().getParent().getParent().getParent();
            } else if (path.endsWith("runtime")) {
                // Running inside a linux package
                path = path.getParent();
            } else if (path.endsWith("support")) {
                // Running inside module-ized package
                path = path.getParent();
            } else {
                URI uri = GCS.class.getProtectionDomain().getCodeSource().getLocation().toURI();
                path = Paths.get(uri).normalize().getParent().toAbsolutePath();
            }
        } catch (Throwable throwable) {
            path = Paths.get(".");
        }
        APP_HOME_PATH = path.normalize().toAbsolutePath();

        long   version = 0;
        String years   = null;
        try (InputStream in = GCS.class.getModule().getResourceAsStream("/META-INF/MANIFEST.MF")) {
            Attributes attributes = new Manifest(in).getMainAttributes();
            version = Version.extract(attributes.getValue("bundle-version"), 0);
            years = attributes.getValue("bundle-copyright-years");
        } catch (Exception exception) {
            // Ignore... we'll fill in default values below
        }
        if (years == null || years.isBlank()) {
            years = "1998-" + DateTimeFormatter.ofPattern("yyyy").format(Instant.now().atZone(ZoneOffset.UTC));
        }
        VERSION = version;
        COPYRIGHT = String.format(I18n.Text("Copyright \u00A9%s by %s"), years, "Richard A. Wilkes");
        COPYRIGHT_BANNER = String.format("%s. All rights reserved.", COPYRIGHT);
        StringBuilder buffer = new StringBuilder();
        buffer.append("GCS ");
        buffer.append(Version.toString(VERSION, false));
        if (VERSION != 0) {
            buffer.append('\n');
            buffer.append(Version.toBuildTimestamp(VERSION));
        }
        buffer.append('\n');
        if (Platform.isWindows()) {
            // The windows command prompt doesn't understand the copyright symbol, so translate it
            // to something it can deal with.
            buffer.append(COPYRIGHT_BANNER.replaceAll("\u00A9", "(c)"));
        } else {
            buffer.append(COPYRIGHT_BANNER);
        }
        APP_BANNER = buffer.toString();
    }

    /**
     * The main entry point for the character sheet.
     *
     * @param args Arguments to the program.
     */
    public static void main(String[] args) {
        boolean      showVersion     = false;
        boolean      showFullVersion = false;
        boolean      generatePDF     = false;
        boolean      generatePNG     = false;
        boolean      generateText    = false;
        Path         template        = null;
        String       margins         = null;
        String       paper           = null;
        List<Path>   files           = new ArrayList<>();
        List<String> msgs            = new ArrayList<>();
        int          length          = args.length;
        for (int i = 0; i < length; i++) {
            String arg = args[i];
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
                    break;
                case "--template":
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
                        msgs.add(I18n.Text("missing argument for --template"));
                    }
                    break;
                case "-v":
                    showVersion = true;
                    break;
                case "--version":
                    showFullVersion = true;
                    break;
                default:
                    msgs.add(I18n.Text("unknown option: ") + parts[0]);
                    break;
                }
            } else {
                files.add(Paths.get(arg));
            }
        }

        if (showVersion || showFullVersion) {
            System.out.println(VERSION != 0 ? Version.toString(VERSION, showFullVersion) : I18n.Text("Development"));
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
            CmdLineExport.export(files, generatePDF, generatePNG, generateText, template, margins, paper);
            System.exit(0);
        }

        if (GraphicsEnvironment.isHeadless()) {
            System.err.println("there is no valid graphics display");
            System.exit(1);
        }
        LaunchProxy launchProxy = new LaunchProxy(files);
        initialize();
        UIApp.startup(launchProxy, files);
    }

    public static void initialize() {
        System.setProperty("apple.laf.useScreenMenuBar", Boolean.TRUE.toString());
        try {
            UIManager.setLookAndFeel(UIManager.getSystemLookAndFeelClassName());
            Font current = UIManager.getFont(Fonts.KEY_STD_TEXT_FIELD);
            UIManager.getDefaults().put(Fonts.KEY_STD_TEXT_FIELD, new Font("SansSerif", current.getStyle(), current.getSize()));
            WiderToolTipUI.installIfNeeded();
        } catch (Exception ex) {
            ex.printStackTrace(System.err);
        }
        Preferences.setPreferenceFile("gcs.pref");
        Fonts.loadFromPreferences();

        // Increase ToolTip time so the user has time to read the skill modifiers
        ToolTipManager.sharedInstance().setDismissDelay(DisplayPreferences.getToolTipTimeout() * 1000);

        if (Library.getRecordedCommit().isBlank()) {
            // No system library present, so download it
            Library.download();
        }
    }

    public static void showHelp() {
        if (VERSION == 0) {
            System.out.println("GCS " + I18n.Text("Development Version"));
        } else {
            System.out.println("GCS " + Version.toString(VERSION, false));
            System.out.println(Version.toBuildTimestamp(VERSION));
        }
        String banner = COPYRIGHT_BANNER;
        if (Platform.isWindows()) {
            // The windows command prompt doesn't understand the copyright symbol, so translate it
            // to something it can deal with.
            banner = banner.replaceAll("\u00A9", "(c)");
        }
        System.out.println(banner);
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
        options.add(I18n.Text("--text"));
        options.add(I18n.Text("Create text versions of sheets specified on the command line."));
        options.add(I18n.Text("--template <file>"));
        options.add(I18n.Text("A template to use when creating text versions of the sheets from the command line. If not specified, the template specified in preferences will be used."));
        options.add(I18n.Text("-v"));
        options.add(I18n.Text("Displays the program version."));
        options.add(I18n.Text("--version"));
        options.add(I18n.Text("Displays the full program version."));
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
}
