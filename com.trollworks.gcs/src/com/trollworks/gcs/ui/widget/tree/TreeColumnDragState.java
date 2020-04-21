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
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.util.ArrayList;

/** Provides temporary storage for dragging {@link TreeColumn}s. */
public class TreeColumnDragState extends TreeDragState {
    private TreeColumn            mColumn;
    private ArrayList<TreeColumn> mOriginal;

    /**
     * Creates a new {@link TreeColumnDragState}.
     *
     * @param panel  The {@link TreePanel} to work with.
     * @param column The {@link TreeColumn} from the drag.
     */
    public TreeColumnDragState(TreePanel panel, TreeColumn column) {
        super(panel);
        mColumn = column;
        mOriginal = new ArrayList<>(panel.getColumns());
        setHeaderFocus(true);
        setContentsFocus(true);
    }

    /** @return The {@link TreeColumn} from the drag. */
    public TreeColumn getColumn() {
        return mColumn;
    }

    /** @return The original {@link TreeColumn}s. */
    public ArrayList<TreeColumn> getOriginalColumns() {
        return mOriginal;
    }

    @Override
    public void dragEnter(DropTargetDragEvent event) {
        event.acceptDrag(DnDConstants.ACTION_MOVE);
    }

    @Override
    public void dragOver(DropTargetDragEvent event) {
        TreePanel  panel      = getPanel();
        int        x          = panel.toHeaderView(new Point(event.getLocation())).x;
        TreeColumn overColumn = panel.overColumn(x);
        if (overColumn != null && overColumn != mColumn) {
            ArrayList<TreeColumn> columns = new ArrayList<>(panel.getColumns());
            int                   over    = columns.indexOf(overColumn);
            int                   cur     = columns.indexOf(mColumn);
            int                   midway  = panel.getColumnStart(over) + overColumn.getWidth() / 2;
            if (over < cur && x < midway || over > cur && x > midway) {
                if (cur < over) {
                    for (int i = cur; i < over; i++) {
                        columns.set(i, columns.get(i + 1));
                    }
                } else {
                    for (int j = cur; j > over; j--) {
                        columns.set(j, columns.get(j - 1));
                    }
                }
                columns.set(over, mColumn);
                panel.restoreColumns(columns);
            }
        }
        event.acceptDrag(DnDConstants.ACTION_MOVE);
    }

    @Override
    public void dragExit(DropTargetEvent event) {
        TreePanel panel = getPanel();
        if (panel.getColumns().equals(mOriginal)) {
            panel.repaintColumn(mColumn);
        } else {
            panel.restoreColumns(mOriginal);
        }
    }

    @Override
    public void dropActionChanged(DropTargetDragEvent event) {
        event.acceptDrag(DnDConstants.ACTION_MOVE);
    }

    @Override
    public boolean drop(DropTargetDropEvent event) {
        getPanel().repaintColumn(mColumn);
        return true;
    }
}
