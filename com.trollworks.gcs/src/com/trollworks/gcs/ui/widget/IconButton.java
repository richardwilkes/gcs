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
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.MouseInfo;
import java.awt.Point;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import javax.swing.JComponent;

public class IconButton extends JComponent implements MouseListener, MouseMotionListener, ComponentListener {
    private RetinaIcon mIcon;
    private Runnable   mClickFunction;
    private int        mMargin = 2;
    private boolean    mInMouseDown;
    private boolean    mPressed;
    private boolean    mShowBorder;

    public IconButton(Img img, String tooltip, Runnable clickFunction) {
        this(new RetinaIcon(img, null), tooltip, clickFunction);
    }

    public IconButton(RetinaIcon icon, String tooltip, Runnable clickFunction) {
        setOpaque(false);
        setBackground(null);
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        setIcon(icon);
        setCursor(Cursor.getDefaultCursor());
        setClickFunction(clickFunction);
        addMouseListener(this);
        addMouseMotionListener(this);
        addComponentListener(this);
    }

    public void setIcon(Img img) {
        setIcon(new RetinaIcon(img, null));
    }

    public void setIcon(RetinaIcon icon) {
        mIcon = icon;
        repaint();
    }

    public void setMargin(int margin) {
        mMargin = margin;
    }

    @Override
    public Dimension getPreferredSize() {
        Scale scale = Scale.get(this);
        return new Dimension(scale.scale(mIcon.getIconWidth() + mMargin * 2), scale.scale(mIcon.getIconHeight() + mMargin * 2));
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
        Scale      scale  = Scale.get(this);
        int        width  = getWidth();
        int        height = getHeight();
        RetinaIcon icon   = mIcon;
        if (isEnabled()) {
            if (mInMouseDown && mPressed) {
                gc.setColor(Colors.adjustBrightness(getBackground(), -0.2f));
                gc.fillRect(0, 0, width, height);
            }
            if (mShowBorder || mInMouseDown) {
                gc.setColor(Colors.adjustBrightness(getBackground(), -0.4f));
                gc.drawRect(0, 0, width, height);
            }
        } else {
            icon = icon.createDisabled();
        }
        icon.paintIcon(this, gc, (width - scale.scale(icon.getIconWidth())) / 2, (height - scale.scale(icon.getIconHeight())) / 2);
    }

    private boolean isOver(int x, int y) {
        return x >= 0 && y >= 0 && x < getWidth() && y < getHeight();
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        mShowBorder = true;
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
            mShowBorder = mPressed;
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
    public void mouseClicked(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        mShowBorder = false;
        repaint();
    }

    private void updateRollOver() {
        boolean wasBorderShown = mShowBorder;
        Point   location       = MouseInfo.getPointerInfo().getLocation();
        UIUtilities.convertPointFromScreen(location, this);
        mShowBorder = isOver(location.x, location.y);
        if (wasBorderShown != mShowBorder) {
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
}
