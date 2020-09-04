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

package com.trollworks.gcs.ui;

import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.WiderToolTipUI;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.Platform;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dialog;
import java.awt.Dialog.ModalityType;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Transparency;
import java.awt.Window;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import javax.swing.AbstractButton;
import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.JViewport;
import javax.swing.RepaintManager;
import javax.swing.UIManager;

/** Various utility methods for the UI. */
public final class UIUtilities {
    private UIUtilities() {
    }

    /** Initialize the UI. */
    public static void initialize() {
        System.setProperty("apple.laf.useScreenMenuBar", Boolean.TRUE.toString());
        try {
            UIManager.setLookAndFeel(UIManager.getSystemLookAndFeelClassName());
            Font current = UIManager.getFont(Fonts.KEY_STD_TEXT_FIELD);
            UIManager.getDefaults().put(Fonts.KEY_STD_TEXT_FIELD, new Font("SansSerif", current.getStyle(), current.getSize()));
            WiderToolTipUI.installIfNeeded();
        } catch (Exception ex) {
            Log.error(ex);
        }
        Theme.current(); // Just here to ensure the theme is loaded
        Fonts.loadFromPreferences();
    }

    /**
     * Selects the tab with the specified title.
     *
     * @param pane  The {@link JTabbedPane} to use.
     * @param title The title to select.
     */
    public static void selectTab(JTabbedPane pane, String title) {
        int count = pane.getTabCount();
        for (int i = 0; i < count; i++) {
            if (pane.getTitleAt(i).equals(title)) {
                pane.setSelectedIndex(i);
                break;
            }
        }
    }

    /**
     * Disables all controls in the specified component and all its children.
     *
     * @param comp The {@link Component} to work on.
     */
    public static void disableControls(Component comp) {
        if (comp instanceof Container) {
            Container container = (Container) comp;
            int       count     = container.getComponentCount();

            for (int i = 0; i < count; i++) {
                disableControls(container.getComponent(i));
            }
        }

        if (comp instanceof AbstractButton || comp instanceof JComboBox || comp instanceof JTextField || comp instanceof IconButton) {
            comp.setEnabled(false);
        }
    }

    /**
     * Sets a {@link Component}'s min, max & preferred sizes to a specific size.
     *
     * @param comp The {@link Component} to work on.
     * @param size The size to set the component to.
     */
    public static void setOnlySize(Component comp, Dimension size) {
        comp.setMinimumSize(size);
        comp.setMaximumSize(size);
        comp.setPreferredSize(size);
    }

    /**
     * Sets a {@link Component}'s min & max sizes to its preferred size. Note that this applies a
     * fudge factor to the width on the Windows platform to attempt to get around highly inaccurate
     * font measurements.
     *
     * @param comp The {@link Component} to work on.
     */
    public static void setToPreferredSizeOnly(Component comp) {
        Dimension size = comp.getPreferredSize();
        if (Platform.isWindows()) {
            size.width += 4; // Fudge the width on Windows, since its text measurement seems to be
            // off in many cases
        }
        setOnlySize(comp, size);
    }

    /** @param comps The {@link Component}s to set to the same size. */
    public static void adjustToSameSize(Component... comps) {
        Dimension best = new Dimension();
        for (Component comp : comps) {
            Dimension size = comp.getPreferredSize();
            if (size.width > best.width) {
                best.width = size.width;
            }
            if (size.height > best.height) {
                best.height = size.height;
            }
        }
        for (Component comp : comps) {
            setOnlySize(comp, best);
        }
    }

    /**
     * Converts a {@link Point} from one component's coordinate system to another's.
     *
     * @param pt   The point to convert.
     * @param from The component the point originated in.
     * @param to   The component the point should be translated to.
     */
    public static void convertPoint(Point pt, Component from, Component to) {
        convertPointToScreen(pt, from);
        convertPointFromScreen(pt, to);
    }

    /**
     * Converts a {@link Point} from on the screen to a position within the component.
     *
     * @param pt        The point to convert.
     * @param component The component the point should be translated to.
     */
    public static void convertPointFromScreen(Point pt, Component component) {
        while (component != null) {
            pt.x -= component.getX();
            pt.y -= component.getY();
            if (component instanceof Window) {
                break;
            }
            component = component.getParent();
        }
    }

    /**
     * Converts a {@link Point} in a component to its position on the screen.
     *
     * @param pt        The point to convert.
     * @param component The component the point originated in.
     */
    public static void convertPointToScreen(Point pt, Component component) {
        while (component != null) {
            pt.x += component.getX();
            pt.y += component.getY();
            if (component instanceof Window) {
                break;
            }
            component = component.getParent();
        }
    }

    /**
     * Converts a {@link Rectangle} from one component's coordinate system to another's.
     *
     * @param bounds The rectangle to convert.
     * @param from   The component the rectangle originated in.
     * @param to     The component the rectangle should be translated to.
     */
    public static void convertRectangle(Rectangle bounds, Component from, Component to) {
        convertRectangleToScreen(bounds, from);
        convertRectangleFromScreen(bounds, to);
    }

    /**
     * Converts a {@link Rectangle} from on the screen to a position within the component.
     *
     * @param bounds    The rectangle to convert.
     * @param component The component the rectangle should be translated to.
     */
    public static void convertRectangleFromScreen(Rectangle bounds, Component component) {
        while (component != null) {
            bounds.x -= component.getX();
            bounds.y -= component.getY();
            if (component instanceof Window) {
                break;
            }
            component = component.getParent();
        }
    }

