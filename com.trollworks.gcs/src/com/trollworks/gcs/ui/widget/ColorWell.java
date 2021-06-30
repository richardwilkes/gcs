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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.Geometry;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import javax.swing.JColorChooser;

public class ColorWell extends Panel implements KeyListener, MouseListener, FocusListener {
    private Color                mColor;
    private ColorChangedListener mListener;
    private boolean              mRollover;

    public ColorWell(Color color, ColorChangedListener listener) {
        this(color, listener, null);
    }

    public ColorWell(Color color, ColorChangedListener listener, String tooltip) {
        mColor = color;
        mListener = listener;
        setToolTipText(tooltip);
        addKeyListener(this);
        addMouseListener(this);
        addFocusListener(this);
        setFocusable(true);
    }

    @Override
    public Dimension getMinimumSize() {
        if (isMinimumSizeSet()) {
            return super.getMinimumSize();
        }
        return new Dimension(20, 20);
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        return new Dimension(20, 20);
    }

    @Override
    public Dimension getMaximumSize() {
        if (isMaximumSizeSet()) {
            return super.getMaximumSize();
        }
        return new Dimension(20, 20);
    }

    public void setColorChangedListener(ColorChangedListener listener) {
        mListener = listener;
    }

    public Color getWellColor() {
        return mColor;
    }

    public void setWellColor(Color color) {
        mColor = color;
        repaint();
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc     = GraphicsUtilities.prepare(g);
        Rectangle  bounds = UIUtilities.getLocalInsetBounds(this);
        gc.setColor(Colors.ON_BACKGROUND);
        gc.fill(bounds);
        Geometry.inset(1, bounds);
        gc.setColor((mRollover || isFocusOwner()) ? Colors.SELECTION : Colors.BACKGROUND);
        gc.fill(bounds);
        Geometry.inset(1, bounds);
        gc.setColor(Color.WHITE);
        gc.fill(bounds);
        gc.setColor(Color.LIGHT_GRAY);
        int xs     = bounds.width / 4;
        int ys     = bounds.height / 4;
        int offset = 0;
        for (int y = bounds.y; y < bounds.y + bounds.height; y += ys) {
            for (int x = bounds.x + offset; x < bounds.x + bounds.width; x += xs * 2) {
                gc.fillRect(x, y, xs - 1, ys - 1);
            }
            offset = offset == 0 ? xs : 0;
        }
        gc.setColor(mColor);
        gc.fill(bounds);
    }

    public void click() {
        Color color = JColorChooser.showDialog(this, null, mColor);
        if (color != null) {
            if (!mColor.equals(color)) {
                mColor = color;
                repaint();
                if (mListener != null) {
                    mListener.colorChanged(mColor);
                }
            }
        }
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        click();
    }

    @Override
    public void mousePressed(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        mRollover = true;
        repaint();
    }

    @Override
    public void mouseExited(MouseEvent event) {
        mRollover = false;
        repaint();
    }

    @Override
    public void focusGained(FocusEvent event) {
        repaint();
    }

    @Override
    public void focusLost(FocusEvent event) {
        repaint();
    }

    @Override
    public void keyTyped(KeyEvent event) {
        if (!event.isConsumed() && (event.getModifiersEx() & Command.COMMAND_MODIFIER) == 0) {
            if (event.getKeyChar() == ' ') {
                event.consume();
                click();
            }
        }
    }

    @Override
    public void keyPressed(KeyEvent event) {
        // Unused
    }

    @Override
    public void keyReleased(KeyEvent event) {
        // Unused
    }

    public interface ColorChangedListener {
        void colorChanged(Color color);
    }
}
