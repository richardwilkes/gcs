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

import java.awt.Container;
import java.awt.Point;
import java.awt.Rectangle;

/** Provides for the display of a transient menu. */
public class TKTransientMenu extends TKBaseMenu {
	private TKMenu					mMenu;
	private TKTransientMenuListener	mTransientMenuListener;

	/**
	 * Create a new transient menu.
	 * 
	 * @param menu The menu to display.
	 * @param listener A listener for this transient menu.
	 */
	protected TKTransientMenu(TKMenu menu, TKTransientMenuListener listener) {
		super();
		mMenu = menu;
		mTransientMenuListener = listener;
	}

	@Override public void close(boolean commandWillBeProcessed) {
		mMenu.close(commandWillBeProcessed);
		super.close(commandWillBeProcessed);
		removeFromParent();
		if (mTransientMenuListener != null) {
			mTransientMenuListener.transientMenuClosed(this);
		}
	}

	/** Always returns <code>null</code>. {@inheritDoc} */
	@Override public TKMenuItem getMenuItem(int index) {
		return null;
	}

	/** Always returns 0. {@inheritDoc} */
	@Override public int getMenuItemCount() {
		return 0;
	}

	/**
	 * Builds a transient menu and shows it.
	 * 
	 * @param owner The owning panel for the transient menu.
	 * @param target The menu target for the transient menu.
	 * @param menu The menu to display.
	 * @param selectedItem The menu item that should be initially selected.
	 * @param where The location, in coordinates local to the <code>owner</code>, to display the
	 *            menu.
	 */
	public static void showTransientMenu(TKPanel owner, TKMenuTarget target, TKMenu menu, int selectedItem, Point where) {
		TKBaseWindow bWindow = owner.getBaseWindow();

		if (bWindow != null) {
			TKTransientMenu transientMenu = new TKTransientMenu(menu, owner instanceof TKTransientMenuListener ? (TKTransientMenuListener) owner : null);
			Container window = (Container) bWindow;
			Rectangle bounds = new Rectangle(0, 0, 1, 1);

			convertRectangle(bounds, owner, window);
			transientMenu.setBounds(bounds);
			window.add(transientMenu, 0);
			menu.display(transientMenu, target, new Rectangle(where.x, where.y, 1, 1), selectedItem);
		}
	}
}
