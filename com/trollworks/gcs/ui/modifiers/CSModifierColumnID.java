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

package com.trollworks.gcs.ui.modifiers;

import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.ui.common.CSHeaderCell;
import com.trollworks.gcs.ui.common.CSMultiCell;
import com.trollworks.gcs.ui.common.CSTextCell;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.outline.TKCell;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKTextCell;

/** Modifier Columns */
public enum CSModifierColumnID {
	/** The advantage name/description. */
	DESCRIPTION(Msgs.DESCRIPTION, Msgs.DESCRIPTION_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSMultiCell();
		}

		@Override public Object getData(CMModifier modifier) {
			return getDataAsText(modifier);
		}

		@Override public String getDataAsText(CMModifier modifier) {
			StringBuilder builder = new StringBuilder();
			String notes = modifier.getNotes();

			builder.append(modifier.toString());
			if (notes.length() > 0) {
				builder.append(" - "); //$NON-NLS-1$
				builder.append(notes);
			}
			return builder.toString();
		}
	},
	/** The total cost modifier. */
	COST_MODIFIER_TOTAL(Msgs.COST_MODIFIER, Msgs.COST_MODIFIER_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public Object getData(CMModifier modifier) {
			return new Integer(modifier.getCostModifier());
		}

		@Override public String getDataAsText(CMModifier modifier) {
			return TKNumberUtils.format(modifier.getCostModifier(), true) + "%"; //$NON-NLS-1$
		}
	},

	/** The page reference. */
	REFERENCE(Msgs.REFERENCE, Msgs.REFERENCE_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public Object getData(CMModifier modifier) {
			return getDataAsText(modifier);
		}

		@Override public String getDataAsText(CMModifier modifier) {
			return modifier.getReference();
		}
	};

	private String	mTitle;
	private String	mToolTip;

	private CSModifierColumnID(String title, String tooltip) {
		mTitle = title;
		mToolTip = tooltip;
	}

	@Override public String toString() {
		return mTitle;
	}

	/**
	 * @param modifier The {@link CMModifier} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(CMModifier modifier);

	/**
	 * @param modifier The {@link CMModifier} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(CMModifier modifier);

	/** @return The tooltip for the column. */
	public String getToolTip() {
		return mToolTip;
	}

	/** @return The {@link TKCell} used to display the data. */
	public abstract TKCell getCell();

	/** @return Whether this column should be displayed for the specified data file. */
	public boolean shouldDisplay() {
		return true;
	}

	/**
	 * Adds all relevant {@link TKColumn}s to a {@link TKOutline}.
	 * 
	 * @param outline The {@link TKOutline} to use.
	 */
	public static void addColumns(TKOutline outline) {
		TKOutlineModel model = outline.getModel();

		for (CSModifierColumnID one : values()) {
			if (one.shouldDisplay()) {
				TKColumn column = new TKColumn(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

				column.setHeaderCell(new CSHeaderCell(false));
				model.addColumn(column);
			}
		}
	}

}
