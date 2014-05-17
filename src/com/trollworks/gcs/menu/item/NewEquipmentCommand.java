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

package com.trollworks.gcs.menu.item;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.toolkit.ui.menu.Command;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Equipment" command. */
public class NewEquipmentCommand extends Command {
	@Localize("New Equipment")
	private static String EQUIPMENT;
	@Localize("New Equipment Container")
	private static String EQUIPMENT_CONTAINER;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String				CMD_NEW_EQUIPMENT			= "NewEquipment";																											//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String				CMD_NEW_EQUIPMENT_CONTAINER	= "NewEquipmentContainer";																									//$NON-NLS-1$
	/** The "New Carried Equipment" command. */
	public static final NewEquipmentCommand	CARRIED_INSTANCE			= new NewEquipmentCommand(false, EQUIPMENT, CMD_NEW_EQUIPMENT, KeyEvent.VK_E, COMMAND_MODIFIER);
	/** The "New Carried Equipment Container" command. */
	public static final NewEquipmentCommand	CARRIED_CONTAINER_INSTANCE	= new NewEquipmentCommand(true, EQUIPMENT_CONTAINER, CMD_NEW_EQUIPMENT_CONTAINER, KeyEvent.VK_E, SHIFTED_COMMAND_MODIFIER);
	private boolean							mContainer;

	private NewEquipmentCommand(boolean container, String title, String cmd, int keyCode, int modifiers) {
		super(title, cmd, keyCode, modifiers);
		mContainer = container;
	}

	@Override
	public void adjust() {
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
