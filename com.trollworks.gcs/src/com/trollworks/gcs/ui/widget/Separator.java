/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.UIUtilities;

import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;

public class Separator extends Panel {
    private boolean mVertical;

    public Separator() {
        super(null, false);
    }

    public Separator(boolean vertical) {
        super(null, false);
        mVertical = vertical;
    }

    @Override
    protected void setStdColors() {
        setBackground(Colors.DIVIDER);
    }

    @Override
    public Dimension getMinimumSize() {
        Insets insets = getInsets();
        return new Dimension(1 + insets.left + insets.right, 1 + insets.top + insets.bottom);
    }

    @Override
    public Dimension getPreferredSize() {
        Insets insets = getInsets();
        return new Dimension(1 + insets.left + insets.right, 1 + insets.top + insets.bottom);
    }

    @Override
    public Dimension getMaximumSize() {
        Insets insets = getInsets();
        return new Dimension(insets.left + insets.right + (mVertical ? 1 : 100000),
                insets.top + insets.bottom + (mVertical ? 100000 : 1));
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        gc.setColor(getBackground());
        gc.fill(UIUtilities.getLocalInsetBounds(this));
    }
}
