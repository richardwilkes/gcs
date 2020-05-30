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

import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComboBox;

/** A skill bonus editor. */
public class SkillBonusEditor extends FeatureEditor {
    private static final String SKILL_BONUS_COMPARISON = "SkillBonusComparison";

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
        String       extra = I18n.Text("to skills whose name ");
        List<String> list  = new ArrayList<>();
        list.add(I18n.Text("to this weapon"));
        StringCriteria criteria  = bonus.getNameCriteria();
        Object         selection = null;
        for (StringCompareType type : StringCompareType.values()) {
            String title = extra + type;
            list.add(title);
            if (type == criteria.getType()) {
                selection = title;
            }
        }
        if (bonus.applyToParentOnly()) {
            selection = list.get(0);
        }
        row.add(addComboBox(SKILL_BONUS_COMPARISON, list.toArray(), selection));
        if (bonus.applyToParentOnly()) {
            row.add(new FlexSpacer(0, 0, true, false));
            grid.add(row, 1, 0);
        } else {
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
            criteria = bonus.getCategoryCriteria();
            row.add(addStringCompareCombo(criteria, I18n.Text("and category ")));
            row.add(addStringCompareField(criteria));
            grid.add(row, 3, 0);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (SKILL_BONUS_COMPARISON.equals(command)) {
            SkillBonus   bonus         = (SkillBonus) getFeature();
            JComboBox<?> combo         = (JComboBox<?>) event.getSource();
            int          selectedIndex = combo.getSelectedIndex();
            boolean      wantRebuild   = bonus.setApplyToParentOnly(selectedIndex == 0);
            if (selectedIndex != 0) {
                bonus.getNameCriteria().setType(StringCompareType.values()[selectedIndex]);
            }
            notifyActionListeners();
            if (wantRebuild) {
                rebuild();
            }
        } else {
            super.actionPerformed(event);
        }
    }
}
