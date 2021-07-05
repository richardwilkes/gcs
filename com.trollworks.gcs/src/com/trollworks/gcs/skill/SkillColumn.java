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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.equipment.FontIconCell;
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

/** Definitions for skill columns. */
public enum SkillColumn {
    /** The skill name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.text("Skills");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The name, specialty, tech level and notes describing a skill");
        }

        @Override
        public String getToolTip(Skill skill) {
            return skill.getDescriptionToolTipText();
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
        public Object getData(Skill skill) {
            return getDataAsText(skill);
        }

        @Override
        public String getDataAsText(Skill skill) {
            return skill.getDescriptionText();
        }
    },
    /** The skill difficulty. */
    DIFFICULTY {
        @Override
        public String toString() {
            return I18n.text("Diff");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The skill difficulty");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.LEFT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            if (dataFile instanceof ListFile) {
                return true;
            }
            return dataFile.getSheetSettings().showDifficulty();
        }

        @Override
        public Object getData(Skill skill) {
            return getDataAsText(skill);
        }

        @Override
        public String getDataAsText(Skill skill) {
            return skill.getDifficultyAsText();
        }
    },
    /** The skill level. */
    LEVEL {
        @Override
        public String toString() {
            return I18n.text("SL");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The skill level");
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
        public Object getData(Skill skill) {
            return Integer.valueOf(skill.canHaveChildren() ? -1 : skill.getLevel());
        }

        @Override
        public String getDataAsText(Skill skill) {
            int level;

            if (skill.canHaveChildren()) {
                return "";
            }
            level = skill.getLevel();
            if (level < 0) {
                return "-";
            }
            return Numbers.format(level);
        }

        @Override
        public String getToolTip(Skill skill) {
            return skill.getLevelToolTip();
        }
    },
    /** The relative skill level. */
    RELATIVE_LEVEL {
        @Override
        public String toString() {
            return I18n.text("RSL");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The relative skill level");
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
        public Object getData(Skill skill) {
            return Integer.valueOf(skill.getAdjustedRelativeLevel());
        }

        @Override
        public String getDataAsText(Skill skill) {
            if (!skill.canHaveChildren()) {
                int           level = skill.getAdjustedRelativeLevel();
                StringBuilder builder;
                if (level == Integer.MIN_VALUE) {
                    return "-";
                }
                builder = new StringBuilder();
                if (!(skill instanceof Technique)) {
                    builder.append(Skill.resolveAttributeName(skill.getDataFile(), skill.getAttribute()));
                }
                builder.append(Numbers.formatWithForcedSign(level));
                return builder.toString();
            }
            return "";
        }

        @Override
        public String getToolTip(Skill skill) {
            return skill.getLevelToolTip();
        }
    },
    /** The points spent in the skill. */
    POINTS {
        @Override
        public String toString() {
            return I18n.text("Pts");
        }

        @Override
        public String getToolTip() {
            return I18n.text("The points spent in the skill");
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
        public Object getData(Skill skill) {
            return Integer.valueOf(skill.getPoints());
        }

        @Override
        public String getDataAsText(Skill skill) {
            return Numbers.format(skill.getPoints());
        }

        @Override
        public String getToolTip(Skill skill) {
            return skill.getPointsToolTip();
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
            return I18n.text("The category or categories the skill belongs to");
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
        public Object getData(Skill skill) {
            return getDataAsText(skill);
        }

        @Override
        public String getDataAsText(Skill skill) {
            return skill.getCategoriesAsString();
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
            return PageRefCell.getStdToolTip(I18n.text("skill"));
        }

        @Override
        public String getToolTip(Skill skill) {
            return PageRefCell.getStdCellToolTip(skill.getReference());
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
        public boolean shouldDisplay(DataFile dataFile) {
            return true;
        }

        @Override
        public Object getData(Skill skill) {
            return getDataAsText(skill);
        }

        @Override
        public String getDataAsText(Skill skill) {
            return skill.getReference();
        }
    };

    /**
     * @param skill The {@link Skill} to get the data from.
     * @return An object representing the data for this column.
     */
    public abstract Object getData(Skill skill);

    /**
     * @param skill The {@link Skill} to get the data from.
     * @return Text representing the data for this column.
     */
    public abstract String getDataAsText(Skill skill);

    /** @return The tooltip for the column. */
    public abstract String getToolTip();

    /**
     * @param skill The {@link Skill} to get the data from.
     * @return The tooltip for a specific row within the column.
     */
    public String getToolTip(Skill skill) {
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
        for (SkillColumn one : values()) {
            if (one.shouldDisplay(dataFile)) {
                Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());
                column.setHeaderCell(one.getHeaderCell(sheetOrTemplate));
                model.addColumn(column);
            }
        }
    }
}
