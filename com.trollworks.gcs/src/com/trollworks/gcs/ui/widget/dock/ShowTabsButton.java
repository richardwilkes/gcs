/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget.dock;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Menu;
import com.trollworks.gcs.ui.widget.MenuItem;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.awt.Component;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.MouseInfo;
import java.awt.Point;
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
import javax.swing.JMenuItem;
import javax.swing.SwingConstants;

public class ShowTabsButton extends Panel implements MouseListener, MouseMotionListener, ComponentListener, ActionListener {
    private static final int MARGIN = 2;

    private Set<DockTab> mHidden = new HashSet<>();
    private boolean      mInMouseDown;
    private boolean      mPressed;
    private boolean      mRollover;

    public ShowTabsButton() {
        super(null, false);
        setToolTipText(I18n.text("Show Hidden Tabs List"));
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
        return new Dimension(getPreferredWidth(), MARGIN + insets.top + TextDrawing.getFontHeight(Fonts.LABEL_PRIMARY.getFont()) + insets.bottom + MARGIN);
    }

    @Override
    public Dimension getMinimumSize() {
        return getPreferredSize();
    }

    public int getPreferredWidth() {
        Insets insets = getInsets();
        return MARGIN + insets.left + TextDrawing.getSimpleWidth(Fonts.LABEL_PRIMARY.getFont(), getText()) + insets.right + MARGIN;
    }

    private String getText() {
        return "»" + Numbers.format(mHidden.size());
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        gc.setFont(Fonts.LABEL_PRIMARY.getFont());
        gc.setColor(UIUtilities.getIconButtonColor(isEnabled(), mInMouseDown, mPressed, mRollover));
        TextDrawing.draw(gc, UIUtilities.getLocalInsetBounds(this), getText(), SwingConstants.CENTER, SwingConstants.CENTER);
    }

    public void click() {
        Menu          menu = new Menu();
        List<DockTab> tabs = new ArrayList<>(mHidden);
        Collections.sort(tabs);
        for (DockTab tab : tabs) {
            Dockable dockable = tab.getDockable();
            menu.addItem(new MenuItem(tab.getIcon(), tab.getFullTitle(),
                    (p) -> dockable.getDockContainer().setCurrentDockable(dockable)));
        }
        menu.presentToUser(this, 0, this::updateRollOver);
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
        mRollover = true;
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
        mRollover = mPressed;
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
}
