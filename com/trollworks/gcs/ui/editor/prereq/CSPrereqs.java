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

package com.trollworks.gcs.ui.editor.prereq;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.prereq.CMPrereq;
import com.trollworks.gcs.model.prereq.CMPrereqList;
import com.trollworks.gcs.ui.editor.CSBandedPanel;

/** Displays and edits {@link CMPrereq} objects. */
public class CSPrereqs extends CSBandedPanel {
	/**
	 * Creates a new prerequisite editor.
	 * 
	 * @param row The row these prerequisites will belong to.
	 * @param prereqs The initial prerequisites to display.
	 */
	public CSPrereqs(CMRow row, CMPrereqList prereqs) {
		super(Msgs.PREREQUISITES);
		addPrereqs(row, new CMPrereqList(null, prereqs), 0);
	}

	/** @return The current prerequisite list. */
	public CMPrereqList getPrereqList() {
		return (CMPrereqList) ((CSListPrereq) getComponent(0)).getPrereq();
	}

	private void addPrereqs(CMRow row, CMPrereqList prereqs, int depth) {
		add(CSBasePrereq.create(row, prereqs, depth++));
		for (CMPrereq prereq : prereqs.getChildren()) {
			if (prereq instanceof CMPrereqList) {
				addPrereqs(row, (CMPrereqList) prereq, depth);
			} else {
				add(CSBasePrereq.create(row, prereq, depth));
			}
		}
	}
}
