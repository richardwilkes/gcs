/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.Selection;
import com.trollworks.gcs.ui.SelectionOwner;
import com.trollworks.gcs.utility.text.Numbers;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Hashtable;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.StringTokenizer;
import javax.swing.undo.StateEditable;

/** The data model underlying a {@link Outline}. */
public class OutlineModel implements SelectionOwner, StateEditable {
    private static final String                          UNDO_KEY_ROWS        = "Rows";
    private static final String                          UNDO_KEY_SELECTION   = "Selection";
    private static final String                          UNDO_KEY_SORT_CONFIG = "SortConfig";
    /** The current config version. */
    public static final  int                             CONFIG_VERSION       = 4;
    private              ArrayList<OutlineModelListener> mListeners;
    private              ArrayList<Column>               mColumns;
    private              ArrayList<Row>                  mRows;
    private              Selection                       mSelection;
    private              Column                          mDragColumn;
    private              Row                             mDragTargetRow;
    private              Row[]                           mDragRows;
    private              boolean                         mLocked;
    private              boolean                         mNotifyOfSelections;
    private              Row                             mSavedAnchorRow;
    private              List<Row>                       mSavedSelection;
    private              boolean                         mShowIndent;
    private              int                             mIndentWidth;
    private              int                             mHierarchyColumnID   = -1;
    private              RowFilter                       mRowFilter;
    private              Map<String, Object>             mProperties          = new HashMap<>();

    /** Creates a new model. */
    public OutlineModel() {
        mListeners = new ArrayList<>();
        mColumns = new ArrayList<>();
        mRows = new ArrayList<>();
        mSelection = new Selection(this);
        mNotifyOfSelections = true;
    }

    public Object getProperty(String key) {
        return mProperties.get(key);
    }

    public void setProperty(String key, Object data) {
        mProperties.put(key, data);
    }

    /**
     * Adds a listener to this model.
     *
     * @param listener The listener to add.
     */
    public void addListener(OutlineModelListener listener) {
        if (!mListeners.contains(listener)) {
            mListeners.add(listener);
        }
    }

    /**
     * Removes a listener from this model.
     *
     * @param listener The listener to remove.
     */
    public void removeListener(OutlineModelListener listener) {
        mListeners.remove(listener);
    }

    private OutlineModelListener[] getCurrentListeners() {
        return mListeners.toArray(new OutlineModelListener[0]);
    }

