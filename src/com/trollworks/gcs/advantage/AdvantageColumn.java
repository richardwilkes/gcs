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
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.outline.Cell;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.ImageCell;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;

import java.awt.Graphics2D;
import java.util.ArrayList;
import java.util.HashMap;

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
			String notes = advantage.getModifierNotes();

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
			return new Integer(advantage.getAdjustedPoints());
		}

		@Override
		public String getDataAsText(Advantage advantage) {
			return Numbers.format(advantage.getAdjustedPoints());
		}
	},
	/** The type. */
	TYPE {
		private HashMap<Integer, StdImage>	mMap;

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
			return new ImageCell(SwingConstants.CENTER, SwingConstants.TOP);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
		}

		@Override
		public Object getData(Advantage advantage) {
			if (!advantage.canHaveChildren()) {
				int type = advantage.getType();
				Integer typeObj;
				StdImage img;

				if (type == 0) {
					return null;
				}
				if (mMap == null) {
					mMap = new HashMap<>();
				}
				typeObj = new Integer(advantage.getType());
				img = mMap.get(typeObj);
				if (img == null) {
					ArrayList<StdImage> list = new ArrayList<>();

					if ((type & Advantage.TYPE_MASK_MENTAL) != 0) {
						list.add(GCSImages.getMentalTypeIcon());
					}
					if ((type & Advantage.TYPE_MASK_PHYSICAL) != 0) {
						list.add(GCSImages.getPhysicalTypeIcon());
					}
					if ((type & Advantage.TYPE_MASK_SOCIAL) != 0) {
						list.add(GCSImages.getSocialTypeIcon());
					}
					if ((type & Advantage.TYPE_MASK_EXOTIC) != 0) {
						list.add(GCSImages.getExoticTypeIcon());
					}
					if ((type & Advantage.TYPE_MASK_SUPERNATURAL) != 0) {
						list.add(GCSImages.getSupernaturalTypeIcon());
					}

					switch (list.size()) {
						case 0:
							break;
						case 1:
							img = list.get(0);
							mMap.put(typeObj, img);
							break;
						default:
							int height = 0;
							int width = 0;
							int x = 0;
							Graphics2D g2d;

							for (StdImage one : list) {
								int tmp;

								width += one.getWidth();
								tmp = one.getHeight();
								if (tmp > height) {
									height = tmp;
								}
							}
							img = StdImage.createTransparent(width, height);
							g2d = img.getGraphics();
							for (StdImage one : list) {
								g2d.drawImage(one, x, (height - one.getHeight()) / 2, null);
								x += one.getWidth();
							}
							g2d.dispose();
							mMap.put(typeObj, img);
							break;
					}
				}
				return img;
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
	static String	DESCRIPTION_TITLE;
	@Localize("The name, level and notes describing an advantage")
	@Localize(locale = "de", value = "Der Name, Stufe und Anmerkungen, die den Vorteil beschreiben")
	@Localize(locale = "ru", value = "Название, уровень и заметки преимущества")
	static String	DESCRIPTION_TOOLTIP;
	@Localize("Pts")
	@Localize(locale = "de", value = "Pkt")
	@Localize(locale = "ru", value = "Очк")
	static String	POINTS_TITLE;
	@Localize("The points spent in the advantage")
	@Localize(locale = "de", value = "Die für den Vorteil aufgewendeten Punkte")
	@Localize(locale = "ru", value = "Потраченые очки на преимущество")
	static String	POINTS_TOOLTIP;
	@Localize("Type")
	@Localize(locale = "de", value = "Typ")
	@Localize(locale = "ru", value = "Тип")
	static String	TYPE_TITLE;
	@Localize("The type of advantage")
	@Localize(locale = "de", value = "Der Typ des Vorteils")
	@Localize(locale = "ru", value = "Тип преимущества")
	static String	TYPE_TOOLTIP;
	@Localize("Category")
	@Localize(locale = "de", value = "Kategorie")
	@Localize(locale = "ru", value = "Категория")
	static String	CATEGORY_TITLE;
	@Localize("The category or categories the advantage belongs to")
	@Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen dieser Vorteil angehört")
	@Localize(locale = "ru", value = "Категория или категории, к которым относится преимущество")
	static String	CATEGORY_TOOLTIP;
	@Localize("Ref")
	@Localize(locale = "de", value = "Ref")
	@Localize(locale = "ru", value = "Ссыл")
	static String	REFERENCE_TITLE;
	@Localize("A reference to the book and page this advantage appears\non (e.g. B22 would refer to \"Basic Set\", page 22)")
	@Localize(locale = "de", value = "Eine Referenz auf das Buch und die Seite, auf der dieser Vorteil beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
	@Localize(locale = "ru", value = "Ссылка на страницу и книгу, описывающая преимущество\n (например, B22 - книга \"Базовые правила\", страница 22)")
	static String	REFERENCE_TOOLTIP;

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
	 * @param outline The {@link Outline} to use.
	 * @param dataFile The {@link DataFile} that data is being displayed for.
	 */
	public static void addColumns(Outline outline, DataFile dataFile) {
		boolean sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
		OutlineModel model = outline.getModel();
		for (AdvantageColumn one : values()) {
			if (one.shouldDisplay(dataFile)) {
				Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());
				column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
