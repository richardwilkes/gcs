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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.OutlineProxy;
import com.trollworks.gcs.widgets.outline.Row;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

import javax.swing.JMenuItem;

/** Provides the "Duplicate" command. */
public class DuplicateCommand extends Command {
	private static String					MSG_DUPLICATE;
	private static String					MSG_DUPLICATE_UNDO;

	static {
		LocalizedMessages.initialize(DuplicateCommand.class);
	}

	/** The singleton {@link DuplicateCommand}. */
	public static final DuplicateCommand	INSTANCE	= new DuplicateCommand();

	private DuplicateCommand() {
		super(MSG_DUPLICATE, KeyEvent.VK_U);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		Component focus = getFocusOwner();
		if (focus instanceof OutlineProxy) {
			focus = ((OutlineProxy) focus).getRealOutline();
		}
		if (focus instanceof ListOutline) {
			OutlineModel model = ((ListOutline) focus).getModel();
			setEnabled(!model.isLocked() && model.hasSelection());
		} else {
			setEnabled(false);
		}
	}

	@Override public void actionPerformed(ActionEvent event) {
		Component focus = getFocusOwner();
		if (focus instanceof OutlineProxy) {
			focus = ((OutlineProxy) focus).getRealOutline();
		}
		ListOutline outline = (ListOutline) focus;
		OutlineModel model = outline.getModel();
		if (!model.isLocked() && model.hasSelection()) {
			ArrayList<Row> rows = new ArrayList<Row>();
			ArrayList<Row> topRows = new ArrayList<Row>();
			outline.getDataFile().startNotify();
			model.setDragRows(model.getSelectionAsList(true).toArray(new Row[0]));
			outline.convertDragRowsToSelf(rows);
			model.setDragRows(null);
			for (Row row : rows) {
				if (row.getDepth() == 0) {
					topRows.add(row);
				}
			}
			outline.addRow(topRows.toArray(new ListRow[0]), MSG_DUPLICATE_UNDO, true);
			model.select(topRows, false);
			outline.scrollSelectionIntoView();
		}
	}
}
