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

package com.trollworks.gcs.ui;

import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.ColorWell;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.ThemePopupMenuSeparatorUI;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.Platform;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dialog;
import java.awt.Dialog.ModalityType;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Transparency;
import java.awt.Window;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import javax.swing.JComponent;
import javax.swing.JViewport;
import javax.swing.RepaintManager;
import javax.swing.UIManager;
import javax.swing.plaf.basic.BasicMenuBarUI;
import javax.swing.plaf.basic.BasicMenuItemUI;
import javax.swing.plaf.basic.BasicMenuUI;
import javax.swing.plaf.basic.BasicPopupMenuUI;

/** Various utility methods for the UI. */
public final class UIUtilities {
    private UIUtilities() {
    }

    /** Initialize the UI. */
    public static void initialize() {
        int scaleFactor = 1;
        if (Platform.isLinux()) {
            // Try to determine the display scaling factor and set it -- but only if none of the
            // properties or environment variables that should control it are set.
            String uiScale = System.getProperty("sun.java2d.uiScale");
            if (uiScale == null || uiScale.isBlank()) {
                String gdkScale = System.getenv("GDK_SCALE");
                if (gdkScale == null || gdkScale.isBlank()) {
                    String gdkDPIScale = System.getenv("GDK_DPI_SCALE");
                    if (gdkDPIScale == null || gdkDPIScale.isBlank()) {
                        int dpi = getXftDPI();
                        if (dpi > 0) {
                            scaleFactor = dpi / 96;
                            System.setProperty("sun.java2d.uiScale", Integer.toString(scaleFactor));
                        }
                    }
                }
            }
        }

        System.setProperty("apple.laf.useScreenMenuBar", Boolean.TRUE.toString());
        try {
            UIManager.setLookAndFeel(UIManager.getSystemLookAndFeelClassName());
        } catch (Exception ex) {
            Log.error(ex);
        }

        // The following two lines are here to ensure the theme is loaded
        Colors.currentThemeColors();
        Fonts.currentThemeFonts();

        UIManager.put("TextComponent.selectionBackgroundInactive", Colors.INACTIVE_SELECTION);

        if (!Platform.isMacintosh()) {
            UIManager.put("MenuBarUI", BasicMenuBarUI.class.getName());
            UIManager.put("MenuUI", BasicMenuUI.class.getName());
            UIManager.put("MenuItemUI", BasicMenuItemUI.class.getName());
            UIManager.put("PopupMenuUI", BasicPopupMenuUI.class.getName());
            UIManager.put("PopupMenuSeparatorUI", ThemePopupMenuSeparatorUI.class.getName());

            UIManager.put("MenuBar.opaque", Boolean.TRUE);
            UIManager.put("MenuBar.background", Colors.BACKGROUND);
            UIManager.put("MenuBar.foreground", Colors.ON_BACKGROUND);
            UIManager.put("MenuBar.shadow", Colors.BACKGROUND);
            UIManager.put("MenuBar.highlight", Colors.BACKGROUND);
            UIManager.put("MenuBar.border", new LineBorder(Colors.DIVIDER, 0, 0, 1, 0));

            UIManager.put("Menu.opaque", Boolean.TRUE);
            UIManager.put("Menu.background", Colors.BACKGROUND);
            UIManager.put("Menu.foreground", Colors.ON_BACKGROUND);
            DynamicColor disabledForeground = new DynamicColor(() -> Colors.getWithAlpha(Colors.ON_BACKGROUND, 128).getRGB());
            UIManager.put("Menu.disabledForeground", disabledForeground);
            UIManager.put("Menu.selectionBackground", Colors.SELECTION);
            UIManager.put("Menu.selectionForeground", Colors.ON_SELECTION);
            UIManager.put("Menu.font", Fonts.LABEL_PRIMARY.getFont());
            EmptyBorder menuBorder = new EmptyBorder(2 * scaleFactor, 4 * scaleFactor, 2 * scaleFactor, 4 * scaleFactor);
            UIManager.put("Menu.border", menuBorder);
            UIManager.put("Menu.borderPainted", Boolean.TRUE);

            UIManager.put("MenuItem.opaque", Boolean.TRUE);
            UIManager.put("MenuItem.borderPainted", Boolean.FALSE);
            UIManager.put("MenuItem.background", Colors.BACKGROUND);
            UIManager.put("MenuItem.foreground", Colors.ON_BACKGROUND);
            UIManager.put("MenuItem.disabledForeground", disabledForeground);
            UIManager.put("MenuItem.selectionBackground", Colors.SELECTION);
            UIManager.put("MenuItem.selectionForeground", Colors.ON_SELECTION);
            UIManager.put("MenuItem.acceleratorForeground", Colors.ON_BACKGROUND);
            UIManager.put("MenuItem.acceleratorSelectionForeground", Colors.ON_SELECTION);
            UIManager.put("MenuItem.font", Fonts.LABEL_PRIMARY.getFont());
            UIManager.put("MenuItem.acceleratorFont", Fonts.LABEL_SECONDARY.getFont());
            UIManager.put("MenuItem.border", menuBorder);
            UIManager.put("MenuItem.borderPainted", Boolean.TRUE);

            UIManager.put("PopupMenu.background", Colors.BACKGROUND);
            UIManager.put("PopupMenu.foreground", Colors.ON_BACKGROUND);
            UIManager.put("PopupMenu.border", new LineBorder(Colors.DIVIDER));

            UIManager.put("Panel.background", Colors.BACKGROUND);
            UIManager.put("Panel.foreground", Colors.ON_BACKGROUND);
        }
    }

