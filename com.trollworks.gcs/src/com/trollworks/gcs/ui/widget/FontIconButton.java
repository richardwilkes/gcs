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
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.MouseInfo;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import javax.swing.SwingConstants;
import javax.swing.event.AncestorEvent;
import javax.swing.event.AncestorListener;

public class FontIconButton extends Panel implements MouseListener, MouseMotionListener, ComponentListener, AncestorListener, FocusListener {
    private String        mText;
    private ClickFunction mClickFunction;
    private int           mSize;
    private int           mMargin;
    private boolean       mInMouseDown;
    private boolean       mPressed;
    private boolean       mRollover;

    public interface ClickFunction {
        void buttonClicked(FontIconButton button);
    }

    public FontIconButton(String text, String tooltip, ClickFunction clickFunction) {
        super(null, false);
        setThemeFont(Fonts.FONT_ICON_STD);
        setToolTipText(tooltip);
        setText(text);
        setCursor(Cursor.getDefaultCursor());
        setClickFunction(clickFunction);
        addMouseListener(this);
        addMouseMotionListener(this);
        addComponentListener(this);
        addAncestorListener(this);
        addFocusListener(this);
    }

    public void setText(String text) {
        mText = text;
        repaint();
    }

    public void setMargin(int margin) {
        mMargin = margin;
    }

    @Override
    public Dimension getPreferredSize() {
        Scale     scale = Scale.get(this);
        Dimension size  = TextDrawing.getPreferredSize(scale.scale(getFont()), mText);
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

    public void setClickFunction(ClickFunction clickFunction) {
        mClickFunction = clickFunction;
    }

    public void click() {
        mClickFunction.buttonClicked(this);
    }

    @Override
    protected void paintComponent(Graphics g) {
        Rectangle bounds = getBounds();
        Insets    insets = getInsets();
        bounds.x = insets.left;
        bounds.y = insets.top;
        bounds.width -= insets.left + insets.right;
        bounds.height -= insets.top + insets.bottom;
        Graphics2D gc = GraphicsUtilities.prepare(g);
        gc.setColor(UIUtilities.getIconButtonColor(isEnabled(), mInMouseDown, mPressed, mRollover));
        Scale scale = Scale.get(this);
        gc.setFont(scale.scale(getFont()));
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

    public void updateRollOver() {
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

    @Override
    public void focusGained(FocusEvent event) {
        scrollRectToVisible(UIUtilities.getLocalBounds(this));
        repaint();
    }

    @Override
    public void focusLost(FocusEvent event) {
        repaint();
    }
}
