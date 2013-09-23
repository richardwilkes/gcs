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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.prereq;

import static com.trollworks.gcs.prereq.SkillPrereqEditor_LS.*;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;

@Localized({
				@LS(key = "WHOSE_NAME", msg = "whose name "),
				@LS(key = "WHOSE_LEVEL", msg = "and whose level "),
				@LS(key = "WHOSE_SPECIALIZATION", msg = "and whose specialization "),
})
/** A skill prerequisite editor panel. */
public class SkillPrereqEditor extends PrereqEditor {
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
