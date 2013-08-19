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

import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKWidgetWindow;

import java.awt.AWTEvent;
import java.awt.Window;
import java.util.ArrayList;

/** The base class for menus. */
public abstract class TKBaseMenu extends TKPanel {
	/** The parent menu of this menu. */
	protected TKBaseMenu				mParentMenu;
	/** The targets of this menu. */
	protected ArrayList<TKMenuTarget>	mMenuTargets;

	/** Creates a base menu. */
	public TKBaseMenu() {
		super();
		mMenuTargets = new ArrayList<TKMenuTarget>(1);
		setOpaque(true);
		setBackground(TKColor.MENU_BACKGROUND);
		enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
	}

	/**
	 * Closes this menu.
	 * 
	 * @param commandWillBeProcessed Pass in <code>true</code> if a command will be processed once
	 *            the menu is closed.
	 */
	public void close(@SuppressWarnings("unused") boolean commandWillBeProcessed) {
		notifyOfUse(mParentMenu);
	}

	/**
	 * Closes this menu and any parent menus that led to this menu.
	 * 
	 * @param commandWillBeProcessed Pass in <code>true</code> if a command will be processed once
	 *            the menu is closed.
	 */
	public void closeCompletely(boolean commandWillBeProcessed) {
		TKBaseMenu menu = getParentMenu();

		if (menu != null) {
			menu.closeCompletely(commandWillBeProcessed);
		} else {
			close(commandWillBeProcessed);
		}
	}

	/**
	 * @param index The menu item's index.
	 * @return The menu item at the specified index.
	 */
	public abstract TKMenuItem getMenuItem(int index);

	/** @return The number of menu items in this menu. */
	public abstract int getMenuItemCount();

	/**
	 * @param command The command to look for.
	 * @return The menu item for the specified command.
	 */
	public TKMenuItem getMenuItemForCommand(String command) {
		int count = getMenuItemCount();

		for (int i = 0; i < count; i++) {
			TKMenuItem item = getMenuItem(i);
			TKBaseMenu menu;

			if (command.equals(item.getCommand())) {
				return item;
			}

			menu = item.getSubMenu();
			if (menu != null) {
				item = menu.getMenuItemForCommand(command);
				if (item != null) {
					return item;
				}
			}
		}
		return null;
	}

	/**
	 * @param keyStroke The key stroke to look for.
	 * @return The menu item for the specified key stroke.
	 */
	public TKMenuItem getMenuItemForKeyStroke(TKKeystroke keyStroke) {
		int count = getMenuItemCount();

		for (int i = 0; i < count; i++) {
			TKMenuItem item = getMenuItem(i);
			TKBaseMenu menu;

			if (keyStroke.equals(item.getKeyStroke())) {
				return item;
			}

			menu = item.getSubMenu();
			if (menu != null) {
				item = menu.getMenuItemForKeyStroke(keyStroke);
				if (item != null) {
					return item;
				}
			}
		}
		return null;
	}

	/** @return The menu targets. */
	public TKMenuTarget[] getMenuTargets() {
		return mMenuTargets.toArray(new TKMenuTarget[0]);
	}

	/** @return The parent menu. */
	public TKBaseMenu getParentMenu() {
		return mParentMenu;
	}

	/** @return The owning base window. */
	public TKBaseWindow getOwningBaseWindow() {
		TKBaseWindow window = getBaseWindow();

		if (window instanceof TKWidgetWindow) {
			window = (TKBaseWindow) ((Window) window).getOwner();
		}
		return window;
	}

	/**
	 * Issue the command for the specified item to this menu's targets.
	 * 
	 * @param item The item the command is being issued for.
	 */
	public void issueCommand(TKMenuItem item) {
		if (item.isEnabled() && item.getSubMenu() == null) {
			closeCompletely(true);
			if (!obeyCommand(item.getCommand(), item)) {
				System.err.println("Unhandled menu selection: " + item.getCommand()); //$NON-NLS-1$
			}
		}
	}

	/**
	 * Notifies the owning window that this menu is in use.
	 * 
	 * @param menu The menu that is in use.
	 */
	protected void notifyOfUse(TKBaseMenu menu) {
		TKBaseWindow window = getOwningBaseWindow();

		if (window != null) {
			window.getUserInputManager().setMenuInUse(menu);
		}
	}

	/**
	 * Adds a menu target.
	 * 
	 * @param target The target to add.
	 */
	public void addMenuTarget(TKMenuTarget target) {
		addMenuTarget(Integer.MAX_VALUE, target);
	}

	/**
	 * Adds a menu target at the specified index.
	 * 
	 * @param index The index to add the target at.
	 * @param target The target to add.
	 */
	public void addMenuTarget(int index, TKMenuTarget target) {
		removeMenuTarget(target);
		if (index < 0) {
			index = 0;
		} else {
			int size = mMenuTargets.size();

			if (index > size) {
				index = size;
			}
		}
		mMenuTargets.add(index, target);
	}

	/**
	 * Removes a menu target.
	 * 
	 * @param target The target to remove.
	 */
	public void removeMenuTarget(TKMenuTarget target) {
		mMenuTargets.remove(target);
	}

	/**
	 * Sets the parent menu.
	 * 
	 * @param menu The parent menu.
	 */
	public void setParentMenu(TKBaseMenu menu) {
		mParentMenu = menu;
	}

	/**
	 * Called before the specified menu item is displayed to allow it to be altered appropriately
	 * for the current state of the application.
	 * 
	 * @param command The command associated with the item.
	 * @param item The menu item to adjust.
	 * @return <code>true</code> if the menu item was dealt with and other {@link TKMenuTarget}s
	 *         should not be invoked.
	 */
	public boolean adjustMenuItem(String command, TKMenuItem item) {
		boolean processed = false;
		TKMenuTarget[] targets = getMenuTargets();

		for (int i = 0; !processed && i < targets.length; i++) {
			try {
				processed = targets[i].adjustMenuItem(command, item);
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
		}
		return processed;
	}

	/** Called prior to a series of menu adjustments starting. */
	public void menusWillBeAdjusted() {
		TKMenuTarget[] targets = getMenuTargets();

		for (TKMenuTarget element : targets) {
			try {
				element.menusWillBeAdjusted();
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
		}
	}

	/** Called after a series of menu adjustments has completed. */
	public void menusWereAdjusted() {
		TKMenuTarget[] targets = getMenuTargets();

		for (TKMenuTarget element : targets) {
			try {
				element.menusWereAdjusted();
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
		}
	}

	/**
	 * Called when the specified menu item has been selected.
	 * 
	 * @param command The command associated with the item.
	 * @param item The menu item that was selected.
	 * @return <code>true</code> if the command was dealt with and other {@link TKMenuTarget}s
	 *         should not be invoked.
	 */
	public boolean obeyCommand(String command, TKMenuItem item) {
		for (TKMenuTarget target : getMenuTargets()) {
			try {
				if (target.obeyCommand(command, item)) {
					return true;
				}
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
		}
		return false;
	}
}
