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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.utility.LocalizedMessages;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "New Equipment" command. */
public class NewEquipmentCommand extends Command {
	/** The action command this command will issue. */
	public static final String				CMD_NEW_EQUIPMENT			= "NewEquipment";																												//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String				CMD_NEW_EQUIPMENT_CONTAINER	= "NewEquipmentContainer";																										//$NON-NLS-1$
	private static String					MSG_EQUIPMENT;
	private static String					MSG_EQUIPMENT_CONTAINER;

	static {
		LocalizedMessages.initialize(NewEquipmentCommand.class);
	}

	/** The "New Carried Equipment" command. */
	public static final NewEquipmentCommand	CARRIED_INSTANCE			= new NewEquipmentCommand(false, MSG_EQUIPMENT, CMD_NEW_EQUIPMENT, KeyEvent.VK_E, COMMAND_MODIFIER);
	/** The "New Carried Equipment Container" command. */
	public static final NewEquipmentCommand	CARRIED_CONTAINER_INSTANCE	= new NewEquipmentCommand(true, MSG_EQUIPMENT_CONTAINER, CMD_NEW_EQUIPMENT_CONTAINER, KeyEvent.VK_E, SHIFTED_COMMAND_MODIFIER);
	private boolean							mContainer;

	private NewEquipmentCommand(boolean container, String title, String cmd, int keyCode, int modifiers) {
		super(title, cmd, keyCode, modifiers);
		mContainer = container;
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			setEnabled(!((LibraryWindow) window).getOutline().getModel().isLocked());
		} else {
			setEnabled(window instanceof SheetWindow || window instanceof TemplateWindow);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		ListOutline outline;
		DataFile dataFile;

		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			LibraryWindow libraryWindow = (LibraryWindow) window;
			libraryWindow.switchToEquipment();
			dataFile = libraryWindow.getLibraryFile();
			outline = libraryWindow.getOutline();
		} else if (window instanceof SheetWindow) {
			SheetWindow sheetWindow = (SheetWindow) window;
			CharacterSheet sheet = sheetWindow.getSheet();
			outline = sheet.getEquipmentOutline();
			dataFile = sheetWindow.getCharacter();
		} else if (window instanceof TemplateWindow) {
			TemplateWindow templateWindow = (TemplateWindow) window;
			outline = templateWindow.getSheet().getEquipmentOutline();
			dataFile = templateWindow.getTemplate();
		} else {
			return;
		}

		Equipment equipment = new Equipment(dataFile, mContainer);
		outline.addRow(equipment, getTitle(), false);
		outline.getModel().select(equipment, false);
		outline.scrollSelectionIntoView();
		outline.openDetailEditor(true);
	}
}
