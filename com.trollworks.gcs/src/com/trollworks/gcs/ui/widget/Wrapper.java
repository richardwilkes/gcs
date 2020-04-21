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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import javax.swing.JPanel;

/** A wrapper panel which is initially transparent. */
public class Wrapper extends JPanel {
    private int mWidth  = -1;
    private int mHeight = -1;

    /** Creates a new {@link Wrapper}. */
    public Wrapper() {
        setOpaque(false);
    }

    /**
     * Creates a new {@link Wrapper}.
     *
     * @param layout The layout to use.
     */
    public Wrapper(LayoutManager2 layout) {
        super(layout);
        setOpaque(false);
    }

    public void setOnlySize(int width, int height) {
        mWidth = width;
        mHeight = height;
    }

    @Override
    public Dimension getMinimumSize() {
        if (mWidth != -1 && mHeight != -1) {
            return getPreferredSize();
        }
        return super.getMinimumSize();
    }

    @Override
    public Dimension getPreferredSize() {
        if (mWidth != -1 && mHeight != -1) {
            Scale  scale  = Scale.get(this);
            Insets insets = getInsets();
            return new Dimension(insets.left + scale.scale(mWidth) + insets.right, insets.top + scale.scale(mHeight) + insets.bottom);
        }
        return super.getPreferredSize();
    }

    @Override
    public Dimension getMaximumSize() {
        if (mWidth != -1 && mHeight != -1) {
            return getPreferredSize();
        }
        return super.getMaximumSize();
    }
}
