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

package com.trollworks.gcs.character;

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.layout.ColumnLayout;

import javax.swing.SwingConstants;

/** The character identity panel. */
public class IdentityPanel extends DropPanel {
	private static String	MSG_IDENTITY;
	private static String	MSG_NAME;
	private static String	MSG_TITLE;
	private static String	MSG_RELIGION;

	static {
		LocalizedMessages.initialize(IdentityPanel.class);
	}

	/**
	 * Creates a new identity panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public IdentityPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0), MSG_IDENTITY);
		createLabelAndField(this, character, Profile.ID_NAME, MSG_NAME, null, SwingConstants.LEFT);
		createLabelAndField(this, character, Profile.ID_TITLE, MSG_TITLE, null, SwingConstants.LEFT);
		createLabelAndField(this, character, Profile.ID_RELIGION, MSG_RELIGION, null, SwingConstants.LEFT);
	}
}
