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
import com.trollworks.gcs.model.feature.CMBonusAttributeType;
import com.trollworks.gcs.model.feature.CMCostReduction;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.text.MessageFormat;

/** An cost reduction editor. */
public class CSCostReduction extends CSBaseFeature {
	private static final String	CHANGE_ATTRIBUTE	= "ChangeAttribute";	//$NON-NLS-1$
	private static final String	CHANGE_PERCENTAGE	= "ChangePercentage";	//$NON-NLS-1$

	/**
	 * Create a new cost reduction editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param feature The feature to edit.
	 */
	public CSCostReduction(CMRow row, CMCostReduction feature) {
		super(row, feature);
	}

	@Override protected void rebuildSelf() {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		addChangeBaseTypePopup(wrapper);
		addAttributePopup(wrapper);
		addPercentagePopup(wrapper);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);
	}

	private void addAttributePopup(TKPanel parent) {
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		for (CMBonusAttributeType one : CMCostReduction.TYPES) {
			TKMenuItem item = new TKMenuItem(one.name(), CHANGE_ATTRIBUTE);
			item.setUserObject(one);
			menu.add(item);
		}

		popup = new TKPopupMenu(menu, this, false);
		popup.setSelectedUserObject(((CMCostReduction) getFeature()).getAttribute());
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	private void addPercentagePopup(TKPanel parent) {
		TKMenu menu = new TKMenu();
		int selection = 0;
		CMCostReduction reduction = (CMCostReduction) getFeature();
		int percentage = reduction.getPercentage();
		TKPopupMenu popup;

		for (int i = 5; i <= 80; i += 5) {
			Integer key = new Integer(i);
			TKMenuItem item = new TKMenuItem(MessageFormat.format(Msgs.BY, key), CHANGE_PERCENTAGE);

			item.setUserObject(key);
			menu.add(item);
			if (i == percentage) {
				selection = i / 5 - 1;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CHANGE_ATTRIBUTE.equals(command)) {
			((CMCostReduction) getFeature()).setAttribute((CMBonusAttributeType) item.getUserObject());
		} else if (CHANGE_PERCENTAGE.equals(command)) {
			((CMCostReduction) getFeature()).setPercentage(((Integer) item.getUserObject()).intValue());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}
}
