/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.skill.SkillPointsTextCell;
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

import javax.swing.SwingConstants;

/** Definitions for spell columns. */
public enum SpellColumn {
    /** The spell name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.Text("Spells");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The name, tech level and notes describing the spell");
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
            String        notes   = spell.getNotes();

            builder.append(spell.toString());
            if (!notes.isEmpty()) {
                builder.append(" - ");
                builder.append(notes);
            }
            return builder.toString();
        }
    },
    /** The spell class/college. */
    CLASS {
        @Override
        public String toString() {
            return I18n.Text("Class");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The class and college of the spell");
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
                return spell.getSpellClass() + "; " + spell.getCollege();
            }
            return "";
        }
    },
    /** The casting time &amp; duration. */
    TIME {
        @Override
        public String toString() {
            return I18n.Text("Time");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The time required to cast the spell and its duration");
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
                return spell.getCastingTime() + "; " + spell.getDuration();
            }
            return "";
        }
    },
    /** The casting &amp; maintenance cost. */
    MANA_COST {
        @Override
        public String toString() {
            return I18n.Text("Mana Cost");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The mana cost to cast and maintain the spell");
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
                return spell.getCastingCost() + "; " + spell.getMaintenance();
            }
            return "";
        }
    },
    /** The spell level. */
    LEVEL {
        @Override
        public String toString() {
            return I18n.Text("SL");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The spell level");
        }

        @Override
        public String getToolTip(Spell spell) {
            return spell.getLevelToolTip();
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
            return Integer.valueOf(spell.canHaveChildren() ? -1 : spell.getLevel());
        }

        @Override
        public String getDataAsText(Spell spell) {
            int level;

            if (spell.canHaveChildren()) {
                return "";
            }
            level = spell.getLevel();
            if (level < 0) {
                return "-";
            }
            return Numbers.format(level);
        }
    },
    /** The relative spell level. */
    RELATIVE_LEVEL {
        @Override
        public String toString() {
            return I18n.Text("RSL");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The relative spell level");
        }

        @Override
        public String getToolTip(Spell spell) {
            return spell.getLevelToolTip();
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
            return Integer.valueOf(getRelativeLevel(spell));
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
                    return "-";
                }
                return spell.getAttribute().toString() + Numbers.formatWithForcedSign(level);
            }
            return "";
        }
    },
    /** The points spent in the spell. */
    POINTS {
        @Override
        public String toString() {
            return I18n.Text("Pts");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The points spent in the spell");
        }

        @Override
        public Cell getCell() {
            return new SkillPointsTextCell();
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof Template || dataFile instanceof GURPSCharacter;
        }

        @Override
        public Object getData(Spell spell) {
            return Integer.valueOf(spell.getPoints());
        }

        @Override
        public String getDataAsText(Spell spell) {
            return Numbers.format(spell.getPoints()); // $NON-NLS-1$
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
            return I18n.Text("The category or categories the spell belongs to");
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
        public String toString() {
            return I18n.Text("Ref");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("A reference to the book and page this spell appears on (e.g. B22 would refer to \"Basic Set\", page 22)");
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

    /**
     * @param spell The {@link Spell} to get the data from.
     * @return The tooltip for a specific row within the column.
     */
    @SuppressWarnings("static-method")
    public String getToolTip(Spell spell) {
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

        for (SpellColumn one : values()) {
            if (one.shouldDisplay(dataFile)) {
                Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

                column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
                model.addColumn(column);
            }
        }
    }
}
