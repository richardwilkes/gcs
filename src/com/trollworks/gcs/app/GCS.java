/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.PrerequisitesThread;
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.template.Template;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.ui.Fonts;
import com.trollworks.toolkit.ui.GraphicsUtilities;
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
import java.nio.file.Path;
import java.nio.file.Paths;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.StringTokenizer;

/** The main entry point for the character sheet. */
public class GCS {
	@Localize("GURPS Character Sheet")
	@Localize(locale = "de", value = "GURPS Charakterblatt")
	@Localize(locale = "ru", value = "GURPS персонаж")
	@Localize(locale = "es", value = "Ficha de Personaje GURPS")
	@Localize(locale = "pt-BR", value = "Ficha de Personagem GURPS")
	private static String	SHEET_DESCRIPTION;
	@Localize("GCS Advantages Library")
	@Localize(locale = "de", value = "GCS Vorteils-Liste")
	@Localize(locale = "ru", value = "GCS библиотека преимуществ")
	@Localize(locale = "es", value = "Biblioteca de Ventajas CGS")
	@Localize(locale = "pt-BR", value = "Biblioteca de Vantagens CGS")
	private static String	ADVANTAGES_LIBRARY_DESCRIPTION;
	@Localize("GCS Equipment Library")
	@Localize(locale = "de", value = "GCS Ausrüstungs-Liste")
	@Localize(locale = "ru", value = "GCS библиотека снаряжения")
	@Localize(locale = "es", value = "Biblioteca de Equipo GCS")
	@Localize(locale = "pt-BR", value = "Biblioteca de Equipamentos GCS")
	private static String	EQUIPMENT_LIBRARY_DESCRIPTION;
	@Localize("GCS Skills Library")
	@Localize(locale = "de", value = "GCS Fertigkeiten-Liste")
	@Localize(locale = "ru", value = "GCS библиотека умений")
	@Localize(locale = "es", value = "Biblioteca de Habilidades GCS")
	@Localize(locale = "pt-BR", value = "Biblioteca de Habilidades GCS")
	private static String	SKILLS_LIBRARY_DESCRIPTION;
	@Localize("GCS Spells Library")
	@Localize(locale = "de", value = "GCS Zauber-Liste")
	@Localize(locale = "ru", value = "GCS библиотека заклинаний")
	@Localize(locale = "es", value = "Biblioteca de Sortilegios GCS")
	@Localize(locale = "pt-BR", value = "Biblioteca de Magias GCS")
	private static String	SPELLS_LIBRARY_DESCRIPTION;
	@Localize("GCS Library")
	@Localize(locale = "de", value = "GCS Listen")
	@Localize(locale = "ru", value = "GCS библиотека")
	@Localize(locale = "es", value = "Biblioteca GCS")
	@Localize(locale = "pt-BR", value = "Biblioteca GCS")
	private static String	LIBRARY_DESCRIPTION;
	@Localize("GCS Character Template")
	@Localize(locale = "de", value = "GCS Charaktervorlage")
	@Localize(locale = "ru", value = "GCS шаблон персонажа")
	@Localize(locale = "es", value = "Plantilla de Personaje GCS")
	@Localize(locale = "pt-BR", value = "Modelo de Personagem GCS")
	private static String	TEMPLATE_DESCRIPTION;
	@Localize("Create PDF versions of sheets specified on the command line.")
	@Localize(locale = "de", value = "Erstelle PDF-Dateien von den auf der Kommandozeile angegebenen Charakterblättern.")
	@Localize(locale = "ru", value = "Создать PDF-версии листов, указанных в командной строке.")
	@Localize(locale = "es", value = "Crear archivo PDF de las hojas especificadas en la linea de mandatos")
	@Localize(locale = "pt-BR", value = "Criar versão em PDF da ficha específicada em linha de comando.")
	private static String	PDF_OPTION_DESCRIPTION;
	@Localize("Create text versions of sheets specified on the command line.")
	@Localize(locale = "pt-BR", value = "Criar versão em texto da ficha específicada em linha de comando.")
	private static String	TEXT_OPTION_DESCRIPTION;
	@Localize("A template to use when creating text versions of the sheets. If this is not specified, then the template found in the data directory will be used by default.")
	@Localize(locale = "pt-BR", value = "Um modelo a ser usado ao criar versões de texto das fichas. Se não for especificado, o modelo encontrado no diretório de dados será usado por padrão.")
	private static String	TEXT_TEMPLATE_OPTION_DESCRIPTION;
	@Localize("FILE")
	@Localize(locale = "de", value = "DATEI")
	@Localize(locale = "ru", value = "ФАЙЛ")
	@Localize(locale = "es", value = "Archivo")
	@Localize(locale = "pt-BR", value = "Arquivo")
	private static String	TEXT_TEMPLATE_ARG;
	@Localize("Create PNG versions of sheets specified on the command line.")
	@Localize(locale = "de", value = "Erstelle PNG-Dateien von den auf der Kommandozeile angegebenen Charakterblättern.")
	@Localize(locale = "ru", value = "Создать PNG-версии листов, указанных в командной строке.")
	@Localize(locale = "es", value = "Crear archivo PNG de las hojas especificadas en la linea de mandatos")
	@Localize(locale = "pt-BR", value = "Criar versão em PNG da ficha específicadas em linha de comando.")
	private static String	PNG_OPTION_DESCRIPTION;
	@Localize("When generating PDF or PNG from the command line, allows you to specify a paper size to use, rather than the one embedded in the file. Valid choices are: LETTER, A4, or the width and height, expressed in inches and separated by an 'x', such as '5x7'.")
	@Localize(locale = "de", value = "Erlaubt, bei der Erstellung von PDF-Dateien oder PNG-Dateien von der Kommandozeile aus, das zu verwendende Papierformat anzugeben, statt das im Charakterblatt angegebene zu verwenden. Gültige Werte sind: LETTER, A4 oder die Höhe und Breite in Zoll und mit einem 'x' getrennt, wie z.B. '5x7'.")
	@Localize(locale = "ru", value = "При создании PDF или PNG из командной строки, позволяет указывать размер бумаги, отличный от того, что содержится в файле. Можно выбрать следующие: Letter, A4 или ширину и высоту в дюймах и разделенную 'х', например '5x7'.")
	@Localize(locale = "es", value = "Cuando se genera en formato PDF o PNG desde la linea de mandatos,se permite especificar el tamaño de papel, en lugar del valor contenido en el archivo. Las elecciones válidas son: LETTER, A4, o ancho y alto en pulgadas separado por una 'x', como por ejemplo '5x7'.")
	@Localize(locale = "pt-BR", value = "Quando gerar PDFs ou PNGs a partir da linha de comando, se permite especificar o tamanho do papel, ao invés dos valores do arquivo. As opções válidas são: CARTA, A4, ou largura e altura, em polegadas e separadas por um 'x', como por exemplo '5x7'.")
	private static String	SIZE_OPTION_DESCRIPTION;
	@Localize("When generating PDF or PNG from the command line, allows you to specify the margins to use, rather than the ones embedded in the file. The top, left, bottom, and right margins must all be specified in inches, separated by colons, such as '1:1:1:1'.")
	@Localize(locale = "de", value = "Erlaubt, bei der Erstellung von PDF-Dateien oder PNG-Dateien von der Kommandozeile aus, die zu verwendenden Ränder anzugeben, statt die im Charakterblatt angegebenen zu verwenden. Der obere, linke, untere und rechte Rand müssen alle in Zoll angegeben werden, getrennt durch Doppelpunke, wie z.B. '1:1:1:1'.")
	@Localize(locale = "ru", value = "При создании PDF или PNG из командной строки, позволяет указать размеры отступов, отличные от тех, что содержатся в файле. Сверху, слева, снизу, и справа должны быть указаны в дюймах, разделенные двоеточиями, например '1:1:1:1'.")
	@Localize(locale = "es", value = "Cuando se genera en formato PDF o PNG desde la linea de mandatos,se permite especificar las dimensiones de los márgenes, en lugar de los valores contenidos en el archivo. Las medidas de los márgenes superior, izquierdo, inferior y derecho, deben especificarse en pulgadas, separadas por dos puntos, como por ejemplo '1:1:1:1' .")
	@Localize(locale = "pt-BR", value = "Quando gerar PDFs ou PNGs a partir da linha de comandos, se permite especificar as margens para uso, ao invés dos valores do arquivo. As margens do topo, esquerda, baixo, e direita, precisam ser especificadas em polegadas, separadas por dois pontos, como por exemplo '1:1:1:1' .")
	private static String	MARGIN_OPTION_DESCRIPTION;
	@Localize("You must specify one or more sheet files to process.")
	@Localize(locale = "de", value = "Es muss mindestens ein Charakterblatt angegeben werden.")
	@Localize(locale = "ru", value = "Вы должны указать один или несколько файлов листов для обработки.")
	@Localize(locale = "es", value = "Debes seleccionar uno o más archivos de Ficha de Personaje para procesarlas.")
	@Localize(locale = "pt-BR", value = "Selecione uma ou mais fichas para processar.")
	private static String	NO_FILES_TO_PROCESS;
	@Localize("Loading \"{0}\"... ")
	@Localize(locale = "de", value = "Lade \"{0}\"... ")
	@Localize(locale = "ru", value = "Загрузка \"{0}\"... ")
	@Localize(locale = "es", value = "Cargando \"{0}\"...")
	@Localize(locale = "pt-BR", value = "Carregando \"{0}\"...")
	private static String	LOADING;
	@Localize("  Creating PDF... ")
	@Localize(locale = "de", value = "  Erstelle PDF-Datei... ")
	@Localize(locale = "ru", value = "  Создание PDF... ")
	@Localize(locale = "es", value = "  Generando archivo PDF... ")
	@Localize(locale = "pt-BR", value = "  Gerando arquivo PDF... ")
	private static String	CREATING_PDF;
	@Localize("  Creating from text template... ")
	@Localize(locale = "pt-BR", value = "  Criado a partir do modelo de texto... ")
	private static String	CREATING_TEXT;
	@Localize("  Creating PNG... ")
	@Localize(locale = "de", value = "  Erstelle PNG-Datei... ")
	@Localize(locale = "ru", value = "  Создание PNG... ")
	@Localize(locale = "es", value = "  Generando archivo PNG... ")
	@Localize(locale = "pt-BR", value = "  Gerando arquivo PNG... ")
	private static String	CREATING_PNG;
	@Localize("  ** ERROR ENCOUNTERED **")
	@Localize(locale = "de", value = "  ** FEHLER AUFGETRETEN **")
	@Localize(locale = "ru", value = "  ** ОБНАРУЖЕНА ОШИБКА **")
	@Localize(locale = "es", value = "  ** SE HA ENCONTRADO UN ERROR **")
	@Localize(locale = "pt-BR", value = "  ** ERRO ENCONTRADO **")
	private static String	PROCESSING_FAILED;
	@Localize("\nDone! {0} overall.")
	@Localize(locale = "de", value = "\nFertig! Benötigte Zeit: {0}.")
	@Localize(locale = "ru", value = "\nГотово! {0} всего.")
	@Localize(locale = "es", value = "\n¡Terminado! {0}")
	@Localize(locale = "pt-BR", value = "\nFinalizado! {0}")
	private static String	FINISHED;
	@Localize("    Created \"{0}\".")
	@Localize(locale = "de", value = "    Erstellt: \"{0}\".")
	@Localize(locale = "ru", value = "    Создано \"{0}\".")
	@Localize(locale = "es", value = "    Creado: \"{0}\".")
	@Localize(locale = "pt-BR", value = "    Criado em: \"{0}\".")
	private static String	CREATED;
	@Localize("WARNING: Invalid paper size specification.")
	@Localize(locale = "de", value = "WARNUNG: Ungültiges Papierformat.")
	@Localize(locale = "ru", value = "ПРЕДУПРЕЖДЕНИЕ: Неверный размер бумаги.")
	@Localize(locale = "es", value = "AVISO: El tamaño del papel especificado no es válido")
	@Localize(locale = "pt-BR", value = "AVISO: Especificação de tamanho de papel inválida.")
	private static String	INVALID_PAPER_SIZE;
	@Localize("WARNING: Invalid paper margins specification.")
	@Localize(locale = "de", value = "WARNUNG: Ungültiger Randbereich.")
	@Localize(locale = "ru", value = "ПРЕДУПРЕЖДЕНИЕ: Неверные поля бумаги.")
	@Localize(locale = "es", value = "AVISO: Los margenes especificados no son válidos")
	@Localize(locale = "pt-BR", value = "AVISO: As margens definidas são inválidas.")
	private static String	INVALID_PAPER_MARGINS;
	@Localize("    Used text template file \"{0}\".")
	@Localize(locale = "pt-BR", value = "    Arquivo de texto de modelo usado \"{0}\".")
	private static String	TEMPLATE_USED;

