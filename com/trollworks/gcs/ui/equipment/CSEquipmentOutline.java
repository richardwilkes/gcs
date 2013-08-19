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

package com.trollworks.gcs.ui.equipment;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMMultipleRowUndo;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.CMRowUndo;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.ui.common.CSNamePostProcessor;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.gcs.ui.common.CSWindow;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for equipment. */
public class CSEquipmentOutline extends CSOutline {
	private boolean	mIsCarried;

	private static TKOutlineModel extractModel(CMDataFile dataFile, boolean isCarried) {
		if (dataFile instanceof CMCharacter) {
			CMCharacter character = (CMCharacter) dataFile;

			return isCarried ? character.getCarriedEquipmentRoot() : character.getOtherEquipmentRoot();
		}
		if (dataFile instanceof CMTemplate) {
			return ((CMTemplate) dataFile).getEquipmentModel();
		}
		return ((CMListFile) dataFile).getModel();
	}

	/**
	 * Create a new equipment outline.
	 * 
	 * @param dataFile The owning data file.
	 * @param isCarried Whether or not this is the carried equipment.
	 */
	public CSEquipmentOutline(CMDataFile dataFile, boolean isCarried) {
		super(dataFile, extractModel(dataFile, isCarried), CMEquipment.ID_LIST_CHANGED);
		mIsCarried = isCarried;
		CSEquipmentColumnID.addColumns(this, dataFile, isCarried);
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		boolean forSheetOrTemplate = mDataFile instanceof CMCharacter || mDataFile instanceof CMTemplate;

		if (CSWindow.CMD_INCREMENT.equals(command)) {
			item.setTitle(Msgs.INCREMENT);
			item.setEnabled(forSheetOrTemplate && selectionHasLeafRows(false));
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			item.setTitle(Msgs.DECREMENT);
			item.setEnabled(forSheetOrTemplate && selectionHasLeafRows(true));
		} else if (CSWindow.CMD_TOGGLE_EQUIPPED.equals(command)) {
			item.setEnabled(mDataFile instanceof CMCharacter && mIsCarried && getModel().hasSelection());
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CSWindow.CMD_INCREMENT.equals(command)) {
			incrementQty();
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			decrementQty();
		} else if (CSWindow.CMD_TOGGLE_EQUIPPED.equals(command)) {
			toggleEquipped();
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private void toggleEquipped() {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();

		for (CMEquipment equipment : new TKFilteredIterator<CMEquipment>(getModel().getSelectionAsList(), CMEquipment.class)) {
			CMRowUndo undo = new CMRowUndo(equipment);

			equipment.setEquipped(!equipment.isEquipped());
			if (undo.finish()) {
				undos.add(undo);
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new CMMultipleRowUndo(undos);
		}
	}

	private boolean selectionHasLeafRows(boolean requireQtyAboveZero) {
		for (CMEquipment equipment : new TKFilteredIterator<CMEquipment>(getModel().getSelectionAsList(), CMEquipment.class)) {
			if (!equipment.canHaveChildren() && (!requireQtyAboveZero || equipment.getQuantity() > 0)) {
				return true;
			}
		}
		return false;
	}

	private void incrementQty() {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();

		for (CMEquipment equipment : new TKFilteredIterator<CMEquipment>(getModel().getSelectionAsList(), CMEquipment.class)) {
			if (!equipment.canHaveChildren()) {
				CMRowUndo undo = new CMRowUndo(equipment);

				equipment.setQuantity(equipment.getQuantity() + 1);
				if (undo.finish()) {
					undos.add(undo);
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new CMMultipleRowUndo(undos);
		}
	}

	private void decrementQty() {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();

		for (CMEquipment equipment : new TKFilteredIterator<CMEquipment>(getModel().getSelectionAsList(), CMEquipment.class)) {
			if (!equipment.canHaveChildren()) {
				int qty = equipment.getQuantity();

				if (qty > 0) {
					CMRowUndo undo = new CMRowUndo(equipment);

					equipment.setQuantity(qty - 1);
					if (undo.finish()) {
						undos.add(undo);
					}
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new CMMultipleRowUndo(undos);
		}
	}

	/** @return Whether or not this is the carried equipment. */
	public boolean isCarried() {
		return mIsCarried;
	}

	@Override protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, TKRow[] rows) {
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof CMEquipment;
	}

	@Override protected void convertDragRowsToSelf(List<TKRow> list) {
		TKOutlineModel model = getModel();
		TKRow[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof CMCharacter || mDataFile instanceof CMTemplate;
		ArrayList<CMRow> process = forSheetOrTemplate ? new ArrayList<CMRow>() : null;

		for (TKRow element : rows) {
			CMEquipment equipment = new CMEquipment(mDataFile, (CMEquipment) element, true);

			model.collectRowsAndSetOwner(list, equipment, false);
			if (forSheetOrTemplate) {
				addRowsToBeProcessed(process, equipment);
			}
		}

		if (forSheetOrTemplate && !process.isEmpty()) {
			EventQueue.invokeLater(new CSNamePostProcessor(this, process));
		}
	}
}
