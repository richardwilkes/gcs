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
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.RenderingHints;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.awt.geom.Path2D;
import javax.swing.SwingConstants;

public class StdButton extends StdPanel implements MouseListener, MouseMotionListener, KeyListener, FocusListener {
    private String        mText;
    private ThemeFont     mThemeFont;
    private ClickFunction mClickFunction;
    private boolean       mInMouseDown;
    private boolean       mPressed;

    public interface ClickFunction {
        void buttonClicked(StdButton button);
    }

    public StdButton(String text, ClickFunction clickFunction) {
        super(null, false);
        setThemeFont(ThemeFont.BUTTON);
        setText(text);
        setCursor(Cursor.getDefaultCursor());
        setClickFunction(clickFunction);
        setFocusable(true);
        addMouseListener(this);
        addMouseMotionListener(this);
        addKeyListener(this);
        addFocusListener(this);
    }

    public final ThemeFont getThemeFont() {
        return mThemeFont;
    }

    public final void setThemeFont(ThemeFont font) {
        mThemeFont = font;
    }

    @Override
    public final Font getFont() {
        return mThemeFont != null ? mThemeFont.getFont() : super.getFont();
    }

    /** @return The text. */
    public String getText() {
        return mText;
    }

    /** @param text The text to use. */
    public void setText(String text) {
        if (text == null) {
            text = "";
        }
        if (!text.equals(mText)) {
            mText = text;
            invalidate();
        }
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        Insets    insets = getInsets();
        Scale     scale  = Scale.get(this);
        Dimension size   = TextDrawing.getPreferredSize(scale.scale(getFont()), mText);
        size.width += insets.left + insets.right + scale.scale(16);
        size.height += insets.top + insets.bottom + scale.scale(4);
        return size;
    }

    @Override
    protected void paintComponent(Graphics g) {
        Rectangle bounds     = UIUtilities.getLocalInsetBounds(this);
        boolean   focusOwner = isFocusOwner();
        boolean   enabled    = isEnabled();
        Color     color;
        Color     onColor;
        if (mPressed) {
            color = ThemeColor.PRESSED_BUTTON;
            onColor = ThemeColor.ON_PRESSED_BUTTON;
        } else if (enabled) {
            if (focusOwner) {
                color = ThemeColor.FOCUS_BUTTON;
                onColor = ThemeColor.ON_FOCUS_BUTTON;
            } else {
                color = ThemeColor.BUTTON;
                onColor = ThemeColor.ON_BUTTON;
            }
        } else {
            color = ThemeColor.BUTTON;
            onColor = ThemeColor.ON_DISABLED_BUTTON;
        }

        Path2D.Double path         = new Path2D.Double();
        double        corner       = bounds.height / 3.0;
        double        top          = bounds.y;
        double        topCorner    = bounds.y + corner;
        double        left         = bounds.x;
        double        leftCorner   = bounds.x + corner;
        double        bottom       = bounds.y + bounds.height - 1;
        double        bottomCorner = bottom - corner;
        double        right        = bounds.x + bounds.width - 1;
        double        rightCorner  = right - corner;
        path.moveTo(leftCorner, top);
        path.lineTo(rightCorner, top);
        path.curveTo(rightCorner, top, right, top, right, topCorner);
        path.lineTo(right, bottomCorner);
        path.curveTo(right, bottomCorner, right, bottom, rightCorner, bottom);
        path.lineTo(leftCorner, bottom);
        path.curveTo(leftCorner, bottom, left, bottom, left, bottomCorner);
        path.lineTo(left, topCorner);
        path.curveTo(left, topCorner, left, top, leftCorner, top);
        path.closePath();

        Graphics2D gc = GraphicsUtilities.prepare(g);
        gc.setColor(color);
        gc.fill(path);

        Scale scale = Scale.get(this);
        Font  font  = scale.scale(getFont());
        gc.setFont(font);
        gc.setColor(onColor);
        Rectangle textBounds = new Rectangle(bounds.x + 2, bounds.y + 1, bounds.width - 4, bounds.height - 2);
        TextDrawing.draw(g, textBounds, TextDrawing.truncateIfNecessary(font, mText, bounds.width, SwingConstants.CENTER), SwingConstants.CENTER, SwingConstants.CENTER);

        gc.setColor(ThemeColor.BUTTON_BORDER);
        RenderingHints saved = GraphicsUtilities.setMaximumQualityForGraphics(gc);
        gc.draw(path);
        gc.setRenderingHints(saved);
    }

    @Override
    public void mouseClicked(MouseEvent event) {
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
    public void mouseReleased(MouseEvent event) {
        if (isEnabled()) {
            mouseDragged(event);
            mInMouseDown = false;
            MouseCapture.stop(this);
            if (mPressed) {
                mPressed = false;
                click();
            }
            repaint();
        }
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        // Unused
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
    public void mouseMoved(MouseEvent event) {
        // Unused
    }

    private boolean isOver(int x, int y) {
        return x >= 0 && y >= 0 && x < getWidth() && y < getHeight();
    }

    public void setClickFunction(ClickFunction clickFunction) {
        mClickFunction = clickFunction;
    }

    public void click() {
        boolean wasPressed = mPressed;
        mPressed = true;
        int width  = getWidth();
        int height = getHeight();
        paintImmediately(0, 0, width, height);
        try {
            Thread.sleep(100);
        } catch (InterruptedException ex) {
            // Ignore
        }
        mPressed = wasPressed;
        paintImmediately(0, 0, width, height);
        mClickFunction.buttonClicked(this);
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
        if (isEnabled() && event.getKeyChar() == ' ' && event.getModifiersEx() == 0) {
            click();
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
}