	static {
		System.setProperty("locale.file", ".gcs_language"); //$NON-NLS-1$ //$NON-NLS-2$
		Localization.initialize();
	}

	private static final CmdLineOption	PDF_OPTION				= new CmdLineOption(PDF_OPTION_DESCRIPTION, null, FileType.PDF_EXTENSION);
	private static final CmdLineOption	TEXT_OPTION				= new CmdLineOption(TEXT_OPTION_DESCRIPTION, null, "text");									//$NON-NLS-1$
	private static final CmdLineOption	TEXT_TEMPLATE_OPTION	= new CmdLineOption(TEXT_TEMPLATE_OPTION_DESCRIPTION, TEXT_TEMPLATE_ARG, "text_template");	//$NON-NLS-1$
	private static final CmdLineOption	PNG_OPTION				= new CmdLineOption(PNG_OPTION_DESCRIPTION, null, FileType.PNG_EXTENSION);
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
		cmdLine.addOptions(TEXT_OPTION, TEXT_TEMPLATE_OPTION, PDF_OPTION, PNG_OPTION, SIZE_OPTION, MARGIN_OPTION);
		cmdLine.processArguments(args);
		if (cmdLine.isOptionUsed(TEXT_OPTION) || cmdLine.isOptionUsed(PDF_OPTION) || cmdLine.isOptionUsed(PNG_OPTION)) {
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
		Preferences.setPreferenceFile("gcs.pref"); //$NON-NLS-1$
		GCSFonts.register();
		GCSImages.getAppIcons(); // Just here to make sure the lookup paths are initialized
		Fonts.loadFromPreferences();
		App.setAboutPanel(AboutPanel.class);
		registerFileTypes(new GCSFileProxyCreator());
	}

