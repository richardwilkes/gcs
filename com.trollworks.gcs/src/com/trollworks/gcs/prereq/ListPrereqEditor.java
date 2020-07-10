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

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.event.ActionEvent;
import javax.swing.JComboBox;
import javax.swing.JComponent;

/** A prerequisite list editor panel. */
public class ListPrereqEditor extends PrereqEditor {
    private static       Class<?> LAST_ITEM_TYPE = AdvantagePrereq.class;
    private static final String   ANY_ALL        = "AnyAll";
    private static final String   WHEN_TL        = "WhenTL";

    /** @param type The last item type created or switched to. */
    public static void setLastItemType(Class<?> type) {
        LAST_ITEM_TYPE = type;
    }

    /**
     * Creates a new prerequisite editor panel.
     *
     * @param row    The owning row.
     * @param prereq The prerequisite to edit.
     * @param depth  The depth of this prerequisite.
     */
    public ListPrereqEditor(ListRow row, PrereqList prereq, int depth) {
        super(row, prereq, depth);
    }

    private static String mapWhenTLToString(IntegerCriteria criteria) {
        if (PrereqList.isWhenTLEnabled(criteria)) {
            switch (criteria.getType()) {
            case IS:
            default:
                return tlIs();
            case AT_LEAST:
                return tlIsAtLeast();
            case AT_MOST:
                return tlIsAtMost();
            }
        }
        return " ";
    }

    private static String tlIs() {
        return I18n.Text("When the Character's TL is");
    }

    private static String tlIsAtLeast() {
        return I18n.Text("When the Character's TL is at least");
    }

    private static String tlIsAtMost() {
        return I18n.Text("When the Character's TL is at most");
    }

    @Override
    protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
        PrereqList      prereqList     = (PrereqList) mPrereq;
        IntegerCriteria whenTLCriteria = prereqList.getWhenTLCriteria();
        left.add(addComboBox(WHEN_TL, new Object[]{" ", tlIs(), tlIsAtLeast(), tlIsAtMost()}, mapWhenTLToString(whenTLCriteria)));
        if (PrereqList.isWhenTLEnabled(whenTLCriteria)) {
            left.add(addNumericCompareField(whenTLCriteria, 0, 99, false));
        }
        String requiresAll        = I18n.Text("Requires all of:");
        String requiresAtLeastOne = I18n.Text("Requires at least one of:");
        left.add(addComboBox(ANY_ALL, new Object[]{requiresAll, requiresAtLeastOne}, prereqList.requiresAll() ? requiresAll : requiresAtLeastOne));

        grid.add(new FlexSpacer(0, 0, true, false), 0, 1);

        IconButton button = new IconButton(Images.MORE, I18n.Text("Add a prerequisite list to this list"), () -> addPrereqList());
        add(button);
        right.add(button);
        button = new IconButton(Images.ADD, I18n.Text("Add a prerequisite to this list"), () -> addPrereq());
        add(button);
        right.add(button);
    }

    private void addPrereqList() {
        addItem(new PrereqList((PrereqList) mPrereq, true));
    }

    private void addPrereq() {
        try {
            Prereq prereq;
            if (LAST_ITEM_TYPE == ContainedWeightPrereq.class) {
                prereq = new ContainedWeightPrereq((PrereqList) mPrereq, mRow.getDataFile().defaultWeightUnits());
            } else {
                prereq = (Prereq) LAST_ITEM_TYPE.getConstructor(PrereqList.class).newInstance((PrereqList) mPrereq);
            }
            addItem(prereq);
        } catch (Exception exception) {
            // Shouldn't have a failure...
            Log.error(exception);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (ANY_ALL.equals(command)) {
            ((PrereqList) mPrereq).setRequiresAll(((JComboBox<?>) event.getSource()).getSelectedIndex() == 0);
            getParent().repaint();
        } else if (WHEN_TL.equals(command)) {
            PrereqList      prereqList     = (PrereqList) mPrereq;
            IntegerCriteria whenTLCriteria = prereqList.getWhenTLCriteria();
            Object          value          = ((JComboBox<?>) event.getSource()).getSelectedItem();
            if (!mapWhenTLToString(whenTLCriteria).equals(value)) {
                if (tlIs().equals(value)) {
                    if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
                        PrereqList.setWhenTLEnabled(whenTLCriteria, true);
                    }
                    whenTLCriteria.setType(NumericCompareType.IS);
                } else if (tlIsAtLeast().equals(value)) {
                    if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
                        PrereqList.setWhenTLEnabled(whenTLCriteria, true);
                    }
                    whenTLCriteria.setType(NumericCompareType.AT_LEAST);
                } else if (tlIsAtMost().equals(value)) {
                    if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
                        PrereqList.setWhenTLEnabled(whenTLCriteria, true);
                    }
                    whenTLCriteria.setType(NumericCompareType.AT_MOST);
                } else {
                    PrereqList.setWhenTLEnabled(whenTLCriteria, false);
                }
                rebuild();
            }
        } else {
            super.actionPerformed(event);
        }
    }

    private void addItem(Prereq prereq) {
        JComponent parent = (JComponent) getParent();
        int        index  = UIUtilities.getIndexOf(parent, this);
        ((PrereqList) mPrereq).add(0, prereq);
        parent.add(create(mRow, prereq, getDepth() + 1), index + 1);
        parent.revalidate();
    }
}
