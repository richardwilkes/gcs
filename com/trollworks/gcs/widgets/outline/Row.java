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

package com.trollworks.gcs.widgets.outline;

import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/** Represents a single row of data within an {@link OutlineModel}. */
public abstract class Row {
	private OutlineModel		mOwner;
	private int					mHeight;
	private boolean				mOpen;
	private Row					mParent;
	/** The children of this row. */
	protected ArrayList<Row>	mChildren;

	/** Create a new outline row. */
	public Row() {
		mHeight = -1;
	}

	/**
	 * @param owner The owning model.
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
	 * @return The icon for the specified column, or <code>null</code>.
	 */
	public BufferedImage getIcon(@SuppressWarnings("unused") Column column) {
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
	 * @param data The data to set.
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
	public int getPreferredHeight(List<Column> columns) {
		int preferredHeight = 0;

		for (Column column : columns) {
			int height = column.getRowCell(this).getPreferredHeight(this, column);

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
				mChildren = new ArrayList<Row>();
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
	ArrayList<Row> getChildList() {
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
		if (mOpen != open && (!open || open && canHaveChildren())) {
			mOpen = open;
			if (mOwner != null) {
				mOwner.rowOpenStateChanged(this, mOpen);
			}
		}
	}

	/**
	 * @param row The child row to determine the index of.
	 * @return The index of the row, or <code>-1</code> if its not an immediate child.
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
	 * @param row The row to add as a child.
	 * @return <code>true</code> if the row was added, <code>false</code> if it was not.
	 */
	public boolean insertChild(int index, Row row) {
		if (canHaveChildren()) {
			int max;

			row.removeFromParent();
			if (index < 0) {
				index = 0;
			}
			max = mChildren.size();
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
	 * @return <code>true</code> if the row was added, <code>false</code> if it was not.
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
	 * @return <code>true</code> if the row was removed, <code>false</code> if it wasn't a child
	 *         of this row.
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
	 * @return <code>true</code> if this row is a child of the specified row.
	 */
	public boolean isChildOf(Row parent) {
		return mParent == parent;
	}

	/**
	 * @param row The descendent row.
	 * @return <code>true</code> if this row is a descendent of the specified row.
	 */
	public boolean isDescendentOf(Row row) {
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
		ArrayList<Row> list = new ArrayList<Row>();
		Row parent = mParent;

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
		int depth = 0;

		while (parent != null) {
			depth++;
			parent = parent.mParent;
		}
		return depth;
	}
}
