/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentDockable;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Equipment" command. */
public class NewEquipmentCommand extends Command {
	@Localize("New Equipment")
	@Localize(locale = "de", value = "Neue Ausrüstung")
	@Localize(locale = "ru", value = "Новое снаряжение")
	private static String					EQUIPMENT;
	@Localize("New Equipment Container")
	@Localize(locale = "de", value = "Neuer Ausrüstungs-Container")
	@Localize(locale = "ru", value = "Новый контейнер снаряжения")
	private static String					EQUIPMENT_CONTAINER;

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
		EquipmentDockable equipment = getTarget(EquipmentDockable.class);
		if (equipment != null) {
			setEnabled(!equipment.getOutline().getModel().isLocked());
		} else {
			SheetDockable sheet = getTarget(SheetDockable.class);
			if (sheet != null) {
				setEnabled(true);
			} else {
				setEnabled(getTarget(TemplateDockable.class) != null);
			}
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		ListOutline outline;
		DataFile dataFile;
		EquipmentDockable eqpDockable = getTarget(EquipmentDockable.class);
		if (eqpDockable != null) {
			dataFile = eqpDockable.getDataFile();
			outline = eqpDockable.getOutline();
			if (outline.getModel().isLocked()) {
				return;
			}
		} else {
			SheetDockable sheet = getTarget(SheetDockable.class);
			if (sheet != null) {
				dataFile = sheet.getDataFile();
				outline = sheet.getSheet().getEquipmentOutline();
			} else {
				TemplateDockable template = getTarget(TemplateDockable.class);
				if (template != null) {
					dataFile = template.getDataFile();
					outline = template.getTemplate().getEquipmentOutline();
				} else {
					return;
				}
			}
		}
		Equipment equipment = new Equipment(dataFile, mContainer);
		outline.addRow(equipment, getTitle(), false);
		outline.getModel().select(equipment, false);
		outline.scrollSelectionIntoView();
		outline.openDetailEditor(true);
	}
}
