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

package com.trollworks.gcs.spell;

import static com.trollworks.gcs.spell.SpellColumn_LS.*;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.widgets.outline.Cell;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "DESCRIPTION", msg = "Spells"),
				@LS(key = "DESCRIPTION_TOOLTIP", msg = "The name, tech level and notes describing the spell"),
				@LS(key = "CLASS", msg = "Class"),
				@LS(key = "CLASS_TOOLTIP", msg = "The class and college of the spell"),
				@LS(key = "MANA_COST", msg = "Mana Cost"),
				@LS(key = "MANA_COST_TOOLTIP", msg = "The mana cost to cast and maintain the spell"),
				@LS(key = "TIME", msg = "Time"),
				@LS(key = "TIME_TOOLTIP", msg = "THe time required to cast the spell and its duration"),
				@LS(key = "POINTS", msg = "Pts"),
				@LS(key = "POINTS_TOOLTIP", msg = "The points spent in the spell"),
				@LS(key = "LEVEL", msg = "SL"),
				@LS(key = "LEVEL_TOOLTIP", msg = "The spell level"),
				@LS(key = "RELATIVE_LEVEL", msg = "RSL"),
				@LS(key = "RELATIVE_LEVEL_TOOLTIP", msg = "The relative spell level"),
				@LS(key = "CATEGORY", msg = "Category"),
				@LS(key = "CATEGORY_TOOLTIP", msg = "The category or categories the spell belongs to"),
				@LS(key = "REFERENCE", msg = "Ref"),
				@LS(key = "REFERENCE_TOOLTIP", msg = "A reference to the book and page this spell appears\non (e.g. B22 would refer to \"Basic Set\", page 22)"),
})
/** Definitions for spell columns. */
public enum SpellColumn {
	/** The spell name/description. */
	DESCRIPTION {
		@Override
		public String getToolTip() {
			return DESCRIPTION_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new MultiCell();
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return true;
		}

		@Override
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			StringBuilder builder = new StringBuilder();
			String notes = spell.getNotes();

			builder.append(spell.toString());
			if (notes.length() > 0) {
				builder.append(" - "); //$NON-NLS-1$
				builder.append(notes);
			}
			return builder.toString();
		}
	},
	/** The spell class/college. */
	CLASS {
		@Override
		public String getToolTip() {
			return CLASS_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new SpellClassCell();
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return true;
		}

		@Override
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			if (!spell.canHaveChildren()) {
				StringBuilder builder = new StringBuilder();

				builder.append(spell.getSpellClass());
				builder.append("; "); //$NON-NLS-1$
				builder.append(spell.getCollege());
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The casting &amp; maintenance cost. */
	MANA_COST {
		@Override
		public String getToolTip() {
			return MANA_COST_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new SpellManaCostCell();
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return true;
		}

		@Override
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			if (!spell.canHaveChildren()) {
				StringBuilder builder = new StringBuilder();

				builder.append(spell.getCastingCost());
				builder.append("; "); //$NON-NLS-1$
				builder.append(spell.getMaintenance());
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The casting time &amp; duration. */
	TIME {
		@Override
		public String getToolTip() {
			return TIME_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new SpellTimeCell();
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return true;
		}

		@Override
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			if (!spell.canHaveChildren()) {
				StringBuilder builder = new StringBuilder();

				builder.append(spell.getCastingTime());
				builder.append("; "); //$NON-NLS-1$
				builder.append(spell.getDuration());
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The spell level. */
	LEVEL {
		@Override
		public String getToolTip() {
			return LEVEL_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Spell spell) {
			return new Integer(spell.canHaveChildren() ? -1 : spell.getLevel());
		}

		@Override
		public String getDataAsText(Spell spell) {
			int level;

			if (spell.canHaveChildren()) {
				return ""; //$NON-NLS-1$
			}
			level = spell.getLevel();
			if (level < 0) {
				return "-"; //$NON-NLS-1$
			}
			return Numbers.format(level);
		}
	},
	/** The relative spell level. */
	RELATIVE_LEVEL {
		@Override
		public String getToolTip() {
			return RELATIVE_LEVEL_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof Template || dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Spell spell) {
			return new Integer(getRelativeLevel(spell));
		}

		private int getRelativeLevel(Spell spell) {
			if (!spell.canHaveChildren()) {
				if (spell.getCharacter() != null) {
					if (spell.getLevel() < 0) {
						return Integer.MIN_VALUE;
					}
					return spell.getRelativeLevel();
				}
			}
			return Integer.MIN_VALUE;
		}

		@Override
		public String getDataAsText(Spell spell) {
			if (!spell.canHaveChildren()) {
				int level = getRelativeLevel(spell);

				if (level == Integer.MIN_VALUE) {
					return "-"; //$NON-NLS-1$
				}
				return "IQ" + Numbers.formatWithForcedSign(level); //$NON-NLS-1$
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The points spent in the spell. */
	POINTS {
		@Override
		public String getToolTip() {
			return POINTS_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof Template || dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Spell spell) {
			return new Integer(spell.canHaveChildren() ? -1 : spell.getPoints());
		}

		@Override
		public String getDataAsText(Spell spell) {
			return spell.canHaveChildren() ? "" : Numbers.format(spell.getPoints()); //$NON-NLS-1$
		}
	},
	/** The category. */
	CATEGORY {
		@Override
		public String getToolTip() {
			return CATEGORY_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.LEFT, true);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
		}

		@Override
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			return spell.getCategoriesAsString();
		}
	},
	/** The page reference. */
	REFERENCE {
		@Override
		public String getToolTip() {
			return REFERENCE_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return true;
		}

		@Override
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			return spell.getReference();
		}
	};

	@Override
	public String toString() {
		return SpellColumn_LS.toString(this);
	}

	/**
	 * @param spell The {@link Spell} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(Spell spell);

	/**
	 * @param spell The {@link Spell} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(Spell spell);

	/** @return The tooltip for the column. */
	public abstract String getToolTip();

	/** @return The {@link Cell} used to display the data. */
	public abstract Cell getCell();

	/**
	 * @param dataFile The {@link DataFile} to use.
	 * @return Whether this column should be displayed for the specified data file.
	 */
	public abstract boolean shouldDisplay(DataFile dataFile);

	/**
	 * Adds all relevant {@link Column}s to a {@link Outline}.
	 * 
	 * @param outline The {@link Outline} to use.
	 * @param dataFile The {@link DataFile} that data is being displayed for.
	 */
	public static void addColumns(Outline outline, DataFile dataFile) {
		boolean sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
		OutlineModel model = outline.getModel();

		for (SpellColumn one : values()) {
			if (one.shouldDisplay(dataFile)) {
				Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

				column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
