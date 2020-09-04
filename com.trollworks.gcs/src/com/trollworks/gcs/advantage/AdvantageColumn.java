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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.IconsCell;
import com.trollworks.gcs.ui.widget.outline.ListHeaderCell;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.MultiCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.util.ArrayList;
import java.util.List;
import javax.swing.SwingConstants;

/** Definitions for advantage columns. */
public enum AdvantageColumn {
    /** The advantage name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.Text("Advantages & Disadvantages");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The name, level and notes describing an advantage");
        }

        @Override
        public String getToolTip(Advantage advantage) {
            StringBuilder builder = new StringBuilder();
            DataFile      df      = advantage.getDataFile();
            if (df.userDescriptionDisplay().tooltip()) {
                String desc = advantage.getUserDesc();
                builder.append(desc);
                if (!desc.isEmpty()) {
                    builder.append('\n');
                }
            }
            if (df.modifiersDisplay().tooltip()) {
                String desc = advantage.getModifierNotes();
                builder.append(desc);
                if (!desc.isEmpty()) {
                    builder.append('\n');
                }
            }
            if (df.notesDisplay().tooltip()) {
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
            DataFile df = advantage.getDataFile();
            if (df.userDescriptionDisplay().inline()) {
                String desc = advantage.getUserDesc();
                if (!desc.isEmpty()) {
                    builder.append(" - ");
                }
                builder.append(desc);
            }
            if (df.modifiersDisplay().inline()) {
                String desc = advantage.getModifierNotes();
                if (!desc.isEmpty()) {
                    builder.append(" - ");
                }
                builder.append(desc);
            }
            if (df.notesDisplay().inline()) {
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
            return I18n.Text("Pts");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The points spent in the advantage");
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
            return I18n.Text("Type");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The type of advantage");
        }

        @Override
        public Cell getCell() {
            return new IconsCell(SwingConstants.CENTER, SwingConstants.TOP);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof ListFile;
        }

        @Override
        public Object getData(Advantage advantage) {
            if (!advantage.canHaveChildren()) {
                int type = advantage.getType();
                if (type == 0) {
                    return null;
                }
                List<RetinaIcon> imgs = new ArrayList<>();
                if ((type & Advantage.TYPE_MASK_MENTAL) != 0) {
                    imgs.add(Images.MENTAL_TYPE);
                }
                if ((type & Advantage.TYPE_MASK_PHYSICAL) != 0) {
                    imgs.add(Images.PHYSICAL_TYPE);
                }
                if ((type & Advantage.TYPE_MASK_SOCIAL) != 0) {
                    imgs.add(Images.SOCIAL_TYPE);
                }
                if ((type & Advantage.TYPE_MASK_EXOTIC) != 0) {
                    imgs.add(Images.EXOTIC_TYPE);
                }
                if ((type & Advantage.TYPE_MASK_SUPERNATURAL) != 0) {
                    imgs.add(Images.SUPERNATURAL_TYPE);
                }
                return imgs;
            }
            return null;
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
            return I18n.Text("Category");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The category or categories the advantage belongs to");
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
            return I18n.Text("Ref");
        }

        @Override
        public String getToolTip() {
            return PageRefCell.getStdToolTip(I18n.Text("advantage"));
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
    @SuppressWarnings("static-method")
    public String getToolTip(Advantage advantage) {
        return null;
    }

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
     * @param outline  The {@link Outline} to use.
     * @param dataFile The {@link DataFile} that data is being displayed for.
     */
    public static void addColumns(Outline outline, DataFile dataFile) {
        boolean      sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
        OutlineModel model           = outline.getModel();
        for (AdvantageColumn one : values()) {
            if (one.shouldDisplay(dataFile)) {
                Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());
                column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
                model.addColumn(column);
            }
        }
    }
}
