/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

import javax.swing.JComboBox;

/** A spell prerequisite editor panel. */
public class SpellPrereqEditor extends PrereqEditor {
	@Localize("whose name")
	@Localize(locale = "de", value = "deren/dessen Name")
	@Localize(locale = "ru", value = "чьё название")
	private static String			WHOSE_SPELL_NAME;
	@Localize("of any kind")
	@Localize(locale = "de", value = "jeglicher Art")
	@Localize(locale = "ru", value = "любого вида")
	private static String			ANY;
	@Localize("whose college name")
	@Localize(locale = "de", value = "deren/dessen Schule")
	@Localize(locale = "ru", value = "чьё название школы")
	private static String			COLLEGE;
	@Localize("from different colleges")
	@Localize(locale = "de", value = "von unterschiedlichen Schulen")
	@Localize(locale = "ru", value = "из разных школ")
	private static String			COLLEGE_COUNT;

	static {
		Localization.initialize();
	}

	private static final String		CHANGE_TYPE	= "ChangeSpellType";																						//$NON-NLS-1$
	private static final String		EMPTY		= "";																										//$NON-NLS-1$
	private static final String[]	TYPES		= { SpellPrereq.TAG_NAME, SpellPrereq.TAG_ANY, SpellPrereq.TAG_COLLEGE, SpellPrereq.TAG_COLLEGE_COUNT };

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
		String[] titles = { WHOSE_SPELL_NAME, ANY, COLLEGE, COLLEGE_COUNT };
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
			String type = TYPES[((JComboBox<?>) event.getSource()).getSelectedIndex()];
			if (!prereq.getType().equals(type)) {
				UIUtilities.forceFocusToAccept();
				prereq.setType(type);
				rebuild();
			}
		} else {
			super.actionPerformed(event);
		}
	}
}
