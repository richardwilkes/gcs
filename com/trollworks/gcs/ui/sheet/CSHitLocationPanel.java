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
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

/** The character hit location panel. */
public class CSHitLocationPanel extends CSDropPanel {
	/** The various hit locations. */
	public static final String[]	LOCATIONS	= new String[] { Msgs.EYE, Msgs.SKULL, Msgs.FACE, Msgs.RIGHT_LEG, Msgs.RIGHT_ARM, Msgs.TORSO, Msgs.GROIN, Msgs.LEFT_ARM, Msgs.LEFT_LEG, Msgs.HAND, Msgs.FOOT, Msgs.NECK, Msgs.VITALS };
	/** The rolls needed for various hit locations. */
	public static final String[]	ROLLS		= new String[] { "-", "3-4", "5", "6-7", "8", "9-10", "11", "12", "13-14", "15", "16", "17-18", "-" };																																																										//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$ //$NON-NLS-6$ //$NON-NLS-7$ //$NON-NLS-8$ //$NON-NLS-9$ //$NON-NLS-10$ //$NON-NLS-11$ //$NON-NLS-12$ //$NON-NLS-13$
	/** The to hit penalties for various hit locations. */
	public static final String[]	PENALTIES	= new String[] { "-9", "-7", "-5", "-2", "-2", "0", "-3", "-2", "-2", "-4", "-4", "-5", "-3" };																																																											//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$ //$NON-NLS-6$ //$NON-NLS-7$ //$NON-NLS-8$ //$NON-NLS-9$ //$NON-NLS-10$ //$NON-NLS-11$ //$NON-NLS-12$ //$NON-NLS-13$
	/** The DR for various hit locations. */
	public static final String[]	DR_KEYS		= new String[] { CMCharacter.ID_EYES_DR, CMCharacter.ID_SKULL_DR, CMCharacter.ID_FACE_DR, CMCharacter.ID_LEG_DR, CMCharacter.ID_ARM_DR, CMCharacter.ID_TORSO_DR, CMCharacter.ID_GROIN_DR, CMCharacter.ID_ARM_DR, CMCharacter.ID_LEG_DR, CMCharacter.ID_HAND_DR, CMCharacter.ID_FOOT_DR, CMCharacter.ID_NECK_DR, CMCharacter.ID_TORSO_DR };

	/**
	 * Creates a new hit location panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSHitLocationPanel(CMCharacter character) {
		super(new TKColumnLayout(7, 2, 0), Msgs.HIT_LOCATION);
		setBorder(new TKCompoundBorder(getBorder(), new TKLineBorder(Color.white, 1, TKLineBorder.TOP_EDGE)));

		int i;

		TKPanel wrapper = new TKPanel(new TKColumnLayout(1, 2, 0));
		addHorizontalBackground(createHeader(wrapper, Msgs.ROLL, null), Color.black);
		for (i = 0; i < LOCATIONS.length; i++) {
			createLabel(wrapper, ROLLS[i], MessageFormat.format(Msgs.ROLL_TOOLTIP, LOCATIONS[i]), TKAlignment.CENTER);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new TKPanel(new TKColumnLayout(1, 2, 0));
		createHeader(wrapper, Msgs.LOCATION, null);
		for (i = 0; i < LOCATIONS.length; i++) {
			createLabel(wrapper, LOCATIONS[i], null, TKAlignment.CENTER);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new TKPanel(new TKColumnLayout(1, 2, 0));
		createHeader(wrapper, Msgs.PENALTY, Msgs.PENALTY_TITLE_TOOLTIP);
		for (i = 0; i < LOCATIONS.length; i++) {
			createLabel(wrapper, PENALTIES[i], MessageFormat.format(Msgs.PENALTY_TOOLTIP, LOCATIONS[i]), TKAlignment.RIGHT);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new TKPanel(new TKColumnLayout(1, 2, 0));
		createHeader(wrapper, Msgs.DR, null);
		for (i = 0; i < LOCATIONS.length; i++) {
			createDisabledField(wrapper, character, DR_KEYS[i], MessageFormat.format(Msgs.DR_TOOLTIP, LOCATIONS[i]));
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = super.getMaximumSizeSelf();

		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		TKPanel panel = new TKPanel();

		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
		addVerticalBackground(panel, Color.black);
	}

	private TKLabel createHeader(TKPanel panel, String title, String tooltip) {
		TKLabel label = new TKLabel(title, CSFont.KEY_LABEL, TKAlignment.CENTER);

		label.setToolTipText(tooltip);
		label.setForeground(Color.white);
		panel.add(label);
		return label;
	}

	private TKLabel createLabel(TKPanel panel, String title, String tooltip, int alignment) {
		TKLabel label = new TKLabel(title, CSFont.KEY_LABEL, alignment);

		label.setToolTipText(tooltip);
		panel.add(label);
		return label;
	}

	private void createDisabledField(TKPanel panel, CMCharacter character, String key, String tooltip) {
		panel.add(new CSIntegerField(character, key, false, 0, 9999, false, tooltip));
	}
}
