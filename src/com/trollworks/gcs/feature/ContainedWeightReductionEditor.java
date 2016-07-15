/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.widget.EditorField;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.units.WeightValue;

import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** An contained weight reduction editor. */
public class ContainedWeightReductionEditor extends FeatureEditor {
	@Localize("Enter a weight or percentage, e.g. \"2 lb\" or \"5%\".")
	private static String WEIGHT_OR_PERCENTAGE;

	static {
		Localization.initialize();
	}

	/**
	 * Create a new contained weight reduction editor.
	 *
	 * @param row The row this feature will belong to.
	 * @param feature The feature to edit.
	 */
	public ContainedWeightReductionEditor(ListRow row, ContainedWeightReduction feature) {
		super(row, feature);
	}

	@Override
	protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		ContainedWeightReduction feature = (ContainedWeightReduction) getFeature();
		FlexRow row = new FlexRow();
		row.add(addChangeBaseTypeCombo());
		EditorField field = new EditorField(new DefaultFormatterFactory(new WeightReductionFormatter()), (event) -> {
			EditorField source = (EditorField) event.getSource();
			if ("value".equals(event.getPropertyName())) { //$NON-NLS-1$
				feature.setValue(source.getValue());
				notifyActionListeners();
			}
		}, SwingConstants.LEFT, feature.getValue(), new WeightValue(999999999, SheetPreferences.getWeightUnits()), WEIGHT_OR_PERCENTAGE);
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		add(field);
		row.add(field);
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 0);
	}
}
