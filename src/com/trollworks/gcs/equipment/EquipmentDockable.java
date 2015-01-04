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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.library.LibraryDockable;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** A list of equipment from a library. */
public class EquipmentDockable extends LibraryDockable {
	@Localize("Untitled Equipment")
	@Localize(locale = "de", value = "Unbenannte Ausrüstungs-Liste")
	@Localize(locale = "ru", value = "Безымянное снаряжение")
	private static String	UNTITLED;

	static {
		Localization.initialize();
	}

	/** Creates a new {@link EquipmentDockable}. */
	public EquipmentDockable(EquipmentList list) {
		super(list);
	}

	@Override
	public EquipmentList getDataFile() {
		return (EquipmentList) super.getDataFile();
	}

	@Override
	protected String getUntitledBaseName() {
		return UNTITLED;
	}

	@Override
	protected ListOutline createOutline() {
		EquipmentList list = getDataFile();
		list.addTarget(this, Equipment.ID_CATEGORY);
		return new EquipmentOutline(list, list.getModel());
	}
}
