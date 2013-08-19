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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.common.ListWindow;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.outline.ListOutline;

/** The skills list window. */
public class SkillListWindow extends ListWindow {
	private static String		MSG_UNTITLED;
	/** The extension for Skill lists. */
	public static final String	EXTENSION				= ".skl";				//$NON-NLS-1$
	/** The command for creating a new skill. */
	public static final String	CMD_NEW_SKILL			= "NewSkill";			//$NON-NLS-1$
	/** The command for creating a new skill container. */
	public static final String	CMD_NEW_SKILL_CONTAINER	= "NewSkillContainer";	//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(SkillListWindow.class);
	}

	/**
	 * Creates a list window.
	 * 
	 * @param list The list to display.
	 */
	public SkillListWindow(SkillList list) {
		super(list, CMD_NEW_SKILL, CMD_NEW_SKILL_CONTAINER);
	}

	@Override protected ListOutline createOutline() {
		return new SkillOutline(mListFile);
	}

	@Override protected String getUntitledName() {
		return MSG_UNTITLED;
	}

	public String[] getAllowedExtensions() {
		return new String[] { EXTENSION };
	}

// @Override public boolean adjustMenuItem(String command, TKMenuItem item) {
// if (CMD_NEW_TECHNIQUE.equals(command)) {
// item.setEnabled(!mOutline.getModel().isLocked());
// } else {
// return super.adjustMenuItem(command, item);
// }
// return true;
// }
//
// @Override public boolean obeyCommand(String command, TKMenuItem item) {
// if (CMD_NEW_TECHNIQUE.equals(command)) {
// addRow(new CMTechnique(mListFile), item.getTitle());
// } else {
// return super.obeyCommand(command, item);
// }
// return true;
// }
}
