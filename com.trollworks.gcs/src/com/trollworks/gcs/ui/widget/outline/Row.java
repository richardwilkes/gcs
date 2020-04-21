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

import com.trollworks.gcs.ui.RetinaIcon;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/** Represents a single row of data within an {@link OutlineModel}. */
public abstract class Row {
    private   OutlineModel   mOwner;
    private   int            mHeight;
    private   boolean        mOpen;
    private   Row            mParent;
    /** The children of this row. */
    protected ArrayList<Row> mChildren;

    /** Create a new outline row. */
    public Row() {
        mHeight = -1;
    }

    @Override
    public final boolean equals(Object obj) {
        return obj == this;
    }

    @Override
    public final int hashCode() {
        return super.hashCode();
    }

    /**
     * @param owner    The owning model.
     * @param snapshot An undo snapshot.
     */
    void applyUndoSnapshot(OutlineModel owner, RowUndoSnapshot snapshot) {
        mOwner = owner;
        mParent = snapshot.getParent();
        mOpen = snapshot.isOpen();
        if (canHaveChildren()) {
            mChildren.clear();
            for (Row child : snapshot.getChildren()) {
                mChildren.add(child);
                child.mParent = this;
            }
        }
    }

    /** @param owner The owning model. */
    void resetOwner(OutlineModel owner) {
        mOwner = owner;
        mParent = null;
    }

    /**
     * @param column The column.
     * @return The icon for the specified column, or {@code null}.
     */
    @SuppressWarnings("static-method")
    public RetinaIcon getIcon(Column column) {
        return null;
    }

    /**
     * @param column The column.
     * @return The data for the specified column.
     */
    public abstract Object getData(Column column);

    /**
     * @param column The column.
     * @return The data for the specified column as text.
     */
    public abstract String getDataAsText(Column column);

    /**
     * Sets the data for the specified column.
     *
     * @param column The column.
     * @param data   The data to set.
     */
    public abstract void setData(Column column, Object data);

    /** @return The height of this row. */
    public int getHeight() {
        return mHeight;
    }

    /**
     * Sets the height of this row.
     *
     * @param height The height to set.
     */
    public void setHeight(int height) {
        mHeight = height;
    }

    /**
     * @param columns The columns used to display this row.
     * @return The preferred height of this row.
     */
    public int getPreferredHeight(Outline outline, List<Column> columns) {
        int preferredHeight = 0;
        for (Column column : columns) {
            int height = column.getRowCell(this).getPreferredHeight(outline, this, column);
            if (height > preferredHeight) {
                preferredHeight = height;
            }
        }
        return preferredHeight;
    }

    /** @return The owning outline model. */
    public OutlineModel getOwner() {
        return mOwner;
    }

    /**
     * Sets the owner.
     *
     * @param owner The owning outline model.
     */
    void setOwner(OutlineModel owner) {
        mOwner = owner;
    }

    /** @return Whether this row can have children or not. */
    public boolean canHaveChildren() {
        return mChildren != null;
    }

    /**
     * Sets whether this row can have children or not.
     *
     * @param canHaveChildren Whether or not children are allowed.
     */
    public void setCanHaveChildren(boolean canHaveChildren) {
        if (canHaveChildren != canHaveChildren()) {
            if (canHaveChildren) {
                mChildren = new ArrayList<>();
            } else {
                setOpen(false);
                mChildren = null;
            }
        }
    }

    /** @return The children of this node. */
    public List<Row> getChildren() {
        return canHaveChildren() ? Collections.unmodifiableList(mChildren) : null;
    }

    /** @return The children of this node. */
    List<Row> getChildList() {
        return canHaveChildren() ? mChildren : null;
    }

    /**
     * @param index The child index.
     * @return The child at the specified index within this node.
     */
    public Row getChild(int index) {
        return index >= 0 && index < getChildCount() ? mChildren.get(index) : null;
    }

    /** @return The number of direct children this node contains. */
    public int getChildCount() {
        return canHaveChildren() ? mChildren.size() : 0;
    }

    /** @return Whether this row has children or not. */
    public boolean hasChildren() {
        return canHaveChildren() && !mChildren.isEmpty();
    }

    /** @return Whether this row is open, showing its children. */
    public boolean isOpen() {
        return mOpen;
    }

    /**
     * Sets whether this row is open, showing its children.
     *
     * @param open Whether this row is open or closed.
     */
    public void setOpen(boolean open) {
        if (mOpen != open && (!open || canHaveChildren())) {
            mOpen = open;
            if (mOwner != null) {
                mOwner.rowOpenStateChanged(this, mOpen);
            }
        }
    }

    /**
     * @param row The child row to determine the index of.
     * @return The index of the row, or {@code -1} if its not an immediate child.
     */
    public int getIndexOfChild(Row row) {
        if (canHaveChildren()) {
            return mChildren.indexOf(row);
        }
        return -1;
    }

    /**
     * Adds a child row to this row.
     *
     * @param index The index to insert at.
     * @param row   The row to add as a child.
     * @return {@code true} if the row was added, {@code false} if it was not.
     */
    public boolean insertChild(int index, Row row) {
        if (canHaveChildren()) {
            row.removeFromParent();
            if (index < 0) {
                index = 0;
            }
            int max = mChildren.size();
            if (index > max) {
                index = max;
            }
            mChildren.add(index, row);
            row.mParent = this;
            return true;
        }
        return false;
    }

    /**
     * Adds a child row to this row.
     *
     * @param row The row to add as a child.
     * @return {@code true} if the row was added, {@code false} if it was not.
     */
    public boolean addChild(Row row) {
        if (canHaveChildren()) {
            row.removeFromParent();
            mChildren.add(row);
            row.mParent = this;
            return true;
        }
        return false;
    }

    /**
     * Removes a child row from this row.
     *
     * @param row The child row to remove.
     * @return {@code true} if the row was removed, {@code false} if it wasn't a child of this row.
     */
    public boolean removeChild(Row row) {
        if (row.isChildOf(this)) {
            mChildren.remove(row);
            row.mParent = null;
            return true;
        }
        return false;
    }

    /**
     * @param parent The parent row.
     * @return {@code true} if this row is a child of the specified row.
     */
    public boolean isChildOf(Row parent) {
        return mParent == parent;
    }

    /**
     * @param row The descendant row.
     * @return {@code true} if this row is a descendant of the specified row.
     */
    public boolean isDescendantOf(Row row) {
        Row parent = mParent;
        while (parent != null) {
            if (parent == row) {
                return true;
            }
            parent = parent.mParent;
        }
        return false;
    }

    /** Removes this row from its parent, if any. */
    public void removeFromParent() {
        if (mParent != null) {
            mParent.removeChild(this);
        }
    }

    /** @return This row's parent row, if any. */
    public Row getParent() {
        return mParent;
    }

    /** @return The path from this row to its top-most parent. */
    public Row[] getPath() {
        ArrayList<Row> list   = new ArrayList<>();
        Row            parent = mParent;
        list.add(this);
        while (parent != null) {
            list.add(0, parent);
            parent = parent.mParent;
        }
        return list.toArray(new Row[0]);
    }

    /** @return The number of parents above this row. */
    public int getDepth() {
        Row parent = mParent;
        int depth  = 0;
        while (parent != null) {
            depth++;
            parent = parent.mParent;
        }
        return depth;
    }

    /** @return The string to display as tooltip. Defaults to {@code null}. */
    @SuppressWarnings({"static-method", "unused"})
    public String getToolTip(Column column) {
        return null;
    }
}
