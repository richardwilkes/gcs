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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.OutlineProxy;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

/** Provides the "Duplicate" command. */
// RAW: This should be reimplemented in terms of the Duplicatable interface
public class DuplicateCommand extends Command {
	@Localize("Duplicate")
	@Localize(locale = "de", value = "Duplizieren")
	@Localize(locale = "ru", value = "Дублировать")
	private static String					DUPLICATE;
	@Localize("Duplicate Rows")
	@Localize(locale = "de", value = "Zeile Duplizieren")
	@Localize(locale = "ru", value = "Дублировать строки")
	private static String					DUPLICATE_UNDO;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String				CMD_DUPLICATE	= "Duplicate";				//$NON-NLS-1$

	/** The singleton {@link DuplicateCommand}. */
	public static final DuplicateCommand	INSTANCE		= new DuplicateCommand();

	private DuplicateCommand() {
		super(DUPLICATE, CMD_DUPLICATE, KeyEvent.VK_U);
	}

	@Override
	public void adjust() {
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

	@Override
	public void actionPerformed(ActionEvent event) {
		Component focus = getFocusOwner();
		if (focus instanceof OutlineProxy) {
			focus = ((OutlineProxy) focus).getRealOutline();
		}
		ListOutline outline = (ListOutline) focus;
		OutlineModel model = outline.getModel();
		if (!model.isLocked() && model.hasSelection()) {
			ArrayList<Row> rows = new ArrayList<>();
			ArrayList<Row> topRows = new ArrayList<>();
			DataFile dataFile = outline.getDataFile();
			dataFile.startNotify();
			model.setDragRows(model.getSelectionAsList(true).toArray(new Row[0]));
			outline.convertDragRowsToSelf(rows);
			model.setDragRows(null);
			for (Row row : rows) {
				if (row.getDepth() == 0) {
					topRows.add(row);
				}
			}
			outline.addRow(topRows.toArray(new ListRow[0]), DUPLICATE_UNDO, true);
			dataFile.endNotify();
			model.select(topRows, false);
			outline.scrollSelectionIntoView();
		}
	}
}
