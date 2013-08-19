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

import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.window.TKBaseWindow;

import java.awt.Component;
import java.awt.Container;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashMap;

/** Manages the preparation and display of contextual menus. */
public class TKContextMenuManager extends TKBaseMenu implements TKMenuTarget {
	/** The key used for the partial menu that applies to all items in a selection. */
	public static final String				ALL_MENU_KEY	= "ALL";	//$NON-NLS-1$
	/** The amount used for the standard indent. */
	public static final int					STD_INDENT		= 20;
	private static TKContextMenuManager		INSTANCE		= null;
	private ArrayList<TKContextMenuHandler>	mHandlers;
	private TKMenu							mMenu;
	private HashMap<String, TKMenu>			mMenus;
	private TKPanel							mSelectionOwner;
	private Collection<? extends Object>	mSelection;
	private boolean							mHandlingContextMenu;

	private TKContextMenuManager() {
		super();
		mHandlers = new ArrayList<TKContextMenuHandler>();
		mMenus = new HashMap<String, TKMenu>();
	}

	/**
	 * Adds a divider to the specified menu if its item count is greater than 0 and the last item is
	 * not a separator.
	 * 
	 * @param menu The menu to operate on.
	 * @param indented Whether or not to indent the menu separator.
	 */
	public static void addDividerIfNecessary(TKMenu menu, boolean indented) {
		int count = menu.getMenuItemCount();

		if (count > 0 && !(menu.getMenuItem(count - 1) instanceof TKMenuSeparator)) {
			if (indented) {
				TKMenuSeparator separator = new TKMenuSeparator(true);

				separator.setIndent(STD_INDENT);
				menu.add(separator);
			} else {
				menu.addSeparator();
			}
		}
	}

	/**
	 * Add a contextual menu handler.
	 * 
	 * @param handler The handler to add.
	 */
	public static void addHandler(TKContextMenuHandler handler) {
		TKContextMenuManager instance = getInstance();

		if (!instance.mHandlers.contains(handler)) {
			instance.mHandlers.add(handler);
		}
	}

	/**
	 * Remove a contextual menu handler.
	 * 
	 * @param handler The handler to remove.
	 */
	public static void removeHandler(TKContextMenuHandler handler) {
		getInstance().mHandlers.remove(handler);
	}

	/**
	 * Adds a new menu item to the specified menu.
	 * 
	 * @param menu The menu to append to.
	 * @param title The menu title.
	 * @param command The command to attach to the menu.
	 * @return The menu item that was added.
	 */
	public static TKMenuItem addMenuItem(TKMenu menu, String title, String command) {
		return addMenuItem(menu, title, command, true, false, true);
	}

	/**
	 * Adds a new menu item to the specified menu.
	 * 
	 * @param menu The menu to append to.
	 * @param title The menu title.
	 * @param command The command to attach to the menu.
	 * @param enabled Pass in <code>true</code> to enable this menu item.
	 * @return The menu item that was added.
	 */
	public static TKMenuItem addMenuItem(TKMenu menu, String title, String command, boolean enabled) {
		return addMenuItem(menu, title, command, enabled, false, true);
	}

	/**
	 * Adds a new menu item to the specified menu.
	 * 
	 * @param menu The menu to append to.
	 * @param title The menu title.
	 * @param command The command to attach to the menu.
	 * @param enabled Pass in <code>true</code> to enable this menu item.
	 * @param marked Pass in <code>true</code> to mark this menu item.
	 * @param indent Pass in <code>true</code> to indent this menu item by {@link #STD_INDENT}
	 *            pixels.
	 * @return The menu item that was added.
	 */
	public static TKMenuItem addMenuItem(TKMenu menu, String title, String command, boolean enabled, boolean marked, boolean indent) {
		TKMenuItem item = new TKMenuItem(title, command);

		if (indent) {
			item.setIndent(STD_INDENT);
		}
		item.setMarked(marked);
		item.setEnabled(enabled);
		menu.add(item);
		return item;
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		return true;
	}

	private static void appendMenu(TKMenu baseMenu, TKMenu menu) {
		if (menu != null && menu.getMenuItemCount() > 0) {
			TKMenuItem item = menu.getTitleItem();
			String title = item.getTitle();
			int count = menu.getMenuItemCount();

			if (!title.equals(ALL_MENU_KEY)) {
				baseMenu.add(new TKMenuSeparator(title, item.getIcon()));
			}

			for (int i = 0; i < count; i++) {
				baseMenu.add(menu.getMenuItem(i));
			}
		}
	}

	@Override public void close(boolean commandWillBeProcessed) {
		mMenu.close(commandWillBeProcessed);
		super.close(commandWillBeProcessed);
		removeFromParent();
		if (!commandWillBeProcessed) {
			notifyOfFinish();
		}
	}

	/**
	 * Returns a new, indented menu.
	 * 
	 * @param title The menu title.
	 * @return The menu that was created.
	 */
	public static TKMenu createIndentedMenu(String title) {
		TKMenu menu = new TKMenu(title);

		menu.getTitleItem().setIndent(STD_INDENT);
		return menu;
	}

	/** @return The one and only instance this manager. */
	public static TKContextMenuManager getInstance() {
		if (INSTANCE == null) {
			INSTANCE = new TKContextMenuManager();
		}
		return INSTANCE;
	}

