/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.widget.outline.Cell;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;

import javax.swing.SwingConstants;

/** Definitions for skill columns. */
public enum SkillColumn {
    /** The skill name/description. */
    DESCRIPTION {
        @Override
        public String toString() {
            return DESCRIPTION_TITLE;
        }

        @Override
        public String getToolTip() {
            return DESCRIPTION_TOOLTIP;
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
            if (notes.length() > 0) {
                builder.append(" - "); //$NON-NLS-1$
                builder.append(notes);
            }
            return builder.toString();
        }
    },
    /** The skill difficulty. */
    DIFFICULTY {
        @Override
        public String toString() {
            return DIFFICULTY_TITLE;
        }

        @Override
        public String getToolTip() {
            return DIFFICULTY_TOOLTIP;
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
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
            return LEVEL_TITLE;
        }

        @Override
        public String getToolTip() {
            return LEVEL_TOOLTIP;
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
                return ""; //$NON-NLS-1$
            }
            level = skill.getLevel();
            if (level < 0) {
                return "-"; //$NON-NLS-1$
            }
            return Numbers.format(level);
        }
    },
    /** The relative skill level. */
    RELATIVE_LEVEL {
        @Override
        public String toString() {
            return RELATIVE_LEVEL_TITLE;
        }

        @Override
        public String getToolTip() {
            return RELATIVE_LEVEL_TOOLTIP;
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
                    return "-"; //$NON-NLS-1$
                }
                builder = new StringBuilder();
                if (!(skill instanceof Technique)) {
                    builder.append(skill.getAttribute());
                }
                builder.append(Numbers.formatWithForcedSign(level));
                return builder.toString();
            }
            return ""; //$NON-NLS-1$
        }
    },
    /** The points spent in the skill. */
    POINTS {
        @Override
        public String toString() {
            return POINTS_TITLE;
        }

        @Override
        public String getToolTip() {
            return POINTS_TOOLTIP;
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
    },
    /** The category. */
    CATEGORY {
        @Override
        public String toString() {
            return CATEGORY_TITLE;
        }

        @Override
        public String getToolTip() {
            return CATEGORY_TOOLTIP;
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.LEFT, true);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
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
            return REFERENCE_TITLE;
        }

        @Override
        public String getToolTip() {
            return REFERENCE_TOOLTIP;
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

    @Localize("Skills")
    @Localize(locale = "de", value = "Fertigkeiten")
    @Localize(locale = "ru", value = "Умения")
    @Localize(locale = "es", value = "Habilidades")
    static String DESCRIPTION_TITLE;
    @Localize("The name, specialty, tech level and notes describing a skill")
    @Localize(locale = "de",
              value = "Der Name, Spezialisierung, Techlevel und Anmerkungen, die die Fertigkeit beschreiben")
    @Localize(locale = "ru", value = "Название, специализация, ТУ и заметки умения")
    @Localize(locale = "es",
              value = "Nombre, especialización, nivel tecnológico y notas que describen una habilidad")
    static String DESCRIPTION_TOOLTIP;
    @Localize("SL")
    @Localize(locale = "de", value = "FW")
    @Localize(locale = "ru", value = "УУ")
    @Localize(locale = "es", value = "NH")
    static String LEVEL_TITLE;
    @Localize("The skill level")
    @Localize(locale = "de", value = "Der Fertigkeitswert")
    @Localize(locale = "ru", value = "Уровень умения")
    @Localize(locale = "es", value = "Nivel de habilidad")
    static String LEVEL_TOOLTIP;
    @Localize("RSL")
    @Localize(locale = "de", value = "RFW")
    @Localize(locale = "ru", value = "ОУУ")
    @Localize(locale = "es", value = "NHR")
    static String RELATIVE_LEVEL_TITLE;
    @Localize("The relative skill level")
    @Localize(locale = "de", value = "Der relative Fertigkeitswert")
    @Localize(locale = "ru", value = "Относительный уровень умения")
    @Localize(locale = "es", value = "Nivel de habilidad relativo")
    static String RELATIVE_LEVEL_TOOLTIP;
    @Localize("Pts")
    @Localize(locale = "de", value = "Pkt")
    @Localize(locale = "ru", value = "Очк")
    @Localize(locale = "es", value = "Ptos")
    static String POINTS_TITLE;
    @Localize("The points spent in the skill")
    @Localize(locale = "de", value = "Die für die Fertigkeit aufgewendeten Punkte")
    @Localize(locale = "ru", value = "Потраченые очки на умение")
    @Localize(locale = "es", value = "Puntos gastados en la habilidad")
    static String POINTS_TOOLTIP;
    @Localize("Diff")
    @Localize(locale = "de", value = "Schwierigkeit")
    @Localize(locale = "ru", value = "Сложн.")
    @Localize(locale = "es", value = "Dificultad")
    static String DIFFICULTY_TITLE;
    @Localize("The skill difficulty")
    @Localize(locale = "de", value = "Die Schwierigkeitsstufe der Fertigkeit")
    @Localize(locale = "ru", value = "Сложность умения")
    @Localize(locale = "es", value = "Dificultad de la habilidad")
    static String DIFFICULTY_TOOLTIP;
    @Localize("Category")
    @Localize(locale = "de", value = "Kategorie")
    @Localize(locale = "ru", value = "Категория")
    @Localize(locale = "es", value = "Categoría")
    static String CATEGORY_TITLE;
    @Localize("The category or categories the skill belongs to")
    @Localize(locale = "de",
              value = "Die Kategorie oder Kategorien, denen diese Fertigkeit angehört")
    @Localize(locale = "ru", value = "Категория или категории, к которым относится умение")
    @Localize(locale = "es", value = "Categoría o categorías a las que pertenece la habilidad")
    static String CATEGORY_TOOLTIP;
    @Localize("Ref")
    @Localize(locale = "de", value = "Ref")
    @Localize(locale = "ru", value = "Ссыл")
    static String REFERENCE_TITLE;
    @Localize("A reference to the book and page this skill appears on (e.g. B22 would refer to \"Basic Set\", page 22)")
    @Localize(locale = "de",
              value = "Eine Referenz auf das Buch und die Seite, auf der dieser Zauber beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
    @Localize(locale = "ru",
              value = "Ссылка на страницу и книгу, описывающая умение (например B22 - книга \"Базовые правила\", страница 22)")
    @Localize(locale = "es",
              value = "Referencia al libro y página en donde se menciona la habilidad (p.e. B22 se refiere al \"Manual Básico\", página 22)")
    static String REFERENCE_TOOLTIP;

    static {
        Localization.initialize();
    }

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

    public boolean showToolTip() {
        return false;
    }

    public String getToolTip(Skill skill) {
        return getToolTip();
    }
}
