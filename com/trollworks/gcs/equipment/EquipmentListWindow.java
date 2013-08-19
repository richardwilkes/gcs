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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.common.ListWindow;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.outline.ListOutline;

/** The equipment list window. */
public class EquipmentListWindow extends ListWindow {
	private static String		MSG_UNTITLED;
	/** The extension for Equipment lists. */
	public static final String	EXTENSION					= ".eqp";					//$NON-NLS-1$
	/** The command for creating a new equipment. */
	public static final String	CMD_NEW_EQUIPMENT			= "NewEquipment";			//$NON-NLS-1$
	/** The command for creating a new equipment container. */
	public static final String	CMD_NEW_EQUIPMENT_CONTAINER	= "NewEquipmentContainer";	//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(EquipmentListWindow.class);
	}

	/**
	 * Creates a list window.
	 * 
	 * @param list The list to display.
	 */
	public EquipmentListWindow(EquipmentList list) {
		super(list, CMD_NEW_EQUIPMENT, CMD_NEW_EQUIPMENT_CONTAINER);
	}

	@Override protected ListOutline createOutline() {
		return new EquipmentOutline(mListFile, false);
	}

	@Override protected String getUntitledName() {
		return MSG_UNTITLED;
	}

	public String[] getAllowedExtensions() {
		return new String[] { EXTENSION };
	}
}
