/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.spell;

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

/** Definitions for spell columns. */
public enum SpellColumn {
	/** The spell name/description. */
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
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			StringBuilder builder = new StringBuilder();
			String notes = spell.getNotes();

			builder.append(spell.toString());
			if (notes.length() > 0) {
				builder.append(" - "); //$NON-NLS-1$
				builder.append(notes);
			}
			return builder.toString();
		}
	},
	/** The spell class/college. */
	CLASS {
		@Override
		public String toString() {
			return CLASS_TITLE;
		}

		@Override
		public String getToolTip() {
			return CLASS_TOOLTIP;
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
				StringBuilder builder = new StringBuilder();

				builder.append(spell.getSpellClass());
				builder.append("; "); //$NON-NLS-1$
				builder.append(spell.getCollege());
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The casting &amp; maintenance cost. */
	MANA_COST {
		@Override
		public String toString() {
			return MANA_COST_TITLE;
		}

		@Override
		public String getToolTip() {
			return MANA_COST_TOOLTIP;
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
				StringBuilder builder = new StringBuilder();

				builder.append(spell.getCastingCost());
				builder.append("; "); //$NON-NLS-1$
				builder.append(spell.getMaintenance());
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The casting time &amp; duration. */
	TIME {
		@Override
		public String toString() {
			return TIME_TITLE;
		}

		@Override
		public String getToolTip() {
			return TIME_TOOLTIP;
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
				StringBuilder builder = new StringBuilder();

				builder.append(spell.getCastingTime());
				builder.append("; "); //$NON-NLS-1$
				builder.append(spell.getDuration());
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The spell level. */
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
		public Object getData(Spell spell) {
			return new Integer(spell.canHaveChildren() ? -1 : spell.getLevel());
		}

		@Override
		public String getDataAsText(Spell spell) {
			int level;

			if (spell.canHaveChildren()) {
				return ""; //$NON-NLS-1$
			}
			level = spell.getLevel();
			if (level < 0) {
				return "-"; //$NON-NLS-1$
			}
			return Numbers.format(level);
		}
	},
	/** The relative spell level. */
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
		public Object getData(Spell spell) {
			return new Integer(getRelativeLevel(spell));
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
					return "-"; //$NON-NLS-1$
				}
				return "IQ" + Numbers.formatWithForcedSign(level); //$NON-NLS-1$
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The points spent in the spell. */
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
			return dataFile instanceof Template || dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Spell spell) {
			return new Integer(spell.canHaveChildren() ? -1 : spell.getPoints());
		}

		@Override
		public String getDataAsText(Spell spell) {
			return spell.canHaveChildren() ? "" : Numbers.format(spell.getPoints()); //$NON-NLS-1$
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
		public Object getData(Spell spell) {
			return getDataAsText(spell);
		}

		@Override
		public String getDataAsText(Spell spell) {
			return spell.getReference();
		}
	};

	@Localize("Spells")
	@Localize(locale = "de", value = "Zauber")
	@Localize(locale = "ru", value = "Заклинания")
	static String	DESCRIPTION_TITLE;
	@Localize("The name, tech level and notes describing the spell")
	@Localize(locale = "de", value = "Der Name, Techlevel und Anmerkungen, die den Zauber beschreiben")
	@Localize(locale = "ru", value = "Название, ТУ и заметки заклинания")
	static String	DESCRIPTION_TOOLTIP;
	@Localize("Class")
	@Localize(locale = "de", value = "Klasse")
	@Localize(locale = "ru", value = "Класс")
	static String	CLASS_TITLE;
	@Localize("The class and college of the spell")
	@Localize(locale = "de", value = "Die Klasse und Schule des Zaubers")
	@Localize(locale = "ru", value = "Класс и школа заклинания")
	static String	CLASS_TOOLTIP;
	@Localize("Mana Cost")
	@Localize(locale = "de", value = "Mana-Kosten")
	@Localize(locale = "ru", value = "Мана")
	static String	MANA_COST_TITLE;
	@Localize("The mana cost to cast and maintain the spell")
	@Localize(locale = "de", value = "Die Mana-Kosten, um den Zauber zu wirken")
	@Localize(locale = "ru", value = "Мана-стоимость сотворения заклинания и его поддержание")
	static String	MANA_COST_TOOLTIP;
	@Localize("Time")
	@Localize(locale = "de", value = "Zeit")
	@Localize(locale = "ru", value = "Время")
	static String	TIME_TITLE;
	@Localize("The time required to cast the spell and its duration")
	@Localize(locale = "de", value = "Die benötigte Zeit, um den Zauber zu wirken und seine Dauer")
	@Localize(locale = "ru", value = "Необходимое время для сотворения заклинания и его длительность")
	static String	TIME_TOOLTIP;
	@Localize("Pts")
	@Localize(locale = "de", value = "Pkt")
	@Localize(locale = "ru", value = "Очк")
	static String	POINTS_TITLE;
	@Localize("The points spent in the spell")
	@Localize(locale = "de", value = "Die für den Zauber aufgewendeten Punkte")
	@Localize(locale = "ru", value = "Потраченные очки на заклинание")
	static String	POINTS_TOOLTIP;
	@Localize("SL")
	@Localize(locale = "de", value = "FW")
	@Localize(locale = "ru", value = "УУ")
	static String	LEVEL_TITLE;
	@Localize("The spell level")
	@Localize(locale = "de", value = "Der Fertigkeitswert des Zaubers")
	@Localize(locale = "ru", value = "Уровень заклинания")
	static String	LEVEL_TOOLTIP;
	@Localize("RSL")
	@Localize(locale = "de", value = "RFW")
	@Localize(locale = "ru", value = "ОУУ")
	static String	RELATIVE_LEVEL_TITLE;
	@Localize("The relative spell level")
	@Localize(locale = "de", value = "Der relative Fertigkeitswert des Zaubers")
	@Localize(locale = "ru", value = "Относительный уровень заклинания")
	static String	RELATIVE_LEVEL_TOOLTIP;
	@Localize("Category")
	@Localize(locale = "de", value = "Kategorie")
	@Localize(locale = "ru", value = "Категория")
	static String	CATEGORY_TITLE;
	@Localize("The category or categories the spell belongs to")
	@Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen dieser Zauber angehört")
	@Localize(locale = "ru", value = "Категория или категории, к которым относится заклинание")
	static String	CATEGORY_TOOLTIP;
	@Localize("Ref")
	@Localize(locale = "de", value = "Ref")
	@Localize(locale = "ru", value = "Ссыл")
	static String	REFERENCE_TITLE;
	@Localize("A reference to the book and page this spell appears\non (e.g. B22 would refer to \"Basic Set\", page 22)")
	@Localize(locale = "de", value = "Eine Referenz auf das Buch und die Seite, auf der dieser Zauber beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
	@Localize(locale = "ru", value = "Ссылка на страницу и книгу, описывающая заклинание\n (например B22 - книга \"Базовые правила\", страница 22)")
	static String	REFERENCE_TOOLTIP;

	static {
		Localization.initialize();
	}

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
	 * @param outline The {@link Outline} to use.
	 * @param dataFile The {@link DataFile} that data is being displayed for.
	 */
	public static void addColumns(Outline outline, DataFile dataFile) {
		boolean sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
		OutlineModel model = outline.getModel();

		for (SpellColumn one : values()) {
			if (one.shouldDisplay(dataFile)) {
				Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

				column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
