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
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Container;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import javax.swing.SwingConstants;

public class MenuItem extends Panel {
    private String            mTitle;
    private SelectionListener mSelectionListener;
    private boolean           mHighlighted;

    public interface SelectionListener {
        void menuItemSelected(MenuItem item);
    }

    public MenuItem(String title, SelectionListener listener) {
        mTitle = title;
        mSelectionListener = listener;
        setThemeFont(ThemeFont.BUTTON);
        setBorder(new EmptyBorder(Button.V_MARGIN, Button.H_MARGIN, Button.V_MARGIN, Button.H_MARGIN));
    }

    public void click() {
        if (mSelectionListener != null) {
            mSelectionListener.menuItemSelected(this);
        }
    }

    public boolean isHighlighted() {
        return mHighlighted;
    }

    public void setHighlighted(boolean highlighted) {
        if (mHighlighted != highlighted) {
            mHighlighted = highlighted;
            setStdColors();
            repaint();
        }
    }

    @Override
    protected void setStdColors() {
        if (mHighlighted) {
            setBackground(ThemeColor.PRESSED_BUTTON);
            setForeground(ThemeColor.ON_PRESSED_BUTTON);
        } else {
            setBackground(ThemeColor.BUTTON);
            setForeground(ThemeColor.ON_BUTTON);
        }
    }

    @Override
    public void repaint(long tm, int x, int y, int width, int height) {
        Container parent = getParent();
        if (parent != null) {
            Rectangle bounds = new Rectangle(x, y, width, height);
            UIUtilities.convertRectangle(bounds, this, parent);
            parent.repaint(tm, bounds.x, bounds.y, bounds.width, bounds.height);
        }
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        Insets    insets = getInsets();
        Scale     scale  = Scale.get(this);
        Dimension size   = TextDrawing.getPreferredSize(scale.scale(getFont()), mTitle);
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        super.paintComponent(gc);
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        Scale     scale  = Scale.get(this);
        Font      font   = scale.scale(getFont());
        gc.setFont(font);
        TextDrawing.draw(gc, bounds, TextDrawing.truncateIfNecessary(font, mTitle, bounds.width,
                SwingConstants.CENTER), SwingConstants.LEFT, SwingConstants.CENTER);
    }
}
