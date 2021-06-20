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

import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

/** A spell prerequisite editor panel. */
public class SpellPrereqEditor extends PrereqEditor {
    private static final String[] TYPES = {
            SpellPrereq.KEY_NAME,
            SpellPrereq.KEY_ANY,
            SpellPrereq.KEY_COLLEGE,
            SpellPrereq.KEY_COLLEGE_COUNT,
            SpellPrereq.KEY_CATEGORY
    };

    /**
     * Creates a new spell prerequisite editor panel.
     *
     * @param row    The owning row.
     * @param prereq The prerequisite to edit.
     * @param depth  The depth of this prerequisite.
     */
    public SpellPrereqEditor(ListRow row, SpellPrereq prereq, int depth) {
        super(row, prereq, depth);
    }

    @Override
    protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
        SpellPrereq prereq = (SpellPrereq) mPrereq;
        String      type   = prereq.getType();

        FlexRow row = new FlexRow();
        row.add(addHasPopup(prereq.has()));
        row.add(addNumericComparePopup(prereq.getQuantityCriteria(), null));
        row.add(addNumericCompareField(prereq.getQuantityCriteria(), 0, 999, false));
        row.add(addChangeBaseTypePopup());
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 1);

        row = new FlexRow();
        row.add(addChangeTypePopup());
        if (SpellPrereq.KEY_NAME.equals(type) || SpellPrereq.KEY_CATEGORY.equals(type) || SpellPrereq.KEY_COLLEGE.equals(type)) {
            row.add(addStringComparePopup(prereq.getStringCriteria(), ""));
            row.add(addStringCompareField(prereq.getStringCriteria()));
        } else {
            row.add(new FlexSpacer(0, 0, true, false));
        }
        grid.add(row, 1, 1);
    }

    private PopupMenu<String> addChangeTypePopup() {
        String[] titles    = {I18n.text("whose name"), I18n.text("of any kind"), I18n.text("whose college name"), I18n.text("from different colleges"), I18n.text("whose category name")};
        int      selection = 0;
        String   current   = ((SpellPrereq) mPrereq).getType();
        int      length    = TYPES.length;
        for (int i = 0; i < length; i++) {
            if (TYPES[i].equals(current)) {
                selection = i;
                break;
            }
        }
        PopupMenu<String> popup = new PopupMenu<>(titles, (p) -> {
            String      type   = TYPES[p.getSelectedIndex()];
            SpellPrereq prereq = (SpellPrereq) mPrereq;
            if (!prereq.getType().equals(type)) {
                Commitable.sendCommitToFocusOwner();
                prereq.setType(type);
                rebuild();
            }
        });
        popup.setSelectedItem(titles[selection], false);
        add(popup);
        return popup;
    }
}
