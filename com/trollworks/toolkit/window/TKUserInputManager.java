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

import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.menu.TKBaseMenu;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.KeyboardFocusManager;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Window;
import java.awt.dnd.DragSource;
import java.awt.dnd.DragSourceDragEvent;
import java.awt.dnd.DragSourceDropEvent;
import java.awt.dnd.DragSourceEvent;
import java.awt.dnd.DragSourceListener;
import java.awt.event.InputEvent;
import java.awt.event.MouseEvent;
import java.awt.event.MouseWheelEvent;
import java.util.ArrayList;

/** Manages user input for base windows. */
public class TKUserInputManager {
	private static final String						MODULE						= "UserInputManager";		//$NON-NLS-1$
	private static final int						MODULE_VERSION				= 1;
	private static final String						TOOLTIPS_ENABLED_KEY		= "ShowToolTips";			//$NON-NLS-1$
	private static final String						TOOLTIP_DELAY_KEY			= "ToolTipDelay";			//$NON-NLS-1$
	private static final String						TOOLTIP_DURATION_KEY		= "ToolTipDuration";		//$NON-NLS-1$
	/** The default value for tooltip delay. */
	public static final long						DEFAULT_TOOLTIP_DELAY		= 1000;
	/** The default value for tooltip duration. */
	public static final long						DEFAULT_TOOLTIP_DURATION	= 10000;
	/** The minimum value for {@link #setToolTipDelay(long)}. */
	public static final long						MINIMUM_TOOLTIP_DELAY		= 0;
	/** The maximum value for {@link #setToolTipDelay(long)}. */
	public static final long						MAXIMUM_TOOLTIP_DELAY		= 5000;
	/** The minimum value for {@link #setToolTipDuration(long)}. */
	public static final long						MINIMUM_TOOLTIP_DURATION	= 5000;
	/** The maximum value for {@link #setToolTipDuration(long)}. */
	public static final long						MAXIMUM_TOOLTIP_DURATION	= 60000;
	private static long								TOOLTIP_DELAY				= DEFAULT_TOOLTIP_DELAY;
	private static long								TOOLTIP_DURATION			= DEFAULT_TOOLTIP_DURATION;
	private static boolean							TOOLTIPS_ENABLED			= true;
	private static boolean							DRAG_IN_PROGRESS			= false;
	private static DragMonitor						DRAG_MONITOR				= null;
	private static ArrayList<TKUserInputMonitor>	MONITORS					= null;
	private TKBaseWindow							mBaseWindow;
	private TKBaseMenu								mMenuInUse;
	private Component								mMouseOver;
	private boolean									mShowToolTipOK;
	private ShowToolTipTask							mCurrentTask;
	private TKTooltip								mToolTip;
	private MouseEvent								mToolTipTrigger;
	private TKPanel									mCurrentToolTipOwner;

	static {
		TKPreferences prefs = TKPreferences.getInstance();

		prefs.resetIfVersionMisMatch(MODULE, MODULE_VERSION);
		TOOLTIPS_ENABLED = prefs.getBooleanValue(MODULE, TOOLTIPS_ENABLED_KEY, true);
		setToolTipDelay(prefs.getLongValue(MODULE, TOOLTIP_DELAY_KEY, DEFAULT_TOOLTIP_DELAY));
		setToolTipDuration(prefs.getLongValue(MODULE, TOOLTIP_DURATION_KEY, DEFAULT_TOOLTIP_DURATION));

		TKKeyDispatcher keyDispatcher = new TKKeyDispatcher();
		KeyboardFocusManager mgr = KeyboardFocusManager.getCurrentKeyboardFocusManager();

		mgr.addKeyEventDispatcher(keyDispatcher);
		mgr.addKeyEventPostProcessor(keyDispatcher);
	}

	/**
	 * Creates a user input manager for the specified base window.
	 * 
	 * @param baseWindow The base window to provide user input management for.
	 */
	public TKUserInputManager(TKBaseWindow baseWindow) {
		mBaseWindow = baseWindow;
		if (DRAG_MONITOR == null) {
			DRAG_MONITOR = new DragMonitor();
		}
	}

	/** @return The maximum amount of time to wait after showing a tooltip before hiding it again. */
	public static long getToolTipDuration() {
		return TOOLTIP_DURATION;
	}

