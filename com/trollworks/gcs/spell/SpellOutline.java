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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.spell;

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
import com.trollworks.ttk.collections.FilteredIterator;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.Row;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for spells. */
public class SpellOutline extends ListOutline implements Incrementable {
	private static String	MSG_INCREMENT;
	private static String	MSG_DECREMENT;

	static {
		LocalizedMessages.initialize(SpellOutline.class);
	}

	private static OutlineModel extractModel(DataFile dataFile) {
		if (dataFile instanceof GURPSCharacter) {
			return ((GURPSCharacter) dataFile).getSpellsRoot();
		}
		if (dataFile instanceof Template) {
			return ((Template) dataFile).getSpellsModel();
		}
		if (dataFile instanceof LibraryFile) {
			return ((LibraryFile) dataFile).getSpellList().getModel();
		}
		return ((ListFile) dataFile).getModel();
	}

	/**
	 * Create a new spells outline.
	 * 
	 * @param dataFile The owning data file.
	 */
	public SpellOutline(DataFile dataFile) {
		this(dataFile, extractModel(dataFile));
	}

	/**
	 * Create a new spells outline.
	 * 
	 * @param dataFile The owning data file.
	 * @param model The {@link OutlineModel} to use.
	 */
	public SpellOutline(DataFile dataFile, OutlineModel model) {
		super(dataFile, model, Spell.ID_LIST_CHANGED);
		SpellColumn.addColumns(this, dataFile);
	}

	@Override
	public String getDecrementTitle() {
		return MSG_DECREMENT;
	}

	@Override
	public String getIncrementTitle() {
		return MSG_INCREMENT;
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
		for (Spell spell : new FilteredIterator<>(getModel().getSelectionAsList(), Spell.class)) {
			if (!spell.canHaveChildren() && (!requirePointsAboveZero || spell.getPoints() > 0)) {
				return true;
			}
		}
		return false;
	}

	@SuppressWarnings("unused")
	@Override
	public void decrement() {
		ArrayList<RowUndo> undos = new ArrayList<RowUndo>();

		for (Spell spell : new FilteredIterator<Spell>(getModel().getSelectionAsList(), Spell.class)) {
			if (!spell.canHaveChildren()) {
				int points = spell.getPoints();

				if (points > 0) {
					RowUndo undo = new RowUndo(spell);

					spell.setPoints(points - 1);
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

		for (Spell spell : new FilteredIterator<Spell>(getModel().getSelectionAsList(), Spell.class)) {
			if (!spell.canHaveChildren()) {
				RowUndo undo = new RowUndo(spell);

				spell.setPoints(spell.getPoints() + 1);
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
		return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof Spell;
	}

	@Override
	public void convertDragRowsToSelf(List<Row> list) {
		OutlineModel model = getModel();
		Row[] rows = model.getDragRows();
		boolean forSheetOrTemplate = mDataFile instanceof GURPSCharacter || mDataFile instanceof Template;
		ArrayList<ListRow> process = new ArrayList<>();

		for (Row element : rows) {
			Spell spell = new Spell(mDataFile, (Spell) element, true, forSheetOrTemplate);

			model.collectRowsAndSetOwner(list, spell, false);
			if (forSheetOrTemplate) {
				addRowsToBeProcessed(process, spell);
			}
		}

		if (forSheetOrTemplate && !process.isEmpty()) {
			EventQueue.invokeLater(new RowPostProcessor(this, process));
		}
	}
}
