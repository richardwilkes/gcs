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

package com.trollworks.toolkit.widget.menu;

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKUserInputManager;
import com.trollworks.toolkit.window.TKWidgetWindow;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Window;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;
import java.util.ArrayList;

/** A standard menu. */
public class TKMenu extends TKBaseMenu implements TKMenuTarget, Runnable {
	/** Constants used when nothing was hit. */
	public static final int			NOTHING_HIT			= -1;
	/** Constants used when the scroll up control was hit. */
	public static final int			SCROLL_UP_HIT		= -2;
	/** Constants used when the scroll down control was hit. */
	public static final int			SCROLL_DOWN_HIT		= -3;
	private static final long		MINIMUM_SCROLL_TIME	= 100;
	private ArrayList<TKMenuItem>	mMenuItems;
	private int						mCurrentMenuItem;
	private boolean					mOpen;
	private TKMenuItem				mTitleItem;
	private boolean					mTopScrollNeeded;
	private boolean					mBottomScrollNeeded;
	private int						mTopVisibleItem;
	private long					mLastScrollTime;
	private boolean					mAutoScrollOK;
	private boolean					mTimerPending;
	private int						mOperation;
	private int						mMaxScroll;
	private boolean					mIgnoreFirstMouseReleased;
	private int						mScrollUpIconWidth;
	private int						mScrollUpIconHeight;
	private int						mScrollDownIconWidth;
	private int						mScrollDownIconHeight;
	private TKWidgetWindow			mWidgetWindow;

	/** Creates a new, empty menu with no title. */
	public TKMenu() {
		this(" "); //$NON-NLS-1$
	}

	/**
	 * Creates a menu with the specified title.
	 * 
	 * @param title The title to use.
	 */
	public TKMenu(String title) {
		this(title, null);
	}

	/**
	 * Creates a menu with the specified icon.
	 * 
	 * @param icon The icon to use.
	 */
	public TKMenu(BufferedImage icon) {
		this(null, icon);
	}

	/**
	 * Creates a menu with the specified title and icon.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 */
	public TKMenu(String title, BufferedImage icon) {
		super();

		BufferedImage aIcon = TKImage.getUpArrowIcon();
		mScrollUpIconWidth = aIcon.getWidth();
		mScrollUpIconHeight = aIcon.getHeight();
		aIcon = TKImage.getDownArrowIcon();
		mScrollDownIconWidth = aIcon.getWidth();
		mScrollDownIconHeight = aIcon.getHeight();
		mCurrentMenuItem = NOTHING_HIT;
		mMenuItems = new ArrayList<TKMenuItem>();
		mTitleItem = new TKMenuItem(title, icon);
		mTitleItem.setSubMenu(this);
	}

	@Override public void setToolTipText(String text) {
		// Ignored.
	}

	/**
	 * Adds the specified menu to this menu.
	 * 
	 * @param menu The menu to add.
	 */
	public void add(TKMenu menu) {
		add(menu.getTitleItem());
	}

	/**
	 * Adds the specified menu item to this menu.
	 * 
	 * @param menuItem The menu item to add.
	 */
	public void add(TKMenuItem menuItem) {
		mMenuItems.add(menuItem);
	}

	/** Adds a separator to this menu. */
	public void addSeparator() {
		add(new TKMenuSeparator());
	}

	/**
	 * Inserts the specified menu into this menu.
	 * 
	 * @param menu The menu to insert.
	 * @param index The index to insert it at.
	 */
	public void insert(TKMenu menu, int index) {
		insert(menu.getTitleItem(), index);
	}

	/**
	 * Inserts the specified menu item into this menu.
	 * 
	 * @param menuItem The menu item to insert.
	 * @param index The index to insert it at.
	 */
	public void insert(TKMenuItem menuItem, int index) {
		mMenuItems.add(index, menuItem);
	}

	/** Give the menu's target a chance to adjust the menu items of this menu. */
	public void adjustMenuItems() {
		int count;

		menusWillBeAdjusted();

		count = getMenuItemCount();
		for (int i = 0; i < count; i++) {
			TKMenuItem item = getMenuItem(i);

			if (!adjustMenuItem(item.getCommand(), item)) {
				item.setEnabled(item.getSubMenu() != null);
			}
		}

		menusWereAdjusted();
	}

