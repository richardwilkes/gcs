/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.toolkit.ui.widget.outline.Cell;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.units.WeightUnits;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** Definitions for equipment columns. */
public enum EquipmentColumn {
    /** The equipped state. */
    EQUIPPED {
        @Override
        public String toString() {
            return I18n.Text("E");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("Whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character.");
        }

        @Override
        public Cell getCell() {
            return new CheckCell(SwingConstants.CENTER, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return carried && dataFile instanceof GURPSCharacter;
        }

        @Override
        public Object getData(Equipment equipment) {
            return equipment.isEquipped() ? Boolean.TRUE : Boolean.FALSE;
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return equipment.isEquipped() ? "\u2713" : "";
        }
    },
    /** The quantity. */
    QUANTITY {
        @Override
        public String toString() {
            return I18n.Text("#");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The quantity of this piece of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return !(dataFile instanceof ListFile);
        }

        @Override
        public Object getData(Equipment equipment) {
            return Integer.valueOf(equipment.getQuantity());
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return Numbers.format(equipment.getQuantity());
        }
    },
    /** The equipment name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.Text("Equipment");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The name and notes describing a piece of equipment");
        }

        @Override
        public String toString(DataFile dataFile, boolean carried) {
            if (dataFile instanceof GURPSCharacter) {
                GURPSCharacter character = (GURPSCharacter) dataFile;
                if (carried) {
                    return MessageFormat.format(I18n.Text("Carried Equipment ({0}; ${1})"), character.getWeightCarried().toString(), Numbers.format(character.getWealthCarried()));
                }
                return MessageFormat.format(I18n.Text("Other Equipment (${0})"), Numbers.format(character.getWealthNotCarried()));
            }
            if (dataFile instanceof Template) {
                return carried ? I18n.Text("Carried Equipment") : I18n.Text("Other Equipment");
            }
            return I18n.Text("Equipment");
        }

        @Override
        public Cell getCell() {
            return new MultiCell();
        }

        @Override
        public boolean isHierarchyColumn() {
            return true;
        }

        @Override
        public Object getData(Equipment equipment) {
            return getDataAsText(equipment);
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            StringBuilder builder = new StringBuilder();
            String        notes   = equipment.getNotes();

            builder.append(equipment.toString());
            if (notes.length() > 0) {
                builder.append(" - ");
                builder.append(notes);
            }
            return builder.toString();
        }
    },
    /** The uses remaining. */
    USES {
        @Override
        public String toString() {
            return I18n.Text("Uses");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The number of uses remaining");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public Object getData(Equipment equipment) {
            return Integer.valueOf(equipment.getUses());
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return equipment.getMaxUses() > 0 ? Numbers.format(equipment.getUses()) : "";
        }
    },
    /** The tech level. */
    TECH_LEVEL {
        @Override
        public String toString() {
            return I18n.Text("TL");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The tech level of this piece of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return dataFile instanceof ListFile;
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
            return I18n.Text("LC");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The legality class of this piece of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return dataFile instanceof ListFile;
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
            return I18n.Text("$");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The value of one of these pieces of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public Object getData(Equipment equipment) {
            return Double.valueOf(equipment.getValue());
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
            return I18n.Text("W");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The weight of one of these pieces of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public Object getData(Equipment equipment) {
            return equipment.getWeight();
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return getDisplayWeight(equipment.getWeight());
        }
    },
    /** The value. */
    EXT_VALUE {
        @Override
        public String toString() {
            return I18n.Text("\u2211 $");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The value of all of these pieces of equipment, plus the value of any contained equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return !(dataFile instanceof ListFile);
        }

        @Override
        public Object getData(Equipment equipment) {
            return Double.valueOf(equipment.getExtendedValue());
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
            return I18n.Text("\u2211 W");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The weight of all of these pieces of equipment, plus the weight of any contained equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return !(dataFile instanceof ListFile);
        }

        @Override
        public Object getData(Equipment equipment) {
            return equipment.getExtendedWeight();
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return getDisplayWeight(equipment.getExtendedWeight());
        }
    },
    /** The category. */
    CATEGORY {
        @Override
        public String toString() {
            return I18n.Text("Category");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The category or categories the equipment belongs to");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.LEFT, true);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return dataFile instanceof ListFile;
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
            return I18n.Text("Ref");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("A reference to the book and page this equipment appears on (e.g. B22 would refer to \"Basic Set\", page 22)");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
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

    /**
     * @param dataFile The {@link DataFile} this equipment list is associated with, or
     *                 <code>null</code>.
     * @param carried  <code>true</code> for the carried equipment, <code>false</code> for the other
     *                 equipment.
     * @return The header title.
     */
    public String toString(DataFile dataFile, boolean carried) {
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

    /**
     * @param equipment The {@link Equipment} to get the data from.
     * @return The tooltip for a specific row within the column.
     */
    @SuppressWarnings("static-method")
    public String getToolTip(Equipment equipment) {
        return null;
    }

    /** @return The {@link Cell} used to display the data. */
    public abstract Cell getCell();

    /**
     * @param dataFile The {@link DataFile} to use.
     * @param carried  <code>true</code> for the carried equipment, <code>false</code> for the other
     *                 equipment.
     * @return Whether this column should be displayed for the specified data file.
     */
    @SuppressWarnings("static-method")
    public boolean shouldDisplay(DataFile dataFile, boolean carried) {
        return true;
    }

    /** @return Whether this column should contain the hierarchy controls. */
    @SuppressWarnings("static-method")
    public boolean isHierarchyColumn() {
        return false;
    }

    /**
     * Adds all relevant {@link Column}s to a {@link Outline}.
     *
     * @param outline  The {@link Outline} to use.
     * @param dataFile The {@link DataFile} that data is being displayed for.
     * @param carried  <code>true</code> for the carried equipment, <code>false</code> for the other
     *                 equipment.
     */
    public static void addColumns(Outline outline, DataFile dataFile, boolean carried) {
        boolean      sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
        OutlineModel model           = outline.getModel();
        for (EquipmentColumn one : values()) {
            if (one.shouldDisplay(dataFile, carried)) {
                Column column = new Column(one.ordinal(), one.toString(dataFile, carried), one.getToolTip(), one.getCell());
                column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
                model.addColumn(column);
                if (one.isHierarchyColumn()) {
                    model.setHierarchyColumn(column);
                }
            }
        }
    }

    public static String getDisplayWeight(WeightValue weight) {
        return getConvertedWeight(weight).toString();
    }

    public static double getNormalizedDisplayWeight(WeightValue weight) {
        return getConvertedWeight(weight).getNormalizedValue();
    }

    public static WeightValue getConvertedWeight(WeightValue weight) {
        WeightUnits defaultWeightUnits = DisplayPreferences.getWeightUnits();
        if (SheetPreferences.areGurpsMetricRulesUsed()) {
            if (defaultWeightUnits.isMetric()) {
                weight = GURPSCharacter.convertToGurpsMetric(weight);
            } else {
                weight = GURPSCharacter.convertFromGurpsMetric(weight);
            }
        } else {
            weight = new WeightValue(weight, defaultWeightUnits);
        }
        return weight;
    }
}
