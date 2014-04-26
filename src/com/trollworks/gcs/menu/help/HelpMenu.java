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

package com.trollworks.gcs.menu.help;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.help.AboutCommand;
import com.trollworks.toolkit.ui.menu.help.OpenURICommand;
import com.trollworks.toolkit.ui.menu.help.UpdateCommand;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Platform;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** The standard "Help" menu. */
public class HelpMenu extends JMenu {
	@Localize("Help")
	private static String	HELP;
	@Localize("Release Notes")
	private static String	RELEASE_NOTES;
	@Localize("Bug Reports")
	private static String	BUGS;
	@Localize("Feature Requests")
	private static String	FEATURES;
	@Localize("License")
	private static String	LICENSE;
	@Localize("Web Site")
	private static String	WEB_SITE;
	@Localize("Mailing Lists")
	private static String	MAILING_LISTS;

	static {
		Localization.initialize();
	}

	/** Creates a new {@link HelpMenu}. */
	public HelpMenu() {
		super(HELP);
		if (!Platform.isMacintosh()) {
			add(new JMenuItem(AboutCommand.INSTANCE));
			addSeparator();
		}
		add(new JMenuItem(UpdateCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(new OpenURICommand(RELEASE_NOTES, "http://gurpscharactersheet.com/Release_Notes"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURICommand(LICENSE, App.getHomePath().resolve("license.html").toUri()))); //$NON-NLS-1$
		addSeparator();
		add(new JMenuItem(new OpenURICommand(WEB_SITE, "http://gurpscharactersheet.com"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURICommand(MAILING_LISTS, "http://sourceforge.net/mail/?group_id=185516"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURICommand(FEATURES, "http://sourceforge.net/tracker/?atid=913592&group_id=185516&func=browse"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURICommand(BUGS, "http://sourceforge.net/tracker/?atid=913589&group_id=185516&func=browse"))); //$NON-NLS-1$
		DynamicMenuEnabler.add(this);
	}
}