	@Override public void close(boolean commandWillBeProcessed) {
		if (isOpen()) {
			TKBaseMenu parentMenu = getParentMenu();
			Container parent = getParent();

			mOpen = false;
			setSelectedMenuItem(NOTHING_HIT);
			super.close(commandWillBeProcessed);
			if (mWidgetWindow != null) {
				mWidgetWindow.dispose();
				mWidgetWindow = null;
				parent = null;
			} else {
				parent.remove(this);
			}
			if (parentMenu != null && parentMenu instanceof TKMenu) {
				((TKMenu) parentMenu).setSelectedMenuItem(NOTHING_HIT);
			}
			if (parent != null) {
				parent.repaint(getX(), getY(), getWidth(), getHeight());
			}
		}
	}

	/**
	 * Displays this menu.
	 * 
	 * @param owner The base menu that is initiating the menu display.
	 * @param target The target of this menu.
	 * @param titleBounds The bounds within <code>owner</code> of this menu's title.
	 */
	public void display(TKBaseMenu owner, TKMenuTarget target, Rectangle titleBounds) {
		display(owner, target, titleBounds, -1);
	}

	/**
	 * Displays this menu.
	 * 
	 * @param owner The base menu that is initiating the menu display.
	 * @param target The target of this menu.
	 * @param titleBounds The bounds within <code>owner</code> of this menu's title.
	 * @param selectedItem The menu item that should be initially selected.
	 */
	public void display(TKBaseMenu owner, TKMenuTarget target, Rectangle titleBounds, int selectedItem) {
		if (!isOpen()) {
			Rectangle windowBounds = TKGraphics.getMaximumWindowBounds(owner, titleBounds);
			Insets insets = getInsets();
			Point pt;
			Dimension size;
			int menuCount;
			int i;

			if (owner instanceof TKMenuBar) {
				Rectangle barBounds = owner.getLocalBounds();
				int delta;

				convertRectangleToScreen(barBounds, owner);
				delta = barBounds.y + barBounds.height - windowBounds.y;
				if (delta >= windowBounds.height) {
					delta = windowBounds.height - 1;
				}
				if (delta > 0) {
					windowBounds.y += delta;
					windowBounds.height -= delta;
				}
			}

			if (selectedItem < 0 || selectedItem >= getMenuItemCount()) {
				selectedItem = -1;
			}

			mTopScrollNeeded = false;
			mBottomScrollNeeded = false;
			mTopVisibleItem = 0;
			setParentMenu(owner);
			addMenuTarget(target);
			adjustMenuItems();

			invalidate();
			size = getPreferredSize();

			menuCount = getMenuItemCount();
			if (owner instanceof TKMenu) {
				pt = new Point(titleBounds.x + titleBounds.width, titleBounds.y);
			} else if (owner instanceof TKMenuBar) {
				pt = new Point(titleBounds.x, titleBounds.y + titleBounds.height);
			} else {
				pt = new Point(titleBounds.x, titleBounds.y);
				if (selectedItem > 0) {
					pt.y -= insets.top;
					for (i = 0; i < selectedItem; i++) {
						pt.y -= getMenuItem(i).getHeight();
					}
				}
				if (size.width < titleBounds.width) {
					size.width = titleBounds.width;
				}
			}

			convertPointToScreen(pt, owner);

			if (pt.x < windowBounds.x) {
				pt.x = windowBounds.x;
			}

			if (pt.x + size.width > windowBounds.x + windowBounds.width) {
				if (owner instanceof TKMenu) {
					pt.x = titleBounds.x - size.width;
					pt.y = titleBounds.y;
					convertPointToScreen(pt, owner);
				} else {
					pt.x = windowBounds.x + windowBounds.width - size.width;
				}
			}

			if (pt.x < windowBounds.x) {
				pt.x = windowBounds.x;
			}

			if (pt.y < windowBounds.y) {
				pt.y = windowBounds.y;
			}

			if (pt.y + size.height > windowBounds.y + windowBounds.height) {
				if (size.height <= windowBounds.height) {
					pt.y = windowBounds.y + windowBounds.height - size.height;
				} else {
					int bottom = windowBounds.y + windowBounds.height - (insets.bottom + mScrollUpIconHeight);

					pt.y = windowBounds.y;
					size.height = windowBounds.height;
					mTopScrollNeeded = false;
					mBottomScrollNeeded = true;

					for (i = menuCount - 1; i >= 0; i--) {
						bottom -= getMenuItem(i).getHeight();
						if (bottom < windowBounds.y + insets.top) {
							mMaxScroll = i + 1;
							break;
						}
					}
				}
			}

			if (pt.y < windowBounds.y) {
				pt.y = windowBounds.y;
			}

			TKBaseWindow owningWindow = owner.getOwningBaseWindow();
			Window window = (Window) owningWindow;
			Rectangle baseWindowBounds = window.getBounds();
			Insets baseWindowInsets = window.getInsets();
			Rectangle menuBounds = new Rectangle(pt.x, pt.y, size.width, size.height);
			baseWindowBounds.x = baseWindowInsets.left;
			baseWindowBounds.y = baseWindowInsets.top;
			baseWindowBounds.width -= baseWindowInsets.left + baseWindowInsets.right;
			baseWindowBounds.height -= baseWindowInsets.top + baseWindowInsets.bottom;
			convertRectangleFromScreen(menuBounds, window);
			if (baseWindowBounds.contains(menuBounds)) {
				setBounds(menuBounds);
				window.add(this, 0);
				mWidgetWindow = null;
			} else {
				mWidgetWindow = new TKWidgetWindow(owningWindow);
				mWidgetWindow.add(this);
				mWidgetWindow.setLocation(pt);
				mWidgetWindow.setSize(size);
			}

			mOpen = true;
			validate();
			repaint();

			if (mWidgetWindow != null) {
				mWidgetWindow.setVisible(true);
			}

			notifyOfUse(this);

			if (selectedItem != -1) {
				setSelectedMenuItem(selectedItem);
				mIgnoreFirstMouseReleased = true;
			}
		}
	}

