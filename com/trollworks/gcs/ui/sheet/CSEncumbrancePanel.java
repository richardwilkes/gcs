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
import com.trollworks.toolkit.notification.TKNotifierTarget;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

/** The character encumbrance panel. */
public class CSEncumbrancePanel extends CSDropPanel implements TKNotifierTarget {
	private static final Color		CURRENT_ENCUMBRANCE_COLOR	= new Color(252, 242, 196);
	/** The various encumbrance titles. */
	public static final String[]	ENCUMBRANCE_TITLES			= new String[] { Msgs.NONE, Msgs.LIGHT, Msgs.MEDIUM, Msgs.HEAVY, Msgs.EXTRA_HEAVY };
	private CMCharacter				mCharacter;
	private TKLabel[]				mMarkers;

	/**
	 * Creates a new encumbrance panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSEncumbrancePanel(CMCharacter character) {
		super(new TKColumnLayout(7, 2, 0), Msgs.ENCUMBRANCE_MOVE_DODGE, true);
		mCharacter = character;
		mMarkers = new TKLabel[CMCharacter.ENCUMBRANCE_LEVELS];
		setBorder(new TKCompoundBorder(getBorder(), new TKLineBorder(Color.white, 1, TKLineBorder.TOP_EDGE)));
		addHorizontalBackground(createHeader(Msgs.ENCUMBRANCE_LEVEL, Msgs.ENCUMBRANCE_TOOLTIP), Color.black);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(Msgs.MAX_CARRY, Msgs.MAX_CARRY_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(Msgs.MOVE, Msgs.MOVE_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(Msgs.DODGE, Msgs.DODGE_TOOLTIP);
		int current = character.getEncumbranceLevel();
		for (int i = 0; i < CMCharacter.ENCUMBRANCE_LEVELS; i++) {
			mMarkers[i] = new TKLabel(MessageFormat.format(i == current ? Msgs.CURRENT_ENCUMBRANCE_FORMAT : Msgs.ENCUMBRANCE_FORMAT, ENCUMBRANCE_TITLES[i], TKNumberUtils.format(i)), CSFont.KEY_LABEL, TKAlignment.RIGHT);

			mMarkers[i].setToolTipText(Msgs.ENCUMBRANCE_TOOLTIP);
			add(mMarkers[i]);
			if (current == i) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			}
			createDivider();
			createWeightField(character, CMCharacter.MAXIMUM_CARRY_PREFIX + i, Msgs.MAX_CARRY, Msgs.MAX_CARRY_TOOLTIP);
			createDivider();
			createField(character, CMCharacter.MOVE_PREFIX + i, Msgs.MOVE, Msgs.MOVE_TOOLTIP);
			createDivider();
			createField(character, CMCharacter.DODGE_PREFIX + i, Msgs.DODGE, Msgs.DODGE_TOOLTIP);
		}
		character.addTarget(this, CMCharacter.ID_CARRIED_WEIGHT, CMCharacter.ID_BASIC_LIFT);
	}

	private TKPanel createDivider() {
		TKPanel panel = new TKPanel();

		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
		return panel;
	}

	private TKLabel createHeader(String title, String tooltip) {
		TKLabel label = new TKLabel(title, CSFont.KEY_LABEL, TKAlignment.CENTER, true);

		label.setForeground(Color.white);
		label.setToolTipText(tooltip);
		add(label);
		return label;
	}

	private CSIntegerField createField(CMCharacter character, String key, String title, String tooltip) {
		CSIntegerField field = new CSIntegerField(character, key, false, 0, 9999, false, title);

		field.setToolTipText(tooltip);
		add(field);
		return field;
	}

	private CSWeightField createWeightField(CMCharacter character, String key, String title, String tooltip) {
		CSWeightField field = new CSWeightField(character, key, TKAlignment.RIGHT, false, title);

		field.setToolTipText(tooltip);
		add(field);
		return field;
	}

	public void handleNotification(Object producer, String type, Object data) {
		int current = mCharacter.getEncumbranceLevel();

		for (int i = 0; i < CMCharacter.ENCUMBRANCE_LEVELS; i++) {
			if (i == current) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			} else {
				removeHorizontalBackground(mMarkers[i]);
			}
		}
		repaint();
	}
}
