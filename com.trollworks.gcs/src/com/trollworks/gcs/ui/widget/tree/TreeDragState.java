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

import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;

/** Temporary storage needed for dragging. */
public abstract class TreeDragState {
    private TreePanel mPanel;
    private boolean   mHeaderFocus;
    private boolean   mContentsFocus;

    /**
     * Creates a new {@link TreeDragState}.
     *
     * @param panel The {@link TreePanel} to work with.
     */
    protected TreeDragState(TreePanel panel) {
        mPanel = panel;
    }

    /** @return The {@link TreePanel} to work with. */
    public TreePanel getPanel() {
        return mPanel;
    }

    /** @return Whether or not the header is the focus of the drag. */
    public final boolean isHeaderFocus() {
        return mHeaderFocus;
    }

    /** @param focus Whether or not the header is the focus of the drag. */
    public final void setHeaderFocus(boolean focus) {
        mHeaderFocus = focus;
    }

    /** @return Whether or not the contents are the focus of the drag. */
    public final boolean isContentsFocus() {
        return mContentsFocus;
    }

    /** @param focus Whether or not the contents are the focus of the drag. */
    public final void setContentsFocus(boolean focus) {
        mContentsFocus = focus;
    }

    /**
     * Called when a drag enters the {@link TreePanel}.
     *
     * @param event The {@link DropTargetDragEvent}.
     */
    public abstract void dragEnter(DropTargetDragEvent event);

    /**
     * Called when a drag is over the {@link TreePanel}.
     *
     * @param event The {@link DropTargetDragEvent}.
     */
    public abstract void dragOver(DropTargetDragEvent event);

    /**
     * Called when a drag leaves the {@link TreePanel}.
     *
     * @param event The {@link DropTargetEvent}.
     */
    public abstract void dragExit(DropTargetEvent event);

    /**
     * Called when the drop action changes.
     *
     * @param event The {@link DropTargetDragEvent}.
     */
    public abstract void dropActionChanged(DropTargetDragEvent event);

    /**
     * Called when a drop occurs on the {@link TreePanel}.
     *
     * @param event The {@link DropTargetDropEvent}.
     * @return Whether or not the drop was successful.
     */
    public abstract boolean drop(DropTargetDropEvent event);
}
