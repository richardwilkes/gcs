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

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.toolkit.collections.TKFilteredList;
import com.trollworks.toolkit.undo.TKUndo;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKSelection;
import com.trollworks.toolkit.widget.border.TKBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKOutlineModelUndo;
import com.trollworks.toolkit.widget.outline.TKOutlineModelUndoSnapshot;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKProxyOutline;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Color;
import java.awt.EventQueue;
import java.awt.Point;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.event.ActionEvent;
import java.util.ArrayList;
import java.util.Collection;

/** Base outline class. */
public class CSOutline extends TKOutline implements Runnable {
	/** The owning data file. */
	protected CMDataFile	mDataFile;
	private String			mRowSetChangedID;
	private TKBorder		mSavedBorder;

	/**
	 * Create a new outline.
	 * 
	 * @param dataFile The owning data file.
	 * @param model The outline model to use.
	 * @param rowSetChangedID The notification ID to use when the row set changes.
	 */
	public CSOutline(CMDataFile dataFile, TKOutlineModel model, String rowSetChangedID) {
		super(model);
		mDataFile = dataFile;
		mRowSetChangedID = rowSetChangedID;
		addActionListener(this);
	}

	@Override public void rowsAdded(TKOutlineModel model, TKRow[] rows) {
		super.rowsAdded(model, rows);
		mDataFile.notifySingle(mRowSetChangedID, null);
	}

	@Override public void rowsWereRemoved(TKOutlineModel model, TKRow[] rows) {
		super.rowsWereRemoved(model, rows);
		mDataFile.notifySingle(mRowSetChangedID, null);
	}

	@Override public void lockedStateWillChange(TKOutlineModel model) {
		stopEditing();
	}

