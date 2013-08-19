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

package com.trollworks.gcs.menu.help;

import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.io.File;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** The standard "Help" menu. */
public class HelpMenu extends JMenu {
	private static String	MSG_HELP;
	private static String	MSG_RELEASE_NOTES;
	private static String	MSG_BUGS;
	private static String	MSG_FEATURES;
	private static String	MSG_USERS_MANUAL;
	private static String	MSG_LICENSE;
	private static String	MSG_WEB_SITE;
	private static String	MSG_MAILING_LISTS;

	static {
		LocalizedMessages.initialize(HelpMenu.class);
	}

	/** Creates a new {@link HelpMenu}. */
	public HelpMenu() {
		super(MSG_HELP);
		if (!Platform.isMacintosh()) {
			add(new JMenuItem(AboutCommand.INSTANCE));
			addSeparator();
		}
		add(new JMenuItem(UpdateCommand.INSTANCE));
		addSeparator();
		File docsDir = new File(System.getProperty("app.home", "."), "docs"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
		add(new JMenuItem(new OpenLocalFileCommand(MSG_RELEASE_NOTES, new File(docsDir, "release_notes.html")))); //$NON-NLS-1$
		add(new JMenuItem(new OpenLocalFileCommand(MSG_USERS_MANUAL, new File(docsDir, "guide.html")))); //$NON-NLS-1$
		add(new JMenuItem(new OpenLocalFileCommand(MSG_LICENSE, new File(docsDir, "license.html")))); //$NON-NLS-1$
		addSeparator();
		add(new JMenuItem(new OpenURLCommand(MSG_WEB_SITE, "http://gcs.trollworks.com"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURLCommand(MSG_MAILING_LISTS, "http://sourceforge.net/mail/?group_id=185516"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURLCommand(MSG_FEATURES, "http://sourceforge.net/tracker/?atid=913592&group_id=185516&func=browse"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURLCommand(MSG_BUGS, "http://sourceforge.net/tracker/?atid=913589&group_id=185516&func=browse"))); //$NON-NLS-1$
		DynamicMenuEnabler.add(this);
	}
}
