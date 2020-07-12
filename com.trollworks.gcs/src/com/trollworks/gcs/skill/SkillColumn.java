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

package com.trollworks.gcs.skill;

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
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import javax.swing.SwingConstants;

/** Definitions for skill columns. */
public enum SkillColumn {
    /** The skill name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return I18n.Text("Skills");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The name, specialty, tech level and notes describing a skill");
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
            StringBuilder builder = new StringBuilder();
            String        notes   = skill.getNotes();

            builder.append(skill.toString());
            if (!notes.isEmpty()) {
                builder.append(" - ");
                builder.append(notes);
            }
            return builder.toString();
        }
    },
    /** The skill difficulty. */
    DIFFICULTY {
        @Override
        public String toString() {
            return I18n.Text("Diff");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The skill difficulty");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof ListFile;
        }

        @Override
        public Object getData(Skill skill) {
            return Integer.valueOf(skill.canHaveChildren() ? -1 : skill.getDifficulty().ordinal() + (skill.getAttribute().ordinal() << 8));
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
            return I18n.Text("SL");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The skill level");
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
            return I18n.Text("RSL");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The relative skill level");
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
            return Integer.valueOf(getRelativeLevel(skill));
        }

        private int getRelativeLevel(Skill skill) {
            if (!skill.canHaveChildren()) {
                if (skill.getCharacter() != null) {
                    if (skill.getLevel() < 0) {
                        return Integer.MIN_VALUE;
                    }
                    int level = skill.getRelativeLevel();
                    if (skill instanceof Technique) {
                        level += ((Technique) skill).getDefault().getModifier();
                    }
                    return level;
                } else if (skill.getTemplate() != null) {
                    int points = skill.getPoints();
                    if (points > 0) {
                        SkillDifficulty difficulty = skill.getDifficulty();
                        int             level;
                        if (skill instanceof Technique) {
                            if (difficulty != SkillDifficulty.A) {
                                points--;
                            }
                            return points + ((Technique) skill).getDefault().getModifier();
                        }
                        level = difficulty.getBaseRelativeLevel();
                        if (difficulty == SkillDifficulty.W) {
                            points /= 3;
                        }
                        if (points > 1) {
                            if (points < 4) {
                                level++;
                            } else {
                                level += 1 + points / 4;
                            }
                        }
                        return level;
                    }
                }
            }
            return Integer.MIN_VALUE;
        }

        @Override
        public String getDataAsText(Skill skill) {
            if (!skill.canHaveChildren()) {
                int           level = getRelativeLevel(skill);
                StringBuilder builder;

                if (level == Integer.MIN_VALUE) {
                    return "-";
                }
                builder = new StringBuilder();
                if (!(skill instanceof Technique)) {
                    builder.append(skill.getAttribute());
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
            return I18n.Text("Pts");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The points spent in the skill");
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
            return I18n.Text("Category");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The category or categories the skill belongs to");
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
            return I18n.Text("Ref");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("A reference to the book and page this skill appears on (e.g. B22 would refer to \"Basic Set\", page 22)");
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
    @SuppressWarnings("static-method")
    public String getToolTip(Skill skill) {
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
        for (SkillColumn one : values()) {
            if (one.shouldDisplay(dataFile)) {
                Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());
                column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
                model.addColumn(column);
            }
        }
    }
}