	/**
	 * Draws the title for this menu into the specified bounds.
	 * 
	 * @param g2d The graphics context to use for drawing.
	 * @param x The starting horizontal location of the area that can be drawn into.
	 * @param y The starting vertical location of the area that can be drawn into.
	 * @param width The width of the area that can be drawn into.
	 * @param height The height of the area that can be drawn into.
	 * @param color If this is not <code>null</code>, it will override the normal color used when
	 *            drawing the text in a menu item.
	 */
	public void drawTitle(Graphics2D g2d, int x, int y, int width, int height, Color color) {
		mTitleItem.draw(g2d, x, y, width, height, color);
	}

	/** @return The current set of menu items. */
	public ArrayList<TKMenuItem> getMenuItems() {
		return new ArrayList<TKMenuItem>(mMenuItems);
	}

	@Override public TKMenuItem getMenuItem(int index) {
		if (index >= 0 && index < mMenuItems.size()) {
			return mMenuItems.get(index);
		}
		return null;
	}

	/**
	 * @param index The menu index.
	 * @return The bounds of the menu item at the specified index, or <code>null</code> if there
	 *         is none.
	 */
	public Rectangle getMenuItemBounds(int index) {
		if (index >= mTopVisibleItem && index < getMenuItemCount()) {
			Insets insets = getInsets();
			Rectangle bounds = new Rectangle(insets.left, insets.top + 2, getWidth() - (insets.left + insets.right), getMenuItem(index).getHeight());

			if (mTopScrollNeeded) {
				bounds.y += mScrollUpIconHeight;
			}

			for (int i = mTopVisibleItem; i < index; i++) {
				bounds.y += getMenuItem(i).getHeight();
			}
			return bounds;
		}
		return null;
	}

	@Override public int getMenuItemCount() {
		return mMenuItems.size();
	}

	/**
	 * @param item The menu item's index.
	 * @return The index of the specified menu item, or {@link #NOTHING_HIT}if it is not part of
	 *         this menu.
	 */
	public int getMenuItemIndex(TKMenuItem item) {
		int count = getMenuItemCount();

		for (int i = 0; i < count; i++) {
			TKMenuItem myItem = getMenuItem(i);

			if (item == myItem) {
				return i;
			}
		}
		return NOTHING_HIT;
	}

