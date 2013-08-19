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

import java.util.Collection;

/** Objects that want to handle contextual menus must implement this interface. */
public interface TKContextMenuHandler {
	/**
	 * Called before a contextual menu is displayed to give each handler a chance to customize the
	 * content.
	 * 
	 * @param owner The owner of the selection.
	 * @param selection The selection.
	 */
	public void contextMenuPrepare(TKPanel owner, Collection<? extends Object> selection);

	/**
	 * Called after a contextual menu has been dealt with to give each handler a chance to clean up
	 * after itself.
	 * 
	 * @param owner The owner of the selection.
	 * @param selection The selection.
	 */
	public void contextMenuDone(TKPanel owner, Collection<? extends Object> selection);

	/**
	 * Called when a menu item has been selected.
	 * 
	 * @param command The command associated with this item.
	 * @param item The menu item that was selected.
	 * @param owner The owner of the selection.
	 * @param selection The selection.
	 * @return Should return <code>true</code> if the command was dealt with and other
	 *         {@link TKMenuTarget}s should not be invoked.
	 */
	public boolean obeyContextMenuCommand(String command, TKMenuItem item, TKPanel owner, Collection<? extends Object> selection);
}
