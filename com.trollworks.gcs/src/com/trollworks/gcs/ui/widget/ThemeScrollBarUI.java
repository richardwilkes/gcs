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

import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import javax.swing.BoundedRangeModel;
import javax.swing.JComponent;
import javax.swing.JScrollBar;
import javax.swing.plaf.ComponentUI;
import javax.swing.plaf.basic.BasicScrollBarUI;

public class ThemeScrollBarUI extends BasicScrollBarUI {
    @SuppressWarnings("MethodOverridesStaticMethodOfSuperclass")
    public static ComponentUI createUI(JComponent c) {
        return new ThemeScrollBarUI();
    }

    @Override
    protected void installDefaults() {
        super.installDefaults();
        scrollBarWidth = 16;
        incrGap = 0;
        decrGap = 0;
        scrollbar.setBorder(null);
    }

    @Override
    protected void configureScrollBarColors() {
        // Unused
    }

    @Override
    protected void installComponents() {
        // Unused
    }

    @Override
    protected void uninstallComponents() {
        // Unused
    }

    @Override
    protected void layoutHScrollbar(JScrollBar sb) {
        trackRect = UIUtilities.getLocalInsetBounds(sb);
        BoundedRangeModel model = sb.getModel();
        double            max   = model.getMaximum();
        int               start = (int) (trackRect.width * (model.getValue() / max));
        int               size  = (int) (trackRect.width * (model.getExtent() / max));
        if (size < 4) {
            size = 4;
        }
        setThumbBounds(start, 3, size, trackRect.height - 6);
    }

    @Override
    protected void layoutVScrollbar(JScrollBar sb) {
        trackRect = UIUtilities.getLocalInsetBounds(sb);
        BoundedRangeModel model = sb.getModel();
        double            max   = model.getMaximum();
        int               start = (int) (trackRect.height * (model.getValue() / max));
        int               size  = (int) (trackRect.height * (model.getExtent() / max));
        if (size < 4) {
            size = 4;
        }
        setThumbBounds(3, start, trackRect.width - 6, size);
    }

    @Override
    public void paint(Graphics g, JComponent c) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        gc.setColor(Colors.BACKGROUND);
        gc.fill(getTrackBounds());
        gc.setColor(isThumbRollover() ? Colors.SCROLL_ROLLOVER : Colors.SCROLL);
        RenderingHints hints = GraphicsUtilities.setMaximumQualityForGraphics(gc);
        gc.fillRoundRect(thumbRect.x, thumbRect.y, thumbRect.width, thumbRect.height, 8, 8);
        gc.setRenderingHints(hints);
    }
}
