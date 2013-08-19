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

package com.trollworks.gcs.ui.editor.prereq;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMNumericCompareType;
import com.trollworks.gcs.model.prereq.CMSpellPrereq;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** A spell prerequisite editor panel. */
public class CSSpellPrereq extends CSBasePrereq {
	private static final String	NAME_PREFIX		= "SpellName";		//$NON-NLS-1$
	private static final String	COLLEGE_PREFIX	= "SpellCollege";	//$NON-NLS-1$
	private static final String	COUNT_PREFIX	= "Count";			//$NON-NLS-1$

	/**
	 * Creates a new spell prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public CSSpellPrereq(CMRow row, CMSpellPrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override protected void rebuildSelf() {
		CMSpellPrereq prereq = (CMSpellPrereq) mPrereq;
		String type = prereq.getType();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(5));

		addHasPopup(wrapper, prereq.has());
		addNumericComparePopups(wrapper, prereq.getQuantityCriteria(), null, COUNT_PREFIX, 3);
		addChangeBaseTypePopup(wrapper);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(3));
		addChangeTypePopup(wrapper);
		if (CMSpellPrereq.TAG_NAME.equals(type)) {
			addStringComparePopups(wrapper, prereq.getStringCriteria(), "", NAME_PREFIX); //$NON-NLS-1$
		} else if (CMSpellPrereq.TAG_COLLEGE.equals(type)) {
			addStringComparePopups(wrapper, prereq.getStringCriteria(), "", COLLEGE_PREFIX); //$NON-NLS-1$
		} else {
			wrapper.add(new TKPanel());
		}
		mCenter.add(wrapper);
	}

	private void addChangeTypePopup(TKPanel parent) {
		String[] keys = { CMSpellPrereq.TAG_NAME, CMSpellPrereq.TAG_ANY, CMSpellPrereq.TAG_COLLEGE, CMSpellPrereq.TAG_COLLEGE_COUNT };
		String[] titles = { Msgs.WHOSE_SPELL_NAME, Msgs.ANY, Msgs.COLLEGE, Msgs.COLLEGE_COUNT };
		TKMenu menu = new TKMenu();
		int selection = 0;
		String current = ((CMSpellPrereq) mPrereq).getType();
		TKPopupMenu popup;

		for (int i = 0; i < keys.length; i++) {
			menu.add(new TKMenuItem(titles[i], keys[i]));
			if (current == keys[i]) {
				selection = i;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		CMSpellPrereq prereq = (CMSpellPrereq) mPrereq;

		if (CMSpellPrereq.TAG_NAME.equals(command) || CMSpellPrereq.TAG_ANY.equals(command) || CMSpellPrereq.TAG_COLLEGE.equals(command) || CMSpellPrereq.TAG_COLLEGE_COUNT.equals(command)) {
			if (!prereq.getType().equals(command)) {
				forceFocusToAccept();
				prereq.setType(command);
				rebuild();
			}
		} else if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(prereq.getStringCriteria(), command.substring(NAME_PREFIX.length()), null);
		} else if (command.startsWith(COLLEGE_PREFIX)) {
			handleStringCompareChange(prereq.getStringCriteria(), command.substring(COLLEGE_PREFIX.length()), null);
		} else if (command.startsWith(COUNT_PREFIX)) {
			prereq.getQuantityCriteria().setType((CMNumericCompareType) item.getUserObject());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	@Override public void actionPerformed(ActionEvent event) {
		CMSpellPrereq prereq = (CMSpellPrereq) mPrereq;
		String command = event.getActionCommand();

		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(prereq.getStringCriteria(), command.substring(NAME_PREFIX.length()), event);
		} else if (command.startsWith(COLLEGE_PREFIX)) {
			handleStringCompareChange(prereq.getStringCriteria(), command.substring(COLLEGE_PREFIX.length()), event);
		} else if (command.startsWith(COUNT_PREFIX)) {
			handleNumericCompareChange(prereq.getQuantityCriteria(), event);
		} else {
			super.actionPerformed(event);
		}
	}
}
