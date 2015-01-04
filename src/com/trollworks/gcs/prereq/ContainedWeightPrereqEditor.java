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

/** A contained weight prerequisite editor panel. */
public class ContainedWeightPrereqEditor extends PrereqEditor {
	@Localize("which ")
	@Localize(locale = "de", value = "die ")
	@Localize(locale = "ru", value = "который")
	private static String	WHICH;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new contained weight prerequisite editor panel.
	 *
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public ContainedWeightPrereqEditor(ListRow row, ContainedWeightPrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override
	protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
		ContainedWeightPrereq prereq = (ContainedWeightPrereq) mPrereq;

		FlexRow row = new FlexRow();
		row.add(addHasCombo(prereq.has()));
		row.add(addChangeBaseTypeCombo());
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 1);

		row = new FlexRow();
		row.add(addNumericCompareCombo(prereq.getWeightCompare(), WHICH));
		row.add(addWeightCompareField(prereq.getWeightCompare()));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 1, 1);
	}
}
