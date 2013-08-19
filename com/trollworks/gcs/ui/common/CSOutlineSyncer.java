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

import com.trollworks.toolkit.widget.outline.TKOutline;

import java.awt.EventQueue;
import java.util.HashSet;

/** Provides synchronization of an outline to its data. */
public class CSOutlineSyncer implements Runnable {
	private static final CSOutlineSyncer	INSTANCE	= new CSOutlineSyncer();
	private static final HashSet<TKOutline>	OUTLINES	= new HashSet<TKOutline>();
	private static boolean					PENDING		= false;

	/**
	 * @param outline The {@link CSOutline} to add to the set of outlines that need to be
	 *            syncronized.
	 */
	public static final void add(TKOutline outline) {
		synchronized (OUTLINES) {
			OUTLINES.add(outline);
			if (!PENDING) {
				PENDING = true;
				EventQueue.invokeLater(INSTANCE);
			}
		}
	}

	/**
	 * @param outline The {@link CSOutline} to remove from the set of outlines that need to be
	 *            syncronized.
	 */
	public static final void remove(TKOutline outline) {
		synchronized (OUTLINES) {
			OUTLINES.remove(outline);
		}
	}

	public void run() {
		HashSet<TKOutline> outlines;

		synchronized (OUTLINES) {
			PENDING = false;
			outlines = new HashSet<TKOutline>(OUTLINES);
			OUTLINES.clear();
		}

		for (TKOutline outline : outlines) {
			outline.sizeColumnsToFit();
			outline.repaint();
		}
	}
}
