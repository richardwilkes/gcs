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
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
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
		String attribute = bonus.getAttribute();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		addChangeBaseTypePopup(wrapper);
		addLeveledAmountPopups(wrapper, bonus.getAmount(), 6, CMAttributeBonus.SPEED.equals(attribute));
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(3));
		wrapper.setBorder(new TKEmptyBorder(0, 20, 0, 0));
		addAttributePopup(wrapper);
		if (CMAttributeBonus.ST.equals(attribute)) {
			addLimitationPopup(wrapper);
		}
		wrapper.add(new TKPanel());
		mCenter.add(wrapper);
	}

	private void addAttributePopup(TKPanel parent) {
		String[] keys = { CMAttributeBonus.ST, CMAttributeBonus.DX, CMAttributeBonus.IQ, CMAttributeBonus.HT, CMAttributeBonus.FP, CMAttributeBonus.HP, CMAttributeBonus.SM, CMAttributeBonus.WILL, CMAttributeBonus.PER, CMAttributeBonus.VISION, CMAttributeBonus.HEARING, CMAttributeBonus.TASTE_SMELL, CMAttributeBonus.TOUCH, CMAttributeBonus.DODGE, CMAttributeBonus.PARRY, CMAttributeBonus.BLOCK, CMAttributeBonus.SPEED, CMAttributeBonus.MOVE };
		String[] titles = { Msgs.ST, Msgs.DX, Msgs.IQ, Msgs.HT, Msgs.FP, Msgs.HP, Msgs.SM, Msgs.WILL, Msgs.PERCEPTION, Msgs.VISION, Msgs.HEARING, Msgs.TASTE_SMELL, Msgs.TOUCH, Msgs.DODGE, Msgs.PARRY, Msgs.BLOCK, Msgs.SPEED, Msgs.MOVE };
		TKMenu menu = new TKMenu();
		int selection = 0;
		CMAttributeBonus bonus = (CMAttributeBonus) getFeature();
		String attribute = bonus.getAttribute();
		TKMenuItem item;
		TKPopupMenu popup;

		for (int i = 0; i < keys.length; i++) {
			item = new TKMenuItem(titles[i], CHANGE_ATTRIBUTE);
			item.setUserObject(keys[i]);
			menu.add(item);
			if (attribute.equals(keys[i])) {
				selection = i;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	private void addLimitationPopup(TKPanel parent) {
		String[] keys = { "NoLimitation", CMAttributeBonus.STRIKING_ONLY, CMAttributeBonus.LIFTING_ONLY }; //$NON-NLS-1$
		String[] titles = { Msgs.NO_LIMITATION, Msgs.STRIKING_ONLY, Msgs.LIFTING_ONLY };
		TKMenu menu = new TKMenu();
		int selection = 0;
		CMAttributeBonus bonus = (CMAttributeBonus) getFeature();
		String limitation = bonus.getLimitation();
		TKMenuItem item;
		TKPopupMenu popup;

		for (int i = 0; i < keys.length; i++) {
			item = new TKMenuItem(titles[i], CHANGE_LIMITATION);
			item.setUserObject(keys[i]);
			menu.add(item);
			if (keys[i].equals(limitation)) {
				selection = i;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CHANGE_ATTRIBUTE.equals(command)) {
			((CMAttributeBonus) getFeature()).setAttribute((String) item.getUserObject());
			forceFocusToAccept();
			rebuild();
		} else if (CHANGE_LIMITATION.equals(command)) {
			((CMAttributeBonus) getFeature()).setLimitation((String) item.getUserObject());
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}
}
