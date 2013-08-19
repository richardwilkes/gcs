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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.utility.LocalizedMessages;

import java.awt.event.ActionEvent;
import java.text.MessageFormat;

import javax.swing.JComboBox;

/** An cost reduction editor. */
public class CostReductionEditor extends FeatureEditor {
	private static String		MSG_BY;
	private static final String	CHANGE_ATTRIBUTE	= "ChangeAttribute";	//$NON-NLS-1$
	private static final String	CHANGE_PERCENTAGE	= "ChangePercentage";	//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(CostReductionEditor.class);
	}

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
			names[i] = CostReduction.TYPES[i].name();
		}
		row.add(addComboBox(CHANGE_ATTRIBUTE, names, feature.getAttribute().name()));
		String[] percents = new String[16];
		for (int i = 0; i < 16; i++) {
			percents[i] = MessageFormat.format(MSG_BY, new Integer((i + 1) * 5));
		}
		row.add(addComboBox(CHANGE_PERCENTAGE, percents, percents[Math.min(80, Math.max(0, feature.getPercentage())) / 5 - 1]));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 0);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (CHANGE_ATTRIBUTE.equals(command)) {
			((CostReduction) getFeature()).setAttribute(CostReduction.TYPES[((JComboBox) event.getSource()).getSelectedIndex()]);
		} else if (CHANGE_PERCENTAGE.equals(command)) {
			((CostReduction) getFeature()).setPercentage((((JComboBox) event.getSource()).getSelectedIndex() + 1) * 5);
		} else {
			super.actionPerformed(event);
		}
	}
}