    /**
     * Converts a {@link Rectangle} in a component to its position on the screen.
     *
     * @param bounds    The rectangle to convert.
     * @param component The component the rectangle originated in.
     */
    public static void convertRectangleToScreen(Rectangle bounds, Component component) {
        while (component != null) {
            bounds.x += component.getX();
            bounds.y += component.getY();
            if (component instanceof Window) {
                break;
            }
            component = component.getParent();
        }
    }

    /**
     * @param parent The parent {@link Container}.
     * @param child  The child {@link Component}.
     * @return The index of the specified {@link Component}. -1 will be returned if the {@link
     *         Component} isn't a direct child.
     */
    public static int getIndexOf(Container parent, Component child) {
        if (parent != null) {
            int count = parent.getComponentCount();
            for (int i = 0; i < count; i++) {
                if (child == parent.getComponent(i)) {
                    return i;
                }
            }
        }
        return -1;
    }

    /**
     * @param comp The component to work with.
     * @return Whether the component should be expanded to fit.
     */
    public static boolean shouldTrackViewportWidth(Component comp) {
        Container parent = comp.getParent();
        if (parent instanceof JViewport) {
            Dimension available = parent.getSize();
            Dimension prefSize  = comp.getPreferredSize();
            return prefSize.width < available.width;
        }
        return false;
    }

    /**
     * @param comp The component to work with.
     * @return Whether the component should be expanded to fit.
     */
    public static boolean shouldTrackViewportHeight(Component comp) {
        Container parent = comp.getParent();
        if (parent instanceof JViewport) {
            Dimension available = parent.getSize();
            Dimension prefSize  = comp.getPreferredSize();
            return prefSize.height < available.height;
        }
        return false;
    }

    /** @param comp The component to revalidate. */
    public static void revalidateImmediately(Component comp) {
        if (comp != null) {
            RepaintManager mgr = RepaintManager.currentManager(comp);
            mgr.validateInvalidComponents();
            mgr.paintDirtyRegions();
        }
    }

    /**
     * @param component The component to be looked at.
     * @param type      The type of component being looked for.
     * @return The first object that matches, starting with the component itself and working up
     *         through its parents, or {@code null}.
     */
    @SuppressWarnings("unchecked")
    public static <T> T getSelfOrAncestorOfType(Component component, Class<T> type) {
        if (component != null) {
            if (type.isAssignableFrom(component.getClass())) {
                return (T) component;
            }
            return getAncestorOfType(component, type);
        }
        return null;
    }

    /**
     * @param component The component whose ancestor chain is to be looked at.
     * @param type      The type of ancestor being looked for.
     * @return The ancestor, or {@code null}.
     */
    @SuppressWarnings("unchecked")
    public static <T> T getAncestorOfType(Component component, Class<T> type) {
        if (component == null) {
            return null;
        }
        Container parent = component.getParent();
        while (parent != null && !type.isAssignableFrom(parent.getClass())) {
            parent = parent.getParent();
        }
        return (T) parent;
    }

    /**
     * Since JComboBox.getSelectedItem() returns a plain Object, this allows us to get the
     * appropriate type of object instead.
     */
    public static <E> E getTypedSelectedItemFromCombo(JComboBox<E> combo) {
        int index = combo.getSelectedIndex();
        return index == -1 ? null : combo.getItemAt(index);
    }

    /**
     * @param component The component to generate an image of.
     * @return The newly created image.
     */
    public static Img getImage(JComponent component) {
        Img offscreen = null;
        synchronized (component.getTreeLock()) {
            Graphics2D gc = null;
            try {
                Rectangle bounds = component.getVisibleRect();
                offscreen = Img.create(component.getGraphicsConfiguration(), bounds.width, bounds.height, Transparency.TRANSLUCENT);
                gc = offscreen.getGraphics();
                gc.translate(-bounds.x, -bounds.y);
                component.paint(gc);
            } catch (Exception exception) {
                Log.error(exception);
            } finally {
                if (gc != null) {
                    gc.dispose();
                }
            }
        }
        return offscreen;
    }

    /**
     * @param obj The object to extract a {@link Component} for.
     * @return The {@link Component} to use for the dialog, or {@code null}.
     */
    public static Component getComponentForDialog(Object obj) {
        return obj instanceof Component ? (Component) obj : null;
    }

    /** @return Whether or not the application is currently in a modal state. */
    public static boolean inModalState() {
        for (Window window : Window.getWindows()) {
            if (window instanceof Dialog) {
                Dialog dialog = (Dialog) window;
                if (dialog.isShowing()) {
                    ModalityType type = dialog.getModalityType();
                    if (type == ModalityType.APPLICATION_MODAL || type == ModalityType.TOOLKIT_MODAL) {
                        return true;
                    }
                }
            }
        }
        return false;
    }

    public static Point convertDropTargetDragPointTo(DropTargetDragEvent dtde, Component comp) {
        Point pt = dtde.getLocation();
        convertPoint(pt, dtde.getDropTargetContext().getComponent(), comp);
        return pt;
    }

    public static void updateDropTargetDragPointTo(DropTargetDragEvent dtde, Component comp) {
        convertPoint(dtde.getLocation(), dtde.getDropTargetContext().getComponent(), comp);
    }

    public static void updateDropTargetDropPointTo(DropTargetDropEvent dtde, Component comp) {
        convertPoint(dtde.getLocation(), dtde.getDropTargetContext().getComponent(), comp);
    }

    /**
     * @param component The {@link JComponent} to work with.
     * @return The local, inset, bounds of the specified {@link JComponent}.
     */
    public static Rectangle getLocalInsetBounds(JComponent component) {
        Insets insets = component.getInsets();
        return new Rectangle(insets.left, insets.top, component.getWidth() - (insets.left + insets.right), component.getHeight() - (insets.top + insets.bottom));
    }
}
