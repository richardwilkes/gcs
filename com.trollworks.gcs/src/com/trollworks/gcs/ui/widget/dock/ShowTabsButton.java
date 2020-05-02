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

package com.trollworks.gcs.ui.widget.dock;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Component;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.MouseInfo;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.JComponent;
import javax.swing.JMenuItem;
import javax.swing.JPopupMenu;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

public class ShowTabsButton extends JComponent implements MouseListener, MouseMotionListener, ComponentListener, ActionListener {
    private static final int          MARGIN  = 2;
    private              boolean      mInMouseDown;
    private              boolean      mPressed;
    private              boolean      mShowBorder;
    private              Set<DockTab> mHidden = new HashSet<>();

    public ShowTabsButton() {
        setOpaque(false);
        setBackground(null);
        setFont(UIManager.getFont("Label.font"));
        setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Show Hidden Tabs List")));
        setCursor(Cursor.getDefaultCursor());
        addMouseListener(this);
        addMouseMotionListener(this);
        addComponentListener(this);
    }

    public void addHidden(DockTab tab) {
        mHidden.add(tab);
        repaint();
    }

    public void clearHidden() {
        mHidden.clear();
        repaint();
    }

    public boolean hasHidden() {
        return !mHidden.isEmpty();
    }

    public boolean isHidden(Component component) {
        if (component instanceof DockTab) {
            return mHidden.contains(component);
        }
        if (component == this) {
            return !hasHidden();
        }
        return false;
    }

    @Override
    public Dimension getPreferredSize() {
        Insets insets = getInsets();
        return new Dimension(getPreferredWidth(), MARGIN + insets.top + TextDrawing.getFontHeight(getFont()) + insets.bottom + MARGIN);
    }

    @Override
    public Dimension getMinimumSize() {
        return getPreferredSize();
    }

    public int getPreferredWidth() {
        Insets insets = getInsets();
        return MARGIN + insets.left + TextDrawing.getSimpleWidth(getFont(), getText()) + insets.right + MARGIN;
    }

    private String getText() {
        return "»" + Numbers.format(mHidden.size());
    }

    @Override
    protected void paintComponent(Graphics gc) {
        Insets insets = getInsets();
        int    x      = insets.left;
        int    y      = insets.top;
        int    width  = getWidth() - (insets.left + insets.right);
        int    height = getHeight() - (insets.top + insets.bottom);
        if (mInMouseDown && mPressed) {
            gc.setColor(Colors.adjustBrightness(getBackground(), -0.2f));
            gc.fillRect(x, y, width, height);
        }
        if (mShowBorder || mInMouseDown) {
            gc.setColor(Colors.adjustBrightness(getBackground(), -0.4f));
            gc.drawRect(x, y, width - 1, height - 1);
        }
        gc.setFont(getFont());
        gc.setColor(Color.BLACK);
        String    text   = getText();
        Rectangle bounds = getBounds();
        bounds.x = insets.left;
        bounds.y = insets.top;
        bounds.width -= insets.left + insets.right;
        bounds.height -= insets.top + insets.bottom;
        TextDrawing.draw(gc, bounds, text, SwingConstants.CENTER, SwingConstants.CENTER);
    }

    public void click() {
        JPopupMenu    menu = new JPopupMenu();
        List<DockTab> tabs = new ArrayList<>(mHidden);
        Collections.sort(tabs);
        for (DockTab tab : tabs) {
            JMenuItem item = new JMenuItem(tab.getFullTitle(), tab.getIcon());
            item.putClientProperty(this, tab.getDockable());
            item.addActionListener(this);
            menu.add(item);
        }
        menu.show(this, 0, 0);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Dockable dockable = (Dockable) ((JMenuItem) event.getSource()).getClientProperty(this);
        dockable.getDockContainer().setCurrentDockable(dockable);
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
        if (!event.isPopupTrigger() && event.getButton() == 1) {
            mInMouseDown = true;
            mPressed = true;
            repaint();
            MouseCapture.start(this, Cursor.getDefaultCursor());
        }
    }

    @Override
    public void mouseDragged(MouseEvent event) {
        boolean wasPressed = mPressed;
        mPressed = isOver(event.getX(), event.getY());
        if (mPressed != wasPressed) {
            repaint();
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
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
