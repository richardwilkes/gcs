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

package com.trollworks.gcs.app;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.PrerequisitesThread;
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Timing;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.LengthUnits;

import java.io.File;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.StringTokenizer;

public class CmdLineExport implements Runnable {
    List<Path> mFiles;
    boolean    mGeneratePDF;
    boolean    mGeneratePNG;
    boolean    mGenerateText;
    Path       mTemplate;
    String     mMargins;
    String     mPaper;

    public CmdLineExport(List<Path> files, boolean generatePDF, boolean generatePNG, boolean generateText, Path template, String margins, String paper) {
        mFiles = files;
        mGeneratePDF = generatePDF;
        mGeneratePNG = generatePNG;
        mGenerateText = generateText;
        mTemplate = template;
        mMargins = margins;
        mPaper = paper;
    }

    public void run() {
        if (mGenerateText || mGeneratePDF || mGeneratePNG) {
            double[] paperSize    = getPaperSize();
            double[] marginsInfo  = getMargins();
            Timing   timing       = new Timing();
            File     textTemplate = null;
            if (mTemplate != null) {
                textTemplate = mTemplate.toFile();
            }
            GraphicsUtilities.setHeadlessPrintMode(true);
            for (Path path : mFiles) {
                File file = path.toFile();
                if (FileType.SHEET.matchExtension(PathUtils.getExtension(file.getName())) && file.canRead()) {
                    System.out.printf(I18n.Text("Loading %s... "), file);
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
                        if (marginsInfo != null && settings != null) {
                            settings.setPageMargins(marginsInfo, LengthUnits.IN);
                        }
                        sheet.rebuild();
                        sheet.setSize(sheet.getPreferredSize());

                        System.out.println(timing);
                        if (mGenerateText) {
                            System.out.print(I18n.Text("  Creating from text template... "));
                            System.out.flush();
                            textTemplate = TextTemplate.resolveTextTemplate(textTemplate);
                            output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), PathUtils.getExtension(textTemplate.getName())));
                            timing.reset();
                            success = new TextTemplate(sheet).export(output, textTemplate);
                            System.out.println(timing);
                            System.out.printf(I18n.Text("    Used text template file: %s\n"), PathUtils.getFullPath(textTemplate));
                            if (success) {
                                System.out.printf(I18n.Text("    Created: %s\n"), output);
                            }
                        }
                        if (mGeneratePDF) {
                            System.out.print(I18n.Text("  Creating PDF... "));
                            System.out.flush();
                            output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), FileType.PDF.getExtension()));
                            timing.reset();
                            success = sheet.saveAsPDF(output);
                            System.out.println(timing);
                            if (success) {
                                System.out.printf(I18n.Text("    Created: %s\n"), output);
                            }
                        }
                        if (mGeneratePNG) {
                            List<File> result = new ArrayList<>();
                            System.out.print(I18n.Text("  Creating PNG... "));
                            System.out.flush();
                            output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), FileType.PNG.getExtension()));
                            timing.reset();
                            success = sheet.saveAsPNG(output, result);
                            System.out.println(timing);
                            if (success) {
                                for (File one : result) {
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
                return new double[]{LengthUnits.IN.convert(LengthUnits.CM, 21), LengthUnits.IN.convert(LengthUnits.CM, 29.7)};
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
