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
import com.trollworks.gcs.model.prereq.CMAttributePrereq;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** An attribute prerequisite editor panel. */
public class CSAttributePrereq extends CSBasePrereq {
	private static final String	SECOND_TYPE_PREFIX	= "SecondType"; //$NON-NLS-1$
	private static final String	VALUE_PREFIX		= "Value";		//$NON-NLS-1$

	/**
	 * Creates a new attribute prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public CSAttributePrereq(CMRow row, CMAttributePrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override protected void rebuildSelf() {
		CMAttributePrereq prereq = (CMAttributePrereq) mPrereq;
		TKPanel wrapper = new TKPanel(new TKColumnLayout(3));

		addHasPopup(wrapper, prereq.has());
		addChangeBaseTypePopup(wrapper);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(5));
		addChangeTypePopup(wrapper);
		addChangeSecondTypePopup(wrapper);
		addNumericComparePopups(wrapper, prereq.getValueCompare(), Msgs.WHICH, VALUE_PREFIX, 5);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);
	}

	private void addChangeTypePopup(TKPanel parent) {
		String[] keys = { CMAttributePrereq.ST, CMAttributePrereq.DX, CMAttributePrereq.IQ, CMAttributePrereq.HT, CMAttributePrereq.WILL };
		String[] titles = { Msgs.ST, Msgs.DX, Msgs.IQ, Msgs.HT, Msgs.WILL };
		TKMenu menu = new TKMenu();
		int selection = 0;
		String current = ((CMAttributePrereq) mPrereq).getWhich();
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

	private void addChangeSecondTypePopup(TKPanel parent) {
		String[] keys = { "None", CMAttributePrereq.ST, CMAttributePrereq.DX, CMAttributePrereq.IQ, CMAttributePrereq.HT, CMAttributePrereq.WILL }; //$NON-NLS-1$
		String[] titles = { " ", Msgs.COMBINED_WITH_ST, Msgs.COMBINED_WITH_DX, Msgs.COMBINED_WITH_IQ, Msgs.COMBINED_WITH_HT, Msgs.COMBINED_WITH_WILL }; //$NON-NLS-1$
		TKMenu menu = new TKMenu();
		int selection = 0;
		String current = ((CMAttributePrereq) mPrereq).getCombinedWith();
		TKPopupMenu popup;

		for (int i = 0; i < keys.length; i++) {
			menu.add(new TKMenuItem(titles[i], SECOND_TYPE_PREFIX + keys[i]));
			if (current == keys[i]) {
				selection = i;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		CMAttributePrereq prereq = (CMAttributePrereq) mPrereq;

		if (CMAttributePrereq.ST.equals(command) || CMAttributePrereq.DX.equals(command) || CMAttributePrereq.IQ.equals(command) || CMAttributePrereq.HT.equals(command) || CMAttributePrereq.WILL.equals(command)) {
			prereq.setWhich(command);
		} else if (command.startsWith(SECOND_TYPE_PREFIX)) {
			prereq.setCombinedWith(command.substring(SECOND_TYPE_PREFIX.length()));
		} else if (command.startsWith(VALUE_PREFIX)) {
			handleNumericCompareChange(prereq.getValueCompare(), command.substring(VALUE_PREFIX.length()), null);
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	@Override public void actionPerformed(ActionEvent event) {
		CMAttributePrereq prereq = (CMAttributePrereq) mPrereq;
		String command = event.getActionCommand();

		if (command.startsWith(VALUE_PREFIX)) {
			handleNumericCompareChange(prereq.getValueCompare(), command.substring(VALUE_PREFIX.length()), event);
		} else {
			super.actionPerformed(event);
		}
	}
}
