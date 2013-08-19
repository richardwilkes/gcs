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

package com.trollworks.toolkit.widget.outline;

import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKSelection;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.StringTokenizer;
import java.util.Map.Entry;

/** The data model underlying a {@link TKOutline}. */
public class TKOutlineModel implements TKSelection.Owner {
	/** The current config version. */
	public static final int						CONFIG_VERSION	= 4;
	private ArrayList<TKOutlineModelListener>	mListeners;
	private ArrayList<TKColumn>					mColumns;
	private ArrayList<TKRow>					mRows;
	private TKSelection							mSelection;
	private TKColumn							mDragColumn;
	private TKRow[]								mDragRows;
	private boolean								mLocked;
	private boolean								mNotifyOfSelections;
	private TKRow								mSavedAnchorRow;
	private List<TKRow>							mSavedSelection;
	private boolean								mShowIndent;
	private int									mIndentWidth;

	/** Creates a new model. */
	public TKOutlineModel() {
		mListeners = new ArrayList<TKOutlineModelListener>();
		mColumns = new ArrayList<TKColumn>();
		mRows = new ArrayList<TKRow>();
		mSelection = new TKSelection(this);
		mNotifyOfSelections = true;
	}

	/**
	 * Adds a listener to this model.
	 * 
	 * @param listener The listener to add.
	 */
	public void addListener(TKOutlineModelListener listener) {
		if (!mListeners.contains(listener)) {
			mListeners.add(listener);
		}
	}

	/**
	 * Removes a listener from this model.
	 * 
	 * @param listener The listener to remove.
	 */
	public void removeListener(TKOutlineModelListener listener) {
		mListeners.remove(listener);
	}

	private TKOutlineModelListener[] getCurrentListeners() {
		return mListeners.toArray(new TKOutlineModelListener[0]);
	}

