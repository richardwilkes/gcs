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
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.prereq.CMAdvantagePrereq;
import com.trollworks.gcs.model.prereq.CMAttributePrereq;
import com.trollworks.gcs.model.prereq.CMContainedWeightPrereq;
import com.trollworks.gcs.model.prereq.CMPrereq;
import com.trollworks.gcs.model.prereq.CMPrereqList;
import com.trollworks.gcs.model.prereq.CMSkillPrereq;
import com.trollworks.gcs.model.prereq.CMSpellPrereq;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.event.ActionEvent;

/** A prerequisite list editor panel. */
public class CSListPrereq extends CSBasePrereq {
	private static String		LAST_ITEM_TYPE	= CMAdvantagePrereq.TAG_ROOT;
	private static final String	ALL				= "All";						//$NON-NLS-1$
	private static final String	ANY				= "Any";						//$NON-NLS-1$
	private static final String	ADD_PREREQ		= "AddPrereq";					//$NON-NLS-1$
	private static final String	ADD_PREREQ_LIST	= "AddPrereqList";				//$NON-NLS-1$

	/** @param type The last item type created or switched to. */
	public static void setLastItemType(String type) {
		LAST_ITEM_TYPE = type;
	}

	/**
	 * Creates a new prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public CSListPrereq(CMRow row, CMPrereqList prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override protected void rebuildSelf() {
		TKMenu menu;
		TKPopupMenu popup;

		menu = new TKMenu();
		addMenuItem(menu, Msgs.REQUIRES_ALL, ALL);
		addMenuItem(menu, Msgs.REQUIRES_ANY, ANY);
		popup = new TKPopupMenu(menu, this, false, ((CMPrereqList) mPrereq).requiresAll() ? 0 : 1);
		popup.setOnlySize(popup.getPreferredSize());
		mLeft.add(popup);

		mCenter.add(new TKPanel());

		addButton(CSImage.getMoreIcon(), ADD_PREREQ_LIST, Msgs.ADD_PREREQ_LIST_TOOLTIP);
		addButton(CSImage.getAddIcon(), ADD_PREREQ, Msgs.ADD_PREREQ_TOOLTIP);
	}

	private void addMenuItem(TKMenu menu, String title, String key) {
		menu.add(new TKMenuItem(title, key));
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (ALL.equals(command)) {
			((CMPrereqList) mPrereq).setRequiresAll(true);
			getParent().repaint();
		} else if (ANY.equals(command)) {
			((CMPrereqList) mPrereq).setRequiresAll(false);
			getParent().repaint();
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	@Override public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (ADD_PREREQ.equals(command)) {
			CMPrereqList list = (CMPrereqList) mPrereq;
			CMPrereq prereq;

			if (LAST_ITEM_TYPE == CMAttributePrereq.TAG_ROOT) {
				prereq = new CMAttributePrereq(list);
			} else if (mRow instanceof CMEquipment && LAST_ITEM_TYPE == CMContainedWeightPrereq.TAG_ROOT) {
				prereq = new CMContainedWeightPrereq(list);
			} else if (LAST_ITEM_TYPE == CMSkillPrereq.TAG_ROOT) {
				prereq = new CMSkillPrereq(list);
			} else if (LAST_ITEM_TYPE == CMSpellPrereq.TAG_ROOT) {
				prereq = new CMSpellPrereq(list);
			} else {
				// Default to an advantage prereq
				prereq = new CMAdvantagePrereq(list);
			}
			addItem(prereq);
			setLastItemType(prereq.getXMLTag());
		} else if (ADD_PREREQ_LIST.equals(command)) {
			addItem(new CMPrereqList((CMPrereqList) mPrereq, true));
		} else {
			super.actionPerformed(event);
		}
	}

	private void addItem(CMPrereq prereq) {
		TKPanel parent = (TKPanel) getParent();
		int index = parent.getIndexOf(this);

		((CMPrereqList) mPrereq).add(0, prereq);
		parent.add(create(mRow, prereq, getDepth() + 1), index + 1);
		parent.revalidate();
	}
}
