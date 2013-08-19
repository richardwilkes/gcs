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
import com.trollworks.gcs.model.prereq.CMAdvantagePrereq;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** A (dis)advantage prerequisite editor panel. */
public class CSAdvantagePrereq extends CSBasePrereq {
	private static final String	NAME_PREFIX		= "Name";	//$NON-NLS-1$
	private static final String	NOTES_PREFIX	= "Notes";	//$NON-NLS-1$
	private static final String	LEVEL_PREFIX	= "Level";	//$NON-NLS-1$

	/**
	 * Creates a new (dis)advantage prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public CSAdvantagePrereq(CMRow row, CMAdvantagePrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override protected void rebuildSelf() {
		CMAdvantagePrereq prereq = (CMAdvantagePrereq) mPrereq;
		TKPanel wrapper = new TKPanel(new TKColumnLayout(3));

		addHasPopup(wrapper, prereq.has());
		addChangeBaseTypePopup(wrapper);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		addStringComparePopups(wrapper, prereq.getNameCriteria(), Msgs.WHOSE_NAME, NAME_PREFIX);
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		addStringComparePopups(wrapper, prereq.getNotesCriteria(), Msgs.WHOSE_NOTES, NOTES_PREFIX);
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(3));
		addNumericComparePopups(wrapper, prereq.getLevelCriteria(), Msgs.WHOSE_LEVEL, LEVEL_PREFIX, 3);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		CMAdvantagePrereq prereq = (CMAdvantagePrereq) mPrereq;

		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(prereq.getNameCriteria(), command.substring(NAME_PREFIX.length()), null);
		} else if (command.startsWith(NOTES_PREFIX)) {
			handleStringCompareChange(prereq.getNotesCriteria(), command.substring(NOTES_PREFIX.length()), null);
		} else if (command.startsWith(LEVEL_PREFIX)) {
			handleNumericCompareChange(prereq.getLevelCriteria(), command.substring(LEVEL_PREFIX.length()), null);
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	@Override public void actionPerformed(ActionEvent event) {
		CMAdvantagePrereq prereq = (CMAdvantagePrereq) mPrereq;
		String command = event.getActionCommand();

		if (command.startsWith(NAME_PREFIX)) {
			handleStringCompareChange(prereq.getNameCriteria(), command.substring(NAME_PREFIX.length()), event);
		} else if (command.startsWith(NOTES_PREFIX)) {
			handleStringCompareChange(prereq.getNotesCriteria(), command.substring(NOTES_PREFIX.length()), event);
		} else if (command.startsWith(LEVEL_PREFIX)) {
			handleNumericCompareChange(prereq.getLevelCriteria(), command.substring(LEVEL_PREFIX.length()), event);
		} else {
			super.actionPerformed(event);
		}
	}
}
