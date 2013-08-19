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

import com.trollworks.toolkit.qa.TKQAMenu;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKPlatform;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.window.TKUserInputManager;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import java.util.HashMap;

/**
 * Holds the menus for a base window. Before a menu is shown, the target of the menu bar will be
 * given a chance to adjust each of the menu items it contains. Once a menu item is selected, the
 * target will be called to process the selection.
 */
public class TKMenuBar extends TKBaseMenu implements TKMenuTarget {
	/** The debug key for showing the QA menu. */
	public static final String		SHOW_QA_MENU_KEY	= "SHOW_QA_MENU";	//$NON-NLS-1$
	private ArrayList<TKMenu>		mMenus;
	private HashMap<String, TKMenu>	mMenuMap;
	private int						mCurrentMenu;
	private TKQAMenu				mQAMenu;

	/**
	 * Creates an empty menu bar.
	 * 
	 * @param target The target of this menu bar.
	 */
	public TKMenuBar(TKMenuTarget target) {
		super();

		mMenus = new ArrayList<TKMenu>();
		mMenuMap = new HashMap<String, TKMenu>();
		mCurrentMenu = -1;
		addMenuTarget(target);

		setBorder(new TKCompoundBorder(new TKLineBorder(TKColor.MENU_BAR_SHADOW, 1, TKLineBorder.BOTTOM_EDGE), new TKEmptyBorder(0, 1, 0, 0)));
	}

	/** Adjusts the title of the menu item that quits the program to be appropriate for the platform. */
	public void adjustQuitMenu() {
		TKMenuItem item = getMenuItemForCommand(TKWindow.CMD_QUIT);

		if (item != null) {
			item.setTitle(getQuitTitle());
		}
	}

	/** @return The appropriate title for the menu item that quits the program. */
	public static String getQuitTitle() {
		return TKPlatform.isMacintosh() ? Msgs.QUIT_TITLE : Msgs.EXIT_TITLE;
	}

	/**
	 * Adds a menu to this menu bar.
	 * 
	 * @param menu The menu to add.
	 */
	public void add(TKMenu menu) {
		mMenus.add(menu);
		menu.getTitleItem().setFullDisplay(false);
	}

	/**
	 * Adds a menu to this menu bar.
	 * 
	 * @param index The index to insert the menu at.
	 * @param menu The menu to add.
	 */
	public void add(int index, TKMenu menu) {
		mMenus.add(index, menu);
		menu.getTitleItem().setFullDisplay(false);
	}

	@Override public void close(boolean commandWillBeProcessed) {
		if (mCurrentMenu != -1) {
			setSelectedMenu(-1);
		}
		super.close(commandWillBeProcessed);
	}

	/**
	 * Maps the specified menu to the specified key.
	 * 
	 * @param menu The menu to map.
	 * @param key The key to map the menu to.
	 */
	public void mapMenuToKey(TKMenu menu, String key) {
		mMenuMap.put(key, menu);
	}

	/**
	 * @param key The menu key.
	 * @return The menu that matches the specified key.
	 */
	public TKMenu getMenu(String key) {
		return mMenuMap.get(key);
	}

	/**
	 * @param index The menu's index.
	 * @return The menu at the specified index.
	 */
	public TKMenu getMenu(int index) {
		if (index >= 0 && index < mMenus.size()) {
			return mMenus.get(index);
		}
		return null;
	}

	/**
	 * @param index The menu's index.
	 * @return The bounds of the menu at the specified index, or <code>null</code> if there is
	 *         none.
	 */
	public Rectangle getMenuBounds(int index) {
		if (index >= 0 && index < getMenuCount()) {
			Insets insets = getInsets();
			Rectangle bounds = new Rectangle(insets.left, insets.top, getMenu(index).getTitleWidth(), getHeight() - (insets.top + insets.bottom));

			for (int i = 0; i < index; i++) {
				bounds.x += getMenu(i).getTitleWidth();
			}
			return bounds;
		}
		return null;
	}

	/** @return The number of menus currently installed. */
	public int getMenuCount() {
		return mMenus.size();
	}

