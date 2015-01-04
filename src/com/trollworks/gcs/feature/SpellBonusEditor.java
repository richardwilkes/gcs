/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Insets;
import java.awt.event.ActionEvent;

import javax.swing.JComboBox;

/** A spell bonus editor. */
public class SpellBonusEditor extends FeatureEditor {
	@Localize("to all colleges")
	@Localize(locale = "de", value = "auf alle Schulen")
	@Localize(locale = "ru", value = "всем школам")
	private static String		ALL_COLLEGES;
	@Localize("to the college whose name")
	@Localize(locale = "de", value = "auf die Schule, deren Namen")
	@Localize(locale = "ru", value = "школе с названием")
	private static String		ONE_COLLEGE;
	@Localize("to the spell whose name")
	@Localize(locale = "de", value = "auf den Zauber, deren Namen")
	@Localize(locale = "ru", value = "заклинанию с названием")
	private static String		SPELL_NAME;
	@Localize("to the power source whose name")
	@Localize(locale = "de", value = "auf die Energiequelle, deren Namen")
	@Localize(locale = "ru", value = "источнику силы, чьё название")
	private static String		POWER_SOURCE_NAME;

	static {
		Localization.initialize();
	}

	private static final String	COLLEGE_TYPE	= "CollegeType";	//$NON-NLS-1$

	/**
	 * Create a new spell bonus editor.
	 *
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public SpellBonusEditor(ListRow row, SpellBonus bonus) {
		super(row, bonus);
	}

	@Override
	protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		SpellBonus bonus = (SpellBonus) getFeature();

		FlexRow row = new FlexRow();
		row.add(addChangeBaseTypeCombo());
		LeveledAmount amount = bonus.getAmount();
		row.add(addLeveledAmountField(amount, -999, 999));
		row.add(addLeveledAmountCombo(amount, false));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 0);

		row = new FlexRow();
		row.setInsets(new Insets(0, 20, 0, 0));

		row.add(addComboBox(COLLEGE_TYPE, new Object[] { ALL_COLLEGES, ONE_COLLEGE, SPELL_NAME, POWER_SOURCE_NAME }, getMatchText(bonus.allColleges(), bonus.getMatchType())));
		if (!bonus.allColleges()) {
			StringCriteria criteria = bonus.getNameCriteria();
			row.add(addStringCompareCombo(criteria, "")); //$NON-NLS-1$
			row.add(addStringCompareField(criteria));
		} else {
			row.add(new FlexSpacer(0, 0, true, false));
		}
		grid.add(row, 1, 0);
	}

	private static String getMatchText(boolean allColleges, String matchType) {
		if (allColleges) {
			return ALL_COLLEGES;
		}
		if (SpellBonus.TAG_COLLEGE_NAME.equals(matchType)) {
			return ONE_COLLEGE;
		}
		if (SpellBonus.TAG_POWER_SOURCE_NAME.equals(matchType)) {
			return POWER_SOURCE_NAME;
		}
		return SPELL_NAME;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (COLLEGE_TYPE.equals(command)) {
			SpellBonus bonus = (SpellBonus) getFeature();
			switch (((JComboBox<?>) event.getSource()).getSelectedIndex()) {
				case 0:
				default:
					if (!bonus.allColleges()) {
						UIUtilities.forceFocusToAccept();
						bonus.allColleges(true);
						rebuild();
					}
					break;
				case 1:
					adjustMatchType(bonus, SpellBonus.TAG_COLLEGE_NAME);
					break;
				case 2:
					adjustMatchType(bonus, SpellBonus.TAG_SPELL_NAME);
					break;
				case 3:
					adjustMatchType(bonus, SpellBonus.TAG_POWER_SOURCE_NAME);
					break;
			}
		} else {
			super.actionPerformed(event);
		}
	}

	private void adjustMatchType(SpellBonus bonus, String type) {
		if (bonus.allColleges() || !type.equals(bonus.getMatchType())) {
			UIUtilities.forceFocusToAccept();
			bonus.allColleges(false);
			bonus.setMatchType(type);
			rebuild();
		}
	}
}
