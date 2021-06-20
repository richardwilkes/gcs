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
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Insets;

/** A skill bonus editor. */
public class SkillBonusEditor extends FeatureEditor {
    /**
     * Create a new skill bonus editor.
     *
     * @param row   The row this feature will belong to.
     * @param bonus The bonus to edit.
     */
    public SkillBonusEditor(ListRow row, SkillBonus bonus) {
        super(row, bonus);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        SkillBonus bonus = (SkillBonus) getFeature();

        FlexRow row = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        LeveledAmount amount = bonus.getAmount();
        row.add(addLeveledAmountField(amount, -999, 999));
        row.add(addLeveledAmountCombo(amount, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        PopupMenu<SkillSelectionType> popup = new PopupMenu<>(SkillSelectionType.values(), (p) -> {
            boolean needRebuild = ((SkillBonus) getFeature()).setSkillSelectionType(p.getSelectedItem());
            notifyActionListeners();
            if (needRebuild) {
                rebuild();
            }
        });
        popup.setSelectedItem(bonus.getSkillSelectionType(), false);
        add(popup);
        row.add(popup);
        grid.add(row, 1, 0);
        switch (bonus.getSkillSelectionType()) {
        case WEAPONS_WITH_NAME -> rebuildWeaponsWithName(grid, row);
        case SKILLS_WITH_NAME -> rebuildSkillsWithName(grid, row);
        default -> row.add(new FlexSpacer(0, 0, true, false));
        }
    }

    private void rebuildWeaponsWithName(FlexGrid grid, FlexRow row) {
        SkillBonus     bonus    = (SkillBonus) getFeature();
        StringCriteria criteria = bonus.getNameCriteria();
        row.add(addStringComparePopup(criteria, null));
        row.add(addStringCompareField(criteria));

        int i = 2;
        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getSpecializationCriteria();
        row.add(addStringComparePopup(criteria, I18n.text("and usage ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, i++, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getCategoryCriteria();
        row.add(addStringComparePopup(criteria, I18n.text("and category ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, i, 0);
    }

    private void rebuildSkillsWithName(FlexGrid grid, FlexRow row) {
        SkillBonus     bonus    = (SkillBonus) getFeature();
        StringCriteria criteria = bonus.getNameCriteria();
        row.add(addStringComparePopup(criteria, null));
        row.add(addStringCompareField(criteria));

        int i = 2;
        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getSpecializationCriteria();
        row.add(addStringComparePopup(criteria, I18n.text("and specialization ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, i++, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getCategoryCriteria();
        row.add(addStringComparePopup(criteria, I18n.text("and category ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, i, 0);
    }
}