	/** @return The path to the GCS library files. */
	public static Path getLibraryRootPath() {
		String library = System.getenv("GCS_LIBRARY"); //$NON-NLS-1$
		if (library != null) {
			return Paths.get(library);
		}
		Path path = App.getHomePath();
		if (BundleInfo.getDefault().getVersion() == 0) {
			path = path.resolve("../gcs_library"); //$NON-NLS-1$
		}
		return path.resolve("Library"); //$NON-NLS-1$
	}

	/**
	 * Registers the file types the app can open.
	 *
	 * @param fileProxyCreator The {@link FileProxyCreator} to use.
	 */
	public static void registerFileTypes(FileProxyCreator fileProxyCreator) {
		FileType.register(GURPSCharacter.EXTENSION, GCSImages.getCharacterSheetDocumentIcons(), SHEET_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true, true);
		FileType.register(AdvantageList.EXTENSION, GCSImages.getAdvantagesDocumentIcons(), ADVANTAGES_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true, true);
		FileType.register(EquipmentList.EXTENSION, GCSImages.getEquipmentDocumentIcons(), EQUIPMENT_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true, true);
		FileType.register(SkillList.EXTENSION, GCSImages.getSkillsDocumentIcons(), SKILLS_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true, true);
		FileType.register(SpellList.EXTENSION, GCSImages.getSpellsDocumentIcons(), SPELLS_LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true, true);
		FileType.register(Template.EXTENSION, GCSImages.getTemplateDocumentIcons(), TEMPLATE_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true, true);
		// For legacy
		FileType.register(LibraryFile.EXTENSION, GCSImages.getAdvantagesDocumentIcons(), LIBRARY_DESCRIPTION, REFERENCE_URL, fileProxyCreator, true, true);

		FileType.registerPdf(GCSImages.getPDFDocumentIcons(), fileProxyCreator, true, false);
		FileType.registerHtml(null, null, false, false);
		FileType.registerPng(null, null, false, false);
	}

