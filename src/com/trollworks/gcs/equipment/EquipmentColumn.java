/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.widget.outline.Cell;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;

import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** Definitions for equipment columns. */
public enum EquipmentColumn {
	/** The equipment name/description. */
	DESCRIPTION {
		@Override
		public String toString() {
			return DESCRIPTION_TITLE;
		}

		@Override
		public String getToolTip() {
			return DESCRIPTION_TOOLTIP;
		}

		@Override
		public String toString(GURPSCharacter character) {
			if (character != null) {
				return MessageFormat.format(DESCRIPTION_TOTALS, character.getWeightCarried().toString(), Numbers.format(character.getWealthCarried()));
			}
			return super.toString(character);
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
		public Object getData(Equipment equipment) {
			return getDataAsText(equipment);
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			StringBuilder builder = new StringBuilder();
			String notes = equipment.getNotes();

			builder.append(equipment.toString());
			if (notes.length() > 0) {
				builder.append(" - "); //$NON-NLS-1$
				builder.append(notes);
			}
			return builder.toString();
		}
	},
	/** The current equipment state. */
	STATE {
		@Override
		public String toString() {
			return STATE_TITLE;
		}

		@Override
		public String getToolTip() {
			return STATE_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.CENTER, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Equipment equipment) {
			return equipment.getState();
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return equipment.getState().toShortName();
		}
	},
	/** The quantity. */
	QUANTITY {
		@Override
		public String toString() {
			return QUANTITY_TITLE;
		}

		@Override
		public String getToolTip() {
			return QUANTITY_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return !(dataFile instanceof ListFile) && !(dataFile instanceof LibraryFile);
		}

		@Override
		public Object getData(Equipment equipment) {
			return new Integer(equipment.getQuantity());
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return Numbers.format(equipment.getQuantity());
		}
	},
	/** The tech level. */
	TECH_LEVEL {
		@Override
		public String toString() {
			return TECH_LEVEL_TITLE;
		}

		@Override
		public String getToolTip() {
			return TECH_LEVEL_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
		}

		@Override
		public Object getData(Equipment equipment) {
			return equipment.getTechLevel();
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return equipment.getTechLevel();
		}
	},
	/** The legality class. */
	LEGALITY_CLASS {
		@Override
		public String toString() {
			return LEGALITY_CLASS_TITLE;
		}

		@Override
		public String getToolTip() {
			return LEGALITY_CLASS_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
		}

		@Override
		public Object getData(Equipment equipment) {
			return equipment.getLegalityClass();
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return equipment.getLegalityClass();
		}
	},
	/** The value. */
	VALUE {
		@Override
		public String toString() {
			return VALUE_TITLE;
		}

		@Override
		public String getToolTip() {
			return VALUE_TOOLTIP;
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
		public Object getData(Equipment equipment) {
			return new Double(equipment.getValue());
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return Numbers.format(equipment.getValue());
		}
	},
	/** The weight. */
	WEIGHT {
		@Override
		public String toString() {
			return WEIGHT_TITLE;
		}

		@Override
		public String getToolTip() {
			return WEIGHT_TOOLTIP;
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
		public Object getData(Equipment equipment) {
			return equipment.getWeight();
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return equipment.getWeight().toString();
		}
	},
	/** The value. */
	EXT_VALUE {
		@Override
		public String toString() {
			return EXT_VALUE_TITLE;
		}

		@Override
		public String getToolTip() {
			return EXT_VALUE_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return !(dataFile instanceof ListFile) && !(dataFile instanceof LibraryFile);
		}

		@Override
		public Object getData(Equipment equipment) {
			return new Double(equipment.getExtendedValue());
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return Numbers.format(equipment.getExtendedValue());
		}
	},
	/** The weight. */
	EXT_WEIGHT {
		@Override
		public String toString() {
			return EXT_WEIGHT_TITLE;
		}

		@Override
		public String getToolTip() {
			return EXT_WEIGHT_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return !(dataFile instanceof ListFile) && !(dataFile instanceof LibraryFile);
		}

		@Override
		public Object getData(Equipment equipment) {
			return equipment.getExtendedWeight();
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return equipment.getExtendedWeight().toString();
		}
	},
	/** The category. */
	CATEGORY {
		@Override
		public String toString() {
			return CATEGORY_TITLE;
		}

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
		public Object getData(Equipment equipment) {
			return getDataAsText(equipment);
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return equipment.getCategoriesAsString();
		}
	},
	/** The page reference. */
	REFERENCE {
		@Override
		public String toString() {
			return REFERENCE_TITLE;
		}

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
		public Object getData(Equipment equipment) {
			return getDataAsText(equipment);
		}

		@Override
		public String getDataAsText(Equipment equipment) {
			return equipment.getReference();
		}
	};

	@Localize("Equipment")
	static String	DESCRIPTION_TITLE;
	@Localize("The name and notes describing a piece of equipment")
	static String	DESCRIPTION_TOOLTIP;
	@Localize("Equipment ({0}; ${1})")
	static String	DESCRIPTION_TOTALS;
	@Localize("?")
	static String	STATE_TITLE;
	@Localize("Whether this piece of equipment is carried & equipped (E), just\ncarried (C), or not carried (-). Items that are not equipped do\nnot apply any features they may normally contribute to the\ncharacter.")
	static String	STATE_TOOLTIP;
	@Localize("TL")
	static String	TECH_LEVEL_TITLE;
	@Localize("The tech level of this piece of equipment")
	static String	TECH_LEVEL_TOOLTIP;
	@Localize("LC")
	static String	LEGALITY_CLASS_TITLE;
	@Localize("The legality class of this piece of equipment")
	static String	LEGALITY_CLASS_TOOLTIP;
	@Localize("#")
	static String	QUANTITY_TITLE;
	@Localize("The quantity of this piece of equipment")
	static String	QUANTITY_TOOLTIP;
	@Localize("$")
	static String	VALUE_TITLE;
	@Localize("The value of one of these pieces of equipment")
	static String	VALUE_TOOLTIP;
	@Localize("W")
	static String	WEIGHT_TITLE;
	@Localize("The weight of one of these pieces of equipment")
	static String	WEIGHT_TOOLTIP;
	@Localize("\u2211 $")
	static String	EXT_VALUE_TITLE;
	@Localize("The value of all of these pieces of equipment,\nplus the value of any contained equipment")
	static String	EXT_VALUE_TOOLTIP;
	@Localize("\u2211 W")
	static String	EXT_WEIGHT_TITLE;
	@Localize("The weight of all of these pieces of equipment\n, plus the weight of any contained equipment")
	static String	EXT_WEIGHT_TOOLTIP;
	@Localize("Category")
	static String	CATEGORY_TITLE;
	@Localize("The category or categories the equipment belongs to")
	static String	CATEGORY_TOOLTIP;
	@Localize("Ref")
	static String	REFERENCE_TITLE;
	@Localize("A reference to the book and page this equipment appears\non (e.g. B22 would refer to \"Basic Set\", page 22)")
	static String	REFERENCE_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * @param character The {@link GURPSCharacter} this equipment list is associated with, or
	 *            <code>null</code>.
	 * @return The header title.
	 */
	public String toString(GURPSCharacter character) {
		return toString();
	}

	/**
	 * @param equipment The {@link Equipment} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(Equipment equipment);

	/**
	 * @param equipment The {@link Equipment} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(Equipment equipment);

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
		GURPSCharacter character = dataFile instanceof GURPSCharacter ? (GURPSCharacter) dataFile : null;
		boolean sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
		OutlineModel model = outline.getModel();

		for (EquipmentColumn one : values()) {
			if (one.shouldDisplay(dataFile)) {
				Column column = new Column(one.ordinal(), one.toString(character), one.getToolTip(), one.getCell());

				column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
