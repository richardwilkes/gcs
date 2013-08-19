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

package com.trollworks.toolkit.window;

import java.util.Collection;

/**
 * Objects that want to allow other objects to learn about their current selection state must
 * implement this interface.
 */
public interface TKSelectionManager {
	/**
	 * Called to obtain the current selection.
	 * 
	 * @return The current selection.
	 */
	public Collection<?> getCurrentSelection();

	/**
	 * Adds a new listener to this manager. This listener should be notified by the manager each
	 * time the selection changes.
	 * 
	 * @param listener The listener to add.
	 */
	public void addSelectionListener(TKSelectionListener listener);

	/**
	 * Removes the specified listener from the manager's listener list.
	 * 
	 * @param listener The listener to remove.
	 */
	public void removeSelectionListener(TKSelectionListener listener);

	/**
	 * Set the selection specified by the collection passed in.
	 * 
	 * @param selectInfo The selection to set.
	 */
	public void setSelection(Collection<?> selectInfo);
}
