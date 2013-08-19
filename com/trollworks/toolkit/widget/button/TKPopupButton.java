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

package com.trollworks.toolkit.widget.button;

import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.widget.menu.TKTransientMenu;
import com.trollworks.toolkit.widget.menu.TKTransientMenuListener;
import com.trollworks.toolkit.window.TKUserInputManager;

import java.awt.AWTEvent;
import java.awt.Point;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;

/** A push button that will present a menu of options when pressed and held. */
public class TKPopupButton extends TKButton implements TKTransientMenuListener {
	private static final long	PRESS_DELAY	= 1000;
	private TKMenu				mMenu;
	private TKMenuTarget		mMenuTarget;
	private int					mItemToSelect;
	private int					mCounter;
	private BufferedImage		mPoppedIcon;
	private boolean				mHasDelay;

	/**
	 * Creates a popup button.
	 * 
	 * @param icon The icon to be displayed.
	 * @param menu The menu to use.
	 * @param target The target for the menu.
	 */
	public TKPopupButton(BufferedImage icon, TKMenu menu, TKMenuTarget target) {
		this(icon, menu, target, -1);
	}

	/**
	 * Creates a popup button.
	 * 
	 * @param icon The icon to be displayed.
	 * @param menu The menu to use.
	 * @param target The target for the menu.
	 * @param initialSelection The initially selected menu item index.
	 */
	public TKPopupButton(BufferedImage icon, TKMenu menu, TKMenuTarget target, int initialSelection) {
		super(icon);
		mMenu = menu;
		mMenuTarget = target;
		mItemToSelect = initialSelection;
		mHasDelay = true;
		enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
	}

	/** @return The initally selected menu item. */
	public int getItemToSelect() {
		return mItemToSelect;
	}

	@Override public BufferedImage getPressedIcon() {
		if (mMenu.isOpen() && mPoppedIcon != null) {
			return mPoppedIcon;
		}
		return super.getPressedIcon();
	}

	@Override public String getToolTipText(MouseEvent event) {
		if (mMenu.isOpen()) {
			return null;
		}
		return super.getToolTipText(event);
	}

	/**
	 * Prepares the menu for display by unmarking all items except the index specified earlier for
	 * selection.
	 */
	protected void prepareMenu() {
		int count = mMenu.getMenuItemCount();

		for (int i = 0; i < count; i++) {
			mMenu.getMenuItem(i).setMarked(i == mItemToSelect);
		}
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		int id = event.getID();

		if (id != MouseEvent.MOUSE_PRESSED && mMenu.isOpen()) {
			if (id == MouseEvent.MOUSE_RELEASED || id == MouseEvent.MOUSE_MOVED || id == MouseEvent.MOUSE_DRAGGED) {
				TKUserInputManager.forwardMouseEvent(event, this, mMenu);
			}
		} else {
			int width = getWidth();

			super.processMouseEventSelf(event);
			if (id == MouseEvent.MOUSE_PRESSED && !mMenu.isOpen() && event.getX() >= width - 10) {
				mCounter++;
				prepareMenu();
				TKTransientMenu.showTransientMenu(TKPopupButton.this, mMenuTarget, mMenu, mItemToSelect, new Point(width, getHeight() / 3));
				repaint();
			}
		}
	}

	/** @param itemIndex The initally selected menu item. */
	public void setItemToSelect(int itemIndex) {
		mItemToSelect = itemIndex;
	}

	/** @param icon The "popped" icon. */
	public void setPoppedIcon(BufferedImage icon) {
		BufferedImage oldValue = mPoppedIcon;

		mPoppedIcon = icon;
		if (mPoppedIcon != oldValue) {
			if (mPoppedIcon == null || oldValue == null || mPoppedIcon.getWidth() != oldValue.getWidth() || mPoppedIcon.getHeight() != oldValue.getHeight()) {
				revalidate();
			}
			repaint();
		}
	}

	@Override public void setPressed(boolean pressed) {
		if (pressed != mPressed) {
			mPressed = pressed;
			repaint();
			if (pressed) {
				if (mHasDelay) {
					TKTimerTask.schedule(new PressDelay(++mCounter), PRESS_DELAY);
				} else {
					delayedPress(++mCounter);
				}
			}
		}
	}

	public void transientMenuClosed(TKTransientMenu menu) {
		unpress();
		mInPress = false;
	}

	/**
	 * Handles a delayed press.
	 * 
	 * @param counter The current counter value.
	 */
	protected void delayedPress(int counter) {
		if (counter == mCounter && isPressed()) {
			prepareMenu();
			TKTransientMenu.showTransientMenu(TKPopupButton.this, mMenuTarget, mMenu, mItemToSelect, new Point(getWidth(), getHeight() / 3));
			repaint();
		}
	}

	/**
	 * @return Whether a delay is used to determine when a popup menu should be displayed.
	 */
	public boolean hasDelay() {
		return mHasDelay;
	}

	/**
	 * @param hasDelay Whether a delay is used to determine when a popup menu should be displayed.
	 */
	public void setHasDelay(boolean hasDelay) {
		mHasDelay = hasDelay;
	}

	private class PressDelay implements Runnable {
		private int	mThisCounter;

		/**
		 * Creates a new delayed press.
		 * 
		 * @param counter The current counter value.
		 */
		PressDelay(int counter) {
			mThisCounter = counter;
		}

		public void run() {
			delayedPress(mThisCounter);
		}
	}
}
