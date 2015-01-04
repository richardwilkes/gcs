/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.spell;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.MultiCell;

/** A cell for displaying the mana cost for casting and maintaining a spell. */
public class SpellManaCostCell extends MultiCell {
	@Override
	protected String getPrimaryText(ListRow row) {
		return row.canHaveChildren() ? "" : ((Spell) row).getCastingCost(); //$NON-NLS-1$
	}

	@Override
	protected String getSecondaryText(ListRow row) {
		return row.canHaveChildren() ? "" : ((Spell) row).getMaintenance(); //$NON-NLS-1$
	}
}
