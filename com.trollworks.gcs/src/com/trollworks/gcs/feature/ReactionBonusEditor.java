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

import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.outline.ListRow;

import java.awt.Insets;
import java.beans.PropertyChangeEvent;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

public class ReactionBonusEditor extends FeatureEditor {
    private static final String      CHANGE_SITUATION = "ChangeSituation";
    private              EditorField mSituationField;

    public ReactionBonusEditor(ListRow row, ReactionBonus bonus) {
        super(row, bonus);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        ReactionBonus bonus = (ReactionBonus) getFeature();
        FlexRow       row   = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        LeveledAmount amount = bonus.getAmount();
        row.add(addLeveledAmountField(amount, -99999, 99999));
        row.add(addLeveledAmountCombo(amount, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));

        DefaultFormatter formatter = new DefaultFormatter();
        formatter.setOverwriteMode(false);
        mSituationField = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, bonus.getSituation(), null);
        add(mSituationField);
        row.add(mSituationField);
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 1, 0);
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        Object source = event.getSource();
        if (source == mSituationField) {
            ((ReactionBonus) getFeature()).setSituation(mSituationField.getText().trim());
        } else {
            super.propertyChange(event);
        }
    }
}
