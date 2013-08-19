/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

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
import com.trollworks.ttk.image.Images;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.Cell;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.ImageCell;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;

import java.awt.Graphics2D;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.HashMap;

import javax.swing.SwingConstants;

/** Definitions for advantage columns. */
public enum AdvantageColumn {
	/** The advantage name/description. */
	DESCRIPTION {
		@Override
		public String toString() {
			return MSG_DESCRIPTION;
		}

		@Override
		public String getToolTip() {
			return MSG_DESCRIPTION_TOOLTIP;
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
			return MSG_POINTS;
		}

		@Override
		public String getToolTip() {
			return MSG_POINTS_TOOLTIP;
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
		private HashMap<Integer, BufferedImage>	mMap;

		@Override
		public String toString() {
			return MSG_TYPE;
		}

		@Override
		public String getToolTip() {
			return MSG_TYPE_TOOLTIP;
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
				BufferedImage img;

				if (type == 0) {
					return null;
				}
				if (mMap == null) {
					mMap = new HashMap<Integer, BufferedImage>();
				}
				typeObj = new Integer(advantage.getType());
				img = mMap.get(typeObj);
				if (img == null) {
					ArrayList<BufferedImage> list = new ArrayList<BufferedImage>();

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

							for (BufferedImage one : list) {
								int tmp;

								width += one.getWidth();
								tmp = one.getHeight();
								if (tmp > height) {
									height = tmp;
								}
							}
							img = Images.create(width, height);
							g2d = (Graphics2D) img.getGraphics();
							g2d.setClip(0, 0, width, height);
							for (BufferedImage one : list) {
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
			return MSG_CATEGORY;
		}

		@Override
		public String getToolTip() {
			return MSG_CATEGORY_TOOLTIP;
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
			return MSG_REFERENCE;
		}

		@Override
		public String getToolTip() {
			return MSG_REFERENCE_TOOLTIP;
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

	static String	MSG_DESCRIPTION;
	static String	MSG_DESCRIPTION_TOOLTIP;
	static String	MSG_POINTS;
	static String	MSG_POINTS_TOOLTIP;
	static String	MSG_TYPE;
	static String	MSG_TYPE_TOOLTIP;
	static String	MSG_CATEGORY;
	static String	MSG_CATEGORY_TOOLTIP;
	static String	MSG_REFERENCE;
	static String	MSG_REFERENCE_TOOLTIP;

	static {
		LocalizedMessages.initialize(AdvantageColumn.class);
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
