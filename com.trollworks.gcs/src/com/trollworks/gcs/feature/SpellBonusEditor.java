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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Insets;
import java.awt.event.ActionEvent;
import javax.swing.JComboBox;

/** A spell bonus editor. */
public class SpellBonusEditor extends FeatureEditor {
    private static final String COLLEGE_TYPE = "CollegeType";

    /**
     * Create a new spell bonus editor.
     *
     * @param row   The row this feature will belong to.
     * @param bonus The bonus to edit.
     */
    public SpellBonusEditor(ListRow row, SpellBonus bonus) {
        super(row, bonus);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        SpellBonus bonus = (SpellBonus) getFeature();

        FlexRow row = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        LeveledAmount amount = bonus.getAmount();
        row.add(addLeveledAmountField(amount, -999, 999));
        row.add(addLeveledAmountCombo(amount, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));

        row.add(addComboBox(COLLEGE_TYPE, new Object[]{getMatchText(true, ""), getMatchText(false, SpellBonus.TAG_COLLEGE_NAME), getMatchText(false, ""), getMatchText(false, SpellBonus.TAG_POWER_SOURCE_NAME)}, getMatchText(bonus.allColleges(), bonus.getMatchType())));
        if (bonus.allColleges()) {
            row.add(new FlexSpacer(0, 0, true, false));
        } else {
            StringCriteria criteria = bonus.getNameCriteria();
            row.add(addStringCompareCombo(criteria, ""));
            row.add(addStringCompareField(criteria));
        }
        grid.add(row, 1, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        StringCriteria criteria = bonus.getCategoryCriteria();
        row.add(addStringCompareCombo(criteria, I18n.Text("and category ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, 2, 0);
    }

    private static String getMatchText(boolean allColleges, String matchType) {
        if (allColleges) {
            return I18n.Text("to all colleges");
        }
        if (SpellBonus.TAG_COLLEGE_NAME.equals(matchType)) {
            return I18n.Text("to the college whose name");
        }
        if (SpellBonus.TAG_POWER_SOURCE_NAME.equals(matchType)) {
            return I18n.Text("to the power source whose name");
        }
        return I18n.Text("to the spell whose name");
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (COLLEGE_TYPE.equals(command)) {
            SpellBonus bonus = (SpellBonus) getFeature();
            switch (((JComboBox<?>) event.getSource()).getSelectedIndex()) {
            case 0:
            default:
                if (!bonus.allColleges()) {
                    Commitable.sendCommitToFocusOwner();
                    bonus.allColleges(true);
                    rebuild();
                }
                break;
            case 1:
                adjustMatchType(bonus, SpellBonus.TAG_COLLEGE_NAME);
                break;
            case 2:
                adjustMatchType(bonus, SpellBonus.TAG_SPELL_NAME);
                break;
            case 3:
                adjustMatchType(bonus, SpellBonus.TAG_POWER_SOURCE_NAME);
                break;
            }
        } else {
            super.actionPerformed(event);
        }
    }

    private void adjustMatchType(SpellBonus bonus, String type) {
        if (bonus.allColleges() || !type.equals(bonus.getMatchType())) {
            Commitable.sendCommitToFocusOwner();
            bonus.allColleges(false);
            bonus.setMatchType(type);
            rebuild();
        }
    }
}
