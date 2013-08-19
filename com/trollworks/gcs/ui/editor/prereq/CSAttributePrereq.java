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
import com.trollworks.gcs.model.feature.CMBonusAttributeType;
import com.trollworks.gcs.model.prereq.CMAttributePrereq;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;
import java.text.MessageFormat;

/** An attribute prerequisite editor panel. */
public class CSAttributePrereq extends CSBasePrereq {
	private static final String	CHANGE_TYPE			= "ChangeType";		//$NON-NLS-1$
	private static final String	CHANGE_SECOND_TYPE	= "ChangeSecondType";	//$NON-NLS-1$
	private static final String	VALUE_PREFIX		= "Value";				//$NON-NLS-1$

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
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		for (CMBonusAttributeType type : CMAttributePrereq.TYPES) {
			TKMenuItem item = new TKMenuItem(type.getPresentationName(), CHANGE_TYPE);
			item.setUserObject(type);
			menu.add(item);
		}
		popup = new TKPopupMenu(menu, this, false);
		popup.setSelectedUserObject(((CMAttributePrereq) mPrereq).getWhich());
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	private void addChangeSecondTypePopup(TKPanel parent) {
		TKMenu menu = new TKMenu();
		CMBonusAttributeType current = ((CMAttributePrereq) mPrereq).getCombinedWith();
		TKPopupMenu popup;

		menu.add(new TKMenuItem(" ", CHANGE_SECOND_TYPE)); //$NON-NLS-1$
		for (CMBonusAttributeType type : CMAttributePrereq.TYPES) {
			TKMenuItem item = new TKMenuItem(MessageFormat.format(Msgs.COMBINED_WITH, type.getPresentationName()), CHANGE_SECOND_TYPE);
			item.setUserObject(type);
			menu.add(item);
		}

		popup = new TKPopupMenu(menu, this, false);
		if (current == null) {
			popup.setSelectedItem(0);
		} else {
			popup.setSelectedUserObject(current);
		}
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		CMAttributePrereq prereq = (CMAttributePrereq) mPrereq;

		if (CHANGE_TYPE.equals(command)) {
			prereq.setWhich((CMBonusAttributeType) item.getUserObject());
		} else if (CHANGE_SECOND_TYPE.equals(command)) {
			prereq.setCombinedWith((CMBonusAttributeType) item.getUserObject());
		} else if (command.startsWith(VALUE_PREFIX)) {
			prereq.getValueCompare().setType((CMNumericCompareType) item.getUserObject());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	@Override public void actionPerformed(ActionEvent event) {
		CMAttributePrereq prereq = (CMAttributePrereq) mPrereq;
		String command = event.getActionCommand();

		if (command.startsWith(VALUE_PREFIX)) {
			handleNumericCompareChange(prereq.getValueCompare(), event);
		} else {
			super.actionPerformed(event);
		}
	}
}
