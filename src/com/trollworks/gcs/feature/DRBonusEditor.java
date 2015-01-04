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
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;

import java.awt.Insets;
import java.awt.event.ActionEvent;

import javax.swing.JComboBox;

/** A DR bonus editor. */
public class DRBonusEditor extends FeatureEditor {
	private static final String	CHANGE_LOCATION	= "ChangeLocation"; //$NON-NLS-1$

	/**
	 * Create a new DR bonus editor.
	 *
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public DRBonusEditor(ListRow row, DRBonus bonus) {
		super(row, bonus);
	}

	@Override
	protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		DRBonus bonus = (DRBonus) getFeature();
		FlexRow row = new FlexRow();
		row.add(addChangeBaseTypeCombo());
		LeveledAmount amount = bonus.getAmount();
		row.add(addLeveledAmountField(amount, -99999, 99999));
		row.add(addLeveledAmountCombo(amount, false));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 0);

		row = new FlexRow();
		row.setInsets(new Insets(0, 20, 0, 0));
		row.add(addComboBox(CHANGE_LOCATION, HitLocation.getChoosableLocations(), ((DRBonus) getFeature()).getLocation()));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 1, 0);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (CHANGE_LOCATION.equals(command)) {
			((DRBonus) getFeature()).setLocation((HitLocation) ((JComboBox<?>) event.getSource()).getSelectedItem());
		} else {
			super.actionPerformed(event);
		}
	}
}
