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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.ui.common.CSDropPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;

/** The character player info panel. */
public class CSPlayerInfoPanel extends CSDropPanel {
	/**
	 * Creates a new player info panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSPlayerInfoPanel(CMCharacter character) {
		super(new TKColumnLayout(2, 2, 0), Msgs.PLAYER_INFO);
		createLabelAndField(character, CMCharacter.ID_PLAYER_NAME, Msgs.PLAYER_NAME);
		createLabelAndField(character, CMCharacter.ID_CAMPAIGN, Msgs.CAMPAIGN);
		createLabelAndField(character, CMCharacter.ID_CREATED_ON, Msgs.CREATED_ON);
	}

	private void createLabelAndField(CMCharacter character, String key, String title) {
		CSField field = new CSField(character, key, null);

		add(new CSLabel(title, field));
		add(field);
	}
}
