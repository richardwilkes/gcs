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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.print.TKPageOrientation;
import com.trollworks.toolkit.print.TKPrintManager;
import com.trollworks.toolkit.qa.TKQAMenu;
import com.trollworks.toolkit.undo.TKUndoManager;
import com.trollworks.toolkit.utility.TKAppWithUI;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.units.TKLengthUnits;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKProgressBar;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.TKToolBarTarget;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKWindowLayout;
import com.trollworks.toolkit.widget.menu.TKBaseMenu;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuBar;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuSeparator;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;

import java.awt.AWTEvent;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsConfiguration;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Toolkit;
import java.awt.Window;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;
import java.awt.event.MouseWheelEvent;
import java.awt.event.WindowEvent;
import java.awt.event.WindowFocusListener;
import java.awt.event.WindowListener;
import java.awt.image.BufferedImage;
import java.awt.print.Printable;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.WeakHashMap;

/** Provides a base OS-level window. */
public class TKWindow extends Frame implements Comparable<TKWindow>, TKBaseWindow, TKMenuTarget, TKToolBarTarget, WindowListener, WindowFocusListener, ComponentListener {
	private static final String								EMPTY							= "";										//$NON-NLS-1$
	/** The standard X stagger amount. */
	public static final int									STAGGER_X						= 15;
	/** The standard Y stagger amount. */
	public static final int									STAGGER_Y						= 30;
	/** The standard open command. */
	public static final String								CMD_OPEN						= "Open";									//$NON-NLS-1$
	/** The standard attempt close command. */
	public static final String								CMD_ATTEMPT_CLOSE				= "AttemptClose";							//$NON-NLS-1$
	/** The standard page setup command. */
	public static final String								CMD_PAGE_SETUP					= "PageSetup";								//$NON-NLS-1$
	/** The standard print command. */
	public static final String								CMD_PRINT						= "Print";									//$NON-NLS-1$
	/** The standard about command. */
	public static final String								CMD_ABOUT						= "About";									//$NON-NLS-1$
	/** The standard quit command. */
	public static final String								CMD_QUIT						= "Quit";									//$NON-NLS-1$
	/** The standard revert command. */
	public static final String								CMD_REVERT						= "Revert";								//$NON-NLS-1$
	/** The standard save command. */
	public static final String								CMD_SAVE						= "Save";									//$NON-NLS-1$
	/** The standard save as command. */
	public static final String								CMD_SAVE_AS						= "SaveAs";								//$NON-NLS-1$
	/** The standard save copy as command. */
	public static final String								CMD_SAVE_COPY_AS				= "SaveCopyAs";							//$NON-NLS-1$
	/** The standard undo command. */
	public static final String								CMD_UNDO						= "Undo";									//$NON-NLS-1$
	/** The standard redo command. */
	public static final String								CMD_REDO						= "Redo";									//$NON-NLS-1$
	/** The standard cut command. */
	public static final String								CMD_CUT							= "Cut";									//$NON-NLS-1$
	/** The standard copy command. */
	public static final String								CMD_COPY						= "Copy";									//$NON-NLS-1$
	/** The standard paste command. */
	public static final String								CMD_PASTE						= "Paste";									//$NON-NLS-1$
	/** The standard clear command. */
	public static final String								CMD_CLEAR						= "Clear";									//$NON-NLS-1$
	/** The standard select all command. */
	public static final String								CMD_SELECT_ALL					= "SelectAll";								//$NON-NLS-1$
	/** The standard duplicate command. */
	public static final String								CMD_DUPLICATE					= "Duplicate";								//$NON-NLS-1$
	/** The standard reset dialog confirmations command. */
	public static final String								CMD_RESET_CONFIRMATION_DIALOGS	= "ResetConfirmationDialogs";				//$NON-NLS-1$
	/** The standard preferences command. */
	public static final String								CMD_PREFERENCES					= "Preferences";							//$NON-NLS-1$
	/** The standard window command. */
	public static final String								CMD_WINDOW						= "Window";								//$NON-NLS-1$
	/** Window preferences */
	public static final String								WINDOW_PREFERENCES				= "WindowPrefs";							//$NON-NLS-1$
	/** Window preferences version */
	public static final int									WINDOW_PREFERENCES_VERSION		= 3;
	/** Window preferences location key */
	public static final String								KEY_LOCATION					= "Location";								//$NON-NLS-1$
	/** Window preferences size key */
	public static final String								KEY_SIZE						= "Size";									//$NON-NLS-1$
	/** Window preferences maximized key */
	public static final String								KEY_MAXIMIZED					= "Maximized";								//$NON-NLS-1$
	/** Window preferences last updated key */
	public static final String								KEY_LAST_UPDATED				= "LastUpdated";							//$NON-NLS-1$
	/** The key used for the standard file menu. */
	public static final String								KEY_STD_MENUBAR_FILE_MENU		= "StdMenuBar.File";						//$NON-NLS-1$
	/** The key used for the standard edit menu. */
	public static final String								KEY_STD_MENUBAR_EDIT_MENU		= "StdMenuBar.Edit";						//$NON-NLS-1$
	/** The key used for the standard window menu. */
	public static final String								KEY_STD_MENUBAR_WINDOW_MENU		= "StdMenuBar.Window";						//$NON-NLS-1$
	/** The key used for the standard help menu. */
	public static final String								KEY_STD_MENUBAR_HELP_MENU		= "StdMenuBar.Help";						//$NON-NLS-1$
	private static final WeakHashMap<TKPreferences, Object>	WINDOW_PREFS_MAP				= new WeakHashMap<TKPreferences, Object>();
	private static BufferedImage							DEFAULT_WINDOW_ICON				= null;
	private static final ArrayList<TKWindow>				WINDOW_LIST						= new ArrayList<TKWindow>();
	private static final HashSet<TKBaseWindowMonitor>		WINDOW_MONITORS					= new HashSet<TKBaseWindowMonitor>();
	private TKPrintManager									mPrintManager;
	private TKPanel											mUserContent;
	private TKRepaintManager								mRepaintManager;
	private TKUndoManager									mUndoManager;
	private TKUserInputManager								mUserInputManager;
	private boolean											mInForeground;
	private boolean											mIsClosed;
	private ArrayList<TKSelectionListener>					mSelectionListeners;
	private TKMenuBar										mMenuBar;
	private TKToolBar										mToolBar;
	private BufferedImage									mTitleIcon;
	private boolean											mWasAlive;

