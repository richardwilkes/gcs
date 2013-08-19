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

import com.trollworks.toolkit.io.TKFileFilter;

import java.util.List;

/**
 * Objects that want to register themselves as being able to handle open requests must implement
 * this interface.
 */
public interface TKWindowOpener {
	/** @return The file filters for this opener, for use in an open dialog. */
	public TKFileFilter[] getFileFilters();

	/**
	 * Attempt to open the specified object in a window.
	 * 
	 * @param obj The object to open.
	 * @param show Will be <code>true</code> if the window should be made visible.
	 * @param finalChance Will be <code>true</code> if this is an attempt to open an object via
	 *            the "default" opener and no more attempts will be made following this call.
	 * @param msgs A list of error and/or warning messages that can be added to for failure to open
	 *            the specified object.
	 * @return The {@link TKWindow} associated with the object that was opened.
	 */
	public TKWindow openWindow(Object obj, boolean show, boolean finalChance, List<String> msgs);
}
