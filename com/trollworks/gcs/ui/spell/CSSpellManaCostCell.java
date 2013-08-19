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

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.ui.common.CSMultiCell;

/** A cell for displaying the mana cost for casting and maintaining a spell. */
public class CSSpellManaCostCell extends CSMultiCell {
	@Override protected String getPrimaryText(CMRow row) {
		return row.canHaveChildren() ? "" : ((CMSpell) row).getCastingCost(); //$NON-NLS-1$
	}

	@Override protected String getSecondaryText(CMRow row) {
		return row.canHaveChildren() ? "" : ((CMSpell) row).getMaintenance(); //$NON-NLS-1$
	}
}
