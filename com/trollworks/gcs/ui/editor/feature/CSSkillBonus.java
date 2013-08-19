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
import com.trollworks.gcs.model.feature.CMSkillBonus;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** A skill bonus editor. */
public class CSSkillBonus extends CSBaseFeature {
	private static final String	NAME_PREFIX				= "SkillName";				//$NON-NLS-1$
	private static final String	SPECIALIZATION_PREFIX	= "SkillSpecialization";	//$NON-NLS-1$

	/**
	 * Create a new skill bonus editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public CSSkillBonus(CMRow row, CMSkillBonus bonus) {
		super(row, bonus);
	}

	@Override protected void rebuildSelf() {
		CMSkillBonus bonus = (CMSkillBonus) getFeature();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		addChangeBaseTypePopup(wrapper);
		addLeveledAmountPopups(wrapper, bonus.getAmount(), 3, false);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addStringComparePopups(wrapper, bonus.getNameCriteria(), Msgs.SKILL_NAME, NAME_PREFIX);
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addStringComparePopups(wrapper, bonus.getSpecializationCriteria(), Msgs.SPECIALIZATION, SPECIALIZATION_PREFIX);
		mCenter.add(wrapper);
	}

	@Override public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(((CMSkillBonus) getFeature()).getNameCriteria(), command.substring(NAME_PREFIX.length()), event);
		} else if (command.startsWith(SPECIALIZATION_PREFIX)) {
			handleStringCompareChange(((CMSkillBonus) getFeature()).getSpecializationCriteria(), command.substring(SPECIALIZATION_PREFIX.length()), event);
		} else {
			super.actionPerformed(event);
		}
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(((CMSkillBonus) getFeature()).getNameCriteria(), command.substring(NAME_PREFIX.length()), null);
		} else if (command.startsWith(SPECIALIZATION_PREFIX)) {
			handleStringCompareChange(((CMSkillBonus) getFeature()).getSpecializationCriteria(), command.substring(SPECIALIZATION_PREFIX.length()), null);
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}
}
