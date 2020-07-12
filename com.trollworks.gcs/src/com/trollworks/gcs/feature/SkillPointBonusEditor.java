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
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Insets;

/** A skill point bonus editor. */
public class SkillPointBonusEditor extends FeatureEditor {
    /**
     * Create a new skill point bonus editor.
     *
     * @param row   The row this feature will belong to.
     * @param bonus The bonus to edit.
     */
    public SkillPointBonusEditor(ListRow row, SkillPointBonus bonus) {
        super(row, bonus);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        SkillPointBonus bonus = (SkillPointBonus) getFeature();
        int i = 0;

        FlexRow row = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        LeveledAmount amount = bonus.getAmount();
        row.add(addLeveledAmountField(amount, -999, 999));
        row.add(addLeveledAmountCombo(amount, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, i++, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        StringCriteria criteria = bonus.getNameCriteria();
        row.add(addStringCompareCombo(criteria, I18n.Text("to skills whose name ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, i++, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getSpecializationCriteria();
        row.add(addStringCompareCombo(criteria, I18n.Text("and specialization ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, i++, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getCategoryCriteria();
        row.add(addStringCompareCombo(criteria, I18n.Text("and category ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, i, 0);
    }
}
