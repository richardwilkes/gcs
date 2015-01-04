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

package com.trollworks.gcs.character.names;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.Alignment;
import com.trollworks.toolkit.ui.layout.FlexColumn;
import com.trollworks.toolkit.ui.layout.FlexComponent;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.layout.LayoutSize;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.TextUtility;

import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;

import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.SwingConstants;

/** Asks the user to name items that have been marked to be customized. */
public class Namer extends JPanel {
	@Localize("Name {0}")
	@Localize(locale = "de", value = "Benenne {0}")
	@Localize(locale = "ru", value = "Имя {0}")
	private static String			NAME_TITLE;
	@Localize("1 item remaining to be named.")
	@Localize(locale = "de", value = "1 weiteres Element zu benennen.")
	@Localize(locale = "ru", value = "осталось назвать 1 элемент.")
	private static String			ONE_REMAINING;
	@Localize("{0} items remaining to be named.")
	@Localize(locale = "de", value = "{0} weitere Elemente zu benennen.")
	@Localize(locale = "ru", value = "{0} элементов осталось назвать.")
	private static String			REMAINING;
	@Localize("Apply")
	@Localize(locale = "de", value = "Anwenden")
	@Localize(locale = "ru", value = "Применить")
	private static String			APPLY;
	@Localize("Cancel")
	@Localize(locale = "de", value = "Abbrechen")
	@Localize(locale = "ru", value = "Отмена")
	private static String			CANCEL;
	@Localize("Cancel Remaining")
	@Localize(locale = "de", value = "Alles Abbrechen")
	@Localize(locale = "ru", value = "Пропустить остальные")
	private static String			CANCEL_REST;

	static {
		Localization.initialize();
	}

	private ListRow					mRow;
	private ArrayList<JTextField>	mFields;

	/**
	 * Brings up a modal naming dialog for each row in the list.
	 *
	 * @param owner The owning component.
	 * @param list The rows to name.
	 * @return Whether anything was modified.
	 */
	static public boolean name(Component owner, ArrayList<ListRow> list) {
		ArrayList<ListRow> rowList = new ArrayList<>();
		ArrayList<HashSet<String>> setList = new ArrayList<>();
		boolean modified = false;
		int count;

		for (ListRow row : list) {
			HashSet<String> set = new HashSet<>();

			row.fillWithNameableKeys(set);
			if (!set.isEmpty()) {
				rowList.add(row);
				setList.add(set);
			}
		}

		count = rowList.size();
		for (int i = 0; i < count; i++) {
			ListRow row = rowList.get(i);
			boolean hasMore = i != count - 1;
			int type = hasMore ? JOptionPane.YES_NO_CANCEL_OPTION : JOptionPane.YES_NO_OPTION;
			String[] options = hasMore ? new String[] { APPLY, CANCEL, CANCEL_REST } : new String[] { APPLY, CANCEL };
			Namer panel = new Namer(row, setList.get(i), count - i - 1);
			switch (WindowUtils.showOptionDialog(owner, panel, MessageFormat.format(NAME_TITLE, row.getLocalizedName()), true, type, JOptionPane.PLAIN_MESSAGE, row.getIcon(true), options, APPLY)) {
				case JOptionPane.YES_OPTION:
					panel.applyChanges();
					modified = true;
					break;
				case JOptionPane.NO_OPTION:
					break;
				case JOptionPane.CANCEL_OPTION:
				case JOptionPane.CLOSED_OPTION:
				default:
					return modified;
			}
		}
		return modified;
	}

	private Namer(ListRow row, HashSet<String> set, int remaining) {
		JLabel label;
		mRow = row;
		mFields = new ArrayList<>();

		FlexColumn column = new FlexColumn();
		if (remaining > 0) {
			label = new JLabel(remaining == 1 ? ONE_REMAINING : MessageFormat.format(REMAINING, new Integer(remaining)), SwingConstants.CENTER);
			Dimension size = label.getMaximumSize();
			size.width = LayoutSize.MAXIMUM_SIZE;
			label.setMaximumSize(size);
			add(label);
			column.add(label);
		}
		label = new JLabel(TextUtility.truncateIfNecessary(row.toString(), 80, SwingConstants.RIGHT), SwingConstants.CENTER);
		Dimension size = label.getMaximumSize();
		size.width = LayoutSize.MAXIMUM_SIZE;
		size.height += 4;
		label.setMaximumSize(size);
		size = label.getPreferredSize();
		size.height += 4;
		label.setPreferredSize(size);
		label.setMinimumSize(size);
		label.setBackground(Color.BLACK);
		label.setForeground(Color.WHITE);
		label.setOpaque(true);
		add(label);
		column.add(label);
		column.add(new FlexSpacer(0, 10, false, false));

		int rowIndex = 0;
		FlexGrid grid = new FlexGrid();
		grid.setFillHorizontal(true);
		ArrayList<String> list = new ArrayList<>(set);
		Collections.sort(list);
		for (String name : list) {
			JTextField field = new JTextField(25);
			field.setName(name);
			size = field.getPreferredSize();
			size.width = LayoutSize.MAXIMUM_SIZE;
			field.setMaximumSize(size);
			mFields.add(field);
			label = new JLabel(name, SwingConstants.RIGHT);
			UIUtilities.setOnlySize(label, label.getPreferredSize());
			add(label);
			add(field);
			grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), rowIndex, 0);
			grid.add(field, rowIndex++, 1);
		}

		column.add(grid);
		column.add(new FlexSpacer(0, 0, false, true));
		column.apply(this);
	}

	private void applyChanges() {
		UIUtilities.forceFocusToAccept();
		HashMap<String, String> map = new HashMap<>();
		for (JTextField field : mFields) {
			map.put(field.getName(), field.getText());
		}
		mRow.applyNameableKeys(map);
	}
}
