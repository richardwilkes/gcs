/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.menu.file.CloseCommand;
import com.trollworks.gcs.menu.window.WindowMenu;
import com.trollworks.gcs.utility.StdUndoManager;
import com.trollworks.gcs.utility.collections.FilteredIterator;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.utility.io.print.PageOrientation;
import com.trollworks.gcs.utility.io.print.PrintManager;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.widgets.layout.FlexRow;

import java.awt.AWTEvent;
import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Frame;
import java.awt.GraphicsConfiguration;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Toolkit;
import java.awt.Window;
import java.awt.event.MouseWheelEvent;
import java.awt.event.WindowEvent;
import java.awt.event.WindowFocusListener;
import java.awt.event.WindowListener;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.List;
import java.util.WeakHashMap;

import javax.swing.JComponent;
import javax.swing.JFrame;
import javax.swing.JToolBar;
import javax.swing.WindowConstants;

/** Provides a base OS-level window. */
public class AppWindow extends JFrame implements Comparable<AppWindow>, WindowListener, WindowFocusListener, Undoable {
	private static final String								EMPTY						= "";										//$NON-NLS-1$
	/** Window preferences */
	public static final String								WINDOW_PREFERENCES			= "WindowPrefs";							//$NON-NLS-1$
	/** Window preferences version */
	public static final int									WINDOW_PREFERENCES_VERSION	= 3;
	/** Window preferences location key */
	public static final String								KEY_LOCATION				= "Location";								//$NON-NLS-1$
	/** Window preferences size key */
	public static final String								KEY_SIZE					= "Size";									//$NON-NLS-1$
	/** Window preferences maximized key */
	public static final String								KEY_MAXIMIZED				= "Maximized";								//$NON-NLS-1$
	/** Window preferences last updated key */
	public static final String								KEY_LAST_UPDATED			= "LastUpdated";							//$NON-NLS-1$
	private static final WeakHashMap<Preferences, Object>	WINDOW_PREFS_MAP			= new WeakHashMap<Preferences, Object>();
	private static BufferedImage							DEFAULT_WINDOW_ICON			= null;
	private static final ArrayList<AppWindow>				WINDOW_LIST					= new ArrayList<AppWindow>();
	private BufferedImage									mWindowIcon;
	private PrintManager									mPrintManager;
	private StdUndoManager									mUndoManager;
	private boolean											mInForeground;
	private boolean											mIsClosed;
	private BufferedImage									mTitleIcon;
	private boolean											mWasAlive;
	private boolean											mIsPrinting;

	/** @return The top-most window. */
	public static AppWindow getTopWindow() {
		if (!WINDOW_LIST.isEmpty()) {
			return WINDOW_LIST.get(0);
		}
		return null;
	}

	/**
	 * Creates a new {@link AppWindow}.
	 * 
	 * @param title The window title. May be <code>null</code>.
	 * @param largeIcon The 32x32 window icon. OK to pass in a 16x16 icon here.
	 * @param smallIcon The 16x16 window icon.
	 */
	public AppWindow(String title, BufferedImage largeIcon, BufferedImage smallIcon) {
		this(title, largeIcon, smallIcon, null, false);
	}

