/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.collections.FilteredIterator;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.menu.file.QuitCommand;
import com.trollworks.gcs.menu.file.SaveCommand;
import com.trollworks.gcs.menu.file.Saveable;
import com.trollworks.gcs.preferences.MenuKeyPreferences;
import com.trollworks.gcs.ui.WindowSizeEnforcer;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.utility.FileProxy;
import com.trollworks.gcs.utility.FileProxyProvider;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.awt.AWTEvent;
import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Frame;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Window;
import java.awt.event.MouseWheelEvent;
import java.awt.event.WindowEvent;
import java.awt.event.WindowFocusListener;
import java.awt.event.WindowListener;
import java.io.File;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.JFrame;
import javax.swing.JToolBar;
import javax.swing.WindowConstants;

/** Provides a base OS-level window. */
public class BaseWindow extends JFrame implements Undoable, Comparable<BaseWindow>, WindowListener, WindowFocusListener {
    private static final String                WINDOW_PREFERENCES         = "WindowPrefs";
    private static final int                   WINDOW_PREFERENCES_VERSION = 3;
    private static final String                KEY_LOCATION               = "Location";
    private static final String                KEY_SIZE                   = "Size";
    private static final String                KEY_LAST_UPDATED           = "LastUpdated";
    private static final ArrayList<BaseWindow> WINDOW_LIST                = new ArrayList<>();
    private              StdUndoManager        mUndoManager;
    private              boolean               mIsClosed;
    boolean mWasAlive;

    /**
     * Creates a new {@link BaseWindow}.
     *
     * @param title The window title. May be {@code null}.
     */
    public BaseWindow(String title) {
        super(title, null);
        setDefaultCloseOperation(WindowConstants.DO_NOTHING_ON_CLOSE);
        setLocationByPlatform(true);
        ((JComponent) getContentPane()).setDoubleBuffered(true);
        getToolkit().setDynamicLayout(true);
        addWindowListener(this);
        addWindowFocusListener(this);
        WindowSizeEnforcer.monitor(this);
        MenuKeyPreferences.loadFromPreferences();
        setJMenuBar(new StdMenuBar());
        setIconImages(Images.APP_ICON_LIST);
        mUndoManager = new StdUndoManager();
        enableEvents(AWTEvent.KEY_EVENT_MASK | AWTEvent.MOUSE_EVENT_MASK);
        WINDOW_LIST.add(this);
    }

    /** @return {@code true} if the window has been closed. */
    public final boolean isClosed() {
        return mIsClosed;
    }

    @Override
    public void setVisible(boolean visible) {
        if (visible) {
            mWasAlive = true;
        }
        super.setVisible(visible);
    }

    /** Call to create the toolbar for this window. */
    protected final void createToolBar() {
        JToolBar toolbar = new JToolBar();
        toolbar.setFloatable(false);
        FlexRow row = new FlexRow();
        row.setInsets(new Insets(2, 5, 2, 5));
        createToolBarContents(toolbar, row);
        row.apply(toolbar);
        add(toolbar, BorderLayout.NORTH);
    }

    /**
     * Called to create the toolbar contents for this window.
     *
     * @param toolbar The {@link JToolBar} to add items to.
     * @param row     The {@link FlexRow} layout to add items to.
     */
    protected void createToolBarContents(JToolBar toolbar, FlexRow row) {
        // Does nothing by default.
    }

    @Override
    public void toFront() {
        if (!isClosed()) {
            if (getExtendedState() == ICONIFIED) {
                setExtendedState(NORMAL);
            }
            super.toFront();
            if (!isActive() || !isFocused()) {
                Component focus = getMostRecentFocusOwner();
                if (focus != null) {
                    focus.requestFocus();
                } else {
                    requestFocus();
                }
            }
        }
    }

    /** @param event The {@link MouseWheelEvent}. */
    protected void processMouseWheelEventSuper(MouseWheelEvent event) {
        super.processMouseWheelEvent(event);
    }

    @Override
    public void dispose() {
        if (!isClosed()) {
            WINDOW_LIST.remove(this);
        }
        if (!mIsClosed) {
            try {
                saveBounds();
                super.dispose();
            } catch (Exception ex) {
                // Necessary, since the AWT appears to sometimes spuriously try
                // to call Container.removeNotify() more than once on itself.
            }
            mIsClosed = true;
        }
    }

    @Override
    public void windowActivated(WindowEvent event) {
        // Unused
    }

