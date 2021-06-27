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

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import javax.swing.JComponent;

/** A prerequisite list editor panel. */
public class ListPrereqEditor extends PrereqEditor {
    private static Class<?> LAST_ITEM_TYPE = AdvantagePrereq.class;

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

    private String mapWhenTLToString() {
        PrereqList prereqList = (PrereqList) mPrereq;
        if (prereqList.isWhenTLEnabled()) {
            return switch (prereqList.getWhenTLCriteria().getType()) {
                case AT_LEAST -> tlIsAtLeast();
                case AT_MOST -> tlIsAtMost();
                default -> tlIs();
            };
        }
        return " ";
    }

    private static String tlIs() {
        return I18n.text("When the Character's TL is");
    }

    private static String tlIsAtLeast() {
        return I18n.text("When the Character's TL is at least");
    }

    private static String tlIsAtMost() {
        return I18n.text("When the Character's TL is at most");
    }

    @Override
    protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
        PopupMenu<String> popup = new PopupMenu<>(new String[]{" ", tlIs(), tlIsAtLeast(), tlIsAtMost()}, (p) -> {
            PrereqList      prereqList     = (PrereqList) mPrereq;
            IntegerCriteria whenTLCriteria = prereqList.getWhenTLCriteria();
            String          value          = p.getSelectedItem();
            if (!mapWhenTLToString().equals(value)) {
                if (tlIs().equals(value)) {
                    prereqList.setWhenTLEnabled(true);
                    whenTLCriteria.setType(NumericCompareType.IS);
                } else if (tlIsAtLeast().equals(value)) {
                    prereqList.setWhenTLEnabled(true);
                    whenTLCriteria.setType(NumericCompareType.AT_LEAST);
                } else if (tlIsAtMost().equals(value)) {
                    prereqList.setWhenTLEnabled(true);
                    whenTLCriteria.setType(NumericCompareType.AT_MOST);
                } else {
                    prereqList.setWhenTLEnabled(false);
                }
                rebuild();
            }
        });
        popup.setSelectedItem(mapWhenTLToString(), false);
        add(popup);
        left.add(popup);

        PrereqList prereqList = (PrereqList) mPrereq;
        if (prereqList.isWhenTLEnabled()) {
            left.add(addNumericCompareField(prereqList.getWhenTLCriteria(), 0, 99, false));
        }
        String requiresAll        = I18n.text("Requires all of:");
        String requiresAtLeastOne = I18n.text("Requires at least one of:");
        popup = new PopupMenu<>(new String[]{requiresAll, requiresAtLeastOne}, (p) -> {
            ((PrereqList) mPrereq).setRequiresAll(p.getSelectedIndex() == 0);
            getParent().repaint();
        });
        popup.setSelectedItem(prereqList.requiresAll() ? requiresAll : requiresAtLeastOne, false);
        add(popup);
        left.add(popup);

        grid.add(new FlexSpacer(0, 0, true, false), 0, 1);

        FontIconButton button = new FontIconButton(FontAwesome.ELLIPSIS_H, I18n.text("Add a prerequisite list to this list"), (b) -> addPrereqList());
        add(button);
        right.add(button);
        button = new FontIconButton(FontAwesome.PLUS_CIRCLE, I18n.text("Add a prerequisite to this list"), (b) -> addPrereq());
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
                prereq = new ContainedWeightPrereq((PrereqList) mPrereq, mRow.getDataFile().getSheetSettings().defaultWeightUnits());
            } else {
                PrereqList prereqList = (PrereqList) mPrereq;
                prereq = (Prereq) LAST_ITEM_TYPE.getConstructor(PrereqList.class).newInstance(prereqList);
            }
            addItem(prereq);
        } catch (Exception exception) {
            // Shouldn't have a failure...
            Log.error(exception);
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