	/**
	 * @param duration The maximum amount of time to wait after showing a tooltip before hiding it
	 *            again.
	 */
	public static void setToolTipDuration(long duration) {
		if (duration < MINIMUM_TOOLTIP_DURATION) {
			duration = MINIMUM_TOOLTIP_DURATION;
		} else if (duration > MAXIMUM_TOOLTIP_DURATION) {
			duration = MAXIMUM_TOOLTIP_DURATION;
		}
		if (TOOLTIP_DURATION != duration) {
			TOOLTIP_DURATION = duration;
			TKPreferences.getInstance().setValue(MODULE, TOOLTIP_DURATION_KEY, duration);
		}
	}

	/** @return The amount of time that needs to pass before a tooltip will be shown. */
	public static long getToolTipDelay() {
		return TOOLTIP_DELAY;
	}

	/** @param delay The amount of time that needs to pass before a tooltip will be shown. */
	public static void setToolTipDelay(long delay) {
		if (delay < MINIMUM_TOOLTIP_DELAY) {
			delay = MINIMUM_TOOLTIP_DELAY;
		} else if (delay > MAXIMUM_TOOLTIP_DELAY) {
			delay = MAXIMUM_TOOLTIP_DELAY;
		}
		if (TOOLTIP_DELAY != delay) {
			TOOLTIP_DELAY = delay;
			TKPreferences.getInstance().setValue(MODULE, TOOLTIP_DELAY_KEY, delay);
		}
	}

	/** @return The menu in use, if any. */
	public TKBaseMenu getMenuInUse() {
		return mMenuInUse;
	}

	/**
	 * Set the current menu being used.
	 * 
	 * @param menu The menu in use.
	 */
	public void setMenuInUse(TKBaseMenu menu) {
		mMenuInUse = menu;
	}

	/**
	 * Processes mouse events.
	 * 
	 * @param event The event to process.
	 */
	public void processMouseEvent(MouseEvent event) {
		int id = event.getID();
		Component srcObj = (Component) event.getSource();
		TKPanel source = srcObj instanceof TKPanel ? (TKPanel) srcObj : null;
		Container window = (Container) mBaseWindow;

		// Ensure we get a {@link MouseEvent#MOUSE_EXITED} on any prior component that received
		// a {@link MouseEvent#MOUSE_ENTERED} before we issue the {@link MouseEvent#MOUSE_ENTERED}
		if (id == MouseEvent.MOUSE_ENTERED) {
			if (mMouseOver != null) {
				processMouseEvent(new MouseEvent(mMouseOver, MouseEvent.MOUSE_EXITED, event.getWhen(), event.getModifiers(), event.getX(), event.getY(), event.getClickCount(), event.isPopupTrigger()));
			}
			mMouseOver = srcObj;
		} else if (id == MouseEvent.MOUSE_EXITED) {
			mMouseOver = null;
		}

		notifyMonitors(event);

		// If the srcObj isn't a {@link TKPanel}, let AWT deal with it.
		if (source == null) {
			if (id == MouseEvent.MOUSE_MOVED || id == MouseEvent.MOUSE_DRAGGED || id == MouseEvent.MOUSE_WHEEL) {
				mBaseWindow.processMouseMotionEventSuper(event);
			} else {
				mBaseWindow.processMouseEventSuper(event);
			}
			return;
		}

		// If the event belongs to a component no longer attached to our window, consume it.
		if (!window.isAncestorOf(source)) {
			event.consume();
			return;
		}

		// Deal with special cases for each type of event.
		// This is primarily for ToolTips, but there is also
		// a bit of code for mouse motion coalescing and menu handling.
		switch (id) {
			case MouseEvent.MOUSE_RELEASED:
				setOKToShowToolTips(true);
				break;
			case MouseEvent.MOUSE_PRESSED:
				TKBaseMenu menu = getMenuInUse();

				setOKToShowToolTips(false);
				hideToolTip();

				if (menu != null && menu != source && !menu.isAncestorOf(source)) {
					if (source instanceof TKBaseMenu) {
						TKBaseMenu pMenu = menu;
						TKBaseMenu tMenu = pMenu.getParentMenu();

						while (tMenu != null && tMenu != source) {
							pMenu = tMenu;
							tMenu = pMenu.getParentMenu();
						}
						pMenu.close(false);
					} else {
						menu.closeCompletely(false);
					}
				}
				break;
			case MouseEvent.MOUSE_MOVED:
			case MouseEvent.MOUSE_DRAGGED:
			case MouseEvent.MOUSE_WHEEL:
				if (mToolTip == null || mToolTip != null && !mToolTip.getTipText().equals(source.getToolTipText(event))) {
					hideToolTip();
					resetToolTipTask(event, source);
				}
				break;
			case MouseEvent.MOUSE_ENTERED:
				resetToolTipTask(event, source);
				break;
			case MouseEvent.MOUSE_EXITED:
				setOKToShowToolTips(false);
				hideToolTip();
				break;
		}

		// Finally, allow the destination component to process the event
		if (source.isEnabled() || source.getMouseEventsEvenWhenDisabled()) {
			source.runMouseListeners(event);
			if (!event.isConsumed()) {
				source.processMouseEventSelf(event);
			}
		}
	}

