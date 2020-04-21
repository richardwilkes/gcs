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
import java.awt.Dimension;
import java.awt.Rectangle;

/** A {@link Component} within a {@link FlexLayout}. */
public class FlexComponent extends FlexCell {
    private Component mComponent;
    private boolean   mOnlyPreferredSize;

    /**
     * Creates a new {@link FlexComponent}.
     *
     * @param component The {@link Component} to wrap.
     */
    public FlexComponent(Component component) {
        mComponent = component;
    }

    /**
     * Creates a new {@link FlexComponent}.
     *
     * @param component         The {@link Component} to wrap.
     * @param onlyPreferredSize Whether only the preferred size is permitted.
     */
    public FlexComponent(Component component, boolean onlyPreferredSize) {
        mComponent = component;
        mOnlyPreferredSize = onlyPreferredSize;
    }

    /**
     * Creates a new {@link FlexComponent}.
     *
     * @param component           The {@link Component} to wrap.
     * @param horizontalAlignment The horizontal {@link Alignment} to use. Pass in {@code null} to
     *                            use the default.
     * @param verticalAlignment   The vertical {@link Alignment} to use. Pass in {@code null} to use
     *                            the default.
     */
    public FlexComponent(Component component, Alignment horizontalAlignment, Alignment verticalAlignment) {
        mComponent = component;
        if (horizontalAlignment != null) {
            setHorizontalAlignment(horizontalAlignment);
        }
        if (verticalAlignment != null) {
            setVerticalAlignment(verticalAlignment);
        }
    }

    /** @param onlyPreferredSize Whether only the preferred size is permitted. */
    public void setOnlyPreferredSize(boolean onlyPreferredSize) {
        mOnlyPreferredSize = onlyPreferredSize;
    }

    @Override
    protected Dimension getSizeSelf(Scale scale, LayoutSize type) {
        if (mOnlyPreferredSize) {
            type = LayoutSize.PREFERRED;
        }
        return type.get(mComponent);
    }

    @Override
    protected void layoutSelf(Scale scale, Rectangle bounds) {
        Rectangle compBounds = new Rectangle(bounds);
        if (!mOnlyPreferredSize) {
            Dimension size = LayoutSize.MINIMUM.get(mComponent);
            if (compBounds.width < size.width) {
                compBounds.width = size.width;
            }
            if (compBounds.height < size.height) {
                compBounds.height = size.height;
            }
            size = LayoutSize.MAXIMUM.get(mComponent);
            if (compBounds.width > size.width) {
                compBounds.width = size.width;
            }
            if (compBounds.height > size.height) {
                compBounds.height = size.height;
            }
        }
        mComponent.setBounds(Alignment.position(bounds, compBounds, getHorizontalAlignment(), getVerticalAlignment()));
    }

    @Override
    public String toString() {
        return mComponent.getClass().getSimpleName();
    }
}