	/** @return The partial menu that applies to all items in a selection. */
	public static TKMenu getAllMenu() {
		return getPartialMenu(ALL_MENU_KEY, null, null);
	}

	/**
	 * @param key The key for the menu to retrieve. The resulting contextual menu is created by
	 *            concatenating all partial menus together in alphabetical order of their keys, with
	 *            the {@link #ALL_MENU_KEY} being forced to the top.
	 * @return The partial menu that applies to a specific type of item in a selection.
	 */
	public static TKMenu getPartialMenu(String key) {
		return getPartialMenu(key, null, null);
	}

	/**
	 * @param key The key for the menu to retrieve. The resulting contextual menu is created by
	 *            concatenating all partial menus together in alphabetical order of their keys, with
	 *            the {@link #ALL_MENU_KEY} being forced to the top.
	 * @param title The title to use when creating the menu, if necessary.
	 * @param icon The icon to use when creating the menu, if necessary.
	 * @return The partial menu that applies to a specific type of item in a selection.
	 */
	public static TKMenu getPartialMenu(String key, String title, BufferedImage icon) {
		HashMap<String, TKMenu> map = getInstance().mMenus;
		TKMenu menu;

		if (key == null) {
			key = ALL_MENU_KEY;
		}
		menu = map.get(key);
		if (menu == null) {
			menu = new TKMenu();
			map.put(key, menu);
			if (title == null) {
				title = key;
			}
		}
		if (title != null) {
			menu.getTitleItem().setTitle(title);
		}
		if (icon != null) {
			menu.getTitleItem().setIcon(icon);
		}
		return menu;
	}

	/**
	 * {@inheritDoc} Always returns <code>null</code>.
	 */
	@Override public TKMenuItem getMenuItem(int index) {
		return null;
	}

	/**
	 * {@inheritDoc} Always returns 0.
	 */
	@Override public int getMenuItemCount() {
		return 0;
	}

	private void notifyOfFinish() {
		for (TKContextMenuHandler handler : mHandlers) {
			handler.contextMenuDone(mSelectionOwner, mSelection);
		}

		mHandlingContextMenu = false;
		mSelectionOwner = null;
		mSelection = null;
		mMenu = null;
		mMenus.clear();
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		boolean processed = false;

		for (TKContextMenuHandler handler : mHandlers) {
			if (handler.obeyContextMenuCommand(command, item, mSelectionOwner, mSelection)) {
				processed = true;
				break;
			}
		}

		notifyOfFinish();
		return processed;
	}

	/**
	 * Builds the context menu and shows it.
	 * 
	 * @param event The mouse event that triggered this context menu.
	 * @param owner The owning panel for this context menu.
	 */
	public static void showContextMenu(MouseEvent event, TKPanel owner) {
		showContextMenu(event, owner, new ArrayList<Object>());
	}

	/**
	 * Builds the context menu and shows it.
	 * 
	 * @param event The mouse event that triggered this context menu.
	 * @param owner The owning panel for this context menu.
	 * @param selection The selection the context menu is operating on.
	 */
	public static void showContextMenu(MouseEvent event, TKPanel owner, Object[] selection) {
		ArrayList<Object> list = new ArrayList<Object>(selection.length);

		for (Object element : selection) {
			list.add(element);
		}

		showContextMenu(event, owner, list);
	}

	/**
	 * Builds the context menu and shows it.
	 * 
	 * @param event The mouse event that triggered this context menu.
	 * @param owner The owning panel for this context menu.
	 * @param selection The selection the context menu is operating on.
	 */
	public static void showContextMenu(MouseEvent event, TKPanel owner, Collection<? extends Object> selection) {
		TKContextMenuManager instance = getInstance();
		String[] keys;
		int i;

		instance.mSelectionOwner = owner;
		instance.mSelection = selection;
		instance.mMenu = new TKMenu();
		instance.mMenus.clear();

		for (TKContextMenuHandler handler : instance.mHandlers) {
			handler.contextMenuPrepare(owner, selection);
		}

		keys = instance.mMenus.keySet().toArray(new String[0]);
		Arrays.sort(keys);

		appendMenu(instance.mMenu, instance.mMenus.get(ALL_MENU_KEY));

		for (i = 0; i < keys.length; i++) {
			if (!ALL_MENU_KEY.equals(keys[i])) {
				appendMenu(instance.mMenu, instance.mMenus.get(keys[i]));
			}
		}

		if (instance.mMenu.getMenuItemCount() > 0) {
			TKBaseWindow bWindow = owner.getBaseWindow();

			if (bWindow != null) {
				Component component = (Component) event.getSource();
				Container window = (Container) bWindow;
				Rectangle bounds = new Rectangle(0, 0, 1, 1);

				convertRectangle(bounds, component, window);
				instance.setBounds(bounds);
				window.add(instance, 0);
				instance.mHandlingContextMenu = true;
				instance.mMenu.display(instance, instance, new Rectangle(event.getX() - STD_INDENT, event.getY(), 1, 1));
			}
		}
	}

	/** @return <code>true</code> if a context menu is being handled. */
	public static boolean isHandlingContextMenu() {
		return getInstance().mHandlingContextMenu;
	}
}
