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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.image.ToolkitImage;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;

import java.awt.event.ActionEvent;

import javax.swing.JComboBox;
import javax.swing.JComponent;

/** A prerequisite list editor panel. */
public class ListPrereqEditor extends PrereqEditor {
	private static String		MSG_REQUIRES_ALL;
	private static String		MSG_REQUIRES_ANY;
	private static String		MSG_ADD_PREREQ_TOOLTIP;
	private static String		MSG_ADD_PREREQ_LIST_TOOLTIP;
	private static String		MSG_NO_TL_PREREQ;
	private static String		MSG_TL_IS;
	private static String		MSG_TL_IS_AT_LEAST;
	private static String		MSG_TL_IS_AT_MOST;
	private static Class<?>		LAST_ITEM_TYPE	= AdvantagePrereq.class;
	private static final String	ANY_ALL			= "AnyAll";				//$NON-NLS-1$
	private static final String	ADD_PREREQ		= "AddPrereq";				//$NON-NLS-1$
	private static final String	ADD_PREREQ_LIST	= "AddPrereqList";			//$NON-NLS-1$
	private static final String	WHEN_TL			= "WhenTL";				//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(ListPrereqEditor.class);
	}

	/** @param type The last item type created or switched to. */
	public static void setLastItemType(Class<?> type) {
		LAST_ITEM_TYPE = type;
	}

	/**
	 * Creates a new prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	public ListPrereqEditor(ListRow row, PrereqList prereq, int depth) {
		super(row, prereq, depth);
	}

	private static String mapWhenTLToString(IntegerCriteria criteria) {
		if (PrereqList.isWhenTLEnabled(criteria)) {
			switch (criteria.getType()) {
				case IS:
				default:
					return MSG_TL_IS;
				case AT_LEAST:
					return MSG_TL_IS_AT_LEAST;
				case AT_MOST:
					return MSG_TL_IS_AT_MOST;
			}
		}
		return MSG_NO_TL_PREREQ;
	}

	@Override
	protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
		PrereqList prereqList = (PrereqList) mPrereq;
		IntegerCriteria whenTLCriteria = prereqList.getWhenTLCriteria();
		left.add(addComboBox(WHEN_TL, new Object[] { MSG_NO_TL_PREREQ, MSG_TL_IS, MSG_TL_IS_AT_LEAST, MSG_TL_IS_AT_MOST }, mapWhenTLToString(whenTLCriteria)));
		if (PrereqList.isWhenTLEnabled(whenTLCriteria)) {
			left.add(addNumericCompareField(whenTLCriteria, 0, 99, false));
		}
		left.add(addComboBox(ANY_ALL, new Object[] { MSG_REQUIRES_ALL, MSG_REQUIRES_ANY }, prereqList.requiresAll() ? MSG_REQUIRES_ALL : MSG_REQUIRES_ANY));

		grid.add(new FlexSpacer(0, 0, true, false), 0, 1);

		right.add(addButton(ToolkitImage.getMoreIcon(), ADD_PREREQ_LIST, MSG_ADD_PREREQ_LIST_TOOLTIP));
		right.add(addButton(ToolkitImage.getAddIcon(), ADD_PREREQ, MSG_ADD_PREREQ_TOOLTIP));
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (ANY_ALL.equals(command)) {
			((PrereqList) mPrereq).setRequiresAll(((JComboBox<Object>) event.getSource()).getSelectedIndex() == 0);
			getParent().repaint();
		} else if (ADD_PREREQ.equals(command)) {
			try {
				addItem((Prereq) LAST_ITEM_TYPE.getConstructor(PrereqList.class).newInstance(mPrereq));
			} catch (Exception exception) {
				// Shouldn't have a failure...
				exception.printStackTrace(System.err);
			}
		} else if (ADD_PREREQ_LIST.equals(command)) {
			addItem(new PrereqList((PrereqList) mPrereq, true));
		} else if (WHEN_TL.equals(command)) {
			PrereqList prereqList = (PrereqList) mPrereq;
			IntegerCriteria whenTLCriteria = prereqList.getWhenTLCriteria();
			Object value = ((JComboBox<Object>) event.getSource()).getSelectedItem();
			if (!mapWhenTLToString(whenTLCriteria).equals(value)) {
				if (MSG_TL_IS.equals(value)) {
					if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
						PrereqList.setWhenTLEnabled(whenTLCriteria, true);
					}
					whenTLCriteria.setType(NumericCompareType.IS);
				} else if (MSG_TL_IS_AT_LEAST.equals(value)) {
					if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
						PrereqList.setWhenTLEnabled(whenTLCriteria, true);
					}
					whenTLCriteria.setType(NumericCompareType.AT_LEAST);
				} else if (MSG_TL_IS_AT_MOST.equals(value)) {
					if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
						PrereqList.setWhenTLEnabled(whenTLCriteria, true);
					}
					whenTLCriteria.setType(NumericCompareType.AT_MOST);
				} else {
					PrereqList.setWhenTLEnabled(whenTLCriteria, false);
				}
				rebuild();
			}
		} else {
			super.actionPerformed(event);
		}
	}

	private void addItem(Prereq prereq) {
		JComponent parent = (JComponent) getParent();
		int index = UIUtilities.getIndexOf(parent, this);
		((PrereqList) mPrereq).add(0, prereq);
		parent.add(create(mRow, prereq, getDepth() + 1), index + 1);
		parent.revalidate();
	}
}
