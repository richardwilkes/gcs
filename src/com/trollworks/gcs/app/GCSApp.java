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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.gcs.menu.edit.EditMenu;
import com.trollworks.gcs.menu.file.FileMenu;
import com.trollworks.gcs.menu.file.NewCharacterSheetCommand;
import com.trollworks.gcs.menu.help.HelpMenu;
import com.trollworks.gcs.menu.item.ItemMenu;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.ui.UpdateChecker;
import com.trollworks.toolkit.ui.WindowsRegistry;
import com.trollworks.toolkit.ui.menu.StdMenuBar;
import com.trollworks.toolkit.ui.menu.window.WindowMenu;
import com.trollworks.toolkit.ui.preferences.FontPreferences;
import com.trollworks.toolkit.ui.preferences.MenuKeyPreferences;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.ui.widget.AppWindow;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.cmdline.CmdLine;

import java.io.File;
import java.util.HashMap;

/** The main application user interface. */
public class GCSApp extends App {
	@Localize("GURPS Character Sheet")
	private static String		SHEET_DESCRIPTION;
	@Localize("GCS Library")
	private static String		LIBRARY_DESCRIPTION;
	@Localize("GCS Character Template")
	private static String		TEMPLATE_DESCRIPTION;
	@Localize("GCS Traits")
	private static String		TRAITS_DESCRIPTION;
	@Localize("GCS Equipment")
	private static String		EQUIPMENT_DESCRIPTION;
	@Localize("GCS Skills")
	private static String		SKILLS_DESCRIPTION;
	@Localize("GCS Spells")
	private static String		SPELLS_DESCRIPTION;

	static {
		Localization.initialize();
	}

	/** The one and only instance of this class. */
	public static final GCSApp	INSTANCE	= new GCSApp();

	private GCSApp() {
		super();
		AppWindow.setDefaultWindowIcon(GCSImages.getDefaultWindowIcon());
	}

	@SuppressWarnings("unchecked")
	@Override
	public void configureApplication(CmdLine cmdLine) {
		HashMap<String, String> map = new HashMap<>();
		map.put(SheetWindow.SHEET_EXTENSION.substring(1), SHEET_DESCRIPTION);
		map.put(LibraryFile.EXTENSION.substring(1), LIBRARY_DESCRIPTION);
		map.put(TemplateWindow.EXTENSION.substring(1), TEMPLATE_DESCRIPTION);
		map.put(Advantage.OLD_ADVANTAGE_EXTENSION.substring(1), TRAITS_DESCRIPTION);
		map.put(Equipment.OLD_EQUIPMENT_EXTENSION.substring(1), EQUIPMENT_DESCRIPTION);
		map.put(Skill.OLD_SKILL_EXTENSION.substring(1), SKILLS_DESCRIPTION);
		map.put(Spell.OLD_SPELL_EXTENSION.substring(1), SPELLS_DESCRIPTION);
		WindowsRegistry.register("GCS", map, new File(APP_HOME_DIR, "gcs.bat"), new File(APP_HOME_DIR, "GURPS Character Sheet.app/Contents/Resources")); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$

		UpdateChecker.check("gcs", "http://gurpscharactersheet.com/current.txt", "http://gurpscharactersheet.com"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$

		ListCollectionThread.get();

		StdMenuBar.configure(FileMenu.class, EditMenu.class, DataMenu.class, ItemMenu.class, WindowMenu.class, HelpMenu.class);
		SheetPreferences.initialize();
		PreferencesWindow.addCategory(SheetPreferences.class);
		PreferencesWindow.addCategory(FontPreferences.class);
		PreferencesWindow.addCategory(MenuKeyPreferences.class);
	}

	@Override
	public void noWindowsAreOpenAtStartup(boolean finalChance) {
		if (finalChance) {
			NewCharacterSheetCommand.newSheet();
		} else {
			StartupDialog sd = new StartupDialog();
			sd.setVisible(true);
		}
	}

	@Override
	public void finalStartup() {
		super.finalStartup();
		setDefaultMenuBar(new StdMenuBar());
	}
}
