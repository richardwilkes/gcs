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

/** A cell for displaying the spell class and college. */
public class SpellClassCell extends MultiCell {
	@Override
	protected String getPrimaryText(ListRow row) {
		return row.canHaveChildren() ? "" : ((Spell) row).getSpellClass(); //$NON-NLS-1$
	}

	@Override
	protected String getSecondaryText(ListRow row) {
		return row.canHaveChildren() ? "" : ((Spell) row).getCollege(); //$NON-NLS-1$
	}

	@Override
	protected String getSortText(ListRow row) {
		return getSecondaryText(row);
	}
}
