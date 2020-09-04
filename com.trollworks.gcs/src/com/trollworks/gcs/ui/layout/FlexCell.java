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

import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.Rectangle;

/** The basic unit within a {@link FlexLayout}. */
public abstract class FlexCell {
    private Alignment mHorizontalAlignment = Alignment.LEFT_TOP;
    private Alignment mVerticalAlignment   = Alignment.CENTER;
    private Insets    mInsets              = new Insets(0, 0, 0, 0);

    /**
     * Creates a new {@link FlexLayout} with this cell as its root cell and applies it to the
     * specified component.
     *
     * @param container The container to apply the {@link FlexLayout} to.
     */
    public void apply(Container container) {
        container.setLayout(new FlexLayout(this));
    }

    /**
     * Layout the cell and its children.
     *
     * @param scale  The {@link Scale} to use.
     * @param bounds The bounds to use for the cell.
     */
    public final void layout(Scale scale, Rectangle bounds) {
        Insets insets = scale.scale(mInsets);
        bounds.x += insets.left;
        bounds.y += insets.top;
        bounds.width -= insets.left + insets.right;
        bounds.height -= insets.top + insets.bottom;
        layoutSelf(scale, bounds);
    }

    /**
     * Called to layout the cell and its children.
     *
     * @param scale  The {@link Scale} to use.
     * @param bounds The bounds to use for the cell. Insets have already been applied.
     */
    protected abstract void layoutSelf(Scale scale, Rectangle bounds);

    /**
     * @param scale The {@link Scale} to use.
     * @param type  The type of size to determine.
     * @return The size for this cell.
     */
    public final Dimension getSize(Scale scale, LayoutSize type) {
        Insets    insets = scale.scale(mInsets);
        Dimension size   = getSizeSelf(scale, type);
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return LayoutSize.sanitizeSize(size);
    }

    /**
     * @param scale The {@link Scale} to use.
     * @param type  The type of size to determine.
     * @return The size for this cell. Do not include the insets from the cell.
     */
    protected abstract Dimension getSizeSelf(Scale scale, LayoutSize type);

    /** @return The horizontal alignment. */
    public Alignment getHorizontalAlignment() {
        return mHorizontalAlignment;
    }

    /** @param alignment The value to set for horizontal alignment. */
    public void setHorizontalAlignment(Alignment alignment) {
        mHorizontalAlignment = alignment;
    }

    /** @return The vertical alignment. */
    public Alignment getVerticalAlignment() {
        return mVerticalAlignment;
    }

    /** @param alignment The value to set for vertical alignment. */
    public void setVerticalAlignment(Alignment alignment) {
        mVerticalAlignment = alignment;
    }

    /** @return The insets. */
    public Insets getInsets() {
        return mInsets;
    }

    /** @param insets The value to set for insets. */
    public void setInsets(Insets insets) {
        if (insets != null) {
            mInsets.set(insets.top, insets.left, insets.bottom, insets.right);
        } else {
            mInsets.set(0, 0, 0, 0);
        }
    }

    @Override
    public String toString() {
        return getClass().getSimpleName();
    }
}