	/** @return The current ToolTip owner. */
	public TKPanel getCurrentToolTipOwner() {
		return mToolTip != null ? mCurrentToolTipOwner : null;
	}

	/**
	 * Sets whether ToolTips are on/off.
	 * 
	 * @param enabled Whether ToolTips are on or off.
	 */
	public static void setToolTipsEnabled(boolean enabled) {
		if (TOOLTIPS_ENABLED != enabled) {
			TOOLTIPS_ENABLED = enabled;
			TKPreferences.getInstance().setValue(MODULE, TOOLTIPS_ENABLED_KEY, enabled);
			if (!enabled) {
				for (TKWindow window : TKWindow.getAllWindows()) {
					window.getUserInputManager().hideToolTip();
				}
			}
		}
	}

	/** @return Whether ToolTips are on/off. */
	public static boolean getToolTipsEnabled() {
		return TOOLTIPS_ENABLED && !isDragInProgress();
	}

	/** Hides the ToolTip if it is showing. */
	public void hideToolTip() {
		if (mToolTip != null) {
			mToolTip.dispose();
			mToolTip = null;
		}
	}

	/**
	 * Reset the ToolTip task.
	 * 
	 * @param event The new event to use as the trigger event.
	 * @param panel The new panel associated with the ToolTip.
	 */
	protected void resetToolTipTask(MouseEvent event, TKPanel panel) {
		if (getToolTipsEnabled()) {
			setOKToShowToolTips(true);
			if (haveCurrentTask()) {
				mCurrentTask.adjustEventAndTime(event, panel);
			} else {
				mCurrentTask = new ShowToolTipTask(event, panel);
			}
		}
	}

	/** @return <code>true</code> if there is a {@link ShowToolTipTask}. */
	protected boolean haveCurrentTask() {
		return mCurrentTask != null;
	}

	/**
	 * @param task The task to check.
	 * @return <code>true</code> if the specified task is the current task.
	 */
	protected boolean isCurrentTask(Runnable task) {
		return mCurrentTask == task;
	}

	/** Removes the current task. */
	protected void flushCurrentTask() {
		mCurrentTask = null;
	}

	/** @return <code>true</code> if tooltips can be shown. */
	protected boolean okToShowToolTips() {
		return mShowToolTipOK;
	}

	/**
	 * Sets whether tooltips can be shown.
	 * 
	 * @param ok Whether showing ToolTips is OK or not.
	 */
	protected void setOKToShowToolTips(boolean ok) {
		mShowToolTipOK = ok;
	}

	/**
	 * Shows the ToolTip for the specified panel.
	 * 
	 * @param event The event that triggered the ToolTip.
	 * @param panel The panel the ToolTip is for.
	 */
	public void showToolTip(MouseEvent event, TKPanel panel) {
		if (getToolTipsEnabled()) {
			Window window = (Window) mBaseWindow;

			if (window.isAncestorOf(panel) && mBaseWindow.isInForeground()) {
				String tooltip = panel.getToolTipText(event);

				if (mToolTip == null && tooltip != null && tooltip.length() > 0) {
					Rectangle screenBounds = TKGraphics.getMaximumWindowBounds(panel, new Rectangle(event.getX(), event.getY(), 1, 1));
					Dimension size;
					Rectangle prohibited;

					mToolTip = panel.getToolTip(event);
					mToolTip.pack();
					size = mToolTip.getSize();
					prohibited = panel.getToolTipProhibitedArea(event);
					TKPanel.convertRectangleToScreen(prohibited, panel);

					// Can it fit on the top?
					if (screenBounds.y + size.height < prohibited.y) {
						mToolTip.setLocation(prohibited.x, prohibited.y - size.height);
						// No? OK, how about the bottom?
					} else if (prohibited.y + prohibited.height + size.height < screenBounds.y + screenBounds.height) {
						mToolTip.setLocation(prohibited.x, prohibited.y + prohibited.height);
						// No? OK, how about the left?
					} else if (screenBounds.x + size.width < prohibited.x) {
						mToolTip.setLocation(prohibited.x - size.width, prohibited.y);
						// No? OK, how about the right?
					} else if (prohibited.x + prohibited.width + size.width < screenBounds.x + screenBounds.width) {
						mToolTip.setLocation(prohibited.x + prohibited.width, prohibited.y);
						// Still no? Geez... just put it on the top and let forceOnScreen() fix it.
					} else {
						mToolTip.setLocation(prohibited.x, prohibited.y - size.height);
					}

					TKGraphics.forceOnScreen(mToolTip);
					mToolTip.setVisible(true);
					mCurrentToolTipOwner = panel;
					mToolTipTrigger = event;
					TKTimerTask.schedule(new HideToolTipTask(mToolTip), getToolTipDuration());
				}
			}
		}
	}

