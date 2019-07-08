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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.RetinaIcon;
import com.trollworks.toolkit.ui.widget.outline.Cell;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.IconsCell;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;

import java.util.ArrayList;
import java.util.List;

import javax.swing.SwingConstants;

/** Definitions for advantage columns. */
public enum AdvantageColumn {
    /** The advantage name/description. */
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
        public Object getData(Advantage advantage) {
            return getDataAsText(advantage);
        }

        @Override
        public String getDataAsText(Advantage advantage) {
            StringBuilder builder = new StringBuilder();
            String        notes   = advantage.getModifierNotes();

            builder.append(advantage.toString());
            if (notes.length() > 0) {
                builder.append(" - "); //$NON-NLS-1$
                builder.append(notes);
            }
            notes = advantage.getNotes();
            if (notes.length() > 0) {
                builder.append(" - "); //$NON-NLS-1$
                builder.append(notes);
            }
            return builder.toString();
        }
    },
    /** The points spent in the advantage. */
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
            return TYPE_TITLE;
        }

        @Override
        public String getToolTip() {
            return TYPE_TOOLTIP;
        }

        @Override
        public Cell getCell() {
            return new IconsCell(SwingConstants.CENTER, SwingConstants.TOP);
        }

        @Override
        public boolean shouldDisplay(DataFile dataFile) {
            return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
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
                    imgs.add(GCSImages.getMentalTypeIcon());
                }
                if ((type & Advantage.TYPE_MASK_PHYSICAL) != 0) {
                    imgs.add(GCSImages.getPhysicalTypeIcon());
                }
                if ((type & Advantage.TYPE_MASK_SOCIAL) != 0) {
                    imgs.add(GCSImages.getSocialTypeIcon());
                }
                if ((type & Advantage.TYPE_MASK_EXOTIC) != 0) {
                    imgs.add(GCSImages.getExoticTypeIcon());
                }
                if ((type & Advantage.TYPE_MASK_SUPERNATURAL) != 0) {
                    imgs.add(GCSImages.getSupernaturalTypeIcon());
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
        public Object getData(Advantage advantage) {
            return getDataAsText(advantage);
        }

        @Override
        public String getDataAsText(Advantage advantage) {
            return advantage.getReference();
        }
    };

    @Localize("Advantages & Disadvantages")
    @Localize(locale = "de", value = "Vorteile & Nachteile")
    @Localize(locale = "ru", value = "Преимущества и недостатки")
    @Localize(locale = "es", value = "Ventajas y Desventajas")
    static String DESCRIPTION_TITLE;
    @Localize("The name, level and notes describing an advantage")
    @Localize(locale = "de", value = "Der Name, Stufe und Anmerkungen, die den Vorteil beschreiben")
    @Localize(locale = "ru", value = "Название, уровень и заметки преимущества")
    @Localize(locale = "es", value = "Nombre, nivel y notas describiendo la ventaja")
    static String DESCRIPTION_TOOLTIP;
    @Localize("Pts")
    @Localize(locale = "de", value = "Pkt")
    @Localize(locale = "ru", value = "Очк")
    @Localize(locale = "es", value = "Ptos")
    static String POINTS_TITLE;
    @Localize("The points spent in the advantage")
    @Localize(locale = "de", value = "Die für den Vorteil aufgewendeten Punkte")
    @Localize(locale = "ru", value = "Потраченые очки на преимущество")
    @Localize(locale = "es", value = "Coste en puntos de la ventaja")
    static String POINTS_TOOLTIP;
    @Localize("Type")
    @Localize(locale = "de", value = "Typ")
    @Localize(locale = "ru", value = "Тип")
    @Localize(locale = "es", value = "Tipo")
    static String TYPE_TITLE;
    @Localize("The type of advantage")
    @Localize(locale = "de", value = "Der Typ des Vorteils")
    @Localize(locale = "ru", value = "Тип преимущества")
    @Localize(locale = "es", value = "Tipo de ventaja")
    static String TYPE_TOOLTIP;
    @Localize("Category")
    @Localize(locale = "de", value = "Kategorie")
    @Localize(locale = "ru", value = "Категория")
    @Localize(locale = "es", value = "Categoría")
    static String CATEGORY_TITLE;
    @Localize("The category or categories the advantage belongs to")
    @Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen dieser Vorteil angehört")
    @Localize(locale = "ru", value = "Категория или категории, к которым относится преимущество")
    @Localize(locale = "es", value = "Categoría o categorías a las que pertenece la ventaja")
    static String CATEGORY_TOOLTIP;
    @Localize("Ref")
    @Localize(locale = "de", value = "Ref")
    @Localize(locale = "ru", value = "Ссыл")
    @Localize(locale = "es", value = "Ref")
    static String REFERENCE_TITLE;
    @Localize("A reference to the book and page this advantage appears on (e.g. B22 would refer to \"Basic Set\", page 22)")
    @Localize(locale = "de",
              value = "Eine Referenz auf das Buch und die Seite, auf der dieser Vorteil beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
    @Localize(locale = "ru",
              value = "Ссылка на страницу и книгу, описывающая преимущество (например, B22 - книга \"Базовые правила\", страница 22)")
    @Localize(locale = "es",
              value = "Referencia al libro y página en donde aparece la ventaja (p.e. B22 se refiere al \"Manual Básico\", página 22)")
    static String REFERENCE_TOOLTIP;

    static {
        Localization.initialize();
    }

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

    public boolean showToolTip() {
        return false;
    }

    public String getToolTip(Advantage advantage) {
        return getToolTip();
    }

}
