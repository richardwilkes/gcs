/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.library.LibraryDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

/** A list of equipment from a library. */
public class EquipmentDockable extends LibraryDockable {
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
        return I18n.Text("Untitled Equipment");
    }

    @Override
    protected ListOutline createOutline() {
        EquipmentList list = getDataFile();
        list.addTarget(this, Equipment.ID_CATEGORY);
        return new EquipmentOutline(list, list.getModel());
    }
}
