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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.equipment.FontAwesomeCell;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.template.Template;
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

import javax.swing.SwingConstants;

/** Definitions for advantage columns. */
public enum AdvantageColumn {
    /** The advantage name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.text("Advantages & Disadvantages");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The name, level and notes describing an advantage");
        }

        @Override
        public String getToolTip(Advantage advantage) {
            StringBuilder builder  = new StringBuilder();
            SheetSettings settings = advantage.getDataFile().getSheetSettings();
            if (settings.userDescriptionDisplay().tooltip()) {
                String desc = advantage.getUserDesc();
                builder.append(desc);
                if (!desc.isEmpty()) {
                    builder.append('\n');
                }
            }
            if (settings.modifiersDisplay().tooltip()) {
                String desc = advantage.getModifierNotes();
                builder.append(desc);
                if (!desc.isEmpty()) {
                    builder.append('\n');
                }
            }
            if (settings.notesDisplay().tooltip()) {
                String desc = advantage.getNotes();
                builder.append(desc);
                if (!desc.isEmpty()) {
                    builder.append('\n');
                }
            }
            if (!builder.isEmpty()) {
                builder.setLength(builder.length() - 1);   // Remove the last '\n'
            }
            return builder.isEmpty() ? null : builder.toString();
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
        public Object getData(Advantage advantage) {
            return getDataAsText(advantage);
        }

        @Override
        public String getDataAsText(Advantage advantage) {
            StringBuilder builder = new StringBuilder();
            builder.append(advantage);
            SheetSettings settings = advantage.getDataFile().getSheetSettings();
            if (settings.userDescriptionDisplay().inline()) {
                String desc = advantage.getUserDesc();
                if (!desc.isEmpty()) {
                    builder.append(" - ");
                }
                builder.append(desc);
            }
            if (settings.modifiersDisplay().inline()) {
                String desc = advantage.getModifierNotes();
                if (!desc.isEmpty()) {
                    builder.append(" - ");
                }
                builder.append(desc);
            }
            if (settings.notesDisplay().inline()) {
                String desc = advantage.getNotes();
                if (!desc.isEmpty()) {
                    builder.append(" - ");
                }
                builder.append(desc);
            }
            return builder.toString();
        }
    },
    /** The points spent in the advantage. */
    POINTS {
        @Override
        public String toString() {
            return I18n.text("Pts");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The points spent in the advantage");
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
        public Object getData(Advantage advantage) {
            return Integer.valueOf(advantage.getAdjustedPoints());
        }

        @Override
        public String getDataAsText(Advantage advantage) {
            return Numbers.format(advantage.getAdjustedPoints());
        }
    },
    /** The type. */
    TYPE {
        @Override
        public String toString() {
            return I18n.text("Type");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The type of advantage");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.LEFT, true);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof ListFile;
        }

        @Override
        public Object getData(Advantage advantage) {
            return getDataAsText(advantage);
        }

        @Override
        public String getDataAsText(Advantage advantage) {
            return advantage.getTypeAsText();
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
            return I18n.text("The category or categories the advantage belongs to");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.LEFT, true);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof ListFile;
        }

        @Override
        public Object getData(Advantage advantage) {
            return getDataAsText(advantage);
        }

        @Override
        public String getDataAsText(Advantage advantage) {
            return advantage.getCategoriesAsString();
        }
    },
    /** The page reference. */
    REFERENCE {
        @Override
        public String toString() {
            return "\uf02e";
        }

        @Override
        public String getToolTip() {
            return PageRefCell.getStdToolTip(I18n.text("advantage"));
        }

        @Override
        public String getToolTip(Advantage advantage) {
            return PageRefCell.getStdCellToolTip(advantage.getReference());
        }

        @Override
        public Cell getCell() {
            return new PageRefCell();
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontAwesomeCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return true;
        }

        @Override
        public Object getData(Advantage advantage) {
            return getDataAsText(advantage);
        }

        @Override
        public String getDataAsText(Advantage advantage) {
            return advantage.getReference();
        }
    };

    /**
     * @param advantage The {@link Advantage} to get the data from.
     * @return An object representing the data for this column.
     */
    public abstract Object getData(Advantage advantage);

    /**
     * @param advantage The {@link Advantage} to get the data from.
     * @return Text representing the data for this column.
     */
    public abstract String getDataAsText(Advantage advantage);

    /** @return The tooltip for the column. */
    public abstract String getToolTip();

    /**
     * @param advantage The {@link Advantage} to get the data from.
     * @return The tooltip for a specific row within the column.
     */
    public String getToolTip(Advantage advantage) {
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
     * @return Whether this column should be displayed for the specified data file.
     */
    public abstract boolean shouldDisplay(DataFile dataFile);

    /**
     * Adds all relevant {@link Column}s to a {@link Outline}.
     *
     * @param outline  The {@link Outline} to use.
     * @param dataFile The {@link DataFile} that data is being displayed for.
     */
    public static void addColumns(Outline outline, DataFile dataFile) {
        boolean      sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
        OutlineModel model           = outline.getModel();
        for (AdvantageColumn one : values()) {
            if (one.shouldDisplay(dataFile)) {
                Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());
                column.setHeaderCell(one.getHeaderCell(sheetOrTemplate));
                model.addColumn(column);
            }
        }
    }
}
