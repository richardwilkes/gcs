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

package com.trollworks.gcs.ui.layout;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;

/** A flexible layout manager. */
public class FlexLayout implements LayoutManager2 {
    private FlexCell mRootCell;

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

    @Override
    public void layoutContainer(Container target) {
        if (mRootCell != null) {
            Rectangle bounds = target.getBounds();
            Insets    insets = target.getInsets();
            bounds.x = insets.left;
            bounds.y = insets.top;
            bounds.width -= insets.left + insets.right;
            bounds.height -= insets.top + insets.bottom;
            mRootCell.layout(Scale.get(target), bounds);
        }
    }

    private Dimension getLayoutSize(Container target, LayoutSize sizeType) {
        Insets    insets = target.getInsets();
        Dimension size   = mRootCell != null ? mRootCell.getSize(Scale.get(target), sizeType) : new Dimension();
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    public Dimension minimumLayoutSize(Container target) {
        return getLayoutSize(target, LayoutSize.MINIMUM);
    }

    @Override
    public Dimension preferredLayoutSize(Container target) {
        return getLayoutSize(target, LayoutSize.PREFERRED);
    }

    @Override
    public Dimension maximumLayoutSize(Container target) {
        return getLayoutSize(target, LayoutSize.MAXIMUM);
    }

    @Override
    public float getLayoutAlignmentX(Container target) {
        return Component.CENTER_ALIGNMENT;
    }

    @Override
    public float getLayoutAlignmentY(Container target) {
        return Component.CENTER_ALIGNMENT;
    }

    @Override
    public void invalidateLayout(Container target) {
        // Not used.
    }

    @Override
    public void addLayoutComponent(Component comp, Object constraints) {
        // Not used.
    }

    @Override
    public void addLayoutComponent(String name, Component comp) {
        // Not used.
    }

    @Override
    public void removeLayoutComponent(Component comp) {
        // Not used.
    }
}
