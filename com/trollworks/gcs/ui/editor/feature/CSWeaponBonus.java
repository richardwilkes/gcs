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

package com.trollworks.gcs.ui.editor.feature;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMIntegerCriteria;
import com.trollworks.gcs.model.criteria.CMNumericCriteria;
import com.trollworks.gcs.model.feature.CMWeaponBonus;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** A weapon bonus editor. */
public class CSWeaponBonus extends CSBaseFeature {
	private static final String	NAME_PREFIX				= "SkillName";				//$NON-NLS-1$
	private static final String	SPECIALIZATION_PREFIX	= "SkillSpecialization";	//$NON-NLS-1$
	private static final String	LEVEL_PREFIX			= "Level";					//$NON-NLS-1$
	private static final String	QUALIFIER_KEY			= "Qualifier";				//$NON-NLS-1$

	/**
	 * Create a new skill bonus editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public CSWeaponBonus(CMRow row, CMWeaponBonus bonus) {
		super(row, bonus);
	}

	@Override protected void rebuildSelf() {
		CMWeaponBonus bonus = (CMWeaponBonus) getFeature();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		addChangeBaseTypePopup(wrapper);
		addLeveledAmountPopups(wrapper, bonus.getAmount(), 3, false, true);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addStringComparePopups(wrapper, bonus.getNameCriteria(), Msgs.WEAPON_SKILL, NAME_PREFIX);
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addStringComparePopups(wrapper, bonus.getSpecializationCriteria(), Msgs.SPECIALIZATION, SPECIALIZATION_PREFIX);
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addNumericComparePopups(wrapper, bonus.getLevelCriteria(), LEVEL_PREFIX, 3);
		mCenter.add(wrapper);
	}

	private void addNumericComparePopups(TKPanel parent, CMIntegerCriteria compare, String keyPrefix, int maxDigits) {
		String[] keys = { CMNumericCriteria.IS, CMNumericCriteria.AT_LEAST, CMNumericCriteria.NO_MORE_THAN };
		String[] titles = { Msgs.IS, Msgs.AT_LEAST, Msgs.AT_MOST };
		TKMenu menu = new TKMenu();
		int selection = 0;
		TKPopupMenu popup;
		TKTextField field;

		for (int i = 0; i < keys.length; i++) {
			menu.add(new TKMenuItem(titles[i], keyPrefix + keys[i]));
			if (compare.getType() == keys[i]) {
				selection = i;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);

		field = new TKTextField(TKTextUtility.makeFiller(maxDigits, 'M'));
		field.setOnlySize(field.getPreferredSize());
		field.setText(compare.getQualifierAsString(true));
		field.setKeyEventFilter(new TKNumberFilter(false, true, maxDigits));
		field.setActionCommand(keyPrefix + QUALIFIER_KEY);
		field.addActionListener(this);
		parent.add(field);
	}

	@Override public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(((CMWeaponBonus) getFeature()).getNameCriteria(), command.substring(NAME_PREFIX.length()), event);
		} else if (command.startsWith(SPECIALIZATION_PREFIX)) {
			handleStringCompareChange(((CMWeaponBonus) getFeature()).getSpecializationCriteria(), command.substring(SPECIALIZATION_PREFIX.length()), event);
		} else if (command.startsWith(LEVEL_PREFIX)) {
			handleNumericCompareChange(((CMWeaponBonus) getFeature()).getLevelCriteria(), command.substring(LEVEL_PREFIX.length()), event);
		} else {
			super.actionPerformed(event);
		}
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(((CMWeaponBonus) getFeature()).getNameCriteria(), command.substring(NAME_PREFIX.length()), null);
		} else if (command.startsWith(SPECIALIZATION_PREFIX)) {
			handleStringCompareChange(((CMWeaponBonus) getFeature()).getSpecializationCriteria(), command.substring(SPECIALIZATION_PREFIX.length()), null);
		} else if (command.startsWith(LEVEL_PREFIX)) {
			handleNumericCompareChange(((CMWeaponBonus) getFeature()).getLevelCriteria(), command.substring(LEVEL_PREFIX.length()), null);
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private void handleNumericCompareChange(CMIntegerCriteria compare, String command, ActionEvent event) {
		if (CMNumericCriteria.IS.equals(command)) {
			compare.setType(CMNumericCriteria.IS);
		} else if (CMNumericCriteria.AT_LEAST.equals(command)) {
			compare.setType(CMNumericCriteria.AT_LEAST);
		} else if (CMNumericCriteria.NO_MORE_THAN.equals(command)) {
			compare.setType(CMNumericCriteria.NO_MORE_THAN);
		} else if (QUALIFIER_KEY.equals(command)) {
			compare.setQualifier(TKNumberUtils.getInteger(((TKTextField) event.getSource()).getText(), 0));
		}
	}
}
