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

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.MouseInfo;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import javax.swing.JComponent;
import javax.swing.SwingConstants;
import javax.swing.event.AncestorEvent;
import javax.swing.event.AncestorListener;

public class FontAwesomeButton extends JComponent implements MouseListener, MouseMotionListener, ComponentListener, AncestorListener {
    private String   mText;
    private Runnable mClickFunction;
    private int      mSize;
    private int      mMargin;
    private boolean  mInMouseDown;
    private boolean  mPressed;
    private boolean  mRollover;

    public FontAwesomeButton(String text, String tooltip, Runnable clickFunction) {
        this(text, 14, tooltip, clickFunction);
    }

    public FontAwesomeButton(String text, int size, String tooltip, Runnable clickFunction) {
        setOpaque(false);
        setBackground(null);
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        setText(text);
        setSize(size);
        setCursor(Cursor.getDefaultCursor());
        setClickFunction(clickFunction);
        addMouseListener(this);
        addMouseMotionListener(this);
        addComponentListener(this);
        addAncestorListener(this);
    }

    public void setText(String text) {
        mText = text;
        repaint();
    }

    public void setSize(int size) {
        mSize = size;
    }

    public void setMargin(int margin) {
        mMargin = margin;
    }

    @Override
    public Dimension getPreferredSize() {
        Scale     scale = Scale.get(this);
        Dimension size  = TextDrawing.getPreferredSize(new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(mSize)), mText);
        if (mMargin != 0) {
            size.width += scale.scale(mMargin) * 2;
            size.height *= scale.scale(mMargin) * 2;
        }
        Insets insets = getInsets();
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    public Dimension getMinimumSize() {
        return getPreferredSize();
    }

    @Override
    public Dimension getMaximumSize() {
        return getPreferredSize();
    }

    public void setClickFunction(Runnable clickFunction) {
        mClickFunction = clickFunction;
    }

    public void click() {
        mClickFunction.run();
    }

    @Override
    protected void paintComponent(Graphics gc) {
        Rectangle bounds = getBounds();
        Insets    insets = getInsets();
        bounds.x = insets.left;
        bounds.y = insets.top;
        bounds.width -= insets.left + insets.right;
        bounds.height -= insets.top + insets.bottom;
        Color color;
        if (isEnabled()) {
            if (mInMouseDown && mPressed) {
                color = ThemeColor.PRESSED_ICON_BUTTON;
            } else if (mRollover) {
                color = ThemeColor.ROLLOVER_ICON_BUTTON;
            } else {
                color = ThemeColor.ICON_BUTTON;
            }
        } else {
            color = ThemeColor.DISABLED_ICON_BUTTON;
        }
        gc.setColor(color);
        Scale scale = Scale.get(this);
        gc.setFont(new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(mSize)));
        TextDrawing.draw(gc, bounds, mText, SwingConstants.CENTER, SwingConstants.CENTER);
    }

    private boolean isOver(int x, int y) {
        return x >= 0 && y >= 0 && x < getWidth() && y < getHeight();
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        mRollover = true;
        repaint();
    }

    @Override
    public void mouseMoved(MouseEvent event) {
        // Unused
    }

    @Override
    public void mousePressed(MouseEvent event) {
        if (isEnabled() && !event.isPopupTrigger() && event.getButton() == 1) {
            mInMouseDown = true;
            mPressed = true;
            repaint();
            MouseCapture.start(this, Cursor.getDefaultCursor());
        }
    }

    @Override
    public void mouseDragged(MouseEvent event) {
        if (isEnabled()) {
            boolean wasPressed = mPressed;
            mPressed = isOver(event.getX(), event.getY());
            if (mPressed != wasPressed) {
                repaint();
            }
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        if (isEnabled()) {
            mouseDragged(event);
            mRollover = mPressed;
            mInMouseDown = false;
            MouseCapture.stop(this);
            if (mPressed) {
                mPressed = false;
                click();
            }
            repaint();
            updateRollOver();
        }
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        mRollover = false;
        repaint();
    }

    private void updateRollOver() {
        boolean wasRollover = mRollover;
        Point   location    = MouseInfo.getPointerInfo().getLocation();
        UIUtilities.convertPointFromScreen(location, this);
        mRollover = isOver(location.x, location.y);
        if (wasRollover != mRollover) {
            repaint();
        }
    }

    @Override
    public void componentResized(ComponentEvent event) {
        updateRollOver();
    }

    @Override
    public void componentMoved(ComponentEvent event) {
        updateRollOver();
    }

    @Override
    public void componentShown(ComponentEvent event) {
        updateRollOver();
    }

    @Override
    public void componentHidden(ComponentEvent event) {
        updateRollOver();
    }

    @Override
    public void ancestorAdded(AncestorEvent event) {
        updateRollOver();
    }

    @Override
    public void ancestorRemoved(AncestorEvent event) {
        updateRollOver();
    }

    @Override
    public void ancestorMoved(AncestorEvent event) {
        updateRollOver();
    }
}