	/** @return The notification ID to use when the row set changes. */
	public String getRowSetChangedID() {
		return mRowSetChangedID;
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (TKWindow.CMD_CLEAR.equals(command) || TKWindow.CMD_DUPLICATE.equals(command)) {
			TKOutlineModel model = getModel();

			item.setEnabled(!model.isLocked() && model.hasSelection());
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (TKWindow.CMD_CLEAR.equals(command)) {
			handleClearCmd();
		} else if (TKWindow.CMD_DUPLICATE.equals(command)) {
			handleDuplicateCmd();
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private void handleClearCmd() {
		TKOutlineModel model = getModel();

		if (!model.isLocked() && model.hasSelection()) {
			TKOutlineModelUndoSnapshot before = new TKOutlineModelUndoSnapshot(model);
			TKRow[] rows = model.getSelectionAsList(true).toArray(new TKRow[0]);

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
			postUndo(new TKOutlineModelUndo(Msgs.CLEAR_UNDO, model, before, new TKOutlineModelUndoSnapshot(model)));
		}
	}

	private void handleDuplicateCmd() {
		TKOutlineModel model = getModel();

		if (!model.isLocked() && model.hasSelection()) {
			ArrayList<TKRow> rows = new ArrayList<TKRow>();
			ArrayList<TKRow> topRows = new ArrayList<TKRow>();

			mDataFile.startNotify();
			model.setDragRows(model.getSelectionAsList(true).toArray(new TKRow[0]));
			convertDragRowsToSelf(rows);
			model.setDragRows(null);
			for (TKRow row : rows) {
				if (row.getDepth() == 0) {
					topRows.add(row);
				}
			}
			addRow(topRows.toArray(new CMRow[0]), Msgs.DUPLICATE_UNDO, true);
			model.select(topRows, false);
			scrollSelectionIntoView();
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
	public int addRow(CMRow row, String name, boolean sibling) {
		return addRow(new CMRow[] { row }, name, sibling);
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
	public int addRow(CMRow[] rows, String name, boolean sibling) {
		TKOutlineModel model = getModel();
		TKOutlineModelUndoSnapshot before = new TKOutlineModelUndoSnapshot(model);
		Point cell = getCellLocationOfEditor();
		TKSelection selection = model.getSelection();
		int count = selection.getCount();
		int insertAt;
		int i;
		TKRow parentRow;

		if (count > 0 || cell != null) {
			insertAt = cell != null ? cell.y : count == 1 ? selection.firstSelectedIndex() : selection.lastSelectedIndex();
			parentRow = model.getRowAtIndex(insertAt++);
			if (!parentRow.canHaveChildren() || !parentRow.isOpen()) {
				parentRow = parentRow.getParent();
			} else if (sibling) {
				insertAt += parentRow.getChildCount();
				parentRow = parentRow.getParent();
			}
			if (parentRow != null && parentRow.canHaveChildren()) {
				for (CMRow row : rows) {
					parentRow.addChild(row);
				}
			}
		} else {
			insertAt = model.getRowCount();
		}

		i = insertAt;
		for (CMRow row : rows) {
			model.addRow(i++, row, true);
		}
		updateAllRows();
		postUndo(new TKOutlineModelUndo(name, model, before, new TKOutlineModelUndoSnapshot(model)));
		revalidateImmediately();
		return insertAt;
	}

	@Override public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();
		String command = event.getActionCommand();

		if (source instanceof TKProxyOutline) {
			source = ((TKProxyOutline) source).getRealOutline();
		}
		if (source == this && TKOutline.CMD_OPEN_SELECTION.equals(command)) {
			openDetailEditor(false);
		} else if (source == this && TKOutline.CMD_DELETE_SELECTION.equals(command)) {
			handleClearCmd();
		} else {
			super.actionPerformed(event);
		}
	}

	/**
	 * Opens detailed editors for the current selection.
	 * 
	 * @param later Whether to call {@link EventQueue#invokeLater(Runnable)} rather than immediately
	 *            opening the editor.
	 */
	public void openDetailEditor(boolean later) {
		mRowsToEdit = new TKFilteredList<CMRow>(getModel().getSelectionAsList(), CMRow.class);
		if (later) {
			EventQueue.invokeLater(this);
		} else {
			run();
		}
	}

	private ArrayList<CMRow>	mRowsToEdit;

	public void run() {
		if (CSRowEditor.edit(mRowsToEdit)) {
			updateRows(mRowsToEdit);
			updateRowHeights(mRowsToEdit);
			mDataFile.notifySingle(mRowSetChangedID, null);
		}
		mRowsToEdit = null;
	}

	@Override public void undoDidHappen(TKOutlineModel model) {
		super.undoDidHappen(model);
		updateAllRows();
		mDataFile.notifySingle(mRowSetChangedID, null);
	}

	@Override protected void rowsWereDropped() {
		updateAllRows();
		mDataFile.notifySingle(mRowSetChangedID, null);
		requestFocusInWindow();
	}

	private void updateAllRows() {
		updateRows(new TKFilteredList<CMRow>(getModel().getTopLevelRows(), CMRow.class));
	}

	private void updateRows(Collection<CMRow> rows) {
		for (CMRow row : rows) {
			updateRow(row);
		}
	}

	private void updateRow(CMRow row) {
		if (row != null) {
			int count = row.getChildCount();

			for (int i = 0; i < count; i++) {
				CMRow child = (CMRow) row.getChild(i);

				updateRow(child);
			}
			row.update();
		}
	}

	@Override protected int dragEnterRow(DropTargetDragEvent dtde) {
		addDragHighlight(this);
		return super.dragEnterRow(dtde);
	}

	@Override protected void dragEnterRow(DropTargetDragEvent dtde, TKProxyOutline proxy) {
		addDragHighlight(proxy);
		super.dragEnterRow(dtde, proxy);
	}

	@Override protected void dragExitRow(DropTargetEvent dte) {
		removeDragHighlight(this);
		super.dragExitRow(dte);
	}

	@Override protected void dragExitRow(DropTargetEvent dte, TKProxyOutline proxy) {
		removeDragHighlight(proxy);
		super.dragExitRow(dte, proxy);
	}

	@Override protected void dropRow(DropTargetDropEvent dtde) {
		removeDragHighlight(this);
		super.dropRow(dtde);
	}

	@Override protected void dropRow(DropTargetDropEvent dtde, TKProxyOutline proxy) {
		removeDragHighlight(proxy);
		super.dropRow(dtde, proxy);
	}

	private void addDragHighlight(TKOutline outline) {
		mSavedBorder = adjustBorderHighlight(outline, new TKLineBorder(TKColor.HIGHLIGHT, 2, TKLineBorder.ALL_EDGES, false));
	}

	private void removeDragHighlight(TKOutline outline) {
		adjustBorderHighlight(outline, mSavedBorder);
	}

	private TKBorder adjustBorderHighlight(TKOutline outline, TKBorder border) {
		TKPanel panel = outline;
		TKBorder savedBorder;

		if (mDataFile instanceof CMListFile) {
			TKScrollPanel scroller = (TKScrollPanel) outline.getAncestorOfType(TKScrollPanel.class);

			if (scroller != null) {
				panel = scroller.getContentBorderView();
			}
		}

		savedBorder = panel.getBorder();
		panel.setBorder(border, true, false);
		return savedBorder;
	}

	/**
	 * Adds the row to the list. Recursively descends the row's children and does the same with
	 * them.
	 * 
	 * @param list The list to add rows to.
	 * @param row The row to check.
	 */
	protected void addRowsToBeProcessed(ArrayList<CMRow> list, CMRow row) {
		int count = row.getChildCount();

		list.add(row);
		for (int i = 0; i < count; i++) {
			addRowsToBeProcessed(list, (CMRow) row.getChild(i));
		}
	}

	@Override public void sorted(TKOutlineModel model, boolean restoring) {
		super.sorted(model, restoring);
		if (!restoring) {
			mDataFile.setModified(true);
		}
	}

	@Override public void postUndo(TKUndo undo) {
		if (mDataFile != null) {
			mDataFile.addEdit(undo);
		}
	}

	@Override public Color getBackground(int rowIndex, boolean selected, boolean active) {
		if (selected && active) {
			return TKColor.HIGHLIGHT;
		}
		return useBanding() ? rowIndex % 2 == 0 ? TKColor.PRIMARY_BANDING : TKColor.SECONDARY_BANDING : Color.white;
	}
}
