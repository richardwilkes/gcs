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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.edit.Incrementable;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.utility.collections.FilteredIterator;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.MultipleRowUndo;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.Row;
import com.trollworks.gcs.widgets.outline.RowPostProcessor;
import com.trollworks.gcs.widgets.outline.RowUndo;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for skills. */
public class SkillOutline extends ListOutline implements Incrementable {
	private static String	MSG_INCREMENT;
	private static String	MSG_DECREMENT;

	static {
		LocalizedMessages.initialize(SkillOutline.class);
	}

	private static OutlineModel extractModel(DataFile dataFile) {
		if (dataFile instanceof GURPSCharacter) {
			return ((GURPSCharacter) dataFile).getSkillsRoot();
		}
		if (dataFile instanceof Template) {
			return ((Template) dataFile).getSkillsModel();
		}
		if (dataFile instanceof LibraryFile) {
			return ((LibraryFile) dataFile).getSkillList().getModel();
		}
		return ((ListFile) dataFile).getModel();
	}

	/**
	 * Create a new skills outline.
	 * 
	 * @param dataFile The owning data file.
	 */
	public SkillOutline(DataFile dataFile) {
		this(dataFile, extractModel(dataFile));
	}

	/**
	 * Create a new skills outline.
	 * 
	 * @param dataFile The owning data file.
	 * @param model The {@link OutlineModel} to use.
	 */
	public SkillOutline(DataFile dataFile, OutlineModel model) {
		super(dataFile, model, Skill.ID_LIST_CHANGED);
		SkillColumn.addColumns(this, dataFile);
	}

	public String getDecrementTitle() {
		return MSG_DECREMENT;
	}

	public String getIncrementTitle() {
		return MSG_INCREMENT;
	}

	public boolean canDecrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeafRows(true);
	}

	public boolean canIncrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeafRows(false);
	}

	private boolean selectionHasLeafRows(boolean requirePointsAboveZero) {
		for (Skill skill : new FilteredIterator<Skill>(getModel().getSelectionAsList(), Skill.class)) {
			if (!skill.canHaveChildren() && (!requirePointsAboveZero || skill.getPoints() > 0)) {
				return true;
			}
		}
		return false;
	}

	public void decrement() {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();
		for (Skill skill : new FilteredIterator<Skill>(getModel().getSelectionAsList(), Skill.class)) {
			if (!skill.canHaveChildren()) {
				int points = skill.getPoints();
				if (points > 0) {
					RowUndo undo = new RowUndo(skill);

					skill.setPoints(points - 1);
					if (undo.finish()) {
						undos.add(undo);
					}
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new MultipleRowUndo(undos);
		}
	}

	public void increment() {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();
		for (Skill skill : new FilteredIterator<Skill>(getModel().getSelectionAsList(), Skill.class)) {
			if (!skill.canHaveChildren()) {
				RowUndo undo = new RowUndo(skill);
				skill.setPoints(skill.getPoints() + 1);
				if (undo.finish()) {
					undos.add(undo);
				}
			}
		}
		if (!undos.isEmpty()) {
			repaintSelection();
			new MultipleRowUndo(undos);
		}
	}

	@Override protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof Skill;
	}

	@Override public void convertDragRowsToSelf(List<Row> list) {
		OutlineModel model = getModel();
		Row[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof GURPSCharacter || mDataFile instanceof Template;
		ArrayList<ListRow> process = forSheetOrTemplate ? new ArrayList<ListRow>() : null;

		for (Row element : rows) {
			ListRow row;

			if (element instanceof Technique) {
				row = new Technique(mDataFile, (Technique) element, forSheetOrTemplate);
			} else {
				row = new Skill(mDataFile, (Skill) element, true, forSheetOrTemplate);
			}

			model.collectRowsAndSetOwner(list, row, false);
			if (forSheetOrTemplate) {
				addRowsToBeProcessed(process, row);
			}
		}

		if (forSheetOrTemplate) {
			EventQueue.invokeLater(new RowPostProcessor(this, process));
		}
	}
}
