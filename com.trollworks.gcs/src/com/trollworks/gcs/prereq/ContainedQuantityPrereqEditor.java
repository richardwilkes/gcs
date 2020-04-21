/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.outline.ListRow;

/** A contained quantity prerequisite editor panel. */
public class ContainedQuantityPrereqEditor extends PrereqEditor {
    /**
     * Creates a new contained quantity prerequisite editor panel.
     *
     * @param row    The owning row.
     * @param prereq The prerequisite to edit.
     * @param depth  The depth of this prerequisite.
     */
    public ContainedQuantityPrereqEditor(ListRow row, ContainedQuantityPrereq prereq, int depth) {
        super(row, prereq, depth);
    }

    @Override
    protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
        ContainedQuantityPrereq prereq = (ContainedQuantityPrereq) mPrereq;

        FlexRow row = new FlexRow();
        row.add(addHasCombo(prereq.has()));
        row.add(addChangeBaseTypeCombo());
        row.add(addNumericCompareCombo(prereq.getQuantityCompare(), null));
        row.add(addNumericCompareField(prereq.getQuantityCompare(), 0, 999999999, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 1);
    }
}