	/** @return The top-most window. */
	public static TKWindow getTopWindow() {
		if (!WINDOW_LIST.isEmpty()) {
			return WINDOW_LIST.get(0);
		}
		return null;
	}

	/**
	 * Creates a new window.
	 * 
	 * @param title The title of the window.
	 * @param icon The icon to use for the window.
	 */
	public TKWindow(String title, BufferedImage icon) {
		this(title, icon, null);
	}

	/**
	 * Creates a new window.
	 * 
	 * @param title The title of the window.
	 * @param icon The icon to use for the window.
	 * @param gc The graphics configuration to use.
	 */
	public TKWindow(String title, BufferedImage icon, GraphicsConfiguration gc) {
		super(title, gc);

		if (icon == null) {
			icon = DEFAULT_WINDOW_ICON;
		}
		if (icon != null) {
			setTitleIcon(icon);
		}

		mRepaintManager = new TKRepaintManager(this);
		mUndoManager = new TKUndoManager();
		mUserInputManager = new TKUserInputManager(this);

		TKPanel content = new TKPanel(new TKCompassLayout());
		content.setOpaque(true);
		setContent(content);
		setLayout(TKWindowLayout.getInstance());

		enableEvents(AWTEvent.KEY_EVENT_MASK | AWTEvent.MOUSE_EVENT_MASK | AWTEvent.COMPONENT_EVENT_MASK);
		addWindowListener(this);
		addWindowFocusListener(this);
		mIsClosed = false;

		setLocation(getDefaultLocation(this));
		Toolkit.getDefaultToolkit().setDynamicLayout(true);

		WINDOW_LIST.add(this);
	}

	/** @return The default window icon. */
	public static BufferedImage getDefaultWindowIcon() {
		return DEFAULT_WINDOW_ICON;
	}

	/** @param image The new default window icon. */
	public static void setDefaultWindowIcon(BufferedImage image) {
		DEFAULT_WINDOW_ICON = image;
	}