	/**
	 * @param where The location.
	 * @return The menu index at the specified location, or -1 if there is none.
	 */
	public int getMenuIndexAt(Point where) {
		return getMenuIndexAt(where.x, where.y);
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The menu index at the specified location, or -1 if there is none.
	 */
	public int getMenuIndexAt(int x, int y) {
		if (x >= 0 && x < getWidth() && y >= 0 && y < getHeight()) {
			int count = getMenuCount();
			int pos = getInsets().left;

			for (int i = 0; i < count; i++) {
				TKMenu menu = getMenu(i);
				int width = menu.getTitleWidth();

				if (x >= pos && x < pos + width) {
					return i;
				}
				pos += width;
			}
		}
		return -1;
	}

	@Override public TKMenuItem getMenuItem(int index) {
		TKMenu menu = getMenu(index);

		if (menu != null) {
			return menu.getTitleItem();
		}
		return null;
	}

	@Override public int getMenuItemCount() {
		return getMenuCount();
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = getPreferredSizeSelf();

		size.width = MAX_SIZE;
		return size;
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return getPreferredSizeSelf();
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		int count = getMenuCount();
		Dimension size = new Dimension(insets.left + insets.right, 0);

		for (int i = 0; i < count; i++) {
			TKMenu menu = getMenu(i);
			int height = menu.getTitleHeight();

			size.width += menu.getTitleWidth();
			if (size.height < height) {
				size.height = height;
			}
		}

		size.height += insets.top + insets.bottom;
		return size;
	}

	/** @return The QA menu, if one is present. */
	public TKQAMenu getQAMenu() {
		return mQAMenu;
	}

	/** Installs the QA menu. */
	public void installQAMenu() {
		if (mQAMenu != null) {
			mMenus.remove(mQAMenu);
		}
		mQAMenu = new TKQAMenu();
		add(mQAMenu);
		revalidate();
	}

	/**
	 * Installs the QA menu.
	 * 
	 * @param onlyIfDebugKeyPresent If <code>true</code>, then the QA menu will only be installed
	 *            if the debug key {@link #SHOW_QA_MENU_KEY} is set.
	 */
	public void installQAMenu(boolean onlyIfDebugKeyPresent) {
		if (!onlyIfDebugKeyPresent || TKDebug.isKeySet(SHOW_QA_MENU_KEY)) {
			installQAMenu();
		}
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Insets insets = getInsets();
		int count = getMenuCount();
		int pos = insets.left;
		int top = insets.top + 2;
		int height = getHeight() - (insets.top + insets.bottom);

		for (int i = 0; i < count; i++) {
			TKMenu menu = getMenu(i);
			int width = menu.getTitleWidth();
			Color titleColor = null;

			if (mCurrentMenu == i && menu.isEnabled()) {
				drawMenuSelection(g2d, pos, top - 1, width - 1, height - 2);
				titleColor = Color.white;
			}
			menu.drawTitle(g2d, pos + 1, top, width, height, titleColor);
			pos += width;
		}
	}

	private void drawMenuSelection(Graphics2D g2d, int x, int y, int w, int h) {
		int ix = x + w - 1;
		int iy = y + h - 1;

		g2d.setPaint(TKColor.MENU_SELECTION_FILL);
		g2d.fillRect(x, y, w, h);
		g2d.setPaint(TKColor.MENU_SELECTION_LINE);
		g2d.drawRect(x, y, w, h);

		g2d.setPaint(TKColor.MENU_SELECTION_SHADOW);
		g2d.drawLine(ix, y + 1, ix, iy);
		g2d.drawLine(x + 1, iy, ix, iy);

		g2d.setPaint(TKColor.MENU_SELECTION_HIGHLIGHT);
		g2d.drawLine(x + 1, y + 1, x + 1, iy);
		g2d.drawLine(x + 1, y + 1, ix, y + 1);
	}

	@SuppressWarnings("fallthrough") @Override public void processMouseEventSelf(MouseEvent event) {
		int index;

		switch (event.getID()) {
			case MouseEvent.MOUSE_PRESSED:
				TKBaseMenu menu = getBaseWindow().getUserInputManager().getMenuInUse();

				if (menu != null) {
					TKBaseMenu parent = menu.getParentMenu();

					while (parent != null) {
						menu = parent;
						parent = menu.getParentMenu();
					}
					if (!(menu instanceof TKMenuBar)) {
						getBaseWindow().getUserInputManager().getMenuInUse().closeCompletely(false);
					}
				}

				index = getMenuIndexAt(event.getPoint());
				if (mCurrentMenu != -1) {
					boolean bail = mCurrentMenu == index;

					setSelectedMenu(-1);
					if (bail) {
						return;
					}
				}

				if (index != -1) {
					setSelectedMenu(index);
				}
				break;
			case MouseEvent.MOUSE_RELEASED:
				index = getMenuIndexAt(event.getPoint());
				if (index != mCurrentMenu && mCurrentMenu != -1) {
					TKUserInputManager.forwardMouseEvent(event, this, getMenu(mCurrentMenu));
					setSelectedMenu(-1);
				}
				break;
			case MouseEvent.MOUSE_MOVED:
				if (mCurrentMenu == -1) {
					break;
				}
				// Intentional fall-through!
			case MouseEvent.MOUSE_DRAGGED:
				index = getMenuIndexAt(event.getPoint());
				if (index != -1) {
					setSelectedMenu(index);
				} else if (mCurrentMenu != -1) {
					TKUserInputManager.forwardMouseEvent(event, this, getMenu(mCurrentMenu));
				}
				break;
		}
	}

	/**
	 * Sets the currently selected menu. Pass in <code>-1</code> to deselect all menus.
	 * 
	 * @param index The menu's index.
	 */
	public void setSelectedMenu(int index) {
		if (index != mCurrentMenu && index >= -1 && index < getMenuCount()) {
			if (mCurrentMenu != -1) {
				getMenu(mCurrentMenu).close(false);
				repaint(getMenuBounds(mCurrentMenu));
			}
			mCurrentMenu = index;
			if (mCurrentMenu != -1) {
				Rectangle bounds = getMenuBounds(mCurrentMenu);

				getMenu(mCurrentMenu).display(this, this, bounds);
				repaint(bounds);
			} else {
				notifyOfUse(null);
			}
		}
	}
}
