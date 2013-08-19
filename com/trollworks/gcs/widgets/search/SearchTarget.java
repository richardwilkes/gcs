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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.search;

import com.trollworks.gcs.menu.edit.JumpToSearchTarget;

import javax.swing.ListCellRenderer;

/** Defines the methods which must be implemented to be the target of a {@link Search} control. */
public interface SearchTarget extends JumpToSearchTarget {
	/**
	 * Called to obtain a {@link ListCellRenderer} for displaying in the drop-down list.
	 * 
	 * @return The item renderer.
	 */
	ListCellRenderer getSearchRenderer();

	/**
	 * Called to have the target search itself with the specified filter and return the matching
	 * objects.
	 * 
	 * @param filter The filter to apply.
	 * @return The matching objects.
	 */
	Object[] search(String filter);

	/**
	 * Called to have the target select the objects specified.
	 * 
	 * @param selection The objects to select.
	 */
	void searchSelect(Object[] selection);
}