    @Override
    public void windowClosed(WindowEvent event) {
        QuitCommand.INSTANCE.quitIfNoSignificantWindowsOpen();
    }

    @Override
    public void windowClosing(WindowEvent event) {
        if (!hasOwnedWindowsShowing(this)) {
            List<Saveable> saveables = new ArrayList<>();
            collectSaveables(this, saveables);
            if (SaveCommand.attemptSave(saveables)) {
                dispose();
            }
        }
    }

    private void collectSaveables(Component component, List<Saveable> saveables) {
        if (component instanceof Container) {
            Container container = (Container) component;
            int       count     = container.getComponentCount();
            for (int i = 0; i < count; i++) {
                collectSaveables(container.getComponent(i), saveables);
            }
        }
        if (component instanceof Saveable) {
            saveables.add((Saveable) component);
        }
    }

    @Override
    public void windowDeactivated(WindowEvent event) {
        Commitable.sendCommitToFocusOwner();
    }

    @Override
    public void windowDeiconified(WindowEvent event) {
        // Unused
    }

    @Override
    public void windowIconified(WindowEvent event) {
        // Unused
    }

    @Override
    public void windowOpened(WindowEvent event) {
        // On windows, this is necessary to prevent the window from opening in the background.
        toFront();
    }

    @Override
    public void windowGainedFocus(WindowEvent event) {
        if (event.getWindow() == this) {
            WINDOW_LIST.remove(this);
            WINDOW_LIST.add(0, this);
        }
    }

    @Override
    public void windowLostFocus(WindowEvent event) {
        Commitable.sendCommitToFocusOwner();
    }

    /**
     * Invalidates the specified component and all of its children, recursively.
     *
     * @param comp The root component to start with.
     */
    public void invalidate(Component comp) {
        comp.invalidate();
        comp.repaint();
        if (comp instanceof Container) {
            for (Component child : ((Container) comp).getComponents()) {
                invalidate(child);
            }
        }
    }

    /** @return The title to be used for this window in the window menu. */
    protected String getTitleForWindowMenu() {
        return getTitle();
    }

    @Override
    public int compareTo(BaseWindow other) {
        if (this != other) {
            String title      = getTitleForWindowMenu();
            String otherTitle = other.getTitleForWindowMenu();
            if (title != null) {
                if (otherTitle == null) {
                    return 1;
                }
                return title.compareTo(otherTitle);
            }
            if (otherTitle != null) {
                return -1;
            }
        }
        return 0;
    }

    /** @return The window's {@link StdUndoManager}. */
    @Override
    public StdUndoManager getUndoManager() {
        return mUndoManager;
    }

    /**
     * @return The prefix for keys in the {@link #WINDOW_PREFERENCES}module for this window. If
     *         {@code null} is returned from this method, then no standard window preferences will
     *         be saved. Returns {@code null} by default.
     */
    @SuppressWarnings("static-method")
    public String getWindowPrefsPrefix() {
        return null;
    }

    /**
     * @return The preference object to use for saving and restoring standard window preferences.
     *         Returns the result of calling {@link Preferences#getInstance()} by default.
     */
    @SuppressWarnings("static-method")
    public Preferences getWindowPreferences() {
        return Preferences.getInstance();
    }

    /**
     * Saves the window bounds to preferences. Preferences must be saved for this to have a lasting
     * effect.
     */
    public void saveBounds() {
        String keyPrefix = getWindowPrefsPrefix();
        if (keyPrefix != null) {
            Preferences prefs        = getWindowPreferences();
            boolean     wasMaximized = (getExtendedState() & MAXIMIZED_BOTH) != 0;
            if (wasMaximized || getExtendedState() == ICONIFIED) {
                setExtendedState(NORMAL);
            }
            prefs.startBatch();
            prefs.setValue(WINDOW_PREFERENCES, keyPrefix + KEY_LOCATION, getLocation());
            prefs.setValue(WINDOW_PREFERENCES, keyPrefix + KEY_SIZE, getSize());
            prefs.setValue(WINDOW_PREFERENCES, keyPrefix + KEY_LAST_UPDATED, System.currentTimeMillis());
            prefs.endBatch();
        }
    }