	/**
	 * @param where The location.
	 * @return The menu item index at the specified location, or {@link #NOTHING_HIT}if there is
	 *         none.
	 */
	public int getMenuItemIndexAt(Point where) {
		return getMenuItemIndexAt(where.x, where.y);
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The menu item index at the specified location, or {@link #NOTHING_HIT}if there is
	 *         none.
	 */
	public int getMenuItemIndexAt(int x, int y) {
		if (x >= 0 && x < getWidth() && y >= 0 && y < getHeight()) {
			int pos = getInsets().top;

			if (mTopScrollNeeded) {
				if (y < mScrollUpIconHeight) {
					return SCROLL_UP_HIT;
				}
				pos += mScrollUpIconHeight;
			}
			if (mBottomScrollNeeded && y >= getHeight() - mScrollDownIconHeight) {
				return SCROLL_DOWN_HIT;
			}

			int count = getMenuItemCount();

			for (int i = mTopVisibleItem; i < count; i++) {
				TKMenuItem item = getMenuItem(i);
				int height = item.getHeight();

				if (y >= pos && y < pos + height) {
					return i;
				}
				pos += height;
			}
		}
		return NOTHING_HIT;
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		Dimension size = new Dimension();
		int count = getMenuItemCount();
		int dividingPoint = 0;
		int i;

		for (i = 0; i < count; i++) {
			int tmp = getMenuItem(i).getPreferredDividingPoint();

			if (tmp > dividingPoint) {
				dividingPoint = tmp;
			}
		}

		for (i = 0; i < count; i++) {
			TKMenuItem item = getMenuItem(i);
			int width = item.getWidth(dividingPoint);

			if (width > size.width) {
				size.width = width;
			}
			size.height += item.getHeight();
		}

		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom + 3;

		return size;
	}

	/** @return The preferred height of the title for this menu. */
	public int getTitleHeight() {
		return mTitleItem.getHeight();
	}

	/** @return The menu item representing this menu's title. */
	public TKMenuItem getTitleItem() {
		return mTitleItem;
	}

	/** @return The preferred width of the title for this menu. */
	public int getTitleWidth() {
		mTitleItem.invalidate();
		return mTitleItem.getWidth(mTitleItem.getPreferredDividingPoint());
	}

	/** @return <code>true</code> if this menu is currently open. */
	public boolean isOpen() {
		return mOpen;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Insets insets = getInsets();
		int count = getMenuItemCount();
		int left = insets.left;
		int top = insets.top;
		int height = getHeight();
		int width = getWidth() - (insets.left + insets.right);
		int bottom = height - insets.bottom;

		drawBackground(g2d, left, top, width - 1, height - 1);

		if (mTopScrollNeeded) {
			g2d.drawImage(TKImage.getUpArrowIcon(), left + (width - mScrollUpIconWidth) / 2, 0, null);
			top += mScrollUpIconHeight;
		}
		if (mBottomScrollNeeded) {
			bottom -= mScrollDownIconHeight;
			g2d.drawImage(TKImage.getDownArrowIcon(), left + (width - mScrollDownIconWidth) / 2, bottom, null);
		}

		top += 2;

		for (int i = mTopVisibleItem; i < count; i++) {
			TKMenuItem item = getMenuItem(i);
			int itemHeight = item.getHeight();
			Color itemColor = null;

			if (mBottomScrollNeeded && top + itemHeight > bottom) {
				break;
			}

			if (mCurrentMenuItem == i && item.isEnabled()) {
				drawMenuSelection(g2d, left + 2, top, width - 5, itemHeight - 2);
				itemColor = Color.white;
			}

			item.getWidth(item.getPreferredDividingPoint()); // Here to insure we have valid info
			// in the item for drawing.
			item.draw(g2d, left, top, width, itemHeight, itemColor);
			top += itemHeight;
		}
	}

	private void drawBackground(Graphics2D g2d, int x, int y, int w, int h) {
		g2d.setPaint(TKColor.MENU_OUTLINE);
		g2d.drawRect(x, y, w, h);

		g2d.setPaint(TKColor.MENU_SHADOW);

		int ix = x + w - 1;
		int iy = y + h - 1;
		g2d.drawLine(ix, y + 1, ix, iy);
		g2d.drawLine(x + 1, iy, ix, iy);

		g2d.setPaint(TKColor.MENU_HIGHLIGHT);
		g2d.drawLine(x + 1, y + 1, x + 1, iy);
		g2d.drawLine(x + 1, y + 1, ix, y + 1);
	}

	private void drawMenuSelection(Graphics2D g2d, int x, int y, int w, int h) {
		g2d.setPaint(TKColor.MENU_SELECTION_FILL);
		g2d.fillRect(x, y, w, h);
		g2d.setPaint(TKColor.MENU_SELECTION_LINE);
		g2d.drawRect(x, y, w, h);

		g2d.setPaint(TKColor.MENU_SELECTION_SHADOW);

		int ix = x + w - 1;
		int iy = y + h - 1;
		g2d.drawLine(ix, y + 1, ix, iy);
		g2d.drawLine(x + 1, iy, ix, iy);

		g2d.setPaint(TKColor.MENU_SELECTION_HIGHLIGHT);
		g2d.drawLine(x + 1, y + 1, x + 1, iy);
		g2d.drawLine(x + 1, y + 1, ix, y + 1);
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		switch (event.getID()) {
			case MouseEvent.MOUSE_PRESSED:
				int which = getMenuItemIndexAt(event.getPoint());

				if (which > NOTHING_HIT) {
					setSelectedMenuItem(which);
				} else {
					scrollMenu(which);
				}
				break;
			case MouseEvent.MOUSE_RELEASED:
				if (!mIgnoreFirstMouseReleased) {
					if (mCurrentMenuItem != NOTHING_HIT && getMenuItemIndexAt(event.getPoint()) == mCurrentMenuItem) {
						issueCommand(getMenuItem(mCurrentMenuItem));
					} else if (mCurrentMenuItem > NOTHING_HIT) {
						TKBaseMenu menu = getMenuItem(mCurrentMenuItem).getSubMenu();

						if (menu != null) {
							TKUserInputManager.forwardMouseEvent(event, this, menu);
						} else {
							setSelectedMenuItem(NOTHING_HIT);
						}
					}
				}
				mIgnoreFirstMouseReleased = false;
				break;
			case MouseEvent.MOUSE_EXITED:
				if (mCurrentMenuItem > NOTHING_HIT) {
					TKMenuItem item = getMenuItem(mCurrentMenuItem);
					if (!item.isEnabled() || item.getSubMenu() == null) {
						setSelectedMenuItem(NOTHING_HIT);
					}
				}
				mIgnoreFirstMouseReleased = false;
				break;
			case MouseEvent.MOUSE_MOVED:
			case MouseEvent.MOUSE_DRAGGED:
				int index = getMenuItemIndexAt(event.getPoint());

				if (index > NOTHING_HIT) {
					setSelectedMenuItem(index);
				} else if (index == SCROLL_UP_HIT || index == SCROLL_DOWN_HIT) {
					scrollMenu(index);
				} else if (mCurrentMenuItem > NOTHING_HIT) {
					TKBaseMenu menu = getMenuItem(mCurrentMenuItem).getSubMenu();

					if (menu != null) {
						TKUserInputManager.forwardMouseEvent(event, this, menu);
					}
				}
				mIgnoreFirstMouseReleased = false;
				break;
		}
	}

	/**
	 * Removes the specified menu from this menu.
	 * 
	 * @param menu The menu to remove.
	 */
	public void remove(TKMenu menu) {
		remove(menu.getTitleItem());
	}

	/**
	 * Removes the specified menu item from this menu.
	 * 
	 * @param menuItem The menu item to remove.
	 */
	public void remove(TKMenuItem menuItem) {
		mMenuItems.remove(menuItem);
	}

	/** Removes all menu items from this menu. */
	@Override public void removeAll() {
		mMenuItems.clear();
	}

	public void run() {
		mTimerPending = false;
		if (mAutoScrollOK) {
			scrollMenu(mOperation);
		}
	}

	private void scrollMenu(int operation) {
		long now = System.currentTimeMillis();

		if (mLastScrollTime + MINIMUM_SCROLL_TIME <= now) {
			mLastScrollTime = now;
			setSelectedMenuItem(NOTHING_HIT);
			mAutoScrollOK = true;
			if (operation == SCROLL_UP_HIT) {
				if (--mTopVisibleItem < 0) {
					mTopVisibleItem = 0;
					mTopScrollNeeded = false;
					mAutoScrollOK = false;
				}
				mBottomScrollNeeded = true;
			} else if (operation == SCROLL_DOWN_HIT) {
				if (++mTopVisibleItem >= mMaxScroll) {
					mTopVisibleItem = mMaxScroll;
					mBottomScrollNeeded = false;
					mAutoScrollOK = false;
				}
				mTopScrollNeeded = true;
			}
			repaint();
			if (!mTimerPending) {
				mTimerPending = true;
				mOperation = operation;
				TKTimerTask.schedule(this, MINIMUM_SCROLL_TIME);
			}
		}
	}

	/**
	 * Sets the currently selected menu item. Pass in {@link #NOTHING_HIT} to deselect all menu
	 * items.
	 * 
	 * @param index The index to select.
	 */
	public void setSelectedMenuItem(int index) {
		if (index != mCurrentMenuItem && index >= NOTHING_HIT && index < getMenuItemCount()) {
			TKBaseMenu menu;
			TKMenuItem item;

			if (mCurrentMenuItem > NOTHING_HIT) {
				item = getMenuItem(mCurrentMenuItem);
				menu = item.getSubMenu();
				if (menu != null) {
					menu.close(false);
				}
				if (item.isEnabled()) {
					repaint(getMenuItemBounds(mCurrentMenuItem));
				}
			}
			mCurrentMenuItem = index;
			if (mCurrentMenuItem > NOTHING_HIT) {
				Rectangle bounds = getMenuItemBounds(mCurrentMenuItem);

				item = getMenuItem(mCurrentMenuItem);
				menu = item.getSubMenu();
				if (menu != null && menu instanceof TKMenu) {
					((TKMenu) menu).display(this, this, bounds);
				}
				if (item.isEnabled()) {
					repaint(bounds);
				}
				mAutoScrollOK = false;
			}
		}
	}
}
