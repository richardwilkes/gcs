/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.template.Template;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.ui.Fonts;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.menu.file.ExportToCommand;
import com.trollworks.toolkit.ui.print.PrintManager;
import com.trollworks.toolkit.utility.BundleInfo;
import com.trollworks.toolkit.utility.Dice;
import com.trollworks.toolkit.utility.FileProxyCreator;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.LaunchProxy;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.Timing;
import com.trollworks.toolkit.utility.cmdline.CmdLine;
import com.trollworks.toolkit.utility.cmdline.CmdLineOption;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.units.LengthUnits;

import java.io.File;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.StringTokenizer;

/** The main entry point for the character sheet. */
public class GCS {
	@Localize("GURPS Character Sheet")
	private static String				SHEET_DESCRIPTION;
	@Localize("GCS Advantages Library")
	private static String				ADVANTAGES_LIBRARY_DESCRIPTION;
	@Localize("GCS Equipment Library")
	private static String				EQUIPMENT_LIBRARY_DESCRIPTION;
	@Localize("GCS Skills Library")
	private static String				SKILLS_LIBRARY_DESCRIPTION;
	@Localize("GCS Spells Library")
	private static String				SPELLS_LIBRARY_DESCRIPTION;
	@Localize("GCS Library")
	private static String				LIBRARY_DESCRIPTION;
	@Localize("GCS Character Template")
	private static String				TEMPLATE_DESCRIPTION;
	@Localize("Create PDF versions of sheets specified on the command line.")
	private static String				PDF_OPTION_DESCRIPTION;
	@Localize("Create HTML versions of sheets specified on the command line.")
	private static String				HTML_OPTION_DESCRIPTION;
	@Localize("A template to use when creating HTML versions of the sheets. If this is not specified, then the template found in the data directory will be used by default.")
	private static String				HTML_TEMPLATE_OPTION_DESCRIPTION;
	@Localize("FILE")
	private static String				HTML_TEMPLATE_ARG;
	@Localize("Create PNG versions of sheets specified on the command line.")
	private static String				PNG_OPTION_DESCRIPTION;
	@Localize("When generating PDF or PNG from the command line, allows you to specify a paper size to use, rather than the one embedded in the file. Valid choices are: LETTER, A4, or the width and height, expressed in inches and separated by an 'x', such as '5x7'.")
	private static String				SIZE_OPTION_DESCRIPTION;
	@Localize("When generating PDF or PNG from the command line, allows you to specify the margins to use, rather than the ones embedded in the file. The top, left, bottom, and right margins must all be specified in inches, separated by colons, such as '1:1:1:1'.")
	private static String				MARGIN_OPTION_DESCRIPTION;
	@Localize("You must specify one or more sheet files to process.")
	private static String				NO_FILES_TO_PROCESS;
	@Localize("Loading \"{0}\"... ")
	private static String				LOADING;
	@Localize("  Creating PDF... ")
	private static String				CREATING_PDF;
	@Localize("  Creating HMTL... ")
	private static String				CREATING_HTML;
	@Localize("  Creating PNG... ")
	private static String				CREATING_PNG;
	@Localize("  ** ERROR ENCOUNTERED **")
	private static String				PROCESSING_FAILED;
	@Localize("\nDone! {0} overall.")
	private static String				FINISHED;
	@Localize("    Created \"{0}\".")
	private static String				CREATED;
	@Localize("WARNING: Invalid paper size specification.")
	private static String				INVALID_PAPER_SIZE;
	@Localize("WARNING: Invalid paper margins specification.")
	private static String				INVALID_PAPER_MARGINS;
	@Localize("    Used template file \"{0}\".")
	private static String				TEMPLATE_USED;

	static {
		Localization.initialize();
	}

