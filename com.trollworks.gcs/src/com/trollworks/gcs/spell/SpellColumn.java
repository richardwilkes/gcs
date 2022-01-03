/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.equipment.FontIconCell;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillPointsTextCell;
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

import javax.swing.SwingConstants;

/** Definitions for spell columns. */
public enum SpellColumn {
    /** The spell name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.text("Spells");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The name, tech level and notes describing the spell");
        }

        @Override
        public String getToolTip(Spell spell) {
            return spell.getDescriptionToolTipText();
        }

        @Override
        public Cell getCell() {
            return new MultiCell();
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            return spell.getDescriptionText();
        }
    },
    /** The resistance. */
    RESIST {
        @Override
        public String toString() {
            return I18n.text("Resist");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The resistance");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return spell.getResist();
            }
            return "";
        }
    },
    /** The spell class. */
    CLASS {
        @Override
        public String toString() {
            return I18n.text("Class");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The class of the spell");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return spell.getSpellClass();
            }
            return "";
        }
    },
    /** The spell college. */
    COLLEGE {
        @Override
        public String toString() {
            return I18n.text("College");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The college of the spell");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return String.join(", ", spell.getColleges());
            }
            return "";
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile.getSheetSettings().showCollegeInSpells();
        }
    },
    /** The casting cost. */
    MANA_COST {
        @Override
        public String toString() {
            return I18n.text("Cost");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The mana cost to cast the spell");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return spell.getCastingCost();
            }
            return "";
        }
    },
    /** The maintenance cost. */
    MAINTAIN {
        @Override
        public String toString() {
            return I18n.text("Maintain");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The mana cost to maintain the spell");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return spell.getMaintenance();
            }
            return "";
        }
    },
    /** The casting time. */
    TIME {
        @Override
        public String toString() {
            return I18n.text("Time");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The time required to cast the spell");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return spell.getCastingTime();
            }
            return "";
        }
    },
    /** The spell duration. */
    DURATION {
        @Override
        public String toString() {
            return I18n.text("Duration");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The spell duration");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return spell.getDuration();
            }
            return "";
        }
    },
    /** The difficulty. */
    DIFFICULTY {
        @Override
        public String toString() {
            return I18n.text("Difficulty");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The difficulty of the spell");
        }

        @Override
        public Object getData(Spell spell) {
            return getDataAsText(spell);
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                return spell.getDifficultyAsText(true);
            }
            return "";
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            if (dataFile instanceof ListFile) {
                return true;
            }
            return dataFile.getSheetSettings().showDifficulty();
        }
    },
    /** The spell level. */
    LEVEL {
        @Override
        public String toString() {
            return I18n.text("SL");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The spell level");
        }

        @Override
        public String getToolTip(Spell spell) {
            if (spell.getDataFile().getSheetSettings().skillLevelAdjustmentsDisplay().tooltip()) {
                return spell.getLevelToolTip();
            }
            return null;
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
            return I18n.text("RSL");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The relative spell level");
        }

        @Override
        public String getToolTip(Spell spell) {
            if (spell.getDataFile().getSheetSettings().skillLevelAdjustmentsDisplay().tooltip()) {
                return spell.getLevelToolTip();
            }
            return null;
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
            return Integer.valueOf(spell.getAdjustedRelativeLevel());
        }

        @Override
        public String getDataAsText(Spell spell) {
            if (!spell.canHaveChildren()) {
                int level = spell.getAdjustedRelativeLevel();
                if (level == Integer.MIN_VALUE) {
                    return "-";
                }
                StringBuilder builder = new StringBuilder();
                if (!(spell instanceof RitualMagicSpell)) {
                    builder.append(Skill.resolveAttributeName(spell.getDataFile(), spell.getAttribute()));
                }
                builder.append(Numbers.formatWithForcedSign(level));
                return builder.toString();
            }
            return "";
        }
    },
    /** The points spent in the spell. */
    POINTS {
        @Override
        public String toString() {
            return I18n.text("Pts");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The points spent in the spell");
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

        @Override
        public String getToolTip(Spell spell) {
            return spell.getPointsToolTip();
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
            return I18n.text("The category or categories the spell belongs to");
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
            return FontAwesome.BOOKMARK;
        }

        @Override
        public String getToolTip() {
            return PageRefCell.getStdToolTip(I18n.text("spell"));
        }

        @Override
        public String getToolTip(Spell spell) {
            return PageRefCell.getStdCellToolTip(spell.getReference());
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
    public String getToolTip(Spell spell) {
        return null;
    }

    /** @return The {@link Cell} used to display the data. */
    public Cell getCell() {
        return new ListTextCell(SwingConstants.LEFT, true);
    }

    /** @return The {@link Cell} used to display the header. */
    public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
        return new ListHeaderCell(sheetOrTemplate);
    }

    /**
     * @param dataFile The {@link DataFile} to use.
     * @return Whether this column should be displayed for the specified data file.
     */
    public boolean shouldDisplay(DataFile dataFile) {
        return true;
    }

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
                column.setHeaderCell(one.getHeaderCell(sheetOrTemplate));
                model.addColumn(column);
            }
        }
    }
}
