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

package com.trollworks.gcs.library;

import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.ScaleRoot;
import com.trollworks.gcs.ui.widget.outline.ListOutline;

import java.awt.BorderLayout;
import java.awt.Dimension;
import java.awt.Rectangle;
import javax.swing.JPanel;
import javax.swing.Scrollable;

public class LibraryContent extends JPanel implements ScaleRoot, Scrollable {
    private ListOutline mOutline;
    private Scale       mScale;

    public LibraryContent(ListOutline outline) {
        mOutline = outline;
        mOutline.setBorder(new LineBorder(ThemeColor.DIVIDER, 0, 0, 0, 1));
        setLayout(new BorderLayout());
        add(mOutline);
        mScale = Preferences.getInstance().getInitialUIScale().getScale();
    }

    @Override
    public Scale getScale() {
        return mScale;
    }

    @Override
    public void setScale(Scale scale) {
        if (mScale.getScale() != scale.getScale()) {
            mScale = scale;
            mOutline.sizeColumnsToFit();
            revalidate();
            repaint();
        }
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return mOutline.getPreferredScrollableViewportSize();
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        return mOutline.getScrollableUnitIncrement(visibleRect, orientation, direction);
    }

    @Override
    public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
        return mOutline.getScrollableBlockIncrement(visibleRect, orientation, direction);
    }

    @Override
    public boolean getScrollableTracksViewportWidth() {
        return mOutline.getScrollableTracksViewportWidth();
    }

    @Override
    public boolean getScrollableTracksViewportHeight() {
        return mOutline.getScrollableTracksViewportHeight();
    }
}
