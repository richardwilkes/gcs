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

package com.trollworks.gcs.menu.help;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.Browser;

import java.awt.event.ActionEvent;

import javax.swing.JMenuItem;

/** A command that will open a specific URL in the preferred browser. */
public class OpenURLCommand extends Command {
	private String	mURL;

	/**
	 * Creates a new {@link OpenURLCommand}.
	 * 
	 * @param title The title to use.
	 * @param url The URL to open.
	 */
	public OpenURLCommand(String title, String url) {
		super(title);
		mURL = url;
	}

	@Override public void adjustForMenu(JMenuItem item) {
		// Not used. Always enabled.
	}

	@Override public void actionPerformed(ActionEvent event) {
		Browser.getPreferredBrowser().openURL(mURL);
	}
}
