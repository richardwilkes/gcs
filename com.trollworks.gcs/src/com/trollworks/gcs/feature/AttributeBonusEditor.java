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

import com.trollworks.gcs.attribute.AttributeChoice;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Insets;

/** An attribute bonus editor. */
public class AttributeBonusEditor extends FeatureEditor {
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
        FlexRow        row   = new FlexRow();
        row.add(addChangeBaseTypePopup());
        LeveledAmount amount = bonus.getAmount();
        row.add(addLeveledAmountField(amount, -999999, 999999));
        row.add(addLeveledAmountPopup(amount, false));
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);

        row = new FlexRow();
        row.setInsets(new Insets(0, 20, 0, 0));
        String attribute = bonus.getAttribute();
        row.add(addAttributePopup(getRow().getDataFile(), I18n.text("to %s"), attribute, false, (p) -> {
            AttributeChoice selectedItem = p.getSelectedItem();
            if (selectedItem != null) {
                ((AttributeBonus) getFeature()).setAttribute(getRow().getDataFile(), selectedItem.getAttribute());
                Commitable.sendCommitToFocusOwner();
                rebuild();
            }
        }));
        if ("st".equals(attribute)) {
            PopupMenu<AttributeBonusLimitation> popup = new PopupMenu<>(AttributeBonusLimitation.values(), (p) -> {
                AttributeBonusLimitation selectedItem = p.getSelectedItem();
                if (selectedItem == null) {
                    selectedItem = AttributeBonusLimitation.NONE;
                }
                ((AttributeBonus) getFeature()).setLimitation(selectedItem);
            });
            popup.setSelectedItem(bonus.getLimitation(), false);
            add(popup);
            row.add(popup);
        }
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 1, 0);
    }
}