	/**
	 * Creates a new {@link AppWindow}.
	 * 
	 * @param title The title of the window.
	 * @param largeIcon The 32x32 window icon. OK to pass in a 16x16 icon here.
	 * @param smallIcon The 16x16 window icon.
	 * @param gc The graphics configuration to use.
	 * @param undecorated Whether to create an undecorated window, without menus.
	 */
	public AppWindow(String title, BufferedImage largeIcon, BufferedImage smallIcon, GraphicsConfiguration gc, boolean undecorated) {
		super(title, gc);
		if (undecorated) {
			setUndecorated(true);
		}
		setDefaultCloseOperation(WindowConstants.DO_NOTHING_ON_CLOSE);
		setLocationByPlatform(true);
		if (!undecorated) {
			setJMenuBar(new StdMenuBar());
		}
		((JComponent) getContentPane()).setDoubleBuffered(true);

		if (largeIcon == null) {
			largeIcon = DEFAULT_WINDOW_ICON;
		}
		if (largeIcon != null) {
			setTitleIcon(largeIcon);
		}

		if (smallIcon == null) {
			smallIcon = largeIcon;
		}
		if (smallIcon != null) {
			setMenuIcon(smallIcon);
		}

		mUndoManager = new StdUndoManager();

		enableEvents(AWTEvent.KEY_EVENT_MASK | AWTEvent.MOUSE_EVENT_MASK | AWTEvent.COMPONENT_EVENT_MASK);
		addWindowListener(this);
		addWindowFocusListener(this);
		new WindowSizeEnforcer(this);

		Toolkit.getDefaultToolkit().setDynamicLayout(true);

		WINDOW_LIST.add(this);
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
	 * @param row The {@link FlexRow} layout to add items to.
	 */
	protected void createToolBarContents(JToolBar toolbar, FlexRow row) {
		// Does nothing by default.
	}

	/** @return The default window icon. */
	public static BufferedImage getDefaultWindowIcon() {
		return DEFAULT_WINDOW_ICON;
	}

	/** @param image The new default window icon. */
	public static void setDefaultWindowIcon(BufferedImage image) {
		DEFAULT_WINDOW_ICON = image;
	}

	/** @param icon The icon representing this window. */
	public void setTitleIcon(BufferedImage icon) {
		if (icon != mTitleIcon) {
			mTitleIcon = icon;
			setIconImage(mTitleIcon);
		}
	}

	/** @return The icon representing this window. */
	public BufferedImage getTitleIcon() {
		return mTitleIcon;
	}

	/** @return The menu icon representing this window. */
	public BufferedImage getMenuIcon() {
		return mWindowIcon;
	}

	/** @param icon The menu icon representing this window. */
	public void setMenuIcon(BufferedImage icon) {
		mWindowIcon = icon;
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

	@Override public void setVisible(boolean visible) {
		if (visible) {
			mWasAlive = true;
		}
		super.setVisible(visible);
		if (visible) {
			WindowMenu.update();
		}
	}

	@Override public void setTitle(String title) {
		super.setTitle(DataMenu.filterTitle(title));
		if (isVisible()) {
			WindowMenu.update();
		}
	}

	@Override public void toFront() {
		if (!isClosed()) {
			Component focus;

			if (getExtendedState() == ICONIFIED) {
				setExtendedState(NORMAL);
			}
			super.toFront();
			if (!isActive() || !isFocused()) {
				focus = getMostRecentFocusOwner();
				if (focus != null) {
					focus.requestFocus();
				} else {
					requestFocus();
				}
			}
		}
	}

	/** @return The {@link PrintManager} for this window. */
	public final PrintManager getPrintManager() {
		if (mPrintManager == null) {
			mPrintManager = createPageSettings();
		}
		return mPrintManager;
	}

	/** @param printManager The {@link PrintManager} to use. */
	public void setPrintManager(PrintManager printManager) {
		mPrintManager = printManager;
	}

	/** @return The default page settings for this window. May return <code>null</code>. */
	protected PrintManager createPageSettings() {
		try {
			return new PrintManager(PageOrientation.PORTRAIT, 0.5, LengthUnits.INCHES);
		} catch (Exception exception) {
			return null;
		}
	}

	/** Called after the page setup has changed. */
	public void adjustToPageSetupChanges() {
		// Does nothing by default.
	}

	/** @return <code>true</code> if the window is currently printing. */
	public boolean isPrinting() {
		return mIsPrinting;
	}

	/** @param printing <code>true</code> if the window is currently printing. */
	public void setPrinting(boolean printing) {
		mIsPrinting = printing;
	}

	/**
	 * @param event
	 */
	protected void processMouseWheelEventSuper(MouseWheelEvent event) {
		super.processMouseWheelEvent(event);
	}

	/**
	 * @param window The window to check.
	 * @return <code>true</code> if an owned window is showing.
	 */
	public static boolean hasOwnedWindowsShowing(Window window) {
		for (Window one : window.getOwnedWindows()) {
			if (one.isShowing()) {
				return true;
			}
		}
		return false;
	}

	/** @return <code>true</code> if the window has been closed. */
	public boolean isClosed() {
		return mIsClosed;
	}

	@Override public void dispose() {
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
		WindowMenu.update();
	}

	/**
	 * @param windowClass The window class to return.
	 * @param <T> The window type.
	 * @return The current visible windows, in order from top to bottom.
	 */
	public static <T extends AppWindow> ArrayList<T> getActiveWindows(Class<T> windowClass) {
		ArrayList<T> list = new ArrayList<T>();

		for (T window : new FilteredIterator<T>(WINDOW_LIST, windowClass)) {
			if (window.isShowing()) {
				list.add(window);
			}
		}
		return list;
	}

	/** @return <code>true</code> if the window is in the foreground. */
	public boolean isInForeground() {
		return mInForeground;
	}

	private void setInForeground(boolean inForeground) {
		if (inForeground != mInForeground) {
			mInForeground = inForeground;
			foregroundStateChanged(inForeground);
		}
	}

	/**
	 * Called when the window changes its foreground state.
	 * 
	 * @param inForeground <code>true</code> will be passed in when the window is in the
	 *            foreground. <code>false</code> will be passed in when the window is not in the
	 *            foreground.
	 */
	protected void foregroundStateChanged(boolean inForeground) {
		// Nothing to do...
	}

	public void windowGainedFocus(WindowEvent event) {
		if (event.getWindow() == this) {
			WINDOW_LIST.remove(this);
			WINDOW_LIST.add(0, this);
			setInForeground(true);
		}
	}

	public void windowLostFocus(WindowEvent event) {
		if (event.getWindow() == this) {
			setInForeground(false);
		}
	}

	public void windowActivated(WindowEvent event) {
		// Nothing to do...
	}

	public void windowClosed(WindowEvent event) {
		// Nothing to do...
	}

	public void windowClosing(WindowEvent event) {
		CloseCommand.INSTANCE.close(this, true);
	}

	public void windowDeactivated(WindowEvent event) {
		// Nothing to do...
	}

	public void windowDeiconified(WindowEvent event) {
		// Nothing to do...
	}

	public void windowIconified(WindowEvent event) {
		// Nothing to do...
	}

	public void windowOpened(WindowEvent event) {
		// Nothing to do...
	}

	/** @return The window's {@link StdUndoManager}. */
	public StdUndoManager getUndoManager() {
		return mUndoManager;
	}

	/** @return A list of all {@link AppWindow}s created by this application. */
	public static ArrayList<AppWindow> getAllWindows() {
		return getWindows(AppWindow.class);
	}

	/**
	 * @param <T> The window type.
	 * @param type The window type to return.
	 * @return A list of all windows of the specified type.
	 */
	public static <T extends AppWindow> ArrayList<T> getWindows(Class<T> type) {
		ArrayList<T> windows = new ArrayList<T>();
		Frame[] frames = Frame.getFrames();

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

	/** @return <code>true</code> if this window is one of a group of windows. */
	protected boolean isPartOfWindowGroup() {
		return false;
	}

	/** @return <code>true</code> if this window is the parent of a group of windows. */
	protected boolean isWindowGroupParent() {
		return false;
	}

	/** @return The window group name this window belongs to. */
	protected String getWindowGroupName() {
		return null;
	}

	/** @return The compare name for the window group this window belongs to. */
	protected String getWindowGroupCompareName() {
		return null;
	}

	/** @return The window group icon this window belongs to. */
	protected BufferedImage getWindowGroupIcon() {
		return null;
	}

	/** @return The title to be used for this window in the window menu. */
	protected String getTitleForWindowMenu() {
		return getTitle();
	}

	public int compareTo(AppWindow other) {
		if (other == this) {
			return 0;
		}

		String group = getWindowGroupCompareName();
		String otherGroup = other.getWindowGroupCompareName();
		int result;

		if (group == null) {
			group = EMPTY;
		}
		if (otherGroup == null) {
			otherGroup = EMPTY;
		}
		result = group.compareTo(otherGroup);
		if (result == 0) {
			if (isWindowGroupParent()) {
				result = -1;
			} else if (other.isWindowGroupParent()) {
				result = 1;
			} else {
				String title = getTitleForWindowMenu();
				String otherTitle = other.getTitleForWindowMenu();

				if (title == null) {
					title = EMPTY;
				}
				if (otherTitle == null) {
					otherTitle = EMPTY;
				}
				result = title.compareTo(otherTitle);
			}
		}
		return result;
	}

	/**
	 * @return The prefix for keys in the {@link #WINDOW_PREFERENCES}module for this window. If
	 *         <code>null</code> is returned from this method, then no standard window preferences
	 *         will be saved. Returns <code>null</code> by default.
	 */
	public String getWindowPrefsPrefix() {
		return null;
	}

	/**
	 * @return The preference object to use for saving and restoring standard window preferences.
	 *         Returns the result of calling {@link Preferences#getInstance()} by default.
	 */
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
			Preferences prefs = getWindowPreferences();
			boolean wasMaximized = (getExtendedState() & MAXIMIZED_BOTH) != 0;

			if (wasMaximized || getExtendedState() == ICONIFIED) {
				setExtendedState(NORMAL);
			}

			prefs.startBatch();
			prefs.setValue(WINDOW_PREFERENCES, keyPrefix + KEY_LOCATION, getLocation());
			prefs.setValue(WINDOW_PREFERENCES, keyPrefix + KEY_SIZE, getSize());
			prefs.setValue(WINDOW_PREFERENCES, keyPrefix + KEY_MAXIMIZED, wasMaximized);
			prefs.setValue(WINDOW_PREFERENCES, keyPrefix + KEY_LAST_UPDATED, System.currentTimeMillis());
			prefs.endBatch();
		}
	}

	/** Restores the window to its saved location and size. */
	public void restoreBounds() {
		Preferences prefs = getWindowPreferences();
		prefs.resetIfVersionMisMatch(WINDOW_PREFERENCES, WINDOW_PREFERENCES_VERSION);
		pruneOldWindowPreferences(prefs);

		boolean needPack = true;
		String keyPrefix = getWindowPrefsPrefix();
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
		GraphicsUtilities.forceOnScreen(this);

		if (prefs.getBooleanValue(WINDOW_PREFERENCES, keyPrefix + KEY_MAXIMIZED, false)) {
			setExtendedState(MAXIMIZED_BOTH);
		}
	}

	private void pruneOldWindowPreferences(Preferences prefs) {
		if (!WINDOW_PREFS_MAP.containsKey(prefs)) {
			List<String> keys = prefs.getModuleKeys(WINDOW_PREFERENCES);
			long cutoff = System.currentTimeMillis() - 3888000000L; // 45 days ago, in milliseconds
			ArrayList<String> list = new ArrayList<String>();

			for (String key : keys) {
				if (key.endsWith(KEY_LAST_UPDATED)) {
					if (prefs.getLongValue(WINDOW_PREFERENCES, key, 0) < cutoff) {
						list.add(key.substring(0, key.length() - KEY_LAST_UPDATED.length()));
					}
				}
			}

			for (String key : keys) {
				for (String prefix : list) {
					if (key.startsWith(prefix)) {
						prefs.removePreference(WINDOW_PREFERENCES, key);
						break;
					}
				}
			}

			WINDOW_PREFS_MAP.put(prefs, null);
		}
	}

	/**
	 * Creates a new, untitled window title.
	 * 
	 * @param <T> The window type.
	 * @param windowClass The window class to use for name comparisons.
	 * @param baseTitle The base untitled name.
	 * @param exclude A window to exclude from naming decisions. May be <code>null</code>.
	 * @return The new window title.
	 */
	public static <T extends AppWindow> String getNextUntitledWindowName(Class<T> windowClass, String baseTitle, AppWindow exclude) {
		List<T> windows = getWindows(windowClass);
		int value = 0;
		String title;

		do {
			title = baseTitle;
			if (++value > 1) {
				title += " " + value; //$NON-NLS-1$
			}
			for (T window : windows) {
				if (window != exclude && title.equals(window.getTitle())) {
					title = null;
					break;
				}
			}
		} while (title == null);
		return title;
	}
}