    private static int getXftDPI() {
        int            dpi     = 0;
        ProcessBuilder builder = new ProcessBuilder("xrdb", "-q");
        builder.redirectOutput(ProcessBuilder.Redirect.PIPE).redirectErrorStream(true);
        try {
            Process process = builder.start();
            try (BufferedReader in = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String prefix = "Xft.dpi:";
                String line;
                while ((line = in.readLine()) != null) {
                    if (line.startsWith(prefix)) {
                        try {
                            dpi = Integer.parseInt(line.substring(prefix.length()).trim());
                        } catch (NumberFormatException nfex) {
                            System.err.println("unable to parse dpi from: " + line);
                        }
                    }
                }
            }
        } catch (IOException exception) {
            exception.printStackTrace(System.err);
        }
        return dpi;
    }

    /**
     * Disables all controls in the specified component and all its children.
     *
     * @param comp The {@link Component} to work on.
     */
    public static void disableControls(Component comp) {
        if (comp instanceof Container container) {
            int       count     = container.getComponentCount();
            for (int i = 0; i < count; i++) {
                disableControls(container.getComponent(i));
            }
        }
        if (comp instanceof Button ||
                comp instanceof Checkbox ||
                comp instanceof ColorWell ||
                comp instanceof EditorField ||
                comp instanceof FontIconButton ||
                comp instanceof MultiLineTextField ||
                comp instanceof PopupMenu) {
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
            if (window instanceof Dialog dialog) {
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

    /**
     * @param component The {@link JComponent} to work with.
     * @return The local bounds of the specified {@link JComponent}.
     */
    public static Rectangle getLocalBounds(JComponent component) {
        return new Rectangle(0, 0, component.getWidth(), component.getHeight());
    }

    public static void invalidateTree(Component comp) {
        comp.invalidate();
        if (comp instanceof Container) {
            for (Component child : ((Container) comp).getComponents()) {
                invalidateTree(child);
            }
        }
    }

    public static Color getIconButtonColor(boolean enabled, boolean inMouseDown, boolean pressed, boolean rollover) {
        if (enabled) {
            if (inMouseDown && pressed) {
                return Colors.ICON_BUTTON_PRESSED;
            }
            if (rollover) {
                return Colors.ICON_BUTTON_ROLLOVER;
            }
            return Colors.ICON_BUTTON;
        }
        return Colors.getWithAlpha(Colors.ICON_BUTTON, 96);
    }
}
