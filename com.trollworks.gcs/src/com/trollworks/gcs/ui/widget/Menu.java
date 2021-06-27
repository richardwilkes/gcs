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
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.SystemEventHandler;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.utility.task.Tasks;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
import java.awt.LayoutManager2;
import java.awt.MouseInfo;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Window;
import java.awt.desktop.AppForegroundEvent;
import java.awt.desktop.AppForegroundListener;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.awt.event.MouseWheelEvent;
import java.awt.event.MouseWheelListener;
import java.util.concurrent.TimeUnit;
import javax.swing.JComponent;
import javax.swing.Popup;
import javax.swing.PopupFactory;
import javax.swing.SwingConstants;
import javax.swing.event.AncestorEvent;
import javax.swing.event.AncestorListener;

public class Menu extends Panel implements Runnable, MouseListener, MouseMotionListener, MouseWheelListener, KeyListener, FocusListener, AncestorListener, ComponentListener, AppForegroundListener, LayoutManager2 {
    private Popup     mPopup;
    private Component mRestoreFocusTo;
    private Runnable  mCallbackWhenDone;
    private MenuItem  mSelection;
    private int       mTop;
    private Rectangle mTopScrollArea;
    private Rectangle mBottomScrollArea;
    private int       mInitialFocusAttemptsRemaining;

    public Menu() {
        setLayout(this);
        addMouseListener(this);
        addMouseMotionListener(this);
        addMouseWheelListener(this);
        addKeyListener(this);
        addAncestorListener(this);
        addFocusListener(this);
        setFocusable(true);
    }

    @Override
    protected void setStdColors() {
        setBackground(Colors.BUTTON);
        setForeground(Colors.ON_BUTTON);
    }

    public MenuItem getSelection() {
        return mSelection;
    }

    public void setTop(int index) {
        if (index < 0) {
            index = 0;
        } else {
            int max = getMaxTop();
            if (index > max) {
                index = max;
            }
        }
        if (mTop != index) {
            mTop = index;
            doLayout();
            repaint();
        }
    }

