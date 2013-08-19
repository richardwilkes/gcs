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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.advantage;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.advantage.CMAdvantageContainerType;
import com.trollworks.gcs.ui.common.CSHeaderCell;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.common.CSMultiCell;
import com.trollworks.gcs.ui.common.CSTextCell;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.outline.TKImageCell;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKCell;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKTextCell;

import java.awt.Graphics2D;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.HashMap;

/** Definitions for advantage columns. */
public enum CSAdvantageColumnID {
	/** The advantage name/description. */
	DESCRIPTION(Msgs.DESCRIPTION, Msgs.DESCRIPTION_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSMultiCell();
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMAdvantage advantage) {
			return getDataAsText(advantage);
		}

		@Override public String getDataAsText(CMAdvantage advantage) {
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
	POINTS(Msgs.POINTS, Msgs.POINTS_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMAdvantage advantage) {
			if (advantage.canHaveChildren()) {
				if (advantage.getContainerType() == CMAdvantageContainerType.GROUP) {
					return new Integer(-1);
				}
			}
			return new Integer(advantage.getAdjustedPoints());
		}

		@Override public String getDataAsText(CMAdvantage advantage) {
			if (advantage.canHaveChildren()) {
				if (advantage.getContainerType() == CMAdvantageContainerType.GROUP) {
					return ""; //$NON-NLS-1$
				}
			}
			return TKNumberUtils.format(advantage.getAdjustedPoints());
		}
	},
	/** The type. */
	TYPE(Msgs.TYPE, Msgs.TYPE_TOOLTIP) {
		private HashMap<Integer, BufferedImage>	mMap;

		@Override public TKCell getCell() {
			return new TKImageCell(true, TKAlignment.CENTER, TKAlignment.TOP);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMListFile;
		}

		@Override public Object getData(CMAdvantage advantage) {
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

					if ((type & CMAdvantage.TYPE_MASK_MENTAL) != 0) {
						list.add(CSImage.getMentalTypeIcon());
					}
					if ((type & CMAdvantage.TYPE_MASK_PHYSICAL) != 0) {
						list.add(CSImage.getPhysicalTypeIcon());
					}
					if ((type & CMAdvantage.TYPE_MASK_SOCIAL) != 0) {
						list.add(CSImage.getSocialTypeIcon());
					}
					if ((type & CMAdvantage.TYPE_MASK_EXOTIC) != 0) {
						list.add(CSImage.getExoticTypeIcon());
					}
					if ((type & CMAdvantage.TYPE_MASK_SUPERNATURAL) != 0) {
						list.add(CSImage.getSupernaturalTypeIcon());
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
							img = TKImage.create(width, height);
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

		@Override public String getDataAsText(CMAdvantage advantage) {
			return advantage.getTypeAsText();
		}
	},
	/** The page reference. */
	REFERENCE(Msgs.REFERENCE, Msgs.REFERENCE_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMAdvantage advantage) {
			return getDataAsText(advantage);
		}

		@Override public String getDataAsText(CMAdvantage advantage) {
			return advantage.getReference();
		}
	};

	private String	mTitle;
	private String	mToolTip;

	private CSAdvantageColumnID(String title, String tooltip) {
		mTitle = title;
		mToolTip = tooltip;
	}

	@Override public String toString() {
		return mTitle;
	}

	/**
	 * @param advantage The {@link CMAdvantage} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(CMAdvantage advantage);

	/**
	 * @param advantage The {@link CMAdvantage} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(CMAdvantage advantage);

	/** @return The tooltip for the column. */
	public String getToolTip() {
		return mToolTip;
	}

	/** @return The {@link TKCell} used to display the data. */
	public abstract TKCell getCell();

	/**
	 * @param dataFile The {@link CMDataFile} to use.
	 * @return Whether this column should be displayed for the specified data file.
	 */
	public abstract boolean shouldDisplay(CMDataFile dataFile);

	/**
	 * Adds all relevant {@link TKColumn}s to a {@link TKOutline}.
	 * 
	 * @param outline The {@link TKOutline} to use.
	 * @param dataFile The {@link CMDataFile} that data is being displayed for.
	 */
	public static void addColumns(TKOutline outline, CMDataFile dataFile) {
		boolean sheetOrTemplate = dataFile instanceof CMCharacter || dataFile instanceof CMTemplate;
		TKOutlineModel model = outline.getModel();

		for (CSAdvantageColumnID one : values()) {
			if (one.shouldDisplay(dataFile)) {
				TKColumn column = new TKColumn(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

				column.setHeaderCell(new CSHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
