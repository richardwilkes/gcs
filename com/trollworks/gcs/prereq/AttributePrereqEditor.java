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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.utility.LocalizedMessages;

import java.awt.event.ActionEvent;
import java.text.MessageFormat;

import javax.swing.JComboBox;

/** An attribute prerequisite editor panel. */
public class AttributePrereqEditor extends PrereqEditor {
	private static String		MSG_COMBINED_WITH;
	private static String		MSG_WHICH;
	private static final String	CHANGE_TYPE			= "ChangeType";		//$NON-NLS-1$
	private static final String	CHANGE_SECOND_TYPE	= "ChangeSecondType";	//$NON-NLS-1$
	private static final String	BLANK				= " ";					//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(AttributePrereqEditor.class);
	}

	/**
	 * Creates a new attribute prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public AttributePrereqEditor(ListRow row, AttributePrereq prereq, int depth) {
		super(row, prereq, depth);
	}

	@Override
	protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
		AttributePrereq prereq = (AttributePrereq) mPrereq;

		FlexRow row = new FlexRow();
		row.add(addHasCombo(prereq.has()));
		row.add(addChangeBaseTypeCombo());
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 0, 1);

		row = new FlexRow();
		row.add(addChangeTypePopup());
		row.add(addChangeSecondTypePopup());
		row.add(addNumericCompareCombo(prereq.getValueCompare(), MSG_WHICH));
		row.add(addNumericCompareField(prereq.getValueCompare(), 0, 99999, false));
		row.add(new FlexSpacer(0, 0, true, false));
		grid.add(row, 1, 1);
	}

	private JComboBox<Object> addChangeTypePopup() {
		BonusAttributeType[] types = AttributePrereq.TYPES;
		String[] titles = new String[types.length];
		for (int i = 0; i < types.length; i++) {
			titles[i] = types[i].getPresentationName();
		}
		return addComboBox(CHANGE_TYPE, titles, ((AttributePrereq) mPrereq).getWhich().getPresentationName());
	}

	private JComboBox<Object> addChangeSecondTypePopup() {
		BonusAttributeType current = ((AttributePrereq) mPrereq).getCombinedWith();
		BonusAttributeType[] types = AttributePrereq.TYPES;
		String[] titles = new String[types.length + 1];
		String selection = BLANK;
		titles[0] = BLANK;
		for (int i = 0; i < types.length; i++) {
			titles[i + 1] = MessageFormat.format(MSG_COMBINED_WITH, types[i].getPresentationName());
			if (current == types[i]) {
				selection = titles[i + 1];
			}
		}
		return addComboBox(CHANGE_SECOND_TYPE, titles, selection);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		AttributePrereq prereq = (AttributePrereq) mPrereq;
		String command = event.getActionCommand();

		if (CHANGE_TYPE.equals(command)) {
			prereq.setWhich(AttributePrereq.TYPES[((JComboBox<Object>) event.getSource()).getSelectedIndex()]);
		} else if (CHANGE_SECOND_TYPE.equals(command)) {
			int which = ((JComboBox<Object>) event.getSource()).getSelectedIndex();
			prereq.setCombinedWith(which == 0 ? null : AttributePrereq.TYPES[which - 1]);
		} else {
			super.actionPerformed(event);
		}
	}
}