	/**
	 * Adds a user input monitor.
	 * 
	 * @param monitor The monitor to add.
	 */
	public static void addMonitor(TKUserInputMonitor monitor) {
		if (MONITORS == null) {
			MONITORS = new ArrayList<TKUserInputMonitor>(1);
		}
		if (!MONITORS.contains(monitor)) {
			MONITORS.add(monitor);
		}
	}

	/**
	 * Notifies a user input monitor of an event.
	 * 
	 * @param event The event to pass to all monitors.
	 */
	public static void notifyMonitors(InputEvent event) {
		if (MONITORS != null) {
			for (TKUserInputMonitor monitor : MONITORS) {
				monitor.userInputEventOccurred(event);
			}
		}
	}

	/**
	 * Removes a user input monitor.
	 * 
	 * @param monitor The monitor to remove.
	 */
	public static void removeMonitor(TKUserInputMonitor monitor) {
		if (MONITORS != null) {
			MONITORS.remove(monitor);
			if (MONITORS.isEmpty()) {
				MONITORS = null;
			}
		}
	}

	/**
	 * Clones a {@link MouseEvent}.
	 * 
	 * @param event The event to clone.
	 * @return The new {@link MouseEvent}.
	 */
	public static final MouseEvent cloneMouseEvent(MouseEvent event) {
		if (event instanceof MouseWheelEvent) {
			MouseWheelEvent old = (MouseWheelEvent) event;

			return new MouseWheelEvent((Component) old.getSource(), old.getID(), System.currentTimeMillis(), old.getModifiers(), old.getX(), old.getY(), old.getClickCount(), old.isPopupTrigger(), old.getScrollType(), old.getScrollAmount(), old.getWheelRotation());
		}
		return new MouseEvent((Component) event.getSource(), event.getID(), System.currentTimeMillis(), event.getModifiers(), event.getX(), event.getY(), event.getClickCount(), event.isPopupTrigger());
	}

