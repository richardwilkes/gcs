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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.menu.file.QuitCommand;
import com.trollworks.gcs.menu.file.SaveCommand;
import com.trollworks.gcs.menu.file.Saveable;
import com.trollworks.gcs.preferences.MenuKeyPreferences;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.WindowSizeEnforcer;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.awt.AWTEvent;
import java.awt.Component;
import java.awt.Container;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.Window;
import java.awt.event.WindowEvent;
import java.awt.event.WindowFocusListener;
import java.awt.event.WindowListener;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.JFrame;
import javax.swing.WindowConstants;

/** Provides a base OS-level window. */
public class BaseWindow extends JFrame implements Undoable, Comparable<BaseWindow>, WindowListener, WindowFocusListener {
    private static final ArrayList<BaseWindow> WINDOW_LIST = new ArrayList<>();
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

    @Override
    public void toFront() {
        if (!mIsClosed) {
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

    @Override
    public void dispose() {
        if (!mIsClosed) {
            WINDOW_LIST.remove(this);
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
     * @return The key for the window preferences. If {@code null} is returned from this method,
     *         then no standard window preferences will be saved. Returns {@code null} by default.
     */
    @SuppressWarnings("static-method")
    public String getWindowPrefsKey() {
        return null;
    }

    /**
     * Saves the window bounds to preferences. Preferences must be saved for this to have a lasting
     * effect.
     */
    public void saveBounds() {
        String key = getWindowPrefsKey();
        if (key != null) {
            boolean wasMaximized = (getExtendedState() & MAXIMIZED_BOTH) != 0;
            if (wasMaximized || getExtendedState() == ICONIFIED) {
                setExtendedState(NORMAL);
            }
            Preferences.getInstance().putBaseWindowPosition(key, new Position(this));
        }
    }

    /** Restores the window to its saved location and size. */
    public void restoreBounds() {
        boolean needPack = true;
        String  key      = getWindowPrefsKey();
        if (key != null) {
            Position info = Preferences.getInstance().getBaseWindowPosition(key);
            if (info != null) {
                info.apply(this);
                needPack = false;
            }
        }
        if (needPack) {
            pack();
        }
        WindowUtils.forceOnScreen(this);
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
        for (Frame frame : frames) {
            if (type.isInstance(frame)) {
                T window = type.cast(frame);
                if (window.mWasAlive && !window.isClosed()) {
                    windows.add(window);
                }
            }
        }
        return windows;
    }

    /** @return A list of all {@link BaseWindow}s created by this application. */
    public static List<BaseWindow> getAllAppWindows() {
        return getWindows(BaseWindow.class);
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

    @Override
    public void paint(Graphics g) {
        // This is just here to try and capture a log message if something goes wrong in the
        // drawing code.
        try {
            super.paint(g);
        } catch(Throwable throwable) {
            Log.error(throwable);
        }
    }

    /** Forces a full repaint and invalidate on all windows, disposing of any window buffers. */
    public static void forceRepaintAndInvalidate() {
        for (BaseWindow window : getAllAppWindows()) {
            window.invalidate(window.getRootPane());
        }
    }

    public static class Position {
        private static final String    X            = "x";
        private static final String    Y            = "y";
        private static final String    WIDTH        = "width";
        private static final String    HEIGHT       = "height";
        private static final String    LAST_UPDATED = "last_updated";
        public               Rectangle mBounds;
        public               long      mLastUpdated;

        public Position(BaseWindow wnd) {
            mBounds = wnd.getBounds();
            mLastUpdated = System.currentTimeMillis();
        }

        public Position(JsonMap m) {
            mBounds = new Rectangle(m.getIntWithDefault(X, 0), m.getIntWithDefault(Y, 0), m.getIntWithDefault(WIDTH, 1), m.getIntWithDefault(HEIGHT, 1));
            mLastUpdated = m.getLong(LAST_UPDATED);
        }

        void apply(BaseWindow wnd) {
            wnd.setBounds(mBounds);
        }

        public void toJSON(JsonWriter w) throws IOException {
            w.startMap();
            w.keyValue(X, mBounds.x);
            w.keyValue(Y, mBounds.y);
            w.keyValue(WIDTH, mBounds.width);
            w.keyValue(HEIGHT, mBounds.height);
            w.keyValue(LAST_UPDATED, mLastUpdated);
            w.endMap();
        }
    }
}
