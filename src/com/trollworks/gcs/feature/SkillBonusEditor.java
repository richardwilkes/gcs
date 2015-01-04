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
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Insets;

/** A skill bonus editor. */
public class SkillBonusEditor extends FeatureEditor {
	@Localize("to skills whose name ")
	@Localize(locale = "de", value = "auf Fertigkeiten, deren Namen ")
	@Localize(locale = "ru", value = "умение с названием ")
	private static String	SKILL_NAME;
	@Localize("and specialization ")
	@Localize(locale = "de", value = "und Spezialisierung ")
	@Localize(locale = "ru", value = "и специализация ")
	private static String	SPECIALIZATION;

	static {
		Localization.initialize();
	}

	/**
	 * Create a new skill bonus editor.
	 *
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public SkillBonusEditor(ListRow row, SkillBonus bonus) {
		super(row, bonus);
	}

	@Override
	protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		SkillBonus bonus = (SkillBonus) getFeature();

		FlexRow row = new FlexRow();
		row.add(addChangeBaseTypeCombo());
		LeveledAmount amount = bonus.getAmount();
		row.add(addLeveledAmountField(amount, -999, 999));
		row.add(addLeveledAmountCombo(amount, false));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 0);

		row = new FlexRow();
		row.setInsets(new Insets(0, 20, 0, 0));
		StringCriteria criteria = bonus.getNameCriteria();
		row.add(addStringCompareCombo(criteria, SKILL_NAME));
		row.add(addStringCompareField(criteria));
		grid.add(row, 1, 0);

		row = new FlexRow();
		row.setInsets(new Insets(0, 20, 0, 0));
		criteria = bonus.getSpecializationCriteria();
		row.add(addStringCompareCombo(criteria, SPECIALIZATION));
		row.add(addStringCompareField(criteria));
		grid.add(row, 2, 0);
	}
}
