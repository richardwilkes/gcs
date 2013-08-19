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

package com.trollworks.toolkit.widget.outline;

import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.menu.TKContextMenuHandler;
import com.trollworks.toolkit.widget.menu.TKContextMenuManager;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.text.MessageFormat;
import java.util.Collection;

/** Provides a contextual menu for {@link TKOutlineHeader}. */
public class TKOutlineHeaderCM implements TKContextMenuHandler {
	private static final String	CMD_HIDE_COLUMN			= "OutlineHeader.HideColumn";		//$NON-NLS-1$
	private static final String	CMD_SHOW_COLUMN			= "OutlineHeader.ShowColumn";		//$NON-NLS-1$
	/** Tag used to show all columns */
	public static final String	CMD_SHOW_ALL_COLUMNS	= "OutlineHeader.ShowAllColumns";	//$NON-NLS-1$
	/** Tag used to reset the columns to their original state. */
	public static final String	CMD_RESET_COLUMNS		= "OutlineHeader.ResetColumns";	//$NON-NLS-1$
	private static boolean		INSTALLED				= false;

	private TKOutlineHeaderCM() {
		// Only allow the install() method to create one of these.
	}

	/**
	 * Used to install a single instance of {@link TKOutlineHeaderCM} to
	 * {@link TKContextMenuManager}.
	 */
	public static void install() {
		if (!INSTALLED) {
			TKContextMenuManager.addHandler(new TKOutlineHeaderCM());
			INSTALLED = true;
		}
	}

	/**
	 * @param selection The secection to trim.
	 * @return the first column in the selection.
	 */
	private TKColumn trimSelection(Collection<?> selection) {
		if (selection != null) {
			for (Object obj : selection) {
				if (obj instanceof TKColumn) {
					return (TKColumn) obj;
				}
			}
		}
		return null;
	}

	public void contextMenuPrepare(TKPanel owner, Collection<?> selection) {
		TKColumn column = trimSelection(selection);

		if (column != null) {
			TKMenu menu = TKContextMenuManager.getAllMenu();
			TKOutline outline = ((TKOutlineHeader) owner).getOwner();
			TKOutlineModel model = outline.getModel();
			Collection<TKColumn> hidden = model.getHiddenColumns();

			TKContextMenuManager.addDividerIfNecessary(menu, false);

			TKContextMenuManager.addMenuItem(menu, Msgs.HIDE_COLUMN_TITLE, CMD_HIDE_COLUMN, model.getVisibleColumnCount() > 1, false, false);
			menu.addSeparator();
			if (!hidden.isEmpty()) {
				for (TKColumn one : hidden) {
					TKMenuItem item = new TKMenuItem(MessageFormat.format(Msgs.SHOW_COLUMN_TITLE, one.getSanitizedName()), CMD_SHOW_COLUMN);

					item.setUserObject(one);
					menu.add(item);
				}
				menu.addSeparator();
				TKContextMenuManager.addMenuItem(menu, Msgs.SHOW_ALL_COLUMNS_TITLE, CMD_SHOW_ALL_COLUMNS, true, false, false);
			}
			TKContextMenuManager.addMenuItem(menu, Msgs.RESET_COLUMNS_TITLE, CMD_RESET_COLUMNS, !outline.getConfig().equals(outline.getDefaultConfig()), false, false);
		}
	}

	public void contextMenuDone(TKPanel owner, Collection<?> selection) {
		// Nothing to do...
	}

	public boolean obeyContextMenuCommand(String command, TKMenuItem item, TKPanel owner, Collection<?> selection) {
		TKOutline outline;
		TKColumn column;

		if (CMD_HIDE_COLUMN.equals(command)) {
			column = trimSelection(selection);
			column.setVisible(false);
			((TKOutlineHeader) owner).getOwner().revalidateView();
		} else if (CMD_SHOW_COLUMN.equals(command)) {
			column = (TKColumn) item.getUserObject();
			column.setVisible(true);
			((TKOutlineHeader) owner).getOwner().revalidateView();
		} else if (CMD_SHOW_ALL_COLUMNS.equals(command)) {
			outline = ((TKOutlineHeader) owner).getOwner();
			for (TKColumn column2 : outline.getModel().getColumns()) {
				column2.setVisible(true);
			}
			outline.revalidateView();
		} else if (CMD_RESET_COLUMNS.equals(command)) {
			outline = ((TKOutlineHeader) owner).getOwner();
			outline.applyConfig(outline.getDefaultConfig());
		} else {
			return false;
		}
		return true;
	}
}
