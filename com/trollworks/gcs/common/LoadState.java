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

package com.trollworks.gcs.common;

import com.trollworks.gcs.weapon.OldWeapon;

import java.util.HashMap;

/** Temporary storage for data needed at load time. */
public class LoadState {
	/** The attribute used for versioning. */
	public static final String			ATTRIBUTE_VERSION	= "version";		//$NON-NLS-1$

	/** The data file version. */
	public int							mDataFileVersion;
	/** The data item version. Used for individual items within a file. */
	public int							mDataItemVersion;
	/** Whether the load is happening to restore undo state. */
	public boolean						mForUndo;
	/** Used to convert old weapon data in equipment lists. */
	public HashMap<Object, OldWeapon>	mOldWeapons			= new HashMap<>();
	/** Used to convert old equipment data. */
	public boolean						mDefaultCarried;
}
