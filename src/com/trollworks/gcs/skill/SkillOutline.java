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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.edit.Incrementable;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.MultipleRowUndo;
import com.trollworks.gcs.widgets.outline.RowPostProcessor;
import com.trollworks.gcs.widgets.outline.RowUndo;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.utility.Localization;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for skills. */
public class SkillOutline extends ListOutline implements Incrementable {
	@Localize("Increment Points")
	@Localize(locale = "de", value = "Punkte erhöhen")
	@Localize(locale = "ru", value = "Увеличить очки")
	private static String	INCREMENT;
	@Localize("Decrement Points")
	@Localize(locale = "de", value = "Punkte verringern")
	@Localize(locale = "ru", value = "Уменьшить очки")
	private static String	DECREMENT;

	static {
		Localization.initialize();
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

	@Override
	public String getDecrementTitle() {
		return DECREMENT;
	}

	@Override
	public String getIncrementTitle() {
		return INCREMENT;
	}

	@Override
	public boolean canDecrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeafRows(true);
	}

	@Override
	public boolean canIncrement() {
		return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeafRows(false);
	}

	private boolean selectionHasLeafRows(boolean requirePointsAboveZero) {
		for (Skill skill : new FilteredIterator<>(getModel().getSelectionAsList(), Skill.class)) {
			if (!skill.canHaveChildren() && (!requirePointsAboveZero || skill.getPoints() > 0)) {
				return true;
			}
		}
		return false;
	}

	@SuppressWarnings("unused")
	@Override
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

	@SuppressWarnings("unused")
	@Override
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

	@Override
	protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof Skill;
	}

	@Override
	public void convertDragRowsToSelf(List<Row> list) {
		OutlineModel model = getModel();
		Row[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof GURPSCharacter || mDataFile instanceof Template;
		ArrayList<ListRow> process = forSheetOrTemplate ? new ArrayList<>() : null;

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
