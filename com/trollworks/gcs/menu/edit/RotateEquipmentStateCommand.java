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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.equipment.EquipmentState;
import com.trollworks.gcs.widgets.outline.MultipleRowUndo;
import com.trollworks.gcs.widgets.outline.RowUndo;
import com.trollworks.ttk.collections.FilteredIterator;
import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.OutlineProxy;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

import javax.swing.JMenuItem;

/** Provides the "Rotate Equipment State" command. */
public class RotateEquipmentStateCommand extends Command {
	/** The action command this command will issue. */
	public static final String						CMD_ROTATE_EQUIPMENT_STATE	= "RotateEquipmentState";				//$NON-NLS-1$
	private static String							MSG_TITLE;

	static {
		LocalizedMessages.initialize(RotateEquipmentStateCommand.class);
	}

	/** The singleton {@link RotateEquipmentStateCommand}. */
	public static final RotateEquipmentStateCommand	INSTANCE					= new RotateEquipmentStateCommand();

	private RotateEquipmentStateCommand() {
		super(MSG_TITLE, CMD_ROTATE_EQUIPMENT_STATE, KeyEvent.VK_QUOTE);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Component focus = getFocusOwner();
		if (focus instanceof OutlineProxy) {
			focus = ((OutlineProxy) focus).getRealOutline();
		}
		if (focus instanceof EquipmentOutline) {
			EquipmentOutline outline = (EquipmentOutline) focus;
			setEnabled(outline.getDataFile() instanceof GURPSCharacter && outline.getModel().hasSelection());
		} else {
			setEnabled(false);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Component focus = getFocusOwner();
		if (focus instanceof OutlineProxy) {
			focus = ((OutlineProxy) focus).getRealOutline();
		}
		EquipmentOutline outline = (EquipmentOutline) focus;
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();
		for (Equipment equipment : new FilteredIterator<Equipment>(outline.getModel().getSelectionAsList(), Equipment.class)) {
			RowUndo undo = new RowUndo(equipment);
			EquipmentState[] values = EquipmentState.values();
			int index = equipment.getState().ordinal() - 1;
			if (index < 0) {
				index = values.length - 1;
			}
			equipment.setState(values[index]);
			if (undo.finish()) {
				undos.add(undo);
			}
		}
		if (!undos.isEmpty()) {
			outline.repaintSelection();
			new MultipleRowUndo(undos);
		}
	}
}