	/**
	 * @param excludeWindow A window to ignore, typically the one being placed.
	 * @return The default location to place a window.
	 */
	public static Point getDefaultLocation(TKWindow excludeWindow) {
		List<TKWindow> windows = getAllWindows();
		Rectangle bounds = TKGraphics.getMaximumWindowBounds();
		int x = bounds.x;
		int y = bounds.y;
		int xcount = 0;
		int ycount = 0;
		boolean again = true;

		while (again) {
			again = false;
			for (TKWindow window : windows) {
				if (window != excludeWindow) {
					Point where = window.getLocation();

					if (where.x == x && where.y == y) {
						again = true;
						x += STAGGER_X;
						y += STAGGER_Y;
						if (x >= bounds.x + bounds.width - 50) {
							x = bounds.x + ++xcount;
						}
						if (y >= bounds.y + bounds.height - 50) {
							y = bounds.x + ++ycount;
						}
						break;
					}
				}
			}
		}

		return new Point(x, y);
	}

	public void menusWillBeAdjusted() {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();

		if (focus instanceof TKMenuTarget && focus != this) {
			((TKMenuTarget) focus).menusWillBeAdjusted();
		}
	}

	public void menusWereAdjusted() {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();

		if (focus instanceof TKMenuTarget && focus != this) {
			((TKMenuTarget) focus).menusWereAdjusted();
		}
	}

	public boolean adjustMenuItem(String command, TKMenuItem item) {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();
		boolean processed = false;

		if (focus != this && focus instanceof TKMenuTarget) {
			processed = ((TKMenuTarget) focus).adjustMenuItem(command, item);
		}

		if (!processed) {
			processed = true;
			if (CMD_ABOUT.equals(command) || CMD_OPEN.equals(command) || CMD_ATTEMPT_CLOSE.equals(command) || CMD_QUIT.equals(command)) {
				item.setEnabled(true);
			} else if (CMD_PAGE_SETUP.equals(command) || CMD_PRINT.equals(command)) {
				item.setEnabled((this instanceof Printable) && getPrintManager() != null);
			} else if (CMD_REDO.equals(command)) {
				item.setEnabled(getUndoManager().canApply(false));
				item.setTitle(getUndoManager().getName(false));
			} else if (CMD_UNDO.equals(command)) {
				item.setEnabled(getUndoManager().canApply(true));
				item.setTitle(getUndoManager().getName(true));
			} else if (CMD_WINDOW.equals(command)) {
				item.setEnabled(true);
				item.setMarked(item.getUserObject() == this);
			} else if (CMD_RESET_CONFIRMATION_DIALOGS.equals(command)) {
				item.setEnabled(TKOptionDialog.isResetNeeded());
			} else {
				TKQAMenu qaMenu = getTKMenuBar().getQAMenu();

				if (qaMenu != null) {
					processed = qaMenu.adjustMenuItem(command, item);
				} else {
					processed = false;
				}
			}
		}
		return processed;
	}

	public boolean obeyCommand(String command, TKMenuItem item) {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();
		boolean processed = false;

		if (focus != this && focus instanceof TKMenuTarget) {
			processed = ((TKMenuTarget) focus).obeyCommand(command, item);
		}

		if (!processed) {
			processed = true;
			if (CMD_ABOUT.equals(command)) {
				TKAppWithUI.getInstanceWithUI().handleAbout(null);
			} else if (CMD_OPEN.equals(command)) {
				open();
			} else if (CMD_ATTEMPT_CLOSE.equals(command)) {
				attemptClose();
			} else if (CMD_QUIT.equals(command)) {
				TKAppWithUI.getInstanceWithUI().handleQuit(null);
			} else if (CMD_WINDOW.equals(command)) {
				((TKWindow) item.getUserObject()).toFront();
			} else if (CMD_RESET_CONFIRMATION_DIALOGS.equals(command)) {
				TKOptionDialog.resetAllDialogs();
			} else if (CMD_PAGE_SETUP.equals(command)) {
				pageSetup();
			} else if (CMD_PRINT.equals(command)) {
				print();
			} else if (CMD_REDO.equals(command)) {
				TKUndoManager rMgr = getUndoManager();

				if (rMgr.canApply(false)) {
					try {
						startProgress(0);
						rMgr.apply(false);
						TKCorrectableManager.getInstance().clearAllCorrectables();
					} finally {
						finishProgress();
					}
				}
			} else if (CMD_UNDO.equals(command)) {
				TKUndoManager uMgr = getUndoManager();

				if (uMgr.canApply(true)) {
					try {
						startProgress(0);
						uMgr.apply(true);
						TKCorrectableManager.getInstance().clearAllCorrectables();
					} finally {
						finishProgress();
					}
				}
			} else if (CMD_WINDOW.equals(command)) {
				((TKWindow) item.getUserObject()).toFront();
			} else {
				TKQAMenu qaMenu = getTKMenuBar().getQAMenu();

				if (qaMenu != null) {
					processed = qaMenu.obeyCommand(command, item);
				} else {
					processed = false;
				}
			}
		}
		return processed;
	}

