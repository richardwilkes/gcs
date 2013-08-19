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

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.advantage.AdvantageListWindow;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.ListWindow;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentListWindow;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.SkillListWindow;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.spell.SpellListWindow;
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
	public static final OpenCommand	INSTANCE	= new OpenCommand();

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
		open(StdFileDialog.choose(focus, true, MSG_OPEN, null, null, SheetWindow.SHEET_EXTENSION, TemplateWindow.EXTENSION, AdvantageListWindow.EXTENSION, SkillListWindow.EXTENSION, SpellListWindow.EXTENSION, EquipmentListWindow.EXTENSION));
	}

	/** @param file The file to open. */
	public void open(File file) {
		if (file != null) {
			try {
				String name = file.getName();
				if (name.matches(StdFileDialog.createExtensionMatcher(SheetWindow.SHEET_EXTENSION))) {
					openCharacterSheet(file);
				} else if (name.matches(StdFileDialog.createExtensionMatcher(TemplateWindow.EXTENSION))) {
					openTemplateSheet(file);
				} else if (name.matches(StdFileDialog.createExtensionMatcher(AdvantageListWindow.EXTENSION))) {
					openAdvantageList(file);
				} else if (name.matches(StdFileDialog.createExtensionMatcher(SkillListWindow.EXTENSION))) {
					openSkillList(file);
				} else if (name.matches(StdFileDialog.createExtensionMatcher(SpellListWindow.EXTENSION))) {
					openSpellList(file);
				} else if (name.matches(StdFileDialog.createExtensionMatcher(EquipmentListWindow.EXTENSION))) {
					openEquipmentList(file);
				} else {
					throw new IOException("Unknown file extension"); //$NON-NLS-1$
				}
			} catch (Exception exception) {
				Debug.diagnoseLoadAndSave(exception);
				StdFileDialog.showCannotOpenMsg(getFocusOwner(), file.getName());
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
	}

	private void openTemplateSheet(File file) throws IOException {
		TemplateWindow window = TemplateWindow.findTemplateWindow(file);
		if (window == null) {
			window = new TemplateWindow(new Template(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
	}

	private void openAdvantageList(File file) throws IOException {
		AdvantageListWindow window = (AdvantageListWindow) ListWindow.findListWindow(file);
		if (window == null) {
			window = new AdvantageListWindow(new AdvantageList(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
	}

	private void openSkillList(File file) throws IOException {
		SkillListWindow window = (SkillListWindow) ListWindow.findListWindow(file);
		if (window == null) {
			window = new SkillListWindow(new SkillList(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
	}

	private void openSpellList(File file) throws IOException {
		SpellListWindow window = (SpellListWindow) ListWindow.findListWindow(file);
		if (window == null) {
			window = new SpellListWindow(new SpellList(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
	}

	private void openEquipmentList(File file) throws IOException {
		EquipmentListWindow window = (EquipmentListWindow) ListWindow.findListWindow(file);
		if (window == null) {
			window = new EquipmentListWindow(new EquipmentList(file));
			window.setVisible(true);
		} else {
			window.toFront();
		}
	}
}
