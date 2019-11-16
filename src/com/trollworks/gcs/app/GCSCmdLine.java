/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.PrerequisitesThread;
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.template.Template;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.ui.Fonts;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.print.PrintManager;
import com.trollworks.toolkit.utility.BundleInfo;
import com.trollworks.toolkit.utility.Dice;
import com.trollworks.toolkit.utility.FileProxyCreator;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.LaunchProxy;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.Timing;
import com.trollworks.toolkit.utility.cmdline.CmdLine;
import com.trollworks.toolkit.utility.cmdline.CmdLineOption;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.units.LengthUnits;

import java.io.File;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.StringTokenizer;

import javax.swing.ToolTipManager;

/**
 * Responsible for most setup and determining whether just the command line or the full graphical
 * app is being launched.
 */
public class GCSCmdLine {
    private static final CmdLineOption PDF_OPTION           = new CmdLineOption(I18n.Text("Create PDF versions of sheets specified on the command line."), null, FileType.PDF_EXTENSION);
    private static final CmdLineOption TEXT_OPTION          = new CmdLineOption(I18n.Text("Create text versions of sheets specified on the command line."), null, "text");
    private static final CmdLineOption TEXT_TEMPLATE_OPTION = new CmdLineOption(I18n.Text("A template to use when creating text versions of the sheets. If this is not specified, then the template found in the data directory will be used by default."), I18n.Text("FILE"), "text_template");
    private static final CmdLineOption PNG_OPTION           = new CmdLineOption(I18n.Text("Create PNG versions of sheets specified on the command line."), null, FileType.PNG_EXTENSION);
    private static final CmdLineOption SIZE_OPTION          = new CmdLineOption(I18n.Text("When generating PDF or PNG from the command line, allows you to specify a paper size to use, rather than the one embedded in the file. Valid choices are: LETTER, A4, or the width and height, expressed in inches and separated by an 'x', such as '5x7'."), I18n.Text("SIZE"), "paper");
    private static final CmdLineOption MARGIN_OPTION        = new CmdLineOption(I18n.Text("When generating PDF or PNG from the command line, allows you to specify the margins to use, rather than the ones embedded in the file. The top, left, bottom, and right margins must all be specified in inches, separated by colons, such as '1:1:1:1'."), I18n.Text("MARGINS"), "margins");

    /**
     * Start the character sheet.
     *
     * @param args Arguments to the program.
     */
    public static void start(String[] args) {
        Dice.setAssumedSideCount(6);
        CmdLine cmdLine = new CmdLine();
        cmdLine.addOptions(TEXT_OPTION, TEXT_TEMPLATE_OPTION, PDF_OPTION, PNG_OPTION, SIZE_OPTION, MARGIN_OPTION);
        cmdLine.processArguments(args);
        if (cmdLine.isOptionUsed(TEXT_OPTION) || cmdLine.isOptionUsed(PDF_OPTION) || cmdLine.isOptionUsed(PNG_OPTION)) {
            System.setProperty("java.awt.headless", Boolean.TRUE.toString());
            initialize();
            Timing timing = new Timing();
            System.out.println(BundleInfo.getDefault().getAppBanner());
            System.out.println();
            if (convert(cmdLine) < 1) {
                System.out.println(I18n.Text("You must specify one or more sheet files to process."));
                System.exit(1);
            }
            System.out.println(MessageFormat.format(I18n.Text("\nDone! {0} overall."), timing));
            System.exit(0);
        } else {
            LaunchProxy.configure(cmdLine.getArgumentsAsFiles());
            if (GraphicsUtilities.areGraphicsSafeToUse()) {
                initialize();
                GCSApp.INSTANCE.startup(cmdLine);
            } else {
                System.err.println(GraphicsUtilities.getReasonForUnsafeGraphics());
                System.exit(1);
            }
        }
    }

    private static void initialize() {
        GraphicsUtilities.configureStandardUI();
        Preferences.setPreferenceFile("gcs.pref");
        GCSFonts.register();
        GCSImages.getAppIcons(); // Just here to make sure the lookup paths are initialized
        Fonts.loadFromPreferences();
        App.setAboutPanel(AboutPanel.class);
        registerFileTypes(new GCSFileProxyCreator());

        // Increase TooTip time so the user has time to read the skill modifiers
        ToolTipManager.sharedInstance().setDismissDelay(DisplayPreferences.getToolTipTimeout() * 1000);
    }