	private static int convert(CmdLine cmdLine) {
		boolean text = cmdLine.isOptionUsed(GCS.TEXT_OPTION);
		boolean pdf = cmdLine.isOptionUsed(GCS.PDF_OPTION);
		boolean png = cmdLine.isOptionUsed(GCS.PNG_OPTION);
		int count = 0;

		if (text || pdf || png) {
			double[] paperSize = getPaperSize(cmdLine);
			double[] margins = getMargins(cmdLine);
			Timing timing = new Timing();
			String textTemplateOption = cmdLine.getOptionArgument(GCS.TEXT_TEMPLATE_OPTION);
			File textTemplate = null;

			if (textTemplateOption != null) {
				textTemplate = new File(textTemplateOption);
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
						if (text) {
							System.out.print(CREATING_TEXT);
							System.out.flush();
							textTemplate = TextTemplate.resolveTextTemplate(textTemplate);
							output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), PathUtils.getExtension(textTemplate.getName())));
							timing.reset();
							success = new TextTemplate(sheet).export(output, textTemplate);
							System.out.println(timing);
							System.out.println(MessageFormat.format(TEMPLATE_USED, PathUtils.getFullPath(textTemplate)));
							if (success) {
								System.out.println(MessageFormat.format(CREATED, output));
								count++;
							}
						}
						if (pdf) {
							System.out.print(CREATING_PDF);
							System.out.flush();
							output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), FileType.PDF_EXTENSION));
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
							output = new File(file.getParentFile(), PathUtils.enforceExtension(PathUtils.getLeafName(file.getName(), false), FileType.PNG_EXTENSION));
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
				double width = Numbers.extractDouble(argument.substring(0, index), -1.0, true);
				double height = Numbers.extractDouble(argument.substring(index + 1), -1.0, true);

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
					values[index] = Numbers.extractDouble(token, -1.0, true);
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