    /** Restores the window to its saved location and size. */
    public void restoreBounds() {
        Preferences prefs = getWindowPreferences();
        prefs.resetIfVersionMisMatch(WINDOW_PREFERENCES, WINDOW_PREFERENCES_VERSION);
        pruneOldWindowPreferences(prefs);
        boolean needPack  = true;
        String  keyPrefix = getWindowPrefsPrefix();
        if (keyPrefix != null) {
            Point location = prefs.getPointValue(WINDOW_PREFERENCES, keyPrefix + KEY_LOCATION);
            if (location != null) {
                setLocation(location);
            }
            Dimension size = prefs.getDimensionValue(WINDOW_PREFERENCES, keyPrefix + KEY_SIZE);
            if (size != null) {
                setSize(size);
                needPack = false;
            }
        }
        if (needPack) {
            pack();
        }
        WindowUtils.forceOnScreen(this);
    }

    private static void pruneOldWindowPreferences(Preferences prefs) {
        // 45 days ago, in milliseconds
        long         cutoff = System.currentTimeMillis() - 1000L * 60L * 60L * 24L * 45L;
        List<String> keys   = prefs.getModuleKeys(WINDOW_PREFERENCES);
        List<String> list   = new ArrayList<>();
        for (String key : keys) {
            if (key.endsWith(KEY_LAST_UPDATED)) {
                if (prefs.getLongValue(WINDOW_PREFERENCES, key, 0) < cutoff) {
                    list.add(key.substring(0, key.length() - KEY_LAST_UPDATED.length()));
                }
            }
        }
        if (!list.isEmpty()) {
            for (String key : keys) {
                for (String prefix : list) {
                    if (key.startsWith(prefix)) {
                        prefs.removePreference(WINDOW_PREFERENCES, key);
                        break;
                    }
                }
            }
        }
    }

    /** @return The top-most window. */
    public static BaseWindow getTopWindow() {
        if (!WINDOW_LIST.isEmpty()) {
            return WINDOW_LIST.get(0);
        }
        return null;
    }

    /**
     * @param <T>  The window type.
     * @param type The window type to return.
     * @return A list of all windows of the specified type.
     */
    public static <T extends BaseWindow> List<T> getWindows(Class<T> type) {
        List<T> windows = new ArrayList<>();
        Frame[] frames  = Frame.getFrames();
        for (Frame element : frames) {
            if (type.isInstance(element)) {
                T window = type.cast(element);
                if (window.mWasAlive && !window.isClosed()) {
                    windows.add(window);
                }
            }
        }
        return windows;
    }

    /**
     * @param windowClass The window class to return.
     * @param <T>         The window type.
     * @return The current visible windows, in order from top to bottom.
     */
    public static <T extends BaseWindow> ArrayList<T> getActiveWindows(Class<T> windowClass) {
        ArrayList<T> list = new ArrayList<>();
        for (T window : new FilteredIterator<>(WINDOW_LIST, windowClass)) {
            if (window.isShowing()) {
                list.add(window);
            }
        }
        return list;
    }

    /** @return A list of all {@link BaseWindow}s created by this application. */
    public static List<BaseWindow> getAllAppWindows() {
        return getWindows(BaseWindow.class);
    }

    /**
     * @param file The backing file to look for.
     * @return The {@link FileProxy} associated with the specified backing file.
     */
    public static FileProxy findFileProxy(File file) {
        String fullPath = PathUtils.getFullPath(file);
        for (BaseWindow window : getAllAppWindows()) {
            if (window instanceof FileProxy) {
                FileProxy proxy = (FileProxy) window;
                File      wFile = proxy.getBackingFile();
                if (wFile != null) {
                    if (PathUtils.getFullPath(wFile).equals(fullPath)) {
                        return proxy;
                    }
                }
            } else if (window instanceof FileProxyProvider) {
                FileProxy proxy = ((FileProxyProvider) window).getFileProxy(file);
                if (proxy != null) {
                    return proxy;
                }
            }
        }
        return null;
    }

    /**
     * @param window The window to check.
     * @return {@code true} if an owned window is showing.
     */
    public static boolean hasOwnedWindowsShowing(Window window) {
        if (window != null) {
            for (Window one : window.getOwnedWindows()) {
                if (one.isShowing() && one.getType() != Window.Type.POPUP) {
                    return true;
                }
            }
        }
        return false;
    }

    /** Forces a full repaint and invalidate on all windows, disposing of any window buffers. */
    public static void forceRepaintAndInvalidate() {
        for (BaseWindow window : getAllAppWindows()) {
            window.invalidate(window.getRootPane());
        }
    }
}
