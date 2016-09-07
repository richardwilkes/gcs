/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu;

import com.trollworks.gcs.app.GCSApp;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.MenuProvider;
import com.trollworks.toolkit.ui.menu.help.AboutCommand;
import com.trollworks.toolkit.ui.menu.help.OpenURICommand;
import com.trollworks.toolkit.ui.menu.help.UpdateCommand;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Platform;

import java.util.Collections;
import java.util.Set;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** Provides the standard "Help" menu. */
public class HelpMenuProvider implements MenuProvider {
	@Localize("Help")
	@Localize(locale = "de", value = "Hilfe")
	@Localize(locale = "ru", value = "Справка")
	@Localize(locale = "es", value = "Ayuda")
	private static String	HELP;
	@Localize("Release Notes")
	@Localize(locale = "de", value = "Hinweise zur Veröffentlichung")
	@Localize(locale = "ru", value = "Примечания к выпуску")
	@Localize(locale = "es", value = "Notas de la versión")
	private static String	RELEASE_NOTES;
	@Localize("Bug Reports & Feature Requests")
	private static String	JIRA;
	@Localize("License")
	@Localize(locale = "de", value = "Lizenz")
	@Localize(locale = "ru", value = "Лицензия")
	@Localize(locale = "es", value = "Licencia")
	private static String	LICENSE;
	@Localize("Web Site")
	@Localize(locale = "de", value = "Webseite")
	@Localize(locale = "ru", value = "Сайт")
	@Localize(locale = "es", value = "Sitio Web")
	private static String	WEB_SITE;
	@Localize("Mailing Lists")
	@Localize(locale = "de", value = "Mailinglisten")
	@Localize(locale = "ru", value = "Списки рассылки")
	@Localize(locale = "es", value = "Listas de correo")
	private static String	MAILING_LISTS;

	static {
		Localization.initialize();
	}

	public static final String NAME = "Help"; //$NON-NLS-1$

	@Override
	public Set<Command> getModifiableCommands() {
		return Collections.emptySet();
	}

	@Override
	public JMenu createMenu() {
		JMenu menu = new JMenu(HELP);
		menu.setName(NAME);
		if (!Platform.isMacintosh()) {
			menu.add(new JMenuItem(AboutCommand.INSTANCE));
			menu.addSeparator();
		}
		menu.add(new JMenuItem(UpdateCommand.INSTANCE));
		menu.addSeparator();
		menu.add(new JMenuItem(new OpenURICommand(RELEASE_NOTES, GCSApp.WEB_SITE + "/release_notes.php"))); //$NON-NLS-1$
		menu.add(new JMenuItem(new OpenURICommand(LICENSE, App.getHomePath().resolve("license.html").toUri()))); //$NON-NLS-1$
		menu.addSeparator();
		menu.add(new JMenuItem(new OpenURICommand(WEB_SITE, GCSApp.WEB_SITE)));
		menu.add(new JMenuItem(new OpenURICommand(MAILING_LISTS, GCSApp.WEB_SITE + "/mailing_lists.php"))); //$NON-NLS-1$
		menu.add(new JMenuItem(new OpenURICommand(JIRA, "https://gurpscharactersheet.atlassian.net"))); //$NON-NLS-1$
		DynamicMenuEnabler.add(menu);
		return menu;
	}
}