	private static final CmdLineOption	PDF_OPTION				= new CmdLineOption(PDF_OPTION_DESCRIPTION, null, "pdf");									//$NON-NLS-1$
	private static final CmdLineOption	HTML_OPTION				= new CmdLineOption(HTML_OPTION_DESCRIPTION, null, "html");								//$NON-NLS-1$
	private static final CmdLineOption	HTML_TEMPLATE_OPTION	= new CmdLineOption(HTML_TEMPLATE_OPTION_DESCRIPTION, HTML_TEMPLATE_ARG, "html_template");	//$NON-NLS-1$
	private static final CmdLineOption	PNG_OPTION				= new CmdLineOption(PNG_OPTION_DESCRIPTION, null, "png");									//$NON-NLS-1$
	private static final CmdLineOption	SIZE_OPTION				= new CmdLineOption(SIZE_OPTION_DESCRIPTION, "SIZE", "paper");								//$NON-NLS-1$ //$NON-NLS-2$
	private static final CmdLineOption	MARGIN_OPTION			= new CmdLineOption(MARGIN_OPTION_DESCRIPTION, "MARGINS", "margins");						//$NON-NLS-1$ //$NON-NLS-2$
	private static final String			REFERENCE_URL			= "http://gcs.trollworks.com";																//$NON-NLS-1$

