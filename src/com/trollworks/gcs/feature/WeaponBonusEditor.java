/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.utility.I18n;

import java.awt.Insets;

/** A weapon bonus editor. */
public class WeaponBonusEditor extends FeatureEditor {
    /**
     * Create a new skill bonus editor.
     *
     * @param row   The row this feature will belong to.
     * @param bonus The bonus to edit.
     */
    public WeaponBonusEditor(ListRow row, WeaponBonus bonus) {
        super(row, bonus);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        WeaponBonus bonus = (WeaponBonus) getFeature();

        FlexRow row = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        LeveledAmount amount = bonus.getAmount();
        row.add(addLeveledAmountField(amount, -999, 999));
        row.add(addLeveledAmountCombo(amount, true));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        StringCriteria criteria = bonus.getNameCriteria();
        row.add(addStringCompareCombo(criteria, I18n.Text("to weapons whose required skill name ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, 1, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getSpecializationCriteria();
        row.add(addStringCompareCombo(criteria, I18n.Text("and specialization ")));
        row.add(addStringCompareField(criteria));
        grid.add(row, 2, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        IntegerCriteria levelCriteria = bonus.getLevelCriteria();
        row.add(addNumericCompareCombo(levelCriteria, I18n.Text("and relative skill level ")));
        row.add(addNumericCompareField(levelCriteria, -999, 999, true));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 3, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        criteria = bonus.getCategoryCriteria();
        row.add(addStringCompareCombo(criteria, I18n.Text("and category ")));
        row.add(addStringCompareField(criteria));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 4, 0);
    }
}
