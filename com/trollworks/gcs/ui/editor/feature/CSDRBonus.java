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
import com.trollworks.gcs.model.feature.CMDRBonus;
import com.trollworks.gcs.model.feature.CMHitLocation;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

/** A DR bonus editor. */
public class CSDRBonus extends CSBaseFeature {
	private static final String	CHANGE_LOCATION	= "ChangeLocation"; //$NON-NLS-1$

	/**
	 * Create a new DR bonus editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public CSDRBonus(CMRow row, CMDRBonus bonus) {
		super(row, bonus);
	}

	@Override protected void rebuildSelf() {
		CMDRBonus bonus = (CMDRBonus) getFeature();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		addChangeBaseTypePopup(wrapper);
		addLeveledAmountPopups(wrapper, bonus.getAmount(), 5, false);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(2));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addLocationPopup(wrapper);
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);
	}

	private void addLocationPopup(TKPanel parent) {
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		for (CMHitLocation location : CMHitLocation.values()) {
			if (location.isChoosable()) {
				TKMenuItem item = new TKMenuItem(location.toString(), CHANGE_LOCATION);
				item.setUserObject(location);
				menu.add(item);
			}
		}

		popup = new TKPopupMenu(menu, this, false);
		popup.setSelectedUserObject(((CMDRBonus) getFeature()).getLocation());
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CHANGE_LOCATION.equals(command)) {
			((CMDRBonus) getFeature()).setLocation((CMHitLocation) item.getUserObject());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}
}
