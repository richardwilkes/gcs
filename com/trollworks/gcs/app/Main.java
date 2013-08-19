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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.app;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.PrerequisitesThread;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.ttk.cmdline.CmdLine;
import com.trollworks.ttk.cmdline.CmdLineOption;
import com.trollworks.ttk.menu.file.FileType;
import com.trollworks.ttk.preferences.Preferences;
import com.trollworks.ttk.print.PrintManager;
import com.trollworks.ttk.text.NumberUtils;
import com.trollworks.ttk.units.LengthUnits;
import com.trollworks.ttk.utility.App;
import com.trollworks.ttk.utility.Fonts;
import com.trollworks.ttk.utility.GraphicsUtilities;
import com.trollworks.ttk.utility.LaunchProxy;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.Path;
import com.trollworks.ttk.utility.Timing;

import java.io.File;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.StringTokenizer;

import javax.swing.UIManager;

/** The main entry point for the character sheet. */
public class Main {
	private static String				MSG_APP_NAME;
	private static String				MSG_APP_VERSION;
	private static String				MSG_APP_COPYRIGHT_YEARS;
	private static String				MSG_APP_COPYRIGHT_OWNER;
	private static String				MSG_LONG_VERSION_FORMAT;
	private static String				MSG_COPYRIGHT_FORMAT;
	private static String				MSG_PDF_OPTION;
	private static String				MSG_HTML_OPTION;
	private static String				MSG_HTML_TEMPLATE_OPTION;
	private static String				MSG_HTML_TEMPLATE_ARG;
	private static String				MSG_PNG_OPTION;
	private static String				MSG_SIZE_OPTION;
	private static String				MSG_MARGIN_OPTION;
	private static String				MSG_NO_FILES_TO_PROCESS;
	private static String				MSG_LOADING;
	private static String				MSG_CREATING_PDF;
	private static String				MSG_CREATING_HTML;
	private static String				MSG_CREATING_PNG;
	private static String				MSG_PROCESSING_FAILED;
	private static String				MSG_FINISHED;
	private static String				MSG_CREATED;
	private static String				MSG_INVALID_PAPER_SIZE;
	private static String				MSG_INVALID_PAPER_MARGINS;
	private static String				MSG_TEMPLATE_USED;

	static {
		LocalizedMessages.initialize(Main.class);
		App.setName(MSG_APP_NAME);
		App.setVersion(MSG_APP_VERSION);
	}

	private static final CmdLineOption	PDF_OPTION				= new CmdLineOption(MSG_PDF_OPTION, null, "pdf");										//$NON-NLS-1$
	private static final CmdLineOption	HTML_OPTION				= new CmdLineOption(MSG_HTML_OPTION, null, "html");									//$NON-NLS-1$
	private static final CmdLineOption	HTML_TEMPLATE_OPTION	= new CmdLineOption(MSG_HTML_TEMPLATE_OPTION, MSG_HTML_TEMPLATE_ARG, "html_template");	//$NON-NLS-1$
	private static final CmdLineOption	PNG_OPTION				= new CmdLineOption(MSG_PNG_OPTION, null, "png");										//$NON-NLS-1$
	private static final CmdLineOption	SIZE_OPTION				= new CmdLineOption(MSG_SIZE_OPTION, "SIZE", "paper");									//$NON-NLS-1$ //$NON-NLS-2$
	private static final CmdLineOption	MARGIN_OPTION			= new CmdLineOption(MSG_MARGIN_OPTION, "MARGINS", "margins");							//$NON-NLS-1$ //$NON-NLS-2$

