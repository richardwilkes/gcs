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

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.WeightValue;

import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** An contained weight reduction editor. */
public class ContainedWeightReductionEditor extends FeatureEditor {
    /**
     * Create a new contained weight reduction editor.
     *
     * @param row     The row this feature will belong to.
     * @param feature The feature to edit.
     */
    public ContainedWeightReductionEditor(ListRow row, ContainedWeightReduction feature) {
        super(row, feature);
    }

    @Override
    protected void rebuildSelf(FlexGrid grid, FlexRow right) {
        ContainedWeightReduction feature = (ContainedWeightReduction) getFeature();
        FlexRow                  row     = new FlexRow();
        row.add(addChangeBaseTypeCombo());
        EditorField field = new EditorField(new DefaultFormatterFactory(new WeightReductionFormatter()), (event) -> {
            EditorField source = (EditorField) event.getSource();
            if ("value".equals(event.getPropertyName())) {
                feature.setValue(source.getValue());
                notifyActionListeners();
            }
        }, SwingConstants.LEFT, feature.getValue(), new WeightValue(new Fixed6(999999999), getRow().getDataFile().defaultWeightUnits()), I18n.Text("Enter a weight or percentage, e.g. \"2 lb\" or \"5%\"."));
        UIUtilities.setToPreferredSizeOnly(field);
        add(field);
        row.add(field);
        row.add(new FlexSpacer(0, 0, true, false));
        grid.add(row, 0, 0);
    }
}
