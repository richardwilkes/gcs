/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListHeaderCell;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.MultiCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.WeightValue;

import java.text.MessageFormat;
import javax.swing.SwingConstants;

/** Definitions for equipment columns. */
public enum EquipmentColumn {
    /** The equipped state. */
    EQUIPPED {
        @Override
        public String toString() {
            return I18n.Text("✓");
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
            return equipment.isEquipped() ? "✓" : "";
        }
    },
    /** The quantity. */
    QUANTITY {
        @Override
        public String toString() {
            return I18n.Text("Qty");
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
        public String getToolTip(Equipment equipment) {
            StringBuilder builder = new StringBuilder();
            DataFile      df      = equipment.getDataFile();
            if (df.modifiersDisplay().tooltip()) {
                String desc = equipment.getModifierNotes();
                builder.append(desc);
                if (!desc.isEmpty()) {
                    builder.append('\n');
                }
            }
            if (df.notesDisplay().tooltip()) {
                String desc = equipment.getNotes();
                builder.append(desc);
                if (!desc.isEmpty()) {
                    builder.append('\n');
                }
            }
            if (builder.length() > 0) {
                builder.setLength(builder.length() - 1);   // Remove the last '\n'
            }
            return builder.length() == 0 ? null : builder.toString();
        }

        @Override
        public String toString(DataFile dataFile, boolean carried) {
            if (dataFile instanceof GURPSCharacter) {
                GURPSCharacter character = (GURPSCharacter) dataFile;
                if (carried) {
                    return MessageFormat.format(I18n.Text("Carried Equipment ({0}; ${1})"), character.getWeightCarried().toString(), character.getWealthCarried().toLocalizedString());
                }
                return MessageFormat.format(I18n.Text("Other Equipment (${0})"), character.getWealthNotCarried().toLocalizedString());
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
            builder.append(equipment);
            DataFile df = equipment.getDataFile();
            if (df.modifiersDisplay().inline()) {
                String desc = equipment.getModifierNotes();
                if (!desc.isEmpty()) {
                    builder.append(" - ");
                }
                builder.append(desc);
            }
            if (df.notesDisplay().inline()) {
                String desc = equipment.getNotes();
                if (!desc.isEmpty()) {
                    builder.append(" - ");
                }
                builder.append(desc);
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
            return equipment.getAdjustedValue();
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return equipment.getAdjustedValue().toLocalizedString();
        }
    },
    /** The weight. */
    WEIGHT {
        @Override
        public String toString() {
            return I18n.Text("Weight");
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
            return equipment.getAdjustedWeight();
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return getDisplayWeight(equipment.getDataFile(), equipment.getAdjustedWeight());
        }
    },
    /** The value. */
    EXT_VALUE {
        @Override
        public String toString() {
            return I18n.Text("∑ $");
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
            return equipment.getExtendedValue();
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return equipment.getExtendedValue().toLocalizedString();
        }
    },
    /** The weight. */
    EXT_WEIGHT {
        @Override
        public String toString() {
            return I18n.Text("∑ Weight");
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
            return getDisplayWeight(equipment.getDataFile(), equipment.getExtendedWeight());
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
     * @param dataFile The {@link DataFile} this equipment list is associated with, or {@code
     *                 null}.
     * @param carried  {@code true} for the carried equipment, {@code false} for the other
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
     * @param carried  {@code true} for the carried equipment, {@code false} for the other
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
     * @param carried  {@code true} for the carried equipment, {@code false} for the other
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

    public static String getDisplayWeight(DataFile df, WeightValue weight) {
        return getConvertedWeight(df, weight).toString();
    }

    public static Fixed6 getNormalizedDisplayWeight(DataFile df, WeightValue weight) {
        return getConvertedWeight(df, weight).getNormalizedValue();
    }

    public static WeightValue getConvertedWeight(DataFile df, WeightValue weight) {
        if (df.useSimpleMetricConversions()) {
            weight = df.defaultWeightUnits().isMetric() ? GURPSCharacter.convertToGurpsMetric(weight) : GURPSCharacter.convertFromGurpsMetric(weight);
        } else {
            weight = new WeightValue(weight, df.defaultWeightUnits());
        }
        return weight;
    }
}
