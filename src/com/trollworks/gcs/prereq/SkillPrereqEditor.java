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
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.utility.Localization;

/** A skill prerequisite editor panel. */
public class SkillPrereqEditor extends PrereqEditor {
	@Localize("whose name ")
	@Localize(locale = "de", value = "dessen Name ")
	@Localize(locale = "ru", value = "чье имя ")
	private static String	WHOSE_NAME;
	@Localize("and whose level ")
	@Localize(locale = "de", value = "und dessen Fertigkeitswert ")
	@Localize(locale = "ru", value = "и чей уровень ")
	private static String	WHOSE_LEVEL;
	@Localize("and whose specialization ")
	@Localize(locale = "de", value = "und dessen Spezialisierung ")
	@Localize(locale = "ru", value = "и чья специализация")
	private static String	WHOSE_SPECIALIZATION;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new skill prerequisite editor panel.
	 *
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public SkillPrereqEditor(ListRow row, SkillPrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override
	protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
		SkillPrereq prereq = (SkillPrereq) mPrereq;

		FlexRow row = new FlexRow();
		row.add(addHasCombo(prereq.has()));
		row.add(addChangeBaseTypeCombo());
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 1);

		row = new FlexRow();
		row.add(addStringCompareCombo(prereq.getNameCriteria(), WHOSE_NAME));
		row.add(addStringCompareField(prereq.getNameCriteria()));
		grid.add(row, 1, 1);

		row = new FlexRow();
		row.add(addStringCompareCombo(prereq.getSpecializationCriteria(), WHOSE_SPECIALIZATION));
		row.add(addStringCompareField(prereq.getSpecializationCriteria()));
		grid.add(row, 2, 1);

		row = new FlexRow();
		row.add(addNumericCompareCombo(prereq.getLevelCriteria(), WHOSE_LEVEL));
		row.add(addNumericCompareField(prereq.getLevelCriteria(), 0, 999, false));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 3, 1);
	}
}
