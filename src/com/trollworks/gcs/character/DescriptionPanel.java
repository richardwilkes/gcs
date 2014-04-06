/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character description panel. */
public class DescriptionPanel extends DropPanel {
	@Localize("Description")
	private static String	DESCRIPTION;
	@Localize("Race:")
	private static String	RACE;
	@Localize("Size:")
	private static String	SIZE_MODIFIER;
	@Localize("The character's size modifier")
	private static String	SIZE_MODIFIER_TOOLTIP;
	@Localize("TL:")
	private static String	TECH_LEVEL;
	@Localize("<html><body>TL0: Stone Age<br>TL1: Bronze Age<br>TL2: Iron Age<br>TL3: Medieval<br>TL4: Age of Sail<br>TL5: Industrial Revolution<br>TL6: Mechanized Age<br>TL7: Nuclear Age<br>TL8: Digital Age<br>TL9: Microtech Age<br>TL10: Robotic Age<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>")
	private static String	TECH_LEVEL_TOOLTIP;
	@Localize("Age:")
	static String			AGE;
	@Localize("Gender:")
	static String			GENDER;
	@Localize("Birthday:")
	static String			BIRTHDAY;
	@Localize("Height:")
	static String			HEIGHT_FIELD;
	@Localize("Weight:")
	static String			WEIGHT;
	@Localize("Hair:")
	static String			HAIR;
	@Localize("The character's hair style and color")
	static String			HAIR_TOOLTIP;
	@Localize("Eyes:")
	static String			EYE_COLOR;
	@Localize("The character's eye color")
	static String			EYE_COLOR_TOOLTIP;
	@Localize("Skin:")
	static String			SKIN_COLOR;
	@Localize("The character's skin color")
	static String			SKIN_COLOR_TOOLTIP;
	@Localize("Hand:")
	static String			HANDEDNESS;
	@Localize("The character's preferred hand")
	static String			HANDEDNESS_TOOLTIP;

	static {
		Localization.initialize();
	}

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
		createLabelAndField(wrapper, character, Profile.ID_HEIGHT, HEIGHT_FIELD, null, SwingConstants.LEFT);
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