    /** @return The path to the GCS library files. */
    public static Path getLibraryRootPath() {
        String library = System.getenv("GCS_LIBRARY");
        if (library != null) {
            return Paths.get(library);
        }
        Path path = App.getHomePath();
        if (BundleInfo.getDefault().getVersion() == 0) {
            path = path.resolve("../gcs_library");
        }
        return path.resolve("Library");
    }

    /**
     * Registers the file types the app can open.
     *
     * @param fileProxyCreator The {@link FileProxyCreator} to use.
     */
    public static void registerFileTypes(FileProxyCreator fileProxyCreator) {
        FileType.register(GURPSCharacter.EXTENSION, GCSImages.getCharacterSheetDocumentIcons(), I18n.Text("GURPS Character Sheet"), GCSApp.WEB_SITE, fileProxyCreator, true, true);
        FileType.register(AdvantageList.EXTENSION, GCSImages.getAdvantagesDocumentIcons(), I18n.Text("GCS Advantages Library"), GCSApp.WEB_SITE, fileProxyCreator, true, true);
        FileType.register(EquipmentList.EXTENSION, GCSImages.getEquipmentDocumentIcons(), I18n.Text("GCS Equipment Library"), GCSApp.WEB_SITE, fileProxyCreator, true, true);
        FileType.register(SkillList.EXTENSION, GCSImages.getSkillsDocumentIcons(), I18n.Text("GCS Skills Library"), GCSApp.WEB_SITE, fileProxyCreator, true, true);
        FileType.register(SpellList.EXTENSION, GCSImages.getSpellsDocumentIcons(), I18n.Text("GCS Spells Library"), GCSApp.WEB_SITE, fileProxyCreator, true, true);
        FileType.register(NoteList.EXTENSION, GCSImages.getNoteDocumentIcons(), I18n.Text("GCS Notes Library"), GCSApp.WEB_SITE, fileProxyCreator, true, true);
        FileType.register(Template.EXTENSION, GCSImages.getTemplateDocumentIcons(), I18n.Text("GCS Character Template"), GCSApp.WEB_SITE, fileProxyCreator, true, true);
        // For legacy
        FileType.register(LibraryFile.EXTENSION, GCSImages.getAdvantagesDocumentIcons(), I18n.Text("GCS Library"), GCSApp.WEB_SITE, fileProxyCreator, true, true);

        FileType.registerPdf(null, fileProxyCreator, true, false);
        FileType.registerHtml(null, null, false, false);
        FileType.registerXml(null, null, false, false);
        FileType.registerPng(null, null, false, false);
    }

