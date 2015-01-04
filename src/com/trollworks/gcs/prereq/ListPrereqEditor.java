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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

import javax.swing.JComboBox;
import javax.swing.JComponent;

/** A prerequisite list editor panel. */
public class ListPrereqEditor extends PrereqEditor {
	@Localize("Requires all of:")
	@Localize(locale = "de", value = "Benötigt alle von:")
	@Localize(locale = "ru", value = "Требует всё из:")
	private static String		REQUIRES_ALL;
	@Localize("Requires at least one of:")
	@Localize(locale = "de", value = "Benötigt mindestens einen von:")
	@Localize(locale = "ru", value = "Требует одно из:")
	private static String		REQUIRES_ANY;
	@Localize("Add a prerequisite to this list")
	@Localize(locale = "de", value = "Füge eine Bedingung zu dieser Liste hinzu")
	@Localize(locale = "ru", value = "Добавить требование в этот список")
	private static String		ADD_PREREQ_TOOLTIP;
	@Localize("Add a prerequisite list to this list")
	@Localize(locale = "de", value = "Füge eine Bedingungs-Liste zu dieser Liste hinzu")
	@Localize(locale = "ru", value = "Добавить список требований в этот список")
	private static String		ADD_PREREQ_LIST_TOOLTIP;
	@Localize(" ")
	@Localize(locale = "de", value = " ")
	private static String		NO_TL_PREREQ;
	@Localize("When the Character's TL is")
	@Localize(locale = "de", value = "Wenn der TL des Charakters ist")
	@Localize(locale = "ru", value = "Когда ТУ персонажа")
	private static String		TL_IS;
	@Localize("When the Character's TL is at least")
	@Localize(locale = "de", value = "Wenn der TL des Charakters ist mindestens")
	@Localize(locale = "ru", value = "Когда ТУ персонажа по крайней мере")
	private static String		TL_IS_AT_LEAST;
	@Localize("When the Character's TL is at most")
	@Localize(locale = "de", value = "Wenn der TL des Charakters ist höchstens")
	@Localize(locale = "ru", value = "Когда ТУ персонажа не более")
	private static String		TL_IS_AT_MOST;

	static {
		Localization.initialize();
	}

	private static Class<?>		LAST_ITEM_TYPE	= AdvantagePrereq.class;
	private static final String	ANY_ALL			= "AnyAll";				//$NON-NLS-1$
	private static final String	WHEN_TL			= "WhenTL";				//$NON-NLS-1$

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
					return TL_IS;
				case AT_LEAST:
					return TL_IS_AT_LEAST;
				case AT_MOST:
					return TL_IS_AT_MOST;
			}
		}
		return NO_TL_PREREQ;
	}

	@Override
	protected void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right) {
		PrereqList prereqList = (PrereqList) mPrereq;
		IntegerCriteria whenTLCriteria = prereqList.getWhenTLCriteria();
		left.add(addComboBox(WHEN_TL, new Object[] { NO_TL_PREREQ, TL_IS, TL_IS_AT_LEAST, TL_IS_AT_MOST }, mapWhenTLToString(whenTLCriteria)));
		if (PrereqList.isWhenTLEnabled(whenTLCriteria)) {
			left.add(addNumericCompareField(whenTLCriteria, 0, 99, false));
		}
		left.add(addComboBox(ANY_ALL, new Object[] { REQUIRES_ALL, REQUIRES_ANY }, prereqList.requiresAll() ? REQUIRES_ALL : REQUIRES_ANY));

		grid.add(new FlexSpacer(0, 0, true, false), 0, 1);

		IconButton button = new IconButton(StdImage.MORE, ADD_PREREQ_LIST_TOOLTIP, () -> addPrereqList());
		add(button);
		right.add(button);
		button = new IconButton(StdImage.ADD, ADD_PREREQ_TOOLTIP, () -> addPrereq());
		add(button);
		right.add(button);
	}

	private void addPrereqList() {
		addItem(new PrereqList((PrereqList) mPrereq, true));
	}

	private void addPrereq() {
		try {
			addItem((Prereq) LAST_ITEM_TYPE.getConstructor(PrereqList.class).newInstance(mPrereq));
		} catch (Exception exception) {
			// Shouldn't have a failure...
			exception.printStackTrace(System.err);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (ANY_ALL.equals(command)) {
			((PrereqList) mPrereq).setRequiresAll(((JComboBox<?>) event.getSource()).getSelectedIndex() == 0);
			getParent().repaint();
		} else if (WHEN_TL.equals(command)) {
			PrereqList prereqList = (PrereqList) mPrereq;
			IntegerCriteria whenTLCriteria = prereqList.getWhenTLCriteria();
			Object value = ((JComboBox<?>) event.getSource()).getSelectedItem();
			if (!mapWhenTLToString(whenTLCriteria).equals(value)) {
				if (TL_IS.equals(value)) {
					if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
						PrereqList.setWhenTLEnabled(whenTLCriteria, true);
					}
					whenTLCriteria.setType(NumericCompareType.IS);
				} else if (TL_IS_AT_LEAST.equals(value)) {
					if (!PrereqList.isWhenTLEnabled(whenTLCriteria)) {
						PrereqList.setWhenTLEnabled(whenTLCriteria, true);
					}
					whenTLCriteria.setType(NumericCompareType.AT_LEAST);
				} else if (TL_IS_AT_MOST.equals(value)) {
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