    private int getMaxTop() {
        int count = getComponentCount();
        if (count != 0) {
            Insets insets    = getInsets();
            int    available = getHeight() - (insets.top + insets.bottom);
            if (count > 1) {
                Scale     scale  = Scale.get(this);
                Font      faFont = new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(getFont()).getSize());
                Dimension faSize = TextDrawing.getPreferredSize(faFont, FontAwesome.CARET_DOWN);
                faSize.height += 2 * scale.scale(Button.V_MARGIN);
                available -= faSize.height;
                if (count > 2) {
                    available -= faSize.height;
                }
            }
            for (int i = getComponentCount() - 1; i >= 0; i--) {
                available -= getComponent(i).getHeight();
                if (available <= 0) {
                    return i;
                }
            }
        }
        return 0;
    }

    private void close() {
        SystemEventHandler.INSTANCE.removeAppForegroundListener(this);
        for (Window window : Window.getWindows()) {
            window.removeComponentListener(this);
        }
        if (mPopup != null) {
            MouseCapture.stop(this);
            mPopup.hide();
            mPopup = null;
            if (mRestoreFocusTo != null) {
                mRestoreFocusTo.requestFocus();
                mRestoreFocusTo = null;
            }
            if (mSelection != null && mSelection.isEnabled()) {
                mSelection.click();
            }
            if (mCallbackWhenDone != null) {
                mCallbackWhenDone.run();
            }
        }
    }

    public void addSeparator() {
        Separator sep = new Separator();
        sep.setBorder(new EmptyBorder(2, 0, 2, 0));
        add(sep, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    public void addSeparatorAt(int index) {
        Separator sep = new Separator();
        sep.setBorder(new EmptyBorder(2, 0, 2, 0));
        add(sep, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true), index);
    }

    public void addItem(MenuItem item) {
        add(item, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    public void addItemAt(int index, MenuItem item) {
        add(item, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true), index);
    }

    public void addCommand(Command cmd) {
        addItem(new MenuItem(cmd));
    }

    public void addCommandAt(int index, Command cmd) {
        addItemAt(index, new MenuItem(cmd));
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        // Unused
    }

    @Override
    public void mousePressed(MouseEvent event) {
        Rectangle bounds = getBounds();
        bounds.x = 0;
        bounds.y = 0;
        Point pt = event.getPoint();
        if (!bounds.contains(pt)) {
            processMouseOver(pt);
            close();
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        processMouseOver(event.getPoint());
        close();
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        processMouseOver(event.getPoint());
    }

    @Override
    public void mouseExited(MouseEvent event) {
        if (mSelection != null) {
            mSelection.setHighlighted(false);
            mSelection = null;
        }
    }

    @Override
    public void mouseDragged(MouseEvent event) {
        processMouseOver(event.getPoint());
    }

    @Override
    public void mouseMoved(MouseEvent event) {
        processMouseOver(event.getPoint());
    }

    private void processMouseOver(Point where) {
        Component over = getComponentAt(where);
        if (over != mSelection) {
            if (mSelection != null) {
                mSelection.setHighlighted(false);
            }
            if (over instanceof MenuItem && over.isEnabled()) {
                mSelection = (MenuItem) over;
                mSelection.setHighlighted(true);
            } else {
                mSelection = null;
            }
        }
    }

    @Override
    public void mouseWheelMoved(MouseWheelEvent event) {
        if (event.getScrollType() == MouseWheelEvent.WHEEL_UNIT_SCROLL) {
            setTop(mTop + event.getUnitsToScroll());
            processMouseOver(event.getPoint());
        }
    }

    @Override
    public void keyTyped(KeyEvent event) {
        // Unused
    }

    @Override
    public void keyPressed(KeyEvent event) {
        if (event.getModifiersEx() == 0) {
            int max     = getComponentCount();
            int keyCode = event.getKeyCode();
            if (keyCode == KeyEvent.VK_UP) {
                int index = (mSelection != null ? UIUtilities.getIndexOf(this, mSelection) : max) - 1;
                if (index >= 0) {
                    while (index >= 0) {
                        Component comp = getComponent(index);
                        if (comp instanceof MenuItem && comp.isEnabled()) {
                            break;
                        }
                        index--;
                    }
                    if (index >= 0) {
                        Component comp = getComponent(index);
                        if (comp instanceof MenuItem) {
                            if (mSelection != null) {
                                mSelection.setHighlighted(false);
                            }
                            mSelection = ((MenuItem) comp);
                            mSelection.setHighlighted(true);
                            scrollToSelection();
                        }
                    }
                }
            } else if (keyCode == KeyEvent.VK_DOWN) {
                int index = (mSelection != null ? UIUtilities.getIndexOf(this, mSelection) : -1) + 1;
                if (index < max) {
                    while (index < max) {
                        Component comp = getComponent(index);
                        if (comp instanceof MenuItem && comp.isEnabled()) {
                            break;
                        }
                        index++;
                    }
                    if (index < max) {
                        Component comp = getComponent(index);
                        if (comp instanceof MenuItem) {
                            if (mSelection != null) {
                                mSelection.setHighlighted(false);
                            }
                            mSelection = ((MenuItem) comp);
                            mSelection.setHighlighted(true);
                            scrollToSelection();
                        }
                    }
                }
            } else if (keyCode == KeyEvent.VK_SPACE || keyCode == KeyEvent.VK_ENTER || keyCode == KeyEvent.VK_COMMA) {
                if (mSelection != null) {
                    close();
                }
            } else if (keyCode == KeyEvent.VK_ESCAPE) {
                mSelection = null;
                close();
            }
        }
    }

    private void scrollToSelection() {
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        if (mTopScrollArea != null) {
            bounds.y = mTopScrollArea.y + mTopScrollArea.height;
            bounds.height -= mTopScrollArea.height;
        }
        if (mBottomScrollArea != null) {
            bounds.height -= mBottomScrollArea.height;
        }
        Rectangle compBounds = mSelection.getBounds();
        int       state      = -1;
        while (!bounds.contains(compBounds)) {
            if (compBounds.y < bounds.y) {
                if (state == 1) {
                    break;
                }
                state = 0;
                int top = mTop;
                setTop(mTop - 1);
                if (top == mTop) {
                    break;
                }
            } else if (compBounds.y + compBounds.height > bounds.y + bounds.height) {
                if (state == 0) {
                    break;
                }
                state = 1;
                int top = mTop;
                setTop(mTop + 1);
                if (top == mTop) {
                    break;
                }
            } else {
                // Shouldn't ever hit this... but just in case.
                break;
            }
            compBounds = mSelection.getBounds();
        }
    }

    @Override
    public void keyReleased(KeyEvent event) {
        // Unused
    }

    @Override
    public void focusGained(FocusEvent event) {
        // Unused
    }

    @Override
    public void focusLost(FocusEvent event) {
        close();
    }

    @Override
    public void ancestorAdded(AncestorEvent event) {
        WindowUtils.getWindowForComponent(this).setFocusableWindowState(true);
        for (Window window : Window.getWindows()) {
            window.addComponentListener(this);
        }
        SystemEventHandler.INSTANCE.addAppForegroundListener(this);
        EventQueue.invokeLater(this);
    }

    @Override
    public void ancestorRemoved(AncestorEvent event) {
        // Unused
    }

    @Override
    public void ancestorMoved(AncestorEvent event) {
        // Unused
    }

    @Override
    public void componentResized(ComponentEvent event) {
        close();
    }

    @Override
    public void componentMoved(ComponentEvent event) {
        close();
    }

    @Override
    public void componentShown(ComponentEvent event) {
        // Unused
    }

    @Override
    public void componentHidden(ComponentEvent event) {
        close();
    }

    @Override
    public void appRaisedToForeground(AppForegroundEvent event) {
        // Unused
    }

    @Override
    public void appMovedToBackground(AppForegroundEvent event) {
        close();
    }

    @Override
    public void addLayoutComponent(String name, Component comp) {
        // Unused
    }

    @Override
    public void addLayoutComponent(Component comp, Object constraints) {
        // Unused
    }

    @Override
    public void removeLayoutComponent(Component comp) {
        // Unused
    }

    @Override
    public float getLayoutAlignmentX(Container target) {
        return 0;
    }

    @Override
    public float getLayoutAlignmentY(Container target) {
        return 0;
    }

    @Override
    public void invalidateLayout(Container target) {
        // Unused
    }

    @Override
    public Dimension minimumLayoutSize(Container target) {
        Dimension size = new Dimension();
        for (Component comp : getComponents()) {
            Dimension compSize = comp.getMinimumSize();
            if (size.width < compSize.width) {
                size.width = compSize.width;
            }
            if (size.height < compSize.height) {
                size.height = compSize.height;
            }
        }
        if (getComponentCount() > 1) {
            Scale     scale  = Scale.get(this);
            Font      faFont = new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(getFont()).getSize());
            Dimension faSize = TextDrawing.getPreferredSize(faFont, FontAwesome.CARET_DOWN);
            faSize.width += 2 * scale.scale(Button.H_MARGIN);
            if (size.width < faSize.width) {
                size.width = faSize.width;
            }
            faSize.height += 2 * scale.scale(Button.V_MARGIN);
            size.height += faSize.height;
            if (getComponentCount() > 2) {
                size.height += faSize.height;
            }
        }
        Insets insets = getInsets();
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    public Dimension preferredLayoutSize(Container target) {
        Dimension size = new Dimension();
        for (Component comp : getComponents()) {
            Dimension compSize = comp.getPreferredSize();
            if (size.width < compSize.width) {
                size.width = compSize.width;
            }
            size.height += compSize.height;
        }
        Insets insets = getInsets();
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    public Dimension maximumLayoutSize(Container target) {
        return preferredLayoutSize(target);
    }

    @Override
    public void layoutContainer(Container target) {
        Scale     scale              = Scale.get(this);
        Font      faFont             = new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(getFont()).getSize());
        Dimension faSize             = TextDrawing.getPreferredSize(faFont, FontAwesome.CARET_DOWN);
        int       scrollMarkerHeight = faSize.height + 2 * scale.scale(Button.V_MARGIN);
        Rectangle bounds             = UIUtilities.getLocalInsetBounds(this);
        int       y                  = bounds.y;
        if (mTop != 0) {
            y += scrollMarkerHeight;
        }
        int max = getComponentCount();
        for (int i = 0; i < max; i++) {
            getComponent(i).setLocation(bounds.x, -10000);
        }
        for (int i = mTop; i < max; i++) {
            Component comp     = getComponent(i);
            Dimension compSize = comp.getPreferredSize();
            comp.setBounds(bounds.x, y, bounds.width, compSize.height);
            y += compSize.height;
        }
    }

    @Override
    protected void paintChildren(Graphics g) {
        super.paintChildren(g);
        Graphics2D gc     = GraphicsUtilities.prepare(g);
        Rectangle  bounds = UIUtilities.getLocalInsetBounds(this);
        Scale      scale  = Scale.get(this);
        Font       faFont = new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(getFont()).getSize());
        Dimension  faSize = TextDrawing.getPreferredSize(faFont, FontAwesome.CARET_DOWN);
        faSize.height += 2 * scale.scale(Button.V_MARGIN);
        gc.setFont(faFont);
        mBottomScrollArea = null;
        if (mTop == 0) {
            mTopScrollArea = null;
        } else {
            mTopScrollArea = new Rectangle(bounds.x, bounds.y, bounds.width, faSize.height);
            TextDrawing.draw(gc, mTopScrollArea, FontAwesome.CARET_UP, SwingConstants.CENTER, SwingConstants.CENTER);
        }
        int count = getComponentCount();
        if (count > 0) {
            Rectangle compBounds = getComponent(count - 1).getBounds();
            if (compBounds.y + compBounds.height > bounds.y + bounds.height) {
                mBottomScrollArea = new Rectangle(bounds.x, bounds.y + bounds.height - faSize.height, bounds.width, faSize.height);
                gc.setColor(getBackground());
                gc.fill(mBottomScrollArea);
                gc.setColor(getForeground());
                TextDrawing.draw(gc, mBottomScrollArea, FontAwesome.CARET_DOWN, SwingConstants.CENTER, SwingConstants.CENTER);
            }
        }
    }

    @Override
    public void run() {
        if (mPopup != null) {
            if (mTopScrollArea != null || mBottomScrollArea != null) {
                Point pt = MouseInfo.getPointerInfo().getLocation();
                UIUtilities.convertPointFromScreen(pt, this);
                if (mTopScrollArea != null && mTopScrollArea.contains(pt)) {
                    setTop(mTop - 1);
                    if (mSelection != null) {
                        mSelection.setHighlighted(false);
                        mSelection = null;
                    }
                } else if (mBottomScrollArea != null && mBottomScrollArea.contains(pt)) {
                    setTop(mTop + 1);
                    if (mSelection != null) {
                        mSelection.setHighlighted(false);
                        mSelection = null;
                    }
                }
            }
            Tasks.scheduleOnUIThread(this, 100, TimeUnit.MILLISECONDS, null);
        }
    }

    public void presentToUser(JComponent owner, int initialIndex, Runnable callbackWhenDone) {
        presentToUser(owner, UIUtilities.getLocalInsetBounds(owner), initialIndex, callbackWhenDone);
    }

    public void presentToUser(JComponent owner, Rectangle controlBounds, int initialIndex, Runnable callbackWhenDone) {
        mCallbackWhenDone = callbackWhenDone;
        mRestoreFocusTo = KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
        for (Component comp : getComponents()) {
            if (comp instanceof MenuItem) {
                ((MenuItem) comp).adjust();
            }
        }
        Dimension prefSize = getPreferredSize();
        if (prefSize.width < controlBounds.width) {
            prefSize.width = controlBounds.width;
        }
        Rectangle maxBounds = WindowUtils.getMaximumWindowBounds(owner, controlBounds);
        if (prefSize.height > maxBounds.height) {
            prefSize.height = maxBounds.height;
        }
        setPreferredSize(prefSize);
        setSize(prefSize);
        doLayout(); // force layout so the calls to get the component locations will give us correct data
        Point pt = new Point(controlBounds.x, controlBounds.y - getComponent(initialIndex).getY());
        UIUtilities.convertPointToScreen(pt, owner);
        if (pt.y < maxBounds.y) {
            int delta = maxBounds.y - pt.y;
            setTop(1);
            for (int j = 0; j < initialIndex; j++) {
                Component comp   = getComponent(j);
                Rectangle bounds = comp.getBounds();
                if (bounds.y >= delta) {
                    setTop(j);
                    pt = new Point(controlBounds.x, controlBounds.y - getComponent(initialIndex).getY());
                    UIUtilities.convertPointToScreen(pt, owner);
                    break;
                }
            }
        }
        if (pt.y + prefSize.height > maxBounds.y + maxBounds.height) {
            prefSize.height = maxBounds.y + maxBounds.height - pt.y;
            int minHeight = getMinimumSize().height;
            if (minHeight > prefSize.height) {
                pt.y -= minHeight - prefSize.height;
                prefSize.height = minHeight;
            }
            setPreferredSize(prefSize);
            setSize(prefSize);
        }
        PopupFactory factory = PopupFactory.getSharedInstance();
        mPopup = factory.getPopup(owner, this, pt.x, pt.y);
        mPopup.show();
        MouseCapture.start(owner, this, this, null);
        mInitialFocusAttemptsRemaining = 5;
        tryInitialFocus();
    }

    private void tryInitialFocus() {
        if (--mInitialFocusAttemptsRemaining > 0 && !isFocusOwner()) {
            requestFocus();
            EventQueue.invokeLater(this::tryInitialFocus);
        }
    }
}
