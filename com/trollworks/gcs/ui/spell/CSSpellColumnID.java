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

package com.trollworks.gcs.ui.spell;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.ui.common.CSHeaderCell;
import com.trollworks.gcs.ui.common.CSMultiCell;
import com.trollworks.gcs.ui.common.CSTextCell;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKCell;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKTextCell;

/** Definitions for spell columns. */
public enum CSSpellColumnID {
	/** The spell name/description. */
	DESCRIPTION(Msgs.SPELLS, Msgs.SPELLS_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSMultiCell();
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMSpell spell) {
			return getDataAsText(spell);
		}

		@Override public String getDataAsText(CMSpell spell) {
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
	CLASS(Msgs.CLASS, Msgs.CLASS_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSSpellClassCell();
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMSpell spell) {
			return getDataAsText(spell);
		}

		@Override public String getDataAsText(CMSpell spell) {
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
	MANA_COST(Msgs.MANA_COST, Msgs.MANA_COST_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSSpellManaCostCell();
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMSpell spell) {
			return getDataAsText(spell);
		}

		@Override public String getDataAsText(CMSpell spell) {
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
	TIME(Msgs.TIME, Msgs.TIME_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSSpellTimeCell();
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMSpell spell) {
			return getDataAsText(spell);
		}

		@Override public String getDataAsText(CMSpell spell) {
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
	LEVEL(Msgs.LEVEL, Msgs.LEVEL_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMCharacter;
		}

		@Override public Object getData(CMSpell spell) {
			return new Integer(spell.canHaveChildren() ? -1 : spell.getLevel());
		}

		@Override public String getDataAsText(CMSpell spell) {
			int level;

			if (spell.canHaveChildren()) {
				return ""; //$NON-NLS-1$
			}
			level = spell.getLevel();
			if (level < 0) {
				return "-"; //$NON-NLS-1$
			}
			return TKNumberUtils.format(level);
		}
	},
	/** The relative spell level. */
	RELATIVE_LEVEL(Msgs.RELATIVE_LEVEL, Msgs.RELATIVE_LEVEL_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMTemplate || dataFile instanceof CMCharacter;
		}

		@Override public Object getData(CMSpell spell) {
			return new Integer(getRelativeLevel(spell));
		}

		private int getRelativeLevel(CMSpell spell) {
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

		@Override public String getDataAsText(CMSpell spell) {
			if (!spell.canHaveChildren()) {
				int level = getRelativeLevel(spell);

				if (level == Integer.MIN_VALUE) {
					return "-"; //$NON-NLS-1$
				}
				return "IQ" + TKNumberUtils.format(level, true); //$NON-NLS-1$
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The points spent in the spell. */
	POINTS(Msgs.POINTS, Msgs.POINTS_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMTemplate || dataFile instanceof CMCharacter;
		}

		@Override public Object getData(CMSpell spell) {
			return new Integer(spell.canHaveChildren() ? -1 : spell.getPoints());
		}

		@Override public String getDataAsText(CMSpell spell) {
			return spell.canHaveChildren() ? "" : TKNumberUtils.format(spell.getPoints()); //$NON-NLS-1$
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

		@Override public Object getData(CMSpell spell) {
			return getDataAsText(spell);
		}

		@Override public String getDataAsText(CMSpell spell) {
			return spell.getReference();
		}
	};

	private String	mTitle;
	private String	mToolTip;

	private CSSpellColumnID(String title, String tooltip) {
		mTitle = title;
		mToolTip = tooltip;
	}

	@Override public String toString() {
		return mTitle;
	}

	/**
	 * @param spell The {@link CMSpell} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(CMSpell spell);

	/**
	 * @param spell The {@link CMSpell} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(CMSpell spell);

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

		for (CSSpellColumnID one : values()) {
			if (one.shouldDisplay(dataFile)) {
				TKColumn column = new TKColumn(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

				column.setHeaderCell(new CSHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
