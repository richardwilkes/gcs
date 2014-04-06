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

package com.trollworks.gcs.menu.data;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;


import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.toolkit.ui.Path;
import com.trollworks.toolkit.ui.image.ToolkitImage;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.StdMenuBar;
import com.trollworks.toolkit.ui.menu.file.OpenDataFileCommand;
import com.trollworks.toolkit.ui.widget.AppWindow;

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
	@Localize("Data")
	private static String DATA;

	static {
		Localization.initialize();
	}

	private static final String	MENU_KEY_SUFFIX_REGEX	= ".*__[A-Za-z0-9]";	//$NON-NLS-1$

	/** Creates a new {@link DataMenu}. */
	public DataMenu() {
		super(DATA);
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
			if (entry instanceof ArrayList<?>) {
				ArrayList<?> subList = (ArrayList<?>) entry;
				JMenu subMenu = new JMenu((String) subList.get(0));
				subMenu.setIcon(new ImageIcon(ToolkitImage.getFolderIcon()));
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
