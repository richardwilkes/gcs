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
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.text.MessageFormat;

/** An cost reduction editor. */
public class CostReductionEditor extends FeatureEditor {
    /**
     * Create a new cost reduction editor.
     *
     * @param row     The row this feature will belong to.
     * @param feature The feature to edit.
     */
    public CostReductionEditor(ListRow row, CostReduction feature) {
        super(row, feature);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        CostReduction feature = (CostReduction) getFeature();
        FlexRow       row     = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        row.add(addAttributePopup(getRow().getDataFile(), "%s", feature.getAttribute(), false, (p) -> {
            AttributeChoice selectedItem = p.getSelectedItem();
            if (selectedItem != null) {
                ((CostReduction) getFeature()).setAttribute(selectedItem.getAttribute());
            }
        }));
        String[] percents = new String[16];
        for (int i = 0; i < 16; i++) {
            percents[i] = MessageFormat.format(I18n.text("by {0}%"), Integer.valueOf((i + 1) * 5));
        }
        PopupMenu<String> popup = new PopupMenu<>(percents,
                (p) -> ((CostReduction) getFeature()).setPercentage((p.getSelectedIndex() + 1) * 5));
        popup.setSelectedItem(percents[Math.min(80, Math.max(0, feature.getPercentage())) / 5 - 1], false);
        add(popup);
        row.add(popup);
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);
    }
}
