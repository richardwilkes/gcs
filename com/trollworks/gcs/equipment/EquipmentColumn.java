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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.Cell;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;

import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** Definitions for equipment columns. */
public enum EquipmentColumn {
	/** The equipment name/description. */
	DESCRIPTION {
		@Override
		public String toString() {
			return MSG_EQUIPMENT;
		}

		@Override
		public String getToolTip() {
			return MSG_EQUIPMENT_TOOLTIP;
		}

		@Override
		public String toString(GURPSCharacter character) {
			if (character != null) {
				return MessageFormat.format(MSG_EQUIPMENT_TOTALS, character.getWeightCarried().toString(), Numbers.format(character.getWealthCarried()));
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
			return MSG_STATE;
		}

		@Override
		public String getToolTip() {
			return MSG_STATE_TOOLTIP;
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
			return equipment.getState().toShortString();
		}
	},
	/** The quantity. */
	QUANTITY {
		@Override
		public String toString() {
			return MSG_QUANTITY;
		}

		@Override
		public String getToolTip() {
			return MSG_QUANTITY_TOOLTIP;
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
	TL {
		@Override
		public String toString() {
			return MSG_TECH_LEVEL;
		}

		@Override
		public String getToolTip() {
			return MSG_TECH_LEVEL_TOOLTIP;
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
	LC {
		@Override
		public String toString() {
			return MSG_LEGALITY_CLASS;
		}

		@Override
		public String getToolTip() {
			return MSG_LEGALITY_CLASS_TOOLTIP;
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
			return MSG_VALUE;
		}

		@Override
		public String getToolTip() {
			return MSG_VALUE_TOOLTIP;
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
			return MSG_WEIGHT;
		}

		@Override
		public String getToolTip() {
			return MSG_WEIGHT_TOOLTIP;
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
			return MSG_EXT_VALUE;
		}

		@Override
		public String getToolTip() {
			return MSG_EXT_VALUE_TOOLTIP;
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
			return MSG_EXT_WEIGHT;
		}

		@Override
		public String getToolTip() {
			return MSG_EXT_WEIGHT_TOOLTIP;
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
			return MSG_CATEGORY;
		}

		@Override
		public String getToolTip() {
			return MSG_CATEGORY_TOOLTIP;
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
			return MSG_REFERENCE;
		}

		@Override
		public String getToolTip() {
			return MSG_REFERENCE_TOOLTIP;
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

	static String	MSG_EQUIPMENT;
	static String	MSG_EQUIPMENT_TOTALS;
	static String	MSG_EQUIPMENT_TOOLTIP;
	static String	MSG_STATE;
	static String	MSG_STATE_TOOLTIP;
	static String	MSG_TECH_LEVEL;
	static String	MSG_TECH_LEVEL_TOOLTIP;
	static String	MSG_LEGALITY_CLASS;
	static String	MSG_LEGALITY_CLASS_TOOLTIP;
	static String	MSG_QUANTITY;
	static String	MSG_QUANTITY_TOOLTIP;
	static String	MSG_VALUE;
	static String	MSG_VALUE_TOOLTIP;
	static String	MSG_WEIGHT;
	static String	MSG_WEIGHT_TOOLTIP;
	static String	MSG_EXT_VALUE;
	static String	MSG_EXT_VALUE_TOOLTIP;
	static String	MSG_EXT_WEIGHT;
	static String	MSG_EXT_WEIGHT_TOOLTIP;
	static String	MSG_CATEGORY;
	static String	MSG_CATEGORY_TOOLTIP;
	static String	MSG_REFERENCE;
	static String	MSG_REFERENCE_TOOLTIP;

	static {
		LocalizedMessages.initialize(EquipmentColumn.class);
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
	 * @param character The {@link GURPSCharacter} this equipment list is associated with, or
	 *            <code>null</code>.
	 * @return The header title.
	 */
	public String toString(GURPSCharacter character) {
		return toString();
	}

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
