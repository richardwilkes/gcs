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

package com.trollworks.gcs.menu;

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
	private static String		HELP;
	@Localize("Release Notes")
	private static String		RELEASE_NOTES;
	@Localize("Bug Reports")
	private static String		BUGS;
	@Localize("Feature Requests")
	private static String		FEATURES;
	@Localize("License")
	private static String		LICENSE;
	@Localize("Web Site")
	private static String		WEB_SITE;
	@Localize("Mailing Lists")
	private static String		MAILING_LISTS;

	static {
		Localization.initialize();
	}

	public static final String	NAME	= "Help";	//$NON-NLS-1$

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
		menu.add(new JMenuItem(new OpenURICommand(RELEASE_NOTES, "http://gurpscharactersheet.com/release_notes.php"))); //$NON-NLS-1$
		menu.add(new JMenuItem(new OpenURICommand(LICENSE, App.getHomePath().resolve("license.html").toUri()))); //$NON-NLS-1$
		menu.addSeparator();
		menu.add(new JMenuItem(new OpenURICommand(WEB_SITE, "http://gurpscharactersheet.com"))); //$NON-NLS-1$
		menu.add(new JMenuItem(new OpenURICommand(MAILING_LISTS, "http://gurpscharactersheet.com/mailing_lists.php"))); //$NON-NLS-1$
		menu.add(new JMenuItem(new OpenURICommand(FEATURES, "http://sourceforge.net/p/gcs-java/feature-requests"))); //$NON-NLS-1$
		menu.add(new JMenuItem(new OpenURICommand(BUGS, "http://sourceforge.net/p/gcs-java/bugs"))); //$NON-NLS-1$
		DynamicMenuEnabler.add(menu);
		return menu;
	}
}
