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

package com.trollworks.gcs.cmdline;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Timing;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.LengthUnits;

import java.awt.EventQueue;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.StringTokenizer;

public class Export implements Runnable {
    List<Path> mFiles;
    boolean    mGeneratePDF;
    boolean    mGeneratePNG;
    boolean    mGenerateText;
    Path       mTemplate;
    String     mMargins;
    String     mPaper;

    public static void process(List<Path> files, boolean generatePDF, boolean generatePNG, boolean generateText, Path template, String margins, String paper) {
        if (files.isEmpty()) {
            System.err.println(I18n.Text("must specify one or more sheet files to process"));
            System.exit(1);
        }
        System.setProperty("java.awt.headless", Boolean.TRUE.toString());
        UIUtilities.initialize();
        try {
            // This is run on the event queue since much of the sheet logic assumes a UI
            // environment and would otherwise cause concurrent modification exceptions, as the
            // detection of whether it was safe to modify data would be inaccurate.
            EventQueue.invokeAndWait(new Export(files, generatePDF, generatePNG, generateText, template, margins, paper));
        } catch (Exception exception) {
            exception.printStackTrace(System.err);
            System.exit(1);
        }
    }

    private Export(List<Path> files, boolean generatePDF, boolean generatePNG, boolean generateText, Path template, String margins, String paper) {
        mFiles = files;
        mGeneratePDF = generatePDF;
        mGeneratePNG = generatePNG;
        mGenerateText = generateText;
        mTemplate = mGenerateText ? template : null;
        mMargins = margins;
        mPaper = paper;
    }

    public void run() {
        if (mGenerateText || mGeneratePDF || mGeneratePNG) {
            double[] paperSize   = getPaperSize();
            double[] marginsInfo = getMargins();
            Timing   timing      = new Timing();
            GraphicsUtilities.setHeadlessPrintMode(true);
            for (Path path : mFiles) {
                if (!FileType.SHEET.matchExtension(PathUtils.getExtension(path)) || !Files.isReadable(path)) {
                    System.out.printf(I18n.Text("Unable to load %s\n"), path);
                    continue;
                }
                System.out.printf(I18n.Text("Loading %s... "), path);
                System.out.flush();
                timing.reset();
                try {
                    GURPSCharacter character = new GURPSCharacter(path);
                    CharacterSheet sheet     = new CharacterSheet(character);
                    PrintManager   settings  = character.getPageSettings();
                    Path           output;
                    boolean        success;

                    sheet.addNotify(); // Required to allow layout to work
                    sheet.rebuild();
                    sheet.getCharacter().processFeaturesAndPrereqs();

                    if (paperSize != null && settings != null) {
                        settings.setPageSize(paperSize, LengthUnits.IN);
                    }
                    if (marginsInfo != null && settings != null) {
                        settings.setPageMargins(marginsInfo, LengthUnits.IN);
                    }
                    sheet.rebuild();
                    sheet.setSize(sheet.getPreferredSize());

                    System.out.println(timing);
                    if (mGenerateText) {
                        System.out.print(I18n.Text("  Creating from text template... "));
                        System.out.flush();
                        output = path.resolveSibling(PathUtils.enforceExtension(PathUtils.getLeafName(path, false), PathUtils.getExtension(mTemplate)));
                        timing.reset();
                        success = new TextTemplate(sheet).export(output, mTemplate);
                        System.out.println(timing);
                        System.out.printf(I18n.Text("    Used text template file: %s\n"), mTemplate.normalize().toAbsolutePath());
                        if (success) {
                            System.out.printf(I18n.Text("    Created: %s\n"), output);
                        }
                    }
                    if (mGeneratePDF) {
                        System.out.print(I18n.Text("  Creating PDF... "));
                        System.out.flush();
                        output = path.resolveSibling(PathUtils.enforceExtension(PathUtils.getLeafName(path, false), FileType.PDF.getExtension()));
                        timing.reset();
                        success = sheet.saveAsPDF(output);
                        System.out.println(timing);
                        if (success) {
                            System.out.printf(I18n.Text("    Created: %s\n"), output);
                        }
                    }
                    if (mGeneratePNG) {
                        List<Path> result = new ArrayList<>();
                        System.out.print(I18n.Text("  Creating PNG... "));
                        System.out.flush();
                        output = path.resolveSibling(PathUtils.enforceExtension(PathUtils.getLeafName(path, false), FileType.PNG.getExtension()));
                        timing.reset();
                        success = sheet.saveAsPNG(output, result);
                        System.out.println(timing);
                        if (success) {
                            for (Path one : result) {
                                System.out.printf(I18n.Text("    Created: %s\n"), one);
                            }
                        }
                    }
                    sheet.dispose();
                } catch (Exception exception) {
                    exception.printStackTrace();
                    System.out.println(I18n.Text("  ** ERROR ENCOUNTERED **"));
                }
            }
            GraphicsUtilities.setHeadlessPrintMode(false);
        }
    }

    private double[] getPaperSize() {
        if (mPaper != null) {
            int index;

            if ("LETTER".equalsIgnoreCase(mPaper)) {
                return new double[]{8.5, 11};
            }

            if ("A4".equalsIgnoreCase(mPaper)) {
                return new double[]{LengthUnits.IN.convert(LengthUnits.CM, new Fixed6(21)).asDouble(), LengthUnits.IN.convert(LengthUnits.CM, new Fixed6(29.7)).asDouble()};
            }

            index = mPaper.indexOf('x');
            if (index == -1) {
                index = mPaper.indexOf('X');
            }
            if (index != -1) {
                double width  = Numbers.extractDouble(mPaper.substring(0, index), -1.0, true);
                double height = Numbers.extractDouble(mPaper.substring(index + 1), -1.0, true);
                if (width > 0.0 && height > 0.0) {
                    return new double[]{width, height};
                }
            }
            System.err.println(I18n.Text("invalid paper size specification: ") + mPaper);
            System.exit(1);
        }
        return null;
    }

    private double[] getMargins() {
        if (mMargins != null) {
            StringTokenizer tokenizer = new StringTokenizer(mMargins, ":");
            double[]        values    = new double[4];
            int             index     = 0;
            while (tokenizer.hasMoreTokens()) {
                String token = tokenizer.nextToken();
                if (index < 4) {
                    values[index] = Numbers.extractDouble(token, -1.0, true);
                    if (values[index] < 0.0) {
                        break;
                    }
                }
                index++;
            }
            if (index == 4) {
                return values;
            }
            System.err.println(I18n.Text("invalid paper margins specification: ") + mMargins);
            System.exit(1);
        }
        return null;
    }
}
