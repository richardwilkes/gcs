/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.ui.sheet.CSPrerequisitesThread;
import com.trollworks.gcs.ui.sheet.CSSheet;
import com.trollworks.gcs.ui.sheet.CSSheetOpener;
import com.trollworks.gcs.ui.sheet.CSSheetWindow;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.cmdline.TKCmdLine;
import com.trollworks.toolkit.print.TKPrintManager;
import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.TKTiming;
import com.trollworks.toolkit.utility.units.TKLengthUnits;

import java.io.File;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.StringTokenizer;

/** The GCS application object when running in batch mode. */
public class CSBatchApplication extends TKApp {
	/**
	 * Creates a new {@link CSBatchApplication}.
	 * 
	 * @param cmdLine The command line.
	 */
	public CSBatchApplication(TKCmdLine cmdLine) {
		super();

		TKTiming timing = new TKTiming();
		System.out.println(TKApp.getVersionBanner(false));
		System.out.println();
		if (convert(cmdLine) < 1) {
			System.out.println(Msgs.NO_FILES_TO_PROCESS);
			System.exit(1);
		}
		System.out.println(MessageFormat.format(Msgs.FINISHED, timing.toSeconds()));
		System.exit(0);
	}

	private int convert(TKCmdLine cmdLine) {
		boolean html = cmdLine.isOptionUsed(CSMain.HTML_OPTION);
		boolean pdf = cmdLine.isOptionUsed(CSMain.PDF_OPTION);
		boolean png = cmdLine.isOptionUsed(CSMain.PNG_OPTION);
		int count = 0;

		if (html || pdf || png) {
			double[] paperSize = getPaperSize(cmdLine);
			double[] margins = getMargins(cmdLine);
			TKTiming timing = new TKTiming();
			String htmlTemplateOption = cmdLine.getOptionArgument(CSMain.HTML_TEMPLATE_OPTION);
			File htmlTemplate = null;

			if (htmlTemplateOption != null) {
				htmlTemplate = new File(htmlTemplateOption);
			}
			TKGraphics.setHeadlessPrintMode(true);
			for (File file : cmdLine.getArgumentsAsFiles()) {
				if (CSSheetOpener.EXTENSION.equals(TKPath.getExtension(file.getName())) && file.canRead()) {
					System.out.print(MessageFormat.format(Msgs.LOADING, file));
					System.out.flush();
					timing.reset();
					try {
						CMCharacter character = new CMCharacter(file);
						CSSheet sheet = new CSSheet(character);
						CSPrerequisitesThread prereqs = new CSPrerequisitesThread(sheet);
						TKPrintManager settings = character.getPageSettings();
						File output;
						boolean success;

						sheet.addNotify(); // Required to allow layout to work
						sheet.rebuild();
						prereqs.start();
						CSPrerequisitesThread.waitForProcessingToFinish(character);

						if (paperSize != null && settings != null) {
							settings.setPageSize(paperSize, TKLengthUnits.INCHES);
						}
						if (margins != null && settings != null) {
							settings.setPageMargins(margins, TKLengthUnits.INCHES);
						}
						sheet.rebuild();
						sheet.setSize(sheet.getPreferredSize());

						System.out.println(timing.toSeconds());
						if (html) {
							StringBuilder builder = new StringBuilder();

							System.out.print(Msgs.CREATING_HTML);
							System.out.flush();
							output = new File(file.getParentFile(), TKPath.getLeafName(file.getName(), false) + CSSheetWindow.HTML_EXTENSION);
							timing.reset();
							success = sheet.saveAsHTML(output, htmlTemplate, builder);
							System.out.println(timing.toSeconds());
							System.out.println(MessageFormat.format(Msgs.TEMPLATE_USED, builder));
							if (success) {
								System.out.println(MessageFormat.format(Msgs.CREATED, output));
								count++;
							}
						}
						if (pdf) {
							System.out.print(Msgs.CREATING_PDF);
							System.out.flush();
							output = new File(file.getParentFile(), TKPath.getLeafName(file.getName(), false) + CSSheetWindow.PDF_EXTENSION);
							timing.reset();
							success = sheet.saveAsPDF(output);
							System.out.println(timing.toSeconds());
							if (success) {
								System.out.println(MessageFormat.format(Msgs.CREATED, output));
								count++;
							}
						}
						if (png) {
							ArrayList<File> result = new ArrayList<File>();

							System.out.print(Msgs.CREATING_PNG);
							System.out.flush();
							output = new File(file.getParentFile(), TKPath.getLeafName(file.getName(), false) + CSSheetWindow.PNG_EXTENSION);
							timing.reset();
							success = sheet.saveAsPNG(output, result);
							System.out.println(timing.toSeconds());
							for (File one : result) {
								System.out.println(MessageFormat.format(Msgs.CREATED, one));
								count++;
							}
						}
						sheet.dispose();
						character.noLongerNeeded();
					} catch (Exception exception) {
						System.out.println(Msgs.PROCESSING_FAILED);
					}
				}
			}
			TKGraphics.setHeadlessPrintMode(false);
		}
		return count;
	}

	private double[] getPaperSize(TKCmdLine cmdLine) {
		if (cmdLine.isOptionUsed(CSMain.SIZE_OPTION)) {
			String argument = cmdLine.getOptionArgument(CSMain.SIZE_OPTION);
			int index;

			if ("LETTER".equalsIgnoreCase(argument)) { //$NON-NLS-1$
				return new double[] { 8.5, 11 };
			}

			if ("A4".equalsIgnoreCase(argument)) { //$NON-NLS-1$
				return new double[] { TKLengthUnits.INCHES.convert(TKLengthUnits.CENTIMETERS, 21), TKLengthUnits.INCHES.convert(TKLengthUnits.CENTIMETERS, 29.7) };
			}

			index = argument.indexOf('x');
			if (index == -1) {
				index = argument.indexOf('X');
			}
			if (index != -1) {
				double width = TKNumberUtils.getDouble(argument.substring(0, index), -1.0);
				double height = TKNumberUtils.getDouble(argument.substring(index + 1), -1.0);

				if (width > 0.0 && height > 0.0) {
					return new double[] { width, height };
				}
			}
			System.out.println(Msgs.INVALID_PAPER_SIZE);
		}
		return null;
	}

	private double[] getMargins(TKCmdLine cmdLine) {
		if (cmdLine.isOptionUsed(CSMain.MARGIN_OPTION)) {
			StringTokenizer tokenizer = new StringTokenizer(cmdLine.getOptionArgument(CSMain.MARGIN_OPTION), ":"); //$NON-NLS-1$
			double[] values = new double[4];
			int index = 0;

			while (tokenizer.hasMoreTokens()) {
				String token = tokenizer.nextToken();

				if (index < 4) {
					values[index] = TKNumberUtils.getDouble(token, -1.0);
					if (values[index] < 0.0) {
						System.out.println(Msgs.INVALID_PAPER_MARGINS);
						return null;
					}
				}
				index++;
			}
			if (index == 4) {
				return values;
			}
			System.out.println(Msgs.INVALID_PAPER_MARGINS);
		}
		return null;
	}
}
