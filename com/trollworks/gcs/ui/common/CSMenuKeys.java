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

package com.trollworks.gcs.ui.common;

import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.widget.menu.TKMenuBar;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutlineHeaderCM;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Event;
import java.awt.event.KeyEvent;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;

/** Menu key mappings. */
public class CSMenuKeys {
	/** The preferences module. */
	public static final String					MODULE		= "CSMenuKeys";						//$NON-NLS-1$
	private static final char					SEPARATOR	= ':';
	private static HashMap<String, TKKeystroke>	KEY_MAP		= new HashMap<String, TKKeystroke>();
	private static HashMap<String, TKKeystroke>	STD_KEY_MAP	= new HashMap<String, TKKeystroke>();
	private static HashMap<String, String>		TITLE_MAP	= new HashMap<String, String>();
	private static HashMap<TKKeystroke, String>	CMD_MAP		= new HashMap<TKKeystroke, String>();
	private static boolean						UPDATE_PREFS;

	static {
		TKPreferences prefs = TKPreferences.getInstance();
		HashSet<String> oldKeys = new HashSet<String>();

		// Localizable strings
		// File menu
		TITLE_MAP.put(CSWindow.CMD_NEW_SHEET, Msgs.NEW_CHARACTER_SHEET);
		TITLE_MAP.put(CSWindow.CMD_NEW_TEMPLATE, Msgs.NEW_CHARACTER_TEMPLATE);
		TITLE_MAP.put(CSWindow.CMD_NEW_LIST, Msgs.NEW_LIST);
		TITLE_MAP.put(TKWindow.CMD_OPEN, Msgs.OPEN);
		TITLE_MAP.put(TKWindow.CMD_ATTEMPT_CLOSE, Msgs.CLOSE);
		TITLE_MAP.put(TKWindow.CMD_SAVE, Msgs.SAVE);
		TITLE_MAP.put(TKWindow.CMD_SAVE_AS, Msgs.SAVE_AS);
		TITLE_MAP.put(TKWindow.CMD_PAGE_SETUP, Msgs.PAGE_SETUP);
		TITLE_MAP.put(TKWindow.CMD_PRINT, Msgs.PRINT);
		TITLE_MAP.put(TKWindow.CMD_QUIT, TKMenuBar.getQuitTitle());

		// Edit menu
		TITLE_MAP.put(TKWindow.CMD_UNDO, Msgs.UNDO);
		TITLE_MAP.put(TKWindow.CMD_REDO, Msgs.REDO);
		TITLE_MAP.put(TKWindow.CMD_CUT, Msgs.CUT);
		TITLE_MAP.put(TKWindow.CMD_COPY, Msgs.COPY);
		TITLE_MAP.put(TKWindow.CMD_PASTE, Msgs.PASTE);
		TITLE_MAP.put(TKWindow.CMD_DUPLICATE, Msgs.DUPLICATE);
		TITLE_MAP.put(TKWindow.CMD_CLEAR, Msgs.CLEAR);
		TITLE_MAP.put(TKWindow.CMD_SELECT_ALL, Msgs.SELECT_ALL);
		TITLE_MAP.put(CSWindow.CMD_INCREMENT, Msgs.INCREMENT);
		TITLE_MAP.put(CSWindow.CMD_DECREMENT, Msgs.DECREMENT);
		TITLE_MAP.put(CSWindow.CMD_TOGGLE_EQUIPPED, Msgs.TOGGLE_EQUIPPED);
		TITLE_MAP.put(CSWindow.CMD_JUMP_TO_FIND, Msgs.JUMP_TO_FIND);
		TITLE_MAP.put(CSWindow.CMD_RANDOMIZE_DESCRIPTION, Msgs.RANDOMIZE_DESCRIPTION);
		TITLE_MAP.put(CSWindow.CMD_RANDOMIZE_FEMALE_NAME, Msgs.RANDOMIZE_FEMALE_NAME);
		TITLE_MAP.put(CSWindow.CMD_RANDOM_MALE_NAME, Msgs.RANDOMIZE_MALE_NAME);
		TITLE_MAP.put(CSWindow.CMD_ADD_NATURAL_PUNCH, Msgs.ADD_NATURAL_PUNCH);
		TITLE_MAP.put(CSWindow.CMD_ADD_NATURAL_KICK, Msgs.ADD_NATURAL_KICK);
		TITLE_MAP.put(CSWindow.CMD_ADD_NATURAL_KICK_WITH_BOOTS, Msgs.ADD_NATURAL_KICK_WITH_BOOTS);
		TITLE_MAP.put(TKOutlineHeaderCM.CMD_RESET_COLUMNS, Msgs.RESET_COLUMNS);
		TITLE_MAP.put(TKWindow.CMD_RESET_CONFIRMATION_DIALOGS, Msgs.RESET_CONFIRMATION_DIALOGS);
		TITLE_MAP.put(TKWindow.CMD_PREFERENCES, Msgs.PREFERENCES);

		// Item menu
		TITLE_MAP.put(CSWindow.CMD_OPEN_EDITOR, Msgs.OPEN_EDITOR);
		TITLE_MAP.put(CSWindow.CMD_COPY_TO_SHEET, Msgs.COPY_TO_SHEET);
		TITLE_MAP.put(CSWindow.CMD_COPY_TO_TEMPLATE, Msgs.COPY_TO_TEMPLATE);
		TITLE_MAP.put(CSWindow.CMD_APPLY_TEMPLATE_TO_SHEET, Msgs.APPLY_TEMPLATE_TO_SHEET);
		TITLE_MAP.put(CSWindow.CMD_NEW_ADVANTAGE, Msgs.NEW_ADVANTAGE);
		TITLE_MAP.put(CSWindow.CMD_NEW_ADVANTAGE_CONTAINER, Msgs.NEW_ADVANTAGE_CONTAINER);
		TITLE_MAP.put(CSWindow.CMD_NEW_SKILL, Msgs.NEW_SKILL);
		TITLE_MAP.put(CSWindow.CMD_NEW_SKILL_CONTAINER, Msgs.NEW_SKILL_CONTAINER);
		TITLE_MAP.put(CSWindow.CMD_NEW_TECHNIQUE, Msgs.NEW_TECHNIQUE);
		TITLE_MAP.put(CSWindow.CMD_NEW_SPELL, Msgs.NEW_SPELL);
		TITLE_MAP.put(CSWindow.CMD_NEW_SPELL_CONTAINER, Msgs.NEW_SPELL_CONTAINER);
		TITLE_MAP.put(CSWindow.CMD_NEW_CARRIED_EQUIPMENT, Msgs.NEW_CARRIED_EQUIPMENT);
		TITLE_MAP.put(CSWindow.CMD_NEW_CARRIED_EQUIPMENT_CONTAINER, Msgs.NEW_CARRIED_EQUIPMENT_CONTAINER);
		TITLE_MAP.put(CSWindow.CMD_NEW_EQUIPMENT, Msgs.NEW_EQUIPMENT);
		TITLE_MAP.put(CSWindow.CMD_NEW_EQUIPMENT_CONTAINER, Msgs.NEW_EQUIPMENT_CONTAINER);

		// Help menu
		TITLE_MAP.put(TKWindow.CMD_ABOUT, MessageFormat.format(Msgs.ABOUT, TKApp.getName()));
		TITLE_MAP.put(CSWindow.CMD_RELEASE_NOTES, Msgs.RELEASE_NOTES);
		TITLE_MAP.put(CSWindow.CMD_USERS_MANUAL, Msgs.USERS_MANUAL);
		TITLE_MAP.put(CSWindow.CMD_LICENSE, Msgs.LICENSE);
		TITLE_MAP.put(CSWindow.CMD_WEB_SITE, Msgs.WEB_SITE);
		TITLE_MAP.put(CSWindow.CMD_FEATURES, Msgs.FEATURES);
		TITLE_MAP.put(CSWindow.CMD_BUGS, Msgs.BUGS);
		TITLE_MAP.put(CSWindow.CMD_MAILING_LISTS, Msgs.MAILING_LISTS);

		reset();
		STD_KEY_MAP.putAll(KEY_MAP);
		for (String key : prefs.getModuleKeys(MODULE)) {
			if (KEY_MAP.containsKey(key)) {
				String buffer = prefs.getStringValue(MODULE, key);
				int separator = buffer.indexOf(SEPARATOR);

				if (separator == -1) {
					put(key, null);
				} else {
					try {
						put(key, new TKKeystroke(Integer.parseInt(buffer.substring(0, separator)), Integer.parseInt(buffer.substring(separator + 1))));
					} catch (Exception exception) {
						put(key, null);
					}
				}
			} else {
				oldKeys.add(key);
			}
		}
		for (String key : oldKeys) {
			prefs.removePreference(MODULE, key);
		}
		UPDATE_PREFS = true;
	}

