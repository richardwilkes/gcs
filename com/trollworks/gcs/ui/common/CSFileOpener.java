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

import com.trollworks.gcs.ui.sheet.CSSheetOpener;
import com.trollworks.gcs.ui.template.CSTemplateOpener;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.window.TKOpenManager;
import com.trollworks.toolkit.window.TKWindow;
import com.trollworks.toolkit.window.TKWindowOpener;

import java.io.File;
import java.util.ArrayList;
import java.util.List;

/** Handles opening of character sheet files. */
public class CSFileOpener implements TKWindowOpener {
	private static final CSFileOpener	INSTANCE;
	/** The file filters for character sheets. */
	public static TKFileFilter[]		FILTERS;

	static {
		// Combine all of the various file filters into one
		ArrayList<TKFileFilter> list = new ArrayList<TKFileFilter>();
		TKFileFilter theFilter;

		for (TKFileFilter filter : CSSheetOpener.FILTERS) {
			list.add(filter);
		}
		for (TKFileFilter filter : CSTemplateOpener.FILTERS) {
			list.add(filter);
		}
		for (TKFileFilter filter : CSListOpener.FILTERS) {
			list.add(filter);
		}
		if (list.size() == 1) {
			theFilter = list.get(0);
		} else {
			theFilter = null;
			for (TKFileFilter filter : list) {
				if (theFilter != null) {
					theFilter = new TKFileFilter(theFilter, filter, Msgs.FILES_DESCRIPTION);
				} else {
					theFilter = filter;
				}
			}
		}
		FILTERS = new TKFileFilter[] { theFilter };
		INSTANCE = new CSFileOpener();
	}

	/** @return The one and only instance of the file opener. */
	public static final CSFileOpener getInstance() {
		return INSTANCE;
	}

	/**
	 * @param filters The filters to choose from.
	 * @return The preferred file filter to use from the array of filters.
	 */
	public static TKFileFilter getPreferredFileFilter(TKFileFilter[] filters) {
		if (filters != null && filters.length > 0) {
			for (TKFileFilter filter : filters) {
				if (filter == FILTERS[0]) {
					return filter;
				}
			}
			return filters[0];
		}
		return null;
	}

	private CSFileOpener() {
		TKOpenManager.setDefaultFileFilter(FILTERS[0]);
		TKOpenManager.setDefaultOpener(this);
	}

	public TKFileFilter[] getFileFilters() {
		return FILTERS;
	}

	public TKWindow openWindow(Object obj, boolean show, boolean finalChance, List<String> msgs) {
		if (obj instanceof String) {
			obj = new File((String) obj);
		}

		if (obj instanceof File) {
			File file = (File) obj;

			for (TKFileFilter filter : CSSheetOpener.FILTERS) {
				if (filter.accept(file)) {
					return CSSheetOpener.getInstance().openWindow(file, show, finalChance, msgs);
				}
			}
			for (TKFileFilter filter : CSTemplateOpener.FILTERS) {
				if (filter.accept(file)) {
					return CSTemplateOpener.getInstance().openWindow(file, show, finalChance, msgs);
				}
			}
			for (TKFileFilter filter : CSListOpener.FILTERS) {
				if (filter.accept(file)) {
					return CSListOpener.getInstance().openWindow(file, show, finalChance, msgs);
				}
			}
		}
		return null;
	}
}
