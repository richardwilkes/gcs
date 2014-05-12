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

import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.image.ToolkitImage;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.MenuProvider;
import com.trollworks.toolkit.ui.menu.StdMenuBar;
import com.trollworks.toolkit.ui.menu.file.OpenDataFileCommand;
import com.trollworks.toolkit.ui.widget.AppWindow;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;

import java.awt.event.InputEvent;
import java.nio.file.Path;
import java.util.Collections;
import java.util.List;
import java.util.Set;

import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.KeyStroke;

/** The standard "Data" menu. */
public class DataMenuProvider implements MenuProvider {
	@Localize("Data")
	private static String		DATA;

	static {
		Localization.initialize();
	}

	public static final String	NAME					= "Data";				//$NON-NLS-1$
	private static final String	MENU_KEY_SUFFIX_REGEX	= ".*__[A-Za-z0-9]";	//$NON-NLS-1$

	/** Updates the available menu items. */
	public static void update() {
		for (AppWindow window : AppWindow.getAllWindows()) {
			JMenu menu = StdMenuBar.findMenuByName(window.getJMenuBar(), NAME);
			if (menu != null) {
				menu.removeAll();
				addToMenu(ListCollectionThread.get().getLists(), menu);
			}
		}
	}

	private static void addToMenu(List<?> list, JMenu menu) {
		int keyMask = menu.getToolkit().getMenuShortcutKeyMask() | InputEvent.ALT_MASK;
		int count = list.size();
		for (int i = 1; i < count; i++) {
			Object entry = list.get(i);
			if (entry instanceof List<?>) {
				List<?> subList = (List<?>) entry;
				JMenu subMenu = new JMenu((String) subList.get(0));
				subMenu.setIcon(ToolkitImage.getFolderIcons().getIcon(16));
				addToMenu(subList, subMenu);
				menu.add(subMenu);
			} else {
				KeyStroke keyStroke = null;
				Path file = (Path) entry;
				String name = PathUtils.getLeafName(file.getFileName(), false);
				if (name.matches(MENU_KEY_SUFFIX_REGEX)) {
					keyStroke = KeyStroke.getKeyStroke(Character.toUpperCase(name.charAt(name.length() - 1)), keyMask);
					name = name.substring(0, name.length() - 3);
				}
				JMenuItem item = new JMenuItem(new OpenDataFileCommand(name, file.toFile()));
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

	@Override
	public Set<Command> getModifiableCommands() {
		return Collections.emptySet();
	}

	@Override
	public JMenu createMenu() {
		JMenu menu = new JMenu(DATA);
		menu.setName(NAME);
		addToMenu(ListCollectionThread.get().getLists(), menu);
		DynamicMenuEnabler.add(menu);
		return menu;
	}
}