	/** Resets all menu keys to their defaults. */
	public static final void reset() {
		int shifted = TKKeystroke.getCommandMask() | Event.SHIFT_MASK;

		KEY_MAP.clear();

		// File menu
		put(CSWindow.CMD_NEW_SHEET, new TKKeystroke('N'));
		put(CSWindow.CMD_NEW_TEMPLATE, new TKKeystroke('N', shifted));
		put(CSWindow.CMD_NEW_LIST, new TKKeystroke('L'));
		put(TKWindow.CMD_OPEN, new TKKeystroke('O'));
		put(TKWindow.CMD_ATTEMPT_CLOSE, new TKKeystroke('W'));
		put(TKWindow.CMD_SAVE, new TKKeystroke('S'));
		put(TKWindow.CMD_SAVE_AS, null);
		put(TKWindow.CMD_PAGE_SETUP, new TKKeystroke('P', shifted));
		put(TKWindow.CMD_PRINT, new TKKeystroke('P'));
		put(TKWindow.CMD_QUIT, new TKKeystroke('Q'));

		// Edit menu
		put(TKWindow.CMD_UNDO, new TKKeystroke('Z'));
		put(TKWindow.CMD_REDO, new TKKeystroke('Y'));
		put(TKWindow.CMD_CUT, new TKKeystroke('X'));
		put(TKWindow.CMD_COPY, new TKKeystroke('C'));
		put(TKWindow.CMD_PASTE, new TKKeystroke('V'));
		put(TKWindow.CMD_DUPLICATE, new TKKeystroke('U'));
		put(TKWindow.CMD_CLEAR, new TKKeystroke(Event.DELETE, 0));
		put(TKWindow.CMD_SELECT_ALL, new TKKeystroke('A'));
		put(CSWindow.CMD_INCREMENT, new TKKeystroke('='));
		put(CSWindow.CMD_DECREMENT, new TKKeystroke('-'));
		put(CSWindow.CMD_TOGGLE_EQUIPPED, new TKKeystroke(KeyEvent.VK_QUOTE));
		put(CSWindow.CMD_JUMP_TO_FIND, new TKKeystroke('J'));
		put(CSWindow.CMD_RANDOMIZE_DESCRIPTION, null);
		put(CSWindow.CMD_RANDOMIZE_FEMALE_NAME, new TKKeystroke('V', shifted));
		put(CSWindow.CMD_RANDOM_MALE_NAME, new TKKeystroke('I', shifted));
		put(CSWindow.CMD_ADD_NATURAL_PUNCH, null);
		put(CSWindow.CMD_ADD_NATURAL_KICK, null);
		put(CSWindow.CMD_ADD_NATURAL_KICK_WITH_BOOTS, null);
		put(TKOutlineHeaderCM.CMD_RESET_COLUMNS, null);
		put(TKWindow.CMD_RESET_CONFIRMATION_DIALOGS, null);
		put(TKWindow.CMD_PREFERENCES, new TKKeystroke(','));

		// Item menu
		put(CSWindow.CMD_OPEN_EDITOR, new TKKeystroke('I'));
		put(CSWindow.CMD_COPY_TO_SHEET, new TKKeystroke('C', shifted));
		put(CSWindow.CMD_COPY_TO_TEMPLATE, new TKKeystroke('T', shifted));
		put(CSWindow.CMD_APPLY_TEMPLATE_TO_SHEET, new TKKeystroke('A', shifted));
		put(CSWindow.CMD_NEW_ADVANTAGE, new TKKeystroke('D'));
		put(CSWindow.CMD_NEW_ADVANTAGE_CONTAINER, new TKKeystroke('D', shifted));
		put(CSWindow.CMD_NEW_SKILL, new TKKeystroke('K'));
		put(CSWindow.CMD_NEW_SKILL_CONTAINER, new TKKeystroke('K', shifted));
		put(CSWindow.CMD_NEW_TECHNIQUE, new TKKeystroke('T'));
		put(CSWindow.CMD_NEW_SPELL, new TKKeystroke('B'));
		put(CSWindow.CMD_NEW_SPELL_CONTAINER, new TKKeystroke('B', shifted));
		put(CSWindow.CMD_NEW_CARRIED_EQUIPMENT, new TKKeystroke('E'));
		put(CSWindow.CMD_NEW_CARRIED_EQUIPMENT_CONTAINER, new TKKeystroke('E', shifted));
		put(CSWindow.CMD_NEW_EQUIPMENT, new TKKeystroke('F'));
		put(CSWindow.CMD_NEW_EQUIPMENT_CONTAINER, new TKKeystroke('F', shifted));

		// Help menu
		put(TKWindow.CMD_ABOUT, null);
		put(CSWindow.CMD_RELEASE_NOTES, null);
		put(CSWindow.CMD_USERS_MANUAL, null);
		put(CSWindow.CMD_LICENSE, null);
		put(CSWindow.CMD_WEB_SITE, null);
		put(CSWindow.CMD_FEATURES, null);
		put(CSWindow.CMD_BUGS, null);
		put(CSWindow.CMD_MAILING_LISTS, null);
	}

