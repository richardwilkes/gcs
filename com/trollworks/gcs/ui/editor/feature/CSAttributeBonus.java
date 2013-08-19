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
import com.trollworks.gcs.model.feature.CMAttributeBonus;
import com.trollworks.gcs.model.feature.CMAttributeBonusLimitation;
import com.trollworks.gcs.model.feature.CMBonusAttributeType;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

/** An attribute bonus editor. */
public class CSAttributeBonus extends CSBaseFeature {
	private static final String	CHANGE_ATTRIBUTE	= "ChangeAttribute";	//$NON-NLS-1$
	private static final String	CHANGE_LIMITATION	= "ChangeLimitation";	//$NON-NLS-1$

	/**
	 * Create a new attribute bonus editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public CSAttributeBonus(CMRow row, CMAttributeBonus bonus) {
		super(row, bonus);
	}

	@Override protected void rebuildSelf() {
		CMAttributeBonus bonus = (CMAttributeBonus) getFeature();
		CMBonusAttributeType attribute = bonus.getAttribute();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		addChangeBaseTypePopup(wrapper);
		addLeveledAmountPopups(wrapper, bonus.getAmount(), 6, !attribute.isIntegerOnly());
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(3));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addAttributePopup(wrapper);
		if (CMBonusAttributeType.ST == attribute) {
			addLimitationPopup(wrapper);
		}
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);
	}

	private void addAttributePopup(TKPanel parent) {
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		for (CMBonusAttributeType one : CMBonusAttributeType.values()) {
			TKMenuItem item = new TKMenuItem(one.toString(), CHANGE_ATTRIBUTE);
			item.setUserObject(one);
			menu.add(item);
		}

		popup = new TKPopupMenu(menu, this, false);
		popup.setSelectedUserObject(((CMAttributeBonus) getFeature()).getAttribute());
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	private void addLimitationPopup(TKPanel parent) {
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		for (CMAttributeBonusLimitation one : CMAttributeBonusLimitation.values()) {
			TKMenuItem item = new TKMenuItem(one.toString(), CHANGE_LIMITATION);
			item.setUserObject(one);
			menu.add(item);
		}

		popup = new TKPopupMenu(menu, this, false);
		popup.setSelectedUserObject(((CMAttributeBonus) getFeature()).getLimitation());
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CHANGE_ATTRIBUTE.equals(command)) {
			((CMAttributeBonus) getFeature()).setAttribute((CMBonusAttributeType) item.getUserObject());
			forceFocusToAccept();
			rebuild();
		} else if (CHANGE_LIMITATION.equals(command)) {
			((CMAttributeBonus) getFeature()).setLimitation((CMAttributeBonusLimitation) item.getUserObject());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}
}
