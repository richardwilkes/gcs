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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.help;

import static com.trollworks.gcs.menu.help.HelpMenu_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.menu.DynamicMenuEnabler;
import com.trollworks.ttk.menu.help.AboutCommand;
import com.trollworks.ttk.menu.help.OpenURICommand;
import com.trollworks.ttk.menu.help.UpdateCommand;
import com.trollworks.ttk.utility.App;
import com.trollworks.ttk.utility.Platform;

import java.io.File;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

@Localized({
				@LS(key = "HELP", msg = "Help"),
				@LS(key = "RELEASE_NOTES", msg = "Release Notes"),
				@LS(key = "BUGS", msg = "Bug Reports"),
				@LS(key = "FEATURES", msg = "Feature Requests"),
				@LS(key = "LICENSE", msg = "License"),
				@LS(key = "WEB_SITE", msg = "Web Site"),
				@LS(key = "MAILING_LISTS", msg = "Mailing Lists"),
})
/** The standard "Help" menu. */
public class HelpMenu extends JMenu {
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
		add(new JMenuItem(new OpenURICommand(LICENSE, new File(App.APP_HOME_DIR, "License.html").toURI()))); //$NON-NLS-1$
		addSeparator();
		add(new JMenuItem(new OpenURICommand(WEB_SITE, "http://gurpscharactersheet.com"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURICommand(MAILING_LISTS, "http://sourceforge.net/mail/?group_id=185516"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURICommand(FEATURES, "http://sourceforge.net/tracker/?atid=913592&group_id=185516&func=browse"))); //$NON-NLS-1$
		add(new JMenuItem(new OpenURICommand(BUGS, "http://sourceforge.net/tracker/?atid=913589&group_id=185516&func=browse"))); //$NON-NLS-1$
		DynamicMenuEnabler.add(this);
	}
}
