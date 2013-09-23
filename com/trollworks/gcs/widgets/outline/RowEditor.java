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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.outline;

import static com.trollworks.gcs.widgets.outline.RowEditor_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.layout.RowDistribution;
import com.trollworks.ttk.widgets.ActionPanel;
import com.trollworks.ttk.widgets.CommitEnforcer;
import com.trollworks.ttk.widgets.WindowUtils;

import java.awt.BorderLayout;
import java.awt.Component;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.SwingConstants;
import javax.swing.border.EmptyBorder;

@Localized({
				@LS(key = "WINDOW_TITLE", msg = "Edit {0}"),
				@LS(key = "CANCEL_REST", msg = "Cancel Remaining"),
				@LS(key = "CANCEL", msg = "Cancel"),
				@LS(key = "APPLY", msg = "Apply"),
				@LS(key = "ONE_REMAINING", msg = "1 item remaining to be edited."),
				@LS(key = "REMAINING", msg = "{0} items remaining to be edited."),
})
/**
 * The base class for all row editors.
 * 
 * @param <T> The row class being edited.
 */
public abstract class RowEditor<T extends ListRow> extends ActionPanel {
	private static HashMap<Class<?>, String>	LAST_TAB_MAP	= new HashMap<>();
	/** Whether the underlying data should be editable. */
	protected boolean							mIsEditable;
	/** The row being edited. */
	protected T									mRow;

	/**
	 * Brings up a modal detailed editor for each row in the list.
	 * 
	 * @param owner The owning component.
	 * @param list The rows to edit.
	 * @return Whether anything was modified.
	 */
	@SuppressWarnings("unused")
	static public boolean edit(Component owner, List<? extends ListRow> list) {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();
		ListRow[] rows = list.toArray(new ListRow[0]);

		for (int i = 0; i < rows.length; i++) {
			boolean hasMore = i != rows.length - 1;
			ListRow row = rows[i];
			RowEditor<? extends ListRow> editor = row.createEditor();
			String title = MessageFormat.format(WINDOW_TITLE, row.getRowType());
			JPanel wrapper = new JPanel(new BorderLayout());

			if (hasMore) {
				int remaining = rows.length - i - 1;
				String msg = remaining == 1 ? ONE_REMAINING : MessageFormat.format(REMAINING, new Integer(remaining));
				JLabel panel = new JLabel(msg, SwingConstants.CENTER);
				panel.setBorder(new EmptyBorder(0, 0, 10, 0));
				wrapper.add(panel, BorderLayout.NORTH);
			}
			wrapper.add(editor, BorderLayout.CENTER);

			int type = hasMore ? JOptionPane.YES_NO_CANCEL_OPTION : JOptionPane.YES_NO_OPTION;
			String[] options = hasMore ? new String[] { APPLY, CANCEL, CANCEL_REST } : new String[] { APPLY, CANCEL };
			switch (WindowUtils.showOptionDialog(owner, wrapper, title, true, type, JOptionPane.PLAIN_MESSAGE, null, options, APPLY)) {
				case JOptionPane.YES_OPTION:
					RowUndo undo = new RowUndo(row);
					if (editor.applyChanges()) {
						if (undo.finish()) {
							undos.add(undo);
						}
					}
					break;
				case JOptionPane.NO_OPTION:
					break;
				case JOptionPane.CANCEL_OPTION:
				case JOptionPane.CLOSED_OPTION:
				default:
					i = rows.length;
					break;
			}
			editor.finished();
		}

		if (!undos.isEmpty()) {
			new MultipleRowUndo(undos);
			return true;
		}
		return false;
	}

	/**
	 * Creates a new {@link RowEditor}.
	 * 
	 * @param row The row being edited.
	 */
	protected RowEditor(T row) {
		super(new ColumnLayout(1, 0, 5, RowDistribution.GIVE_EXCESS_TO_LAST));
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
		CommitEnforcer.forceFocusToAccept();
		boolean modified = applyChangesSelf();
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
