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

package com.trollworks.gcs.menu;

import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.gcs.menu.edit.EditMenu;
import com.trollworks.gcs.menu.file.FileMenu;
import com.trollworks.gcs.menu.help.HelpMenu;
import com.trollworks.gcs.menu.item.ItemMenu;
import com.trollworks.gcs.menu.window.WindowMenu;

import javax.swing.JMenu;
import javax.swing.JMenuBar;

/** The standard menu bar. */
public class StdMenuBar extends JMenuBar {
	/** Creates a new {@link StdMenuBar}. */
	public StdMenuBar() {
		add(new FileMenu());
		add(new EditMenu());
		add(new DataMenu());
		add(new ItemMenu());
		add(new WindowMenu());
		add(new HelpMenu());
	}

	/**
	 * @param bar The {@link JMenuBar} to search.
	 * @param type The {@link Class} to look for as a top-level {@link JMenu}.
	 * @return The found {@link JMenu}, or <code>null</code>.
	 */
	public static JMenu findMenu(JMenuBar bar, Class<? extends JMenu> type) {
		if (bar != null) {
			int count = bar.getMenuCount();
			for (int i = 0; i < count; i++) {
				JMenu menu = bar.getMenu(i);
				if (type.isInstance(menu)) {
					return menu;
				}
			}
		}
		return null;
	}
}
