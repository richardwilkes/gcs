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

import java.awt.Dimension;
import java.awt.LayoutManager;
import java.awt.Rectangle;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;

public class ScrollContent extends ContentPanel implements Scrollable {
    private int     mHScrollUnit;
    private int     mVScrollUnit;
    private boolean mTrackWidth;
    private boolean mTrackHeight;

    public ScrollContent(LayoutManager layoutManager) {
        super(layoutManager);
        mHScrollUnit = 16;
        mVScrollUnit = 16;
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return getPreferredSize();
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        return orientation == SwingConstants.HORIZONTAL ? mHScrollUnit : mVScrollUnit;
    }

    public void setHScrollUnit(int amount) {
        mHScrollUnit = amount;
    }

    public void setVScrollUnit(int amount) {
        mVScrollUnit = amount;
    }

    @Override
    public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
        return orientation == SwingConstants.HORIZONTAL ? visibleRect.width : visibleRect.height;
    }

    @Override
    public boolean getScrollableTracksViewportWidth() {
        return mTrackWidth;
    }

    public void setScrollableTracksViewportWidth(boolean track) {
        mTrackWidth = track;
    }

    @Override
    public boolean getScrollableTracksViewportHeight() {
        return mTrackHeight;
    }

    public void setScrollableTracksViewportHeight(boolean track) {
        mTrackHeight = track;
    }
}