	/**
	 * Clones a {@link MouseEvent}.
	 * 
	 * @param event The event to clone.
	 * @param refreshTime Pass in <code>true</code> to generate a new time stamp.
	 * @return The new {@link MouseEvent}.
	 */
	public static final MouseEvent cloneMouseEvent(MouseEvent event, boolean refreshTime) {
		if (event instanceof MouseWheelEvent) {
			MouseWheelEvent old = (MouseWheelEvent) event;

			return new MouseWheelEvent((Component) old.getSource(), old.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), old.getModifiers(), old.getX(), old.getY(), old.getClickCount(), old.isPopupTrigger(), old.getScrollType(), old.getScrollAmount(), old.getWheelRotation());
		}
		return new MouseEvent((Component) event.getSource(), event.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), event.getModifiers(), event.getX(), event.getY(), event.getClickCount(), event.isPopupTrigger());
	}

	/**
	 * Clones a {@link MouseEvent}.
	 * 
	 * @param event The event to clone.
	 * @param source Pass in a new source.
	 * @param where Pass in a new location.
	 * @param refreshTime Pass in <code>true</code> to generate a new time stamp.
	 * @return The new {@link MouseEvent}.
	 */
	public static final MouseEvent cloneMouseEvent(MouseEvent event, Component source, Point where, boolean refreshTime) {
		if (event instanceof MouseWheelEvent) {
			MouseWheelEvent old = (MouseWheelEvent) event;

			return new MouseWheelEvent(source, old.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), old.getModifiers(), where.x, where.y, old.getClickCount(), old.isPopupTrigger(), old.getScrollType(), old.getScrollAmount(), old.getWheelRotation());
		}
		return new MouseEvent(source, event.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), event.getModifiers(), where.x, where.y, event.getClickCount(), event.isPopupTrigger());
	}

	/**
	 * Forwards a {@link MouseEvent} from one component to another.
	 * 
	 * @param event The event to forward.
	 * @param from The component that originally received the event.
	 * @param to The component the event should be forwarded to.
	 */
	static public void forwardMouseEvent(MouseEvent event, Component from, Component to) {
		translateMouseEvent(event, from, to);
		to.dispatchEvent(event);
	}

	/**
	 * Translates a {@link MouseEvent} from one component to another.
	 * 
	 * @param event The event that will be forwarded.
	 * @param from The component that originally received the event.
	 * @param to The component the event should be forwarded to.
	 */
	static public void translateMouseEvent(MouseEvent event, Component from, Component to) {
		Point evtPt = event.getPoint();
		Point pt = new Point(evtPt);

		TKPanel.convertPoint(pt, from, to);
		event.setSource(to);
		event.translatePoint(pt.x - evtPt.x, pt.y - evtPt.y);
	}

	/** @return <code>true</code> if a drag is in progress. */
	public static boolean isDragInProgress() {
		return DRAG_IN_PROGRESS;
	}

	/**
	 * Sets whether a drag is in progress or not.
	 * 
	 * @param dragInProgress Whether a drag is in progress or not.
	 */
	public static void setDragInProgress(boolean dragInProgress) {
		if (DRAG_IN_PROGRESS != dragInProgress) {
			DRAG_IN_PROGRESS = dragInProgress;
			if (dragInProgress) {
				for (TKWindow window : TKWindow.getAllWindows()) {
					TKUserInputManager mgr = window.getUserInputManager();

					mgr.setOKToShowToolTips(false);
					mgr.hideToolTip();
				}
			}
		}
	}

	/** @return The current tooltip being displayed. */
	protected TKTooltip getCurrentToolTip() {
		return mToolTip;
	}

	/** @return The mouse event that triggered the current tooltip. */
	public MouseEvent getToolTipTrigger() {
		return mToolTip != null ? mToolTipTrigger : null;
	}

	private class HideToolTipTask implements Runnable {
		private TKTooltip	mToolTipToHide;

		/**
		 * Create a new task to hide the tooltip.
		 * 
		 * @param tooltip The tooltip to hide.
		 */
		HideToolTipTask(TKTooltip tooltip) {
			mToolTipToHide = tooltip;
		}

		public void run() {
			if (mToolTipToHide == getCurrentToolTip()) {
				hideToolTip();
			}
		}
	}

	private class ShowToolTipTask implements Runnable {
		private MouseEvent	mToolTipEvent;
		private TKPanel		mToolTipComponent;
		private long		mWhen;

		/**
		 * Create a new task to show the tooltip.
		 * 
		 * @param event The event that triggered the tooltip to show.
		 * @param panel The panel to show the tooltip on.
		 */
		ShowToolTipTask(MouseEvent event, TKPanel panel) {
			mToolTipEvent = event;
			mToolTipComponent = panel;
			mWhen = System.currentTimeMillis() + getToolTipDelay();
			TKTimerTask.schedule(this, getToolTipDelay());
		}

		/**
		 * Adjust the tooltip event and panel information, reseting the triggering time.
		 * 
		 * @param event The event that triggered the tooltip to show.
		 * @param panel The panel to show the tooltip on.
		 */
		void adjustEventAndTime(MouseEvent event, TKPanel panel) {
			mToolTipEvent = event;
			mToolTipComponent = panel;
			mWhen = System.currentTimeMillis() + getToolTipDelay();
		}

		public void run() {
			if (isCurrentTask(this)) {
				if (okToShowToolTips()) {
					if (mWhen <= System.currentTimeMillis()) {
						setOKToShowToolTips(false);
						flushCurrentTask();
						showToolTip(mToolTipEvent, mToolTipComponent);
					} else {
						long delta = getToolTipDelay() - (mWhen - System.currentTimeMillis());

						TKTimerTask.schedule(this, delta < 1 ? 1 : delta);
					}
				} else {
					flushCurrentTask();
				}
			}
		}
	}

	private class DragMonitor implements DragSourceListener {
		/** Create the drag monitor. */
		DragMonitor() {
			DragSource.getDefaultDragSource().addDragSourceListener(this);
		}

		public void dragEnter(DragSourceDragEvent dsde) {
			setDragInProgress(true);
		}

		public void dragOver(DragSourceDragEvent dsde) {
			// Does nothing
		}

		public void dropActionChanged(DragSourceDragEvent dsde) {
			// Does nothing
		}

		public void dragDropEnd(DragSourceDropEvent dsde) {
			setDragInProgress(false);
		}

		public void dragExit(DragSourceEvent dse) {
			// Does nothing
		}
	}
}