	private void notifyOfRowAdditions(TKRow[] rows) {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.rowsAdded(this, rows);
		}
	}

	private void notifyOfSortCleared() {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.sortCleared(this);
		}
	}

	private void notifyOfLockedStateWillChange() {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.lockedStateWillChange(this);
		}
	}

	private void notifyOfLockedStateDidChange() {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.lockedStateDidChange(this);
		}
	}

	private void notifyOfSort(boolean restoring) {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.sorted(this, restoring);
		}
	}

	private void notifyOfRowsWillBeRemoved(TKRow[] rows) {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.rowsWillBeRemoved(this, rows);
		}
	}

	private void notifyOfRowsWereRemoved(TKRow[] rows) {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.rowsWereRemoved(this, rows);
		}
	}

	private void notifyOfUndoWillHappen() {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.undoWillHappen(this);
		}
	}

	private void notifyOfUndoDidHappen() {
		TKOutlineModelListener[] listeners = getCurrentListeners();

		for (TKOutlineModelListener element : listeners) {
			element.undoDidHappen(this);
		}
	}

	/**
	 * Adds the specified column.
	 * 
	 * @param column The column to add.
	 */
	public void addColumn(TKColumn column) {
		assert !mColumns.contains(column);
		mColumns.add(column);
	}

	/**
	 * Adds the specified columns.
	 * 
	 * @param columns The columns to add.
	 */
	public void addColumns(List<TKColumn> columns) {
		mColumns.addAll(columns);
	}

	/**
	 * Removes the specified column.
	 * 
	 * @param column The column to remove.
	 */
	public void removeColumn(TKColumn column) {
		assert mColumns.contains(column);
		mColumns.remove(column);
	}

	/** @return The columns contained by the model. */
	public List<TKColumn> getColumns() {
		return mColumns;
	}

	/** @return The total number of columns present in the model. */
	public int getColumnCount() {
		return mColumns.size();
	}

	/**
	 * @param id The column ID.
	 * @return The column with the specified user-supplied ID.
	 */
	public TKColumn getColumnWithID(int id) {
		int count = getColumnCount();

		for (int i = 0; i < count; i++) {
			TKColumn column = getColumnAtIndex(i);

			if (column.getID() == id) {
				return column;
			}
		}
		return null;
	}

	/**
	 * @param index The index of the column.
	 * @return The column at the specified index.
	 */
	public TKColumn getColumnAtIndex(int index) {
		return mColumns.get(index);
	}

	/**
	 * @param column The column.
	 * @return The column index of the specified column.
	 */
	public int getIndexOfColumn(TKColumn column) {
		return mColumns.indexOf(column);
	}

	/** @return The number of columns that can be displayed. */
	public int getVisibleColumnCount() {
		int count = 0;

		for (TKColumn column : mColumns) {
			if (column.isVisible()) {
				count++;
			}
		}
		return count;
	}

	/** @return The columns that can have been hidden. */
	public Collection<TKColumn> getHiddenColumns() {
		ArrayList<TKColumn> list = new ArrayList<TKColumn>();

		for (TKColumn column : mColumns) {
			if (!column.isVisible()) {
				list.add(column);
			}
		}
		return list;
	}

	/**
	 * Adds the specified row.
	 * 
	 * @param row The row to add.
	 */
	public void addRow(TKRow row) {
		addRow(row, false);
	}

	/**
	 * Adds the specified row.
	 * 
	 * @param row The row to add.
	 * @param includeChildren Whether children of open rows are added as well.
	 */
	public void addRow(TKRow row, boolean includeChildren) {
		addRow(mRows.size(), row, includeChildren);
	}

	/**
	 * Adds the specified row.
	 * 
	 * @param index The index to add the row at.
	 * @param row The row to add.
	 */
	public void addRow(int index, TKRow row) {
		addRow(index, row, false);
	}

	/**
	 * Adds the specified row.
	 * 
	 * @param index The index to add the row at.
	 * @param row The row to add.
	 * @param includeChildren Whether children of open rows are added as well.
	 */
	public void addRow(int index, TKRow row, boolean includeChildren) {
		ArrayList<TKRow> list = new ArrayList<TKRow>();

		assert !mRows.contains(row);

		if (includeChildren) {
			collectRowsAndSetOwner(list, row, false);
		} else {
			list.add(row);
			row.setOwner(this);
		}
		preserveSelection();
		mRows.addAll(index, list);
		mSelection.setSize(mRows.size());
		restoreSelection();
		notifyOfRowAdditions(list.toArray(new TKRow[0]));
		clearSort();
	}

	private void addChildren(TKRow row) {
		List<TKRow> list = collectRowsAndSetOwner(new ArrayList<TKRow>(), row, true);

		preserveSelection();
		mRows.addAll(getIndexOfRow(row) + 1, list);
		mSelection.setSize(mRows.size());
		restoreSelection();
		notifyOfRowAdditions(list.toArray(new TKRow[0]));
	}

	/**
	 * Adds the specified row to the passed in list, along with its children if the row is open.
	 * Each row added to the list also has its owner set to this outline model.
	 * 
	 * @param list The list to add it to.
	 * @param row The row to add.
	 * @param childrenOnly <code>false</code> to include the passed in row as well as its
	 *            children.
	 * @return The passed in list.
	 */
	public List<TKRow> collectRowsAndSetOwner(List<TKRow> list, TKRow row, boolean childrenOnly) {
		if (!childrenOnly) {
			list.add(row);
			row.setOwner(this);
		}
		if (row.isOpen() && row.hasChildren()) {
			for (TKRow row2 : row.getChildren()) {
				collectRowsAndSetOwner(list, row2, false);
			}
		}
		return list;
	}

	/**
	 * Removes the specified row.
	 * 
	 * @param row The row to remove.
	 */
	public void removeRow(TKRow row) {
		removeRows(new TKRow[] { row });
	}

	/**
	 * Removes the specified row.
	 * 
	 * @param index The row index to remove.
	 */
	public void removeRow(int index) {
		removeRows(new int[] { index });
	}

	/**
	 * Removes the specified rows.
	 * 
	 * @param rows The rows to remove.
	 */
	public void removeRows(TKRow[] rows) {
		HashSet<TKRow> set = new HashSet<TKRow>();
		int i;

		for (i = 0; i < rows.length; i++) {
			int index = getIndexOfRow(rows[i]);

			if (index > -1) {
				collectRows(set, index);
			}
		}

		int[] indexes = new int[set.size()];
		i = 0;
		for (TKRow row : set) {
			indexes[i++] = getIndexOfRow(row);
		}
		removeRowsInternal(indexes);
	}

	/**
	 * Removes the specified rows.
	 * 
	 * @param indexes The row indexes to remove.
	 */
	public void removeRows(int[] indexes) {
		HashSet<TKRow> set = new HashSet<TKRow>();
		int max = mRows.size();
		int i;

		for (i = 0; i < indexes.length; i++) {
			int index = indexes[i];

			if (index > -1 && index < max) {
				collectRows(set, index);
			}
		}

		int[] rows = new int[set.size()];
		i = 0;
		for (TKRow row : set) {
			rows[i++] = getIndexOfRow(row);
		}
		removeRowsInternal(rows);
	}

	private int collectRows(HashSet<TKRow> set, int index) {
		TKRow row = getRowAtIndex(index);
		int max = mRows.size();

		set.add(row);
		index++;
		while (index < max) {
			TKRow next = getRowAtIndex(index);

			if (next.isDescendentOf(row)) {
				set.add(next);
				index++;
			} else {
				break;
			}
		}
		return index;
	}

	private void removeRowsInternal(int[] indexes) {
		TKRow[] rows = new TKRow[indexes.length];
		int i;

		Arrays.sort(indexes);
		for (i = 0; i < indexes.length; i++) {
			rows[i] = getRowAtIndex(indexes[i]);
		}

		preserveSelection();
		notifyOfRowsWillBeRemoved(rows);
		for (i = indexes.length - 1; i >= 0; i--) {
			mRows.remove(indexes[i]);
			rows[i].setOwner(null);
		}
		mSelection.setSize(mRows.size());
		restoreSelection();
		notifyOfRowsWereRemoved(rows);
	}

	/** Removes all rows. */
	public void removeAllRows() {
		TKRow[] rows = mRows.toArray(new TKRow[0]);

		mSelection.deselect();
		mSelection.setSize(0);
		notifyOfRowsWillBeRemoved(rows);
		mRows.clear();
		for (TKRow element : rows) {
			element.setOwner(null);
		}
		notifyOfRowsWereRemoved(rows);
	}

	/** Removes the selection from the model. */
	public void removeSelection() {
		int[] indexes = mSelection.getSelectedIndexes();

		mSelection.deselect();
		removeRows(indexes);
	}

	/** @return The rows contained by the model. */
	public List<TKRow> getRows() {
		return mRows;
	}

	/** @return The total number of rows present in the outline. */
	public int getRowCount() {
		return mRows.size();
	}

	/**
	 * @param index The index of the row.
	 * @return The row at the specified index.
	 */
	public TKRow getRowAtIndex(int index) {
		return mRows.get(index);
	}

	/**
	 * @param row The row.
	 * @return The row index of the specified row.
	 */
	public int getIndexOfRow(TKRow row) {
		return mRows.indexOf(row);
	}

	/** @return The top-level rows (i.e. those with a <code>null</code> parent). */
	public List<TKRow> getTopLevelRows() {
		ArrayList<TKRow> list = new ArrayList<TKRow>();

		for (TKRow row : mRows) {
			if (row.getParent() == null) {
				list.add(row);
			}
		}
		return list;
	}

	/** @return The current selection. */
	public TKSelection getSelection() {
		return mSelection;
	}

	/** @return The column being dragged. */
	public TKColumn getDragColumn() {
		return mDragColumn;
	}

	/** @param column The column being dragged. */
	protected void setDragColumn(TKColumn column) {
		mDragColumn = column;
	}

	/** @return The rows being dragged. */
	public TKRow[] getDragRows() {
		return mDragRows;
	}

	/** @param rows The rows being dragged. */
	public void setDragRows(TKRow[] rows) {
		mDragRows = rows;
	}

	/** Clears the sort criteria on the columns. */
	public void clearSort() {
		if (clearSortInternal()) {
			notifyOfSortCleared();
		}
	}

	private boolean clearSortInternal() {
		int count = getColumnCount();
		boolean notify = false;

		for (int i = 0; i < count; i++) {
			TKColumn column = getColumnAtIndex(i);

			if (column.getSortSequence() != -1) {
				column.setSortCriteria(-1, column.isSortAscending());
				notify = true;
			}
		}
		return notify;
	}

	/** Sorts the model, if needed. */
	public void sortIfNeeded() {
		sortInternal(false);
	}

	/** Sorts the model. */
	public void sort() {
		sortInternal(false);
	}

	private void sortInternal(boolean restoring) {
		preserveSelection();
		TKRowSorter.sort(mColumns, mRows, true);
		restoreSelection();
		notifyOfSort(restoring);
	}

	/**
	 * @return A configuration string that can be used to restore the current sort configuration.
	 *         Returns <code>null</code> if there was no current sort applied.
	 */
	public String getSortConfig() {
		StringBuilder buffer = new StringBuilder();
		int count = mColumns.size();
		boolean hasSort = false;

		buffer.append('S');
		buffer.append(CONFIG_VERSION);
		buffer.append('\t');
		buffer.append(count);
		for (int i = 0; i < count; i++) {
			TKColumn column = getColumnAtIndex(i);
			int sequence = column.getSortSequence();

			buffer.append('\t');
			buffer.append(column.getID());
			buffer.append('\t');
			buffer.append(sequence);
			buffer.append('\t');
			buffer.append(column.isSortAscending());
			if (sequence > -1) {
				hasSort = true;
			}
		}

		return hasSort ? buffer.toString() : null;
	}

	/**
	 * Attempts to restore the specified sort configuration.
	 * 
	 * @param config The configuration to restore.
	 */
	public void applySortConfig(String config) {
		int result = applySortConfigInternal(config);

		if (result == 0) {
			notifyOfSortCleared();
		} else if (result == 1) {
			sort();
		}
	}

	private int applySortConfigInternal(String config) {
		int result = -1;

		if (config != null && config.startsWith("S")) { //$NON-NLS-1$
			try {
				StringTokenizer tokenizer = new StringTokenizer(config, "\t"); //$NON-NLS-1$

				if (TKNumberUtils.getInteger(tokenizer.nextToken().substring(1), 0) == CONFIG_VERSION) {
					int count = TKNumberUtils.getInteger(tokenizer.nextToken(), 0);

					if (clearSortInternal()) {
						result = 0;
					}
					for (int i = 0; i < count; i++) {
						TKColumn column = getColumnWithID(TKNumberUtils.getInteger(tokenizer.nextToken(), 0));

						if (column == null) {
							throw new Exception();
						}
						column.setSortCriteria(TKNumberUtils.getInteger(tokenizer.nextToken(), -1), TKNumberUtils.getBoolean(tokenizer.nextToken()));
						if (column.getSortSequence() != -1) {
							result = 1;
						}
					}
				}
			} catch (Exception exception) {
				// It's a bad config... ignore it.
			}
		}
		return result;
	}

	/**
	 * If the first row capable of having children is open, then all rows will be closed, otherwise
	 * all rows will be opened.
	 */
	public void toggleRowOpenState() {
		boolean first = true;
		boolean open = true;

		for (int i = 0; i < mRows.size(); i++) {
			TKRow row = getRowAtIndex(i);

			if (row.canHaveChildren()) {
				if (first) {
					open = !row.isOpen();
					first = false;
				}
				row.setOpen(open);
			}
		}
	}

	/**
	 * Called when a row's open state changes.
	 * 
	 * @param row The row being changed.
	 * @param open The new open state.
	 */
	public void rowOpenStateChanged(TKRow row, boolean open) {
		if (row.hasChildren() && mRows.contains(row)) {
			if (open) {
				addChildren(row);
			} else {
				removeRows(row.getChildren().toArray(new TKRow[0]));
			}
		}
	}

	/**
	 * @param index The index of the row to check.
	 * @return <code>true</code> if the specified row is selected.
	 */
	public boolean isRowSelected(int index) {
		return mSelection.isSelected(index);
	}

	/**
	 * @param row The row to check.
	 * @return <code>true</code> if the specified row is selected.
	 */
	public boolean isRowSelected(TKRow row) {
		return mSelection.isSelected(getIndexOfRow(row));
	}

	/**
	 * @param index The index of the row to check.
	 * @return <code>true</code> if the specified row is selected, either directly or due to one
	 *         of its parents being selected.
	 */
	public boolean isExtendedRowSelected(int index) {
		if (index < 0 || index >= mRows.size()) {
			return false;
		}
		return isExtendedRowSelected(getRowAtIndex(index));
	}

	/**
	 * @param row The row to check.
	 * @return <code>true</code> if the specified row is selected, either directly or due to one
	 *         of its parents being selected.
	 */
	public boolean isExtendedRowSelected(TKRow row) {
		while (row != null) {
			if (isRowSelected(row)) {
				return true;
			}
			row = row.getParent();
		}
		return false;
	}

	/** @return <code>true</code> if a selection is present. */
	public boolean hasSelection() {
		return !mSelection.isEmpty();
	}

	/** @return <code>true</code> if one or more rows are not currently selected. */
	public boolean canSelectAll() {
		return mSelection.canSelectAll();
	}

	/** @return The first selected row index. Returns -1 if no row is selected. */
	public int getFirstSelectedRowIndex() {
		return mSelection.firstSelectedIndex();
	}

	/** @return The first selected row. Returns <code>null</code> if no row is selected. */
	public TKRow getFirstSelectedRow() {
		int index = getFirstSelectedRowIndex();

		return index == -1 ? null : getRowAtIndex(index);
	}

	/** @return The last selected row index. Returns -1 if no row is selected. */
	public int getLastSelectedRowIndex() {
		return mSelection.lastSelectedIndex();
	}

	/** @return The last selected row. Returns <code>null</code> if no row is selected. */
	public TKRow getLastSelectedRow() {
		int index = getLastSelectedRowIndex();

		return index == -1 ? null : getRowAtIndex(index);
	}

	/** @return The current selection. */
	public List<TKRow> getSelectionAsList() {
		return getSelectionAsList(false);
	}

	/**
	 * @param minimal Pass in <code>true</code> to prevent children of selected nodes from being
	 *            included.
	 * @return The current selection.
	 */
	public List<TKRow> getSelectionAsList(boolean minimal) {
		ArrayList<TKRow> list = new ArrayList<TKRow>(mSelection.getCount());
		int index = mSelection.firstSelectedIndex();

		while (index != -1) {
			TKRow row = getRowAtIndex(index);
			boolean add = true;

			if (minimal) {
				TKRow parent = row.getParent();

				while (parent != null) {
					if (mSelection.isSelected(getIndexOfRow(parent))) {
						add = false;
						break;
					}
					parent = parent.getParent();
				}
			}

			if (add) {
				list.add(row);
			}
			index = mSelection.nextSelectedIndex(index + 1);
		}
		return list;
	}

	/** @return The selection count. */
	public int getSelectionCount() {
		return mSelection.getCount();
	}

	/** Selects all rows in the outline. */
	public void select() {
		mSelection.select();
	}

	/**
	 * Selects the specified row.
	 * 
	 * @param rowIndex The row to select.
	 * @param add Pass in <code>true</code> to add to the current selection or <code>false</code>
	 *            to replace the current selection.
	 */
	public void select(int rowIndex, boolean add) {
		mSelection.select(rowIndex, add);
	}

	/**
	 * Selects the specified row.
	 * 
	 * @param row The row to select.
	 * @param add Pass in <code>true</code> to add to the current selection or <code>false</code>
	 *            to replace the current selection.
	 */
	public void select(TKRow row, boolean add) {
		mSelection.select(getIndexOfRow(row), add);
	}

	/**
	 * Selects the specified rows.
	 * 
	 * @param rows The rows to select.
	 * @param add Pass in <code>true</code> to add to the current selection or <code>false</code>
	 *            to replace the current selection.
	 */
	public void select(Collection<? extends TKRow> rows, boolean add) {
		HashSet<TKRow> set = new HashSet<TKRow>(rows);
		int[] indexes = new int[set.size()];
		int i = 0;

		for (TKRow row : set) {
			indexes[i++] = getIndexOfRow(row);
		}
		mSelection.select(indexes, add);
	}

	/**
	 * Selects the specified rows.
	 * 
	 * @param from The first row in the range to select.
	 * @param to The last row in the range to select.
	 * @param add Pass in <code>true</code> to add to the current selection or <code>false</code>
	 *            to replace the current selection.
	 */
	public void select(int from, int to, boolean add) {
		mSelection.select(from, to, add);
	}

	/** Deselects all rows in the outline. */
	public void deselect() {
		mSelection.deselect();
	}

	/**
	 * Deselects the specified row.
	 * 
	 * @param rowIndex The row index to deselect.
	 */
	public void deselect(int rowIndex) {
		mSelection.deselect(rowIndex);
	}

	/**
	 * Deselects the specified row.
	 * 
	 * @param row The row to deselect.
	 */
	public void deselect(TKRow row) {
		mSelection.deselect(getIndexOfRow(row));
	}

	/**
	 * Deselects the specified rows.
	 * 
	 * @param rows The rows to deselect.
	 */
	public void deselect(List<TKRow> rows) {
		HashSet<TKRow> set = new HashSet<TKRow>(rows);
		int[] indexes = new int[set.size()];
		int i = 0;

		for (TKRow row : set) {
			indexes[i++] = getIndexOfRow(row);
		}
		mSelection.deselect(indexes);
	}

	/**
	 * Deselects the specified rows.
	 * 
	 * @param from The first row in the range to deselect.
	 * @param to The last row in the range to deselect.
	 */
	public void deselect(int from, int to) {
		mSelection.deselect(from, to);
	}

	private void preserveSelection() {
		int anchor = mSelection.getAnchor();

		mSavedAnchorRow = anchor != -1 ? getRowAtIndex(anchor) : null;
		mSavedSelection = getSelectionAsList();
		mNotifyOfSelections = false;
		deselect();
	}

	private void restoreSelection() {
		select(mSavedSelection, false);
		if (mSavedAnchorRow != null) {
			mSelection.setAnchor(getIndexOfRow(mSavedAnchorRow));
		}
		mSavedAnchorRow = null;
		mSavedSelection = null;
		mNotifyOfSelections = true;
	}

	public void selectionAboutToChange() {
		if (mNotifyOfSelections) {
			TKOutlineModelListener[] listeners = getCurrentListeners();

			for (TKOutlineModelListener element : listeners) {
				element.selectionWillChange(this);
			}
		}
	}

	public void selectionDidChange() {
		if (mNotifyOfSelections) {
			TKOutlineModelListener[] listeners = getCurrentListeners();

			for (TKOutlineModelListener element : listeners) {
				element.selectionDidChange(this);
			}
		}
	}

	/** @return Whether the model is "locked" or not. */
	public boolean isLocked() {
		return mLocked;
	}

	/** @param locked Whether the model is "locked" or not. */
	public void setLocked(boolean locked) {
		if (mLocked != locked) {
			notifyOfLockedStateWillChange();
			mLocked = locked;
			notifyOfLockedStateDidChange();
		}
	}

	/**
	 * @param column The column to check.
	 * @return <code>true</code> if the specified column is the first visible column.
	 */
	public boolean isFirstColumn(TKColumn column) {
		if (column != null) {
			for (TKColumn col : mColumns) {
				if (col == column) {
					return true;
				}
				if (col.isVisible()) {
					return false;
				}
			}
		}
		return false;
	}

	/** @return <code>true</code> if hierarchy indention (and controls) will be shown. */
	public boolean showIndent() {
		return mShowIndent;
	}

	/** @param show <code>true</code> if hierarchy indention (and controls) will be shown. */
	public void setShowIndent(boolean show) {
		mShowIndent = show;
	}

	/** @return The width used to indent each level of hierarchy. */
	public int getIndentWidth() {
		return mIndentWidth;
	}

	/** @param width The new indent width. */
	public void setIndentWidth(int width) {
		mIndentWidth = width;
	}

	/**
	 * @param row The row.
	 * @param column The column.
	 * @return The amount the column for this row is indented.
	 */
	public int getIndentWidth(TKRow row, TKColumn column) {
		if (mShowIndent && isFirstColumn(column)) {
			return getIndentWidth() * (1 + row.getDepth());
		}
		return 0;
	}

	/** @param snapshot The snapshot to apply. */
	public void applyUndoSnapshot(TKOutlineModelUndoSnapshot snapshot) {
		boolean sortCleared;
		int result;

		notifyOfUndoWillHappen();

		sortCleared = clearSortInternal();
		mRows = new ArrayList<TKRow>(snapshot.getRows());
		for (TKRow row : mRows) {
			row.applyUndoSnapshot(this);
		}
		for (Entry<TKRow, TKRowUndoSnapshot> entry : snapshot.getMap().entrySet()) {
			entry.getKey().applyUndoSnapshot(this, entry.getValue());
		}

		mSelection = new TKSelection(snapshot.getSelection());
		result = applySortConfigInternal(snapshot.getSortConfig());
		if (result == 0) {
			sortCleared = true;
		} else if (result == 1) {
			sortCleared = false;
			notifyOfSort(false);
		}
		if (sortCleared) {
			notifyOfSortCleared();
		}

		notifyOfUndoDidHappen();
	}
}