	/**
	 * @param filters The filters to choose from.
	 * @return The preferred file filter to use from the array of filters. By default, just returns
	 *         the first one.
	 */
	public TKFileFilter getPreferredFileFilter(TKFileFilter[] filters) {
		return filters != null && filters.length > 0 ? filters[0] : null;
	}

	/** Brings up the open file dialog and allows the user to choose a file to open. */
	public void open() {
		open(false);
	}

	/**
	 * Brings up the open file dialog and allows the user to choose a file to open.
	 * 
	 * @param onEventQueue Whether the opening should occur on the event queue.
	 */
	public void open(boolean onEventQueue) {
		TKFileDialog fileDialog = new TKFileDialog(this, true);
		TKFileFilter[] filters = TKOpenManager.getFileFilters();

		for (TKFileFilter element : filters) {
			fileDialog.addFileFilter(element);
		}
		fileDialog.setActiveFileFilter(getPreferredFileFilter(filters));

		if (fileDialog.doModal() == TKDialog.OK) {
			TKOpenManager.openFiles(this, fileDialog.getSelectedItems(), onEventQueue);
		}
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

	/** Force a repaint of this window. */
	public void forceRepaint() {
		mRepaintManager.forceRepaint();
	}

	/** Force a repaint and invalidate of this window. */
	public void forceRepaintAndInvalidate() {
		invalidate(this);
		mRepaintManager.forceRepaint();
	}

	/**
	 * Invalidates the specified component and all of its children, recursively.
	 * 
	 * @param comp The root component to start with.
	 */
	public void invalidate(Component comp) {
		comp.invalidate();
		if (comp instanceof Container) {
			for (Component child : ((Container) comp).getComponents()) {
				invalidate(child);
			}
		}
	}

	public Rectangle getLocalBounds(boolean insets) {
		Rectangle bounds = getBounds();

		bounds.x = 0;
		bounds.y = 0;
		if (insets) {
			Insets theInsets = getInsets();

			bounds.x += theInsets.left;
			bounds.y += theInsets.top;
			bounds.width -= theInsets.left + theInsets.right;
			bounds.height -= theInsets.top + theInsets.bottom;
		}
		return bounds;
	}

	public TKPanel getContent() {
		return mUserContent;
	}

	public void setContent(TKPanel panel) {
		if (mUserContent != null) {
			mUserContent.removeFromParent();
			mUserContent.removeComponentListener(this);
		}
		mUserContent = panel;
		mUserContent.addComponentListener(this);
		add(panel);
	}

	@Override public void setVisible(boolean visible) {
		if (visible) {
			mWasAlive = true;
		}
		super.setVisible(visible);
		if (visible) {
			adjustWindowMenus();
		}
	}

	@Override public void setTitle(String title) {
		super.setTitle(title);
		adjustWindowMenus();
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

	@Override public void paint(Graphics graphics) {
		mRepaintManager.paint(graphics);
	}

	/**
	 * Paints the specified area within the window.
	 * 
	 * @param g2d The graphics context to use.
	 * @param clips The areas to draw.
	 */
	public void paint(Graphics2D g2d, Rectangle[] clips) {
		mRepaintManager.paint(g2d, clips);
	}

	/**
	 * Paints the specified area within the window without waiting for any lazy updates that may be
	 * pending.
	 * 
	 * @param bounds The area to draw.
	 */
	public void paintImmediately(Rectangle bounds) {
		mRepaintManager.paintImmediately(bounds);
	}

	/**
	 * Paints the specified area within the window without waiting for any lazy updates that may be
	 * pending.
	 * 
	 * @param bounds The area to draw.
	 */
	public void paintImmediately(Rectangle[] bounds) {
		mRepaintManager.paintImmediately(bounds);
	}

	/** @return The {@link TKPrintManager} for this window. */
	public final TKPrintManager getPrintManager() {
		if (mPrintManager == null) {
			mPrintManager = createPageSettings();
		}
		return mPrintManager;
	}

	/** @param printManager The {@link TKPrintManager} to use. */
	public void setPrintManager(TKPrintManager printManager) {
		mPrintManager = printManager;
	}

	/** @return The default page settings for this window. May return <code>null</code>. */
	protected TKPrintManager createPageSettings() {
		try {
			return new TKPrintManager(TKPageOrientation.PORTRAIT, 0.5, TKLengthUnits.INCHES);
		} catch (Exception exception) {
			return null;
		}
	}

	/**
	 * If the window implements the {@link Printable} interface, present the page setup dialog and
	 * allow the user to change the settings.
	 */
	public void pageSetup() {
		if (this instanceof Printable) {
			TKPrintManager mgr = getPrintManager();
			if (mgr != null) {
				mgr.pageSetup(this);
			} else {
				TKOptionDialog.error(this, Msgs.NO_PRINTER_SELECTED);
			}
		}
	}

	/** Called after the page setup has changed. */
	public void adjustToPageSetupChanges() {
		// Does nothing by default.
	}

	/**
	 * If the window implements the {@link Printable} interface, present the print dialog and allow
	 * the user to print the window's contents.
	 */
	public void print() {
		if (this instanceof Printable) {
			TKPrintManager mgr = getPrintManager();
			if (mgr != null) {
				mgr.print(this, getTitle(), (Printable) this);
			} else {
				TKOptionDialog.error(this, Msgs.NO_PRINTER_SELECTED);
			}
		}
	}

	public boolean isPrinting() {
		return mRepaintManager.isPrinting();
	}

	@Override public void repaint(long maxDelay, int x, int y, int width, int height) {
		mRepaintManager.repaint(x, y, width, height);
	}

	/**
	 * Marks the specified area within the window as needing to be painted.
	 * 
	 * @param bounds The area to redraw.
	 */
	public void repaint(Rectangle bounds) {
		mRepaintManager.repaint(bounds);
	}

	/**
	 * Marks the specified area within the window as no longer needing to be painted.
	 * 
	 * @param bounds The area to redraw.
	 */
	public void stopRepaint(Rectangle bounds) {
		mRepaintManager.stopRepaint(bounds);
	}

	/**
	 * Marks the specified area within the window as no longer needing to be painted.
	 * 
	 * @param bounds The area to redraw.
	 */
	public void stopRepaint(Rectangle[] bounds) {
		mRepaintManager.stopRepaint(bounds);
	}

	@Override public void update(Graphics graphics) {
		mRepaintManager.paint(graphics);
	}

	/** Forces all areas that need repainting to be painted. */
	public void updateImmediately() {
		mRepaintManager.updateImmediately();
	}

	public void processWindowKeyEvent(KeyEvent event) {
		// Not used by default.
	}

	@Override protected void processMouseEvent(MouseEvent event) {
		mUserInputManager.processMouseEvent(event);
	}

	@Override protected void processMouseMotionEvent(MouseEvent event) {
		mUserInputManager.processMouseEvent(event);
	}

	public void processMouseEventSuper(MouseEvent event) {
		super.processMouseEvent(event);
	}

	public void processMouseMotionEventSuper(MouseEvent event) {
		super.processMouseMotionEvent(event);
	}

	@Override protected void processMouseWheelEvent(MouseWheelEvent event) {
		mUserInputManager.processMouseEvent(event);
	}

	/**
	 * @param event
	 */
	protected void processMouseWheelEventSuper(MouseWheelEvent event) {
		super.processMouseWheelEvent(event);
	}

	/**
	 * Attempts to close the window. If the window does not want to be closed at this time, it will
	 * remain open and return <code>false</code>, otherwise it will close (the default
	 * implementation calls {@link #dispose()}).
	 * 
	 * @return <code>true</code> if the window closed successfully, <code>false</code> if it
	 *         remains open.
	 */
	public boolean attemptClose() {
		dispose();
		return true;
	}

	/** @return Whether this window has owned windows showing. */
	public boolean hasOwnedWindowsShowing() {
		for (Window window : getOwnedWindows()) {
			if (window.isShowing()) {
				return true;
			}
		}
		return false;
	}

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
			notifyOfWindowClosed(this);
		}
		adjustWindowMenus();
	}

