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

package com.trollworks.gcs.menu.window;

import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.AppWindow;

import java.util.ArrayList;
import java.util.Collections;

import javax.swing.JCheckBoxMenuItem;
import javax.swing.JMenu;

/** The standard "Window" menu. */
public class WindowMenu extends JMenu {
	private static String	MSG_WINDOW;

	static {
		LocalizedMessages.initialize(WindowMenu.class);
	}

	/** Creates a new {@link WindowMenu}. */
	public WindowMenu() {
		super(MSG_WINDOW);
		DynamicMenuEnabler.add(this);
	}

	/** Updates the available menu items. */
	public static void update() {
		ArrayList<AppWindow> windows = AppWindow.getAllWindows();
		Collections.sort(windows);
		for (AppWindow window : windows) {
			WindowMenu windowMenu = (WindowMenu) StdMenuBar.findMenu(window.getJMenuBar(), WindowMenu.class);
			if (windowMenu != null) {
				windowMenu.removeAll();
				for (AppWindow one : windows) {
					windowMenu.add(new JCheckBoxMenuItem(new SwitchToWindowCommand(one)));
				}
			}
		}
	}
}
