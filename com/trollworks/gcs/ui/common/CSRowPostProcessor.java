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

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.ui.sheet.CSSheetWindow;
import com.trollworks.toolkit.collections.TKFilteredList;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.window.TKBaseWindow;

import java.util.ArrayList;
import java.util.HashMap;

/** Helper for causing the row post-processing to occur. */
public class CSRowPostProcessor implements Runnable {
	private HashMap<TKOutline, ArrayList<CMRow>>	mMap;

	/**
	 * Creates a new post processor for name substitution.
	 * 
	 * @param map The map to process.
	 */
	public CSRowPostProcessor(HashMap<TKOutline, ArrayList<CMRow>> map) {
		mMap = map;
	}

	/**
	 * Creates a new post processor for name substitution.
	 * 
	 * @param outline The outline containing the rows.
	 * @param list The list to process.
	 */
	public CSRowPostProcessor(TKOutline outline, ArrayList<CMRow> list) {
		mMap = new HashMap<TKOutline, ArrayList<CMRow>>();
		mMap.put(outline, list);
	}

	public void run() {
		for (TKOutline outline : mMap.keySet()) {
			TKBaseWindow window = outline.getBaseWindow();
			ArrayList<CMRow> rows = mMap.get(outline);
			boolean modified = CSModifierEnabler.process(window, new TKFilteredList<CMAdvantage>(rows, CMAdvantage.class));

			modified |= CSNamer.name(window, rows);
			if (modified) {
				outline.updateRowHeights(rows);
				outline.repaint();

				if (window instanceof CSSheetWindow) {
					CSSheetWindow sheetWindow = (CSSheetWindow) window;

					sheetWindow.notifyOfPrereqOrFeatureModification();
				}
			}
		}
	}
}
