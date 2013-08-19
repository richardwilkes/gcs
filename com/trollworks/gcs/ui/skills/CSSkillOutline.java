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

package com.trollworks.gcs.ui.skills;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMMultipleRowUndo;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.CMRowUndo;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMTechnique;
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

/** An outline specifically for skills. */
public class CSSkillOutline extends CSOutline {
	private static TKOutlineModel extractModel(CMDataFile dataFile) {
		if (dataFile instanceof CMCharacter) {
			return ((CMCharacter) dataFile).getSkillsRoot();
		}
		if (dataFile instanceof CMTemplate) {
			return ((CMTemplate) dataFile).getSkillsModel();
		}
		return ((CMListFile) dataFile).getModel();
	}

	/**
	 * Create a new skills outline.
	 * 
	 * @param dataFile The owning data file.
	 */
	public CSSkillOutline(CMDataFile dataFile) {
		super(dataFile, extractModel(dataFile), CMSkill.ID_LIST_CHANGED);
		CSSkillColumnID.addColumns(this, dataFile);
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		boolean forSheetOrTemplate = mDataFile instanceof CMCharacter || mDataFile instanceof CMTemplate;

		if (CSWindow.CMD_INCREMENT.equals(command)) {
			item.setTitle(Msgs.INCREMENT);
			item.setEnabled(forSheetOrTemplate && selectionHasLeafRows(false));
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			item.setTitle(Msgs.DECREMENT);
			item.setEnabled(forSheetOrTemplate && selectionHasLeafRows(true));
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CSWindow.CMD_INCREMENT.equals(command)) {
			incrementPoints();
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			decrementPoints();
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private boolean selectionHasLeafRows(boolean requirePointsAboveZero) {
		for (CMSkill skill : new TKFilteredIterator<CMSkill>(getModel().getSelectionAsList(), CMSkill.class)) {
			if (!skill.canHaveChildren() && (!requirePointsAboveZero || skill.getPoints() > 0)) {
				return true;
			}
		}
		return false;
	}

	private void incrementPoints() {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();

		for (CMSkill skill : new TKFilteredIterator<CMSkill>(getModel().getSelectionAsList(), CMSkill.class)) {
			if (!skill.canHaveChildren()) {
				CMRowUndo undo = new CMRowUndo(skill);

				skill.setPoints(skill.getPoints() + 1);
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

	private void decrementPoints() {
		ArrayList<CMRowUndo> undos = new ArrayList<CMRowUndo>();

		for (CMSkill skill : new TKFilteredIterator<CMSkill>(getModel().getSelectionAsList(), CMSkill.class)) {
			if (!skill.canHaveChildren()) {
				int points = skill.getPoints();

				if (points > 0) {
					CMRowUndo undo = new CMRowUndo(skill);

					skill.setPoints(points - 1);
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
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof CMSkill;
	}

	@Override protected void convertDragRowsToSelf(List<TKRow> list) {
		TKOutlineModel model = getModel();
		TKRow[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof CMCharacter || mDataFile instanceof CMTemplate;
		ArrayList<CMRow> process = forSheetOrTemplate ? new ArrayList<CMRow>() : null;

		for (TKRow element : rows) {
			CMRow row;

			if (element instanceof CMTechnique) {
				row = new CMTechnique(mDataFile, (CMTechnique) element, forSheetOrTemplate);
			} else {
				row = new CMSkill(mDataFile, (CMSkill) element, true, forSheetOrTemplate);
			}

			model.collectRowsAndSetOwner(list, row, false);
			if (forSheetOrTemplate) {
				addRowsToBeProcessed(process, row);
			}
		}

		if (forSheetOrTemplate) {
			EventQueue.invokeLater(new CSRowPostProcessor(this, process));
		}
	}
}
