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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Insets;

/** A spell point bonus editor. */
public class SpellPointBonusEditor extends FeatureEditor {
    /**
     * Create a new spell point bonus editor.
     *
     * @param row   The row this feature will belong to.
     * @param bonus The bonus to edit.
     */
    public SpellPointBonusEditor(ListRow row, SpellPointBonus bonus) {
        super(row, bonus);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        SpellPointBonus bonus = (SpellPointBonus) getFeature();

        FlexRow row = new FlexRow();
        row.add(addChangeBaseTypePopup());
        LeveledAmount amount = bonus.getAmount();
        row.add(addLeveledAmountField(amount, -999, 999));
        row.add(addLeveledAmountPopup(amount, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));

        PopupMenu<String> popup = new PopupMenu<>(new String[]{getMatchText(true, ""),
                getMatchText(false, SpellPointBonus.KEY_COLLEGE_NAME), getMatchText(false, ""),
                getMatchText(false, SpellPointBonus.KEY_POWER_SOURCE_NAME)}, (p) -> {
            SpellPointBonus b = (SpellPointBonus) getFeature();
            switch (p.getSelectedIndex()) {
            case 0:
            default:
                if (!b.allColleges()) {
                    Commitable.sendCommitToFocusOwner();
                    b.allColleges(true);
                    rebuild();
                }
                break;
            case 1:
                adjustMatchType(b, SpellPointBonus.KEY_COLLEGE_NAME);
                break;
            case 2:
                adjustMatchType(b, SpellPointBonus.KEY_SPELL_NAME);
                break;
            case 3:
                adjustMatchType(b, SpellPointBonus.KEY_POWER_SOURCE_NAME);
                break;
            }
        });
        popup.setSelectedItem(getMatchText(bonus.allColleges(), bonus.getMatchType()), false);
        add(popup);
        row.add(popup);
        if (bonus.allColleges()) {
            row.add(new FlexSpacer(0, 0, true, false));
        } else {
            StringCriteria criteria = bonus.getNameCriteria();
            row.add(addStringComparePopup(criteria, ""));
            row.add(addStringCompareField(criteria));
        }
        grid.add(row, 1, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        StringCriteria criteria = bonus.getCategoryCriteria();
        row.add(addStringComparePopup(criteria, I18n.text("and category ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, 2, 0);
    }

    private static String getMatchText(boolean allColleges, String matchType) {
        if (allColleges) {
            return I18n.text("to all colleges");
        }
        if (SpellPointBonus.KEY_COLLEGE_NAME.equals(matchType)) {
            return I18n.text("to the college whose name");
        }
        if (SpellPointBonus.KEY_POWER_SOURCE_NAME.equals(matchType)) {
            return I18n.text("to the power source whose name");
        }
        return I18n.text("to the spell whose name");
    }

    private void adjustMatchType(SpellPointBonus bonus, String type) {
        if (bonus.allColleges() || !type.equals(bonus.getMatchType())) {
            Commitable.sendCommitToFocusOwner();
            bonus.allColleges(false);
            bonus.setMatchType(type);
            rebuild();
        }
    }
}
