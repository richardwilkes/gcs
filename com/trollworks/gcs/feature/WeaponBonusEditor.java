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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.layout.FlexGrid;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.layout.FlexSpacer;
import com.trollworks.gcs.widgets.outline.ListRow;

import java.awt.Insets;

/** A weapon bonus editor. */
public class WeaponBonusEditor extends FeatureEditor {
	private static String	MSG_WEAPON_SKILL;
	private static String	MSG_RELATIVE_SKILL_LEVEL;
	private static String	MSG_SPECIALIZATION;

	static {
		LocalizedMessages.initialize(WeaponBonusEditor.class);
	}

	/**
	 * Create a new skill bonus editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public WeaponBonusEditor(ListRow row, WeaponBonus bonus) {
		super(row, bonus);
	}

	@Override protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		WeaponBonus bonus = (WeaponBonus) getFeature();

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
		row.add(addStringCompareCombo(criteria, MSG_WEAPON_SKILL));
		row.add(addStringCompareField(criteria));
		grid.add(row, 1, 0);

		row = new FlexRow();
		row.setInsets(new Insets(0, 20, 0, 0));
		criteria = bonus.getSpecializationCriteria();
		row.add(addStringCompareCombo(criteria, MSG_SPECIALIZATION));
		row.add(addStringCompareField(criteria));
		grid.add(row, 2, 0);

		row = new FlexRow();
		row.setInsets(new Insets(0, 20, 0, 0));
		IntegerCriteria levelCriteria = bonus.getLevelCriteria();
		row.add(addNumericCompareCombo(levelCriteria, MSG_RELATIVE_SKILL_LEVEL));
		row.add(addNumericCompareField(levelCriteria, -999, 999, true));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 3, 0);
	}
}
