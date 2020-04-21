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

import com.trollworks.gcs.ui.RetinaIcon;

import java.awt.Graphics2D;
import java.awt.Point;

public class IconTreeColumn extends TreeColumn {
    public static final int            HMARGIN = 2;
    public static final int            VMARGIN = 1;
    private             IconInteractor mIconInteractor;

    public IconTreeColumn(String name, IconInteractor iconInteractor) {
        super(name);
        mIconInteractor = iconInteractor;
    }

    protected RetinaIcon getIcon(TreeRow row) {
        return mIconInteractor != null ? mIconInteractor.getIcon(row) : null;
    }

    @Override
    public int compare(TreeRow o1, TreeRow o2) {
        return 0;
    }

    @Override
    public int calculatePreferredWidth(TreeRow row) {
        RetinaIcon icon = getIcon(row);
        return HMARGIN + (icon != null ? icon.getIconWidth() : 0) + HMARGIN;
    }

    @Override
    public int calculatePreferredHeight(TreeRow row, int width) {
        RetinaIcon icon = getIcon(row);
        return VMARGIN + (icon != null ? icon.getIconHeight() : 0) + VMARGIN;
    }

    @Override
    public void draw(Graphics2D gc, TreePanel panel, TreeRow row, int position, int top, int left, int width, boolean selected, boolean active) {
        RetinaIcon icon = getIcon(row);
        if (icon != null) {
            icon.paintIcon(panel, gc, left + (width - icon.getIconWidth()) / 2, top + (panel.getRowHeight(row) - icon.getIconHeight()) / 2);
        }
    }

    @Override
    public boolean mousePress(TreeRow row, Point where) {
        if (mIconInteractor != null) {
            return mIconInteractor.mousePress(row, this, where);
        }
        return false;
    }
}
