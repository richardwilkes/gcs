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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.layout;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;

/** A flexible layout manager. */
public class FlexLayout implements LayoutManager2 {
	private FlexCell	mRootCell;

	/** Creates a new {@link FlexLayout}. */
	public FlexLayout() {
		// Does nothing.
	}

	/**
	 * Creates a new {@link FlexLayout}.
	 * 
	 * @param rootCell The root cell to layout.
	 */
	public FlexLayout(FlexCell rootCell) {
		mRootCell = rootCell;
	}

	/** @return The root cell. */
	public FlexCell getRootCell() {
		return mRootCell;
	}

	/** @param rootCell The value to set for the root cell. */
	public void setRootCell(FlexCell rootCell) {
		mRootCell = rootCell;
	}

	public void layoutContainer(Container target) {
		if (mRootCell != null) {
			Rectangle bounds = target.getBounds();
			Insets insets = target.getInsets();
			bounds.x = insets.left;
			bounds.y = insets.top;
			bounds.width -= insets.left + insets.right;
			bounds.height -= insets.top + insets.bottom;
			mRootCell.layout(bounds);
		}
	}

	private Dimension getLayoutSize(Container target, LayoutSize sizeType) {
		Insets insets = target.getInsets();
		Dimension size = mRootCell != null ? mRootCell.getSize(sizeType) : new Dimension();
		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom;
		return size;
	}

	public Dimension minimumLayoutSize(Container target) {
		return getLayoutSize(target, LayoutSize.MINIMUM);
	}

	public Dimension preferredLayoutSize(Container target) {
		return getLayoutSize(target, LayoutSize.PREFERRED);
	}

	public Dimension maximumLayoutSize(Container target) {
		return getLayoutSize(target, LayoutSize.MAXIMUM);
	}

	public float getLayoutAlignmentX(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public void invalidateLayout(Container target) {
		// Not used.
	}

	public void addLayoutComponent(Component comp, Object constraints) {
		// Not used.
	}

	public void addLayoutComponent(String name, Component comp) {
		// Not used.
	}

	public void removeLayoutComponent(Component comp) {
		// Not used.
	}
}
