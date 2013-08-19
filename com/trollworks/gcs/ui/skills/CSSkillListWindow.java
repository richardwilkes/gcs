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

package com.trollworks.gcs.ui.skills;

import com.trollworks.gcs.model.skill.CMSkillList;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.ui.common.CSListOpener;
import com.trollworks.gcs.ui.common.CSListWindow;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

/** The skills list window. */
public class CSSkillListWindow extends CSListWindow {
	/**
	 * Creates a list window.
	 * 
	 * @param list The list to display.
	 */
	public CSSkillListWindow(CMSkillList list) {
		super(list, CMD_NEW_SKILL, CMD_NEW_SKILL_CONTAINER);
	}

	@Override protected CSOutline createOutline() {
		return new CSSkillOutline(mListFile);
	}

	@Override protected String getUntitledName() {
		return Msgs.UNTITLED;
	}

	@Override public TKFileFilter[] getFileFilters() {
		return new TKFileFilter[] { CSListOpener.FILTERS[CSListOpener.SKILL_FILTER] };
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (CMD_NEW_TECHNIQUE.equals(command)) {
			item.setEnabled(!mOutline.getModel().isLocked());
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CMD_NEW_TECHNIQUE.equals(command)) {
			addRow(new CMTechnique(mListFile), item.getTitle());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}
}
