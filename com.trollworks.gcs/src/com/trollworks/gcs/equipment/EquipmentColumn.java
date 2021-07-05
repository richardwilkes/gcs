/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.HeaderCell;
import com.trollworks.gcs.ui.widget.outline.ListHeaderCell;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.MultiCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
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
            return FontAwesome.CHECK_CIRCLE;
        }

        @Override
        public String getToolTip() {
            return I18n.text("Whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character.");
        }

        @Override
        public Cell getCell() {
            return new CheckCell(SwingConstants.CENTER, false);
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
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
            return equipment.isEquipped() ? FontAwesome.CHECK_CIRCLE : "";
        }
    },
    /** The quantity. */
    QUANTITY {
        @Override
        public String toString() {
            return FontAwesome.SLACK_HASH;
        }

        @Override
        public String getToolTip() {
            return I18n.text("The quantity of this piece of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_BRANDS, sheetOrTemplate);
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
            return I18n.text("Equipment");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The name and notes describing a piece of equipment");
        }

        @Override
        public String getToolTip(Equipment equipment) {
            return equipment.getDescriptionToolTipText();
        }

        @Override
        public String toString(DataFile dataFile, boolean carried) {
            if (dataFile instanceof GURPSCharacter) {
                GURPSCharacter character = (GURPSCharacter) dataFile;
                if (carried) {
                    return MessageFormat.format(I18n.text("Carried Equipment ({0}; ${1})"), character.getWeightCarried(false).toString(), character.getWealthCarried().toLocalizedString());
                }
                return MessageFormat.format(I18n.text("Other Equipment (${0})"), character.getWealthNotCarried().toLocalizedString());
            }
            if (dataFile instanceof Template) {
                return carried ? I18n.text("Carried Equipment") : I18n.text("Other Equipment");
            }
            return I18n.text("Equipment");
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
            return equipment.getDescriptionText();
        }
    },
    /** The uses remaining. */
    USES {
        @Override
        public String toString() {
            return I18n.text("Uses");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The number of uses remaining");
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
            return I18n.text("TL");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The tech level of this piece of equipment");
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
            return I18n.text("LC");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The legality class of this piece of equipment");
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
            return FontAwesome.DOLLAR_SIGN;
        }

        @Override
        public String getToolTip() {
            return I18n.text("The value of one of these pieces of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
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
            return FontAwesome.WEIGHT_HANGING;
        }

        @Override
        public String getToolTip() {
            return I18n.text("The weight of one of these pieces of equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
        }

        @Override
        public Object getData(Equipment equipment) {
            return equipment.getAdjustedWeight(false);
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return getDisplayWeight(equipment.getDataFile(), equipment.getAdjustedWeight(false));
        }
    },
    /** The value. */
    EXT_VALUE {
        @Override
        public String toString() {
            return FontAwesome.LAYER_GROUP + " " + FontAwesome.DOLLAR_SIGN;
        }

        @Override
        public String getToolTip() {
            return I18n.text("The value of all of these pieces of equipment, plus the value of any contained equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
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
            return FontAwesome.LAYER_GROUP + " " + FontAwesome.WEIGHT_HANGING;
        }

        @Override
        public String getToolTip() {
            return I18n.text("The weight of all of these pieces of equipment, plus the weight of any contained equipment");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile, boolean carried) {
            return !(dataFile instanceof ListFile);
        }

        @Override
        public Object getData(Equipment equipment) {
            return equipment.getExtendedWeight(false);
        }

        @Override
        public String getDataAsText(Equipment equipment) {
            return getDisplayWeight(equipment.getDataFile(), equipment.getExtendedWeight(false));
        }
    },
    /** The category. */
    CATEGORY {
        @Override
        public String toString() {
            return I18n.text("Category");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The category or categories the equipment belongs to");
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
            return FontAwesome.BOOKMARK;
        }

        @Override
        public String getToolTip() {
            return PageRefCell.getStdToolTip(I18n.text("equipment"));
        }

        @Override
        public String getToolTip(Equipment equipment) {
            return PageRefCell.getStdCellToolTip(equipment.getReference());
        }

        @Override
        public Cell getCell() {
            return new PageRefCell();
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
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
    public String getToolTip(Equipment equipment) {
        return null;
    }

    /** @return The {@link Cell} used to display the data. */
    public abstract Cell getCell();

    /** @return The {@link Cell} used to display the header. */
    public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
        return new ListHeaderCell(sheetOrTemplate);
    }

    /**
     * @param dataFile The {@link DataFile} to use.
     * @param carried  {@code true} for the carried equipment, {@code false} for the other
     *                 equipment.
     * @return Whether this column should be displayed for the specified data file.
     */
    public boolean shouldDisplay(DataFile dataFile, boolean carried) {
        return true;
    }

    /** @return Whether this column should contain the hierarchy controls. */
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
                column.setHeaderCell(one.getHeaderCell(sheetOrTemplate));
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

    public static WeightValue getConvertedWeight(DataFile df, WeightValue weight) {
        SheetSettings settings = df.getSheetSettings();
        if (settings.useSimpleMetricConversions()) {
            weight = settings.defaultWeightUnits().isMetric() ? GURPSCharacter.convertToGurpsMetric(weight) : GURPSCharacter.convertFromGurpsMetric(weight);
        } else {
            weight = new WeightValue(weight, settings.defaultWeightUnits());
        }
        return weight;
    }
}
