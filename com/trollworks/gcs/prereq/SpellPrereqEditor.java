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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.CommitEnforcer;

import java.awt.event.ActionEvent;

import javax.swing.JComboBox;

/** A spell prerequisite editor panel. */
public class SpellPrereqEditor extends PrereqEditor {
	private static String			MSG_WHOSE_SPELL_NAME;
	private static String			MSG_ANY;
	private static String			MSG_COLLEGE;
	private static String			MSG_COLLEGE_COUNT;
	private static final String		CHANGE_TYPE	= "ChangeSpellType";																						//$NON-NLS-1$
	private static final String		EMPTY		= "";																										//$NON-NLS-1$
	private static final String[]	TYPES		= { SpellPrereq.TAG_NAME, SpellPrereq.TAG_ANY, SpellPrereq.TAG_COLLEGE, SpellPrereq.TAG_COLLEGE_COUNT };

	static {
		LocalizedMessages.initialize(SpellPrereqEditor.class);
	}

	/**
	 * Creates a new spell prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public SpellPrereqEditor(ListRow row, SpellPrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override
	protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
		SpellPrereq prereq = (SpellPrereq) mPrereq;
		String type = prereq.getType();

		FlexRow row = new FlexRow();
		row.add(addHasCombo(prereq.has()));
		row.add(addNumericCompareCombo(prereq.getQuantityCriteria(), null));
		row.add(addNumericCompareField(prereq.getQuantityCriteria(), 0, 999, false));
		row.add(addChangeBaseTypeCombo());
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 1);

		row = new FlexRow();
		row.add(addChangeTypePopup());
		if (SpellPrereq.TAG_NAME.equals(type)) {
			row.add(addStringCompareCombo(prereq.getStringCriteria(), EMPTY));
			row.add(addStringCompareField(prereq.getStringCriteria()));
		} else if (SpellPrereq.TAG_COLLEGE.equals(type)) {
			row.add(addStringCompareCombo(prereq.getStringCriteria(), EMPTY));
			row.add(addStringCompareField(prereq.getStringCriteria()));
		} else {
			row.add(new FlexSpacer(0, 0, true, false));
		}
		grid.add(row, 1, 1);
	}

	private JComboBox<Object> addChangeTypePopup() {
		String[] titles = { MSG_WHOSE_SPELL_NAME, MSG_ANY, MSG_COLLEGE, MSG_COLLEGE_COUNT };
		int selection = 0;
		String current = ((SpellPrereq) mPrereq).getType();
		for (int i = 0; i < TYPES.length; i++) {
			if (TYPES[i].equals(current)) {
				selection = i;
				break;
			}
		}
		return addComboBox(CHANGE_TYPE, titles, titles[selection]);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		SpellPrereq prereq = (SpellPrereq) mPrereq;
		String command = event.getActionCommand();

		if (CHANGE_TYPE.equals(command)) {
			String type = TYPES[((JComboBox<Object>) event.getSource()).getSelectedIndex()];
			if (!prereq.getType().equals(type)) {
				CommitEnforcer.forceFocusToAccept();
				prereq.setType(type);
				rebuild();
			}
		} else {
			super.actionPerformed(event);
		}
	}
}
