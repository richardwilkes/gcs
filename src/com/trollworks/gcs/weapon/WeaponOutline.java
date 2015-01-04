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

package com.trollworks.gcs.weapon;

import com.trollworks.toolkit.ui.widget.outline.Outline;

class WeaponOutline extends Outline {
	WeaponOutline() {
		super(false);
		setAllowColumnDrag(false);
		setAllowColumnResize(false);
		setAllowRowDrag(false);
	}

	@Override
	public boolean canDeleteSelection() {
		return getModel().hasSelection();
	}

	@Override
	public void deleteSelection() {
		if (canDeleteSelection()) {
			getModel().removeSelection();
			sizeColumnsToFit();
		}
	}
}
