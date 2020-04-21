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

import java.awt.Dimension;
import java.awt.Rectangle;

/** A spacer within a {@link FlexLayout}. */
public class FlexSpacer extends FlexCell {
    private int     mWidth;
    private int     mHeight;
    private boolean mGrowWidth;
    private boolean mGrowHeight;

    /**
     * Creates a new {@link FlexSpacer}.
     *
     * @param width      The width of the spacer.
     * @param height     The height of the spacer.
     * @param growWidth  Whether the width of the spacer can grow.
     * @param growHeight Whether the height of the spacer can grow.
     */
    public FlexSpacer(int width, int height, boolean growWidth, boolean growHeight) {
        mWidth = width;
        mHeight = height;
        mGrowWidth = growWidth;
        mGrowHeight = growHeight;
    }

    @Override
    protected Dimension getSizeSelf(Scale scale, LayoutSize type) {
        int width  = scale.scale(mWidth);
        int height = scale.scale(mHeight);
        if (type == LayoutSize.MAXIMUM) {
            if (mGrowWidth) {
                width = Integer.MAX_VALUE;
            }
            if (mGrowHeight) {
                height = Integer.MAX_VALUE;
            }
        }
        return new Dimension(width, height);
    }

    @Override
    protected void layoutSelf(Scale scale, Rectangle bounds) {
        // Does nothing
    }
}