    private void notifyOfRowAdditions(Row[] rows) {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.rowsAdded(this, rows);
        }
    }

    private void notifyOfSortCleared() {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.sortCleared(this);
        }
    }

    private void notifyOfLockedStateWillChange() {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.lockedStateWillChange(this);
        }
    }

    private void notifyOfLockedStateDidChange() {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.lockedStateDidChange(this);
        }
    }

    private void notifyOfSort() {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.sorted(this);
        }
    }

    private void notifyOfRowsWillBeRemoved(Row[] rows) {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.rowsWillBeRemoved(this, rows);
        }
    }

    private void notifyOfRowsWereRemoved(Row[] rows) {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.rowsWereRemoved(this, rows);
        }
    }

    /**
     * @param row    The {@link Row} that was modified.
     * @param column The {@link Column} that was modified.
     */
    public void notifyOfRowModification(Row row, Column column) {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.rowWasModified(this, row, column);
        }
    }

    private void notifyOfUndoWillHappen() {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.undoWillHappen(this);
        }
    }

    private void notifyOfUndoDidHappen() {
        for (OutlineModelListener listener : getCurrentListeners()) {
            listener.undoDidHappen(this);
        }
    }

    /**
     * Adds the specified column.
     *
     * @param column The column to add.
     */
    public void addColumn(Column column) {
        mColumns.add(column);
    }

    /**
     * Adds the specified columns.
     *
     * @param columns The columns to add.
     */
    public void addColumns(List<Column> columns) {
        mColumns.addAll(columns);
    }

    /**
     * Removes the specified column.
     *
     * @param column The column to remove.
     */
    public void removeColumn(Column column) {
        mColumns.remove(column);
    }

    public void removeAllColumns() {
        mColumns.clear();
    }

    /** @return The columns contained by the model. */
    public List<Column> getColumns() {
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
    public Column getColumnWithID(int id) {
        int count = getColumnCount();
        for (int i = 0; i < count; i++) {
            Column column = getColumnAtIndex(i);
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
    public Column getColumnAtIndex(int index) {
        return mColumns.get(index);
    }

    /**
     * @param column The column.
     * @return The column index of the specified column.
     */
    public int getIndexOfColumn(Column column) {
        return mColumns.indexOf(column);
    }

    /** @return The number of columns that can be displayed. */
    public int getVisibleColumnCount() {
        int count = 0;
        for (Column column : mColumns) {
            if (column.isVisible()) {
                count++;
            }
        }
        return count;
    }

    /** @return The columns that can have been hidden. */
    public Collection<Column> getHiddenColumns() {
        List<Column> list = new ArrayList<>();
        for (Column column : mColumns) {
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
    public void addRow(Row row) {
        addRow(row, false);
    }

    /**
     * Adds the specified row.
     *
     * @param row             The row to add.
     * @param includeChildren Whether children of open rows are added as well.
     */
    public void addRow(Row row, boolean includeChildren) {
        addRow(mRows.size(), row, includeChildren);
    }

    /**
     * Adds the specified row.
     *
     * @param index The index to add the row at.
     * @param row   The row to add.
     */
    public void addRow(int index, Row row) {
        addRow(index, row, false);
    }

    /**
     * Adds the specified row.
     *
     * @param index           The index to add the row at.
     * @param row             The row to add.
     * @param includeChildren Whether children of open rows are added as well.
     */
    public void addRow(int index, Row row, boolean includeChildren) {
        ArrayList<Row> list = new ArrayList<>();
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
        notifyOfRowAdditions(list.toArray(new Row[0]));
        clearSort();
    }

    private void addChildren(Row row) {
        List<Row> list = collectRowsAndSetOwner(new ArrayList<>(), row, true);
        preserveSelection();
        mRows.addAll(getIndexOfRow(row) + 1, list);
        mSelection.setSize(mRows.size());
        restoreSelection();
        notifyOfRowAdditions(list.toArray(new Row[0]));
    }

    /**
     * Adds the specified row to the passed in list, along with its children if the row is open.
     * Each row added to the list also has its owner set to this outline model.
     *
     * @param list         The list to add it to.
     * @param row          The row to add.
     * @param childrenOnly {@code false} to include the passed in row as well as its children.
     * @return The passed in list.
     */
    public List<Row> collectRowsAndSetOwner(List<Row> list, Row row, boolean childrenOnly) {
        if (!childrenOnly) {
            list.add(row);
            row.setOwner(this);
        }
        if (row.isOpen() && row.hasChildren()) {
            for (Row row2 : row.getChildren()) {
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
    public void removeRow(Row row) {
        removeRows(new Row[]{row});
    }

    /**
     * Removes the specified row.
     *
     * @param index The row index to remove.
     */
    public void removeRow(int index) {
        removeRows(new int[]{index});
    }

    /**
     * Removes the specified rows.
     *
     * @param rows The rows to remove.
     */
    public void removeRows(Row[] rows) {
        Set<Row> set    = new HashSet<>();
        int      i;
        int      length = rows.length;
        for (i = 0; i < length; i++) {
            int index = getIndexOfRow(rows[i]);
            if (index > -1) {
                collectRowAndDescendantsAtIndex(set, index);
            }
        }
        int[] indexes = new int[set.size()];
        i = 0;
        for (Row row : set) {
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
        Set<Row> set    = new HashSet<>();
        int      max    = mRows.size();
        int      i;
        int      length = indexes.length;
        for (i = 0; i < length; i++) {
            int index = indexes[i];
            if (index > -1 && index < max) {
                collectRowAndDescendantsAtIndex(set, index);
            }
        }
        int[] rows = new int[set.size()];
        i = 0;
        for (Row row : set) {
            rows[i++] = getIndexOfRow(row);
        }
        removeRowsInternal(rows);
    }

    /**
     * Adds the specified row index to the provided set as well as any descendant rows.
     *
     * @param set   The set to add the rows to.
     * @param index The index to start at.
     */
    public void collectRowAndDescendantsAtIndex(Set<Row> set, int index) {
        Row row = getRowAtIndex(index);
        int max = mRows.size();
        set.add(row);
        while (++index < max) {
            Row next = getRowAtIndex(index);
            if (!next.isDescendantOf(row)) {
                break;
            }
            set.add(next);
        }
    }

    private void removeRowsInternal(int[] indexes) {
        int   length = indexes.length;
        Row[] rows   = new Row[length];
        int   i;

        Arrays.sort(indexes);
        for (i = 0; i < length; i++) {
            rows[i] = getRowAtIndex(indexes[i]);
        }

        preserveSelection();
        notifyOfRowsWillBeRemoved(rows);
        for (i = length - 1; i >= 0; i--) {
            mRows.remove(indexes[i]);
            rows[i].setOwner(null);
        }
        mSelection.setSize(mRows.size());
        restoreSelection();
        notifyOfRowsWereRemoved(rows);
    }

    /** Removes all rows. */
    public void removeAllRows() {
        Row[] rows = mRows.toArray(new Row[0]);

        mSelection.deselect();
        mSelection.setSize(0);
        notifyOfRowsWillBeRemoved(rows);
        mRows.clear();
        for (Row element : rows) {
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
    public List<Row> getRows() {
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
    public Row getRowAtIndex(int index) {
        return mRows.get(index);
    }

    /**
     * @param row The row.
     * @return The row index of the specified row.
     */
    public int getIndexOfRow(Row row) {
        return mRows.indexOf(row);
    }

    /** @return The top-level rows (i.e. those with a {@code null} parent). */
    public List<Row> getTopLevelRows() {
        List<Row> list = new ArrayList<>();
        for (Row row : mRows) {
            if (row.getParent() == null) {
                list.add(row);
            }
        }
        return list;
    }

    /** @return The current selection. */
    public Selection getSelection() {
        return mSelection;
    }

    /** @return The column being dragged. */
    public Column getDragColumn() {
        return mDragColumn;
    }

    /** @param column The column being dragged. */
    protected void setDragColumn(Column column) {
        mDragColumn = column;
    }

    /** @return The rows being dragged. */
    public Row[] getDragRows() {
        return mDragRows;
    }

    /** @param rows The rows being dragged. */
    public void setDragRows(Row[] rows) {
        mDragRows = rows;
    }

    /** Clears the sort criteria on the columns. */
    public void clearSort() {
        if (clearSortInternal()) {
            notifyOfSortCleared();
        }
    }

    private boolean clearSortInternal() {
        int     count  = getColumnCount();
        boolean notify = false;

        for (int i = 0; i < count; i++) {
            Column column = getColumnAtIndex(i);

            if (column.getSortSequence() != -1) {
                column.setSortCriteria(-1, column.isSortAscending());
                notify = true;
            }
        }
        return notify;
    }

    /** Sorts the model, if needed. */
    public void sortIfNeeded() {
        sortInternal();
    }

    /** Sorts the model. */
    public void sort() {
        sortInternal();
    }

    private void sortInternal() {
        preserveSelection();
        RowSorter.sort(mColumns, mRows, true);
        restoreSelection();
        notifyOfSort();
    }

    /**
     * @return A configuration string that can be used to restore the current sort configuration.
     *         Returns {@code null} if there was no current sort applied.
     */
    public String getSortConfig() {
        StringBuilder buffer  = new StringBuilder();
        int           count   = mColumns.size();
        boolean       hasSort = false;

        buffer.append('S');
        buffer.append(CONFIG_VERSION);
        buffer.append('\t');
        buffer.append(count);
        for (int i = 0; i < count; i++) {
            Column column   = getColumnAtIndex(i);
            int    sequence = column.getSortSequence();

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
        if (config != null && config.startsWith("S")) {
            try {
                StringTokenizer tokenizer = new StringTokenizer(config, "\t");
                if (Numbers.extractInteger(tokenizer.nextToken().substring(1), 0, false) == CONFIG_VERSION) {
                    int count = Numbers.extractInteger(tokenizer.nextToken(), 0, false);
                    if (clearSortInternal()) {
                        result = 0;
                    }
                    for (int i = 0; i < count; i++) {
                        Column column = getColumnWithID(Numbers.extractInteger(tokenizer.nextToken(), 0, false));
                        if (column == null) {
                            throw new Exception();
                        }
                        column.setSortCriteria(Numbers.extractInteger(tokenizer.nextToken(), -1, false), Numbers.extractBoolean(tokenizer.nextToken()));
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
        boolean open  = true;

        for (int i = 0; i < mRows.size(); i++) {
            Row row = getRowAtIndex(i);
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
     * @param row  The row being changed.
     * @param open The new open state.
     */
    public void rowOpenStateChanged(Row row, boolean open) {
        if (row.hasChildren() && mRows.contains(row)) {
            if (open) {
                addChildren(row);
            } else {
                removeRows(row.getChildren().toArray(new Row[0]));
            }
        }
    }

    /**
     * @param index The index of the row to check.
     * @return {@code true} if the specified row is selected.
     */
    public boolean isRowSelected(int index) {
        return mSelection.isSelected(index);
    }

    /**
     * @param row The row to check.
     * @return {@code true} if the specified row is selected.
     */
    public boolean isRowSelected(Row row) {
        return mSelection.isSelected(getIndexOfRow(row));
    }

    /**
     * @param index The index of the row to check.
     * @return {@code true} if the specified row is selected, either directly or due to one of its
     *         parents being selected.
     */
    public boolean isExtendedRowSelected(int index) {
        if (index < 0 || index >= mRows.size()) {
            return false;
        }
        return isExtendedRowSelected(getRowAtIndex(index));
    }

    /**
     * @param row The row to check.
     * @return {@code true} if the specified row is selected, either directly or due to one of its
     *         parents being selected.
     */
    public boolean isExtendedRowSelected(Row row) {
        while (row != null) {
            if (isRowSelected(row)) {
                return true;
            }
            row = row.getParent();
        }
        return false;
    }

    /** @return {@code true} if a selection is present. */
    public boolean hasSelection() {
        return !mSelection.isEmpty();
    }

    /** @return {@code true} if one or more rows are not currently selected. */
    public boolean canSelectAll() {
        return mSelection.canSelectAll();
    }

    /** @return The first selected row index. Returns -1 if no row is selected. */
    public int getFirstSelectedRowIndex() {
        return mSelection.firstSelectedIndex();
    }

    /** @return The first selected row. Returns {@code null} if no row is selected. */
    public Row getFirstSelectedRow() {
        int index = getFirstSelectedRowIndex();
        return index == -1 ? null : getRowAtIndex(index);
    }

    /** @return The last selected row index. Returns -1 if no row is selected. */
    public int getLastSelectedRowIndex() {
        return mSelection.lastSelectedIndex();
    }

    /** @return The last selected row. Returns {@code null} if no row is selected. */
    public Row getLastSelectedRow() {
        int index = getLastSelectedRowIndex();
        return index == -1 ? null : getRowAtIndex(index);
    }

    /** @return The current selection. */
    public List<Row> getSelectionAsList() {
        return getSelectionAsList(false);
    }

    /**
     * @param minimal Pass in {@code true} to prevent children of selected nodes from being
     *                included.
     * @return The current selection.
     */
    public List<Row> getSelectionAsList(boolean minimal) {
        List<Row> list  = new ArrayList<>(mSelection.getCount());
        int       index = mSelection.firstSelectedIndex();

        while (index != -1) {
            Row     row = getRowAtIndex(index);
            boolean add = true;

            if (minimal) {
                Row parent = row.getParent();
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
        reapplyRowFilter();
    }

    /**
     * Selects the specified row.
     *
     * @param rowIndex The row to select.
     * @param add      Pass in {@code true} to add to the current selection or {@code false} to
     *                 replace the current selection.
     */
    public void select(int rowIndex, boolean add) {
        mSelection.select(rowIndex, add);
        reapplyRowFilter();
    }

    /**
     * Selects the specified row.
     *
     * @param row The row to select.
     * @param add Pass in {@code true} to add to the current selection or {@code false} to replace
     *            the current selection.
     */
    public void select(Row row, boolean add) {
        mSelection.select(getIndexOfRow(row), add);
        reapplyRowFilter();
    }

    /**
     * Selects the specified rows.
     *
     * @param rows The rows to select.
     * @param add  Pass in {@code true} to add to the current selection or {@code false} to replace
     *             the current selection.
     */
    public void select(Collection<? extends Row> rows, boolean add) {
        Set<Row> set     = new HashSet<>(rows);
        int[]    indexes = new int[set.size()];
        int      i       = 0;

        for (Row row : set) {
            indexes[i++] = getIndexOfRow(row);
        }
        mSelection.select(indexes, add);
        reapplyRowFilter();
    }

    /**
     * Selects the specified rows.
     *
     * @param from The first row in the range to select.
     * @param to   The last row in the range to select.
     * @param add  Pass in {@code true} to add to the current selection or {@code false} to replace
     *             the current selection.
     */
    public void select(int from, int to, boolean add) {
        mSelection.select(from, to, add);
        reapplyRowFilter();
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
    public void deselect(Row row) {
        mSelection.deselect(getIndexOfRow(row));
    }

    /**
     * Deselects the specified rows.
     *
     * @param rows The rows to deselect.
     */
    public void deselect(List<Row> rows) {
        Set<Row> set     = new HashSet<>(rows);
        int[]    indexes = new int[set.size()];
        int      i       = 0;

        for (Row row : set) {
            indexes[i++] = getIndexOfRow(row);
        }
        mSelection.deselect(indexes);
    }

    /**
     * Deselects the specified rows.
     *
     * @param from The first row in the range to deselect.
     * @param to   The last row in the range to deselect.
     */
    public void deselect(int from, int to) {
        mSelection.deselect(from, to);
    }

    private void preserveSelection() {
        int anchor = mSelection.getAnchor();

        mSavedAnchorRow = anchor == -1 ? null : getRowAtIndex(anchor);
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

    @Override
    public void selectionAboutToChange() {
        if (mNotifyOfSelections) {
            OutlineModelListener[] listeners = getCurrentListeners();

            for (OutlineModelListener element : listeners) {
                element.selectionWillChange(this);
            }
        }
    }

    @Override
    public void selectionDidChange() {
        if (mNotifyOfSelections) {
            OutlineModelListener[] listeners = getCurrentListeners();

            for (OutlineModelListener element : listeners) {
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

    /** @return The column that contains the hierarchy controls. */
    public Column getHierarchyColumn() {
        if (mHierarchyColumnID != -1) {
            Column column = getColumnWithID(mHierarchyColumnID);
            if (column != null) {
                return column;
            }
        }
        return getColumnAtIndex(0);
    }

    /** @param column The column that should contain the hierarchy controls. */
    public void setHierarchyColumn(Column column) {
        mHierarchyColumnID = column == null ? -1 : column.getID();
    }

    /**
     * @param column The column to check.
     * @return {@code true} if the specified column is the hierarchy column.
     */
    public boolean isHierarchyColumn(Column column) {
        if (column != null) {
            if (mHierarchyColumnID == -1) {
                for (Column col : mColumns) {
                    if (col == column) {
                        return true;
                    }
                    if (col.isVisible()) {
                        return false;
                    }
                }
            } else {
                return mHierarchyColumnID == column.getID();
            }
        }
        return false;
    }

    /** @return {@code true} if hierarchy indention (and controls) will be shown. */
    public boolean showIndent() {
        return mShowIndent;
    }

    /** @param show {@code true} if hierarchy indention (and controls) will be shown. */
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
     * @param row    The row.
     * @param column The column.
     * @return The amount the column for this row is indented.
     */
    public int getIndentWidth(Row row, Column column) {
        if (mShowIndent && isHierarchyColumn(column)) {
            return getIndentWidth() * (1 + row.getDepth());
        }
        return 0;
    }

    @Override
    public void storeState(Hashtable<Object, Object> state) {
        List<Row> rows = getRows();
        state.put(UNDO_KEY_ROWS, new ArrayList<>(rows));
        state.put(UNDO_KEY_SELECTION, new Selection(getSelection()));
        String sortConfig = getSortConfig();
        if (sortConfig != null) {
            state.put(UNDO_KEY_SORT_CONFIG, sortConfig);
        }
        for (Row row : RowSorter.collectContainerRows(rows, new HashSet<>())) {
            state.put(row, new RowUndoSnapshot(row));
        }
    }

    @Override
    public void restoreState(Hashtable<?, ?> state) {
        notifyOfUndoWillHappen();

        String  origSortConfig = getSortConfig();
        boolean sortCleared    = clearSortInternal();

        @SuppressWarnings("unchecked") ArrayList<Row> rows = (ArrayList<Row>) state.get(UNDO_KEY_ROWS);
        if (rows != null) {
            mRows = new ArrayList<>(rows);
        }
        for (Row row : mRows) {
            row.resetOwner(this);
        }
        for (Map.Entry<?, ?> entry : state.entrySet()) {
            Object key = entry.getKey();
            if (key instanceof Row) {
                ((Row) key).applyUndoSnapshot(this, (RowUndoSnapshot) entry.getValue());
            }
        }

        Selection selection = (Selection) state.get(UNDO_KEY_SELECTION);
        if (selection != null) {
            mSelection = new Selection(selection);
        }

        String sortConfig = (String) state.get(UNDO_KEY_SORT_CONFIG);
        if (sortConfig == null) {
            sortConfig = origSortConfig;
        }
        int result = applySortConfigInternal(sortConfig);
        if (result == 0) {
            sortCleared = true;
        } else if (result == 1) {
            sortCleared = false;
            notifyOfSort();
        }
        if (sortCleared) {
            notifyOfSortCleared();
        }

        notifyOfUndoDidHappen();
    }

    /** @return The {@link RowFilter} being used. */
    public RowFilter getRowFilter() {
        return mRowFilter;
    }

    /** @param filter The {@link RowFilter} to use. */
    public void setRowFilter(RowFilter filter) {
        mRowFilter = filter;
    }

    /**
     * @param row The {@link Row} to check.
     * @return Whether the {@link Row} should be filtered from view.
     */
    public boolean isRowFiltered(Row row) {
        if (mRowFilter != null) {
            return mRowFilter.isRowFiltered(row);
        }
        return false;
    }

    /** Causes the {@link RowFilter} to be re-applied to the selection. */
    public void reapplyRowFilter() {
        if (mRowFilter != null) {
            List<Row> list  = new ArrayList<>(mSelection.getCount());
            int       index = mSelection.firstSelectedIndex();
            while (index != -1) {
                Row row = getRowAtIndex(index);
                if (mRowFilter.isRowFiltered(row)) {
                    list.add(row);
                }
                index = mSelection.nextSelectedIndex(index + 1);
            }
            if (!list.isEmpty()) {
                int anchor = mSelection.getAnchor();
                deselect(list);
                mSelection.setAnchor(anchor);
            }
        }
    }

    public Row getDragTargetRow() {
        return mDragTargetRow;
    }

    public void setDragTargetRow(Row dragTargetRow) {
        mDragTargetRow = dragTargetRow;
    }
}
