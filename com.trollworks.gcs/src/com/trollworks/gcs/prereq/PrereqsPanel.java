/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

/** Displays and edits {@link Prereq} objects. */
public class PrereqsPanel extends BandedPanel {
    /**
     * Creates a new prerequisite editor.
     *
     * @param row     The row these prerequisites will belong to.
     * @param prereqs The initial prerequisites to display.
     */
    public PrereqsPanel(ListRow row, PrereqList prereqs) {
        super(I18n.Text("Prerequisites"));
        addPrereqs(row, new PrereqList(null, prereqs), 0);
    }

    /** @return The current prerequisite list. */
    public PrereqList getPrereqList() {
        return (PrereqList) ((ListPrereqEditor) getComponent(0)).getPrereq();
    }

    private void addPrereqs(ListRow row, PrereqList prereqs, int depth) {
        add(PrereqEditor.create(row, prereqs, depth++));
        for (Prereq prereq : prereqs.getChildren()) {
            if (prereq instanceof PrereqList) {
                addPrereqs(row, (PrereqList) prereq, depth);
            } else {
                add(PrereqEditor.create(row, prereq, depth));
            }
        }
    }
}
