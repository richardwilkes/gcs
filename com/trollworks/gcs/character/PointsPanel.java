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

import static com.trollworks.gcs.character.PointsPanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.layout.RowDistribution;
import com.trollworks.ttk.notification.NotifierTarget;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "POINTS", msg = "{0} Points"),
				@LS(key = "ATTRIBUTE_POINTS", msg = "Attributes:"),
				@LS(key = "ATTRIBUTE_POINTS_TOOLTIP", msg = "A summary of all points spent on attributes for this character"),
				@LS(key = "ADVANTAGE_POINTS", msg = "Advantages:"),
				@LS(key = "ADVANTAGE_POINTS_TOOLTIP", msg = "A summary of all points spent on advantages for this character"),
				@LS(key = "DISADVANTAGE_POINTS", msg = "Disadvantages:"),
				@LS(key = "DISADVANTAGE_POINTS_TOOLTIP", msg = "A summary of all points spent on disadvantages for this character"),
				@LS(key = "QUIRK_POINTS", msg = "Quirks:"),
				@LS(key = "QUIRK_POINTS_TOOLTIP", msg = "A summary of all points spent on quirks for this character"),
				@LS(key = "SKILL_POINTS", msg = "Skills:"),
				@LS(key = "SKILL_POINTS_TOOLTIP", msg = "A summary of all points spent on skills for this character"),
				@LS(key = "SPELL_POINTS", msg = "Spells:"),
				@LS(key = "SPELL_POINTS_TOOLTIP", msg = "A summary of all points spent on spells for this character"),
				@LS(key = "RACE_POINTS", msg = "Race:"),
				@LS(key = "RACE_POINTS_TOOLTIP", msg = "A summary of all points spent on a racial package for this character"),
				@LS(key = "EARNED_POINTS", msg = "Earned:"),
				@LS(key = "EARNED_POINTS_TOOLTIP", msg = "Points that have been earned but not yet been spent"),
})
/** The character points panel. */
public class PointsPanel extends DropPanel implements NotifierTarget {
	private GURPSCharacter	mCharacter;

	/**
	 * Creates a new points panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public PointsPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), getTitle(character));
		mCharacter = character;
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ATTRIBUTE_POINTS, ATTRIBUTE_POINTS, ATTRIBUTE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ADVANTAGE_POINTS, ADVANTAGE_POINTS, ADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DISADVANTAGE_POINTS, DISADVANTAGE_POINTS, DISADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_QUIRK_POINTS, QUIRK_POINTS, QUIRK_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SKILL_POINTS, SKILL_POINTS, SKILL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SPELL_POINTS, SPELL_POINTS, SPELL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_RACE_POINTS, RACE_POINTS, RACE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndField(this, character, GURPSCharacter.ID_EARNED_POINTS, EARNED_POINTS, EARNED_POINTS_TOOLTIP, SwingConstants.RIGHT);
		mCharacter.addTarget(this, GURPSCharacter.ID_TOTAL_POINTS);
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
		addHorizontalBackground(panel, Color.black);
		panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
	}

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		getBoxedDropShadowBorder().setTitle(getTitle(mCharacter));
		invalidate();
		repaint();
	}

	private static String getTitle(GURPSCharacter character) {
		return MessageFormat.format(POINTS, Numbers.format(character.getTotalPoints()));
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
