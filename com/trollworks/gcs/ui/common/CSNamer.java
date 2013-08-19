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

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKOptionDialog;

import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;

/** Asks the user to name items that have been marked to be customized. */
public class CSNamer extends TKPanel {
	private CMRow					mRow;
	private ArrayList<TKTextField>	mFields;

	/**
	 * Brings up a modal naming dialog for each row in the list.
	 * 
	 * @param list The rows to name.
	 * @return Whether anything was modified.
	 */
	static public boolean name(ArrayList<CMRow> list) {
		ArrayList<CMRow> rowList = new ArrayList<CMRow>();
		ArrayList<HashSet<String>> setList = new ArrayList<HashSet<String>>();
		boolean modified = false;
		int count;

		for (CMRow row : list) {
			HashSet<String> set = new HashSet<String>();

			row.fillWithNameableKeys(set);
			if (!set.isEmpty()) {
				rowList.add(row);
				setList.add(set);
			}
		}

		count = rowList.size();
		for (int i = 0; i < count; i++) {
			CMRow row = rowList.get(i);
			boolean hasMore = i != count - 1;
			TKOptionDialog dialog = new TKOptionDialog(MessageFormat.format(Msgs.NAME_TITLE, row.getLocalizedName()), hasMore ? TKOptionDialog.TYPE_YES_NO_CANCEL : TKOptionDialog.TYPE_YES_NO);
			CSNamer panel = new CSNamer(row, setList.get(i), count - i - 1);

			if (hasMore) {
				dialog.setCancelButtonTitle(Msgs.CANCEL_REST);
			}
			dialog.setYesButtonTitle(Msgs.APPLY);
			dialog.setNoButtonTitle(Msgs.CANCEL);
			dialog.setResizable(true);
			switch (dialog.doModal(row.getImage(true), panel)) {
				case TKOptionDialog.YES:
					panel.applyChanges();
					modified = true;
					break;
				case TKOptionDialog.NO:
					break;
				case TKDialog.CANCEL:
					return modified;
			}
		}
		return modified;
	}

	private CSNamer(CMRow row, HashSet<String> set, int remaining) {
		super(new TKCompassLayout());
		mRow = row;
		mFields = new ArrayList<TKTextField>();
		add(createTop(row, remaining), TKCompassPosition.NORTH);
		add(createCenter(set), TKCompassPosition.CENTER);
	}

	private TKPanel createTop(CMRow row, int remaining) {
		TKPanel top = new TKPanel(new TKColumnLayout());
		TKLabel label = new TKLabel(TKTextUtility.truncateIfNecessary(row.toString(), 80, TKAlignment.RIGHT), TKFont.CONTROL_FONT_KEY, TKAlignment.LEFT, true);

		top.setBorder(new TKEmptyBorder(0, 0, 15, 0));
		if (remaining > 0) {
			String msg;

			if (remaining == 1) {
				msg = Msgs.ONE_REMAINING;
			} else {
				msg = MessageFormat.format(Msgs.REMAINING, new Integer(remaining));
			}
			top.add(new TKLabel(msg, TKFont.CONTROL_FONT_KEY, TKAlignment.CENTER));
		}
		label.setBorder(new TKCompoundBorder(TKLineBorder.getSharedBorder(true), new TKEmptyBorder(0, 2, 0, 2)));
		label.setBackground(TKColor.CONTROL_ROLL);
		label.setOpaque(true);
		top.add(new TKPanel());
		top.add(label);
		return top;
	}

	private TKPanel createCenter(HashSet<String> set) {
		ArrayList<String> list = new ArrayList<String>(set);
		TKPanel center = new TKPanel(new TKColumnLayout(2));

		Collections.sort(list);
		for (String name : list) {
			TKTextField field = new TKTextField(100);

			field.setName(name);
			mFields.add(field);
			center.add(new TKLabel(name, TKAlignment.RIGHT));
			center.add(field);
		}
		return center;
	}

	private void applyChanges() {
		HashMap<String, String> map = new HashMap<String, String>();
		TKBaseWindow window = getBaseWindow();

		if (window != null) {
			window.forceFocusToAccept();
		}

		for (TKTextField field : mFields) {
			map.put(field.getName(), field.getText());
		}
		mRow.applyNameableKeys(map);
	}
}
