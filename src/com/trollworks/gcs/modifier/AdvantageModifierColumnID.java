/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.toolkit.ui.widget.outline.Cell;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.TextCell;
import com.trollworks.toolkit.utility.I18n;

import javax.swing.SwingConstants;

/** AdvantageModifier Columns */
public enum AdvantageModifierColumnID {
    /** The enabled/disabled column. */
    ENABLED {
        @Override
        public String toString() {
            return I18n.Text("Enabled");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("Whether this modifier has been enabled or not");
        }

        @Override
        public Cell getCell(boolean forEditor) {
            if (forEditor) {
                return new TextCell(SwingConstants.CENTER, false);
            }
            return new ListTextCell(SwingConstants.CENTER, false);
        }

        @Override
        public String getDataAsText(AdvantageModifier modifier) {
            return modifier.isEnabled() ? "\u2713" : "";
        }
    },
    /** The advantage name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.Text("Enhancements & Limitations");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The name and notes describing this enhancement or limitation");
        }

        @Override
        public Cell getCell(boolean forEditor) {
            return new MultiCell(forEditor);
        }

        @Override
        public String getDataAsText(AdvantageModifier modifier) {
            StringBuilder builder = new StringBuilder();
            String        notes   = modifier.getNotes();

            builder.append(modifier.toString());
            if (!notes.isEmpty()) {
                builder.append(" (");
                builder.append(notes);
                builder.append(')');
            }
            return builder.toString();
        }
    },
    /** The total cost modifier. */
    COST_MODIFIER_TOTAL {
        @Override
        public String toString() {
            return I18n.Text("Cost Modifier");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The cost modifier for this enhancement or limitation");
        }

        @Override
        public Cell getCell(boolean forEditor) {
            if (forEditor) {
                return new TextCell(SwingConstants.LEFT, false);
            }
            return new ListTextCell(SwingConstants.LEFT, false);
        }

        @Override
        public String getDataAsText(AdvantageModifier modifier) {
            return modifier.getCostDescription();
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
            return I18n.Text("A reference to the book and page this modifier appears on (e.g. B22 would refer to \"Basic Set\", page 22)");
        }

        @Override
        public Cell getCell(boolean forEditor) {
            if (forEditor) {
                return new TextCell(SwingConstants.RIGHT, false);
            }
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public String getDataAsText(AdvantageModifier modifier) {
            return modifier.getReference();
        }
    };

    /**
     * @param modifier The {@link AdvantageModifier} to get the data from.
     * @return An object representing the data for this column.
     */
    public Object getData(AdvantageModifier modifier) {
        return getDataAsText(modifier);
    }

    /**
     * @param modifier The {@link AdvantageModifier} to get the data from.
     * @return Text representing the data for this column.
     */
    public abstract String getDataAsText(AdvantageModifier modifier);

    /** @return The tooltip for the column. */
    public abstract String getToolTip();

    /**
     * @param forEditor Whether this is for an editor or not.
     * @return The {@link Cell} used to display the data.
     */
    public abstract Cell getCell(boolean forEditor);

    /** @return Whether this column should be displayed for the specified data file. */
    @SuppressWarnings("static-method")
    public boolean shouldDisplay() {
        return true;
    }

    /**
     * Adds all relevant {@link Column}s to a {@link Outline}.
     *
     * @param outline   The {@link Outline} to use.
     * @param forEditor Whether this is for an editor or not.
     */
    public static void addColumns(Outline outline, boolean forEditor) {
        OutlineModel model = outline.getModel();

        for (AdvantageModifierColumnID one : values()) {
            if (one.shouldDisplay()) {
                Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell(forEditor));

                if (!forEditor) {
                    column.setHeaderCell(new ListHeaderCell(true));
                }
                model.addColumn(column);
            }
        }
    }

}