	/**
	 * The main entry point for the character sheet.
	 * 
	 * @param args Arguments to the program. None used at the moment.
	 */
	public static void main(String[] args) {
		ArrayList<CmdLineOption> options = new ArrayList<CmdLineOption>();
		options.add(HTML_OPTION);
		options.add(HTML_TEMPLATE_OPTION);
		options.add(PDF_OPTION);
		options.add(PNG_OPTION);
		options.add(SIZE_OPTION);
		options.add(MARGIN_OPTION);

		String versionBanner = getVersionBanner(false);
		CmdLine cmdLine = new CmdLine(args, options, versionBanner);
		if (cmdLine.isOptionUsed(HTML_OPTION) || cmdLine.isOptionUsed(PDF_OPTION) || cmdLine.isOptionUsed(PNG_OPTION)) {
			System.setProperty("java.awt.headless", Boolean.TRUE.toString()); //$NON-NLS-1$
			initialize();
			Timing timing = new Timing();
			System.out.println(versionBanner);
			System.out.println();
			if (convert(cmdLine) < 1) {
				System.out.println(MSG_NO_FILES_TO_PROCESS);
				System.exit(1);
			}
			System.out.println(MessageFormat.format(MSG_FINISHED, timing));
			System.exit(0);
		} else {
			LaunchProxy.configure("GCSLaunchProxy", 1, cmdLine.getArgumentsAsFiles().toArray(new File[0])); //$NON-NLS-1$
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
		System.setProperty("apple.laf.useScreenMenuBar", Boolean.TRUE.toString()); //$NON-NLS-1$
		// System.setProperty("apple.awt.brushMetalLook", Boolean.TRUE.toString());
		// System.setProperty("apple.awt.showGrowBox", Boolean.FALSE.toString());
		try {
			UIManager.setLookAndFeel(UIManager.getSystemLookAndFeelClassName());
		} catch (Exception ex) {
			ex.printStackTrace(System.err);
		}
		Preferences.setPreferenceFile("gcs.pref"); //$NON-NLS-1$
		GCSFonts.register();
		Fonts.loadFromPreferences();
		App.setAboutPanel(AboutPanel.class);
		FileType.register(SheetWindow.SHEET_EXTENSION, GCSImages.getCharacterSheetIcon(false), SheetWindow.class, true);
		FileType.register(LibraryFile.EXTENSION, GCSImages.getLibraryIcon(false), LibraryWindow.class, true);
		FileType.register(TemplateWindow.EXTENSION, GCSImages.getTemplateIcon(false), TemplateWindow.class, true);
		FileType.register(Advantage.OLD_ADVANTAGE_EXTENSION, GCSImages.getAdvantageIcon(false, false), LibraryWindow.class, true);
		FileType.register(Equipment.OLD_EQUIPMENT_EXTENSION, GCSImages.getEquipmentIcon(false, false), LibraryWindow.class, true);
		FileType.register(Skill.OLD_SKILL_EXTENSION, GCSImages.getSkillIcon(false, false), LibraryWindow.class, true);
		FileType.register(Spell.OLD_SPELL_EXTENSION, GCSImages.getSpellIcon(false, false), LibraryWindow.class, true);
	}

	private static int convert(CmdLine cmdLine) {
		boolean html = cmdLine.isOptionUsed(Main.HTML_OPTION);
		boolean pdf = cmdLine.isOptionUsed(Main.PDF_OPTION);
		boolean png = cmdLine.isOptionUsed(Main.PNG_OPTION);
		int count = 0;

		if (html || pdf || png) {
			double[] paperSize = getPaperSize(cmdLine);
			double[] margins = getMargins(cmdLine);
			Timing timing = new Timing();
			String htmlTemplateOption = cmdLine.getOptionArgument(Main.HTML_TEMPLATE_OPTION);
			File htmlTemplate = null;

			if (htmlTemplateOption != null) {
				htmlTemplate = new File(htmlTemplateOption);
			}
			GraphicsUtilities.setHeadlessPrintMode(true);
			for (File file : cmdLine.getArgumentsAsFiles()) {
				if (SheetWindow.SHEET_EXTENSION.equals(Path.getExtension(file.getName())) && file.canRead()) {
					System.out.print(MessageFormat.format(MSG_LOADING, file));
					System.out.flush();
					timing.reset();
					try {
						GURPSCharacter character = new GURPSCharacter(file);
						CharacterSheet sheet = new CharacterSheet(character);
						PrerequisitesThread prereqs = new PrerequisitesThread(sheet);
						PrintManager settings = character.getPageSettings();
						File output;
						boolean success;

						sheet.addNotify(); // Required to allow layout to work
						sheet.rebuild();
						prereqs.start();
						PrerequisitesThread.waitForProcessingToFinish(character);

						if (paperSize != null && settings != null) {
							settings.setPageSize(paperSize, LengthUnits.INCHES);
						}
						if (margins != null && settings != null) {
							settings.setPageMargins(margins, LengthUnits.INCHES);
						}
						sheet.rebuild();
						sheet.setSize(sheet.getPreferredSize());

						System.out.println(timing);
						if (html) {
							StringBuilder builder = new StringBuilder();

							System.out.print(MSG_CREATING_HTML);
							System.out.flush();
							output = new File(file.getParentFile(), Path.getLeafName(file.getName(), false) + SheetWindow.HTML_EXTENSION);
							timing.reset();
							success = sheet.saveAsHTML(output, htmlTemplate, builder);
							System.out.println(timing);
							System.out.println(MessageFormat.format(MSG_TEMPLATE_USED, builder));
							if (success) {
								System.out.println(MessageFormat.format(MSG_CREATED, output));
								count++;
							}
						}
						if (pdf) {
							System.out.print(MSG_CREATING_PDF);
							System.out.flush();
							output = new File(file.getParentFile(), Path.getLeafName(file.getName(), false) + SheetWindow.PDF_EXTENSION);
							timing.reset();
							success = sheet.saveAsPDF(output);
							System.out.println(timing);
							if (success) {
								System.out.println(MessageFormat.format(MSG_CREATED, output));
								count++;
							}
						}
						if (png) {
							ArrayList<File> result = new ArrayList<File>();

							System.out.print(MSG_CREATING_PNG);
							System.out.flush();
							output = new File(file.getParentFile(), Path.getLeafName(file.getName(), false) + SheetWindow.PNG_EXTENSION);
							timing.reset();
							success = sheet.saveAsPNG(output, result);
							System.out.println(timing);
							for (File one : result) {
								System.out.println(MessageFormat.format(MSG_CREATED, one));
								count++;
							}
						}
						sheet.dispose();
					} catch (Exception exception) {
						exception.printStackTrace();
						System.out.println(MSG_PROCESSING_FAILED);
					}
				}
			}
			GraphicsUtilities.setHeadlessPrintMode(false);
		}
		return count;
	}

	private static double[] getPaperSize(CmdLine cmdLine) {
		if (cmdLine.isOptionUsed(Main.SIZE_OPTION)) {
			String argument = cmdLine.getOptionArgument(Main.SIZE_OPTION);
			int index;

			if ("LETTER".equalsIgnoreCase(argument)) { //$NON-NLS-1$
				return new double[] { 8.5, 11 };
			}

			if ("A4".equalsIgnoreCase(argument)) { //$NON-NLS-1$
				return new double[] { LengthUnits.INCHES.convert(LengthUnits.CENTIMETERS, 21), LengthUnits.INCHES.convert(LengthUnits.CENTIMETERS, 29.7) };
			}

			index = argument.indexOf('x');
			if (index == -1) {
				index = argument.indexOf('X');
			}
			if (index != -1) {
				double width = NumberUtils.getDouble(argument.substring(0, index), -1.0);
				double height = NumberUtils.getDouble(argument.substring(index + 1), -1.0);

				if (width > 0.0 && height > 0.0) {
					return new double[] { width, height };
				}
			}
			System.out.println(MSG_INVALID_PAPER_SIZE);
		}
		return null;
	}

	private static double[] getMargins(CmdLine cmdLine) {
		if (cmdLine.isOptionUsed(Main.MARGIN_OPTION)) {
			StringTokenizer tokenizer = new StringTokenizer(cmdLine.getOptionArgument(Main.MARGIN_OPTION), ":"); //$NON-NLS-1$
			double[] values = new double[4];
			int index = 0;

			while (tokenizer.hasMoreTokens()) {
				String token = tokenizer.nextToken();

				if (index < 4) {
					values[index] = NumberUtils.getDouble(token, -1.0);
					if (values[index] < 0.0) {
						System.out.println(MSG_INVALID_PAPER_MARGINS);
						return null;
					}
				}
				index++;
			}
			if (index == 4) {
				return values;
			}
			System.out.println(MSG_INVALID_PAPER_MARGINS);
		}
		return null;
	}

	/**
	 * @param useRealCopyrightSymbol Whether or not a real copyright symbol should be used.
	 * @return A version banner for this application.
	 */
	public static String getVersionBanner(boolean useRealCopyrightSymbol) {
		return MessageFormat.format(MSG_LONG_VERSION_FORMAT, App.getName(), App.getVersion(), getCopyrightBanner(useRealCopyrightSymbol));
	}

	/**
	 * @param useRealCopyrightSymbol Whether or not a real copyright symbol should be used.
	 * @return A copyright banner for this application.
	 */
	public static String getCopyrightBanner(boolean useRealCopyrightSymbol) {
		String banner = MessageFormat.format(MSG_COPYRIGHT_FORMAT, MSG_APP_COPYRIGHT_YEARS, MSG_APP_COPYRIGHT_OWNER);

		if (useRealCopyrightSymbol) {
			banner = banner.replaceAll("\\(c\\)", "\u00A9"); //$NON-NLS-1$ //$NON-NLS-2$
		}
		return banner;
	}
}
