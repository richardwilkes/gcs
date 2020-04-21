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

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexLayout;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.LayoutManager;
import java.awt.Rectangle;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;

/** A simple panel that draws banded colors behind its contents. */
public class BandedPanel extends ActionPanel implements Scrollable {
    private String mTitle;

    /**
     * Creates a new {@link BandedPanel}.
     *
     * @param title The title for this panel.
     */
    public BandedPanel(String title) {
        super(new ColumnLayout(1, 0, 0));
        setOpaque(true);
        setBackground(Color.white);
        mTitle = title;
    }

    @Override
    protected void paintComponent(Graphics gc) {
        super.paintComponent(GraphicsUtilities.prepare(gc));
        Rectangle bounds = getBounds();
        bounds.x = 0;
        bounds.y = 0;
        int step  = getStep();
        int count = getComponentCount();
        for (int i = 0; i < count; i += step) {
            Rectangle compBounds = getComponent(i).getBounds();
            bounds.y = compBounds.y;
            bounds.height = compBounds.height;
            int logical = i / step;
            gc.setColor(Colors.getBanding(logical % 2 == 0));
            gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
        }
    }

    private int getStep() {
        LayoutManager layout = getLayout();
        if (layout instanceof ColumnLayout) {
            return ((ColumnLayout) layout).getColumns();
        } else if (layout instanceof FlexLayout && ((FlexLayout) layout).getRootCell() instanceof FlexGrid) {
            int columns = ((FlexGrid) ((FlexLayout) layout).getRootCell()).getColumnCount();
            return columns - columns / 2;
        }
        return 1;
    }

    @Override
    public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
        return orientation == SwingConstants.VERTICAL ? visibleRect.height : visibleRect.width;
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return getPreferredSize();
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        int height = 10;
        int count  = getComponentCount();
        if (count > 0) {
            count = Math.min(getStep(), count);
            for (int i = 0; i < count; i++) {
                int tmp = getComponent(i).getHeight();
                if (tmp > height) {
                    height = tmp;
                }
            }
        }
        return height;
    }

    @Override
    public boolean getScrollableTracksViewportHeight() {
        return UIUtilities.shouldTrackViewportHeight(this);
    }

    @Override
    public boolean getScrollableTracksViewportWidth() {
        return UIUtilities.shouldTrackViewportWidth(this);
    }

    @Override
    public String toString() {
        return mTitle;
    }
}
