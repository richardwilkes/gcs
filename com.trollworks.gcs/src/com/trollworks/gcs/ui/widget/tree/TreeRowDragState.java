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

package com.trollworks.gcs.ui.widget.tree;

import java.awt.Point;
import java.awt.Rectangle;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/** Provides temporary storage for dragging {@link TreeRow}s. */
public class TreeRowDragState extends TreeDragState {
    private TreeRowSelection mRowSelection;
    private TreeContainerRow mParentRow;
    private int              mChildInsertIndex;

    /**
     * Creates a new {@link TreeRowDragState}.
     *
     * @param panel        The {@link TreePanel} to work with.
     * @param rowSelection The {@link TreeRowSelection} from the drag.
     */
    public TreeRowDragState(TreePanel panel, TreeRowSelection rowSelection) {
        super(panel);
        mRowSelection = rowSelection;
        setContentsFocus(true);
    }

    /** @return The parent {@link TreeContainerRow} for the insertion. */
    public TreeContainerRow getParentRow() {
        return mParentRow;
    }

    /** @param parentRow The parent {@link TreeContainerRow} for the insertion. */
    public void setParentRow(TreeContainerRow parentRow) {
        mParentRow = parentRow;
    }

    /** @return The child insertion index. */
    public int getChildInsertIndex() {
        return mChildInsertIndex;
    }

    /** @param childInsertIndex The child insertion index. */
    public void setChildInsertIndex(int childInsertIndex) {
        mChildInsertIndex = childInsertIndex;
    }

    @Override
    public void dragEnter(DropTargetDragEvent event) {
        acceptDrag(event);
    }

    private boolean acceptDrag(DropTargetDragEvent event) {
        int allowed = getPanel().getAllowedRowDropTypes();
        if ((allowed & event.getDropAction()) == 0) {
            if ((allowed & DnDConstants.ACTION_MOVE) != 0) {
                event.acceptDrag(DnDConstants.ACTION_MOVE);
            } else if ((allowed & DnDConstants.ACTION_COPY) != 0) {
                event.acceptDrag(DnDConstants.ACTION_COPY);
            } else {
                event.acceptDrag(DnDConstants.ACTION_NONE);
                return false;
            }
        }
        return true;
    }

    @Override
    public void dragOver(DropTargetDragEvent event) {
        TreePanel        panel       = getPanel();
        TreeContainerRow parentRow   = null;
        int              insertIndex = -1;
        if (acceptDrag(event)) {
            Point    pt      = panel.toContentView(new Point(event.getLocation()));
            TreeRoot root    = panel.getRoot();
            TreeRow  overRow = panel.overRow(pt.y);
            if (overRow != null) {
                if (mRowSelection.getRows().get(0).getTreeRoot() != root || !panel.isRowOrAncestorSelected(overRow)) {
                    int indent = TreePanel.INDENT * overRow.getDepth();
                    if (panel.areDisclosureControlsShowing()) {
                        indent += TreePanel.INDENT;
                    }
                    if (overRow instanceof TreeContainerRow && pt.x > indent) {
                        parentRow = (TreeContainerRow) overRow;
                        insertIndex = 0;
                    } else {
                        parentRow = overRow.getParent();
                        insertIndex = overRow.getIndex();
                        Rectangle bounds = panel.getRowBounds(overRow);
                        if (pt.y > bounds.y + bounds.height / 2) {
                            insertIndex++;
                        }
                    }
                }
            } else {
                parentRow = root;
                insertIndex = root.getChildCount();
            }
        }
        panel.adjustInsertionMarker(parentRow, insertIndex);
    }

    @Override
    public void dragExit(DropTargetEvent event) {
        getPanel().adjustInsertionMarker(null, -1);
    }

    @Override
    public void dropActionChanged(DropTargetDragEvent event) {
        acceptDrag(event);
    }

    @Override
    public boolean drop(DropTargetDropEvent event) {
        TreeContainerRow parentRow = mParentRow;
        if (parentRow != null) {
            int       childInsertIndex = mChildInsertIndex;
            TreePanel panel            = getPanel();
            TreeRow   indexRow         = null;
            if (childInsertIndex < parentRow.getChildCount()) {
                indexRow = parentRow.getChild(childInsertIndex);
            }
            panel.adjustInsertionMarker(null, -1);
            List<TreeRow>         rows       = mRowSelection.getRows();
            Set<TreeContainerRow> openRows   = mRowSelection.getOpenRows();
            boolean               fromSelf   = rows.get(0).getTreeRoot() == panel.getRoot();
            int                   dropAction = event.getDropAction() & panel.getAllowedRowDropTypes();
            if (dropAction == DnDConstants.ACTION_COPY || !fromSelf && dropAction == DnDConstants.ACTION_MOVE) {
                List<TreeRow> copiedRows = new ArrayList<>(rows.size());
                for (TreeRow row : rows) {
                    copiedRows.add(row.clone());
                }
                List<TreeContainerRow> originalContainers = getContainers(rows);
                List<TreeContainerRow> copiedContainers   = getContainers(copiedRows);
                rows = copiedRows;
                Set<TreeContainerRow> open  = new HashSet<>();
                int                   count = originalContainers.size();
                for (int i = 0; i < count; i++) {
                    if (openRows.contains(originalContainers.get(i))) {
                        open.add(copiedContainers.get(i));
                    }
                }
                openRows = open;
            } else if (dropAction == DnDConstants.ACTION_MOVE) {
                for (TreeRow row : rows) {
                    TreeContainerRow parent = row.getParent();
                    if (row == indexRow) {
                        int index = indexRow.getIndex();
                        parent.removeRow(row);
                        indexRow = index < parent.getChildCount() ? parent.getChild(index) : null;
                    } else {
                        parent.removeRow(row);
                    }
                }
            } else {
                return false;
            }
            childInsertIndex = indexRow != null ? indexRow.getIndex() : parentRow.getChildCount();
            parentRow.addRow(childInsertIndex, rows);
            panel.select(rows.get(0), rows.get(rows.size() - 1), false);
            panel.setOpen(true, openRows);
            panel.getTreeSorter().clearSort();
            panel.pack();
            panel.notify(TreeNotificationKeys.ROW_DROP, rows);
            return true;
        }
        return false;
    }

    private static List<TreeContainerRow> getContainers(List<TreeRow> rows) {
        List<TreeContainerRow> containers = new ArrayList<>();
        for (TreeRow row : new TreeRowIterator(rows)) {
            if (row instanceof TreeContainerRow) {
                containers.add((TreeContainerRow) row);
            }
        }
        return containers;
    }
}
