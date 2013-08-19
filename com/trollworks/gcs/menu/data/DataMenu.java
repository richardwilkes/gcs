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

package com.trollworks.gcs.menu.data;

import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.widgets.AppWindow;

import java.awt.Toolkit;
import java.awt.event.InputEvent;
import java.io.File;
import java.util.ArrayList;

import javax.swing.ImageIcon;
import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.KeyStroke;

/** The standard "Data" menu. */
public class DataMenu extends JMenu {
	private static String		MSG_DATA;
	private static final String	MENU_KEY_SUFFIX_REGEX	= ".*__[A-Za-z0-9]";	//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(DataMenu.class);
	}

	/** Creates a new {@link DataMenu}. */
	public DataMenu() {
		super(MSG_DATA);
		addToMenu(ListCollectionThread.get().getLists(), this);
		DynamicMenuEnabler.add(this);
	}

	/** Updates the available menu items. */
	public static void update() {
		for (AppWindow window : AppWindow.getAllWindows()) {
			DataMenu menu = (DataMenu) StdMenuBar.findMenu(window.getJMenuBar(), DataMenu.class);
			if (menu != null) {
				menu.removeAll();
				menu.addToMenu(ListCollectionThread.get().getLists(), menu);
			}
		}
	}

	private void addToMenu(ArrayList<?> list, JMenu menu) {
		int keyMask = Toolkit.getDefaultToolkit().getMenuShortcutKeyMask() | InputEvent.ALT_MASK;
		int count = list.size();
		for (int i = 1; i < count; i++) {
			Object entry = list.get(i);
			if (entry instanceof ArrayList) {
				ArrayList<?> subList = (ArrayList<?>) entry;
				JMenu subMenu = new JMenu((String) subList.get(0));
				subMenu.setIcon(new ImageIcon(Images.getFolderIcon()));
				addToMenu(subList, subMenu);
				menu.add(subMenu);
			} else {
				KeyStroke keyStroke = null;
				File file = (File) entry;
				String name = Path.getLeafName(file.getName(), false);
				if (name.matches(MENU_KEY_SUFFIX_REGEX)) {
					keyStroke = KeyStroke.getKeyStroke(Character.toUpperCase(name.charAt(name.length() - 1)), keyMask);
					name = name.substring(0, name.length() - 3);
				}
				JMenuItem item = new JMenuItem(new OpenDataFileCommand(name, file));
				if (keyStroke != null) {
					item.setAccelerator(keyStroke);
				}
				menu.add(item);
			}
		}
	}

	/**
	 * @param title The title to filter for menu key suffixes.
	 * @return The filtered title.
	 */
	public static String filterTitle(String title) {
		if (title.matches(MENU_KEY_SUFFIX_REGEX)) {
			return title.substring(0, title.length() - 3);
		}
		return title;
	}
}
