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

package com.trollworks.toolkit.widget.search;

import com.trollworks.toolkit.widget.TKItemRenderer;

import java.util.Collection;

/** Defines the methods which must be implemented to be the target of a {@link TKSearch} control. */
public interface TKSearchTarget {
	/**
	 * Called to obtain a {@link TKItemRenderer} for displaying in the drop-down list.
	 * 
	 * @return The item renderer.
	 */
	public TKItemRenderer getSearchRenderer();

	/**
	 * Called to have the target search itself with the specified filter and return a collection of
	 * matching objects.
	 * 
	 * @param filter The filter to apply.
	 * @return A collection of matching objects.
	 */
	public Collection<Object> search(String filter);

	/**
	 * Called to have the target select the collection of objects specified.
	 * 
	 * @param selection The collection of objects to select.
	 */
	public void searchSelect(Collection<Object> selection);
}
