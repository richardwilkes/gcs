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

package com.trollworks.gcs.ui.advantage;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMMultipleRowUndo;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.CMRowUndo;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.ui.common.CSRowPostProcessor;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.gcs.ui.common.CSWindow;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for (Dis)Advantages. */
public class CSAdvantageOutline extends CSOutline {
	private static TKOutlineModel extractModel(CMDataFile dataFile) {
		if (dataFile instanceof CMCharacter) {
			return ((CMCharacter) dataFile).getAdvantagesModel();
		}
		if (dataFile instanceof CMTemplate) {
			return ((CMTemplate) dataFile).getAdvantagesModel();
		}
		return ((CMListFile) dataFile).getModel();
	}

	/**
	 * Create a new Advantages, Disadvantages & Quirks outline.
	 * 
	 * @param dataFile The owning data file.
	 */
	public CSAdvantageOutline(CMDataFile dataFile) {
		super(dataFile, extractModel(dataFile), CMAdvantage.ID_LIST_CHANGED);
		CSAdvantageColumnID.addColumns(this, dataFile);
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		boolean forSheetOrTemplate = mDataFile instanceof CMCharacter || mDataFile instanceof CMTemplate;

		if (CSWindow.CMD_INCREMENT.equals(command)) {
			item.setTitle(Msgs.INCREMENT);
			item.setEnabled(forSheetOrTemplate && selectionHasLeveledRows(false));
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			item.setTitle(Msgs.DECREMENT);
			item.setEnabled(forSheetOrTemplate && selectionHasLeveledRows(true));
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CSWindow.CMD_INCREMENT.equals(command)) {
			incrementLevel();
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			decrementLevel();
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private boolean selectionHasLeveledRows(boolean requireLevelAboveZero) {
		for (CMAdvantage advantage : new TKFilteredIterator<CMAdvantage>(getModel().getSelectionAsList(), CMAdvantage.class)) {
			if (!advantage.canHaveChildren() && advantage.isLeveled() && (!requireLevelAboveZero || advantage.getLevels() > 0)) {
				return true;
			}
		}
		return false;
	}

	private void incrementLevel() {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();

		for (CMAdvantage advantage : new TKFilteredIterator<CMAdvantage>(getModel().getSelectionAsList(), CMAdvantage.class)) {
			if (!advantage.canHaveChildren() && advantage.isLeveled()) {
				CMRowUndo undo = new CMRowUndo(advantage);

				advantage.setLevels(advantage.getLevels() + 1);
				if (undo.finish()) {
					undos.add(undo);
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new CMMultipleRowUndo(undos);
		}
	}

	private void decrementLevel() {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();

		for (CMAdvantage advantage : new TKFilteredIterator<CMAdvantage>(getModel().getSelectionAsList(), CMAdvantage.class)) {
			if (!advantage.canHaveChildren() && advantage.isLeveled()) {
				int levels = advantage.getLevels();

				if (levels > 0) {
					CMRowUndo undo = new CMRowUndo(advantage);

					advantage.setLevels(levels - 1);
					if (undo.finish()) {
						undos.add(undo);
					}
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new CMMultipleRowUndo(undos);
		}
	}

	@Override protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, TKRow[] rows) {
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof CMAdvantage;
	}

	@Override protected void convertDragRowsToSelf(List<TKRow> list) {
		TKOutlineModel model = getModel();
		TKRow[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof CMCharacter || mDataFile instanceof CMTemplate;
		ArrayList<CMRow> process = forSheetOrTemplate ? new ArrayList<CMRow>() : null;

		for (TKRow element : rows) {
			CMAdvantage advantage = new CMAdvantage(mDataFile, (CMAdvantage) element, true);

			model.collectRowsAndSetOwner(list, advantage, false);
			if (forSheetOrTemplate) {
				addRowsToBeProcessed(process, advantage);
			}
		}

		if (forSheetOrTemplate && !process.isEmpty()) {
			EventQueue.invokeLater(new CSRowPostProcessor(this, process));
		}
	}
}
