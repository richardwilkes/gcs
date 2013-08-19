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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.CommitEnforcer;

import java.awt.Insets;
import java.awt.event.ActionEvent;

import javax.swing.JComboBox;

/** A spell bonus editor. */
public class SpellBonusEditor extends FeatureEditor {
	private static String		MSG_ALL_COLLEGES;
	private static String		MSG_ONE_COLLEGE;
	private static String		MSG_POWER_SOURCE_NAME;
	private static String		MSG_SPELL_NAME;
	private static final String	COLLEGE_TYPE	= "CollegeType";	//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(SpellBonusEditor.class);
	}

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

		row.add(addComboBox(COLLEGE_TYPE, new Object[] { MSG_ALL_COLLEGES, MSG_ONE_COLLEGE, MSG_SPELL_NAME, MSG_POWER_SOURCE_NAME }, getMatchText(bonus.allColleges(), bonus.getMatchType())));
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
			return MSG_ALL_COLLEGES;
		}
		if (SpellBonus.TAG_COLLEGE_NAME.equals(matchType)) {
			return MSG_ONE_COLLEGE;
		}
		if (SpellBonus.TAG_POWER_SOURCE_NAME.equals(matchType)) {
			return MSG_POWER_SOURCE_NAME;
		}
		return MSG_SPELL_NAME;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (COLLEGE_TYPE.equals(command)) {
			SpellBonus bonus = (SpellBonus) getFeature();
			switch (((JComboBox<Object>) event.getSource()).getSelectedIndex()) {
				case 0:
				default:
					if (!bonus.allColleges()) {
						CommitEnforcer.forceFocusToAccept();
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
			CommitEnforcer.forceFocusToAccept();
			bonus.allColleges(false);
			bonus.setMatchType(type);
			rebuild();
		}
	}
}
