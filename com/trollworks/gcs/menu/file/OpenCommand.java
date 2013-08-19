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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.utility.Debug;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.StdFileDialog;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.File;
import java.io.IOException;

import javax.swing.JMenuItem;

/** Provides the "Open..." command. */
public class OpenCommand extends Command {
	private static String			MSG_OPEN;

	static {
		LocalizedMessages.initialize(OpenCommand.class);
	}

	/** The singleton {@link OpenCommand}. */
	public static final OpenCommand	INSTANCE			= new OpenCommand();
	private static final String		CHARACTER_MATCHER	= StdFileDialog.createExtensionMatcher(SheetWindow.SHEET_EXTENSION);
	private static final String		TEMPLATE_MATCHER	= StdFileDialog.createExtensionMatcher(TemplateWindow.EXTENSION);
	private static final String		LIBRARY_MATCHER		= StdFileDialog.createExtensionMatcher(LibraryFile.EXTENSION);
	private static final String		ADVANTAGE_MATCHER	= StdFileDialog.createExtensionMatcher(Advantage.OLD_ADVANTAGE_EXTENSION);
	private static final String		SKILL_MATCHER		= StdFileDialog.createExtensionMatcher(Skill.OLD_SKILL_EXTENSION);
	private static final String		SPELL_MATCHER		= StdFileDialog.createExtensionMatcher(Spell.OLD_SPELL_EXTENSION);
	private static final String		EQUIPMENT_MATCHER	= StdFileDialog.createExtensionMatcher(Equipment.OLD_EQUIPMENT_EXTENSION);

	private OpenCommand() {
		super(MSG_OPEN, KeyEvent.VK_O);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		// Do nothing. Always enabled.
	}

	@Override public void actionPerformed(ActionEvent event) {
		open();
	}

	/** Ask the user to open a file. */
	public void open() {
		Component focus = getFocusOwner();
		open(StdFileDialog.choose(focus, true, MSG_OPEN, null, null, SheetWindow.SHEET_EXTENSION, LibraryFile.EXTENSION, TemplateWindow.EXTENSION, Advantage.OLD_ADVANTAGE_EXTENSION, Skill.OLD_SKILL_EXTENSION, Spell.OLD_SPELL_EXTENSION, Equipment.OLD_EQUIPMENT_EXTENSION));
	}

	/** @param file The file to open. */
	public void open(File file) {
		if (file != null) {
			try {
				String name = file.getName();
				if (name.matches(CHARACTER_MATCHER)) {
					openCharacterSheet(file);
				} else if (name.matches(LIBRARY_MATCHER) || name.matches(ADVANTAGE_MATCHER) || name.matches(SKILL_MATCHER) || name.matches(SPELL_MATCHER) || name.matches(EQUIPMENT_MATCHER)) {
					openLibrary(file);
				} else if (name.matches(TEMPLATE_MATCHER)) {
					openTemplateSheet(file);
				} else {
					throw new IOException("Unknown file extension"); //$NON-NLS-1$
				}
			} catch (Exception exception) {
				Debug.diagnoseLoadAndSave(exception);
				StdFileDialog.showCannotOpenMsg(getFocusOwner(), file.getName(), exception);
			}
		}
	}

	private void openCharacterSheet(File file) throws IOException {
		SheetWindow window = SheetWindow.findSheetWindow(file);
		if (window == null) {
			window = new SheetWindow(new GURPSCharacter(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
		RecentFilesMenu.addRecent(file);
	}

	private void openLibrary(File file) throws IOException {
		LibraryWindow window = LibraryWindow.findLibraryWindow(file);
		if (window == null) {
			window = new LibraryWindow(new LibraryFile(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
		RecentFilesMenu.addRecent(file);
	}

	private void openTemplateSheet(File file) throws IOException {
		TemplateWindow window = TemplateWindow.findTemplateWindow(file);
		if (window == null) {
			window = new TemplateWindow(new Template(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
		RecentFilesMenu.addRecent(file);
	}
}
