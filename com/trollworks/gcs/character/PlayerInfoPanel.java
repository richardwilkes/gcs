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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import static com.trollworks.gcs.character.PlayerInfoPanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "PLAYER_INFO", msg = "Player Information"),
				@LS(key = "PLAYER_NAME", msg = "Player:"),
				@LS(key = "CAMPAIGN", msg = "Campaign:"),
				@LS(key = "CREATED_ON", msg = "Created On:"),
})
/** The character player info panel. */
public class PlayerInfoPanel extends DropPanel {
	/**
	 * Creates a new player info panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public PlayerInfoPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0), PLAYER_INFO);
		createLabelAndField(this, character, Profile.ID_PLAYER_NAME, PLAYER_NAME, null, SwingConstants.LEFT);
		createLabelAndField(this, character, Profile.ID_CAMPAIGN, CAMPAIGN, null, SwingConstants.LEFT);
		createLabelAndField(this, character, GURPSCharacter.ID_CREATED_ON, CREATED_ON, null, SwingConstants.LEFT);
	}
}
