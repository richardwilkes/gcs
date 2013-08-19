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

import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.advantage.CMAdvantageList;
import com.trollworks.gcs.ui.common.CSListOpener;
import com.trollworks.gcs.ui.common.CSListWindow;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.notification.TKBatchNotifierTarget;

/** The (dis)advantages list window. */
public class CSAdvantageListWindow extends CSListWindow implements TKBatchNotifierTarget {
	/**
	 * Creates a list window.
	 * 
	 * @param list The list to display.
	 */
	public CSAdvantageListWindow(CMAdvantageList list) {
		super(list, CMD_NEW_ADVANTAGE, CMD_NEW_ADVANTAGE_CONTAINER);
		list.addTarget(this, CMAdvantage.ID_TYPE);
	}

	@Override protected CSOutline createOutline() {
		return new CSAdvantageOutline(mListFile);
	}

	@Override protected String getUntitledName() {
		return Msgs.UNTITLED;
	}

	@Override public TKFileFilter[] getFileFilters() {
		return new TKFileFilter[] { CSListOpener.FILTERS[CSListOpener.ADVANTAGE_FILTER] };
	}

	public void enterBatchMode() {
		// Not needed.
	}

	public void leaveBatchMode() {
		// Not needed.
	}

	public void handleNotification(Object producer, String type, Object data) {
		repaint();
	}
}
