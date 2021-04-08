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

import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Insets;
import java.awt.event.ActionEvent;
import javax.swing.JComboBox;

/** An attribute bonus editor. */
public class AttributeBonusEditor extends FeatureEditor {
    private static final String CHANGE_ATTRIBUTE  = "ChangeAttribute";
    private static final String CHANGE_LIMITATION = "ChangeLimitation";

    /**
     * Create a new attribute bonus editor.
     *
     * @param row   The row this feature will belong to.
     * @param bonus The bonus to edit.
     */
    public AttributeBonusEditor(ListRow row, AttributeBonus bonus) {
        super(row, bonus);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        AttributeBonus bonus = (AttributeBonus) getFeature();

        FlexRow row = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        LeveledAmount      amount    = bonus.getAmount();
        BonusAttributeType attribute = bonus.getAttribute();
        row.add(addLeveledAmountField(amount, -999999, 999999));
        row.add(addLeveledAmountCombo(amount, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        int      length = BonusAttributeType.values().length;
        String[] names  = new String[length];
        for (int i = 0; i < length; i++) {
            names[i] = I18n.Text("to ") + BonusAttributeType.values()[i];
        }
        row.add(addComboBox(CHANGE_ATTRIBUTE, names, names[attribute.ordinal()]));
        if (BonusAttributeType.ST == attribute) {
            row.add(addComboBox(CHANGE_LIMITATION, AttributeBonusLimitation.values(), bonus.getLimitation()));
        }
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 1, 0);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (CHANGE_ATTRIBUTE.equals(command)) {
            ((AttributeBonus) getFeature()).setAttribute(BonusAttributeType.values()[((JComboBox<?>) event.getSource()).getSelectedIndex()]);
            Commitable.sendCommitToFocusOwner();
            rebuild();
        } else if (CHANGE_LIMITATION.equals(command)) {
            ((AttributeBonus) getFeature()).setLimitation((AttributeBonusLimitation) ((JComboBox<?>) event.getSource()).getSelectedItem());
        } else {
            super.actionPerformed(event);
        }
    }
}
