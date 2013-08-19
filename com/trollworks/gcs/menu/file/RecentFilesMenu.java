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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.advantage.AdvantageListWindow;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.equipment.EquipmentListWindow;
import com.trollworks.gcs.menu.data.OpenDataFileCommand;
import com.trollworks.gcs.skill.SkillListWindow;
import com.trollworks.gcs.spell.SpellListWindow;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.Preferences;

import java.io.File;
import java.util.ArrayList;

import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.event.MenuEvent;
import javax.swing.event.MenuListener;

/** The standard "Recent Files" menu. */
public class RecentFilesMenu extends JMenu implements MenuListener {
	private static final String				PREFS_MODULE		= "RecentFiles";																																									//$NON-NLS-1$
	private static final int				PREFS_VERSION		= 1;
	private static String					MSG_TITLE;
	private static final int				MAX_RECENTS			= 20;
	private static final ArrayList<File>	RECENTS				= new ArrayList<File>();
	private static final String[]			ALLOWED_EXTENSIONS	= { SheetWindow.SHEET_EXTENSION, TemplateWindow.EXTENSION, AdvantageListWindow.EXTENSION, SkillListWindow.EXTENSION, SpellListWindow.EXTENSION, EquipmentListWindow.EXTENSION };

	static {
		LocalizedMessages.initialize(RecentFilesMenu.class);
		Preferences prefs = Preferences.getInstance();
		prefs.resetIfVersionMisMatch(PREFS_MODULE, PREFS_VERSION);
		for (int i = 0; i < MAX_RECENTS; i++) {
			String path = prefs.getStringValue(PREFS_MODULE, Integer.toString(i));
			if (path == null) {
				break;
			}
			addRecent(new File(path));
		}
	}

	/** @return The current set of recents. */
	public static ArrayList<File> getRecents() {
		return new ArrayList<File>(RECENTS);
	}

	/** @return The number of recents currently present. */
	public static int getRecentCount() {
		return RECENTS.size();
	}

	/** Removes all recents. */
	public static void clearRecents() {
		RECENTS.clear();
	}

	/** @param file The {@link File} to add to the recents list. */
	public static void addRecent(File file) {
		String extension = Path.getExtension(file.getName());
		if (Platform.isMacintosh() || Platform.isWindows()) {
			extension = extension.toLowerCase();
		}
		for (String allowed : ALLOWED_EXTENSIONS) {
			if (allowed.equals(extension)) {
				if (file.canRead()) {
					file = file.getAbsoluteFile();
					RECENTS.remove(file);
					RECENTS.add(0, file);
					if (RECENTS.size() > MAX_RECENTS) {
						RECENTS.remove(MAX_RECENTS);
					}
				}
				break;
			}
		}
	}

	/** Saves the current set of recents to preferences. */
	public static void saveToPreferences() {
		Preferences prefs = Preferences.getInstance();
		prefs.startBatch();
		prefs.removePreferences(PREFS_MODULE);
		int count = RECENTS.size();
		for (int i = 0; i < count; i++) {
			prefs.setValue(PREFS_MODULE, Integer.toString(i), RECENTS.get(i).getAbsolutePath());
		}
		prefs.endBatch();
	}

	/** Creates a new {@link RecentFilesMenu}. */
	public RecentFilesMenu() {
		super(MSG_TITLE);
		addMenuListener(this);
	}

	public void menuCanceled(MenuEvent event) {
		// Nothing to do.
	}

	public void menuDeselected(MenuEvent event) {
		// Nothing to do.
	}

	public void menuSelected(MenuEvent event) {
		removeAll();
		ArrayList<File> list = new ArrayList<File>();
		for (File file : RECENTS) {
			if (file.canRead()) {
				list.add(file);
				add(new JMenuItem(new OpenDataFileCommand(Path.getLeafName(file.getName(), false), file)));
				if (list.size() == MAX_RECENTS) {
					break;
				}
			}
		}
		RECENTS.clear();
		RECENTS.addAll(list);

		if (getRecentCount() > 0) {
			addSeparator();
		}
		JMenuItem item = new JMenuItem(ClearRecentFilesMenuCommand.INSTANCE);
		ClearRecentFilesMenuCommand.INSTANCE.adjustForMenu(item);
		add(item);
	}
}
