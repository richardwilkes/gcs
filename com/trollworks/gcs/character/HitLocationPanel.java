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

import static com.trollworks.gcs.character.HitLocationPanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "HIT_LOCATION", msg = "Hit Location"),
				@LS(key = "ROLL", msg = "Roll"),
				@LS(key = "ROLL_TOOLTIP", msg = "<html><body>The random roll needed to hit the <b>{0}</b> hit location</body></html>"),
				@LS(key = "LOCATION", msg = "Where"),
				@LS(key = "PENALTY", msg = "-"),
				@LS(key = "PENALTY_TITLE_TOOLTIP", msg = "The hit penalty for targeting a specific hit location"),
				@LS(key = "PENALTY_TOOLTIP", msg = "<html><body>The hit penalty for targeting the <b>{0}</b> hit location</body></html>"),
				@LS(key = "DR", msg = "DR"),
				@LS(key = "DR_TOOLTIP", msg = "<html><body>The total DR protecting the <b>{0}</b> hit location</body></html>"),
				@LS(key = "EYE", msg = "Eye"),
				@LS(key = "SKULL", msg = "Skull"),
				@LS(key = "FACE", msg = "Face"),
				@LS(key = "RIGHT_LEG", msg = "R. Leg"),
				@LS(key = "RIGHT_ARM", msg = "R. Arm"),
				@LS(key = "TORSO", msg = "Torso"),
				@LS(key = "GROIN", msg = "Groin"),
				@LS(key = "LEFT_ARM", msg = "L. Arm"),
				@LS(key = "LEFT_LEG", msg = "L. Leg"),
				@LS(key = "HAND", msg = "Hand"),
				@LS(key = "FOOT", msg = "Foot"),
				@LS(key = "NECK", msg = "Neck"),
				@LS(key = "VITALS", msg = "Vitals"),
})
/** The character hit location panel. */
public class HitLocationPanel extends DropPanel {
	/** The various hit locations. */
	public static final String[]	LOCATIONS	= new String[] { EYE, SKULL, FACE, RIGHT_LEG, RIGHT_ARM, TORSO, GROIN, LEFT_ARM, LEFT_LEG, HAND, FOOT, NECK, VITALS };
	/** The rolls needed for various hit locations. */
	public static final String[]	ROLLS		= new String[] { "-", "3-4", "5", "6-7", "8", "9-10", "11", "12", "13-14", "15", "16", "17-18", "-" };																																							//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$ //$NON-NLS-6$ //$NON-NLS-7$ //$NON-NLS-8$ //$NON-NLS-9$ //$NON-NLS-10$ //$NON-NLS-11$ //$NON-NLS-12$ //$NON-NLS-13$
	/** The to hit penalties for various hit locations. */
	public static final String[]	PENALTIES	= new String[] { "-9", "-7", "-5", "-2", "-2", "0", "-3", "-2", "-2", "-4", "-4", "-5", "-3" };																																								//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$ //$NON-NLS-6$ //$NON-NLS-7$ //$NON-NLS-8$ //$NON-NLS-9$ //$NON-NLS-10$ //$NON-NLS-11$ //$NON-NLS-12$ //$NON-NLS-13$
	/** The DR for various hit locations. */
	public static final String[]	DR_KEYS		= new String[] { Armor.ID_EYES_DR, Armor.ID_SKULL_DR, Armor.ID_FACE_DR, Armor.ID_LEG_DR, Armor.ID_ARM_DR, Armor.ID_TORSO_DR, Armor.ID_GROIN_DR, Armor.ID_ARM_DR, Armor.ID_LEG_DR, Armor.ID_HAND_DR, Armor.ID_FOOT_DR, Armor.ID_NECK_DR, Armor.ID_TORSO_DR };

	/**
	 * Creates a new hit location panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public HitLocationPanel(GURPSCharacter character) {
		super(new ColumnLayout(7, 2, 0), HIT_LOCATION);

		int i;

		Wrapper wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		PageHeader header = createHeader(wrapper, ROLL, null);
		addHorizontalBackground(header, Color.black);
		for (i = 0; i < LOCATIONS.length; i++) {
			createLabel(wrapper, ROLLS[i], MessageFormat.format(ROLL_TOOLTIP, LOCATIONS[i]), SwingConstants.CENTER);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, LOCATION, null);
		for (i = 0; i < LOCATIONS.length; i++) {
			wrapper.add(new PageLabel(LOCATIONS[i], header, SwingConstants.CENTER));
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, PENALTY, PENALTY_TITLE_TOOLTIP);
		for (i = 0; i < LOCATIONS.length; i++) {
			createLabel(wrapper, PENALTIES[i], MessageFormat.format(PENALTY_TOOLTIP, LOCATIONS[i]), SwingConstants.RIGHT);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, DR, null);
		for (i = 0; i < LOCATIONS.length; i++) {
			createDisabledField(wrapper, character, DR_KEYS[i], MessageFormat.format(DR_TOOLTIP, LOCATIONS[i]), SwingConstants.RIGHT);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);
	}

	@Override
	public Dimension getMaximumSize() {
		Dimension size = super.getMaximumSize();
		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		addVerticalBackground(panel, Color.black);
	}

	private static void createLabel(Container panel, String title, String tooltip, int alignment) {
		PageLabel label = new PageLabel(title, null);
		label.setHorizontalAlignment(alignment);
		label.setToolTipText(tooltip);
		panel.add(label);
	}
}