    private static int convert(CmdLine cmdLine) {
        boolean text  = cmdLine.isOptionUsed(TEXT_OPTION);
        boolean pdf   = cmdLine.isOptionUsed(PDF_OPTION);
        boolean png   = cmdLine.isOptionUsed(PNG_OPTION);
        int     count = 0;

        if (text || pdf || png) {
            double[] paperSize          = getPaperSize(cmdLine);
            double[] margins            = getMargins(cmdLine);
            Timing   timing             = new Timing();
            String   textTemplateOption = cmdLine.getOptionArgument(TEXT_TEMPLATE_OPTION);
            File     textTemplate       = null;

            if (textTemplateOption != null) {
                textTemplate = new File(textTemplateOption);
            }
            GraphicsUtilities.setHeadlessPrintMode(true);
            for (File file : cmdLine.getArgumentsAsFiles()) {
                if (GURPSCharacter.EXTENSION.equals(PathUtils.getExtension(file.getName())) && file.canRead()) {
                    System.out.print(MessageFormat.format(I18n.Text("Loading \"{0}\"... "), file));
                    System.out.flush();
                    timing.reset();
                    try {
                        GURPSCharacter      character = new GURPSCharacter(file);
                        CharacterSheet      sheet     = new CharacterSheet(character);
                        PrerequisitesThread prereqs   = new PrerequisitesThread(sheet);
                        PrintManager        settings  = character.getPageSettings();
                        File                output;
                        boolean             success;

                        sheet.addNotify(); // Required to allow layout to work
                        sheet.rebuild();
                        prereqs.start();
                        PrerequisitesThread.waitForProcessingToFinish(character);

                        if (paperSize != null && settings != null) {
                            settings.setPageSize(paperSize, LengthUnits.IN);
                        }
                        if (margins != null && settings != null) {
                            settings.setPageMargins(margins, LengthUnits.IN);
                        }
                        sheet.rebuild();
                        sheet.setSize(sheet.getPreferredSize());

                        System.out.println(timing);
                        if (text) {
                            System.out.print(I18n.Text("  Creating from text template... "));
                            System.out.flush();
                            textTemplate = TextTemplate.resolveTextTemplate(textTemplate);
                            output       = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), PathUtils.getExtension(textTemplate.getName())));
                            timing.reset();
                            success = new TextTemplate(sheet).export(output, textTemplate);
                            System.out.println(timing);
                            System.out.println(MessageFormat.format(I18n.Text("    Used text template file \"{0}\"."), PathUtils.getFullPath(textTemplate)));
                            if (success) {
                                System.out.println(MessageFormat.format(I18n.Text("    Created \"{0}\"."), output));
                                count++;
                            }
                        }
                        if (pdf) {
                            System.out.print(I18n.Text("  Creating PDF... "));
                            System.out.flush();
                            output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), FileType.PDF_EXTENSION));
                            timing.reset();
                            success = sheet.saveAsPDF(output);
                            System.out.println(timing);
                            if (success) {
                                System.out.println(MessageFormat.format(I18n.Text("    Created \"{0}\"."), output));
                                count++;
                            }
                        }
                        if (png) {
                            ArrayList<File> result = new ArrayList<>();

                            System.out.print(I18n.Text("  Creating PNG... "));
                            System.out.flush();
                            output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), FileType.PNG_EXTENSION));
                            timing.reset();
                            success = sheet.saveAsPNG(output, result);
                            System.out.println(timing);
                            for (File one : result) {
                                System.out.println(MessageFormat.format(I18n.Text("    Created \"{0}\"."), one));
                                count++;
                            }
                        }
                        sheet.dispose();
                    } catch (Exception exception) {
                        exception.printStackTrace();
                        System.out.println(I18n.Text("  ** ERROR ENCOUNTERED **"));
                    }
                }
            }
            GraphicsUtilities.setHeadlessPrintMode(false);
        }
        return count;
    }

    private static double[] getPaperSize(CmdLine cmdLine) {
        if (cmdLine.isOptionUsed(SIZE_OPTION)) {
            String argument = cmdLine.getOptionArgument(SIZE_OPTION);
            int    index;

            if ("LETTER".equalsIgnoreCase(argument)) {
                return new double[] { 8.5, 11 };
            }

            if ("A4".equalsIgnoreCase(argument)) {
                return new double[] { LengthUnits.IN.convert(LengthUnits.CM, 21), LengthUnits.IN.convert(LengthUnits.CM, 29.7) };
            }

            index = argument.indexOf('x');
            if (index == -1) {
                index = argument.indexOf('X');
            }
            if (index != -1) {
                double width  = Numbers.extractDouble(argument.substring(0, index), -1.0, true);
                double height = Numbers.extractDouble(argument.substring(index + 1), -1.0, true);

                if (width > 0.0 && height > 0.0) {
                    return new double[] { width, height };
                }
            }
            System.out.println(I18n.Text("WARNING: Invalid paper size specification."));
        }
        return null;
    }

    private static double[] getMargins(CmdLine cmdLine) {
        if (cmdLine.isOptionUsed(MARGIN_OPTION)) {
            StringTokenizer tokenizer = new StringTokenizer(cmdLine.getOptionArgument(MARGIN_OPTION), ":");
            double[]        values    = new double[4];
            int             index     = 0;

            while (tokenizer.hasMoreTokens()) {
                String token = tokenizer.nextToken();

                if (index < 4) {
                    values[index] = Numbers.extractDouble(token, -1.0, true);
                    if (values[index] < 0.0) {
                        System.out.println(I18n.Text("WARNING: Invalid paper margins specification."));
                        return null;
                    }
                }
                index++;
            }
            if (index == 4) {
                return values;
            }
            System.out.println(I18n.Text("WARNING: Invalid paper margins specification."));
        }
        return null;
    }
}
