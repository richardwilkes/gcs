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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.character.names.Namer;
import com.trollworks.gcs.modifier.ModifierEnabler;
import com.trollworks.ttk.collections.FilteredList;
import com.trollworks.ttk.widgets.outline.Outline;

import java.awt.Container;
import java.util.ArrayList;
import java.util.HashMap;

/** Helper for causing the row post-processing to occur. */
public class RowPostProcessor implements Runnable {
	private HashMap<Outline, ArrayList<ListRow>>	mMap;

	/**
	 * Creates a new post processor for name substitution.
	 * 
	 * @param map The map to process.
	 */
	public RowPostProcessor(HashMap<Outline, ArrayList<ListRow>> map) {
		mMap = map;
	}

	/**
	 * Creates a new post processor for name substitution.
	 * 
	 * @param outline The outline containing the rows.
	 * @param list The list to process.
	 */
	public RowPostProcessor(Outline outline, ArrayList<ListRow> list) {
		mMap = new HashMap<Outline, ArrayList<ListRow>>();
		mMap.put(outline, list);
	}

	public void run() {
		for (Outline outline : mMap.keySet()) {
			ArrayList<ListRow> rows = mMap.get(outline);
			boolean modified = ModifierEnabler.process(outline, new FilteredList<Advantage>(rows, Advantage.class));
			modified |= Namer.name(outline, rows);
			if (modified) {
				outline.updateRowHeights(rows);
				outline.repaint();
				Container window = outline.getTopLevelAncestor();
				if (window instanceof SheetWindow) {
					SheetWindow sheetWindow = (SheetWindow) window;
					sheetWindow.notifyOfPrereqOrFeatureModification();
				}
			}
		}
	}
}