	/**
	 * @param windowClass The window class to return.
	 * @param <T> The window type.
	 * @return The current visible windows, in order from top to bottom.
	 */
	public static <T extends TKWindow> ArrayList<T> getActiveWindows(Class<T> windowClass) {
		ArrayList<T> list = new ArrayList<T>();

		for (T window : new TKFilteredIterator<T>(WINDOW_LIST, windowClass)) {
			if (window.isShowing()) {
				list.add(window);
			}
		}
		return list;
	}

	public boolean isInForeground() {
		return mInForeground;
	}

	private void setInForeground(boolean inForeground) {
		if (inForeground != mInForeground) {
			mInForeground = inForeground;
			foregroundStateChanged(inForeground);

			if (!mInForeground) {
				TKBaseMenu menu = mUserInputManager.getMenuInUse();

				if (menu != null) {
					menu.closeCompletely(false);
				}
			}

			Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();
			if (focus instanceof TKPanel) {
				((TKPanel) focus).windowFocus(inForeground);
			}
		}
	}

	/**
	 * Called when the window changes its foreground state.
	 * 
	 * @param inForeground <code>true</code> will be passed in when the window is in the
	 *            foreground. <code>false</code> will be passed in when the window is not in the
	 *            foreground.
	 */
	protected void foregroundStateChanged(@SuppressWarnings("unused") boolean inForeground) {
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
		if (event.getWindow() == this) {
			attemptClose();
		}
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

	/**
	 * Adds a new selection listener to the window. This selection listener will be notified by the
	 * window each time the selection changes. This is a helper routine and is provided in case
	 * someone wants to implement the {@link TKSelectionManager} interface.
	 * 
	 * @param listener The selection listener to add to the window.
	 */
	public void addSelectionListener(TKSelectionListener listener) {
		if (mSelectionListeners == null) {
			mSelectionListeners = new ArrayList<TKSelectionListener>(1);
		}
		if (!mSelectionListeners.contains(listener)) {
			mSelectionListeners.add(listener);
		}
	}

	/**
	 * Dispatches a selection changed event to all of the registered {@link TKSelectionListener}s.
	 * 
	 * @param selectedObjects The new list of selected objects.
	 */
	public void fireSelectionChangedEvent(Collection<?> selectedObjects) {
		if (this instanceof TKSelectionManager && mSelectionListeners != null) {
			for (TKSelectionListener listener : new ArrayList<TKSelectionListener>(mSelectionListeners)) {
				listener.selectionChanged((TKSelectionManager) this, selectedObjects);
			}
		}
	}

	/**
	 * Removes the given selection listener from the selection manager's listener list. This is a
	 * helper routine and is provided in case someone wants to implement the
	 * {@link TKSelectionManager} interface.
	 * 
	 * @param listener The selection listener to remove from the window.
	 */
	public void removeSelectionListener(TKSelectionListener listener) {
		if (mSelectionListeners != null) {
			mSelectionListeners.remove(listener);
			if (mSelectionListeners.isEmpty()) {
				mSelectionListeners = null;
			}
		}
	}

	/**
	 * Overrides {@link #processComponentEvent(ComponentEvent)} to enforce minimum/maximum sizes.
	 * <p>
	 * {@inheritDoc}
	 */
	@Override protected void processComponentEvent(ComponentEvent event) {
		if (event.getID() == ComponentEvent.COMPONENT_RESIZED) {
			enforceMinMaxSize();
		}
		super.processComponentEvent(event);
	}

	public void componentHidden(ComponentEvent event) {
		// Not used.
	}

	public void componentMoved(ComponentEvent event) {
		// Not used.
	}

	public void componentResized(ComponentEvent event) {
		enforceMinMaxSize();
	}

	public void componentShown(ComponentEvent event) {
		// Not used.
	}

	private void enforceMinMaxSize() {
		Dimension origSize = getSize();
		Dimension otherSize = getMinimumSize();
		int width = origSize.width;
		int height = origSize.height;

		if (width < otherSize.width) {
			width = otherSize.width;
		}
		if (height < otherSize.height) {
			height = otherSize.height;
		}
		otherSize = getMaximumSize();
		if (width > otherSize.width) {
			width = otherSize.width;
		}
		if (height > otherSize.height) {
			height = otherSize.height;
		}
		if (width != origSize.width || height != origSize.height) {
			setSize(width, height);
		}
	}

	public void forceFocusToAccept() {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();

		if (focus instanceof TKTextField) {
			((TKTextField) focus).notifyActionListeners();
		}
	}

	public TKRepaintManager getRepaintManager() {
		return mRepaintManager;
	}

	public TKUndoManager getUndoManager() {
		return mUndoManager;
	}

	/** @param mgr The new undo manager. */
	public void setUndoManager(TKUndoManager mgr) {
		mUndoManager = mgr;
	}

	public TKUserInputManager getUserInputManager() {
		return mUserInputManager;
	}

	public TKMenuBar getTKMenuBar() {
		return mMenuBar;
	}

	public void setTKMenuBar(TKMenuBar menuBar) {
		TKMenuBar oldValue = mMenuBar;

		mMenuBar = menuBar;
		if (mMenuBar != null) {
			mMenuBar.setOpaque(true);
			mMenuBar.setVisible(true);
		}
		if (oldValue != null) {
			remove(oldValue);
		}
		if (mMenuBar != null) {
			add(mMenuBar);
		}
	}

	public TKToolBar getTKToolBar() {
		return mToolBar;
	}

	public void setTKToolBar(TKToolBar toolBar) {
		TKToolBar oldValue = mToolBar;

		mToolBar = toolBar;
		if (mToolBar != null) {
			mToolBar.setOpaque(true);
			mToolBar.setVisible(true);
		}
		if (oldValue != null) {
			remove(oldValue);
		}
		if (mToolBar != null) {
			add(mToolBar);
		}
	}

	/** @return The first progress bar found in the toolbar, if any. */
	public TKProgressBar getProgressBar() {
		if (mToolBar != null) {
			return mToolBar.getProgressBar();
		}
		return null;
	}

	/**
	 * Start the progress bar, with a delay of one second.
	 * 
	 * @param length The length of the operation.
	 */
	public void startProgress(int length) {
		startProgress(length, 1000);
	}

	/**
	 * Start the progress bar.
	 * 
	 * @param length The length of the operation.
	 * @param showAfter Only show the progress bar after this number of milliseconds have elapsed.
	 */
	public void startProgress(int length, long showAfter) {
		TKProgressBar progress = getProgressBar();

		if (progress != null) {
			if (length == 0) {
				progress.start(showAfter);
			} else {
				progress.setLength(length, showAfter);
			}
		}
	}

	/** Increment the progress bar. */
	public void incrementProgress() {
		TKProgressBar progress = getProgressBar();

		if (progress != null) {
			progress.increment();
		}
	}

	/** Finish the progress bar. */
	public void finishProgress() {
		TKProgressBar progress = getProgressBar();

		if (progress != null) {
			progress.finished();
		}
	}

	public boolean adjustToolBarItem(String command, TKPanel item) {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();
		boolean processed = false;

		if (focus != this && focus instanceof TKToolBarTarget) {
			processed = ((TKToolBarTarget) focus).adjustToolBarItem(command, item);
		}
		return processed;
	}

	public boolean obeyToolBarCommand(String command, TKPanel item) {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();
		boolean processed = false;

		if (focus != this && focus instanceof TKToolBarTarget) {
			processed = ((TKToolBarTarget) focus).obeyToolBarCommand(command, item);
		}
		return processed;
	}

	/**
	 * Asks each {@link TKWindow} to release an undo edit, is possible.
	 * 
	 * @return <code>true</code> if at least one edit was released.
	 */
	public static boolean releaseAnEditFromEachWindow() {
		boolean success = false;

		for (TKWindow window : getAllWindows()) {
			if (window.getUndoManager().releaseAnEdit()) {
				success = true;
			}
		}
		return success;
	}

	/** @return A list of all {@link TKWindow}s created by this application. */
	public static ArrayList<TKWindow> getAllWindows() {
		return getWindows(TKWindow.class);
	}

	/**
	 * @param <T> The window type.
	 * @param type The window type to return.
	 * @return A list of all windows of the specified type.
	 */
	public static <T extends TKWindow> ArrayList<T> getWindows(Class<T> type) {
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

	/** Adjusts the window menu in all windows. */
	public static void adjustWindowMenus() {
		TKWindow[] windows = TKWindow.getAllWindows().toArray(new TKWindow[0]);

		Arrays.sort(windows);

		for (TKWindow element : windows) {
			element.adjustWindowMenu(windows);
		}
	}

	/**
	 * Adjusts the window menu to contain the currently visible window set.
	 * 
	 * @param windows The current windows.
	 */
	public void adjustWindowMenu(TKWindow[] windows) {
		TKMenu menu = getWindowMenu();

		if (menu != null) {
			menu.removeAll();
			for (TKWindow element : windows) {
				TKMenuItem item = new TKMenuItem(element.getTitleForWindowMenu(), element.getTitleIcon(), CMD_WINDOW);

				if (element.isPartOfWindowGroup()) {
					item.setIndent(element.isWindowGroupParent() ? 23 : 33);
				}
				item.setUserObject(element);

				if (element.isWindowGroupParent()) {
					menu.add(new TKMenuSeparator(element.getWindowGroupName(), element.getWindowGroupIcon()));
				}
				menu.add(item);
			}
		}
	}

	/** @return The standard "Window" menu, if there is one. */
	public TKMenu getWindowMenu() {
		TKMenuBar menuBar = getTKMenuBar();

		if (menuBar != null) {
			return menuBar.getMenu(KEY_STD_MENUBAR_WINDOW_MENU);
		}
		return null;
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

	public int compareTo(TKWindow other) {
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
	 *         Returns the result of calling {@link TKPreferences#getInstance()} by default.
	 */
	public TKPreferences getWindowPreferences() {
		return TKPreferences.getInstance();
	}

	/**
	 * Saves the window bounds to preferences. Preferences must be saved for this to have a lasting
	 * effect.
	 */
	public void saveBounds() {
		String keyPrefix = getWindowPrefsPrefix();

		if (keyPrefix != null) {
			TKPreferences prefs = getWindowPreferences();
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
		restoreBounds(null);
	}

	/**
	 * Restores the window to its saved location and size.
	 * 
	 * @param defaultBounds The default bounds to use if none were previously saved
	 */
	public void restoreBounds(Rectangle defaultBounds) {
		TKPreferences prefs = getWindowPreferences();
		String keyPrefix = getWindowPrefsPrefix();

		prefs.resetIfVersionMisMatch(WINDOW_PREFERENCES, WINDOW_PREFERENCES_VERSION);
		pruneOldWindowPreferences(prefs);

		if (defaultBounds == null) {
			pack();
			defaultBounds = getBounds();
		}

		if (keyPrefix != null) {
			Point location = prefs.getPointValue(WINDOW_PREFERENCES, keyPrefix + KEY_LOCATION);
			Dimension size = prefs.getDimensionValue(WINDOW_PREFERENCES, keyPrefix + KEY_SIZE);

			if (location != null) {
				defaultBounds.x = location.x;
				defaultBounds.y = location.y;
			}

			if (size != null) {
				defaultBounds.width = size.width;
				defaultBounds.height = size.height;
			}
		}

		setBounds(defaultBounds);
		TKGraphics.forceOnScreen(this);

		if (prefs.getBooleanValue(WINDOW_PREFERENCES, keyPrefix + KEY_MAXIMIZED, false)) {
			setExtendedState(MAXIMIZED_BOTH);
		}
	}

	private void pruneOldWindowPreferences(TKPreferences prefs) {
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
	public static <T extends TKWindow> String getNextUntitledWindowName(Class<T> windowClass, String baseTitle, TKWindow exclude) {
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

	/**
	 * @return Whether the window is significant and should prevent the app from quitting if it is
	 *         currently displayed.
	 */
	public boolean isSignificant() {
		return true;
	}

	/**
	 * @param exclude A window to exclude from the check.
	 * @return Whether a significant window is still open.
	 */
	public static boolean isSignificantWindowOpen(TKWindow exclude) {
		for (TKWindow window : TKWindow.getAllWindows()) {
			if (exclude != window && window.isVisible() && window.isSignificant()) {
				return true;
			}
		}
		return false;
	}

	/**
	 * Adds a monitor to the list of monitors.
	 * 
	 * @param monitor The base window monitor to add.
	 */
	public static final void addBaseWindowMonitor(TKBaseWindowMonitor monitor) {
		synchronized (WINDOW_MONITORS) {
			if (monitor != null) {
				WINDOW_MONITORS.add(monitor);
			}
		}
	}

	/**
	 * Removes a monitor from the list of monitors.
	 * 
	 * @param monitor The base window monitor to remove.
	 */
	public static final void removeBaseWindowMonitor(TKBaseWindowMonitor monitor) {
		if (monitor != null) {
			synchronized (WINDOW_MONITORS) {
				WINDOW_MONITORS.remove(monitor);
			}
		}
	}

	/**
	 * Called whenever a base window is closed.
	 * 
	 * @param window The window that was closed.
	 */
	static final void notifyOfWindowClosed(TKBaseWindow window) {
		TKBaseWindowMonitor[] monitors;

		synchronized (WINDOW_MONITORS) {
			monitors = WINDOW_MONITORS.toArray(new TKBaseWindowMonitor[0]);
		}

		for (TKBaseWindowMonitor element : monitors) {
			try {
				element.windowClosed(window);
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
		}
	}
}
