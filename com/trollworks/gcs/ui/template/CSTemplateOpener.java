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

package com.trollworks.gcs.ui.template;

import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.window.TKFileDialog;
import com.trollworks.toolkit.window.TKWindow;
import com.trollworks.toolkit.window.TKWindowOpener;

import java.io.File;
import java.util.List;

/** Handles opening of template files. */
public class CSTemplateOpener implements TKWindowOpener {
	/** The extension for templates. */
	public static final String				EXTENSION	= ".gct";																//$NON-NLS-1$
	/** The file filters for character sheets. */
	public static TKFileFilter[]			FILTERS		= new TKFileFilter[] { new TKFileFilter(Msgs.TEMPLATES, EXTENSION) };
	private static final CSTemplateOpener	INSTANCE	= new CSTemplateOpener();

	/** @return The one and only instance of the template opener. */
	public static final CSTemplateOpener getInstance() {
		return INSTANCE;
	}

	private CSTemplateOpener() {
		TKFileDialog.setIconForFileExtension(EXTENSION, CSImage.getTemplateIcon(false));
	}

	/**
	 * @param filters The filters to choose from.
	 * @return The preferred file filter to use from the array of filters.
	 */
	public static TKFileFilter getPreferredFileFilter(TKFileFilter[] filters) {
		if (filters != null) {
			for (TKFileFilter element : filters) {
				if (element == FILTERS[0]) {
					return element;
				}
			}
		}
		return null;
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

			if (finalChance || EXTENSION.equals(TKPath.getExtension(file.getName()))) {
				CSTemplateWindow window = CSTemplateWindow.findTemplateWindow(file);

				if (window == null) {
					try {
						window = new CSTemplateWindow(new CMTemplate(file));
					} catch (Exception exception) {
						if (TKDebug.isKeySet(TKDebug.KEY_DIAGNOSE_LOAD_SAVE)) {
							exception.printStackTrace(System.err);
						}
						return null;
					}
				}

				if (show) {
					window.setVisible(true);
				}
				return window;
			}
		}

		return null;
	}
}
