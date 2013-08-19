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

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.names.Namer;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.ttk.collections.FilteredList;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.Selection;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.OutlineProxy;
import com.trollworks.ttk.widgets.outline.Row;

import java.awt.EventQueue;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.Collection;

import javax.swing.undo.StateEdit;

/** Base outline class. */
public class ListOutline extends Outline implements Runnable, ActionListener {
	private static String	MSG_CLEAR_UNDO;
	/** The owning data file. */
	protected DataFile		mDataFile;
	private String			mRowSetChangedID;

	static {
		LocalizedMessages.initialize(ListOutline.class);
	}

	/**
	 * Create a new outline.
	 * 
	 * @param dataFile The owning data file.
	 * @param model The outline model to use.
	 * @param rowSetChangedID The notification ID to use when the row set changes.
	 */
	public ListOutline(DataFile dataFile, OutlineModel model, String rowSetChangedID) {
		super(model);
		mDataFile = dataFile;
		mRowSetChangedID = rowSetChangedID;
		addActionListener(this);
	}

	/** @return The owning data file. */
	public DataFile getDataFile() {
		return mDataFile;
	}

	@Override
	public void rowsAdded(OutlineModel model, Row[] rows) {
		super.rowsAdded(model, rows);
		mDataFile.notifySingle(mRowSetChangedID, null);
	}

	@Override
	public void rowsWereRemoved(OutlineModel model, Row[] rows) {
		super.rowsWereRemoved(model, rows);
		mDataFile.notifySingle(mRowSetChangedID, null);
	}

	/** @return The notification ID to use when the row set changes. */
	public String getRowSetChangedID() {
		return mRowSetChangedID;
	}

	@Override
	public boolean canDeleteSelection() {
		OutlineModel model = getModel();
		return !model.isLocked() && model.hasSelection();
	}

	@Override
	public void deleteSelection() {
		if (canDeleteSelection()) {
			OutlineModel model = getModel();
			StateEdit edit = new StateEdit(model, MSG_CLEAR_UNDO);
			Row[] rows = model.getSelectionAsList(true).toArray(new Row[0]);
			mDataFile.startNotify();
			model.removeSelection();
			for (int i = rows.length - 1; i >= 0; i--) {
				rows[i].removeFromParent();
			}
			if (model.getRowCount() > 0) {
				updateAllRows();
			}
			// Send it out again, since we have a few chicken-and-egg
			// scenarios to deal with... <sigh>
			mDataFile.notify(mRowSetChangedID, null);
			mDataFile.endNotify();
			edit.end();
			postUndo(edit);
		}
	}

	/**
	 * Adds a row at the "best" place (i.e. looks at the selection).
	 * 
	 * @param row The row to add.
	 * @param name The name for the undo event.
	 * @param sibling If the current selection is a container, whether to insert into it, or as a
	 *            sibling.
	 * @return The index of the row that was added.
	 */
	public int addRow(ListRow row, String name, boolean sibling) {
		return addRow(new ListRow[] { row }, name, sibling);
	}

	/**
	 * Adds rows at the "best" place (i.e. looks at the selection).
	 * 
	 * @param rows The rows to add.
	 * @param name The name for the undo event.
	 * @param sibling If the current selection is a container, whether to insert into it, or as a
	 *            sibling.
	 * @return The index of the first row that was added.
	 */
	public int addRow(ListRow[] rows, String name, boolean sibling) {
		OutlineModel model = getModel();
		StateEdit edit = new StateEdit(model, name);
		Selection selection = model.getSelection();
		int count = selection.getCount();
		int insertAt;
		int i;
		Row parentRow;

		if (count > 0) {
			insertAt = count == 1 ? selection.firstSelectedIndex() : selection.lastSelectedIndex();
			parentRow = model.getRowAtIndex(insertAt++);
			if (!parentRow.canHaveChildren() || !parentRow.isOpen()) {
				parentRow = parentRow.getParent();
			} else if (sibling) {
				insertAt += parentRow.getChildCount();
				parentRow = parentRow.getParent();
			}
			if (parentRow != null && parentRow.canHaveChildren()) {
				for (ListRow row : rows) {
					parentRow.addChild(row);
				}
			}
		} else {
			insertAt = model.getRowCount();
		}

		i = insertAt;
		for (ListRow row : rows) {
			model.addRow(i++, row, true);
		}
		updateAllRows();
		edit.end();
		postUndo(edit);
		revalidate();
		return insertAt;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();
		String command = event.getActionCommand();

		if (source instanceof OutlineProxy) {
			source = ((OutlineProxy) source).getRealOutline();
		}
		if (source == this && Outline.CMD_OPEN_SELECTION.equals(command)) {
			openDetailEditor(false);
		}
	}

	/**
	 * Opens detailed editors for the current selection.
	 * 
	 * @param later Whether to call {@link EventQueue#invokeLater(Runnable)} rather than immediately
	 *            opening the editor.
	 */
	public void openDetailEditor(boolean later) {
		mRowsToEdit = new FilteredList<>(getModel().getSelectionAsList(), ListRow.class);
		if (later) {
			EventQueue.invokeLater(this);
		} else {
			run();
		}
	}

	private ArrayList<ListRow>	mRowsToEdit;

	@Override
	public void run() {
		if (RowEditor.edit(this, mRowsToEdit)) {
			if (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) {
				Namer.name(this, mRowsToEdit);
			}
			updateRows(mRowsToEdit);
			updateRowHeights(mRowsToEdit);
			repaint();
			mDataFile.notifySingle(mRowSetChangedID, null);
		}
		mRowsToEdit = null;
	}

	@Override
	public void undoDidHappen(OutlineModel model) {
		super.undoDidHappen(model);
		updateAllRows();
		mDataFile.notifySingle(mRowSetChangedID, null);
	}

	@Override
	protected void rowsWereDropped() {
		updateAllRows();
		mDataFile.notifySingle(mRowSetChangedID, null);
		requestFocusInWindow();
	}

	private void updateAllRows() {
		updateRows(new FilteredList<>(getModel().getTopLevelRows(), ListRow.class));
	}

	private void updateRows(Collection<ListRow> rows) {
		for (ListRow row : rows) {
			updateRow(row);
		}
	}

	private void updateRow(ListRow row) {
		if (row != null) {
			int count = row.getChildCount();
			for (int i = 0; i < count; i++) {
				ListRow child = (ListRow) row.getChild(i);
				updateRow(child);
			}
			row.update();
		}
	}

	/**
	 * Adds the row to the list. Recursively descends the row's children and does the same with
	 * them.
	 * 
	 * @param list The list to add rows to.
	 * @param row The row to check.
	 */
	protected void addRowsToBeProcessed(ArrayList<ListRow> list, ListRow row) {
		int count = row.getChildCount();
		list.add(row);
		for (int i = 0; i < count; i++) {
			addRowsToBeProcessed(list, (ListRow) row.getChild(i));
		}
	}

	@Override
	public void sorted(OutlineModel model, boolean restoring) {
		super.sorted(model, restoring);
		if (!restoring) {
			mDataFile.setModified(true);
		}
	}
}
