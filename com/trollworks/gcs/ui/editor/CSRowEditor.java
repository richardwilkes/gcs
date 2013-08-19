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

package com.trollworks.gcs.ui.editor;

import com.trollworks.gcs.model.CMMultipleRowUndo;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.CMRowUndo;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.widget.layout.TKRowDistribution;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKDialog;

import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

/**
 * The base class for all row editors.
 * 
 * @param <T> The row class being edited.
 */
public abstract class CSRowEditor<T extends CMRow> extends TKPanel {
	private static HashMap<Class<?>, String>	LAST_TAB_MAP	= new HashMap<Class<?>, String>();
	/** Whether the underlying data should be editable. */
	protected boolean							mIsEditable;
	/** The row being edited. */
	protected T									mRow;

	/**
	 * Brings up a modal detailed editor for each row in the list.
	 * 
	 * @param list The rows to edit.
	 * @return Whether anything was modified.
	 */
	static public boolean edit(List<? extends CMRow> list) {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();
		CMRow[] rows = list.toArray(new CMRow[0]);

		for (int i = 0; i < rows.length; i++) {
			boolean hasMore = i != rows.length - 1;
			CMRow row = rows[i];
			CSRowEditor<? extends CMRow> editor = row.createEditor();
			TKOptionDialog dialog = new TKOptionDialog(MessageFormat.format(Msgs.WINDOW_TITLE, row.getRowType()), hasMore ? TKOptionDialog.TYPE_YES_NO_CANCEL : TKOptionDialog.TYPE_YES_NO);
			TKPanel wrapper = new TKPanel(new TKCompassLayout());

			if (hasMore) {
				dialog.setCancelButtonTitle(Msgs.CANCEL_REST);
			}
			dialog.setYesButtonTitle(Msgs.APPLY);
			dialog.setNoButtonTitle(Msgs.CANCEL);
			dialog.setResizable(true);
			if (row.getOwner().isLocked()) {
				dialog.getYesButton().setEnabled(false);
			}
			if (hasMore) {
				int remaining = rows.length - i - 1;
				String msg = remaining == 1 ? Msgs.ONE_REMAINING : MessageFormat.format(Msgs.REMAINING, new Integer(remaining));
				TKLabel panel = new TKLabel(msg, TKFont.CONTROL_FONT_KEY, TKAlignment.CENTER);

				panel.setBorder(new TKCompoundBorder(new TKEmptyBorder(0, 0, 15, 0), new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.BOTTOM_EDGE)));
				wrapper.add(panel, TKCompassPosition.NORTH);
			}
			wrapper.add(editor, TKCompassPosition.CENTER);
			switch (dialog.doModal(null, wrapper)) {
				case TKOptionDialog.YES:
					CMRowUndo undo = new CMRowUndo(row);

					if (editor.applyChanges()) {
						if (undo.finish()) {
							undos.add(undo);
						}
					}
					break;
				case TKOptionDialog.NO:
					break;
				case TKDialog.CANCEL:
					i = rows.length;
					break;
			}
			editor.finished();
		}

		if (!undos.isEmpty()) {
			new CMMultipleRowUndo(undos);
			return true;
		}
		return false;
	}

	/**
	 * Creates a new {@link CSRowEditor}.
	 * 
	 * @param row The row being edited.
	 */
	protected CSRowEditor(T row) {
		super(new TKColumnLayout(1, 0, 5, TKRowDistribution.GIVE_EXCESS_TO_LAST));
		mRow = row;
		mIsEditable = !mRow.getOwner().isLocked();
	}

	/** @return The last tab showing for this specific row editor class. */
	public String getLastTabName() {
		return LAST_TAB_MAP.get(getClass());
	}

	/** @param name The last tab showing for this specific row editor class. */
	public void updateLastTabName(String name) {
		LAST_TAB_MAP.put(getClass(), name);
	}

	/**
	 * Called to apply any changes that were made.
	 * 
	 * @return Whether anything was modified.
	 */
	public final boolean applyChanges() {
		boolean modified;

		forceFocusToAccept();
		modified = applyChangesSelf();
		if (modified) {
			mRow.getDataFile().setModified(true);
		}
		return modified;
	}

	/**
	 * Called to apply any changes that were made.
	 * 
	 * @return Whether anything was modified.
	 */
	protected abstract boolean applyChangesSelf();

	/** Called when the editor is no longer needed. */
	public abstract void finished();
}
