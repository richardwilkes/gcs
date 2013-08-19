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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.edit.Incrementable;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.MultipleRowUndo;
import com.trollworks.gcs.widgets.outline.RowPostProcessor;
import com.trollworks.gcs.widgets.outline.RowUndo;
import com.trollworks.ttk.collections.FilteredIterator;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.Row;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for Advantages. */
public class AdvantageOutline extends ListOutline implements Incrementable {
	private static String	MSG_INCREMENT;
	private static String	MSG_DECREMENT;

	static {
		LocalizedMessages.initialize(AdvantageOutline.class);
	}

	private static OutlineModel extractModel(DataFile dataFile) {
		if (dataFile instanceof GURPSCharacter) {
			return ((GURPSCharacter) dataFile).getAdvantagesModel();
		}
		if (dataFile instanceof Template) {
			return ((Template) dataFile).getAdvantagesModel();
		}
		if (dataFile instanceof LibraryFile) {
			return ((LibraryFile) dataFile).getAdvantageList().getModel();
		}
		return ((ListFile) dataFile).getModel();
	}

	/**
	 * Create a new Advantages, Disadvantages & Quirks outline.
	 * 
	 * @param dataFile The owning data file.
	 */
	public AdvantageOutline(DataFile dataFile) {
		this(dataFile, extractModel(dataFile));
	}

	/**
	 * Create a new Advantages, Disadvantages & Quirks outline.
	 * 
	 * @param dataFile The owning data file.
	 * @param model The {@link OutlineModel} to use.
	 */
	public AdvantageOutline(DataFile dataFile, OutlineModel model) {
		super(dataFile, model, Advantage.ID_LIST_CHANGED);
		AdvantageColumn.addColumns(this, dataFile);
	}

	@Override
	protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof Advantage;
	}

	@Override
	public void convertDragRowsToSelf(List<Row> list) {
		OutlineModel model = getModel();
		Row[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof GURPSCharacter || mDataFile instanceof Template;
		ArrayList<ListRow> process = forSheetOrTemplate ? new ArrayList<ListRow>() : null;

		for (Row element : rows) {
			Advantage advantage = new Advantage(mDataFile, (Advantage) element, true);

			model.collectRowsAndSetOwner(list, advantage, false);
			if (forSheetOrTemplate) {
				addRowsToBeProcessed(process, advantage);
			}
		}

		if (forSheetOrTemplate && !process.isEmpty()) {
			EventQueue.invokeLater(new RowPostProcessor(this, process));
		}
	}

	@Override
	public String getIncrementTitle() {
		return MSG_INCREMENT;
	}

	@Override
	public String getDecrementTitle() {
		return MSG_DECREMENT;
	}

	@Override
	public boolean canIncrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeveledRows(false);
	}

	@Override
	public boolean canDecrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeveledRows(true);
	}

	private boolean selectionHasLeveledRows(boolean requireLevelAboveZero) {
		for (Advantage advantage : new FilteredIterator<Advantage>(getModel().getSelectionAsList(), Advantage.class)) {
			if (!advantage.canHaveChildren() && advantage.isLeveled() && (!requireLevelAboveZero || advantage.getLevels() > 0)) {
				return true;
			}
		}
		return false;
	}

	@Override
	public void increment() {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();
		for (Advantage advantage : new FilteredIterator<Advantage>(getModel().getSelectionAsList(), Advantage.class)) {
			if (!advantage.canHaveChildren() && advantage.isLeveled()) {
				RowUndo undo = new RowUndo(advantage);

				advantage.setLevels(advantage.getLevels() + 1);
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

	@Override
	public void decrement() {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();
		for (Advantage advantage : new FilteredIterator<Advantage>(getModel().getSelectionAsList(), Advantage.class)) {
			if (!advantage.canHaveChildren() && advantage.isLeveled()) {
				int levels = advantage.getLevels();

				if (levels > 0) {
					RowUndo undo = new RowUndo(advantage);

					advantage.setLevels(levels - 1);
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
}
