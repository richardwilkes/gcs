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
import com.trollworks.gcs.model.prereq.CMContainedWeightPrereq;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** A contained weight prerequisite editor panel. */
public class CSContainedWeightPrereq extends CSBasePrereq {
	private static final String	WEIGHT_PREFIX	= "Weight"; //$NON-NLS-1$

	/**
	 * Creates a new contained weight prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public CSContainedWeightPrereq(CMRow row, CMContainedWeightPrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override protected void rebuildSelf() {
		CMContainedWeightPrereq prereq = (CMContainedWeightPrereq) mPrereq;
		TKPanel wrapper = new TKPanel(new TKColumnLayout(3));

		addHasPopup(wrapper, prereq.has());
		addChangeBaseTypePopup(wrapper);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(3));
		addNumericComparePopups(wrapper, prereq.getWeightCompare(), Msgs.WHICH, WEIGHT_PREFIX, 15);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		CMContainedWeightPrereq prereq = (CMContainedWeightPrereq) mPrereq;

		if (command.startsWith(WEIGHT_PREFIX)) {
			prereq.getWeightCompare().setType((CMNumericCompareType) item.getUserObject());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	@Override public void actionPerformed(ActionEvent event) {
		CMContainedWeightPrereq prereq = (CMContainedWeightPrereq) mPrereq;
		String command = event.getActionCommand();

		if (command.startsWith(WEIGHT_PREFIX)) {
			handleNumericCompareChange(prereq.getWeightCompare(), event);
		} else {
			super.actionPerformed(event);
		}
	}
}
