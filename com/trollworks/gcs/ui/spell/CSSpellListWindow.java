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

import com.trollworks.gcs.model.spell.CMSpellList;
import com.trollworks.gcs.ui.common.CSListOpener;
import com.trollworks.gcs.ui.common.CSListWindow;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.toolkit.io.TKFileFilter;

/** The spells list window. */
public class CSSpellListWindow extends CSListWindow {
	/**
	 * Creates a list window.
	 * 
	 * @param list The list to display.
	 */
	public CSSpellListWindow(CMSpellList list) {
		super(list, CMD_NEW_SPELL, CMD_NEW_SPELL_CONTAINER);
	}

	@Override protected CSOutline createOutline() {
		return new CSSpellOutline(mListFile);
	}

	@Override protected String getUntitledName() {
		return Msgs.UNTITLED;
	}

	@Override public TKFileFilter[] getFileFilters() {
		return new TKFileFilter[] { CSListOpener.FILTERS[CSListOpener.SPELL_FILTER] };
	}
}
