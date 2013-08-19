/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.widgets.CommitEnforcer;
import com.trollworks.gcs.widgets.layout.FlexGrid;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.layout.FlexSpacer;
import com.trollworks.gcs.widgets.outline.ListRow;

import java.awt.Insets;
import java.awt.event.ActionEvent;

import javax.swing.JComboBox;

/** An attribute bonus editor. */
public class AttributeBonusEditor extends FeatureEditor {
	private static final String	CHANGE_ATTRIBUTE	= "ChangeAttribute";	//$NON-NLS-1$
	private static final String	CHANGE_LIMITATION	= "ChangeLimitation";	//$NON-NLS-1$

	/**
	 * Create a new attribute bonus editor.
	 * 
	 * @param row The row this feature will belong to.
	 * @param bonus The bonus to edit.
	 */
	public AttributeBonusEditor(ListRow row, AttributeBonus bonus) {
		super(row, bonus);
	}

	@Override protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		AttributeBonus bonus = (AttributeBonus) getFeature();

		FlexRow row = new FlexRow();
		row.add(addChangeBaseTypeCombo());
		LeveledAmount amount = bonus.getAmount();
		BonusAttributeType attribute = bonus.getAttribute();
		row.add(addLeveledAmountField(amount, -999999, 999999));
		row.add(addLeveledAmountCombo(amount, !attribute.isIntegerOnly()));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 0);

		row = new FlexRow();
		row.setInsets(new Insets(0, 20, 0, 0));
		row.add(addComboBox(CHANGE_ATTRIBUTE, BonusAttributeType.values(), attribute));
		if (BonusAttributeType.ST == attribute) {
			row.add(addComboBox(CHANGE_LIMITATION, AttributeBonusLimitation.values(), bonus.getLimitation()));
		}
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 1, 0);
	}

	@Override public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (CHANGE_ATTRIBUTE.equals(command)) {
			((AttributeBonus) getFeature()).setAttribute((BonusAttributeType) ((JComboBox) event.getSource()).getSelectedItem());
			CommitEnforcer.forceFocusToAccept();
			rebuild();
		} else if (CHANGE_LIMITATION.equals(command)) {
			((AttributeBonus) getFeature()).setLimitation((AttributeBonusLimitation) ((JComboBox) event.getSource()).getSelectedItem());
		} else {
			super.actionPerformed(event);
		}
	}
}
