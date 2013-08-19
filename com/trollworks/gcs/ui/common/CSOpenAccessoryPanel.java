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

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.advantage.CMAdvantageList;
import com.trollworks.gcs.model.equipment.CMEquipmentList;
import com.trollworks.gcs.model.skill.CMSkillList;
import com.trollworks.gcs.model.spell.CMSpellList;
import com.trollworks.gcs.ui.sheet.CSSheetWindow;
import com.trollworks.gcs.ui.template.CSTemplateWindow;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKTitledBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.window.TKFileDialog;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

/** Provides a panel for creating new files in the open dialog. */
public class CSOpenAccessoryPanel extends TKPanel implements ActionListener {
	private TKFileDialog	mOwner;
	private TKButton		mSheetButton;
	private TKButton		mTemplateButton;
	private TKButton		mAdvantageListButton;
	private TKButton		mSkillListButton;
	private TKButton		mSpellListButton;
	private TKButton		mEquipmentListButton;

	/**
	 * Creates a new {@link CSOpenAccessoryPanel}.
	 * 
	 * @param owner The owning {@link TKFileDialog}.
	 */
	public CSOpenAccessoryPanel(TKFileDialog owner) {
		super(new TKColumnLayout(4));
		setBorder(new TKCompoundBorder(new TKTitledBorder(Msgs.NEW_TITLE, TKFont.lookup(TKFont.TEXT_FONT_KEY)), new TKEmptyBorder(5)));

		mOwner = owner;
		add(new TKPanel());
		mSheetButton = createButton(Msgs.NEW_SHEET);
		mSkillListButton = createButton(Msgs.NEW_SKILL_LIST);
		add(new TKPanel());
		add(new TKPanel());
		mTemplateButton = createButton(Msgs.NEW_TEMPLATE);
		mSpellListButton = createButton(Msgs.NEW_SPELL_LIST);
		add(new TKPanel());
		add(new TKPanel());
		mAdvantageListButton = createButton(Msgs.NEW_ADVANTAGE_LIST);
		mEquipmentListButton = createButton(Msgs.NEW_EQUIPMENT_LIST);
		add(new TKPanel());
		adjustToSameSize(new TKPanel[] { mSheetButton, mSkillListButton, mTemplateButton, mSpellListButton, mAdvantageListButton, mEquipmentListButton });

		owner.setAccessoryPanel(this);
	}

	private TKButton createButton(String title) {
		TKButton button = new TKButton(title);

		add(button);
		button.addActionListener(this);
		return button;
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();

		if (src == mSheetButton) {
			if (mOwner.attemptClose()) {
				CSSheetWindow.displaySheetWindow(new CMCharacter());
			}
		} else if (src == mTemplateButton) {
			if (mOwner.attemptClose()) {
				CSTemplateWindow.displayTemplateWindow(new CMTemplate());
			}
		} else if (src == mAdvantageListButton) {
			if (mOwner.attemptClose()) {
				CSListWindow.displayListWindow(new CMAdvantageList());
			}
		} else if (src == mSkillListButton) {
			if (mOwner.attemptClose()) {
				CSListWindow.displayListWindow(new CMSkillList());
			}
		} else if (src == mSpellListButton) {
			if (mOwner.attemptClose()) {
				CSListWindow.displayListWindow(new CMSpellList());
			}
		} else if (src == mEquipmentListButton) {
			if (mOwner.attemptClose()) {
				CSListWindow.displayListWindow(new CMEquipmentList());
			}
		}
	}
}
