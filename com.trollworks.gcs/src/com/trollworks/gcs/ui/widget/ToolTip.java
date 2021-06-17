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

import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;

import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import javax.swing.JComponent;
import javax.swing.JToolTip;
import javax.swing.SwingConstants;

public class ToolTip extends JToolTip {
    private String mWrappedText;

    public ToolTip(JComponent comp) {
        setUI(null);
        setBorder(new EmptyBorder(4, 8, 4, 8));
        setComponent(comp);
    }

    @Override
    public void setTipText(String tipText) {
        super.setTipText(tipText);
        mWrappedText = tipText != null ? TextDrawing.wrapToPixelWidth(ThemeFont.TOOLTIP.getFont(), tipText, 600) : "";
    }

    @Override
    public Dimension getPreferredSize() {
        if (mWrappedText.isBlank()) {
            return new Dimension();
        }
        Insets    insets = getInsets();
        Dimension size   = TextDrawing.getPreferredSize(ThemeFont.TOOLTIP.getFont(), mWrappedText);
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    protected void paintComponent(Graphics g) {
        if (!mWrappedText.isBlank()) {
            Graphics2D gc = GraphicsUtilities.prepare(g);
            gc.setColor(ThemeColor.TOOLTIP);
            Rectangle bounds = getBounds();
            bounds.x = 0;
            bounds.y = 0;
            gc.fill(bounds);
            gc.setFont(ThemeFont.TOOLTIP.getFont());
            gc.setColor(ThemeColor.ON_TOOLTIP);
            TextDrawing.draw(gc, UIUtilities.getLocalInsetBounds(this), mWrappedText, SwingConstants.LEFT, SwingConstants.CENTER);
        }
    }
}