	/**
	 * Puts a particular mapping into the menu key map.
	 * 
	 * @param cmd The command to alter.
	 * @param keystroke The new keystroke.
	 */
	public static final void put(String cmd, TKKeystroke keystroke) {
		TKKeystroke old = KEY_MAP.get(cmd);

		KEY_MAP.put(cmd, keystroke);
		CMD_MAP.remove(old);
		if (keystroke != null) {
			CMD_MAP.put(keystroke, cmd);
		}

		if (UPDATE_PREFS) {
			StringBuffer buffer = new StringBuffer();

			if (keystroke != null) {
				buffer.append(keystroke.getKeyCode());
				buffer.append(SEPARATOR);
				buffer.append(keystroke.getModifiers());
			}
			TKPreferences.getInstance().setValue(MODULE, cmd, buffer.toString());
		}

		for (TKWindow window : TKWindow.getAllWindows()) {
			TKMenuBar menuBar = window.getTKMenuBar();

			if (menuBar != null) {
				TKMenuItem item = menuBar.getMenuItemForCommand(cmd);

				if (item != null) {
					item.setKeyStroke(keystroke);
				}
			}
		}
	}

	/**
	 * @param cmd The command.
	 * @return The keystroke to use for the specified command.
	 */
	public static final TKKeystroke getKeyStroke(String cmd) {
		return KEY_MAP.get(cmd);
	}

	/**
	 * @param keyStroke The key stroke.
	 * @return The command for the specified key stroke.
	 */
	public static final String getCommand(TKKeystroke keyStroke) {
		return CMD_MAP.get(keyStroke);
	}

	/**
	 * @param cmd The command.
	 * @return The default title to use for the specified command.
	 */
	public static final String getTitle(String cmd) {
		return TITLE_MAP.get(cmd);
	}

	/** @return All of the commands present in the menu key map, ordered by title. */
	public static final String[] getCommands() {
		HashMap<String, String> titleToCmdMap = new HashMap<String, String>();
		ArrayList<String> titles = new ArrayList<String>(TITLE_MAP.values());
		ArrayList<String> cmds = new ArrayList<String>(titles.size());

		Collections.sort(titles);
		for (String cmd : KEY_MAP.keySet()) {
			titleToCmdMap.put(getTitle(cmd), cmd);
		}
		for (String title : titles) {
			cmds.add(titleToCmdMap.get(title));
		}
		return cmds.toArray(new String[0]);
	}

	/** @return Whether the key map is currently set to defaults or not. */
	public static final boolean isSetToDefaults() {
		return KEY_MAP.equals(STD_KEY_MAP);
	}
}
