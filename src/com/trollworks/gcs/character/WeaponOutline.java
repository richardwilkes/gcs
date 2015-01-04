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

package com.trollworks.gcs.character;

import com.trollworks.gcs.weapon.WeaponColumn;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.Row;

import java.awt.dnd.DropTargetDragEvent;

/** An outline for weapons. */
public class WeaponOutline extends Outline {
	/**
	 * Creates a new weapon outline.
	 * 
	 * @param weaponClass The class of weapon to generate an outline for.
	 */
	public WeaponOutline(Class<? extends WeaponStats> weaponClass) {
		super(false);
		WeaponColumn.addColumns(this, weaponClass, false);
		getModel().getColumnWithID(WeaponColumn.DESCRIPTION.ordinal()).setSortCriteria(0, true);
		setEnabled(false);
	}

	@Override
	protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
		return false;
	}
}
