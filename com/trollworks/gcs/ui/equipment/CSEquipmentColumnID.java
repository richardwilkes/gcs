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

package com.trollworks.gcs.ui.equipment;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.ui.common.CSHeaderCell;
import com.trollworks.gcs.ui.common.CSMultiCell;
import com.trollworks.gcs.ui.common.CSTextCell;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.units.TKWeightUnits;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKCell;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKTextCell;

import java.text.MessageFormat;

/** Definitions for equipment columns. */
public enum CSEquipmentColumnID {
	/** The equipment name/description. */
	DESCRIPTION(Msgs.EQUIPMENT, Msgs.EQUIPMENT_TOOLTIP) {
		@Override public String toString(CMCharacter character, boolean carried) {
			if (carried && character != null) {
				return MessageFormat.format(Msgs.CARRIED_EQUIPMENT, TKWeightUnits.POUNDS.format(character.getWeightCarried()), TKNumberUtils.format(character.getWealthCarried()));
			}
			return super.toString(character, carried);
		}

		@Override public TKCell getCell() {
			return new CSMultiCell();
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return true;
		}

		@Override public Object getData(CMEquipment equipment) {
			return getDataAsText(equipment);
		}

		@Override public String getDataAsText(CMEquipment equipment) {
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
	/** Whether the equipment is equipped or not (only for carried lists). */
	EQUIPPED(Msgs.EQUIPPED, Msgs.EQUIPPED_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.CENTER, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return carried && dataFile instanceof CMCharacter;
		}

		@Override public Object getData(CMEquipment equipment) {
			return equipment.isEquipped() ? Boolean.TRUE : Boolean.FALSE;
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return equipment.isFullyEquipped() ? Msgs.EQUIPPED : ""; //$NON-NLS-1$
		}
	},
	/** The quantity. */
	QUANTITY(Msgs.QUANTITY, Msgs.QUANTITY_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return !(dataFile instanceof CMListFile);
		}

		@Override public Object getData(CMEquipment equipment) {
			return new Integer(equipment.getQuantity());
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return TKNumberUtils.format(equipment.getQuantity());
		}
	},
	/** The tech level. */
	TL(Msgs.TECH_LEVEL, Msgs.TECH_LEVEL_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return dataFile instanceof CMListFile;
		}

		@Override public Object getData(CMEquipment equipment) {
			return equipment.getTechLevel();
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return equipment.getTechLevel();
		}
	},
	/** The legality class. */
	LC(Msgs.LEGALITY_CLASS, Msgs.LEGALITY_CLASS_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return dataFile instanceof CMListFile;
		}

		@Override public Object getData(CMEquipment equipment) {
			return equipment.getLegalityClass();
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return equipment.getLegalityClass();
		}
	},
	/** The value. */
	VALUE(Msgs.VALUE, Msgs.VALUE_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_FLOAT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return true;
		}

		@Override public Object getData(CMEquipment equipment) {
			return new Double(equipment.getValue());
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return TKNumberUtils.format(equipment.getValue());
		}
	},
	/** The weight. */
	WEIGHT(Msgs.WEIGHT, Msgs.WEIGHT_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_FLOAT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return true;
		}

		@Override public Object getData(CMEquipment equipment) {
			return new Double(equipment.getWeight());
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return TKNumberUtils.format(equipment.getWeight());
		}
	},
	/** The value. */
	EXT_VALUE(Msgs.EXT_VALUE, Msgs.EXT_VALUE_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_FLOAT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return !(dataFile instanceof CMListFile);
		}

		@Override public Object getData(CMEquipment equipment) {
			return new Double(equipment.getExtendedValue());
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return TKNumberUtils.format(equipment.getExtendedValue());
		}
	},
	/** The weight. */
	EXT_WEIGHT(Msgs.EXT_WEIGHT, Msgs.EXT_WEIGHT_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_FLOAT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return !(dataFile instanceof CMListFile);
		}

		@Override public Object getData(CMEquipment equipment) {
			return new Double(equipment.getExtendedWeight());
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return TKNumberUtils.format(equipment.getExtendedWeight());
		}
	},
	/** The page reference. */
	REFERENCE(Msgs.REFERENCE, Msgs.REFERENCE_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile, boolean carried) {
			return true;
		}

		@Override public Object getData(CMEquipment equipment) {
			return getDataAsText(equipment);
		}

		@Override public String getDataAsText(CMEquipment equipment) {
			return equipment.getReference();
		}
	};

	private String	mTitle;
	private String	mToolTip;

	private CSEquipmentColumnID(String title, String tooltip) {
		mTitle = title;
		mToolTip = tooltip;
	}

	@Override public String toString() {
		return mTitle;
	}

	/**
	 * @param equipment The {@link CMEquipment} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(CMEquipment equipment);

	/**
	 * @param equipment The {@link CMEquipment} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(CMEquipment equipment);

	/** @return The tooltip for the column. */
	public String getToolTip() {
		return mToolTip;
	}

	/** @return The {@link TKCell} used to display the data. */
	public abstract TKCell getCell();

	/**
	 * @param dataFile The {@link CMDataFile} to use.
	 * @param carried Whether this is for the "carried" equipment list.
	 * @return Whether this column should be displayed for the specified data file.
	 */
	public abstract boolean shouldDisplay(CMDataFile dataFile, boolean carried);

	/**
	 * @param character The {@link CMCharacter} this equipment list is associated with, or
	 *            <code>null</code>.
	 * @param carried Whether this is for the "carried" equipment list.
	 * @return The header title.
	 */
	public String toString(@SuppressWarnings("unused") CMCharacter character, @SuppressWarnings("unused") boolean carried) {
		return toString();
	}

	/**
	 * Adds all relevant {@link TKColumn}s to a {@link TKOutline}.
	 * 
	 * @param outline The {@link TKOutline} to use.
	 * @param dataFile The {@link CMDataFile} that data is being displayed for.
	 * @param carried Whether this is for the "carried" equipment list.
	 */
	public static void addColumns(TKOutline outline, CMDataFile dataFile, boolean carried) {
		CMCharacter character = dataFile instanceof CMCharacter ? (CMCharacter) dataFile : null;
		boolean sheetOrTemplate = dataFile instanceof CMCharacter || dataFile instanceof CMTemplate;
		TKOutlineModel model = outline.getModel();

		for (CSEquipmentColumnID one : values()) {
			if (one.shouldDisplay(dataFile, carried)) {
				TKColumn column = new TKColumn(one.ordinal(), one.toString(character, carried), one.getToolTip(), one.getCell());

				column.setHeaderCell(new CSHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
