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

/** All objects which want to be targets of menu operations must implement this interface. */
public interface TKMenuTarget {
	/** Called prior to a series of menu adjustments starting. */
	public void menusWillBeAdjusted();

	/**
	 * Called before the specified menu item is displayed to allow it to be altered appropriately
	 * for the current state of the application.
	 * 
	 * @param command The command associated with the item.
	 * @param item The menu item to adjust.
	 * @return <code>true</code> if the menu item was dealt with and other {@link TKMenuTarget}s
	 *         should not be invoked.
	 */
	public boolean adjustMenuItem(String command, TKMenuItem item);

	/** Called after a series of menu adjustments has completed. */
	public void menusWereAdjusted();

	/**
	 * Called when the specified menu item has been selected.
	 * 
	 * @param command The command associated with the item.
	 * @param item The menu item that was selected.
	 * @return <code>true</code> if the command was dealt with and other {@link TKMenuTarget}s
	 *         should not be invoked.
	 */
	public boolean obeyCommand(String command, TKMenuItem item);
}
