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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.edit.Incrementable;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.utility.collections.FilteredIterator;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.MultipleRowUndo;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.Row;
import com.trollworks.gcs.widgets.outline.RowPostProcessor;
import com.trollworks.gcs.widgets.outline.RowUndo;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for equipment. */
public class EquipmentOutline extends ListOutline implements Incrementable {
	private static String	MSG_INCREMENT;
	private static String	MSG_DECREMENT;

	static {
		LocalizedMessages.initialize(EquipmentOutline.class);
	}

	private static OutlineModel extractModel(DataFile dataFile) {
		if (dataFile instanceof GURPSCharacter) {
			return ((GURPSCharacter) dataFile).getEquipmentRoot();
		}
		if (dataFile instanceof Template) {
			return ((Template) dataFile).getEquipmentModel();
		}
		if (dataFile instanceof LibraryFile) {
			return ((LibraryFile) dataFile).getEquipmentList().getModel();
		}
		return ((ListFile) dataFile).getModel();
	}

	/**
	 * Create a new equipment outline.
	 * 
	 * @param dataFile The owning data file.
	 */
	public EquipmentOutline(DataFile dataFile) {
		this(dataFile, extractModel(dataFile));
	}

	/**
	 * Create a new equipment outline.
	 * 
	 * @param dataFile The owning data file.
	 * @param model The {@link OutlineModel} to use.
	 */
	public EquipmentOutline(DataFile dataFile, OutlineModel model) {
		super(dataFile, model, Equipment.ID_LIST_CHANGED);
		EquipmentColumn.addColumns(this, dataFile);
	}

	public String getDecrementTitle() {
		return MSG_DECREMENT;
	}

	public String getIncrementTitle() {
		return MSG_INCREMENT;
	}

	public boolean canDecrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeafRows(true);
	}

	public boolean canIncrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeafRows(false);
	}

	private boolean selectionHasLeafRows(boolean requireQtyAboveZero) {
		for (Equipment equipment : new FilteredIterator<Equipment>(getModel().getSelectionAsList(), Equipment.class)) {
			if (!equipment.canHaveChildren() && (!requireQtyAboveZero || equipment.getQuantity() > 0)) {
				return true;
			}
		}
		return false;
	}

	public void decrement() {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();

		for (Equipment equipment : new FilteredIterator<Equipment>(getModel().getSelectionAsList(), Equipment.class)) {
			if (!equipment.canHaveChildren()) {
				int qty = equipment.getQuantity();

				if (qty > 0) {
					RowUndo undo = new RowUndo(equipment);

					equipment.setQuantity(qty - 1);
					if (undo.finish()) {
						undos.add(undo);
					}
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new MultipleRowUndo(undos);
		}
	}

	public void increment() {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();

		for (Equipment equipment : new FilteredIterator<Equipment>(getModel().getSelectionAsList(), Equipment.class)) {
			if (!equipment.canHaveChildren()) {
				RowUndo undo = new RowUndo(equipment);

				equipment.setQuantity(equipment.getQuantity() + 1);
				if (undo.finish()) {
					undos.add(undo);
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new MultipleRowUndo(undos);
		}
	}

	@Override protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof Equipment;
	}

	@Override public void convertDragRowsToSelf(List<Row> list) {
		OutlineModel model = getModel();
		Row[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof GURPSCharacter || mDataFile instanceof Template;
		ArrayList<ListRow> process = forSheetOrTemplate ? new ArrayList<ListRow>() : null;

		for (Row element : rows) {
			Equipment equipment = new Equipment(mDataFile, (Equipment) element, true);

			model.collectRowsAndSetOwner(list, equipment, false);
			if (forSheetOrTemplate) {
				addRowsToBeProcessed(process, equipment);
			}
		}

		if (forSheetOrTemplate && !process.isEmpty()) {
			EventQueue.invokeLater(new RowPostProcessor(this, process));
		}
	}
}
