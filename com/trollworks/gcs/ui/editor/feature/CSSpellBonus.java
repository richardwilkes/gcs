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
import com.trollworks.gcs.model.feature.CMSpellBonus;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** A spell bonus editor. */
public class CSSpellBonus extends CSBaseFeature {
	private static final String	CMD_ALL_COLLEGES	= "AllColleges";	//$NON-NLS-1$
	private static final String	CMD_ONE_COLLEGE		= "OneCollege";	//$NON-NLS-1$
	private static final String	CMD_NAME			= "SpellNamed";	//$NON-NLS-1$
	private static final String	NAME_PREFIX			= "CollegeName";	//$NON-NLS-1$

	/**
	 * Create a new spell bonus editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public CSSpellBonus(CMRow row, CMSpellBonus bonus) {
		super(row, bonus);
	}

	@Override protected void rebuildSelf() {
		CMSpellBonus bonus = (CMSpellBonus) getFeature();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		addChangeBaseTypePopup(wrapper);
		addLeveledAmountPopups(wrapper, bonus.getAmount(), 3, false);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(3));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addCollegeTypePopup(wrapper);
		if (!bonus.allColleges()) {
			addStringComparePopups(wrapper, bonus.getNameCriteria(), "", NAME_PREFIX); //$NON-NLS-1$
		} else {
			wrapper.add(new TKPanel());
		}
		mCenter.add(wrapper);
	}

	private void addCollegeTypePopup(TKPanel parent) {
		CMSpellBonus bonus = (CMSpellBonus) getFeature();
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;
		int index;

		menu.add(new TKMenuItem(Msgs.ALL_COLLEGES, CMD_ALL_COLLEGES));
		menu.add(new TKMenuItem(Msgs.ONE_COLLEGE, CMD_ONE_COLLEGE));
		menu.add(new TKMenuItem(Msgs.SPELL_NAME, CMD_NAME));
		if (bonus.allColleges()) {
			index = 0;
		} else if (bonus.matchesCollegeName()) {
			index = 1;
		} else {
			index = 2;
		}
		popup = new TKPopupMenu(menu, this, false, index);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(((CMSpellBonus) getFeature()).getNameCriteria(), command.substring(NAME_PREFIX.length()), event);
		} else {
			super.actionPerformed(event);
		}
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		CMSpellBonus bonus = (CMSpellBonus) getFeature();

		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(bonus.getNameCriteria(), command.substring(NAME_PREFIX.length()), null);
		} else if (CMD_ALL_COLLEGES.equals(command)) {
			if (!bonus.allColleges()) {
				forceFocusToAccept();
				bonus.allColleges(true);
				rebuild();
			}
		} else if (CMD_ONE_COLLEGE.equals(command)) {
			if (bonus.allColleges() || !bonus.matchesCollegeName()) {
				forceFocusToAccept();
				bonus.allColleges(false);
				bonus.matchesCollegeName(true);
				rebuild();
			}
		} else if (CMD_NAME.equals(command)) {
			if (bonus.allColleges() || bonus.matchesCollegeName()) {
				forceFocusToAccept();
				bonus.allColleges(false);
				bonus.matchesCollegeName(false);
				rebuild();
			}
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}
}
