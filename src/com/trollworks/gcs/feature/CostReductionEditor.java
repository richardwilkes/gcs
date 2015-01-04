/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.text.MessageFormat;

import javax.swing.JComboBox;

/** An cost reduction editor. */
public class CostReductionEditor extends FeatureEditor {
	@Localize("by {0}%")
	@Localize(locale = "de", value = "um {0}%")
	@Localize(locale = "ru", value = "на {0}% ")
	private static String		BY;

	static {
		Localization.initialize();
	}

	private static final String	CHANGE_ATTRIBUTE	= "ChangeAttribute";	//$NON-NLS-1$
	private static final String	CHANGE_PERCENTAGE	= "ChangePercentage";	//$NON-NLS-1$

	/**
	 * Create a new cost reduction editor.
	 *
	 * @param row The row this feature will belong to.
	 * @param feature The feature to edit.
	 */
	public CostReductionEditor(ListRow row, CostReduction feature) {
		super(row, feature);
	}

	@Override
	protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		CostReduction feature = (CostReduction) getFeature();
		FlexRow row = new FlexRow();
		row.add(addChangeBaseTypeCombo());
		String[] names = new String[CostReduction.TYPES.length];
		for (int i = 0; i < CostReduction.TYPES.length; i++) {
			names[i] = CostReduction.TYPES[i].toString();
		}
		row.add(addComboBox(CHANGE_ATTRIBUTE, names, feature.getAttribute().name()));
		String[] percents = new String[16];
		for (int i = 0; i < 16; i++) {
			percents[i] = MessageFormat.format(BY, new Integer((i + 1) * 5));
		}
		row.add(addComboBox(CHANGE_PERCENTAGE, percents, percents[Math.min(80, Math.max(0, feature.getPercentage())) / 5 - 1]));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 0);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (CHANGE_ATTRIBUTE.equals(command)) {
			((CostReduction) getFeature()).setAttribute(CostReduction.TYPES[((JComboBox<?>) event.getSource()).getSelectedIndex()]);
		} else if (CHANGE_PERCENTAGE.equals(command)) {
			((CostReduction) getFeature()).setPercentage((((JComboBox<?>) event.getSource()).getSelectedIndex() + 1) * 5);
		} else {
			super.actionPerformed(event);
		}
	}
}
