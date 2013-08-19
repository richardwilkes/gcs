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

package com.trollworks.gcs.character.names;

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.text.TextUtility;
import com.trollworks.gcs.widgets.CommitEnforcer;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.WindowUtils;
import com.trollworks.gcs.widgets.layout.Alignment;
import com.trollworks.gcs.widgets.layout.FlexColumn;
import com.trollworks.gcs.widgets.layout.FlexComponent;
import com.trollworks.gcs.widgets.layout.FlexGrid;
import com.trollworks.gcs.widgets.layout.FlexSpacer;
import com.trollworks.gcs.widgets.layout.LayoutSize;
import com.trollworks.gcs.widgets.outline.ListRow;

import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;

import javax.swing.ImageIcon;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.SwingConstants;

/** Asks the user to name items that have been marked to be customized. */
public class Namer extends JPanel {
	private static String			MSG_NAME_TITLE;
	private static String			MSG_ONE_REMAINING;
	private static String			MSG_REMAINING;
	private static String			MSG_APPLY;
	private static String			MSG_CANCEL;
	private static String			MSG_CANCEL_REST;
	private ListRow					mRow;
	private ArrayList<JTextField>	mFields;

	static {
		LocalizedMessages.initialize(Namer.class);
	}

	/**
	 * Brings up a modal naming dialog for each row in the list.
	 * 
	 * @param owner The owning component.
	 * @param list The rows to name.
	 * @return Whether anything was modified.
	 */
	static public boolean name(Component owner, ArrayList<ListRow> list) {
		ArrayList<ListRow> rowList = new ArrayList<ListRow>();
		ArrayList<HashSet<String>> setList = new ArrayList<HashSet<String>>();
		boolean modified = false;
		int count;

		for (ListRow row : list) {
			HashSet<String> set = new HashSet<String>();

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
			String[] options = hasMore ? new String[] { MSG_APPLY, MSG_CANCEL, MSG_CANCEL_REST } : new String[] { MSG_APPLY, MSG_CANCEL };
			Namer panel = new Namer(row, setList.get(i), count - i - 1);
			switch (WindowUtils.showOptionDialog(owner, panel, MessageFormat.format(MSG_NAME_TITLE, row.getLocalizedName()), true, type, JOptionPane.PLAIN_MESSAGE, new ImageIcon(row.getImage(true)), options, MSG_APPLY)) {
				case JOptionPane.YES_OPTION:
					panel.applyChanges();
					modified = true;
					break;
				case JOptionPane.NO_OPTION:
					break;
				case JOptionPane.CANCEL_OPTION:
				case JOptionPane.CLOSED_OPTION:
					return modified;
			}
		}
		return modified;
	}

	private Namer(ListRow row, HashSet<String> set, int remaining) {
		JLabel label;
		mRow = row;
		mFields = new ArrayList<JTextField>();

		FlexColumn column = new FlexColumn();
		if (remaining > 0) {
			label = new JLabel(remaining == 1 ? MSG_ONE_REMAINING : MessageFormat.format(MSG_REMAINING, new Integer(remaining)), SwingConstants.CENTER);
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
		ArrayList<String> list = new ArrayList<String>(set);
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
		CommitEnforcer.forceFocusToAccept();
		HashMap<String, String> map = new HashMap<String, String>();
		for (JTextField field : mFields) {
			map.put(field.getName(), field.getText());
		}
		mRow.applyNameableKeys(map);
	}
}