	/**
	 * The main entry point for the character sheet.
	 *
	 * @param args Arguments to the program.
	 */
	public static void main(String[] args) {
		App.setup(GCS.class);
		Dice.setAssumedSideCount(6);
		CmdLine cmdLine = new CmdLine();
		cmdLine.addOptions(HTML_OPTION, HTML_TEMPLATE_OPTION, PDF_OPTION, PNG_OPTION, SIZE_OPTION, MARGIN_OPTION);
		cmdLine.processArguments(args);
		if (cmdLine.isOptionUsed(HTML_OPTION) || cmdLine.isOptionUsed(PDF_OPTION) || cmdLine.isOptionUsed(PNG_OPTION)) {
			System.setProperty("java.awt.headless", Boolean.TRUE.toString()); //$NON-NLS-1$
			initialize();
			Timing timing = new Timing();
			System.out.println(BundleInfo.getDefault().getAppBanner());
			System.out.println();
			if (convert(cmdLine) < 1) {
				System.out.println(NO_FILES_TO_PROCESS);
				System.exit(1);
			}
			System.out.println(MessageFormat.format(FINISHED, timing));
			System.exit(0);
		} else {
			LaunchProxy.configure(cmdLine.getArgumentsAsFiles().toArray(new File[0]));
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
		Preferences.setPreferenceFile("gcs.pref"); //$NON-NLS-1$
		GCSFonts.register();
		Fonts.loadFromPreferences();
		App.setAboutPanel(AboutPanel.class);
		registerFileTypes(new GCSFileProxyCreator());
	}

	/**
	 * Registers the file types the app can open.
	 *
	 * @param fileProxyCreator The {@link FileProxyCreator} to use.
	 */
	public static void registerFileTypes(FileProxyCreator fileProxyCreator) {
		FileType.register(GURPSCharacter.EXTENSION, GCSImages.getCharacterSheetDocumentIcons(), SHEET_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true);
		FileType.register(AdvantageList.EXTENSION, GCSImages.getAdvantagesDocumentIcons(), ADVANTAGES_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true);
		FileType.register(EquipmentList.EXTENSION, GCSImages.getEquipmentDocumentIcons(), EQUIPMENT_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true);
		FileType.register(SkillList.EXTENSION, GCSImages.getSkillsDocumentIcons(), SKILLS_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true);
		FileType.register(SpellList.EXTENSION, GCSImages.getSpellsDocumentIcons(), SPELLS_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true);
		FileType.register(Template.EXTENSION, GCSImages.getTemplateDocumentIcons(), TEMPLATE_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true);
		// For legacy
		FileType.register(LibraryFile.EXTENSION, GCSImages.getAdvantagesDocumentIcons(), LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true);
	}

	private static int convert(CmdLine cmdLine) {
		boolean html = cmdLine.isOptionUsed(GCS.HTML_OPTION);
		boolean pdf = cmdLine.isOptionUsed(GCS.PDF_OPTION);
		boolean png = cmdLine.isOptionUsed(GCS.PNG_OPTION);
		int count = 0;

		if (html || pdf || png) {
			double[] paperSize = getPaperSize(cmdLine);
			double[] margins = getMargins(cmdLine);
			Timing timing = new Timing();
			String htmlTemplateOption = cmdLine.getOptionArgument(GCS.HTML_TEMPLATE_OPTION);
			File htmlTemplate = null;

			if (htmlTemplateOption != null) {
				htmlTemplate = new File(htmlTemplateOption);
			}
			GraphicsUtilities.setHeadlessPrintMode(true);
			for (File file : cmdLine.getArgumentsAsFiles()) {
				if (GURPSCharacter.EXTENSION.equals(PathUtils.getExtension(file.getName())) && file.canRead()) {
					System.out.print(MessageFormat.format(LOADING, file));
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
							settings.setPageSize(paperSize, LengthUnits.IN);
						}
						if (margins != null && settings != null) {
							settings.setPageMargins(margins, LengthUnits.IN);
						}
						sheet.rebuild();
						sheet.setSize(sheet.getPreferredSize());

						System.out.println(timing);
						if (html) {
							StringBuilder builder = new StringBuilder();

							System.out.print(CREATING_HTML);
							System.out.flush();
							output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), ExportToCommand.HTML_EXTENSION));
							timing.reset();
							success = sheet.saveAsHTML(output, htmlTemplate, builder);
							System.out.println(timing);
							System.out.println(MessageFormat.format(TEMPLATE_USED, builder));
							if (success) {
								System.out.println(MessageFormat.format(CREATED, output));
								count++;
							}
						}
						if (pdf) {
							System.out.print(CREATING_PDF);
							System.out.flush();
							output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), ExportToCommand.PDF_EXTENSION));
							timing.reset();
							success = sheet.saveAsPDF(output);
							System.out.println(timing);
							if (success) {
								System.out.println(MessageFormat.format(CREATED, output));
								count++;
							}
						}
						if (png) {
							ArrayList<File> result = new ArrayList<>();

							System.out.print(CREATING_PNG);
							System.out.flush();
							output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), ExportToCommand.PNG_EXTENSION));
							timing.reset();
							success = sheet.saveAsPNG(output, result);
							System.out.println(timing);
							for (File one : result) {
								System.out.println(MessageFormat.format(CREATED, one));
								count++;
							}
						}
						sheet.dispose();
					} catch (Exception exception) {
						exception.printStackTrace();
						System.out.println(PROCESSING_FAILED);
					}
				}
			}
			GraphicsUtilities.setHeadlessPrintMode(false);
		}
		return count;
	}

	private static double[] getPaperSize(CmdLine cmdLine) {
		if (cmdLine.isOptionUsed(GCS.SIZE_OPTION)) {
			String argument = cmdLine.getOptionArgument(GCS.SIZE_OPTION);
			int index;

			if ("LETTER".equalsIgnoreCase(argument)) { //$NON-NLS-1$
				return new double[] { 8.5, 11 };
			}

			if ("A4".equalsIgnoreCase(argument)) { //$NON-NLS-1$
				return new double[] { LengthUnits.IN.convert(LengthUnits.CM, 21), LengthUnits.IN.convert(LengthUnits.CM, 29.7) };
			}

			index = argument.indexOf('x');
			if (index == -1) {
				index = argument.indexOf('X');
			}
			if (index != -1) {
				double width = Numbers.getLocalizedDouble(argument.substring(0, index), -1.0);
				double height = Numbers.getLocalizedDouble(argument.substring(index + 1), -1.0);

				if (width > 0.0 && height > 0.0) {
					return new double[] { width, height };
				}
			}
			System.out.println(INVALID_PAPER_SIZE);
		}
		return null;
	}

	private static double[] getMargins(CmdLine cmdLine) {
		if (cmdLine.isOptionUsed(GCS.MARGIN_OPTION)) {
			StringTokenizer tokenizer = new StringTokenizer(cmdLine.getOptionArgument(GCS.MARGIN_OPTION), ":"); //$NON-NLS-1$
			double[] values = new double[4];
			int index = 0;

			while (tokenizer.hasMoreTokens()) {
				String token = tokenizer.nextToken();

				if (index < 4) {
					values[index] = Numbers.getLocalizedDouble(token, -1.0);
					if (values[index] < 0.0) {
						System.out.println(INVALID_PAPER_MARGINS);
						return null;
					}
				}
				index++;
			}
			if (index == 4) {
				return values;
			}
			System.out.println(INVALID_PAPER_MARGINS);
		}
		return null;
	}
}
