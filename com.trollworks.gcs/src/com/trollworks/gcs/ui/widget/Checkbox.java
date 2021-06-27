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
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.TextDrawing;
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
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import javax.swing.SwingConstants;

public class Checkbox extends Panel implements MouseListener, MouseMotionListener, KeyListener, FocusListener {
    private String        mText;
    private ClickFunction mClickFunction;
    private boolean       mInMouseDown;
    private boolean       mPressed;
    private boolean       mChecked;

    public interface ClickFunction {
        void checkboxClicked(Checkbox checkbox);
    }

    public Checkbox(String text, boolean checked, ClickFunction clickFunction) {
        super(null, false);
        setText(text);
        setCursor(Cursor.getDefaultCursor());
        setClickFunction(clickFunction);
        setFocusable(true);
        addMouseListener(this);
        addMouseMotionListener(this);
        addKeyListener(this);
        addFocusListener(this);
        mChecked = checked;
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

    public boolean isChecked() {
        return mChecked;
    }

    public void setChecked(boolean checked) {
        mChecked = checked;
        repaint();
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
        mChecked = !mChecked;
        mPressed = wasPressed;
        paintImmediately(0, 0, width, height);
        if (mClickFunction != null) {
            mClickFunction.checkboxClicked(this);
        }
    }

    @Override
    public void focusGained(FocusEvent event) {
        Rectangle bounds = getBounds();
        bounds.x = 0;
        bounds.y = 0;
        scrollRectToVisible(bounds);
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

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        Insets    insets = getInsets();
        Scale     scale  = Scale.get(this);
        Font      font   = scale.scale(getFont());
        Dimension size   = TextDrawing.getPreferredSize(new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, font.getSize()), FontAwesome.CHECK_CIRCLE);
        if (!mText.isBlank()) {
            Dimension textSize = TextDrawing.getPreferredSize(font, mText);
            size.width += textSize.width + scale.scale(4);
            size.height = Math.max(size.height, textSize.height);
        }
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom + scale.scale(4);
        return size;
    }

    @Override
    protected void paintComponent(Graphics g) {
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        Color     color;
        if (mPressed) {
            color = Colors.PRESSED_ICON_BUTTON;
        } else if (isEnabled()) {
            color = Colors.ICON_BUTTON;
        } else {
            color = Colors.getWithAlpha(Colors.ICON_BUTTON, 96);
        }
        Graphics2D gc         = GraphicsUtilities.prepare(g);
        Scale      scale      = Scale.get(this);
        Font       font       = scale.scale(getFont());
        boolean    focusOwner = isFocusOwner();
        Font       iconFont   = new Font(focusOwner ? Fonts.FONT_AWESOME_SOLID : Fonts.FONT_AWESOME_REGULAR, Font.PLAIN, font.getSize());
        Dimension  size       = TextDrawing.getPreferredSize(iconFont, FontAwesome.CHECK_CIRCLE);
        gc.setFont(iconFont);
        gc.setColor(color);
        Rectangle textBounds = new Rectangle(bounds.x, bounds.y, size.width, bounds.height);
        String    unchecked  = focusOwner ? FontAwesome.DOT_CIRCLE : FontAwesome.CIRCLE;
        TextDrawing.draw(gc, textBounds, mChecked ? FontAwesome.CHECK_CIRCLE : unchecked, SwingConstants.CENTER, SwingConstants.CENTER);
        if (!mText.isBlank()) {
            gc.setFont(font);
            gc.setColor(color);
            bounds.x += size.width + scale.scale(4);
            bounds.width -= size.width + scale.scale(4);
            TextDrawing.draw(gc, bounds, TextDrawing.truncateIfNecessary(font, mText, bounds.width, SwingConstants.CENTER), SwingConstants.LEFT, SwingConstants.CENTER);
        }
    }
}
