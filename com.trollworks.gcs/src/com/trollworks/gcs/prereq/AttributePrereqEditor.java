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

import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.text.MessageFormat;
import javax.swing.JComboBox;

/** An attribute prerequisite editor panel. */
public class AttributePrereqEditor extends PrereqEditor {
    private static final String CHANGE_TYPE        = "ChangeType";
    private static final String CHANGE_SECOND_TYPE = "ChangeSecondType";
    private static final String BLANK              = " ";

    /**
     * Creates a new attribute prerequisite editor panel.
     *
     * @param row    The owning row.
     * @param prereq The prerequisite to edit.
     * @param depth  The depth of this prerequisite.
     */
    public AttributePrereqEditor(ListRow row, AttributePrereq prereq, int depth) {
        super(row, prereq, depth);
    }

    @Override
    protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
        AttributePrereq prereq = (AttributePrereq) mPrereq;

        FlexRow row = new FlexRow();
        row.add(addHasCombo(prereq.has()));
        row.add(addChangeBaseTypeCombo());
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 1);

        row = new FlexRow();
        row.add(addChangeTypePopup());
        row.add(addChangeSecondTypePopup());
        row.add(addNumericCompareCombo(prereq.getValueCompare(), I18n.Text("which ")));
        row.add(addNumericCompareField(prereq.getValueCompare(), 0, 99999, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 1, 1);
    }

    private JComboBox<Object> addChangeTypePopup() {
        BonusAttributeType[] types  = AttributePrereq.TYPES;
        int                  length = types.length;
        String[]             titles = new String[length];
        for (int i = 0; i < length; i++) {
            titles[i] = types[i].getPresentationName();
        }
        return addComboBox(CHANGE_TYPE, titles, ((AttributePrereq) mPrereq).getWhich().getPresentationName());
    }

    private JComboBox<Object> addChangeSecondTypePopup() {
        BonusAttributeType   current   = ((AttributePrereq) mPrereq).getCombinedWith();
        BonusAttributeType[] types     = AttributePrereq.TYPES;
        int                  length    = types.length;
        String[]             titles    = new String[length + 1];
        String               selection = BLANK;
        titles[0] = BLANK;
        for (int i = 0; i < length; i++) {
            titles[i + 1] = MessageFormat.format(I18n.Text("combined with {0}"), types[i].getPresentationName());
            if (current == types[i]) {
                selection = titles[i + 1];
            }
        }
        return addComboBox(CHANGE_SECOND_TYPE, titles, selection);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        AttributePrereq prereq  = (AttributePrereq) mPrereq;
        String          command = event.getActionCommand();

        if (CHANGE_TYPE.equals(command)) {
            prereq.setWhich(AttributePrereq.TYPES[((JComboBox<?>) event.getSource()).getSelectedIndex()]);
        } else if (CHANGE_SECOND_TYPE.equals(command)) {
            int which = ((JComboBox<?>) event.getSource()).getSelectedIndex();
            prereq.setCombinedWith(which == 0 ? null : AttributePrereq.TYPES[which - 1]);
        } else {
            super.actionPerformed(event);
        }
    }
}
