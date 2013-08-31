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

import static com.trollworks.gcs.character.DescriptionPanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "DESCRIPTION", msg = "Description"),
				@LS(key = "RACE", msg = "Race:"),
				@LS(key = "SIZE_MODIFIER", msg = "Size:"),
				@LS(key = "SIZE_MODIFIER_TOOLTIP", msg = "The character's size modifier"),
				@LS(key = "TECH_LEVEL", msg = "TL:"),
				@LS(key = "TECH_LEVEL_TOOLTIP", msg = "<html><body>TL0: Stone Age<br>TL1: Bronze Age<br>TL2: Iron Age<br>TL3: Medieval<br>TL4: Age of Sail<br>TL5: Industrial Revolution<br>TL6: Mechanized Age<br>TL7: Nuclear Age<br>TL8: Digital Age<br>TL9: Microtech Age<br>TL10: Robotic Age<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>"),
				@LS(key = "AGE", msg = "Age:"),
				@LS(key = "GENDER", msg = "Gender:"),
				@LS(key = "BIRTHDAY", msg = "Birthday:"),
				@LS(key = "HEIGHT", msg = "Height:"),
				@LS(key = "WEIGHT", msg = "Weight:"),
				@LS(key = "HAIR", msg = "Hair:"),
				@LS(key = "HAIR_TOOLTIP", msg = "The character's hair style and color"),
				@LS(key = "EYE_COLOR", msg = "Eyes:"),
				@LS(key = "EYE_COLOR_TOOLTIP", msg = "The character's eye color"),
				@LS(key = "SKIN_COLOR", msg = "Skin:"),
				@LS(key = "SKIN_COLOR_TOOLTIP", msg = "The character's skin color"),
				@LS(key = "HANDEDNESS", msg = "Hand:"),
				@LS(key = "HANDEDNESS_TOOLTIP", msg = "The character's preferred hand"),
})
/** The character description panel. */
public class DescriptionPanel extends DropPanel {
	/**
	 * Creates a new description panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public DescriptionPanel(GURPSCharacter character) {
		super(new ColumnLayout(5, 2, 0), DESCRIPTION);

		Wrapper wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_RACE, RACE, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_GENDER, GENDER, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_AGE, AGE, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_BIRTHDAY, BIRTHDAY, null, SwingConstants.LEFT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_HEIGHT, DescriptionPanel_LS.HEIGHT, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_WEIGHT, WEIGHT, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_SIZE_MODIFIER, SIZE_MODIFIER, SIZE_MODIFIER_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_TECH_LEVEL, TECH_LEVEL, TECH_LEVEL_TOOLTIP, SwingConstants.LEFT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_HAIR, HAIR, HAIR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_EYE_COLOR, EYE_COLOR, EYE_COLOR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_SKIN_COLOR, SKIN_COLOR, SKIN_COLOR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_HANDEDNESS, HANDEDNESS, HANDEDNESS_TOOLTIP, SwingConstants.LEFT);
		add(wrapper);
	}

	private void createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		addVerticalBackground(panel, Color.black);
	}
}
